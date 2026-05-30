package mocks

import (
	"context"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/mock"
)

// MockRoleRepository is a mock implementation of RoleRepository
type MockRoleRepository struct {
	mock.Mock
}

// AssignRole assigns a role to a user for a client
func (m *MockRoleRepository) AssignRole(ctx context.Context, assignment *entities.UserRoleAssignment) error {
	args := m.Called(ctx, assignment)
	return args.Error(0)
}

// RevokeRole removes a role assignment
func (m *MockRoleRepository) RevokeRole(ctx context.Context, userID, clientID, role string) error {
	args := m.Called(ctx, userID, clientID, role)
	return args.Error(0)
}

// GetUserRoles retrieves all roles for a user in a client
func (m *MockRoleRepository) GetUserRoles(ctx context.Context, userID, clientID string) ([]*entities.UserRoleAssignment, error) {
	args := m.Called(ctx, userID, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserRoleAssignment), args.Error(1)
}

// HasRole checks if a user has a specific role for a client
func (m *MockRoleRepository) HasRole(ctx context.Context, userID, clientID, role string) (bool, error) {
	args := m.Called(ctx, userID, clientID, role)
	return args.Bool(0), args.Error(1)
}

// ListRolesByClient retrieves all role assignments for a client
func (m *MockRoleRepository) ListRolesByClient(ctx context.Context, clientID string) ([]*entities.UserRoleAssignment, error) {
	args := m.Called(ctx, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserRoleAssignment), args.Error(1)
}

// ListRolesByUser retrieves all role assignments for a user across all clients
func (m *MockRoleRepository) ListRolesByUser(ctx context.Context, userID string) ([]*entities.UserRoleAssignment, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserRoleAssignment), args.Error(1)
}

// HasAnyRole checks if a user has any role for a client
func (m *MockRoleRepository) HasAnyRole(ctx context.Context, userID, clientID string) (bool, error) {
	args := m.Called(ctx, userID, clientID)
	return args.Bool(0), args.Error(1)
}

// Assign assigns a role to a user
func (m *MockRoleRepository) Assign(ctx context.Context, assignment *entities.UserRoleAssignment) error {
	args := m.Called(ctx, assignment)
	return args.Error(0)
}

// Revoke removes a specific role assignment
func (m *MockRoleRepository) Revoke(ctx context.Context, assignmentID int64, revokedByUserID string) error {
	args := m.Called(ctx, assignmentID, revokedByUserID)
	return args.Error(0)
}

// ListByUser retrieves all role assignments for a user
func (m *MockRoleRepository) ListByUser(ctx context.Context, userID string) ([]*entities.UserRoleAssignment, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserRoleAssignment), args.Error(1)
}

// ListByClient retrieves paginated role assignments for a client
func (m *MockRoleRepository) ListByClient(ctx context.Context, clientID string, page, pageSize int) ([]*entities.UserRoleAssignment, int, error) {
	args := m.Called(ctx, clientID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*entities.UserRoleAssignment), args.Int(1), args.Error(2)
}

// RevokeAllForUser revokes all role assignments for a user
func (m *MockRoleRepository) RevokeAllForUser(ctx context.Context, userID, revokedByUserID string) error {
	args := m.Called(ctx, userID, revokedByUserID)
	return args.Error(0)
}

// GetActiveRoles retrieves active role names for a user in a client
func (m *MockRoleRepository) GetActiveRoles(ctx context.Context, userID, clientID string) ([]string, error) {
	args := m.Called(ctx, userID, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
