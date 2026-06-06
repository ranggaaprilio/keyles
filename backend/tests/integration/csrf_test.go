package integration

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/infrastructure/config"
	"github.com/ranggaaprilio/keyles/interfaces/http/middleware"
	"github.com/stretchr/testify/assert"
)

func TestCSRFProtectionOnAdminEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CSRFCookieName:       "keyles_csrf",
		CSRFHeaderName:       "X-CSRF-Token",
		CSRFTokenLength:      32,
		SecurityCookieSecure: false,
	}

	router := gin.New()
	router.Use(middleware.CSRF(cfg))
	router.POST("/api/v1/admin/clients", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Without CSRF token should fail
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/admin/clients", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "CSRF_TOKEN_INVALID")

	// With valid CSRF token should succeed
	cookie := w.Header().Get("Set-Cookie")
	var token string
	if cookie != "" {
		parts := strings.Split(cookie, ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "keyles_csrf=") {
				token = strings.TrimPrefix(part, "keyles_csrf=")
			}
		}
	}

	if token != "" {
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("POST", "/api/v1/admin/clients", nil)
		req2.Header.Set("X-CSRF-Token", token)
		req2.Header.Set("Cookie", "keyles_csrf="+token)
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)
	}
}

func TestCSRFExemptionOnOAuthEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CSRFCookieName:       "keyles_csrf",
		CSRFHeaderName:       "X-CSRF-Token",
		CSRFTokenLength:      32,
		SecurityCookieSecure: false,
	}

	exemptPaths := []string{
		"/oauth2/token",
		"/oauth2/revoke",
		"/oauth2/introspect",
		"/health",
		"/.well-known/openid-configuration",
		"/api/v1/register",
		"/api/v1/check-availability",
		"/api/v1/verify-otp",
		"/api/v1/resend-otp",
		"/api/v1/login",
	}

	for _, path := range exemptPaths {
		router := gin.New()
		router.Use(middleware.CSRF(cfg))
		router.POST(path, func(c *gin.Context) {
			c.String(http.StatusOK, "OK")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", path, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Path %s should be exempt from CSRF", path)
	}
}
