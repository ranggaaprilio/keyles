package mocks

import (
	"context"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/mock"
)

// MockUserEventRepository is a mock implementation of UserEventRepository
type MockUserEventRepository struct {
	mock.Mock
}

// Record stores a new user event
func (m *MockUserEventRepository) Record(ctx context.Context, event *entities.UserEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// ListByUser retrieves paginated events for a specific user
func (m *MockUserEventRepository) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*entities.UserEvent, int, error) {
	args := m.Called(ctx, userID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*entities.UserEvent), args.Int(1), args.Error(2)
}

// DeleteOlderThan removes events older than the specified time
func (m *MockUserEventRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}
