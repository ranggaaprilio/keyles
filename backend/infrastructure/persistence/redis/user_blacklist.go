package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/ranggaaprilio/keyles/domain/services"
	"github.com/redis/go-redis/v9"
)

const (
	userBlacklistKeyPrefix = "user_blacklist:"
)

// UserBlacklistCache provides Redis-backed user blacklisting
type UserBlacklistCache struct {
	client *redis.Client
}

// NewUserBlacklistCache creates a new UserBlacklistCache
func NewUserBlacklistCache(client *redis.Client) services.UserBlacklist {
	return &UserBlacklistCache{client: client}
}

// Add blacklists a user for the specified TTL duration
func (c *UserBlacklistCache) Add(ctx context.Context, userID string, ttl time.Duration) error {
	key := c.key(userID)
	return c.client.Set(ctx, key, "1", ttl).Err()
}

// IsBlacklisted checks if a user is currently blacklisted
func (c *UserBlacklistCache) IsBlacklisted(ctx context.Context, userID string) (bool, error) {
	key := c.key(userID)
	val, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check user blacklist: %w", err)
	}
	return val > 0, nil
}

func (c *UserBlacklistCache) key(userID string) string {
	return fmt.Sprintf("%s%s", userBlacklistKeyPrefix, userID)
}

// Verify interface compliance
var _ services.UserBlacklist = (*UserBlacklistCache)(nil)
