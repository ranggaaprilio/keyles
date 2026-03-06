package user

import (
"context"
"errors"
"fmt"

"github.com/ranggaaprilio/keyles/domain/entities"
"github.com/ranggaaprilio/keyles/domain/repositories"
)

// RevokeSessionInput represents the request to revoke a session
type RevokeSessionInput struct {
	UserID    string
	TenantID  string
	TokenID   int64
	RevokedBy string
}

// RevokeSession handles revoking a single user session
type RevokeSession struct {
	endUserRepo      repositories.EndUserRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	eventRepo        repositories.UserEventRepository
}

// NewRevokeSession creates a new RevokeSession use case
func NewRevokeSession(
endUserRepo repositories.EndUserRepository,
refreshTokenRepo repositories.RefreshTokenRepository,
eventRepo repositories.UserEventRepository,
) *RevokeSession {
	return &RevokeSession{
		endUserRepo:      endUserRepo,
		refreshTokenRepo: refreshTokenRepo,
		eventRepo:        eventRepo,
	}
}

// Execute revokes a specific session (refresh token)
func (uc *RevokeSession) Execute(ctx context.Context, input RevokeSessionInput) error {
	if input.UserID == "" {
		return errors.New("user_id is required")
	}
	if input.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if input.TokenID == 0 {
		return errors.New("token_id is required")
	}

	// Verify user exists and belongs to tenant
	user, err := uc.endUserRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if user.TenantID != input.TenantID {
		return errors.New("user not found")
	}

	// Fetch the token to verify it belongs to the user
	token, err := uc.refreshTokenRepo.GetByID(ctx, input.TokenID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}
	if token.UserID != input.UserID {
		return errors.New("session not found")
	}
	if token.IsRevoked() {
		return errors.New("session already revoked")
	}

	// Revoke the token
	if err := uc.refreshTokenRepo.Revoke(ctx, token.Token, input.RevokedBy); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	// Record session_terminated event
	event := &entities.UserEvent{
		TenantID:  input.TenantID,
		UserID:    input.UserID,
		EventType: entities.EventTypeSessionTerminated,
		Details:   map[string]any{"token_id": input.TokenID, "revoked_by": input.RevokedBy},
	}
	_ = uc.eventRepo.Record(ctx, event)

	return nil
}
