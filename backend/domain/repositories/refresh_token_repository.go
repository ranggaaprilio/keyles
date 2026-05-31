package repositories

import (
	"context"
	"errors"

	"github.com/ranggaaprilio/keyles/domain/entities"
)

var ErrRefreshTokenReplay = errors.New("refresh token replay detected")

// RefreshTokenRepository defines the interface for refresh token data access
type RefreshTokenRepository interface {
	// Create stores a new refresh token
	Create(ctx context.Context, token *entities.RefreshToken) error

	// GetByToken retrieves a refresh token by its token value (hashed)
	GetByToken(ctx context.Context, tokenHash string) (*entities.RefreshToken, error)

	// Revoke marks a refresh token as revoked
	Revoke(ctx context.Context, tokenHash string, revokedBy string) error

	// RevokeAllForUser revokes all refresh tokens for a user-client combination
	RevokeAllForUser(ctx context.Context, userID string, clientID string) error

	// DeleteExpired removes expired tokens from the database
	// Should be called periodically via cron job
	DeleteExpired(ctx context.Context) (int64, error)

	// IsRevoked checks if a token is revoked
	IsRevoked(ctx context.Context, tokenHash string) (bool, error)

	// UpdateLastUsed updates the last_used_at timestamp
	UpdateLastUsed(ctx context.Context, tokenHash string) error

	// RevokeByClientID revokes all refresh tokens issued to a specific client
	RevokeByClientID(ctx context.Context, clientID string) error

	// RevokeByUserID revokes all active refresh tokens for a user across all client applications.
	// Used by disable_user and delete_user use cases.
	RevokeByUserID(ctx context.Context, userID string) error

	// ListByUserID returns active, unexpired refresh tokens for a user.
	// Used by list_sessions use case to display current sessions.
	ListByUserID(ctx context.Context, userID string) ([]*entities.RefreshToken, error)

	// GetByID retrieves a refresh token by its database ID.
	// Used by revoke_session use case.
	GetByID(ctx context.Context, id int64) (*entities.RefreshToken, error)
}

// RefreshTokenRotationRepository atomically replaces single-use refresh tokens.
// A replay must revoke all active descendants in the token family.
type RefreshTokenRotationRepository interface {
	Rotate(ctx context.Context, currentTokenHash string, replacement *entities.RefreshToken) error
	RevokeFamily(ctx context.Context, familyID string, reason string) error
}
