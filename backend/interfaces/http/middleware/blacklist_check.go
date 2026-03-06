package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/domain/services"
)

// BlacklistCheckMiddleware checks if the authenticated user has been blacklisted
// (disabled or deleted). It should execute after the auth middleware has set the
// user_id in the Gin context. Blacklisted users receive an immediate 401.
func BlacklistCheckMiddleware(blacklist services.UserBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			// No user_id in context — let auth middleware handle it
			c.Next()
			return
		}

		uid, ok := userID.(string)
		if !ok || uid == "" {
			c.Next()
			return
		}

		blacklisted, err := blacklist.IsBlacklisted(c.Request.Context(), uid)
		if err != nil {
			// On Redis failure, fail open to avoid blocking all authenticated requests.
			// The access token TTL (15min) provides a bounded window of risk.
			c.Next()
			return
		}

		if blacklisted {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":             "token_invalid",
				"error_description": "account has been revoked",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
