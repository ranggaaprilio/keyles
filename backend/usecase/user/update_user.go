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

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	UserID      string
	TenantID    string
	DisplayName string
}

// UpdateUserResponse contains the updated user data
type UpdateUserResponse struct {
	ID          string
	Email       string
	DisplayName *string
	Status      entities.UserStatus
	LastLoginAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UpdateUser handles updating a user's display name (T042)
type UpdateUser struct {
	userRepo repositories.EndUserRepository
}

// NewUpdateUser creates a new UpdateUser use case
func NewUpdateUser(userRepo repositories.EndUserRepository) *UpdateUser {
	return &UpdateUser{
		userRepo: userRepo,
	}
}

// Execute updates a user's display name
func (uc *UpdateUser) Execute(ctx context.Context, req UpdateUserRequest) (*UpdateUserResponse, error) {
	if req.UserID == "" {
		return nil, errors.New("user_id is required")
	}
	if req.TenantID == "" {
		return nil, errors.New("tenant_id is required")
	}

	if req.DisplayName != "" {
		if err := entities.ValidateFullName(req.DisplayName); err != nil {
			return nil, fmt.Errorf("invalid display_name: %w", err)
		}
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

	if req.DisplayName != "" {
		dn := req.DisplayName
		user.DisplayName = &dn
	} else {
		user.DisplayName = nil
	}
	user.UpdatedAt = time.Now()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &UpdateUserResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Status:      user.Status,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}
