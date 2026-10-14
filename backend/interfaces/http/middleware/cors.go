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
		if isPublicOIDCPath(c.Request.URL.Path) {
			// Any relying party may call these; they never rely on cookies, so no credentials.
			c.Header("Access-Control-Allow-Origin", "*")
		} else if allowedOrigin != "" {
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

// isPublicOIDCPath reports whether a path is an OIDC endpoint that relying parties
// call cross-origin with bearer tokens or client credentials rather than cookies.
func isPublicOIDCPath(path string) bool {
	switch path {
	case "/oauth2/token", "/oauth2/userinfo", "/oauth2/revoke":
		return true
	}
	return strings.HasPrefix(path, "/.well-known/")
}
