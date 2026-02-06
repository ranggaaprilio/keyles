package usecase_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/services"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// hashTokenForRefresh hashes a token for testing
func hashTokenForRefresh(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// TestRefreshToken_Success tests successful token refresh
func TestRefreshToken_Success(t *testing.T) {
	// Setup mocks
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockClientRepo := new(MockClientRepositoryForToken)
	mockTokenService := new(MockTokenService)

	ctx := context.Background()
	issuer := "https://auth.example.com"

	// Test data
	rawRefreshToken := "test-refresh-token-12345"
	tokenHash := hashTokenForRefresh(rawRefreshToken)
	clientID := "client-123"
	clientSecret := "secret-456"
	userID := "user-789"
	tenantID := "tenant-abc"
	scope := "openid profile email"

	// Create stored refresh token
	storedToken := &entities.RefreshToken{
		ID:        1,
		Token:     tokenHash,
		UserID:    userID,
		ClientID:  clientID,
		TenantID:  tenantID,
		Scope:     scope,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}

	client := &entities.Client{
		ClientID:   clientID,
		TenantID:   tenantID,
		ClientName: "Test Client",
		IsActive:   true,
	}

	// Setup mock expectations
	mockRefreshTokenRepo.On("GetByToken", ctx, tokenHash).Return(storedToken, nil)
	mockRefreshTokenRepo.On("UpdateLastUsed", ctx, tokenHash).Return(nil)
	mockClientRepo.On("ValidateCredentials", ctx, clientID, clientSecret).Return(client, nil)
	mockTokenService.On("SignAccessToken", ctx, mock.AnythingOfType("*services.TokenClaims")).Return("new-access-token", nil)

	// Create use case
	uc := auth.NewRefreshToken(mockRefreshTokenRepo, mockClientRepo, mockTokenService, issuer)

	// Execute
	req := auth.RefreshTokenRequest{
		RefreshToken: rawRefreshToken,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
	resp, err := uc.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "new-access-token", resp.AccessToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, 900, resp.ExpiresIn) // 15 minutes in seconds
	assert.Equal(t, scope, resp.Scope)

	mockRefreshTokenRepo.AssertExpectations(t)
	mockClientRepo.AssertExpectations(t)
	mockTokenService.AssertExpectations(t)
}

// TestRefreshToken_TokenNotFound tests refresh with non-existent token
func TestRefreshToken_TokenNotFound(t *testing.T) {
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockClientRepo := new(MockClientRepositoryForToken)
	mockTokenService := new(MockTokenService)

	ctx := context.Background()
	issuer := "https://auth.example.com"

	rawRefreshToken := "invalid-token"
	tokenHash := hashTokenForRefresh(rawRefreshToken)

	// Token not found
	mockRefreshTokenRepo.On("GetByToken", ctx, tokenHash).Return(nil, errors.New("token not found"))

	uc := auth.NewRefreshToken(mockRefreshTokenRepo, mockClientRepo, mockTokenService, issuer)

	req := auth.RefreshTokenRequest{
		RefreshToken: rawRefreshToken,
		ClientID:     "client-123",
		ClientSecret: "secret-456",
	}
	resp, err := uc.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	oauthErr, ok := err.(*auth.OAuthError)
	assert.True(t, ok)
	assert.Equal(t, auth.ErrInvalidGrant, oauthErr.Code)

	mockRefreshTokenRepo.AssertExpectations(t)
}

// TestRefreshToken_TokenRevoked tests refresh with revoked token (FR-046)
func TestRefreshToken_TokenRevoked(t *testing.T) {
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockClientRepo := new(MockClientRepositoryForToken)
	mockTokenService := new(MockTokenService)

	ctx := context.Background()
	issuer := "https://auth.example.com"

	rawRefreshToken := "revoked-token"
	tokenHash := hashTokenForRefresh(rawRefreshToken)

	revokedAt := time.Now().Add(-1 * time.Hour)
	storedToken := &entities.RefreshToken{
		ID:          1,
		Token:       tokenHash,
		UserID:      "user-789",
		ClientID:    "client-123",
		TenantID:    "tenant-abc",
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		RevokedFlag: true,
		RevokedAt:   &revokedAt,
	}

	mockRefreshTokenRepo.On("GetByToken", ctx, tokenHash).Return(storedToken, nil)

	uc := auth.NewRefreshToken(mockRefreshTokenRepo, mockClientRepo, mockTokenService, issuer)

	req := auth.RefreshTokenRequest{
		RefreshToken: rawRefreshToken,
		ClientID:     "client-123",
		ClientSecret: "secret-456",
	}
	resp, err := uc.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	oauthErr, ok := err.(*auth.OAuthError)
	assert.True(t, ok)
	assert.Equal(t, auth.ErrInvalidGrant, oauthErr.Code)
	assert.Contains(t, oauthErr.Description, "revoked")

	mockRefreshTokenRepo.AssertExpectations(t)
}

// TestRefreshToken_TokenExpired tests refresh with expired token
func TestRefreshToken_TokenExpired(t *testing.T) {
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockClientRepo := new(MockClientRepositoryForToken)
	mockTokenService := new(MockTokenService)

	ctx := context.Background()
	issuer := "https://auth.example.com"

	rawRefreshToken := "expired-token"
	tokenHash := hashTokenForRefresh(rawRefreshToken)

	storedToken := &entities.RefreshToken{
		ID:        1,
		Token:     tokenHash,
		UserID:    "user-789",
		ClientID:  "client-123",
		TenantID:  "tenant-abc",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		CreatedAt: time.Now().Add(-8 * 24 * time.Hour),
	}

	mockRefreshTokenRepo.On("GetByToken", ctx, tokenHash).Return(storedToken, nil)

	uc := auth.NewRefreshToken(mockRefreshTokenRepo, mockClientRepo, mockTokenService, issuer)

	req := auth.RefreshTokenRequest{
		RefreshToken: rawRefreshToken,
		ClientID:     "client-123",
		ClientSecret: "secret-456",
	}
	resp, err := uc.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	oauthErr, ok := err.(*auth.OAuthError)
	assert.True(t, ok)
	assert.Equal(t, auth.ErrInvalidGrant, oauthErr.Code)
	assert.Contains(t, oauthErr.Description, "expired")

	mockRefreshTokenRepo.AssertExpectations(t)
}

// TestRefreshToken_ClientIDMismatch tests client_id validation (FR-047)
func TestRefreshToken_ClientIDMismatch(t *testing.T) {
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockClientRepo := new(MockClientRepositoryForToken)
	mockTokenService := new(MockTokenService)

	ctx := context.Background()
	issuer := "https://auth.example.com"

	rawRefreshToken := "test-token"
	tokenHash := hashTokenForRefresh(rawRefreshToken)

	storedToken := &entities.RefreshToken{
		ID:        1,
		Token:     tokenHash,
		UserID:    "user-789",
		ClientID:  "original-client", // Different from request
		TenantID:  "tenant-abc",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}

	mockRefreshTokenRepo.On("GetByToken", ctx, tokenHash).Return(storedToken, nil)

	uc := auth.NewRefreshToken(mockRefreshTokenRepo, mockClientRepo, mockTokenService, issuer)

	req := auth.RefreshTokenRequest{
		RefreshToken: rawRefreshToken,
		ClientID:     "different-client", // Mismatch!
		ClientSecret: "secret-456",
	}
	resp, err := uc.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	oauthErr, ok := err.(*auth.OAuthError)
	assert.True(t, ok)
	assert.Equal(t, auth.ErrInvalidGrant, oauthErr.Code)
	assert.Contains(t, oauthErr.Description, "client_id")

	mockRefreshTokenRepo.AssertExpectations(t)
}

// TestRefreshToken_InvalidClientCredentials tests client authentication failure
func TestRefreshToken_InvalidClientCredentials(t *testing.T) {
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockClientRepo := new(MockClientRepositoryForToken)
	mockTokenService := new(MockTokenService)

	ctx := context.Background()
	issuer := "https://auth.example.com"

	rawRefreshToken := "test-token"
	tokenHash := hashTokenForRefresh(rawRefreshToken)
	clientID := "client-123"

	storedToken := &entities.RefreshToken{
		ID:        1,
		Token:     tokenHash,
		UserID:    "user-789",
		ClientID:  clientID,
		TenantID:  "tenant-abc",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}

	mockRefreshTokenRepo.On("GetByToken", ctx, tokenHash).Return(storedToken, nil)
	mockClientRepo.On("ValidateCredentials", ctx, clientID, "wrong-secret").Return(nil, errors.New("invalid credentials"))

	uc := auth.NewRefreshToken(mockRefreshTokenRepo, mockClientRepo, mockTokenService, issuer)

	req := auth.RefreshTokenRequest{
		RefreshToken: rawRefreshToken,
		ClientID:     clientID,
		ClientSecret: "wrong-secret",
	}
	resp, err := uc.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	oauthErr, ok := err.(*auth.OAuthError)
	assert.True(t, ok)
	assert.Equal(t, auth.ErrInvalidClient, oauthErr.Code)

	mockRefreshTokenRepo.AssertExpectations(t)
	mockClientRepo.AssertExpectations(t)
}

// TestRefreshToken_MissingRefreshToken tests missing refresh_token parameter
func TestRefreshToken_MissingRefreshToken(t *testing.T) {
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockClientRepo := new(MockClientRepositoryForToken)
	mockTokenService := new(MockTokenService)

	ctx := context.Background()
	issuer := "https://auth.example.com"

	uc := auth.NewRefreshToken(mockRefreshTokenRepo, mockClientRepo, mockTokenService, issuer)

	req := auth.RefreshTokenRequest{
		RefreshToken: "", // Missing
		ClientID:     "client-123",
		ClientSecret: "secret-456",
	}
	resp, err := uc.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	oauthErr, ok := err.(*auth.OAuthError)
	assert.True(t, ok)
	assert.Equal(t, auth.ErrInvalidRequest, oauthErr.Code)
}

// TestRefreshToken_MissingClientID tests missing client_id parameter
func TestRefreshToken_MissingClientID(t *testing.T) {
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockClientRepo := new(MockClientRepositoryForToken)
	mockTokenService := new(MockTokenService)

	ctx := context.Background()
	issuer := "https://auth.example.com"

	uc := auth.NewRefreshToken(mockRefreshTokenRepo, mockClientRepo, mockTokenService, issuer)

	req := auth.RefreshTokenRequest{
		RefreshToken: "some-token",
		ClientID:     "", // Missing
		ClientSecret: "secret-456",
	}
	resp, err := uc.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	oauthErr, ok := err.(*auth.OAuthError)
	assert.True(t, ok)
	assert.Equal(t, auth.ErrInvalidRequest, oauthErr.Code)
}

// TestRefreshToken_MissingClientSecret tests missing client_secret parameter
func TestRefreshToken_MissingClientSecret(t *testing.T) {
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockClientRepo := new(MockClientRepositoryForToken)
	mockTokenService := new(MockTokenService)

	ctx := context.Background()
	issuer := "https://auth.example.com"

	uc := auth.NewRefreshToken(mockRefreshTokenRepo, mockClientRepo, mockTokenService, issuer)

	req := auth.RefreshTokenRequest{
		RefreshToken: "some-token",
		ClientID:     "client-123",
		ClientSecret: "", // Missing
	}
	resp, err := uc.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	oauthErr, ok := err.(*auth.OAuthError)
	assert.True(t, ok)
	assert.Equal(t, auth.ErrInvalidRequest, oauthErr.Code)
}

// TestRefreshToken_TokenServiceError tests token service failure
func TestRefreshToken_TokenServiceError(t *testing.T) {
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockClientRepo := new(MockClientRepositoryForToken)
	mockTokenService := new(MockTokenService)

	ctx := context.Background()
	issuer := "https://auth.example.com"

	rawRefreshToken := "test-token"
	tokenHash := hashTokenForRefresh(rawRefreshToken)
	clientID := "client-123"
	clientSecret := "secret-456"

	storedToken := &entities.RefreshToken{
		ID:        1,
		Token:     tokenHash,
		UserID:    "user-789",
		ClientID:  clientID,
		TenantID:  "tenant-abc",
		Scope:     "openid",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}

	client := &entities.Client{
		ClientID:   clientID,
		TenantID:   "tenant-abc",
		ClientName: "Test Client",
		IsActive:   true,
	}

	mockRefreshTokenRepo.On("GetByToken", ctx, tokenHash).Return(storedToken, nil)
	mockClientRepo.On("ValidateCredentials", ctx, clientID, clientSecret).Return(client, nil)
	mockTokenService.On("SignAccessToken", ctx, mock.AnythingOfType("*services.TokenClaims")).Return("", errors.New("signing failed"))

	uc := auth.NewRefreshToken(mockRefreshTokenRepo, mockClientRepo, mockTokenService, issuer)

	req := auth.RefreshTokenRequest{
		RefreshToken: rawRefreshToken,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
	resp, err := uc.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	oauthErr, ok := err.(*auth.OAuthError)
	assert.True(t, ok)
	assert.Equal(t, auth.ErrServerError, oauthErr.Code)

	mockRefreshTokenRepo.AssertExpectations(t)
	mockClientRepo.AssertExpectations(t)
	mockTokenService.AssertExpectations(t)
}

// Suppress unused import warning for services
var _ services.TokenService = (*MockTokenService)(nil)
