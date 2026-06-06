package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/infrastructure/config"
	"github.com/stretchr/testify/assert"
)

func TestCSRFMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CSRFCookieName:  "keyles_csrf",
		CSRFHeaderName:  "X-CSRF-Token",
		CSRFTokenLength: 32,
	}

	router := gin.New()
	router.Use(CSRF(cfg))
	router.POST("/api/v1/admin/clients", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/admin/clients", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "CSRF_TOKEN_INVALID")
}

func TestCSRFValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CSRFCookieName:  "keyles_csrf",
		CSRFHeaderName:  "X-CSRF-Token",
		CSRFTokenLength: 32,
	}

	router := gin.New()
	router.Use(CSRF(cfg))
	router.POST("/api/v1/admin/clients", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// First request to get the CSRF cookie
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/api/v1/admin/clients", nil)
	router.ServeHTTP(w1, req1)

	cookie := w1.Header().Get("Set-Cookie")
	assert.NotEmpty(t, cookie)

	// Extract token from cookie
	token := extractTokenFromCookie(cookie)
	assert.NotEmpty(t, token)

	// Second request with valid token
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/v1/admin/clients", nil)
	req2.Header.Set("X-CSRF-Token", token)
	req2.Header.Set("Cookie", cookie)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestCSRFMismatchedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CSRFCookieName:  "keyles_csrf",
		CSRFHeaderName:  "X-CSRF-Token",
		CSRFTokenLength: 32,
	}

	router := gin.New()
	router.Use(CSRF(cfg))
	router.POST("/api/v1/admin/clients", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/admin/clients", nil)
	req.Header.Set("X-CSRF-Token", "invalid-token")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "CSRF_TOKEN_INVALID")
}

func TestCSRFGETExempt(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CSRFCookieName:  "keyles_csrf",
		CSRFHeaderName:  "X-CSRF-Token",
		CSRFTokenLength: 32,
	}

	router := gin.New()
	router.Use(CSRF(cfg))
	router.GET("/api/v1/admin/clients", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/clients", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFHEADExempt(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CSRFCookieName:  "keyles_csrf",
		CSRFHeaderName:  "X-CSRF-Token",
		CSRFTokenLength: 32,
	}

	router := gin.New()
	router.Use(CSRF(cfg))
	router.HEAD("/api/v1/admin/clients", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("HEAD", "/api/v1/admin/clients", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFOPTIONSExempt(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CSRFCookieName:  "keyles_csrf",
		CSRFHeaderName:  "X-CSRF-Token",
		CSRFTokenLength: 32,
	}

	router := gin.New()
	router.Use(CSRF(cfg))
	router.OPTIONS("/api/v1/admin/clients", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/v1/admin/clients", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFExemptPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CSRFCookieName:  "keyles_csrf",
		CSRFHeaderName:  "X-CSRF-Token",
		CSRFTokenLength: 32,
	}

	exemptPaths := []string{
		"/oauth2/auth",
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
		router.Use(CSRF(cfg))
		router.POST(path, func(c *gin.Context) {
			c.String(http.StatusOK, "OK")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", path, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Path %s should be exempt from CSRF", path)
	}
}

func extractTokenFromCookie(cookie string) string {
	// Simple extraction: keyles_csrf=<token>; ...
	parts := strings.Split(cookie, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "keyles_csrf=") {
			return strings.TrimPrefix(part, "keyles_csrf=")
		}
	}
	return ""
}
