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

func TestDeleteUser_HappyPath(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	blacklist := new(mocks.MockUserBlacklist)
	auditRepo := new(mocks.MockAuditRepository)

	targetID := "target-user"
	adminID := "admin-user"
	u := &entities.User{ID: targetID, TenantID: "t1", Email: "target@example.com"}

	endUserRepo.On("GetByID", ctx, targetID).Return(u, nil)
	roleRepo.On("RevokeAllForUser", ctx, targetID, adminID).Return(nil)
	refreshTokenRepo.On("RevokeByUserID", ctx, targetID).Return(nil)
	blacklist.On("Add", ctx, targetID, 900*time.Second).Return(nil)
	endUserRepo.On("Delete", ctx, targetID).Return(nil)
	auditRepo.On("Create", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	uc := user.NewDeleteUser(endUserRepo, roleRepo, refreshTokenRepo, blacklist, auditRepo)
	err := uc.Execute(ctx, user.DeleteUserInput{
TargetUserID: targetID,
AdminUserID:  adminID,
TenantID:     "t1",
})

	assert.NoError(t, err)
	endUserRepo.AssertCalled(t, "Delete", ctx, targetID)
	roleRepo.AssertCalled(t, "RevokeAllForUser", ctx, targetID, adminID)
	refreshTokenRepo.AssertCalled(t, "RevokeByUserID", ctx, targetID)
	blacklist.AssertCalled(t, "Add", ctx, targetID, 900*time.Second)
}

func TestDeleteUser_CannotDeleteSelf(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	blacklist := new(mocks.MockUserBlacklist)
	auditRepo := new(mocks.MockAuditRepository)

	uc := user.NewDeleteUser(endUserRepo, roleRepo, refreshTokenRepo, blacklist, auditRepo)
	err := uc.Execute(ctx, user.DeleteUserInput{
TargetUserID: "user1",
AdminUserID:  "user1",
TenantID:     "t1",
})

	assert.ErrorIs(t, err, user.ErrCannotDeleteSelf)
}

func TestDeleteUser_TenantIsolation(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	blacklist := new(mocks.MockUserBlacklist)
	auditRepo := new(mocks.MockAuditRepository)

	u := &entities.User{ID: "target", TenantID: "other-tenant"}
	endUserRepo.On("GetByID", ctx, "target").Return(u, nil)

	uc := user.NewDeleteUser(endUserRepo, roleRepo, refreshTokenRepo, blacklist, auditRepo)
	err := uc.Execute(ctx, user.DeleteUserInput{
TargetUserID: "target",
AdminUserID:  "admin1",
TenantID:     "t1",
})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestDeleteUser_UserNotFound(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	blacklist := new(mocks.MockUserBlacklist)
	auditRepo := new(mocks.MockAuditRepository)

	endUserRepo.On("GetByID", ctx, "target").Return(nil, errors.New("not found"))

	uc := user.NewDeleteUser(endUserRepo, roleRepo, refreshTokenRepo, blacklist, auditRepo)
	err := uc.Execute(ctx, user.DeleteUserInput{
TargetUserID: "target",
AdminUserID:  "admin1",
TenantID:     "t1",
})

	assert.Error(t, err)
}

func TestDeleteUser_DeleteFailure(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	blacklist := new(mocks.MockUserBlacklist)
	auditRepo := new(mocks.MockAuditRepository)

	targetID := "target-user"
	u := &entities.User{ID: targetID, TenantID: "t1", Email: "t@e.com"}

	endUserRepo.On("GetByID", ctx, targetID).Return(u, nil)
	roleRepo.On("RevokeAllForUser", ctx, targetID, "admin1").Return(nil)
	refreshTokenRepo.On("RevokeByUserID", ctx, targetID).Return(nil)
	blacklist.On("Add", ctx, targetID, 900*time.Second).Return(nil)
	endUserRepo.On("Delete", ctx, targetID).Return(errors.New("db error"))

	uc := user.NewDeleteUser(endUserRepo, roleRepo, refreshTokenRepo, blacklist, auditRepo)
	err := uc.Execute(ctx, user.DeleteUserInput{
TargetUserID: targetID,
AdminUserID:  "admin1",
TenantID:     "t1",
})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete user")
}

func TestDeleteUser_MissingFields(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	blacklist := new(mocks.MockUserBlacklist)
	auditRepo := new(mocks.MockAuditRepository)

	uc := user.NewDeleteUser(endUserRepo, roleRepo, refreshTokenRepo, blacklist, auditRepo)

	err := uc.Execute(ctx, user.DeleteUserInput{AdminUserID: "a", TenantID: "t"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target_user_id is required")

	err = uc.Execute(ctx, user.DeleteUserInput{TargetUserID: "u", TenantID: "t"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "admin_user_id is required")

	err = uc.Execute(ctx, user.DeleteUserInput{TargetUserID: "u", AdminUserID: "a"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant_id is required")
}
