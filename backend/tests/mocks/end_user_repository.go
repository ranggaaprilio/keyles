package mocks

import (
	"context"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/mock"
)

// MockEndUserRepository is a mock implementation of EndUserRepository
type MockEndUserRepository struct {
	mock.Mock
}

func (m *MockEndUserRepository) GetByID(ctx context.Context, userID string) (*entities.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockEndUserRepository) GetByEmail(ctx context.Context, tenantID, email string) (*entities.User, error) {
	args := m.Called(ctx, tenantID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockEndUserRepository) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockEndUserRepository) Update(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockEndUserRepository) ListByTenant(ctx context.Context, tenantID, search string, status entities.UserStatus, page, pageSize int) ([]*entities.User, int, error) {
	args := m.Called(ctx, tenantID, search, status, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*entities.User), args.Int(1), args.Error(2)
}

func (m *MockEndUserRepository) CountByTenant(ctx context.Context, tenantID string) (int, error) {
	args := m.Called(ctx, tenantID)
	return args.Int(0), args.Error(1)
}

func (m *MockEndUserRepository) UpdateStatus(ctx context.Context, userID string, status entities.UserStatus) error {
	args := m.Called(ctx, userID, status)
	return args.Error(0)
}

func (m *MockEndUserRepository) UpdateLastLogin(ctx context.Context, userID string, at time.Time) error {
	args := m.Called(ctx, userID, at)
	return args.Error(0)
}

func (m *MockEndUserRepository) Delete(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
