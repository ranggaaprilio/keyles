package mocks

import (
	"context"
	"time"

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

// MockUserBlacklist mocks the Redis user blacklist
type MockUserBlacklist struct {
	mock.Mock
}

func (m *MockUserBlacklist) Add(ctx context.Context, userID string, ttl time.Duration) error {
	args := m.Called(ctx, userID, ttl)
	return args.Error(0)
}

func (m *MockUserBlacklist) IsBlacklisted(ctx context.Context, userID string) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

// MockUserCountCache mocks the Redis user count cache
type MockUserCountCache struct {
	mock.Mock
}

func (m *MockUserCountCache) Get(ctx context.Context, tenantID string) (int, bool, error) {
	args := m.Called(ctx, tenantID)
	return args.Int(0), args.Bool(1), args.Error(2)
}

func (m *MockUserCountCache) Set(ctx context.Context, tenantID string, count int) error {
	args := m.Called(ctx, tenantID, count)
	return args.Error(0)
}

func (m *MockUserCountCache) Invalidate(ctx context.Context, tenantID string) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}
