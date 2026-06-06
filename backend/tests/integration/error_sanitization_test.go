package integration

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/interfaces/http/middleware"
	"github.com/stretchr/testify/assert"
)

func TestErrorSanitization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.ErrorHandler())

	router.GET("/error/email", func(c *gin.Context) {
		c.Error(errors.New("User user@example.com not found"))
		c.Status(http.StatusBadRequest)
	})

	router.GET("/error/path", func(c *gin.Context) {
		c.Error(errors.New("File /Users/admin/data.txt not found"))
		c.Status(http.StatusInternalServerError)
	})

	router.GET("/error/sensitive", func(c *gin.Context) {
		c.Error(errors.New("Database password is invalid"))
		c.Status(http.StatusInternalServerError)
	})

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		shouldNotContain []string
	}{
		{
			name:           "email masked",
			path:           "/error/email",
			expectedStatus: http.StatusBadRequest,
			shouldNotContain: []string{"user@example.com"},
		},
		{
			name:           "path redacted",
			path:           "/error/path",
			expectedStatus: http.StatusInternalServerError,
			shouldNotContain: []string{"/Users/admin"},
		},
		{
			name:           "sensitive pattern hidden",
			path:           "/error/sensitive",
			expectedStatus: http.StatusInternalServerError,
			shouldNotContain: []string{"password"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			body := w.Body.String()
			for _, s := range tt.shouldNotContain {
				assert.NotContains(t, body, s, "Response should not contain sensitive data: %s", s)
			}
		})
	}
}
