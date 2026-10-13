package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole allows the request through only when the JWT role set by
// AuthMiddleware is one of the given roles. Otherwise it responds 403.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		if _, ok := allowed[c.GetString("role")]; !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Admin access required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
