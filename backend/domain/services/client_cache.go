package services

import "context"

// ClientCountCache provides caching for tenant client counts
type ClientCountCache interface {
	// Get returns cached client count for a tenant (-1 if not cached)
	Get(ctx context.Context, tenantID string) (int, error)
	// Set stores the client count for a tenant
	Set(ctx context.Context, tenantID string, count int) error
	// Invalidate removes the cached count for a tenant
	Invalidate(ctx context.Context, tenantID string) error
}

// RevokedClientCache provides caching for revoked client IDs
type RevokedClientCache interface {
	// Revoke marks a client as revoked in the cache
	Revoke(ctx context.Context, clientID string) error
	// IsRevoked checks if a client has been revoked
	IsRevoked(ctx context.Context, clientID string) (bool, error)
}
