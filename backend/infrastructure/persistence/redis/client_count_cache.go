package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	clientCountKeyPrefix = "client_count:"
	clientCountTTL       = 60 * time.Second
)

// ClientCountCache provides caching for tenant client counts
type ClientCountCache struct {
	client *redis.Client
}

// NewClientCountCache creates a new ClientCountCache
func NewClientCountCache(client *redis.Client) *ClientCountCache {
	return &ClientCountCache{client: client}
}

// Get retrieves the cached client count for a tenant
// Returns -1 if not cached (cache miss)
func (c *ClientCountCache) Get(ctx context.Context, tenantID string) (int, error) {
	key := c.key(tenantID)
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return -1, nil // cache miss
		}
		return -1, fmt.Errorf("failed to get client count from cache: %w", err)
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return -1, fmt.Errorf("invalid cached count value: %w", err)
	}
	return count, nil
}

// Set stores the client count for a tenant in cache
func (c *ClientCountCache) Set(ctx context.Context, tenantID string, count int) error {
	key := c.key(tenantID)
	return c.client.Set(ctx, key, strconv.Itoa(count), clientCountTTL).Err()
}

// Invalidate removes the cached client count for a tenant
func (c *ClientCountCache) Invalidate(ctx context.Context, tenantID string) error {
	key := c.key(tenantID)
	return c.client.Del(ctx, key).Err()
}

func (c *ClientCountCache) key(tenantID string) string {
	return fmt.Sprintf("%s%s", clientCountKeyPrefix, tenantID)
}
