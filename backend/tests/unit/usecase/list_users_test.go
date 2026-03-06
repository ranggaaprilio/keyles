package usecase_test

import (
"context"
"testing"

"github.com/ranggaaprilio/keyles/domain/entities"
"github.com/ranggaaprilio/keyles/tests/mocks"
"github.com/ranggaaprilio/keyles/usecase/user"
"github.com/stretchr/testify/assert"
)

func TestListUsers_DefaultPagination(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)

	users := []*entities.User{{ID: "u1", TenantID: "t1", Email: "a@b.com"}}
	endUserRepo.On("ListByTenant", ctx, "t1", "", entities.UserStatus(""), 1, 25).Return(users, 1, nil)

	uc := user.NewListUsers(endUserRepo)
	output, err := uc.Execute(ctx, user.ListUsersInput{TenantID: "t1"})

	assert.NoError(t, err)
	assert.Equal(t, 1, output.Total)
	assert.Equal(t, 1, output.Page)
	assert.Equal(t, 25, output.PageSize)
	assert.Len(t, output.Users, 1)
}

func TestListUsers_CustomPageSize(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)

	endUserRepo.On("ListByTenant", ctx, "t1", "", entities.UserStatus(""), 2, 10).Return([]*entities.User{}, 0, nil)

	uc := user.NewListUsers(endUserRepo)
	output, err := uc.Execute(ctx, user.ListUsersInput{TenantID: "t1", Page: 2, PageSize: 10})

	assert.NoError(t, err)
	assert.Equal(t, 2, output.Page)
	assert.Equal(t, 10, output.PageSize)
}

func TestListUsers_PageSizeClampedTo100(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)

	endUserRepo.On("ListByTenant", ctx, "t1", "", entities.UserStatus(""), 1, 100).Return([]*entities.User{}, 0, nil)

	uc := user.NewListUsers(endUserRepo)
	output, err := uc.Execute(ctx, user.ListUsersInput{TenantID: "t1", PageSize: 200})

	assert.NoError(t, err)
	assert.Equal(t, 100, output.PageSize)
}

func TestListUsers_SearchFilter(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)

	endUserRepo.On("ListByTenant", ctx, "t1", "alice", entities.UserStatus(""), 1, 25).Return([]*entities.User{}, 0, nil)

	uc := user.NewListUsers(endUserRepo)
	_, err := uc.Execute(ctx, user.ListUsersInput{TenantID: "t1", Search: "alice"})

	assert.NoError(t, err)
	endUserRepo.AssertExpectations(t)
}

func TestListUsers_StatusFilter(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)

	endUserRepo.On("ListByTenant", ctx, "t1", "", entities.UserStatusPending, 1, 25).Return([]*entities.User{}, 0, nil)

	uc := user.NewListUsers(endUserRepo)
	_, err := uc.Execute(ctx, user.ListUsersInput{TenantID: "t1", StatusFilter: entities.UserStatusPending})

	assert.NoError(t, err)
	endUserRepo.AssertExpectations(t)
}

func TestListUsers_MissingTenantID(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)

	uc := user.NewListUsers(endUserRepo)
	_, err := uc.Execute(ctx, user.ListUsersInput{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant_id is required")
}

func TestListUsers_TotalPages(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)

	endUserRepo.On("ListByTenant", ctx, "t1", "", entities.UserStatus(""), 1, 25).Return([]*entities.User{}, 51, nil)

	uc := user.NewListUsers(endUserRepo)
	output, err := uc.Execute(ctx, user.ListUsersInput{TenantID: "t1"})

	assert.NoError(t, err)
	assert.Equal(t, 3, output.TotalPages) // ceil(51/25) = 3
}
