package repositories

import (
	"context"
	"time"
)

// Session represents a client-agnostic end-user SSO session stored in Redis.
// Sessions are ephemeral with a configurable TTL (default 8 hours).
// A single session can be reused across multiple OAuth clients in the same tenant
// as long as the end-user remains active and has a valid role assignment.
type Session struct {
	SessionID      string
	UserID         string
	TenantID       string
	AuthenticatedAt time.Time
	CreatedAt      time.Time
	ExpiresAt      time.Time
	Metadata       map[string]interface{}
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