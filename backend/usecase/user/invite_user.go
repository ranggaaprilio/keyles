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

// InviteUserInput represents the request to invite a new user
type InviteUserInput struct {
	TenantID      string
	Email         string
	DisplayName   string
	InvitedBy     string // admin user ID performing the invitation
	InviteBaseURL string // base URL for the invitation link
}

// InviteUserOutput represents the response after inviting a user
type InviteUserOutput struct {
	UserID       string
	Email        string
	DisplayName  string
	Status       entities.UserStatus
	InvitationID string
	ExpiresAt    time.Time
}

var (
ErrQuotaExceeded           = errors.New("user quota exceeded: maximum 10,000 users per tenant")
ErrEmailAlreadyExists      = errors.New("a user with this email already exists in this tenant")
ErrPendingInvitationExists = errors.New("a pending invitation already exists for this email")
)

// InviteUser handles the user invitation workflow
type InviteUser struct {
	endUserRepo    repositories.EndUserRepository
	invitationRepo repositories.InvitationRepository
	eventRepo      repositories.UserEventRepository
	emailService   services.EmailService
	countCache     services.UserCountCache
}

// NewInviteUser creates a new InviteUser use case
func NewInviteUser(
endUserRepo repositories.EndUserRepository,
invitationRepo repositories.InvitationRepository,
eventRepo repositories.UserEventRepository,
emailService services.EmailService,
countCache services.UserCountCache,
) *InviteUser {
	return &InviteUser{
		endUserRepo:    endUserRepo,
		invitationRepo: invitationRepo,
		eventRepo:      eventRepo,
		emailService:   emailService,
		countCache:     countCache,
	}
}

// Execute invites a new user to the tenant
func (uc *InviteUser) Execute(ctx context.Context, input InviteUserInput) (*InviteUserOutput, error) {
	// Validate input
	if input.TenantID == "" {
		return nil, errors.New("tenant_id is required")
	}
	if input.Email == "" {
		return nil, errors.New("email is required")
	}
	if input.InvitedBy == "" {
		return nil, errors.New("invited_by is required")
	}

	// Check tenant user quota (cache-first, fall back to DB)
	count, hit, _ := uc.countCache.Get(ctx, input.TenantID)
	if !hit {
		var err error
		count, err = uc.endUserRepo.CountByTenant(ctx, input.TenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to check user count: %w", err)
		}
		_ = uc.countCache.Set(ctx, input.TenantID, count)
	}
	if count >= entities.MaxUsersPerTenant {
		return nil, ErrQuotaExceeded
	}

	// Check if email already exists in this tenant
	existingUser, err := uc.endUserRepo.GetByEmail(ctx, input.TenantID, input.Email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Check for pending invitation for same email
	existingInv, err := uc.invitationRepo.GetPendingByEmail(ctx, input.TenantID, input.Email)
	if err == nil && existingInv != nil {
		return nil, ErrPendingInvitationExists
	}

	// Generate 32-byte cryptographically secure token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}
	plainToken := base64.URLEncoding.EncodeToString(tokenBytes)

	// Hash the token with bcrypt
	tokenHash, err := bcrypt.GenerateFromPassword([]byte(plainToken), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash invitation token: %w", err)
	}

	// Create pending user
	user := entities.NewUser(input.TenantID, input.Email, input.DisplayName)
	user.ID = uuid.New().String()

	if err := uc.endUserRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create invitation record
	now := time.Now()
	invitation := &entities.Invitation{
		ID:          uuid.New().String(),
		TenantID:    input.TenantID,
		Email:       input.Email,
		DisplayName: input.DisplayName,
		TokenHash:   string(tokenHash),
		Status:      entities.InvitationStatusPending,
		InvitedBy:   input.InvitedBy,
		ExpiresAt:   now.Add(entities.InvitationTTL),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Send invitation email
	inviteURL := fmt.Sprintf("%s/%s/accept", input.InviteBaseURL, plainToken)
	_ = uc.emailService.SendInvitationEmail(ctx, input.Email, input.DisplayName, inviteURL, input.TenantID)

	// Record user_invited event (fire-and-forget)
	event := &entities.UserEvent{
		TenantID:   input.TenantID,
		UserID:     user.ID,
		EventType:  entities.EventTypeUserInvited,
		Details:    map[string]any{"invited_by": input.InvitedBy, "email": input.Email},
		OccurredAt: now,
	}
	_ = uc.eventRepo.Record(ctx, event)

	// Invalidate user count cache
	_ = uc.countCache.Invalidate(ctx, input.TenantID)

	return &InviteUserOutput{
		UserID:       user.ID,
		Email:        user.Email,
		DisplayName:  user.DisplayName,
		Status:       user.Status,
		InvitationID: invitation.ID,
		ExpiresAt:    invitation.ExpiresAt,
	}, nil
}
