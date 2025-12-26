package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

// RedisRateLimiter implements rate limiting using Redis
type RedisRateLimiter struct {
	client  *redis.Client
	limiter *limiter.Limiter
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
// rate is the max number of requests per duration (e.g., "10-M" for 10 per minute)
func NewRedisRateLimiter(client *redis.Client, rate limiter.Rate) (*RedisRateLimiter, error) {
	store, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:   "oauth:ratelimit",
		MaxRetry: 3,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create rate limiter store: %w", err)
	}

	instance := limiter.New(store, rate)

	return &RedisRateLimiter{
		client:  client,
		limiter: instance,
	}, nil
}

// rateLimitKey generates the Redis key for rate limiting
func (r *RedisRateLimiter) rateLimitKey(clientID string) string {
	return fmt.Sprintf("oauth:ratelimit:%s", clientID)
}

// Allow checks if a request is allowed for the given client ID
func (r *RedisRateLimiter) Allow(ctx context.Context, clientID string) (bool, error) {
	key := r.rateLimitKey(clientID)
	
	limiterCtx, err := r.limiter.Get(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}

	return limiterCtx.Reached == false, nil
}

// GetLimit returns the current rate limit status for a client
func (r *RedisRateLimiter) GetLimit(ctx context.Context, clientID string) (limit int64, remaining int64, resetTime int64, err error) {
	key := r.rateLimitKey(clientID)
	
	limiterCtx, err := r.limiter.Get(ctx, key)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get rate limit: %w", err)
	}

	return limiterCtx.Limit, limiterCtx.Remaining, limiterCtx.Reset, nil
}

// Peek returns the current rate limit status without incrementing the counter
func (r *RedisRateLimiter) Peek(ctx context.Context, clientID string) (limit int64, remaining int64, resetTime int64, err error) {
	key := r.rateLimitKey(clientID)
	
	limiterCtx, err := r.limiter.Peek(ctx, key)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to peek rate limit: %w", err)
	}

	return limiterCtx.Limit, limiterCtx.Remaining, limiterCtx.Reset, nil
}

// Reset resets the rate limit for a client
func (r *RedisRateLimiter) Reset(ctx context.Context, clientID string) error {
	key := r.rateLimitKey(clientID)
	_, err := r.limiter.Reset(ctx, key)
	return err
}
