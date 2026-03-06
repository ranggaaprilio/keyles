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

func TestRevokeSession_HappyPath(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	eventRepo := new(mocks.MockUserEventRepository)

	u := &entities.User{ID: "u1", TenantID: "t1"}
	token := &entities.RefreshToken{
		ID:          1,
		Token:       "hashed-token",
		UserID:      "u1",
		ClientID:    "c1",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
		RevokedFlag: false,
	}

	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)
	refreshTokenRepo.On("GetByID", ctx, int64(1)).Return(token, nil)
	refreshTokenRepo.On("Revoke", ctx, "hashed-token", "admin1").Return(nil)
	eventRepo.On("Record", ctx, mock.AnythingOfType("*entities.UserEvent")).Return(nil)

	uc := user.NewRevokeSession(endUserRepo, refreshTokenRepo, eventRepo)
	err := uc.Execute(ctx, user.RevokeSessionInput{
UserID:    "u1",
TenantID:  "t1",
TokenID:   1,
RevokedBy: "admin1",
})

	assert.NoError(t, err)
	refreshTokenRepo.AssertCalled(t, "Revoke", ctx, "hashed-token", "admin1")
	eventRepo.AssertCalled(t, "Record", ctx, mock.AnythingOfType("*entities.UserEvent"))
}

func TestRevokeSession_TenantIsolation(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	eventRepo := new(mocks.MockUserEventRepository)

	u := &entities.User{ID: "u1", TenantID: "other-tenant"}
	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)

	uc := user.NewRevokeSession(endUserRepo, refreshTokenRepo, eventRepo)
	err := uc.Execute(ctx, user.RevokeSessionInput{
UserID:    "u1",
TenantID:  "t1",
TokenID:   1,
RevokedBy: "admin1",
})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestRevokeSession_AlreadyRevoked(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	eventRepo := new(mocks.MockUserEventRepository)

	u := &entities.User{ID: "u1", TenantID: "t1"}
	token := &entities.RefreshToken{
		ID:          1,
		Token:       "hashed-token",
		UserID:      "u1",
		ClientID:    "c1",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
		RevokedFlag: true, // already revoked
	}

	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)
	refreshTokenRepo.On("GetByID", ctx, int64(1)).Return(token, nil)

	uc := user.NewRevokeSession(endUserRepo, refreshTokenRepo, eventRepo)
	err := uc.Execute(ctx, user.RevokeSessionInput{
UserID:    "u1",
TenantID:  "t1",
TokenID:   1,
RevokedBy: "admin1",
})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already revoked")
}

func TestRevokeSession_SessionBelongsToDifferentUser(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	eventRepo := new(mocks.MockUserEventRepository)

	u := &entities.User{ID: "u1", TenantID: "t1"}
	token := &entities.RefreshToken{
		ID:          1,
		Token:       "hashed-token",
		UserID:      "u2", // belongs to different user
		ClientID:    "c1",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
		RevokedFlag: false,
	}

	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)
	refreshTokenRepo.On("GetByID", ctx, int64(1)).Return(token, nil)

	uc := user.NewRevokeSession(endUserRepo, refreshTokenRepo, eventRepo)
	err := uc.Execute(ctx, user.RevokeSessionInput{
UserID:    "u1",
TenantID:  "t1",
TokenID:   1,
RevokedBy: "admin1",
})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestRevokeSession_MissingFields(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	eventRepo := new(mocks.MockUserEventRepository)

	uc := user.NewRevokeSession(endUserRepo, refreshTokenRepo, eventRepo)

	err := uc.Execute(ctx, user.RevokeSessionInput{TenantID: "t1", TokenID: 1, RevokedBy: "a"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is required")

	err = uc.Execute(ctx, user.RevokeSessionInput{UserID: "u1", TokenID: 1, RevokedBy: "a"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant_id is required")

	err = uc.Execute(ctx, user.RevokeSessionInput{UserID: "u1", TenantID: "t1", RevokedBy: "a"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token_id is required")
}

func TestRevokeSession_TokenNotFound(t *testing.T) {
	ctx := context.Background()
	endUserRepo := new(mocks.MockEndUserRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	eventRepo := new(mocks.MockUserEventRepository)

	u := &entities.User{ID: "u1", TenantID: "t1"}
	endUserRepo.On("GetByID", ctx, "u1").Return(u, nil)
	refreshTokenRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

	uc := user.NewRevokeSession(endUserRepo, refreshTokenRepo, eventRepo)
	err := uc.Execute(ctx, user.RevokeSessionInput{
UserID:    "u1",
TenantID:  "t1",
TokenID:   999,
RevokedBy: "admin1",
})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}
