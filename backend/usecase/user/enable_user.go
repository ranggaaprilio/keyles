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

// EnableUser handles enabling a previously disabled user account (US5)
type EnableUser struct {
	userRepo  repositories.EndUserRepository
	eventRepo repositories.UserEventRepository
}

// NewEnableUser creates a new EnableUser use case
func NewEnableUser(
	userRepo repositories.EndUserRepository,
	eventRepo repositories.UserEventRepository,
) *EnableUser {
	return &EnableUser{
		userRepo:  userRepo,
		eventRepo: eventRepo,
	}
}

// Execute enables a user account
// NOTE: previously revoked sessions are NOT restored
func (uc *EnableUser) Execute(ctx context.Context, userID string, tenantID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user_id format")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return errors.New("invalid tenant_id format")
	}

	user, err := uc.userRepo.GetByID(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	if user.TenantID != tenantUUID {
		return errors.New("tenant mismatch")
	}

	if err := uc.userRepo.UpdateStatus(ctx, userUUID, entities.UserStatusActive); err != nil {
		return fmt.Errorf("failed to enable user: %w", err)
	}

	event := &entities.UserEvent{
		TenantID:   tenantID,
		UserID:     userID,
		EventType:  entities.EventTypeAccountEnabled,
		Details:    map[string]any{},
		OccurredAt: time.Now(),
	}
	if err := uc.eventRepo.Record(ctx, event); err != nil {
		return fmt.Errorf("failed to record user event: %w", err)
	}

	return nil
}
