package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
)

var (
	ErrInvitationExpired  = errors.New("invitation has expired")
	ErrInvitationAccepted = errors.New("invitation has already been accepted")
	ErrWeakPassword       = errors.New("password does not meet complexity requirements")
)

// AcceptInvitationInput represents the request to accept an invitation
type AcceptInvitationInput struct {
	Token    string // plain-text invitation token
	Password string // new password for the user account
}

// AcceptInvitation handles accepting a user invitation and activating the account
type AcceptInvitation struct {
	endUserRepo    repositories.EndUserRepository
	invitationRepo repositories.InvitationRepository
	eventRepo      repositories.UserEventRepository
	passwordSvc    services.PasswordService
}

// NewAcceptInvitation creates a new AcceptInvitation use case
func NewAcceptInvitation(
	endUserRepo repositories.EndUserRepository,
	invitationRepo repositories.InvitationRepository,
	eventRepo repositories.UserEventRepository,
	passwordSvc services.PasswordService,
) *AcceptInvitation {
	return &AcceptInvitation{
		endUserRepo:    endUserRepo,
		invitationRepo: invitationRepo,
		eventRepo:      eventRepo,
		passwordSvc:    passwordSvc,
	}
}

// Execute accepts an invitation and activates the user account
func (uc *AcceptInvitation) Execute(ctx context.Context, input AcceptInvitationInput) error {
	if input.Token == "" {
		return errors.New("token is required")
	}
	if input.Password == "" {
		return errors.New("password is required")
	}

	// Validate password strength using existing validation
	if err := entities.ValidatePassword(input.Password); err != nil {
		return fmt.Errorf("%w: %v", ErrWeakPassword, err)
	}

	// Look up invitation by token (bcrypt comparison happens in repository)
	invitation, err := uc.invitationRepo.GetByToken(ctx, input.Token)
	if err != nil {
		return fmt.Errorf("invalid invitation token: %w", err)
	}

	// Check invitation status
	if invitation.IsAccepted() {
		return ErrInvitationAccepted
	}
	if invitation.IsExpired() {
		return ErrInvitationExpired
	}
	if invitation.Status != entities.InvitationStatusPending {
		return errors.New("invitation is not in a valid state")
	}

	// Hash password with bcrypt (cost >= 12)
	passwordHash, err := uc.passwordSvc.Hash(input.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Fetch the pending user by email
	user, err := uc.endUserRepo.GetByEmail(ctx, invitation.TenantID, invitation.Email)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Update user: set password hash and activate
	user.PasswordHash = passwordHash
	user.Status = entities.UserStatusActive
	user.UpdatedAt = time.Now()

	if err := uc.endUserRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	// Mark invitation as accepted
	now := time.Now()
	if err := uc.invitationRepo.UpdateStatus(ctx, invitation.ID, entities.InvitationStatusAccepted, &now); err != nil {
		return fmt.Errorf("failed to update invitation status: %w", err)
	}

	// Record invitation_accepted event (fire-and-forget)
	event := &entities.UserEvent{
		TenantID:   invitation.TenantID,
		UserID:     user.ID,
		EventType:  entities.EventTypeInvitationAccepted,
		Details:    map[string]any{"invitation_id": invitation.ID},
		OccurredAt: now,
	}
	_ = uc.eventRepo.Record(ctx, event)

	return nil
}
