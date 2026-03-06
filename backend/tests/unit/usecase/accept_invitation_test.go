package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAcceptInvitation_HappyPath(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	invitationRepo := new(mocks.MockInvitationRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	passwordSvc := new(mocks.MockPasswordService)

	invitation := &entities.Invitation{
		ID:        "inv-1",
		TenantID:  "tenant-1",
		Email:     "user@example.com",
		TokenHash: "$2a$10$validhash",
		Status:    entities.InvitationStatusPending,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	pendingUser := &entities.User{
		ID:       "user-1",
		TenantID: "tenant-1",
		Email:    "user@example.com",
		Status:   entities.UserStatusPending,
	}

	invitationRepo.On("GetByToken", ctx, "valid-token").Return(invitation, nil)
	passwordSvc.On("Hash", "StrongP@ss1").Return("$2a$12$hashedpassword", nil)
	endUserRepo.On("GetByEmail", ctx, "tenant-1", "user@example.com").Return(pendingUser, nil)
	endUserRepo.On("Update", ctx, mock.MatchedBy(func(u *entities.User) bool {
return u.ID == "user-1" && u.Status == entities.UserStatusActive && u.PasswordHash == "$2a$12$hashedpassword"
	})).Return(nil)
	invitationRepo.On("UpdateStatus", ctx, "inv-1", entities.InvitationStatusAccepted, mock.AnythingOfType("*time.Time")).Return(nil)
	eventRepo.On("Record", ctx, mock.MatchedBy(func(e *entities.UserEvent) bool {
return e.EventType == entities.EventTypeInvitationAccepted && e.UserID == "user-1"
	})).Return(nil)

	uc := user.NewAcceptInvitation(endUserRepo, invitationRepo, eventRepo, passwordSvc)
	err := uc.Execute(ctx, user.AcceptInvitationInput{Token: "valid-token", Password: "StrongP@ss1"})

	assert.NoError(t, err)
	endUserRepo.AssertExpectations(t)
	invitationRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
	passwordSvc.AssertExpectations(t)
}

func TestAcceptInvitation_ExpiredInvitation(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	invitationRepo := new(mocks.MockInvitationRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	passwordSvc := new(mocks.MockPasswordService)

	invitation := &entities.Invitation{
		ID:        "inv-1",
		TenantID:  "tenant-1",
		Email:     "user@example.com",
		Status:    entities.InvitationStatusPending,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	invitationRepo.On("GetByToken", ctx, "expired-token").Return(invitation, nil)

	uc := user.NewAcceptInvitation(endUserRepo, invitationRepo, eventRepo, passwordSvc)
	err := uc.Execute(ctx, user.AcceptInvitationInput{Token: "expired-token", Password: "StrongP@ss1"})

	assert.ErrorIs(t, err, user.ErrInvitationExpired)
}

func TestAcceptInvitation_AlreadyAccepted(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	invitationRepo := new(mocks.MockInvitationRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	passwordSvc := new(mocks.MockPasswordService)

	invitation := &entities.Invitation{
		ID:        "inv-1",
		TenantID:  "tenant-1",
		Email:     "user@example.com",
		Status:    entities.InvitationStatusAccepted,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	invitationRepo.On("GetByToken", ctx, "used-token").Return(invitation, nil)

	uc := user.NewAcceptInvitation(endUserRepo, invitationRepo, eventRepo, passwordSvc)
	err := uc.Execute(ctx, user.AcceptInvitationInput{Token: "used-token", Password: "StrongP@ss1"})

	assert.ErrorIs(t, err, user.ErrInvitationAccepted)
}

func TestAcceptInvitation_WeakPassword(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	invitationRepo := new(mocks.MockInvitationRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	passwordSvc := new(mocks.MockPasswordService)

	uc := user.NewAcceptInvitation(endUserRepo, invitationRepo, eventRepo, passwordSvc)
	err := uc.Execute(ctx, user.AcceptInvitationInput{Token: "some-token", Password: "weak"})

	assert.ErrorIs(t, err, user.ErrWeakPassword)
}

func TestAcceptInvitation_InvalidToken(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	invitationRepo := new(mocks.MockInvitationRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	passwordSvc := new(mocks.MockPasswordService)

	invitationRepo.On("GetByToken", ctx, "invalid-token").Return(nil, errors.New("not found"))

	uc := user.NewAcceptInvitation(endUserRepo, invitationRepo, eventRepo, passwordSvc)
	err := uc.Execute(ctx, user.AcceptInvitationInput{Token: "invalid-token", Password: "StrongP@ss1"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid invitation token")
}

func TestAcceptInvitation_MissingFields(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	invitationRepo := new(mocks.MockInvitationRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	passwordSvc := new(mocks.MockPasswordService)

	uc := user.NewAcceptInvitation(endUserRepo, invitationRepo, eventRepo, passwordSvc)

	err := uc.Execute(ctx, user.AcceptInvitationInput{Token: "", Password: "StrongP@ss1"})
	assert.Contains(t, err.Error(), "token is required")

	err = uc.Execute(ctx, user.AcceptInvitationInput{Token: "tok", Password: ""})
	assert.Contains(t, err.Error(), "password is required")
}
