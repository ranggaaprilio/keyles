package repositories

import (
	"context"

	"github.com/ranggaaprilio/keyles/domain/entities"
)

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
}
