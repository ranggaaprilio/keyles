package repositories

import (
	"context"
	"time"
)

// Session represents a user session stored in Redis
type Session struct {
	SessionID string
	UserID    string
	TenantID  string
	ClientID  string
	CreatedAt time.Time
	ExpiresAt time.Time
	Metadata  map[string]interface{}
}

// SessionRepository defines the interface for session storage in Redis
// Sessions are ephemeral with 8-hour TTL
type SessionRepository interface {
	// Create stores a new session with automatic expiration
	Create(ctx context.Context, session *Session, ttl time.Duration) error

	// Get retrieves a session by session ID
	Get(ctx context.Context, sessionID string) (*Session, error)

	// Delete removes a session (logout)
	Delete(ctx context.Context, sessionID string) error

	// Exists checks if a session exists and is valid
	Exists(ctx context.Context, sessionID string) (bool, error)

	// Extend extends the TTL of an existing session
	Extend(ctx context.Context, sessionID string, ttl time.Duration) error
}
