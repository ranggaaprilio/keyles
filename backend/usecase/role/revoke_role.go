package role

import (
	"context"
	"errors"

	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// RevokeRoleRequest represents a request to revoke a role from a user
type RevokeRoleRequest struct {
	UserID   string
	ClientID string
	Role     string
}

// RevokeRole handles role revocation from users (FR-006b)
type RevokeRole struct {
	roleRepo         repositories.RoleRepository
	refreshTokenRepo repositories.RefreshTokenRepository
}

// NewRevokeRole creates a new RevokeRole use case
func NewRevokeRole(
	roleRepo repositories.RoleRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
) *RevokeRole {
	return &RevokeRole{
		roleRepo:         roleRepo,
		refreshTokenRepo: refreshTokenRepo,
	}
}

// Execute revokes a role from a user for a client
func (uc *RevokeRole) Execute(ctx context.Context, req RevokeRoleRequest) error {
	// Validate required fields
	if req.UserID == "" {
		return errors.New("user_id is required")
	}
	if req.ClientID == "" {
		return errors.New("client_id is required")
	}
	if req.Role == "" {
		return errors.New("role is required")
	}

	// Check if user has the role
	hasRole, err := uc.roleRepo.HasRole(ctx, req.UserID, req.ClientID, req.Role)
	if err != nil {
		return errors.New("failed to check role: " + err.Error())
	}
	if !hasRole {
		return errors.New("role not found: user does not have this role for the client")
	}

	// Revoke the role (FR-006b)
	if err := uc.roleRepo.RevokeRole(ctx, req.UserID, req.ClientID, req.Role); err != nil {
		return errors.New("failed to revoke role: " + err.Error())
	}

	// Cascade revocation of refresh tokens (FR-006e)
	// This ensures user's active sessions are terminated when their role is revoked
	// Best effort - don't fail if this errors
	_ = uc.refreshTokenRepo.RevokeAllForUser(ctx, req.UserID, req.ClientID)

	return nil
}
