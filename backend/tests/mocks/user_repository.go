package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository interface
type MockUserRepository struct {
	mock.Mock
}

// Create creates a new user in the database
func (m *MockUserRepository) Create(ctx context.Context, user *entities.AdminUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// FindByID retrieves a user by ID
func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.AdminUser, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.AdminUser), args.Error(1)
}

// FindByEmail retrieves a user by email address
func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.AdminUser, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.AdminUser), args.Error(1)
}

// FindByTenantID retrieves all users for a tenant
func (m *MockUserRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entities.AdminUser, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.AdminUser), args.Error(1)
}

// Update updates an existing user
func (m *MockUserRepository) Update(ctx context.Context, user *entities.AdminUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// EmailExists checks if an email address already exists
func (m *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

// Delete deletes a user
func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
