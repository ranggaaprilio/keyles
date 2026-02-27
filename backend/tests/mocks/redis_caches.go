package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockClientCountCache mocks the Redis client count cache
type MockClientCountCache struct {
	mock.Mock
}

func (m *MockClientCountCache) Get(ctx context.Context, tenantID string) (int, error) {
	args := m.Called(ctx, tenantID)
	return args.Int(0), args.Error(1)
}

func (m *MockClientCountCache) Set(ctx context.Context, tenantID string, count int) error {
	args := m.Called(ctx, tenantID, count)
	return args.Error(0)
}

func (m *MockClientCountCache) Invalidate(ctx context.Context, tenantID string) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

// MockRevokedClientCache mocks the Redis revoked client cache
type MockRevokedClientCache struct {
	mock.Mock
}

func (m *MockRevokedClientCache) Revoke(ctx context.Context, clientID string) error {
	args := m.Called(ctx, clientID)
	return args.Error(0)
}

func (m *MockRevokedClientCache) IsRevoked(ctx context.Context, clientID string) (bool, error) {
	args := m.Called(ctx, clientID)
	return args.Bool(0), args.Error(1)
}
