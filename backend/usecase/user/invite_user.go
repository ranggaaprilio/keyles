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

// InviteUserRequest represents a request to invite a new user to a tenant
type InviteUserRequest struct {
	TenantID    string
	Email       string
	DisplayName string
	InvitedBy   string
}

// InviteUserResponse contains the created invitation and the plain token
type InviteUserResponse struct {
	Invitation *entities.Invitation
	PlainToken string
}

// InviteUser handles inviting a new user to a tenant (T035)
type InviteUser struct {
	userRepo          repositories.EndUserRepository
	invRepo           repositories.InvitationRepository
	eventRepo         repositories.UserEventRepository
	emailService      services.EmailService
	tenantRepo        repositories.TenantRepository
	invitationBaseURL string
}

// NewInviteUser creates a new InviteUser use case
func NewInviteUser(
	userRepo repositories.EndUserRepository,
	invRepo repositories.InvitationRepository,
	eventRepo repositories.UserEventRepository,
	emailService services.EmailService,
	tenantRepo repositories.TenantRepository,
	invitationBaseURL string,
) *InviteUser {
	return &InviteUser{
		userRepo:          userRepo,
		invRepo:           invRepo,
		eventRepo:         eventRepo,
		emailService:      emailService,
		tenantRepo:        tenantRepo,
		invitationBaseURL: invitationBaseURL,
	}
}

// Execute invites a new user to a tenant
func (uc *InviteUser) Execute(ctx context.Context, req InviteUserRequest) (*InviteUserResponse, error) {
	if req.TenantID == "" {
		return nil, errors.New("tenant_id is required")
	}
	if req.Email == "" {
		return nil, errors.New("email is required")
	}
	if req.InvitedBy == "" {
		return nil, errors.New("invited_by is required")
	}

	if err := entities.ValidateEmail(req.Email); err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	if req.DisplayName != "" {
		if err := entities.ValidateFullName(req.DisplayName); err != nil {
			return nil, fmt.Errorf("invalid display_name: %w", err)
		}
	}

	tenantUUID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, errors.New("invalid tenant_id format")
	}

	invitedByUUID, err := uuid.Parse(req.InvitedBy)
	if err != nil {
		return nil, errors.New("invalid invited_by format")
	}

	count, err := uc.userRepo.CountByTenant(ctx, tenantUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user quota: %w", err)
	}
	if count >= entities.MaxUsersPerTenant {
		return nil, errors.New("user quota exceeded")
	}

	normalizedEmail := entities.NormalizeEmail(req.Email)
	existingUser, err := uc.userRepo.GetByEmail(ctx, tenantUUID, normalizedEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	existingInv, err := uc.invRepo.GetPendingByEmail(ctx, tenantUUID, normalizedEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to check pending invitation: %w", err)
	}
	if existingInv != nil {
		return nil, errors.New("pending invitation already exists for this email")
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}
	plainToken := hex.EncodeToString(tokenBytes)

	tokenHash, err := bcrypt.GenerateFromPassword([]byte(plainToken), 10)
	if err != nil {
		return nil, fmt.Errorf("failed to hash invitation token: %w", err)
	}

	now := time.Now()
	var displayNamePtr *string
	if req.DisplayName != "" {
		dn := req.DisplayName
		displayNamePtr = &dn
	}
	invitation := &entities.Invitation{
		ID:          uuid.New(),
		TenantID:    tenantUUID,
		Email:       normalizedEmail,
		DisplayName: displayNamePtr,
		TokenHash:   string(tokenHash),
		Status:      entities.InvitationStatusPending,
		InvitedBy:   invitedByUUID,
		ExpiresAt:   now.Add(entities.InvitationTTL),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.invRepo.Create(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	user, err := entities.NewUser(tenantUUID, normalizedEmail, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	if displayNamePtr != nil {
		user.DisplayName = displayNamePtr
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	event := &entities.UserEvent{
		TenantID:   req.TenantID,
		UserID:     user.ID.String(),
		EventType:  entities.EventTypeUserInvited,
		Details:    map[string]any{"invited_by": req.InvitedBy, "email": normalizedEmail},
		OccurredAt: now,
	}
	if err := uc.eventRepo.Record(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to record user event: %w", err)
	}

	tenant, err := uc.tenantRepo.FindByID(ctx, tenantUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	inviteURL := fmt.Sprintf("%s/accept?token=%s", uc.invitationBaseURL, plainToken)

	toName := req.DisplayName
	if toName == "" {
		toName = normalizedEmail
	}
	if err := uc.emailService.SendInvitationEmail(ctx, normalizedEmail, toName, inviteURL, tenant.OrganizationName); err != nil {
		return nil, fmt.Errorf("failed to send invitation email: %w", err)
	}

	return &InviteUserResponse{
		Invitation: invitation,
		PlainToken: plainToken,
	}, nil
}
