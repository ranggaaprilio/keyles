package user

import (
"context"
"errors"
"fmt"
"time"

"github.com/ranggaaprilio/keyles/domain/entities"
"github.com/ranggaaprilio/keyles/domain/repositories"
)

var (
ErrCanOnlyEnableDisabled = errors.New("can only enable disabled users")
)

// EnableUserInput represents the request to enable a user
type EnableUserInput struct {
	TargetUserID string
	AdminUserID  string
	TenantID     string
}

// EnableUser handles re-enabling a disabled user account
type EnableUser struct {
	endUserRepo repositories.EndUserRepository
	eventRepo   repositories.UserEventRepository
	auditRepo   repositories.AuditRepository
}

// NewEnableUser creates a new EnableUser use case
func NewEnableUser(
endUserRepo repositories.EndUserRepository,
eventRepo repositories.UserEventRepository,
auditRepo repositories.AuditRepository,
) *EnableUser {
	return &EnableUser{
		endUserRepo: endUserRepo,
		eventRepo:   eventRepo,
		auditRepo:   auditRepo,
	}
}

// Execute enables a disabled user account
func (uc *EnableUser) Execute(ctx context.Context, input EnableUserInput) error {
	if input.TargetUserID == "" {
		return errors.New("target_user_id is required")
	}
	if input.TenantID == "" {
		return errors.New("tenant_id is required")
	}

	// Verify user exists and belongs to tenant
	user, err := uc.endUserRepo.GetByID(ctx, input.TargetUserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if user.TenantID != input.TenantID {
		return errors.New("user not found")
	}

	// Can only enable disabled users
	if user.Status == entities.UserStatusActive {
		return nil // idempotent
	}
	if user.Status == entities.UserStatusPending {
		return ErrCanOnlyEnableDisabled
	}

	// Update status to active
	if err := uc.endUserRepo.UpdateStatus(ctx, input.TargetUserID, entities.UserStatusActive); err != nil {
		return fmt.Errorf("failed to enable user: %w", err)
	}

	// Record event
	now := time.Now()
	event := &entities.UserEvent{
		TenantID:   input.TenantID,
		UserID:     input.TargetUserID,
		EventType:  entities.EventTypeAccountEnabled,
		Details:    map[string]any{"enabled_by": input.AdminUserID},
		OccurredAt: now,
	}
	_ = uc.eventRepo.Record(ctx, event)

	// Audit log
	auditLog := entities.NewAuditLog("account_enabled", "", "")
	auditLog.WithData("user_id", input.TargetUserID)
	auditLog.WithData("enabled_by", input.AdminUserID)
	_ = uc.auditRepo.Create(ctx, auditLog)

	return nil
}
