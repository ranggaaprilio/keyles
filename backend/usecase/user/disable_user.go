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

// DisableUserRequest represents a request to disable a user account
type DisableUserRequest struct {
	TargetUserID string
	AdminUserID  string
	TenantID     string
}

// DisableUser handles disabling a user account (US5)
type DisableUser struct {
	userRepo         repositories.EndUserRepository
	roleRepo         repositories.RoleRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	blacklist        services.UserBlacklist
	eventRepo        repositories.UserEventRepository
}

// NewDisableUser creates a new DisableUser use case
func NewDisableUser(
	userRepo repositories.EndUserRepository,
	roleRepo repositories.RoleRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
	blacklist services.UserBlacklist,
	eventRepo repositories.UserEventRepository,
) *DisableUser {
	return &DisableUser{
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		refreshTokenRepo: refreshTokenRepo,
		blacklist:        blacklist,
		eventRepo:        eventRepo,
	}
}

// Execute disables a user account, revokes all sessions, and blacklists the user
func (uc *DisableUser) Execute(ctx context.Context, req DisableUserRequest) error {
	if req.TargetUserID == req.AdminUserID {
		return errors.New("cannot disable your own account")
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

	if err := uc.userRepo.UpdateStatus(ctx, targetUUID, entities.UserStatusDisabled); err != nil {
		return fmt.Errorf("failed to disable user: %w", err)
	}

	if err := uc.refreshTokenRepo.RevokeByUserID(ctx, req.TargetUserID); err != nil {
		return fmt.Errorf("failed to revoke sessions: %w", err)
	}

	if err := uc.blacklist.Add(ctx, req.TargetUserID, 900*time.Second); err != nil {
		return fmt.Errorf("failed to blacklist user: %w", err)
	}

	event := &entities.UserEvent{
		TenantID:   req.TenantID,
		UserID:     req.TargetUserID,
		EventType:  entities.EventTypeAccountDisabled,
		Details:    map[string]any{"disabled_by": req.AdminUserID},
		OccurredAt: time.Now(),
	}
	if err := uc.eventRepo.Record(ctx, event); err != nil {
		return fmt.Errorf("failed to record user event: %w", err)
	}

	return nil
}
