package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/auth"
)

// TestRevokeToken_Success tests successful token revocation
func TestRevokeToken_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	validToken := &entities.RefreshToken{
		ID:          1,
		Token:       "hashed-token",
		UserID:      "user-123",
		ClientID:    "client-123",
		TenantID:    "tenant-123",
		RevokedFlag: false,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	mockRefreshTokenRepo.On("GetByToken", ctx, mock.AnythingOfType("string")).Return(validToken, nil)
	mockRefreshTokenRepo.On("Revoke", ctx, mock.AnythingOfType("string"), "client_request").Return(nil)

	uc := auth.NewRevokeToken(mockRefreshTokenRepo)

	// Act
	err := uc.Execute(ctx, auth.RevokeTokenRequest{
		Token:         "raw-refresh-token",
		TokenTypeHint: "refresh_token",
	})

	// Assert
	assert.NoError(t, err)
	mockRefreshTokenRepo.AssertExpectations(t)
}

// TestRevokeToken_TokenNotFound tests RFC 7009 compliant behavior for non-existent tokens
func TestRevokeToken_TokenNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	mockRefreshTokenRepo.On("GetByToken", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("token not found"))

	uc := auth.NewRevokeToken(mockRefreshTokenRepo)

	// Act - per RFC 7009, non-existent token should return success
	err := uc.Execute(ctx, auth.RevokeTokenRequest{
		Token:         "invalid-token",
		TokenTypeHint: "refresh_token",
	})

	// Assert - per RFC 7009, server responds with 200 OK even for invalid tokens
	assert.NoError(t, err)
	mockRefreshTokenRepo.AssertExpectations(t)
}

// TestRevokeToken_AlreadyRevoked tests idempotent behavior for already-revoked tokens
func TestRevokeToken_AlreadyRevoked(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	revokedToken := &entities.RefreshToken{
		ID:          1,
		Token:       "hashed-token",
		UserID:      "user-123",
		ClientID:    "client-123",
		TenantID:    "tenant-123",
		RevokedFlag: true,
		RevokedAt:   func() *time.Time { t := time.Now(); return &t }(),
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	mockRefreshTokenRepo.On("GetByToken", ctx, mock.AnythingOfType("string")).Return(revokedToken, nil)

	uc := auth.NewRevokeToken(mockRefreshTokenRepo)

	// Act
	err := uc.Execute(ctx, auth.RevokeTokenRequest{
		Token:         "already-revoked-token",
		TokenTypeHint: "refresh_token",
	})

	// Assert - should succeed without calling Revoke again (idempotent)
	assert.NoError(t, err)
	mockRefreshTokenRepo.AssertExpectations(t)
	mockRefreshTokenRepo.AssertNotCalled(t, "Revoke")
}

func TestRevokeToken_DifferentClientDoesNotRevoke(t *testing.T) {
	ctx := context.Background()
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	token := &entities.RefreshToken{
		Token:       "hashed-token",
		ClientID:    "owning-client",
		RevokedFlag: false,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	mockRefreshTokenRepo.On("GetByToken", ctx, mock.AnythingOfType("string")).Return(token, nil)

	uc := auth.NewRevokeToken(mockRefreshTokenRepo)
	err := uc.Execute(ctx, auth.RevokeTokenRequest{
		Token:         "raw-refresh-token",
		TokenTypeHint: "refresh_token",
		ClientID:      "different-client",
	})

	assert.NoError(t, err)
	mockRefreshTokenRepo.AssertExpectations(t)
	mockRefreshTokenRepo.AssertNotCalled(t, "Revoke")
}

// TestRevokeToken_MissingToken tests error handling for missing token parameter
func TestRevokeToken_MissingToken(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	uc := auth.NewRevokeToken(mockRefreshTokenRepo)

	// Act
	err := uc.Execute(ctx, auth.RevokeTokenRequest{
		Token: "",
	})

	// Assert - should return invalid_request error
	assert.Error(t, err)
	var oauthErr *auth.OAuthError
	assert.True(t, errors.As(err, &oauthErr))
	assert.Equal(t, auth.ErrInvalidRequest, oauthErr.Code)
}

// TestRevokeToken_RevokeAllForUser tests cascade revocation
func TestRevokeToken_RevokeAllForUser(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	mockRefreshTokenRepo.On("RevokeAllForUser", ctx, "user-123", "client-123").Return(nil)

	uc := auth.NewRevokeToken(mockRefreshTokenRepo)

	// Act
	err := uc.RevokeAllForUser(ctx, "user-123", "client-123")

	// Assert
	assert.NoError(t, err)
	mockRefreshTokenRepo.AssertExpectations(t)
}

// TestRevokeToken_RevokeAllForUser_Error tests error handling in cascade revocation
func TestRevokeToken_RevokeAllForUser_Error(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	mockRefreshTokenRepo.On("RevokeAllForUser", ctx, "user-123", "client-123").Return(errors.New("database error"))

	uc := auth.NewRevokeToken(mockRefreshTokenRepo)

	// Act
	err := uc.RevokeAllForUser(ctx, "user-123", "client-123")

	// Assert
	assert.Error(t, err)
	mockRefreshTokenRepo.AssertExpectations(t)
}

// TestRevokeToken_DatabaseError tests error handling during token revocation
func TestRevokeToken_DatabaseError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	validToken := &entities.RefreshToken{
		ID:          1,
		Token:       "hashed-token",
		UserID:      "user-123",
		ClientID:    "client-123",
		TenantID:    "tenant-123",
		RevokedFlag: false,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	mockRefreshTokenRepo.On("GetByToken", ctx, mock.AnythingOfType("string")).Return(validToken, nil)
	mockRefreshTokenRepo.On("Revoke", ctx, mock.AnythingOfType("string"), "client_request").Return(errors.New("database error"))

	uc := auth.NewRevokeToken(mockRefreshTokenRepo)

	// Act
	err := uc.Execute(ctx, auth.RevokeTokenRequest{
		Token:         "valid-token",
		TokenTypeHint: "refresh_token",
	})

	// Assert - should return server_error
	assert.Error(t, err)
	var oauthErr *auth.OAuthError
	assert.True(t, errors.As(err, &oauthErr))
	assert.Equal(t, auth.ErrServerError, oauthErr.Code)
}

// TestRevokeToken_AccessTokenHint tests handling of access_token hint
func TestRevokeToken_AccessTokenHint(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	uc := auth.NewRevokeToken(mockRefreshTokenRepo)

	// Act - access tokens are JWTs, cannot be revoked
	err := uc.Execute(ctx, auth.RevokeTokenRequest{
		Token:         "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
		TokenTypeHint: "access_token",
	})

	// Assert - should succeed (but not actually revoke anything)
	assert.NoError(t, err)
	mockRefreshTokenRepo.AssertNotCalled(t, "GetByToken")
}

// TestRevokeToken_NoHint tests default behavior when no hint is provided
func TestRevokeToken_NoHint(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	validToken := &entities.RefreshToken{
		ID:          1,
		Token:       "hashed-token",
		UserID:      "user-123",
		ClientID:    "client-123",
		TenantID:    "tenant-123",
		RevokedFlag: false,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	mockRefreshTokenRepo.On("GetByToken", ctx, mock.AnythingOfType("string")).Return(validToken, nil)
	mockRefreshTokenRepo.On("Revoke", ctx, mock.AnythingOfType("string"), "client_request").Return(nil)

	uc := auth.NewRevokeToken(mockRefreshTokenRepo)

	// Act - should default to treating as refresh token
	err := uc.Execute(ctx, auth.RevokeTokenRequest{
		Token: "raw-refresh-token",
	})

	// Assert
	assert.NoError(t, err)
	mockRefreshTokenRepo.AssertExpectations(t)
}
