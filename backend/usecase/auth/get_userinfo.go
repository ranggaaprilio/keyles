package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// UserInfoClaims represents the OIDC UserInfo response (FR-052)
type UserInfoClaims struct {
	Sub           string   `json:"sub"`             // Subject - User ID
	Email         string   `json:"email"`           // User email
	EmailVerified bool     `json:"email_verified"`  // Email verification status
	TenantID      string   `json:"tenant_id"`       // Tenant ID (custom claim)
	Roles         []string `json:"roles,omitempty"` // Active roles for the client (feature 005)
}

// GetUserInfo use case extracts user profile from access token
type GetUserInfo struct {
	userRepo interface{}
	roleRepo repositories.RoleRepository
}

// NewGetUserInfo creates a new GetUserInfo use case
func NewGetUserInfo(
	userRepo interface{},
	roleRepo repositories.RoleRepository,
) *GetUserInfo {
	return &GetUserInfo{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// Execute retrieves user information for the UserInfo endpoint
// The user ID and client ID should be extracted from the validated access token by the handler
func (uc *GetUserInfo) Execute(ctx context.Context, userID string, clientID string) (*UserInfoClaims, error) {
	// Validate input
	if userID == "" {
		return nil, errors.New("user_id is required")
	}

	// Retrieve user from repository
	user, err := uc.getUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Fetch active roles for the user-client pair
	roles, _ := uc.roleRepo.GetActiveRoles(ctx, userID, clientID)
	if roles == nil {
		roles = []string{}
	}

	// Map to UserInfo claims
	claims := &UserInfoClaims{
		Sub:           user.ID,
		Email:         user.Email,
		EmailVerified: true, // AdminUsers are verified by default
		TenantID:      user.TenantID,
		Roles:         roles,
	}

	return claims, nil
}

func (uc *GetUserInfo) getUser(ctx context.Context, userID string) (*entities.User, error) {
	switch repo := uc.userRepo.(type) {
	case repositories.EndUserRepository:
		return repo.GetByID(ctx, userID)
	case repositories.UserRepository:
		id, err := uuid.Parse(userID)
		if err != nil {
			return nil, errors.New("invalid user_id format")
		}
		user, err := repo.FindByID(ctx, id)
		if err != nil || user == nil {
			return nil, err
		}
		return &entities.User{ID: user.ID.String(), TenantID: user.TenantID.String(), Email: user.Email}, nil
	default:
		return nil, errors.New("unsupported user repository")
	}
}
