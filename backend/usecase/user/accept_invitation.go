package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// AcceptInvitationRequest represents a request to accept an invitation
type AcceptInvitationRequest struct {
	PlainToken string
	Password   string
}

// AcceptInvitation handles accepting a user invitation (T036)
type AcceptInvitation struct {
	userRepo repositories.EndUserRepository
	invRepo  repositories.InvitationRepository
}

// NewAcceptInvitation creates a new AcceptInvitation use case
func NewAcceptInvitation(
	userRepo repositories.EndUserRepository,
	invRepo repositories.InvitationRepository,
) *AcceptInvitation {
	return &AcceptInvitation{
		userRepo: userRepo,
		invRepo:  invRepo,
	}
}

// Execute accepts an invitation and activates the user account
func (uc *AcceptInvitation) Execute(ctx context.Context, req AcceptInvitationRequest) error {
	if req.PlainToken == "" {
		return errors.New("token is required")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}

	if err := entities.ValidatePassword(req.Password); err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}

	invitation, err := uc.invRepo.GetByToken(ctx, req.PlainToken)
	if err != nil {
		return fmt.Errorf("failed to retrieve invitation: %w", err)
	}
	if invitation == nil {
		return errors.New("invitation not found")
	}

	if invitation.IsExpired() {
		return errors.New("invitation has expired")
	}

	if invitation.IsAccepted() {
		return errors.New("invitation has already been accepted")
	}

	if invitation.Status != entities.InvitationStatusPending {
		return errors.New("invitation is no longer valid")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := uc.userRepo.GetByEmail(ctx, invitation.TenantID, invitation.Email)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return errors.New("user not found for this invitation")
	}

	if user.Status != entities.UserStatusPending {
		return errors.New("user is not in pending status")
	}

	user.PasswordHash = string(passwordHash)
	user.Status = entities.UserStatusActive
	user.UpdatedAt = time.Now()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	now := time.Now()
	if err := uc.invRepo.UpdateStatus(ctx, invitation.ID, entities.InvitationStatusAccepted, &now); err != nil {
		return fmt.Errorf("failed to update invitation status: %w", err)
	}

	return nil
}
