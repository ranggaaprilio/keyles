package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		role string
		want int
	}{
		{"admin", http.StatusOK},
		{"owner", http.StatusOK},
		{"member", http.StatusForbidden},
		{"viewer", http.StatusForbidden},
		{"", http.StatusForbidden},
	}
	for _, tc := range cases {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			if tc.role != "" {
				c.Set("role", tc.role)
			}
		})
		router.Use(RequireRole("admin", "owner"))
		router.GET("/admin", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, tc.want, w.Code, "role=%q", tc.role)
	}
}
