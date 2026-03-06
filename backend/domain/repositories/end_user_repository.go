package repositories

import (
	"context"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
)

// EndUserRepository defines the interface for end-user persistence operations.
// This is separate from UserRepository (which serves AdminUser for admin authentication)
// to prevent regression in the admin authentication flow while cleanly separating concerns.
type EndUserRepository interface {
	// GetByID retrieves an end-user by their unique ID
	GetByID(ctx context.Context, userID string) (*entities.User, error)

	// GetByEmail retrieves an end-user by email within a specific tenant
	GetByEmail(ctx context.Context, tenantID, email string) (*entities.User, error)

	// Create inserts a new end-user record
	Create(ctx context.Context, user *entities.User) error

	// Update modifies an existing end-user record (display name, password hash, etc.)
	Update(ctx context.Context, user *entities.User) error

	// ListByTenant returns paginated users filtered by optional status and search term.
	// search: case-insensitive partial match on display_name or email (empty = no filter)
	// status: filter by account status (empty = all statuses)
	// Returns: users, total count across all pages, error
	ListByTenant(ctx context.Context, tenantID, search string, status entities.UserStatus, page, pageSize int) ([]*entities.User, int, error)

	// CountByTenant returns the total number of users (all statuses) in a tenant
	CountByTenant(ctx context.Context, tenantID string) (int, error)

	// UpdateStatus updates the account status of a user
	UpdateStatus(ctx context.Context, userID string, status entities.UserStatus) error

	// UpdateLastLogin records the most recent successful authentication timestamp
	UpdateLastLogin(ctx context.Context, userID string, at time.Time) error

	// Delete permanently removes the user record
	Delete(ctx context.Context, userID string) error
}
