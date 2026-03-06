package services

import "context"

// UserCountCache abstracts the Redis-backed tenant user count cache behind a
// domain interface for Dependency Inversion Principle (DIP) compliance.
//
// The cache avoids frequent COUNT queries against the users table when enforcing
// the 10,000-user-per-tenant quota during invitation.
//
// Redis key pattern: user_count:{tenant_id}
// TTL: 60 seconds
type UserCountCache interface {
	// Get retrieves the cached user count for a tenant.
	// Returns count, cache-hit boolean, and error.
	// On cache miss, returns 0, false, nil.
	Get(ctx context.Context, tenantID string) (int, bool, error)

	// Set stores the user count for a tenant in cache
	Set(ctx context.Context, tenantID string, count int) error

	// Invalidate removes the cached user count for a tenant
	Invalidate(ctx context.Context, tenantID string) error
}
