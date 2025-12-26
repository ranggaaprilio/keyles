package repositories

import (
	"context"

	"github.com/ranggaaprilio/keyles/domain/entities"
)

// ClientRepository defines the interface for client data access
type ClientRepository interface {
	// Create creates a new OAuth client
	Create(ctx context.Context, client *entities.Client) error

	// GetByID retrieves a client by database ID (for internal use)
	GetByID(ctx context.Context, id string) (*entities.Client, error)

	// GetByClientID retrieves a client by client_id and tenant_id
	GetByClientID(ctx context.Context, clientID string, tenantID string) (*entities.Client, error)

	// Update updates an existing client
	Update(ctx context.Context, client *entities.Client) error

	// Delete soft-deletes a client (sets is_active = false)
	Delete(ctx context.Context, clientID string) error

	// ListByTenant retrieves all clients for a tenant
	ListByTenant(ctx context.Context, tenantID string) ([]*entities.Client, error)

	// ValidateCredentials validates client credentials (client_id + client_secret)
	// Returns the client if valid, error if invalid
	ValidateCredentials(ctx context.Context, clientID string, clientSecret string) (*entities.Client, error)
}
