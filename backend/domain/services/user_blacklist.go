package services

import (
	"context"
	"time"
)

// UserBlacklist defines the interface for managing blacklisted users.
// The Redis key pattern is user_blacklist:{user_id} with 900s TTL
// matching the max access token lifetime per FR-037.
type UserBlacklist interface {
	// Add blacklists a user for the specified TTL duration
	Add(ctx context.Context, userID string, ttl time.Duration) error

	// IsBlacklisted checks if a user is currently blacklisted
	IsBlacklisted(ctx context.Context, userID string) (bool, error)
}
