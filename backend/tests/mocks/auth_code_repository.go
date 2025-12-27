package mocks

import (
	"context"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/mock"
)

// MockAuthCodeRepository is a mock implementation of AuthCodeRepository
type MockAuthCodeRepository struct {
	mock.Mock
}

// Store saves an authorization code with automatic expiration
func (m *MockAuthCodeRepository) Store(ctx context.Context, code *entities.AuthorizationCode, ttl time.Duration) error {
	args := m.Called(ctx, code, ttl)
	return args.Error(0)
}

// Get retrieves an authorization code
func (m *MockAuthCodeRepository) Get(ctx context.Context, code string) (*entities.AuthorizationCode, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.AuthorizationCode), args.Error(1)
}

// MarkAsUsed marks an authorization code as used
func (m *MockAuthCodeRepository) MarkAsUsed(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

// Delete removes an authorization code
func (m *MockAuthCodeRepository) Delete(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

// Exists checks if an authorization code exists
func (m *MockAuthCodeRepository) Exists(ctx context.Context, code string) (bool, error) {
	args := m.Called(ctx, code)
	return args.Bool(0), args.Error(1)
}
