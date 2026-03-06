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

func TestEnableUser_HappyPath(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	auditRepo := new(mocks.MockAuditRepository)

	u := &entities.User{ID: "u1", TenantID: "t1", Status: entities.UserStatusDisabled}
	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)
	endUserRepo.On("UpdateStatus", ctx, "u1", entities.UserStatusActive).Return(nil)
	eventRepo.On("Record", ctx, mock.AnythingOfType("*entities.UserEvent")).Return(nil)
	auditRepo.On("Create", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	uc := user.NewEnableUser(endUserRepo, eventRepo, auditRepo)
	err := uc.Execute(ctx, user.EnableUserInput{
TargetUserID: "u1",
AdminUserID:  "admin1",
TenantID:     "t1",
})

	assert.NoError(t, err)
	endUserRepo.AssertCalled(t, "UpdateStatus", ctx, "u1", entities.UserStatusActive)
}

func TestEnableUser_AlreadyActive(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	auditRepo := new(mocks.MockAuditRepository)

	u := &entities.User{ID: "u1", TenantID: "t1", Status: entities.UserStatusActive}
	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)

	uc := user.NewEnableUser(endUserRepo, eventRepo, auditRepo)
	err := uc.Execute(ctx, user.EnableUserInput{
TargetUserID: "u1",
AdminUserID:  "admin1",
TenantID:     "t1",
})

	assert.NoError(t, err) // idempotent
	endUserRepo.AssertNotCalled(t, "UpdateStatus")
}

func TestEnableUser_PendingUser(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	auditRepo := new(mocks.MockAuditRepository)

	u := &entities.User{ID: "u1", TenantID: "t1", Status: entities.UserStatusPending}
	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)

	uc := user.NewEnableUser(endUserRepo, eventRepo, auditRepo)
	err := uc.Execute(ctx, user.EnableUserInput{
TargetUserID: "u1",
AdminUserID:  "admin1",
TenantID:     "t1",
})

	assert.ErrorIs(t, err, user.ErrCanOnlyEnableDisabled)
}

func TestEnableUser_TenantIsolation(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	auditRepo := new(mocks.MockAuditRepository)

	u := &entities.User{ID: "u1", TenantID: "other-tenant"}
	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)

	uc := user.NewEnableUser(endUserRepo, eventRepo, auditRepo)
	err := uc.Execute(ctx, user.EnableUserInput{
TargetUserID: "u1",
AdminUserID:  "admin1",
TenantID:     "t1",
})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestEnableUser_UserNotFound(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	auditRepo := new(mocks.MockAuditRepository)

	endUserRepo.On("GetByID", ctx, "u1").Return(nil, errors.New("not found"))

	uc := user.NewEnableUser(endUserRepo, eventRepo, auditRepo)
	err := uc.Execute(ctx, user.EnableUserInput{
TargetUserID: "u1",
AdminUserID:  "admin1",
TenantID:     "t1",
})

	assert.Error(t, err)
}

func TestEnableUser_MissingFields(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	eventRepo := new(mocks.MockUserEventRepository)
	auditRepo := new(mocks.MockAuditRepository)

	uc := user.NewEnableUser(endUserRepo, eventRepo, auditRepo)

	err := uc.Execute(ctx, user.EnableUserInput{AdminUserID: "a", TenantID: "t"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target_user_id is required")

	err = uc.Execute(ctx, user.EnableUserInput{TargetUserID: "u"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant_id is required")
}
