package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/user"
	"github.com/stretchr/testify/assert"
)

func TestGetUser_HappyPath(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)

	u := &entities.User{ID: "u1", TenantID: "t1", Email: "a@b.com", Status: entities.UserStatusActive}
	roles := []*entities.UserRoleAssignment{
		{ID: 1, UserID: "u1", ClientID: "c1", Role: "editor", IsActive: true},
	}

	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)
	roleRepo.On("ListByUser", ctx, "u1").Return(roles, nil)

	uc := user.NewGetUser(endUserRepo, roleRepo)
	output, err := uc.Execute(ctx, "u1", "t1")

	assert.NoError(t, err)
	assert.Equal(t, "u1", output.User.ID)
	assert.Len(t, output.RoleAssignments, 1)
}

func TestGetUser_CrossTenantRejected(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)

	u := &entities.User{ID: "u1", TenantID: "other-tenant", Email: "a@b.com"}
	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)

	uc := user.NewGetUser(endUserRepo, roleRepo)
	_, err := uc.Execute(ctx, "u1", "t1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestGetUser_NotFound(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)

	endUserRepo.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	uc := user.NewGetUser(endUserRepo, roleRepo)
	_, err := uc.Execute(ctx, "nonexistent", "t1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestGetUser_RolesErrorDefaultsToEmpty(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)

	u := &entities.User{ID: "u1", TenantID: "t1", Email: "a@b.com"}
	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)
	roleRepo.On("ListByUser", ctx, "u1").Return(nil, errors.New("db error"))

	uc := user.NewGetUser(endUserRepo, roleRepo)
	output, err := uc.Execute(ctx, "u1", "t1")

	assert.NoError(t, err)
	assert.Empty(t, output.RoleAssignments)
}

func TestGetUser_EmptyUserID(t *testing.T) {
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	uc := user.NewGetUser(endUserRepo, roleRepo)
	_, err := uc.Execute(context.Background(), "", "t1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestGetUser_EmptyTenantID(t *testing.T) {
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	uc := user.NewGetUser(endUserRepo, roleRepo)
	_, err := uc.Execute(context.Background(), "u1", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant_id is required")
}
