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

func TestListUserActivity_HappyPath(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	eventRepo := new(mocks.MockUserEventRepository)

	u := &entities.User{ID: "u1", TenantID: "t1"}
	events := []*entities.UserEvent{
		{TenantID: "t1", UserID: "u1", EventType: entities.EventTypeLoginSuccess},
		{TenantID: "t1", UserID: "u1", EventType: entities.EventTypeAccountEnabled},
	}

	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)
	eventRepo.On("ListByUser", ctx, "u1", 1, 25).Return(events, 2, nil)

	uc := user.NewListUserActivity(endUserRepo, eventRepo)
	out, err := uc.Execute(ctx, user.ListUserActivityInput{
UserID:   "u1",
TenantID: "t1",
})

	assert.NoError(t, err)
	assert.Len(t, out.Events, 2)
	assert.Equal(t, 2, out.Total)
	assert.Equal(t, 1, out.Page)
	assert.Equal(t, 25, out.PageSize)
	assert.Equal(t, 1, out.TotalPages)
}

func TestListUserActivity_CustomPagination(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	eventRepo := new(mocks.MockUserEventRepository)

	u := &entities.User{ID: "u1", TenantID: "t1"}
	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)
	eventRepo.On("ListByUser", ctx, "u1", 2, 10).Return([]*entities.UserEvent{}, 15, nil)

	uc := user.NewListUserActivity(endUserRepo, eventRepo)
	out, err := uc.Execute(ctx, user.ListUserActivityInput{
UserID:   "u1",
TenantID: "t1",
Page:     2,
PageSize: 10,
})

	assert.NoError(t, err)
	assert.Equal(t, 2, out.Page)
	assert.Equal(t, 10, out.PageSize)
	assert.Equal(t, 2, out.TotalPages) // ceil(15/10) = 2
}

func TestListUserActivity_PageSizeClampedTo100(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	eventRepo := new(mocks.MockUserEventRepository)

	u := &entities.User{ID: "u1", TenantID: "t1"}
	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)
	eventRepo.On("ListByUser", ctx, "u1", 1, 100).Return([]*entities.UserEvent{}, 0, nil)

	uc := user.NewListUserActivity(endUserRepo, eventRepo)
	out, err := uc.Execute(ctx, user.ListUserActivityInput{
UserID:   "u1",
TenantID: "t1",
PageSize: 500,
})

	assert.NoError(t, err)
	assert.Equal(t, 100, out.PageSize)
}

func TestListUserActivity_TenantIsolation(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	eventRepo := new(mocks.MockUserEventRepository)

	u := &entities.User{ID: "u1", TenantID: "other-tenant"}
	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)

	uc := user.NewListUserActivity(endUserRepo, eventRepo)
	_, err := uc.Execute(ctx, user.ListUserActivityInput{
UserID:   "u1",
TenantID: "t1",
})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestListUserActivity_UserNotFound(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	eventRepo := new(mocks.MockUserEventRepository)

	endUserRepo.On("GetByID", ctx, "u1").Return(nil, errors.New("not found"))

	uc := user.NewListUserActivity(endUserRepo, eventRepo)
	_, err := uc.Execute(ctx, user.ListUserActivityInput{
UserID:   "u1",
TenantID: "t1",
})

	assert.Error(t, err)
}

func TestListUserActivity_MissingFields(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	eventRepo := new(mocks.MockUserEventRepository)

	uc := user.NewListUserActivity(endUserRepo, eventRepo)

	_, err := uc.Execute(ctx, user.ListUserActivityInput{TenantID: "t1"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is required")

	_, err = uc.Execute(ctx, user.ListUserActivityInput{UserID: "u1"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant_id is required")
}
