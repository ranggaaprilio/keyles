package mocks

import (
	"context"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/mock"
)

// MockTenantRepository is a mock implementation of TenantRepository
type MockTenantRepository struct {
	mock.Mock
}

func (m *MockTenantRepository) Create(ctx context.Context, tenant *entities.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) FindByID(id string) (*entities.Tenant, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

func (m *MockTenantRepository) FindByOrganizationName(ctx context.Context, name string) (*entities.Tenant, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

func (m *MockTenantRepository) Update(tenant *entities.Tenant) error {
	args := m.Called(tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTenantRepository) List(ctx context.Context, offset, limit int) ([]*entities.Tenant, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Tenant), args.Error(1)
}
