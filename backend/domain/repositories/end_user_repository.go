package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
)

// EndUserRepository defines the interface for end-user persistence operations
type EndUserRepository interface {
	// GetByID retrieves an end user by their unique ID
	GetByID(ctx context.Context, userID uuid.UUID) (*entities.User, error)

	// GetByEmail retrieves an end user by email within a tenant
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entities.User, error)

	// Create creates a new end user
	Create(ctx context.Context, user *entities.User) error

	// Update updates an existing end user
	Update(ctx context.Context, user *entities.User) error

	// ListByTenant retrieves paginated end users for a tenant with optional filtering
	ListByTenant(ctx context.Context, tenantID uuid.UUID, search string, status entities.UserStatus, page, pageSize int) ([]*entities.User, int, error)

	// CountByTenant returns the total number of end users in a tenant
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)

	// UpdateStatus updates the status of an end user
	UpdateStatus(ctx context.Context, userID uuid.UUID, status entities.UserStatus) error

	// UpdateLastLogin updates the last login timestamp for an end user
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, at time.Time) error

	// Delete removes an end user
	Delete(ctx context.Context, userID uuid.UUID) error
}
