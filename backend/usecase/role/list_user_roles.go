package role

import (
	"context"
	"errors"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// ListUserRolesRequest represents a request to list roles for a user
type ListUserRolesRequest struct {
	UserID   string
	ClientID string // Optional: filter by client
}

// ListUserRolesResponse represents the response containing user roles
type ListUserRolesResponse struct {
	Roles []*entities.UserRoleAssignment
}

// ListUserRoles retrieves all active roles for a user (FR-006b)
type ListUserRoles struct {
	roleRepo repositories.RoleRepository
}

// NewListUserRoles creates a new ListUserRoles use case
func NewListUserRoles(roleRepo repositories.RoleRepository) *ListUserRoles {
	return &ListUserRoles{
		roleRepo: roleRepo,
	}
}

// Execute retrieves all active roles for a user, optionally filtered by client
func (uc *ListUserRoles) Execute(ctx context.Context, req ListUserRolesRequest) (*ListUserRolesResponse, error) {
	// Validate required fields
	if req.UserID == "" {
		return nil, errors.New("user_id is required")
	}

	var roles []*entities.UserRoleAssignment
	var err error

	if req.ClientID != "" {
		// Filter by specific client
		roles, err = uc.roleRepo.GetUserRoles(ctx, req.UserID, req.ClientID)
	} else {
		// Get all roles across all clients
		roles, err = uc.roleRepo.ListRolesByUser(ctx, req.UserID)
	}

	if err != nil {
		return nil, errors.New("failed to retrieve roles: " + err.Error())
	}

	// Filter for only active roles
	activeRoles := make([]*entities.UserRoleAssignment, 0)
	for _, r := range roles {
		if r.IsActive {
			activeRoles = append(activeRoles, r)
		}
	}

	return &ListUserRolesResponse{
		Roles: activeRoles,
	}, nil
}
