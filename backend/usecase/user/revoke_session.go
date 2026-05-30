package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// RevokeSession handles revoking a single user session (US4)
type RevokeSession struct {
	refreshTokenRepo repositories.RefreshTokenRepository
	eventRepo        repositories.UserEventRepository
}

// NewRevokeSession creates a new RevokeSession use case
func NewRevokeSession(
	refreshTokenRepo repositories.RefreshTokenRepository,
	eventRepo repositories.UserEventRepository,
) *RevokeSession {
	return &RevokeSession{
		refreshTokenRepo: refreshTokenRepo,
		eventRepo:        eventRepo,
	}
}

// Execute revokes a specific refresh token and records the event
func (uc *RevokeSession) Execute(ctx context.Context, userID string, sessionTokenHash string, adminUserID string) error {
	if userID == "" {
		return errors.New("user_id is required")
	}
	if sessionTokenHash == "" {
		return errors.New("session_token_hash is required")
	}

	if err := uc.refreshTokenRepo.Revoke(ctx, sessionTokenHash, adminUserID); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	event := &entities.UserEvent{
		TenantID:   "",
		UserID:     userID,
		EventType:  entities.EventTypeSessionTerminated,
		Details:    map[string]any{"revoked_by": adminUserID, "token_hash": sessionTokenHash},
		OccurredAt: time.Now(),
	}
	if err := uc.eventRepo.Record(ctx, event); err != nil {
		return fmt.Errorf("failed to record user event: %w", err)
	}

	return nil
}
