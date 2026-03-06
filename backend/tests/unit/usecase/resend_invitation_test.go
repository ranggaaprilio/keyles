package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResendInvitation_HappyPath(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	invitationRepo := new(mocks.MockInvitationRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	emailService := new(mocks.MockEmailService)

	pendingUser := &entities.User{
		ID:          "user-1",
		TenantID:    "tenant-1",
		Email:       "user@example.com",
		DisplayName: "Test User",
		Status:      entities.UserStatusPending,
	}

	existingInv := &entities.Invitation{
		ID:       "old-inv-1",
		TenantID: "tenant-1",
		Email:    "user@example.com",
		Status:   entities.InvitationStatusPending,
	}

	endUserRepo.On("GetByID", ctx, "user-1").Return(pendingUser, nil)
	invitationRepo.On("GetPendingByEmail", ctx, "tenant-1", "user@example.com").Return(existingInv, nil)
	invitationRepo.On("UpdateStatus", ctx, "old-inv-1", entities.InvitationStatusExpired, mock.Anything).Return(nil)

	invitationRepo.On("Create", ctx, mock.MatchedBy(func(inv *entities.Invitation) bool {
return inv.TenantID == "tenant-1" && inv.Email == "user@example.com" && inv.Status == entities.InvitationStatusPending && inv.TokenHash != ""
	})).Return(nil)

	emailService.On("SendInvitationEmail", ctx, "user@example.com", "Test User", mock.AnythingOfType("string"), "tenant-1").Return(nil)

	eventRepo.On("Record", ctx, mock.MatchedBy(func(e *entities.UserEvent) bool {
return e.EventType == entities.EventTypeInvitationResent && e.UserID == "user-1"
	})).Return(nil)

	uc := user.NewResendInvitation(endUserRepo, invitationRepo, eventRepo, emailService)
	err := uc.Execute(ctx, user.ResendInvitationInput{
UserID:        "user-1",
TenantID:      "tenant-1",
InviteBaseURL: "https://auth.example.com/invite",
})

	assert.NoError(t, err)
	endUserRepo.AssertExpectations(t)
	invitationRepo.AssertExpectations(t)
	emailService.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestResendInvitation_UserNotPending(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	invitationRepo := new(mocks.MockInvitationRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	emailService := new(mocks.MockEmailService)

	activeUser := &entities.User{
		ID:       "user-1",
		TenantID: "tenant-1",
		Email:    "user@example.com",
		Status:   entities.UserStatusActive,
	}

	endUserRepo.On("GetByID", ctx, "user-1").Return(activeUser, nil)

	uc := user.NewResendInvitation(endUserRepo, invitationRepo, eventRepo, emailService)
	err := uc.Execute(ctx, user.ResendInvitationInput{
UserID:   "user-1",
TenantID: "tenant-1",
})

	assert.ErrorIs(t, err, user.ErrUserNotPending)
}

func TestResendInvitation_UserNotFound(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	invitationRepo := new(mocks.MockInvitationRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	emailService := new(mocks.MockEmailService)

	endUserRepo.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	uc := user.NewResendInvitation(endUserRepo, invitationRepo, eventRepo, emailService)
	err := uc.Execute(ctx, user.ResendInvitationInput{
UserID:   "nonexistent",
TenantID: "tenant-1",
})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestResendInvitation_CrossTenantRejected(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	invitationRepo := new(mocks.MockInvitationRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	emailService := new(mocks.MockEmailService)

	otherTenantUser := &entities.User{
		ID:       "user-1",
		TenantID: "other-tenant",
		Email:    "user@example.com",
		Status:   entities.UserStatusPending,
	}

	endUserRepo.On("GetByID", ctx, "user-1").Return(otherTenantUser, nil)

	uc := user.NewResendInvitation(endUserRepo, invitationRepo, eventRepo, emailService)
	err := uc.Execute(ctx, user.ResendInvitationInput{
UserID:   "user-1",
TenantID: "tenant-1",
})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestResendInvitation_MissingFields(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	invitationRepo := new(mocks.MockInvitationRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	emailService := new(mocks.MockEmailService)

	uc := user.NewResendInvitation(endUserRepo, invitationRepo, eventRepo, emailService)

	err := uc.Execute(ctx, user.ResendInvitationInput{UserID: "", TenantID: "t1"})
	assert.Contains(t, err.Error(), "user_id is required")

	err = uc.Execute(ctx, user.ResendInvitationInput{UserID: "u1", TenantID: ""})
	assert.Contains(t, err.Error(), "tenant_id is required")
}
