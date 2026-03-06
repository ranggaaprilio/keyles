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
)

func TestListSessions_HappyPath(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	u := &entities.User{ID: "u1", TenantID: "t1"}
	tokens := []*entities.RefreshToken{
		{ID: 1, UserID: "u1", ClientID: "c1", ExpiresAt: time.Now().Add(1 * time.Hour), RevokedFlag: false},
		{ID: 2, UserID: "u1", ClientID: "c2", ExpiresAt: time.Now().Add(2 * time.Hour), RevokedFlag: false},
		{ID: 3, UserID: "u1", ClientID: "c1", ExpiresAt: time.Now().Add(-1 * time.Hour), RevokedFlag: false}, // expired
		{ID: 4, UserID: "u1", ClientID: "c2", ExpiresAt: time.Now().Add(1 * time.Hour), RevokedFlag: true},   // revoked
	}

	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)
	refreshTokenRepo.On("ListByUserID", ctx, "u1").Return(tokens, nil)

	uc := user.NewListSessions(endUserRepo, refreshTokenRepo)
	sessions, err := uc.Execute(ctx, user.ListSessionsInput{UserID: "u1", TenantID: "t1"})

	assert.NoError(t, err)
	assert.Len(t, sessions, 2) // only non-revoked, non-expired
}

func TestListSessions_TenantIsolation(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	u := &entities.User{ID: "u1", TenantID: "other-tenant"}
	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)

	uc := user.NewListSessions(endUserRepo, refreshTokenRepo)
	_, err := uc.Execute(ctx, user.ListSessionsInput{UserID: "u1", TenantID: "t1"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestListSessions_EmptyResult(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	u := &entities.User{ID: "u1", TenantID: "t1"}
	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)
	refreshTokenRepo.On("ListByUserID", ctx, "u1").Return([]*entities.RefreshToken{}, nil)

	uc := user.NewListSessions(endUserRepo, refreshTokenRepo)
	sessions, err := uc.Execute(ctx, user.ListSessionsInput{UserID: "u1", TenantID: "t1"})

	assert.NoError(t, err)
	assert.NotNil(t, sessions)
	assert.Len(t, sessions, 0)
}

func TestListSessions_UserNotFound(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	endUserRepo.On("GetByID", ctx, "u1").Return(nil, errors.New("not found"))

	uc := user.NewListSessions(endUserRepo, refreshTokenRepo)
	_, err := uc.Execute(ctx, user.ListSessionsInput{UserID: "u1", TenantID: "t1"})

	assert.Error(t, err)
}

func TestListSessions_EmptyUserID(t *testing.T) {
	endUserRepo := new(mocks.MockEndUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	uc := user.NewListSessions(endUserRepo, refreshTokenRepo)
	_, err := uc.Execute(context.Background(), user.ListSessionsInput{UserID: "", TenantID: "t1"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestListSessions_EmptyTenantID(t *testing.T) {
	endUserRepo := new(mocks.MockEndUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	uc := user.NewListSessions(endUserRepo, refreshTokenRepo)
	_, err := uc.Execute(context.Background(), user.ListSessionsInput{UserID: "u1", TenantID: ""})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant_id is required")
}

func TestListSessions_RepoError(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	u := &entities.User{ID: "u1", TenantID: "t1"}
	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)
	refreshTokenRepo.On("ListByUserID", ctx, "u1").Return(nil, errors.New("db error"))
	uc := user.NewListSessions(endUserRepo, refreshTokenRepo)
	_, err := uc.Execute(ctx, user.ListSessionsInput{UserID: "u1", TenantID: "t1"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list sessions")
}
