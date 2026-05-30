package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/mock"
)

// MockEndUserRepository is a mock implementation of EndUserRepository
type MockEndUserRepository struct {
	mock.Mock
}

// GetByID retrieves an end user by their unique ID
func (m *MockEndUserRepository) GetByID(ctx context.Context, userID uuid.UUID) (*entities.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

// GetByEmail retrieves an end user by email within a tenant
func (m *MockEndUserRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entities.User, error) {
	args := m.Called(ctx, tenantID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

// Create creates a new end user
func (m *MockEndUserRepository) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// Update updates an existing end user
func (m *MockEndUserRepository) Update(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// ListByTenant retrieves paginated end users for a tenant with optional filtering
func (m *MockEndUserRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, search string, status entities.UserStatus, page, pageSize int) ([]*entities.User, int, error) {
	args := m.Called(ctx, tenantID, search, status, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*entities.User), args.Int(1), args.Error(2)
}

// CountByTenant returns the total number of end users in a tenant
func (m *MockEndUserRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	args := m.Called(ctx, tenantID)
	return args.Int(0), args.Error(1)
}

// UpdateStatus updates the status of an end user
func (m *MockEndUserRepository) UpdateStatus(ctx context.Context, userID uuid.UUID, status entities.UserStatus) error {
	args := m.Called(ctx, userID, status)
	return args.Error(0)
}

// UpdateLastLogin updates the last login timestamp for an end user
func (m *MockEndUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, at time.Time) error {
	args := m.Called(ctx, userID, at)
	return args.Error(0)
}

// Delete removes an end user
func (m *MockEndUserRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
