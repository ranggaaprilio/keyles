package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	redis *redis.Client
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{redis: redisClient}
}

// Limit creates a rate limiting middleware
func (rl *RateLimiter) Limit(maxRequests int, window time.Duration, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		if key == "" {
			c.Next()
			return
		}

		ctx := context.Background()
		rateLimitKey := fmt.Sprintf("ratelimit:%s", key)

		// Increment counter
		pipe := rl.redis.Pipeline()
		incr := pipe.Incr(ctx, rateLimitKey)
		pipe.Expire(ctx, rateLimitKey, window)
		
		_, err := pipe.Exec(ctx)
		if err != nil {
			// Log error but don't block request if Redis fails
			c.Next()
			return
		}

		count := int(incr.Val())

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, maxRequests-count)))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

		// Check if limit exceeded
		if count > maxRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":       "RATE_LIMIT_EXCEEDED",
					"message":    "Too many requests. Please try again later.",
					"retryAfter": int(window.Seconds()),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IPBasedLimit creates IP-based rate limiting
func (rl *RateLimiter) IPBasedLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	return rl.Limit(maxRequests, window, func(c *gin.Context) string {
		return fmt.Sprintf("ip:%s:%s", c.ClientIP(), c.Request.URL.Path)
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
