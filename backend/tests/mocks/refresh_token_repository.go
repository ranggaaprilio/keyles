package mocks

import (
	"context"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/mock"
)

// MockRefreshTokenRepository is a mock implementation of RefreshTokenRepository
type MockRefreshTokenRepository struct {
	mock.Mock
}

// Create stores a new refresh token
func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// GetByToken retrieves a refresh token by its token value (hashed)
func (m *MockRefreshTokenRepository) GetByToken(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.RefreshToken), args.Error(1)
}

// Revoke marks a refresh token as revoked
func (m *MockRefreshTokenRepository) Revoke(ctx context.Context, tokenHash string, revokedBy string) error {
	args := m.Called(ctx, tokenHash, revokedBy)
	return args.Error(0)
}

// RevokeAllForUser revokes all refresh tokens for a user-client combination
func (m *MockRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string, clientID string) error {
	args := m.Called(ctx, userID, clientID)
	return args.Error(0)
}

// DeleteExpired removes expired tokens from the database
func (m *MockRefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// IsRevoked checks if a token is revoked
func (m *MockRefreshTokenRepository) IsRevoked(ctx context.Context, tokenHash string) (bool, error) {
	args := m.Called(ctx, tokenHash)
	return args.Bool(0), args.Error(1)
}

// UpdateLastUsed updates the last_used_at timestamp
func (m *MockRefreshTokenRepository) UpdateLastUsed(ctx context.Context, tokenHash string) error {
	args := m.Called(ctx, tokenHash)
	return args.Error(0)
}
