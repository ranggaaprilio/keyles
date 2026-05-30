package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockUserBlacklist is a mock implementation of UserBlacklist
type MockUserBlacklist struct {
	mock.Mock
}

// Add blacklists a user for the specified TTL duration
func (m *MockUserBlacklist) Add(ctx context.Context, userID string, ttl time.Duration) error {
	args := m.Called(ctx, userID, ttl)
	return args.Error(0)
}

// IsBlacklisted checks if a user is currently blacklisted
func (m *MockUserBlacklist) IsBlacklisted(ctx context.Context, userID string) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}
