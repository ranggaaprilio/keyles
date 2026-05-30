package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// UserInfoClaims represents the OIDC UserInfo response (FR-052)
type UserInfoClaims struct {
	Sub           string   `json:"sub"`            // Subject - User ID
	Email         string   `json:"email"`          // User email
	EmailVerified bool     `json:"email_verified"` // Email verification status
	TenantID      string   `json:"tenant_id"`      // Tenant ID (custom claim)
	Roles         []string `json:"roles,omitempty"`
}

// GetUserInfo use case extracts user profile from access token
type GetUserInfo struct {
	userRepo repositories.UserRepository
	roleRepo repositories.RoleRepository
}

// NewGetUserInfo creates a new GetUserInfo use case
func NewGetUserInfo(userRepo repositories.UserRepository, roleRepo repositories.RoleRepository) *GetUserInfo {
	return &GetUserInfo{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// Execute retrieves user information for the UserInfo endpoint
// The user ID should be extracted from the validated access token by the handler
func (uc *GetUserInfo) Execute(ctx context.Context, userID, clientID string) (*UserInfoClaims, error) {
	// Validate input
	if userID == "" {
		return nil, errors.New("user_id is required")
	}

	// Parse user ID as UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user_id format")
	}

	// Retrieve user from repository
	user, err := uc.userRepo.FindByID(ctx, userUUID)
	if err != nil {
		return nil, errors.New("user not found: " + err.Error())
	}

	roles, err := uc.roleRepo.GetActiveRoles(ctx, userID, clientID)
	if err != nil {
		return nil, errors.New("failed to retrieve user roles: " + err.Error())
	}
	if roles == nil {
		roles = []string{}
	}

	// Map to UserInfo claims
	claims := &UserInfoClaims{
		Sub:           user.ID.String(),
		Email:         user.Email,
		EmailVerified: true, // AdminUsers are verified by default
		TenantID:      user.TenantID.String(),
		Roles:         roles,
	}

	return claims, nil
}
