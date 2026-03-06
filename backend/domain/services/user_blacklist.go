package services

import (
	"context"
	"time"
)

// UserBlacklist provides immediate token invalidation for deleted or disabled users.
// When a user is disabled or deleted, their user ID is added to the blacklist with a TTL
// matching the maximum access token lifetime (900 seconds / 15 minutes).
//
// Redis key pattern: user_blacklist:{user_id}
// TTL: 900 seconds (matches AccessTokenTTL)
//
// The blacklist is checked on every authenticated request by the BlacklistCheckMiddleware.
// Keys expire automatically; no manual cleanup is required.
type UserBlacklist interface {
	// Add adds a user to the blacklist with the given TTL
	Add(ctx context.Context, userID string, ttl time.Duration) error

	// IsBlacklisted checks if a user is in the blacklist
	IsBlacklisted(ctx context.Context, userID string) (bool, error)
}
