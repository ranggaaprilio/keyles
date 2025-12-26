package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisTokenCacheRepository implements token caching using Redis
type RedisTokenCacheRepository struct {
	client *redis.Client
}

// NewRedisTokenCacheRepository creates a new Redis token cache repository
func NewRedisTokenCacheRepository(client *redis.Client) *RedisTokenCacheRepository {
	return &RedisTokenCacheRepository{client: client}
}

// tokenCacheKey generates the Redis key for a cached token
func (r *RedisTokenCacheRepository) tokenCacheKey(tokenHash string) string {
	return fmt.Sprintf("oauth:token:cache:%s", tokenHash)
}

// Set caches a refresh token with TTL
func (r *RedisTokenCacheRepository) Set(ctx context.Context, tokenHash string, userID string, ttl time.Duration) error {
	key := r.tokenCacheKey(tokenHash)
	return r.client.Set(ctx, key, userID, ttl).Err()
}

// Get retrieves a cached token
func (r *RedisTokenCacheRepository) Get(ctx context.Context, tokenHash string) (string, error) {
	key := r.tokenCacheKey(tokenHash)
	
	userID, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // Token not in cache
		}
		return "", fmt.Errorf("failed to get token from cache: %w", err)
	}

	return userID, nil
}

// Delete removes a token from cache
func (r *RedisTokenCacheRepository) Delete(ctx context.Context, tokenHash string) error {
	key := r.tokenCacheKey(tokenHash)
	return r.client.Del(ctx, key).Err()
}

// Exists checks if a token exists in cache
func (r *RedisTokenCacheRepository) Exists(ctx context.Context, tokenHash string) (bool, error) {
	key := r.tokenCacheKey(tokenHash)
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
