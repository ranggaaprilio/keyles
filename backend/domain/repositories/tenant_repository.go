package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
)

// TenantRepository defines the interface for tenant persistence operations
type TenantRepository interface {
	// Create creates a new tenant in the database
	Create(ctx context.Context, tenant *entities.Tenant) error

	// FindByID retrieves a tenant by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Tenant, error)

	// FindByOrganizationName retrieves a tenant by organization name (case-insensitive)
	FindByOrganizationName(ctx context.Context, name string) (*entities.Tenant, error)

	// Update updates an existing tenant
	Update(ctx context.Context, tenant *entities.Tenant) error

	// OrganizationNameExists checks if an organization name already exists
	OrganizationNameExists(ctx context.Context, name string) (bool, error)

	// Delete soft deletes a tenant (sets status to deleted)
	Delete(ctx context.Context, id uuid.UUID) error
}
