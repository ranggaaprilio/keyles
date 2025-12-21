package mocks

import (
	"context"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/mock"
)

// MockAuditRepository is a mock implementation of AuditRepository
type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) Create(ctx context.Context, log *entities.AuditLog) error {
	args := m.Called(log)
	return args.Error(0)
}

func (m *MockAuditRepository) FindByTenantID(ctx context.Context, tenantID string, limit int) ([]*entities.AuditLog, error) {
	args := m.Called(ctx, tenantID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) FindByEventType(ctx context.Context, eventType string, limit int) ([]*entities.AuditLog, error) {
	args := m.Called(ctx, eventType, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.AuditLog), args.Error(1)
}
