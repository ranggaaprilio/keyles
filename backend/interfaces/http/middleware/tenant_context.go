package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// TenantContextKey is the context key for tenant ID
type TenantContextKey string

const (
	// TenantIDKey is the context key for storing tenant ID
	TenantIDKey TenantContextKey = "tenant_id"
)

// TenantContextMiddleware extracts tenant_id from client_id and adds to request context
func TenantContextMiddleware(clientRepo repositories.ClientRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract client_id from request
		clientID := extractClientIDForTenant(c)
		if clientID == "" {
			// No client_id, continue without setting tenant context
			c.Next()
			return
		}

		// Look up client to get tenant_id
		client, err := clientRepo.GetByID(c.Request.Context(), clientID)
		if err != nil || client == nil {
			// Client not found or error, continue without setting tenant context
			// The actual authentication will handle invalid clients
			c.Next()
			return
		}

		// Add tenant_id to context
		ctx := context.WithValue(c.Request.Context(), TenantIDKey, client.TenantID)
		c.Request = c.Request.WithContext(ctx)

		// Also store in Gin context for easier access
		c.Set("tenant_id", client.TenantID)
		c.Set("client_id", clientID)

		c.Next()
	}
}

// GetTenantIDFromContext retrieves tenant ID from context
func GetTenantIDFromContext(ctx context.Context) string {
	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok {
		return tenantID
	}
	return ""
}

// GetTenantIDFromGinContext retrieves tenant ID from Gin context
func GetTenantIDFromGinContext(c *gin.Context) string {
	if tenantID, exists := c.Get("tenant_id"); exists {
		if tid, ok := tenantID.(string); ok {
			return tid
		}
	}
	return ""
}

// extractClientIDForTenant extracts the client_id from the request
func extractClientIDForTenant(c *gin.Context) string {
	// Try to get from form data (token endpoint)
	if clientID := c.PostForm("client_id"); clientID != "" {
		return clientID
	}

	// Try to get from query parameters (authorization endpoint)
	if clientID := c.Query("client_id"); clientID != "" {
		return clientID
	}

	// Try to get from Basic Auth
	clientID, _, ok := c.Request.BasicAuth()
	if ok {
		return clientID
	}

	// Try to get from JSON body
	var body struct {
		ClientID string `json:"client_id"`
	}
	if err := c.ShouldBindJSON(&body); err == nil && body.ClientID != "" {
		return body.ClientID
	}

	return ""
}
