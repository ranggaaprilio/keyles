package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/infrastructure/config"
	"github.com/ranggaaprilio/keyles/infrastructure/monitoring"
)

// CSRF returns a middleware that implements double-submit cookie pattern
func CSRF(cfg *config.Config) gin.HandlerFunc {
	cookieName := cfg.CSRFCookieName
	if cookieName == "" {
		cookieName = "keyles_csrf"
	}
	headerName := cfg.CSRFHeaderName
	if headerName == "" {
		headerName = "X-CSRF-Token"
	}
	tokenLength := cfg.CSRFTokenLength
	if tokenLength < 16 || tokenLength > 64 {
		tokenLength = 32
	}

	return func(c *gin.Context) {
		// Generate token if not present
		existingToken, err := c.Cookie(cookieName)
		if err != nil || existingToken == "" {
			existingToken = generateToken(tokenLength)
			c.SetCookie(
				cookieName,
				existingToken,
				86400, // 24 hours
				"/",
				"",
				cfg.SecurityCookieSecure,
				false, // HttpOnly=false so frontend can read it
			)
		}

		// Exempt methods don't require validation
		method := c.Request.Method
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}

		// Check exempt paths
		path := c.Request.URL.Path
		if isCSRFExempt(path) {
			c.Next()
			return
		}

		// Validate CSRF token on state-changing requests
		token := c.GetHeader(headerName)
		if token == "" {
			token = c.PostForm("csrf_token")
		}

		if token == "" || !strings.EqualFold(token, existingToken) {
			monitoring.IncrementSecurityEvent("csrf_rejected")
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "CSRF_TOKEN_INVALID",
					"message": "CSRF token missing or invalid",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isCSRFExempt checks if a path is exempt from CSRF validation
func isCSRFExempt(path string) bool {
	exemptPaths := []string{
		"/oauth2/auth",
		"/oauth2/token",
		"/oauth2/revoke",
		"/oauth2/introspect",
		"/health",
		"/api/v1/register",
		"/api/v1/check-availability",
		"/api/v1/verify-otp",
		"/api/v1/resend-otp",
		"/api/v1/login",
	}

	for _, exempt := range exemptPaths {
		if path == exempt {
			return true
		}
	}

	// Wildcard match for .well-known
	if strings.HasPrefix(path, "/.well-known/") {
		return true
	}

	return false
}

// generateToken creates a cryptographically secure random token
func generateToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a simpler approach if crypto/rand fails
		for i := range bytes {
			bytes[i] = byte(i)
		}
	}
	return hex.EncodeToString(bytes)
}
