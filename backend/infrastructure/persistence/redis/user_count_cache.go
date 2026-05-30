package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	userCountKeyPrefix = "user_count:"
	userCountTTL       = 60 * time.Second
)

// UserCountCache provides caching for tenant user counts
type UserCountCache struct {
	client *redis.Client
}

// NewUserCountCache creates a new UserCountCache
func NewUserCountCache(client *redis.Client) *UserCountCache {
	return &UserCountCache{client: client}
}

// Get retrieves the cached user count for a tenant
// Returns -1 if not cached (cache miss)
func (c *UserCountCache) Get(ctx context.Context, tenantID string) (int, error) {
	key := c.key(tenantID)
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return -1, nil // cache miss
		}
		return -1, fmt.Errorf("failed to get user count from cache: %w", err)
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return -1, fmt.Errorf("invalid cached count value: %w", err)
	}
	return count, nil
}

// Set stores the user count for a tenant in cache
func (c *UserCountCache) Set(ctx context.Context, tenantID string, count int) error {
	key := c.key(tenantID)
	return c.client.Set(ctx, key, strconv.Itoa(count), userCountTTL).Err()
}

// Invalidate removes the cached user count for a tenant
func (c *UserCountCache) Invalidate(ctx context.Context, tenantID string) error {
	key := c.key(tenantID)
	return c.client.Del(ctx, key).Err()
}

func (c *UserCountCache) key(tenantID string) string {
	return fmt.Sprintf("%s%s", userCountKeyPrefix, tenantID)
}
