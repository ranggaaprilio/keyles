package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
)

var ErrAuthorizationCodeUnavailable = errors.New("authorization code is missing, expired, or already consumed")

// AuthCodeRepository defines the interface for authorization code storage in Redis
// Authorization codes are ephemeral with 5-minute TTL
type AuthCodeRepository interface {
	// Store saves an authorization code with automatic expiration
	Store(ctx context.Context, code *entities.AuthorizationCode, ttl time.Duration) error

	// Get retrieves an authorization code
	Get(ctx context.Context, code string) (*entities.AuthorizationCode, error)

	// MarkAsUsed marks an authorization code as used (prevents replay attacks)
	MarkAsUsed(ctx context.Context, code string) error

	// Delete removes an authorization code
	Delete(ctx context.Context, code string) error

	// Exists checks if an authorization code exists
	Exists(ctx context.Context, code string) (bool, error)
}

// AtomicAuthCodeRepository atomically consumes one-time authorization codes.
// Implementations must return an error when the code is missing, expired, or
// has already been consumed.
type AtomicAuthCodeRepository interface {
	Consume(ctx context.Context, code string) (*entities.AuthorizationCode, error)
}
