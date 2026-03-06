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

func TestInviteUser_QuotaExceeded_CacheHit(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	invitationRepo := new(mocks.MockInvitationRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	emailService := new(mocks.MockEmailService)
	countCache := new(mocks.MockUserCountCache)

	// Cache returns quota-exceeded count (hit=true)
	countCache.On("Get", ctx, "tenant-1").Return(10000, true, nil)

	uc := user.NewInviteUser(endUserRepo, invitationRepo, eventRepo, emailService, countCache)
	_, err := uc.Execute(ctx, user.InviteUserInput{
TenantID:  "tenant-1",
Email:     "new@example.com",
InvitedBy: "admin-1",
})

	assert.ErrorIs(t, err, user.ErrQuotaExceeded)
	countCache.AssertExpectations(t)
}

func TestInviteUser_QuotaExceeded_CacheMiss_DBFallback(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	invitationRepo := new(mocks.MockInvitationRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	emailService := new(mocks.MockEmailService)
	countCache := new(mocks.MockUserCountCache)

	// Cache miss, DB returns quota-exceeded count
	countCache.On("Get", ctx, "tenant-1").Return(0, false, nil)
	endUserRepo.On("CountByTenant", ctx, "tenant-1").Return(10000, nil)
	countCache.On("Set", ctx, "tenant-1", 10000).Return(nil)

	uc := user.NewInviteUser(endUserRepo, invitationRepo, eventRepo, emailService, countCache)
	_, err := uc.Execute(ctx, user.InviteUserInput{
TenantID:  "tenant-1",
Email:     "new@example.com",
InvitedBy: "admin-1",
})

	assert.ErrorIs(t, err, user.ErrQuotaExceeded)
	endUserRepo.AssertExpectations(t)
	countCache.AssertExpectations(t)
}

func TestInviteUser_EmailAlreadyExists(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	invitationRepo := new(mocks.MockInvitationRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	emailService := new(mocks.MockEmailService)
	countCache := new(mocks.MockUserCountCache)

	countCache.On("Get", ctx, "tenant-1").Return(5, true, nil)
	endUserRepo.On("GetByEmail", ctx, "tenant-1", "existing@example.com").Return(&entities.User{
		ID:       "user-1",
		TenantID: "tenant-1",
		Email:    "existing@example.com",
		Status:   entities.UserStatusActive,
	}, nil)

	uc := user.NewInviteUser(endUserRepo, invitationRepo, eventRepo, emailService, countCache)
	_, err := uc.Execute(ctx, user.InviteUserInput{
TenantID:  "tenant-1",
Email:     "existing@example.com",
InvitedBy: "admin-1",
})

	assert.ErrorIs(t, err, user.ErrEmailAlreadyExists)
}

func TestInviteUser_PendingInvitationExists(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	invitationRepo := new(mocks.MockInvitationRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	emailService := new(mocks.MockEmailService)
	countCache := new(mocks.MockUserCountCache)

	countCache.On("Get", ctx, "tenant-1").Return(5, true, nil)
	endUserRepo.On("GetByEmail", ctx, "tenant-1", "new@example.com").Return(nil, errors.New("not found"))
	invitationRepo.On("GetPendingByEmail", ctx, "tenant-1", "new@example.com").Return(&entities.Invitation{
		ID:       "inv-1",
		TenantID: "tenant-1",
		Email:    "new@example.com",
		Status:   entities.InvitationStatusPending,
	}, nil)

	uc := user.NewInviteUser(endUserRepo, invitationRepo, eventRepo, emailService, countCache)
	_, err := uc.Execute(ctx, user.InviteUserInput{
TenantID:  "tenant-1",
Email:     "new@example.com",
InvitedBy: "admin-1",
})

	assert.ErrorIs(t, err, user.ErrPendingInvitationExists)
}

func TestInviteUser_HappyPath(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	invitationRepo := new(mocks.MockInvitationRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	emailService := new(mocks.MockEmailService)
	countCache := new(mocks.MockUserCountCache)

	countCache.On("Get", ctx, "tenant-1").Return(5, true, nil)
	endUserRepo.On("GetByEmail", ctx, "tenant-1", "new@example.com").Return(nil, errors.New("not found"))
	invitationRepo.On("GetPendingByEmail", ctx, "tenant-1", "new@example.com").Return(nil, errors.New("not found"))

	endUserRepo.On("Create", ctx, mock.MatchedBy(func(u *entities.User) bool {
return u.TenantID == "tenant-1" && u.Email == "new@example.com" && u.Status == entities.UserStatusPending
	})).Return(nil)

	invitationRepo.On("Create", ctx, mock.MatchedBy(func(inv *entities.Invitation) bool {
return inv.TenantID == "tenant-1" && inv.Email == "new@example.com" && inv.Status == entities.InvitationStatusPending && inv.InvitedBy == "admin-1" && inv.TokenHash != ""
	})).Return(nil)

	emailService.On("SendInvitationEmail", ctx, "new@example.com", "New User", mock.AnythingOfType("string"), "tenant-1").Return(nil)

	eventRepo.On("Record", ctx, mock.MatchedBy(func(e *entities.UserEvent) bool {
return e.EventType == entities.EventTypeUserInvited && e.TenantID == "tenant-1"
	})).Return(nil)

	countCache.On("Invalidate", ctx, "tenant-1").Return(nil)

	uc := user.NewInviteUser(endUserRepo, invitationRepo, eventRepo, emailService, countCache)
	output, err := uc.Execute(ctx, user.InviteUserInput{
TenantID:      "tenant-1",
Email:         "new@example.com",
DisplayName:   "New User",
InvitedBy:     "admin-1",
InviteBaseURL: "https://auth.example.com/invite",
})

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "new@example.com", output.Email)
	assert.Equal(t, "New User", output.DisplayName)
	assert.Equal(t, entities.UserStatusPending, output.Status)
	assert.NotEmpty(t, output.UserID)
	assert.NotEmpty(t, output.InvitationID)

	endUserRepo.AssertExpectations(t)
	invitationRepo.AssertExpectations(t)
	emailService.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
	countCache.AssertExpectations(t)
}

func TestInviteUser_MissingFields(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	invitationRepo := new(mocks.MockInvitationRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	emailService := new(mocks.MockEmailService)
	countCache := new(mocks.MockUserCountCache)

	uc := user.NewInviteUser(endUserRepo, invitationRepo, eventRepo, emailService, countCache)

	_, err := uc.Execute(ctx, user.InviteUserInput{TenantID: "", Email: "a@b.com", InvitedBy: "admin"})
	assert.Contains(t, err.Error(), "tenant_id is required")

	_, err = uc.Execute(ctx, user.InviteUserInput{TenantID: "t1", Email: "", InvitedBy: "admin"})
	assert.Contains(t, err.Error(), "email is required")

	_, err = uc.Execute(ctx, user.InviteUserInput{TenantID: "t1", Email: "a@b.com", InvitedBy: ""})
	assert.Contains(t, err.Error(), "invited_by is required")
}
