package user

import (
"context"
"errors"
"fmt"

"github.com/ranggaaprilio/keyles/domain/entities"
"github.com/ranggaaprilio/keyles/domain/repositories"
)

// GetUserOutput represents the detailed user response
type GetUserOutput struct {
	User            *entities.User
	RoleAssignments []*entities.UserRoleAssignment
}

// GetUser handles fetching a single user with their role assignments
type GetUser struct {
	endUserRepo repositories.EndUserRepository
	roleRepo    repositories.RoleRepository
}

// NewGetUser creates a new GetUser use case
func NewGetUser(endUserRepo repositories.EndUserRepository, roleRepo repositories.RoleRepository) *GetUser {
	return &GetUser{
		endUserRepo: endUserRepo,
		roleRepo:    roleRepo,
	}
}

// Execute retrieves a user by ID, scoped to the caller's tenant
func (uc *GetUser) Execute(ctx context.Context, userID, tenantID string) (*GetUserOutput, error) {
if userID == "" {
return nil, errors.New("user_id is required")
}
if tenantID == "" {
return nil, errors.New("tenant_id is required")
}

user, err := uc.endUserRepo.GetByID(ctx, userID)
if err != nil {
return nil, fmt.Errorf("user not found: %w", err)
}

// Tenant isolation
if user.TenantID != tenantID {
return nil, errors.New("user not found")
}

// Fetch role assignments
roles, err := uc.roleRepo.ListByUser(ctx, userID)
if err != nil {
// Non-fatal: return user without roles
roles = []*entities.UserRoleAssignment{}
}

return &GetUserOutput{
User:            user,
RoleAssignments: roles,
}, nil
}
