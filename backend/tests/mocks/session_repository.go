package mocks

import (
	"context"
	"time"

	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/stretchr/testify/mock"
)

// MockSessionRepository is a mock implementation of SessionRepository
type MockSessionRepository struct {
	mock.Mock
}

// Create stores a new session with automatic expiration
func (m *MockSessionRepository) Create(ctx context.Context, session *repositories.Session, ttl time.Duration) error {
	args := m.Called(ctx, session, ttl)
	return args.Error(0)
}

// Get retrieves a session by session ID
func (m *MockSessionRepository) Get(ctx context.Context, sessionID string) (*repositories.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.Session), args.Error(1)
}

// Delete removes a session (logout)
func (m *MockSessionRepository) Delete(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

// Exists checks if a session exists and is valid
func (m *MockSessionRepository) Exists(ctx context.Context, sessionID string) (bool, error) {
	args := m.Called(ctx, sessionID)
	return args.Bool(0), args.Error(1)
}

// Extend extends the TTL of an existing session
func (m *MockSessionRepository) Extend(ctx context.Context, sessionID string, ttl time.Duration) error {
	args := m.Called(ctx, sessionID, ttl)
	return args.Error(0)
}
