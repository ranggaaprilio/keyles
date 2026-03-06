package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateUser_HappyPath(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	auditRepo := new(mocks.MockAuditRepository)

	existing := &entities.User{ID: "u1", TenantID: "t1", Email: "a@b.com", DisplayName: "Old Name"}
	endUserRepo.On("GetByID", ctx, "u1").Return(existing, nil)
	endUserRepo.On("Update", ctx, mock.MatchedBy(func(u *entities.User) bool {
return u.ID == "u1" && u.DisplayName == "New Name"
	})).Return(nil)
	auditRepo.On("Create", ctx, mock.Anything).Return(nil)

	uc := user.NewUpdateUser(endUserRepo, auditRepo)
	newName := "New Name"
	result, err := uc.Execute(ctx, user.UpdateUserInput{UserID: "u1", TenantID: "t1", DisplayName: &newName})

	assert.NoError(t, err)
	assert.Equal(t, "New Name", result.DisplayName)
}

func TestUpdateUser_CrossTenantRejected(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	auditRepo := new(mocks.MockAuditRepository)

	existing := &entities.User{ID: "u1", TenantID: "other-tenant"}
	endUserRepo.On("GetByID", ctx, "u1").Return(existing, nil)

	uc := user.NewUpdateUser(endUserRepo, auditRepo)
	newName := "Name"
	_, err := uc.Execute(ctx, user.UpdateUserInput{UserID: "u1", TenantID: "t1", DisplayName: &newName})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestUpdateUser_DisplayNameTooLong(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	auditRepo := new(mocks.MockAuditRepository)

	uc := user.NewUpdateUser(endUserRepo, auditRepo)
	longName := strings.Repeat("x", 256)
	_, err := uc.Execute(ctx, user.UpdateUserInput{UserID: "u1", TenantID: "t1", DisplayName: &longName})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "display name must not exceed 255")
}

func TestUpdateUser_EmptyDisplayNameAllowed(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	auditRepo := new(mocks.MockAuditRepository)

	existing := &entities.User{ID: "u1", TenantID: "t1", DisplayName: "Old"}
	endUserRepo.On("GetByID", ctx, "u1").Return(existing, nil)
	endUserRepo.On("Update", ctx, mock.MatchedBy(func(u *entities.User) bool {
return u.DisplayName == ""
})).Return(nil)
	auditRepo.On("Create", ctx, mock.Anything).Return(nil)

	uc := user.NewUpdateUser(endUserRepo, auditRepo)
	empty := ""
	result, err := uc.Execute(ctx, user.UpdateUserInput{UserID: "u1", TenantID: "t1", DisplayName: &empty})

	assert.NoError(t, err)
	assert.Equal(t, "", result.DisplayName)
}

func TestUpdateUser_NilDisplayName(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	auditRepo := new(mocks.MockAuditRepository)

	existing := &entities.User{ID: "u1", TenantID: "t1", DisplayName: "Unchanged"}
	endUserRepo.On("GetByID", ctx, "u1").Return(existing, nil)
	endUserRepo.On("Update", ctx, mock.MatchedBy(func(u *entities.User) bool {
		return u.DisplayName == "Unchanged"
	})).Return(nil)
	auditRepo.On("Create", ctx, mock.Anything).Return(nil)

	uc := user.NewUpdateUser(endUserRepo, auditRepo)
	result, err := uc.Execute(ctx, user.UpdateUserInput{UserID: "u1", TenantID: "t1", DisplayName: nil})

	assert.NoError(t, err)
	assert.Equal(t, "Unchanged", result.DisplayName)
}

func TestUpdateUser_RepoUpdateError(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	auditRepo := new(mocks.MockAuditRepository)

	existing := &entities.User{ID: "u1", TenantID: "t1", DisplayName: "Old"}
	endUserRepo.On("GetByID", ctx, "u1").Return(existing, nil)
	endUserRepo.On("Update", ctx, mock.Anything).Return(errors.New("db error"))

	uc := user.NewUpdateUser(endUserRepo, auditRepo)
	newName := "New"
	_, err := uc.Execute(ctx, user.UpdateUserInput{UserID: "u1", TenantID: "t1", DisplayName: &newName})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update user")
}

func TestUpdateUser_UserNotFound(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	auditRepo := new(mocks.MockAuditRepository)

	endUserRepo.On("GetByID", ctx, "u1").Return(nil, errors.New("not found"))

	uc := user.NewUpdateUser(endUserRepo, auditRepo)
	newName := "Name"
	_, err := uc.Execute(ctx, user.UpdateUserInput{UserID: "u1", TenantID: "t1", DisplayName: &newName})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}
