package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
)

// UserRepository defines the interface for user persistence operations
type UserRepository interface {
	// Create creates a new user in the database
	Create(ctx context.Context, user *entities.AdminUser) error

	// FindByID retrieves a user by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.AdminUser, error)

	// FindByEmail retrieves a user by email address (case-insensitive)
	FindByEmail(ctx context.Context, email string) (*entities.AdminUser, error)

	// FindByTenantID retrieves all users for a tenant
	FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entities.AdminUser, error)

	// Update updates an existing user
	Update(ctx context.Context, user *entities.AdminUser) error

	// EmailExists checks if an email address already exists
	EmailExists(ctx context.Context, email string) (bool, error)

	// Delete deletes a user
	Delete(ctx context.Context, id uuid.UUID) error
}
