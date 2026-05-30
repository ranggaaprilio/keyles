package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
)

// DeleteUserRequest represents a request to delete a user account
type DeleteUserRequest struct {
	TargetUserID string
	AdminUserID  string
	TenantID     string
}

// DeleteUser handles hard-deleting a user account (US6)
type DeleteUser struct {
	userRepo         repositories.EndUserRepository
	roleRepo         repositories.RoleRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	blacklist        services.UserBlacklist
	eventRepo        repositories.UserEventRepository
}

// NewDeleteUser creates a new DeleteUser use case
func NewDeleteUser(
	userRepo repositories.EndUserRepository,
	roleRepo repositories.RoleRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
	blacklist services.UserBlacklist,
	eventRepo repositories.UserEventRepository,
) *DeleteUser {
	return &DeleteUser{
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		refreshTokenRepo: refreshTokenRepo,
		blacklist:        blacklist,
		eventRepo:        eventRepo,
	}
}

// Execute hard-deletes a user account after revoking roles and sessions
func (uc *DeleteUser) Execute(ctx context.Context, req DeleteUserRequest) error {
	if req.TargetUserID == req.AdminUserID {
		return errors.New("cannot delete your own account")
	}

	targetUUID, err := uuid.Parse(req.TargetUserID)
	if err != nil {
		return errors.New("invalid target_user_id format")
	}

	tenantUUID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return errors.New("invalid tenant_id format")
	}

	user, err := uc.userRepo.GetByID(ctx, targetUUID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	if user.TenantID != tenantUUID {
		return errors.New("tenant mismatch")
	}

	if err := uc.roleRepo.RevokeAllForUser(ctx, req.TargetUserID, req.AdminUserID); err != nil {
		return fmt.Errorf("failed to revoke roles: %w", err)
	}

	if err := uc.refreshTokenRepo.RevokeByUserID(ctx, req.TargetUserID); err != nil {
		return fmt.Errorf("failed to revoke sessions: %w", err)
	}

	if err := uc.blacklist.Add(ctx, req.TargetUserID, 900*time.Second); err != nil {
		return fmt.Errorf("failed to blacklist user: %w", err)
	}

	if err := uc.userRepo.Delete(ctx, targetUUID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	event := &entities.UserEvent{
		TenantID:  req.TenantID,
		UserID:    req.TargetUserID,
		EventType: entities.EventTypeUserDeleted,
		Details: map[string]any{
			"deleted_by": req.AdminUserID,
			"email":      user.Email,
			"user_id":    req.TargetUserID,
		},
		OccurredAt: time.Now(),
	}
	if err := uc.eventRepo.Record(ctx, event); err != nil {
		return fmt.Errorf("failed to record user event: %w", err)
	}

	return nil
}
