package user

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotPending = errors.New("user is not in pending status; cannot resend invitation")
)

// ResendInvitationInput represents the request to resend an invitation
type ResendInvitationInput struct {
	UserID        string
	TenantID      string
	InviteBaseURL string
}

// ResendInvitation handles resending an invitation to a pending user
type ResendInvitation struct {
	endUserRepo    repositories.EndUserRepository
	invitationRepo repositories.InvitationRepository
	eventRepo      repositories.UserEventRepository
	emailService   services.EmailService
}

// NewResendInvitation creates a new ResendInvitation use case
func NewResendInvitation(
	endUserRepo repositories.EndUserRepository,
	invitationRepo repositories.InvitationRepository,
	eventRepo repositories.UserEventRepository,
	emailService services.EmailService,
) *ResendInvitation {
	return &ResendInvitation{
		endUserRepo:    endUserRepo,
		invitationRepo: invitationRepo,
		eventRepo:      eventRepo,
		emailService:   emailService,
	}
}

// Execute resends an invitation to a pending user
func (uc *ResendInvitation) Execute(ctx context.Context, input ResendInvitationInput) error {
	if input.UserID == "" {
		return errors.New("user_id is required")
	}
	if input.TenantID == "" {
		return errors.New("tenant_id is required")
	}

	// Fetch user to verify they exist and are pending
	user, err := uc.endUserRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify tenant isolation
	if user.TenantID != input.TenantID {
		return errors.New("user not found")
	}

	if user.Status != entities.UserStatusPending {
		return ErrUserNotPending
	}

	// Expire current pending invitation
	existingInv, err := uc.invitationRepo.GetPendingByEmail(ctx, input.TenantID, user.Email)
	if err == nil && existingInv != nil {
		_ = uc.invitationRepo.UpdateStatus(ctx, existingInv.ID, entities.InvitationStatusExpired, nil)
	}

	// Generate new token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate invitation token: %w", err)
	}
	plainToken := base64.URLEncoding.EncodeToString(tokenBytes)

	tokenHash, err := bcrypt.GenerateFromPassword([]byte(plainToken), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash invitation token: %w", err)
	}

	// Create new invitation with fresh 72h expiry
	now := time.Now()
	invitation := &entities.Invitation{
		ID:          uuid.New().String(),
		TenantID:    input.TenantID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		TokenHash:   string(tokenHash),
		Status:      entities.InvitationStatusPending,
		InvitedBy:   input.UserID, // track who resent
		ExpiresAt:   now.Add(entities.InvitationTTL),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.invitationRepo.Create(ctx, invitation); err != nil {
		return fmt.Errorf("failed to create invitation: %w", err)
	}

	// Send invitation email
	inviteURL := fmt.Sprintf("%s/%s/accept", input.InviteBaseURL, plainToken)
	_ = uc.emailService.SendInvitationEmail(ctx, user.Email, user.DisplayName, inviteURL, input.TenantID)

	// Record invitation_resent event
	event := &entities.UserEvent{
		TenantID:   input.TenantID,
		UserID:     user.ID,
		EventType:  entities.EventTypeInvitationResent,
		Details:    map[string]any{"invitation_id": invitation.ID},
		OccurredAt: now,
	}
	_ = uc.eventRepo.Record(ctx, event)

	return nil
}
