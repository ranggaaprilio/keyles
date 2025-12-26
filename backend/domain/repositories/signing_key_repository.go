package repositories

import (
	"context"

	"github.com/ranggaaprilio/keyles/domain/entities"
)

// SigningKeyRepository defines the interface for signing key data access
type SigningKeyRepository interface {
	// Create stores a new signing key
	Create(ctx context.Context, key *entities.SigningKey) error

	// GetActive retrieves the currently active signing key for a specific algorithm
	GetActive(ctx context.Context, algorithm string) (*entities.SigningKey, error)

	// GetByKeyID retrieves a signing key by its key ID
	GetByKeyID(ctx context.Context, keyID string) (*entities.SigningKey, error)

	// ListActive retrieves all active signing keys (for JWKS endpoint)
	ListActive(ctx context.Context) ([]*entities.SigningKey, error)

	// Deactivate marks a signing key as inactive (for key rotation)
	Deactivate(ctx context.Context, keyID string) error

	// DeleteExpired removes expired keys from the database
	DeleteExpired(ctx context.Context) (int64, error)
}
