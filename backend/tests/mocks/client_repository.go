package mocks

import (
	"context"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/mock"
)

// MockClientRepository is a mock implementation of ClientRepository
type MockClientRepository struct {
	mock.Mock
}

// Create creates a new OAuth client
func (m *MockClientRepository) Create(ctx context.Context, client *entities.Client) error {
	args := m.Called(ctx, client)
	return args.Error(0)
}

// GetByID retrieves a client by database ID
func (m *MockClientRepository) GetByID(ctx context.Context, id string) (*entities.Client, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Client), args.Error(1)
}

// GetByClientID retrieves a client by client_id and tenant_id
func (m *MockClientRepository) GetByClientID(ctx context.Context, clientID string, tenantID string) (*entities.Client, error) {
	args := m.Called(ctx, clientID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Client), args.Error(1)
}

// Update updates an existing client
func (m *MockClientRepository) Update(ctx context.Context, client *entities.Client) error {
	args := m.Called(ctx, client)
	return args.Error(0)
}

// Delete soft-deletes a client
func (m *MockClientRepository) Delete(ctx context.Context, clientID string) error {
	args := m.Called(ctx, clientID)
	return args.Error(0)
}

// ListByTenant retrieves all clients for a tenant
func (m *MockClientRepository) ListByTenant(ctx context.Context, tenantID string) ([]*entities.Client, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Client), args.Error(1)
}

// ValidateCredentials validates client credentials
func (m *MockClientRepository) ValidateCredentials(ctx context.Context, clientID string, clientSecret string) (*entities.Client, error) {
	args := m.Called(ctx, clientID, clientSecret)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Client), args.Error(1)
}
