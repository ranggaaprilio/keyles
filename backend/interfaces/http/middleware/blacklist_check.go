package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/domain/services"
)

// BlacklistCheckMiddleware creates a middleware that checks if the user is blacklisted
func BlacklistCheckMiddleware(blacklist services.UserBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok || userIDStr == "" {
			c.Next()
			return
		}

		isBlacklisted, err := blacklist.IsBlacklisted(c.Request.Context(), userIDStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "server_error",
			})
			c.Abort()
			return
		}

		if isBlacklisted {
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
