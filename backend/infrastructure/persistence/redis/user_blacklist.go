package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	userBlacklistKeyPrefix = "user:blacklisted:"
)

// UserBlacklist provides a Redis-backed blacklist for disabled user accounts.
// When a user is disabled, their ID is added with a TTL equal to the maximum
// access-token lifetime so that in-flight tokens are rejected without a DB lookup.
type UserBlacklist struct {
	client *redis.Client
}

// NewUserBlacklist creates a new UserBlacklist
func NewUserBlacklist(client *redis.Client) *UserBlacklist {
	return &UserBlacklist{client: client}
}

// Add places a user ID on the blacklist for the given duration.
func (b *UserBlacklist) Add(ctx context.Context, userID string, ttl time.Duration) error {
	key := b.key(userID)
	return b.client.Set(ctx, key, "1", ttl).Err()
}

// IsBlacklisted returns true if the user is currently on the blacklist.
func (b *UserBlacklist) IsBlacklisted(ctx context.Context, userID string) (bool, error) {
	key := b.key(userID)
	count, err := b.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check user blacklist: %w", err)
	}
	return count > 0, nil
}

func (b *UserBlacklist) key(userID string) string {
	return fmt.Sprintf("%s%s", userBlacklistKeyPrefix, userID)
}
