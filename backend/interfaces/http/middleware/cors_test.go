package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newCORSTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS("https://admin.example.com", "GET,POST,OPTIONS", "Origin,Content-Type,Accept,Authorization"))
	ok := func(c *gin.Context) { c.Status(http.StatusOK) }
	router.GET("/.well-known/openid-configuration", ok)
	router.GET("/.well-known/jwks.json", ok)
	router.POST("/oauth2/token", ok)
	router.GET("/oauth2/userinfo", ok)
	router.POST("/oauth2/login", ok)
	router.GET("/api/v1/clients", ok)
	return router
}

func corsRequest(router *gin.Engine, method, path, origin string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Origin", origin)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestCORSPublicOIDCEndpointsAllowAnyOriginWithoutCredentials(t *testing.T) {
	router := newCORSTestRouter()

	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/.well-known/openid-configuration"},
		{http.MethodGet, "/.well-known/jwks.json"},
		{http.MethodPost, "/oauth2/token"},
		{http.MethodGet, "/oauth2/userinfo"},
		{http.MethodOptions, "/oauth2/token"},
	} {
		w := corsRequest(router, tc.method, tc.path, "http://localhost:5173")
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"), tc.path)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"), tc.path)
	}
}

func TestCORSPrivateEndpointsUseAllowlist(t *testing.T) {
	router := newCORSTestRouter()

	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, "/oauth2/login"},
		{http.MethodGet, "/api/v1/clients"},
	} {
		w := corsRequest(router, tc.method, tc.path, "http://localhost:5173")
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"), tc.path)

		w = corsRequest(router, tc.method, tc.path, "https://admin.example.com")
		assert.Equal(t, "https://admin.example.com", w.Header().Get("Access-Control-Allow-Origin"), tc.path)
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"), tc.path)
	}
}
