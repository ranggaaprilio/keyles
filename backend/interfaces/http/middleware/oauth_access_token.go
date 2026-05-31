package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/usecase/auth"
)

func OAuthAccessTokenMiddleware(validator *auth.AccessTokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		if validator == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			c.Abort()
			return
		}

		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			c.Abort()
			return
		}

		claims, err := validator.Validate(c.Request.Context(), strings.TrimPrefix(header, "Bearer "), "")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.Subject)
		c.Set("tenant_id", claims.TenantID)
		c.Set("client_id", claims.ClientID)
		c.Set("scope", claims.Scope)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}
