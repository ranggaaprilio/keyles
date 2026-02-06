package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/infrastructure/persistence/redis"
)

// RateLimiterMiddleware implements rate limiting per client_id
// Implements FR-057: 10 requests/minute per client_id on token endpoint
func RateLimiterMiddleware(rateLimiter *redis.RedisRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract client_id from request
		clientID := extractClientID(c)
		if clientID == "" {
			// If no client_id, allow the request (will be caught by auth)
			c.Next()
			return
		}

		// Check rate limit
		allowed, err := rateLimiter.Allow(c.Request.Context(), clientID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":             "rate_limit_error",
				"error_description": "Failed to check rate limit",
			})
			c.Abort()
			return
		}

		if !allowed {
			// Get rate limit details for response headers
			limit, remaining, reset, _ := rateLimiter.GetLimit(c.Request.Context(), clientID)

			// Set rate limit headers
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", reset))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":             "rate_limit_exceeded",
				"error_description": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractClientID extracts the client_id from the request
func extractClientID(c *gin.Context) string {
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
