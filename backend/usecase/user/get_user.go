package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// GetUserRequest represents a request to get a single user
type GetUserRequest struct {
	UserID   string
	TenantID string
}

// GetUserResponse contains the user data and role assignments
type GetUserResponse struct {
	ID            string
	Email         string
	DisplayName   *string
	Status        entities.UserStatus
	LastLoginAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	RolesByClient map[string][]string
}

// GetUser handles retrieving a single user with roles (T041)
type GetUser struct {
	userRepo repositories.EndUserRepository
	roleRepo repositories.RoleRepository
}

// NewGetUser creates a new GetUser use case
func NewGetUser(
	userRepo repositories.EndUserRepository,
	roleRepo repositories.RoleRepository,
) *GetUser {
	return &GetUser{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// Execute retrieves a user by ID with role assignments
func (uc *GetUser) Execute(ctx context.Context, req GetUserRequest) (*GetUserResponse, error) {
	if req.UserID == "" {
		return nil, errors.New("user_id is required")
	}
	if req.TenantID == "" {
		return nil, errors.New("tenant_id is required")
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, errors.New("invalid user_id format")
	}

	tenantUUID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, errors.New("invalid tenant_id format")
	}

	user, err := uc.userRepo.GetByID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if user.TenantID != tenantUUID {
		return nil, errors.New("tenant mismatch")
	}

	roleAssignments, err := uc.roleRepo.ListByUser(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user roles: %w", err)
	}

	rolesByClient := make(map[string][]string)
	for _, ra := range roleAssignments {
		if ra.IsActive {
			rolesByClient[ra.ClientID] = append(rolesByClient[ra.ClientID], ra.Role)
		}
	}

	return &GetUserResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		Status:        user.Status,
		LastLoginAt:   user.LastLoginAt,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		RolesByClient: rolesByClient,
	}, nil
}
