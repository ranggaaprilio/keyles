package repositories

import (
	"context"
	"time"
)

// PendingSignup holds an end-user self sign-up awaiting email verification.
// The user record is only created once the emailed code is verified, so an
// abandoned sign-up never reserves the (globally unique) email address.
type PendingSignup struct {
	TenantID      string
	TransactionID string // authorization transaction that started the sign-up; the code is only valid there
	Email         string // normalized (lowercase, trimmed)
	DisplayName   string
	PasswordHash  string
	CodeHash      string // SHA-256 hex of the emailed code; the plaintext code is never stored
}

// PendingSignupStore keeps pending sign-ups in short-lived storage (Redis).
type PendingSignupStore interface {
	// AllowSend reports whether another code may be emailed to this address now,
	// enforcing a per-email cooldown and an hourly send cap.
	AllowSend(ctx context.Context, tenantID, email string, cooldown time.Duration, maxPerHour int) (bool, error)

	// Save stores (or replaces) a pending sign-up and resets its attempt counter.
	Save(ctx context.Context, signup *PendingSignup, ttl time.Duration) error

	// Get returns the pending sign-up, or nil, nil if none exists.
	Get(ctx context.Context, tenantID, email string) (*PendingSignup, error)

	// Verify atomically checks codeHash. On a match the record is deleted and
	// returned. On a mismatch the attempt counter is incremented and the record
	// is deleted once maxAttempts is reached. Returns nil, nil when the code is
	// invalid or no record exists.
	Verify(ctx context.Context, tenantID, email, codeHash string, maxAttempts int) (*PendingSignup, error)

	// Delete removes a pending sign-up.
	Delete(ctx context.Context, tenantID, email string) error
}
