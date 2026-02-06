package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/mock"
)

// MockAuditRepository is a mock implementation of AuditRepository
type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) Create(ctx context.Context, log *entities.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAuditRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entities.AuditLog, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) FindByEventType(ctx context.Context, eventType entities.EventType, limit, offset int) ([]*entities.AuditLog, error) {
	args := m.Called(ctx, eventType, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) FindRecent(ctx context.Context, limit int) ([]*entities.AuditLog, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.AuditLog), args.Error(1)
}
