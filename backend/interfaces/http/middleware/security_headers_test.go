package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/infrastructure/config"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		SecurityHeadersCSP:  "default-src 'self'; script-src 'self';",
		SecurityHeadersHSTS: "max-age=31536000; includeSubDomains",
	}

	router := gin.New()
	router.Use(SecurityHeaders(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "default-src 'self'; script-src 'self';", w.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "camera=(), microphone=(), geolocation=(), payment=()", w.Header().Get("Permissions-Policy"))
	assert.Equal(t, "same-origin", w.Header().Get("Cross-Origin-Opener-Policy"))
	assert.Equal(t, "require-corp", w.Header().Get("Cross-Origin-Embedder-Policy"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}

func TestSecurityHeadersDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{}

	router := gin.New()
	router.Use(SecurityHeaders(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
	assert.NotEmpty(t, w.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
}
