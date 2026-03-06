package usecase_test

import (
"context"
"testing"
"time"

"github.com/google/uuid"
"github.com/ranggaaprilio/keyles/domain/entities"
"github.com/ranggaaprilio/keyles/tests/mocks"
"github.com/ranggaaprilio/keyles/usecase/user"
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/mock"
)

func TestDisableUser_HappyPath(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	userRepo := new(mocks.MockUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	blacklist := new(mocks.MockUserBlacklist)
	eventRepo := new(mocks.MockUserEventRepository)
	auditRepo := new(mocks.MockAuditRepository)

	targetID := "target-user-id"
	adminID := "admin-user-id"

	// uuid.Parse fails for non-UUID strings so admin check is skipped
	u := &entities.User{ID: targetID, TenantID: "t1", Status: entities.UserStatusActive}
	endUserRepo.On("GetByID", ctx, targetID).Return(u, nil)
	endUserRepo.On("UpdateStatus", ctx, targetID, entities.UserStatusDisabled).Return(nil)
	refreshTokenRepo.On("RevokeByUserID", ctx, targetID).Return(nil)
	blacklist.On("Add", ctx, targetID, 900*time.Second).Return(nil)
	eventRepo.On("Record", ctx, mock.AnythingOfType("*entities.UserEvent")).Return(nil)
	auditRepo.On("Create", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	uc := user.NewDisableUser(endUserRepo, userRepo, refreshTokenRepo, blacklist, eventRepo, auditRepo)
	err := uc.Execute(ctx, user.DisableUserInput{
TargetUserID: targetID,
AdminUserID:  adminID,
TenantID:     "t1",
})

	assert.NoError(t, err)
	endUserRepo.AssertCalled(t, "UpdateStatus", ctx, targetID, entities.UserStatusDisabled)
	refreshTokenRepo.AssertCalled(t, "RevokeByUserID", ctx, targetID)
	blacklist.AssertCalled(t, "Add", ctx, targetID, 900*time.Second)
}

func TestDisableUser_CannotDisableSelf(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	userRepo := new(mocks.MockUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	blacklist := new(mocks.MockUserBlacklist)
	eventRepo := new(mocks.MockUserEventRepository)
	auditRepo := new(mocks.MockAuditRepository)

	uc := user.NewDisableUser(endUserRepo, userRepo, refreshTokenRepo, blacklist, eventRepo, auditRepo)
	err := uc.Execute(ctx, user.DisableUserInput{
TargetUserID: "user1",
AdminUserID:  "user1",
TenantID:     "t1",
})

	assert.ErrorIs(t, err, user.ErrCannotDisableSelf)
}

func TestDisableUser_CannotDisableAdmin(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	userRepo := new(mocks.MockUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	blacklist := new(mocks.MockUserBlacklist)
	eventRepo := new(mocks.MockUserEventRepository)
	auditRepo := new(mocks.MockAuditRepository)

	targetUUID := uuid.New()
	adminUser := &entities.AdminUser{
		ID:   targetUUID,
		Role: entities.UserRoleAdmin,
	}
	userRepo.On("FindByID", ctx, targetUUID).Return(adminUser, nil)

	uc := user.NewDisableUser(endUserRepo, userRepo, refreshTokenRepo, blacklist, eventRepo, auditRepo)
	err := uc.Execute(ctx, user.DisableUserInput{
TargetUserID: targetUUID.String(),
		AdminUserID:  "other-admin",
		TenantID:     "t1",
	})

	assert.ErrorIs(t, err, user.ErrCannotDisableAdmin)
}

func TestDisableUser_CannotDisableOwner(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	userRepo := new(mocks.MockUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	blacklist := new(mocks.MockUserBlacklist)
	eventRepo := new(mocks.MockUserEventRepository)
	auditRepo := new(mocks.MockAuditRepository)

	targetUUID := uuid.New()
	adminUser := &entities.AdminUser{
		ID:   targetUUID,
		Role: entities.UserRoleOwner,
	}
	userRepo.On("FindByID", ctx, targetUUID).Return(adminUser, nil)

	uc := user.NewDisableUser(endUserRepo, userRepo, refreshTokenRepo, blacklist, eventRepo, auditRepo)
	err := uc.Execute(ctx, user.DisableUserInput{
TargetUserID: targetUUID.String(),
		AdminUserID:  "other-admin",
		TenantID:     "t1",
	})

	assert.ErrorIs(t, err, user.ErrCannotDisableAdmin)
}

func TestDisableUser_TenantIsolation(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	userRepo := new(mocks.MockUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	blacklist := new(mocks.MockUserBlacklist)
	eventRepo := new(mocks.MockUserEventRepository)
	auditRepo := new(mocks.MockAuditRepository)

	u := &entities.User{ID: "target", TenantID: "other-tenant"}
	endUserRepo.On("GetByID", ctx, "target").Return(u, nil)

	uc := user.NewDisableUser(endUserRepo, userRepo, refreshTokenRepo, blacklist, eventRepo, auditRepo)
	err := uc.Execute(ctx, user.DisableUserInput{
TargetUserID: "target",
AdminUserID:  "admin1",
TenantID:     "t1",
})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestDisableUser_MissingFields(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	userRepo := new(mocks.MockUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	blacklist := new(mocks.MockUserBlacklist)
	eventRepo := new(mocks.MockUserEventRepository)
	auditRepo := new(mocks.MockAuditRepository)

	uc := user.NewDisableUser(endUserRepo, userRepo, refreshTokenRepo, blacklist, eventRepo, auditRepo)

	err := uc.Execute(ctx, user.DisableUserInput{AdminUserID: "a", TenantID: "t"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target_user_id is required")

	err = uc.Execute(ctx, user.DisableUserInput{TargetUserID: "u", TenantID: "t"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "admin_user_id is required")

	err = uc.Execute(ctx, user.DisableUserInput{TargetUserID: "u", AdminUserID: "a"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant_id is required")
}
