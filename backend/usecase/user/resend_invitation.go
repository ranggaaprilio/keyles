package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
)

// ResendInvitationRequest represents a request to resend an invitation
type ResendInvitationRequest struct {
	UserID   string
	TenantID string
}

// ResendInvitation handles resending an invitation to a pending user (T037)
type ResendInvitation struct {
	userRepo          repositories.EndUserRepository
	invRepo           repositories.InvitationRepository
	eventRepo         repositories.UserEventRepository
	emailService      services.EmailService
	tenantRepo        repositories.TenantRepository
	invitationBaseURL string
}

// NewResendInvitation creates a new ResendInvitation use case
func NewResendInvitation(
	userRepo repositories.EndUserRepository,
	invRepo repositories.InvitationRepository,
	eventRepo repositories.UserEventRepository,
	emailService services.EmailService,
	tenantRepo repositories.TenantRepository,
	invitationBaseURL string,
) *ResendInvitation {
	return &ResendInvitation{
		userRepo:          userRepo,
		invRepo:           invRepo,
		eventRepo:         eventRepo,
		emailService:      emailService,
		tenantRepo:        tenantRepo,
		invitationBaseURL: invitationBaseURL,
	}
}

// Execute resends an invitation to a pending user
func (uc *ResendInvitation) Execute(ctx context.Context, req ResendInvitationRequest) error {
	if req.UserID == "" {
		return errors.New("user_id is required")
	}
	if req.TenantID == "" {
		return errors.New("tenant_id is required")
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return errors.New("invalid user_id format")
	}

	tenantUUID, err := uuid.Parse(req.TenantID)
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

	if user.Status != entities.UserStatusPending {
		return errors.New("user is not in pending status")
	}

	invitation, err := uc.invRepo.GetPendingByEmail(ctx, tenantUUID, user.Email)
	if err != nil {
		return fmt.Errorf("failed to find pending invitation: %w", err)
	}
	if invitation == nil {
		return errors.New("no pending invitation found for this user")
	}

	if err := uc.invRepo.UpdateStatus(ctx, invitation.ID, entities.InvitationStatusExpired, nil); err != nil {
		return fmt.Errorf("failed to expire current invitation: %w", err)
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate invitation token: %w", err)
	}
	plainToken := hex.EncodeToString(tokenBytes)

	tokenHash, err := bcrypt.GenerateFromPassword([]byte(plainToken), 10)
	if err != nil {
		return fmt.Errorf("failed to hash invitation token: %w", err)
	}

	now := time.Now()
	newInvitation := &entities.Invitation{
		ID:          uuid.New(),
		TenantID:    tenantUUID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		TokenHash:   string(tokenHash),
		Status:      entities.InvitationStatusPending,
		InvitedBy:   invitation.InvitedBy,
		ExpiresAt:   now.Add(entities.InvitationTTL),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.invRepo.Create(ctx, newInvitation); err != nil {
		return fmt.Errorf("failed to create new invitation: %w", err)
	}

	event := &entities.UserEvent{
		TenantID:   req.TenantID,
		UserID:     req.UserID,
		EventType:  entities.EventTypeInvitationResent,
		Details:    map[string]any{"email": user.Email},
		OccurredAt: now,
	}
	if err := uc.eventRepo.Record(ctx, event); err != nil {
		return fmt.Errorf("failed to record user event: %w", err)
	}

	tenant, err := uc.tenantRepo.FindByID(ctx, tenantUUID)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	inviteURL := fmt.Sprintf("%s/accept?token=%s", uc.invitationBaseURL, plainToken)

	toName := user.Email
	if user.DisplayName != nil && *user.DisplayName != "" {
		toName = *user.DisplayName
	}
	if err := uc.emailService.SendInvitationEmail(ctx, user.Email, toName, inviteURL, tenant.OrganizationName); err != nil {
		return fmt.Errorf("failed to send invitation email: %w", err)
	}

	return nil
}
