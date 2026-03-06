package repositories

import (
	"context"

	"github.com/ranggaaprilio/keyles/domain/entities"
)

// RoleRepository defines the interface for user role operations
type RoleRepository interface {
	// AssignRole assigns a role to a user for a client (legacy — use Assign for new code)
	AssignRole(ctx context.Context, assignment *entities.UserRoleAssignment) error

	// RevokeRole removes a role assignment (legacy — use Revoke for new code)
	RevokeRole(ctx context.Context, userID, clientID, role string) error

	// GetUserRoles retrieves all roles for a user in a client
	GetUserRoles(ctx context.Context, userID, clientID string) ([]*entities.UserRoleAssignment, error)

	// HasRole checks if a user has a specific role for a client
	HasRole(ctx context.Context, userID, clientID, role string) (bool, error)

	// HasAnyRole checks if a user has any active role for a client
	HasAnyRole(ctx context.Context, userID, clientID string) (bool, error)

	// ListRolesByClient retrieves all role assignments for a client
	ListRolesByClient(ctx context.Context, clientID string) ([]*entities.UserRoleAssignment, error)

	// ListRolesByUser retrieves all role assignments for a user across all clients
	ListRolesByUser(ctx context.Context, userID string) ([]*entities.UserRoleAssignment, error)

	// --- New methods for feature 005 ---

	// Assign creates a new role assignment. Returns ErrDuplicateRole if already active.
	Assign(ctx context.Context, assignment *entities.UserRoleAssignment) error

	// Revoke soft-deletes a role assignment by ID, recording revokedBy and revokedAt.
	Revoke(ctx context.Context, assignmentID int64, revokedByUserID string) error

	// ListByUser returns all role assignments for a user, including inactive ones.
	ListByUser(ctx context.Context, userID string) ([]*entities.UserRoleAssignment, error)

	// ListByClient returns all active role assignments for a client application (admin view).
	ListByClient(ctx context.Context, clientID string, page, pageSize int) ([]*entities.UserRoleAssignment, int, error)

	// RevokeAllForUser revokes all active role assignments for a user (used on account disable/delete).
	RevokeAllForUser(ctx context.Context, userID, revokedByUserID string) error

	// GetActiveRoles returns only active role name strings for a user-client pair.
	// Used by issue_token.go and get_userinfo.go for JWT claim injection.
	GetActiveRoles(ctx context.Context, userID, clientID string) ([]string, error)
}

