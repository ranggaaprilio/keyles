package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	revokedClientKeyPrefix = "revoked_client:"
	revokedClientTTL       = 900 * time.Second // 15 minutes = max access token lifetime
)

// RevokedClientCache provides blacklisting for deleted/revoked clients
type RevokedClientCache struct {
	client *redis.Client
}

// NewRevokedClientCache creates a new RevokedClientCache
func NewRevokedClientCache(client *redis.Client) *RevokedClientCache {
	return &RevokedClientCache{client: client}
}

// Revoke adds a client to the revocation blacklist
func (c *RevokedClientCache) Revoke(ctx context.Context, clientID string) error {
	key := c.key(clientID)
	return c.client.Set(ctx, key, "1", revokedClientTTL).Err()
}

// IsRevoked checks if a client has been revoked/deleted
func (c *RevokedClientCache) IsRevoked(ctx context.Context, clientID string) (bool, error) {
	key := c.key(clientID)
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check revoked client: %w", err)
	}
	return count > 0, nil
}

func (c *RevokedClientCache) key(clientID string) string {
	return fmt.Sprintf("%s%s", revokedClientKeyPrefix, clientID)
}
