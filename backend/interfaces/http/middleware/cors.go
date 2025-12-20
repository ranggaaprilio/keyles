package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS middleware handles Cross-Origin Resource Sharing
func CORS(allowedOrigins, allowedMethods, allowedHeaders string) gin.HandlerFunc {
	origins := strings.Split(allowedOrigins, ",")
	methods := strings.Split(allowedMethods, ",")
	headers := strings.Split(allowedHeaders, ",")

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowedOrigin := ""
		for _, o := range origins {
			trimmedOrigin := strings.TrimSpace(o)
			if trimmedOrigin == "*" {
				allowedOrigin = "*"
				break
			}
			if trimmedOrigin == origin {
				allowedOrigin = origin
				break
			}
		}

		// Set CORS headers
		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Header("Access-Control-Allow-Methods", strings.Join(methods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(headers, ", "))
		c.Header("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
