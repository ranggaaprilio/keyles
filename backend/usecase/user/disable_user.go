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

var (
ErrCannotDisableSelf  = errors.New("cannot disable your own account")
ErrCannotDisableAdmin = errors.New("cannot disable an administrator account")
)

// DisableUserInput represents the request to disable a user
type DisableUserInput struct {
	TargetUserID string
	AdminUserID  string
	TenantID     string
}

// DisableUser handles disabling a user account
type DisableUser struct {
	endUserRepo      repositories.EndUserRepository
	userRepo         repositories.UserRepository // for admin user check
	refreshTokenRepo repositories.RefreshTokenRepository
	userBlacklist    services.UserBlacklist
	eventRepo        repositories.UserEventRepository
	auditRepo        repositories.AuditRepository
}

// NewDisableUser creates a new DisableUser use case
func NewDisableUser(
endUserRepo repositories.EndUserRepository,
userRepo repositories.UserRepository,
refreshTokenRepo repositories.RefreshTokenRepository,
userBlacklist services.UserBlacklist,
eventRepo repositories.UserEventRepository,
auditRepo repositories.AuditRepository,
) *DisableUser {
	return &DisableUser{
		endUserRepo:      endUserRepo,
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		userBlacklist:    userBlacklist,
		eventRepo:        eventRepo,
		auditRepo:        auditRepo,
	}
}

// Execute disables a user account
func (uc *DisableUser) Execute(ctx context.Context, input DisableUserInput) error {
	if input.TargetUserID == "" {
		return errors.New("target_user_id is required")
	}
	if input.AdminUserID == "" {
		return errors.New("admin_user_id is required")
	}
	if input.TenantID == "" {
		return errors.New("tenant_id is required")
	}

	// Cannot disable self
	if input.TargetUserID == input.AdminUserID {
		return ErrCannotDisableSelf
	}

	// Check if target is an admin user (admin_users table)
	targetUUID, err := uuid.Parse(input.TargetUserID)
	if err == nil {
		adminUser, err := uc.userRepo.FindByID(ctx, targetUUID)
		if err == nil && adminUser != nil {
			if adminUser.Role == entities.UserRoleAdmin || adminUser.Role == entities.UserRoleOwner {
				return ErrCannotDisableAdmin
			}
		}
	}

	// Verify user exists and belongs to tenant
	user, err := uc.endUserRepo.GetByID(ctx, input.TargetUserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if user.TenantID != input.TenantID {
		return errors.New("user not found")
	}

	// Update status to disabled
	if err := uc.endUserRepo.UpdateStatus(ctx, input.TargetUserID, entities.UserStatusDisabled); err != nil {
		return fmt.Errorf("failed to disable user: %w", err)
	}

	// Revoke all refresh tokens
	_ = uc.refreshTokenRepo.RevokeByUserID(ctx, input.TargetUserID)

	// Add to blacklist (15 min TTL matching access token lifetime)
	_ = uc.userBlacklist.Add(ctx, input.TargetUserID, 900*time.Second)

	// Record event
	now := time.Now()
	event := &entities.UserEvent{
		TenantID:   input.TenantID,
		UserID:     input.TargetUserID,
		EventType:  entities.EventTypeAccountDisabled,
		Details:    map[string]any{"disabled_by": input.AdminUserID},
		OccurredAt: now,
	}
	_ = uc.eventRepo.Record(ctx, event)

	// Audit log
	auditLog := entities.NewAuditLog("account_disabled", "", "")
	auditLog.WithData("user_id", input.TargetUserID)
	auditLog.WithData("disabled_by", input.AdminUserID)
	_ = uc.auditRepo.Create(ctx, auditLog)

	return nil
}
