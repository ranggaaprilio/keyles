package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/infrastructure/config"
	"github.com/ranggaaprilio/keyles/interfaces/http/middleware"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeadersOnAllResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		SecurityHeadersCSP:  "default-src 'self'; script-src 'self';",
		SecurityHeadersHSTS: "max-age=31536000; includeSubDomains",
	}

	router := gin.New()
	router.Use(middleware.SecurityHeaders(cfg))

	endpoints := []string{"/health", "/api/v1/login", "/oauth2/token", "/.well-known/openid-configuration"}
	for _, endpoint := range endpoints {
		router.GET(endpoint, func(c *gin.Context) {
			c.String(http.StatusOK, "OK")
		})
	}

	for _, endpoint := range endpoints {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", endpoint, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Endpoint %s should return 200", endpoint)
		assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"), "Endpoint %s should have CSP header", endpoint)
		assert.NotEmpty(t, w.Header().Get("Strict-Transport-Security"), "Endpoint %s should have HSTS header", endpoint)
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"), "Endpoint %s should have X-Frame-Options header", endpoint)
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"), "Endpoint %s should have X-Content-Type-Options header", endpoint)
		assert.NotEmpty(t, w.Header().Get("Permissions-Policy"), "Endpoint %s should have Permissions-Policy header", endpoint)
		assert.Equal(t, "same-origin", w.Header().Get("Cross-Origin-Opener-Policy"), "Endpoint %s should have COOP header", endpoint)
		assert.Equal(t, "require-corp", w.Header().Get("Cross-Origin-Embedder-Policy"), "Endpoint %s should have COEP header", endpoint)
	}
}
