package repositories

import (
	"context"

	"github.com/ranggaaprilio/keyles/domain/entities"
)

// RoleRepository defines the interface for user role operations
type RoleRepository interface {
	// AssignRole assigns a role to a user for a client
	AssignRole(ctx context.Context, assignment *entities.UserRoleAssignment) error

	// RevokeRole removes a role assignment
	RevokeRole(ctx context.Context, userID, clientID, role string) error

	// GetUserRoles retrieves all roles for a user in a client
	GetUserRoles(ctx context.Context, userID, clientID string) ([]*entities.UserRoleAssignment, error)

	// HasRole checks if a user has a specific role for a client
	HasRole(ctx context.Context, userID, clientID, role string) (bool, error)

	// ListRolesByClient retrieves all role assignments for a client
	ListRolesByClient(ctx context.Context, clientID string) ([]*entities.UserRoleAssignment, error)

	// ListRolesByUser retrieves all role assignments for a user across all clients
	ListRolesByUser(ctx context.Context, userID string) ([]*entities.UserRoleAssignment, error)
}
