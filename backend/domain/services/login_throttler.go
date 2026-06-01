package services

import "context"

// LoginThrottler defines the domain abstraction for OAuth end-user login
// rate limiting. It manages fixed-window counters for both source IP address
// and tenant-scoped normalized email. A login attempt is rejected when either
// counter reaches the configured maximum.
//
// The TTL is set only when a counter is first created; subsequent failures
// increment the counter without extending the window.
type LoginThrottler interface {
	// IsThrottled returns true when either the source-IP counter or the
	// tenant-email counter has reached or exceeded the configured maximum.
	// Callers MUST check before verifying credentials.
	IsThrottled(ctx context.Context, sourceIP, tenantID, normalizedEmail string) (bool, error)

	// RecordFailure atomically increments both the source-IP counter and the
	// tenant-email counter. It creates each key with the configured window TTL
	// if the key does not already exist; later increments do not extend the TTL.
	RecordFailure(ctx context.Context, sourceIP, tenantID, normalizedEmail string) error

	// ClearEmailBucket removes the tenant-email counter after a successful
	// login so the user is not penalized by their own past failures when
	// they next authenticate.
	ClearEmailBucket(ctx context.Context, tenantID, normalizedEmail string) error
}