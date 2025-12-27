package usecase_test

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthCodeRepository is a mock for AuthCodeRepository
type MockAuthCodeRepository struct {
	mock.Mock
}

func (m *MockAuthCodeRepository) Store(ctx context.Context, code *entities.AuthorizationCode, ttl time.Duration) error {
	args := m.Called(ctx, code, ttl)
	return args.Error(0)
}

func (m *MockAuthCodeRepository) Get(ctx context.Context, code string) (*entities.AuthorizationCode, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.AuthorizationCode), args.Error(1)
}

func (m *MockAuthCodeRepository) MarkAsUsed(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func (m *MockAuthCodeRepository) Delete(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func (m *MockAuthCodeRepository) Exists(ctx context.Context, code string) (bool, error) {
	args := m.Called(ctx, code)
	return args.Bool(0), args.Error(1)
}

// MockRefreshTokenRepository is a mock for RefreshTokenRepository
type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) GetByToken(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) Revoke(ctx context.Context, tokenHash string, revokedBy string) error {
	args := m.Called(ctx, tokenHash, revokedBy)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string, clientID string) error {
	args := m.Called(ctx, userID, clientID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRefreshTokenRepository) IsRevoked(ctx context.Context, tokenHash string) (bool, error) {
	args := m.Called(ctx, tokenHash)
	return args.Bool(0), args.Error(1)
}

func (m *MockRefreshTokenRepository) UpdateLastUsed(ctx context.Context, tokenHash string) error {
	args := m.Called(ctx, tokenHash)
	return args.Error(0)
}

// MockTokenService is a mock for TokenService
type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) SignIDToken(ctx context.Context, claims *services.TokenClaims) (string, error) {
	args := m.Called(ctx, claims)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) SignAccessToken(ctx context.Context, claims *services.TokenClaims) (string, error) {
	args := m.Called(ctx, claims)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) ValidateTokenSignature(ctx context.Context, token string) (*services.TokenClaims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.TokenClaims), args.Error(1)
}

func (m *MockTokenService) GetPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	args := m.Called(ctx, keyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rsa.PublicKey), args.Error(1)
}

func (m *MockTokenService) GetJWKS(ctx context.Context) (*services.JWKS, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.JWKS), args.Error(1)
}

func (m *MockTokenService) GetActiveKeyID(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

// MockClientRepositoryForToken is a mock for ClientRepository
type MockClientRepositoryForToken struct {
	mock.Mock
}

func (m *MockClientRepositoryForToken) Create(ctx context.Context, client *entities.Client) error {
	args := m.Called(ctx, client)
	return args.Error(0)
}

func (m *MockClientRepositoryForToken) GetByID(ctx context.Context, id string) (*entities.Client, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Client), args.Error(1)
}

func (m *MockClientRepositoryForToken) GetByClientID(ctx context.Context, clientID string, tenantID string) (*entities.Client, error) {
	args := m.Called(ctx, clientID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Client), args.Error(1)
}

func (m *MockClientRepositoryForToken) Update(ctx context.Context, client *entities.Client) error {
	args := m.Called(ctx, client)
	return args.Error(0)
}

func (m *MockClientRepositoryForToken) Delete(ctx context.Context, clientID string) error {
	args := m.Called(ctx, clientID)
	return args.Error(0)
}

func (m *MockClientRepositoryForToken) ListByTenant(ctx context.Context, tenantID string) ([]*entities.Client, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Client), args.Error(1)
}

func (m *MockClientRepositoryForToken) ValidateCredentials(ctx context.Context, clientID string, clientSecret string) (*entities.Client, error) {
	args := m.Called(ctx, clientID, clientSecret)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Client), args.Error(1)
}

// MockUserRepository is a mock for fetching user info
type MockUserRepositoryForToken struct {
	mock.Mock
}

func (m *MockUserRepositoryForToken) GetByID(ctx context.Context, id string) (*entities.AdminUser, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.AdminUser), args.Error(1)
}

// Helper function to generate valid PKCE code verifier and challenge
func generatePKCEPair() (verifier string, challenge string) {
	verifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	hash := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(hash[:])
	return
}

// Test helper to create a valid authorization code
func createValidAuthCode(verifier string) *entities.AuthorizationCode {
	_, challenge := generatePKCEPair()
	return &entities.AuthorizationCode{
		Code:                "test-auth-code",
		ClientID:            "test-client-id",
		TenantID:            "test-tenant-id",
		UserID:              "test-user-id",
		RedirectURI:         "https://app.example.com/callback",
		Scope:               "openid profile email",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		ExpiresAt:           time.Now().Add(5 * time.Minute),
		CreatedAt:           time.Now(),
		UsedFlag:            false,
	}
}

func createValidClient() *entities.Client {
	return &entities.Client{
		TenantID:            "test-tenant-id",
		ClientID:            "test-client-id",
		ClientSecretHash:    "hashed-secret",
		ClientName:          "Test Client",
		AllowedRedirectURIs: []string{"https://app.example.com/callback"},
		IsActive:            true,
		CreatedAt:           time.Now(),
	}
}

func createValidUser() *entities.AdminUser {
	return &entities.AdminUser{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Email:    "user@example.com",
		FullName: "Test User",
	}
}

func TestIssueToken_Execute_Success(t *testing.T) {
	ctx := context.Background()
	verifier, _ := generatePKCEPair()

	authCodeRepo := new(MockAuthCodeRepository)
	clientRepo := new(MockClientRepositoryForToken)
	refreshTokenRepo := new(MockRefreshTokenRepository)
	tokenService := new(MockTokenService)

	authCode := createValidAuthCode(verifier)
	client := createValidClient()

	// Setup expectations
	authCodeRepo.On("Get", ctx, "test-auth-code").Return(authCode, nil)
	clientRepo.On("ValidateCredentials", ctx, "test-client-id", "test-client-secret").Return(client, nil)
	authCodeRepo.On("MarkAsUsed", ctx, "test-auth-code").Return(nil)
	tokenService.On("SignIDToken", ctx, mock.AnythingOfType("*services.TokenClaims")).Return("id-token-jwt", nil)
	tokenService.On("SignAccessToken", ctx, mock.AnythingOfType("*services.TokenClaims")).Return("access-token-jwt", nil)
	refreshTokenRepo.On("Create", ctx, mock.AnythingOfType("*entities.RefreshToken")).Return(nil)

	useCase := auth.NewIssueToken(authCodeRepo, clientRepo, refreshTokenRepo, tokenService, "https://sso.example.com")

	req := auth.TokenRequest{
		GrantType:    "authorization_code",
		Code:         "test-auth-code",
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		CodeVerifier: verifier,
	}

	resp, err := useCase.Execute(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "id-token-jwt", resp.IDToken)
	assert.Equal(t, "access-token-jwt", resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, 900, resp.ExpiresIn) // 15 minutes
	assert.Equal(t, "openid profile email", resp.Scope)

	authCodeRepo.AssertExpectations(t)
	clientRepo.AssertExpectations(t)
	refreshTokenRepo.AssertExpectations(t)
	tokenService.AssertExpectations(t)
}

func TestIssueToken_Execute_InvalidGrantType(t *testing.T) {
	ctx := context.Background()

	authCodeRepo := new(MockAuthCodeRepository)
	clientRepo := new(MockClientRepositoryForToken)
	refreshTokenRepo := new(MockRefreshTokenRepository)
	tokenService := new(MockTokenService)

	useCase := auth.NewIssueToken(authCodeRepo, clientRepo, refreshTokenRepo, tokenService, "https://sso.example.com")

	req := auth.TokenRequest{
		GrantType: "invalid_grant",
		Code:      "test-auth-code",
	}

	resp, err := useCase.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)

	var oauthErr *auth.OAuthError
	assert.True(t, errors.As(err, &oauthErr))
	assert.Equal(t, auth.ErrUnsupportedGrantType, oauthErr.Code)
}

func TestIssueToken_Execute_MissingCode(t *testing.T) {
	ctx := context.Background()

	authCodeRepo := new(MockAuthCodeRepository)
	clientRepo := new(MockClientRepositoryForToken)
	refreshTokenRepo := new(MockRefreshTokenRepository)
	tokenService := new(MockTokenService)

	useCase := auth.NewIssueToken(authCodeRepo, clientRepo, refreshTokenRepo, tokenService, "https://sso.example.com")

	req := auth.TokenRequest{
		GrantType:    "authorization_code",
		Code:         "", // Missing
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		CodeVerifier: "some-verifier",
	}

	resp, err := useCase.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)

	var oauthErr *auth.OAuthError
	assert.True(t, errors.As(err, &oauthErr))
	assert.Equal(t, auth.ErrInvalidRequest, oauthErr.Code)
}

func TestIssueToken_Execute_AuthCodeNotFound(t *testing.T) {
	ctx := context.Background()

	authCodeRepo := new(MockAuthCodeRepository)
	clientRepo := new(MockClientRepositoryForToken)
	refreshTokenRepo := new(MockRefreshTokenRepository)
	tokenService := new(MockTokenService)

	authCodeRepo.On("Get", ctx, "invalid-code").Return(nil, errors.New("not found"))

	useCase := auth.NewIssueToken(authCodeRepo, clientRepo, refreshTokenRepo, tokenService, "https://sso.example.com")

	req := auth.TokenRequest{
		GrantType:    "authorization_code",
		Code:         "invalid-code",
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		CodeVerifier: "some-verifier",
	}

	resp, err := useCase.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)

	var oauthErr *auth.OAuthError
	assert.True(t, errors.As(err, &oauthErr))
	assert.Equal(t, auth.ErrInvalidGrant, oauthErr.Code)

	authCodeRepo.AssertExpectations(t)
}

func TestIssueToken_Execute_AuthCodeExpired(t *testing.T) {
	ctx := context.Background()
	verifier, _ := generatePKCEPair()

	authCodeRepo := new(MockAuthCodeRepository)
	clientRepo := new(MockClientRepositoryForToken)
	refreshTokenRepo := new(MockRefreshTokenRepository)
	tokenService := new(MockTokenService)

	authCode := createValidAuthCode(verifier)
	authCode.ExpiresAt = time.Now().Add(-1 * time.Minute) // Expired

	authCodeRepo.On("Get", ctx, "test-auth-code").Return(authCode, nil)

	useCase := auth.NewIssueToken(authCodeRepo, clientRepo, refreshTokenRepo, tokenService, "https://sso.example.com")

	req := auth.TokenRequest{
		GrantType:    "authorization_code",
		Code:         "test-auth-code",
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		CodeVerifier: verifier,
	}

	resp, err := useCase.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)

	var oauthErr *auth.OAuthError
	assert.True(t, errors.As(err, &oauthErr))
	assert.Equal(t, auth.ErrInvalidGrant, oauthErr.Code)
	assert.Contains(t, oauthErr.Description, "expired")

	authCodeRepo.AssertExpectations(t)
}

func TestIssueToken_Execute_AuthCodeAlreadyUsed(t *testing.T) {
	ctx := context.Background()
	verifier, _ := generatePKCEPair()

	authCodeRepo := new(MockAuthCodeRepository)
	clientRepo := new(MockClientRepositoryForToken)
	refreshTokenRepo := new(MockRefreshTokenRepository)
	tokenService := new(MockTokenService)

	authCode := createValidAuthCode(verifier)
	authCode.UsedFlag = true // Already used

	authCodeRepo.On("Get", ctx, "test-auth-code").Return(authCode, nil)

	useCase := auth.NewIssueToken(authCodeRepo, clientRepo, refreshTokenRepo, tokenService, "https://sso.example.com")

	req := auth.TokenRequest{
		GrantType:    "authorization_code",
		Code:         "test-auth-code",
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		CodeVerifier: verifier,
	}

	resp, err := useCase.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)

	var oauthErr *auth.OAuthError
	assert.True(t, errors.As(err, &oauthErr))
	assert.Equal(t, auth.ErrInvalidGrant, oauthErr.Code)
	assert.Contains(t, oauthErr.Description, "already been used")

	authCodeRepo.AssertExpectations(t)
}

func TestIssueToken_Execute_InvalidClientCredentials(t *testing.T) {
	ctx := context.Background()
	verifier, _ := generatePKCEPair()

	authCodeRepo := new(MockAuthCodeRepository)
	clientRepo := new(MockClientRepositoryForToken)
	refreshTokenRepo := new(MockRefreshTokenRepository)
	tokenService := new(MockTokenService)

	authCode := createValidAuthCode(verifier)

	authCodeRepo.On("Get", ctx, "test-auth-code").Return(authCode, nil)
	clientRepo.On("ValidateCredentials", ctx, "test-client-id", "wrong-secret").Return(nil, errors.New("invalid credentials"))

	useCase := auth.NewIssueToken(authCodeRepo, clientRepo, refreshTokenRepo, tokenService, "https://sso.example.com")

	req := auth.TokenRequest{
		GrantType:    "authorization_code",
		Code:         "test-auth-code",
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     "test-client-id",
		ClientSecret: "wrong-secret",
		CodeVerifier: verifier,
	}

	resp, err := useCase.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)

	var oauthErr *auth.OAuthError
	assert.True(t, errors.As(err, &oauthErr))
	assert.Equal(t, auth.ErrInvalidClient, oauthErr.Code)

	authCodeRepo.AssertExpectations(t)
	clientRepo.AssertExpectations(t)
}

func TestIssueToken_Execute_ClientIDMismatch(t *testing.T) {
	ctx := context.Background()
	verifier, _ := generatePKCEPair()

	authCodeRepo := new(MockAuthCodeRepository)
	clientRepo := new(MockClientRepositoryForToken)
	refreshTokenRepo := new(MockRefreshTokenRepository)
	tokenService := new(MockTokenService)

	authCode := createValidAuthCode(verifier)
	client := createValidClient()

	authCodeRepo.On("Get", ctx, "test-auth-code").Return(authCode, nil)
	clientRepo.On("ValidateCredentials", ctx, "different-client-id", "test-client-secret").Return(client, nil)

	useCase := auth.NewIssueToken(authCodeRepo, clientRepo, refreshTokenRepo, tokenService, "https://sso.example.com")

	req := auth.TokenRequest{
		GrantType:    "authorization_code",
		Code:         "test-auth-code",
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     "different-client-id", // Mismatch with auth code
		ClientSecret: "test-client-secret",
		CodeVerifier: verifier,
	}

	resp, err := useCase.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)

	var oauthErr *auth.OAuthError
	assert.True(t, errors.As(err, &oauthErr))
	assert.Equal(t, auth.ErrInvalidGrant, oauthErr.Code)
	assert.Contains(t, oauthErr.Description, "client_id")

	authCodeRepo.AssertExpectations(t)
	clientRepo.AssertExpectations(t)
}

func TestIssueToken_Execute_RedirectURIMismatch(t *testing.T) {
	ctx := context.Background()
	verifier, _ := generatePKCEPair()

	authCodeRepo := new(MockAuthCodeRepository)
	clientRepo := new(MockClientRepositoryForToken)
	refreshTokenRepo := new(MockRefreshTokenRepository)
	tokenService := new(MockTokenService)

	authCode := createValidAuthCode(verifier)
	client := createValidClient()

	authCodeRepo.On("Get", ctx, "test-auth-code").Return(authCode, nil)
	clientRepo.On("ValidateCredentials", ctx, "test-client-id", "test-client-secret").Return(client, nil)

	useCase := auth.NewIssueToken(authCodeRepo, clientRepo, refreshTokenRepo, tokenService, "https://sso.example.com")

	req := auth.TokenRequest{
		GrantType:    "authorization_code",
		Code:         "test-auth-code",
		RedirectURI:  "https://different.example.com/callback", // Mismatch
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		CodeVerifier: verifier,
	}

	resp, err := useCase.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)

	var oauthErr *auth.OAuthError
	assert.True(t, errors.As(err, &oauthErr))
	assert.Equal(t, auth.ErrInvalidGrant, oauthErr.Code)
	assert.Contains(t, oauthErr.Description, "redirect_uri")

	authCodeRepo.AssertExpectations(t)
	clientRepo.AssertExpectations(t)
}

func TestIssueToken_Execute_InvalidPKCEVerifier(t *testing.T) {
	ctx := context.Background()
	verifier, _ := generatePKCEPair()

	authCodeRepo := new(MockAuthCodeRepository)
	clientRepo := new(MockClientRepositoryForToken)
	refreshTokenRepo := new(MockRefreshTokenRepository)
	tokenService := new(MockTokenService)

	authCode := createValidAuthCode(verifier)
	client := createValidClient()

	authCodeRepo.On("Get", ctx, "test-auth-code").Return(authCode, nil)
	clientRepo.On("ValidateCredentials", ctx, "test-client-id", "test-client-secret").Return(client, nil)

	useCase := auth.NewIssueToken(authCodeRepo, clientRepo, refreshTokenRepo, tokenService, "https://sso.example.com")

	req := auth.TokenRequest{
		GrantType:    "authorization_code",
		Code:         "test-auth-code",
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		CodeVerifier: "wrong-verifier", // Invalid
	}

	resp, err := useCase.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)

	var oauthErr *auth.OAuthError
	assert.True(t, errors.As(err, &oauthErr))
	assert.Equal(t, auth.ErrInvalidGrant, oauthErr.Code)
	assert.Contains(t, oauthErr.Description, "code_verifier")

	authCodeRepo.AssertExpectations(t)
	clientRepo.AssertExpectations(t)
}

func TestIssueToken_Execute_MissingCodeVerifier(t *testing.T) {
	ctx := context.Background()

	authCodeRepo := new(MockAuthCodeRepository)
	clientRepo := new(MockClientRepositoryForToken)
	refreshTokenRepo := new(MockRefreshTokenRepository)
	tokenService := new(MockTokenService)

	// No need to set up mock expectations - validation fails early
	useCase := auth.NewIssueToken(authCodeRepo, clientRepo, refreshTokenRepo, tokenService, "https://sso.example.com")

	req := auth.TokenRequest{
		GrantType:    "authorization_code",
		Code:         "test-auth-code",
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		CodeVerifier: "", // Missing
	}

	resp, err := useCase.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)

	var oauthErr *auth.OAuthError
	assert.True(t, errors.As(err, &oauthErr))
	assert.Equal(t, auth.ErrInvalidRequest, oauthErr.Code)
	assert.Contains(t, oauthErr.Description, "code_verifier")
}

func TestIssueToken_Execute_TokenSigningFailure(t *testing.T) {
	ctx := context.Background()
	verifier, _ := generatePKCEPair()

	authCodeRepo := new(MockAuthCodeRepository)
	clientRepo := new(MockClientRepositoryForToken)
	refreshTokenRepo := new(MockRefreshTokenRepository)
	tokenService := new(MockTokenService)

	authCode := createValidAuthCode(verifier)
	client := createValidClient()

	authCodeRepo.On("Get", ctx, "test-auth-code").Return(authCode, nil)
	clientRepo.On("ValidateCredentials", ctx, "test-client-id", "test-client-secret").Return(client, nil)
	authCodeRepo.On("MarkAsUsed", ctx, "test-auth-code").Return(nil)
	tokenService.On("SignIDToken", ctx, mock.AnythingOfType("*services.TokenClaims")).Return("", errors.New("signing error"))

	useCase := auth.NewIssueToken(authCodeRepo, clientRepo, refreshTokenRepo, tokenService, "https://sso.example.com")

	req := auth.TokenRequest{
		GrantType:    "authorization_code",
		Code:         "test-auth-code",
		RedirectURI:  "https://app.example.com/callback",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		CodeVerifier: verifier,
	}

	resp, err := useCase.Execute(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)

	var oauthErr *auth.OAuthError
	assert.True(t, errors.As(err, &oauthErr))
	assert.Equal(t, auth.ErrServerError, oauthErr.Code)

	authCodeRepo.AssertExpectations(t)
	clientRepo.AssertExpectations(t)
	tokenService.AssertExpectations(t)
}

func TestValidateTokenRequest_MissingFields(t *testing.T) {
	tests := []struct {
		name    string
		req     auth.TokenRequest
		wantErr string
	}{
		{
			name: "missing client_id",
			req: auth.TokenRequest{
				GrantType:    "authorization_code",
				Code:         "code",
				RedirectURI:  "https://app.example.com/callback",
				ClientSecret: "secret",
				CodeVerifier: "verifier",
			},
			wantErr: "client_id",
		},
		{
			name: "missing client_secret",
			req: auth.TokenRequest{
				GrantType:    "authorization_code",
				Code:         "code",
				RedirectURI:  "https://app.example.com/callback",
				ClientID:     "client",
				CodeVerifier: "verifier",
			},
			wantErr: "client_secret",
		},
		{
			name: "missing redirect_uri",
			req: auth.TokenRequest{
				GrantType:    "authorization_code",
				Code:         "code",
				ClientID:     "client",
				ClientSecret: "secret",
				CodeVerifier: "verifier",
			},
			wantErr: "redirect_uri",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auth.ValidateTokenRequest(tt.req)
			assert.Error(t, err)

			var oauthErr *auth.OAuthError
			assert.True(t, errors.As(err, &oauthErr))
			assert.Equal(t, auth.ErrInvalidRequest, oauthErr.Code)
			assert.Contains(t, oauthErr.Description, tt.wantErr)
		})
	}
}

func TestVerifyPKCE_S256_Valid(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	err := auth.VerifyPKCE(verifier, challenge, "S256")
	assert.NoError(t, err)
}

func TestVerifyPKCE_S256_Invalid(t *testing.T) {
	verifier := "wrong-verifier"
	hash := sha256.Sum256([]byte("different-verifier"))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	err := auth.VerifyPKCE(verifier, challenge, "S256")
	assert.Error(t, err)
}

func TestVerifyPKCE_UnsupportedMethod(t *testing.T) {
	err := auth.VerifyPKCE("verifier", "challenge", "plain")
	assert.Error(t, err)

	var oauthErr *auth.OAuthError
	assert.True(t, errors.As(err, &oauthErr))
	assert.Contains(t, oauthErr.Description, "S256")
}

// Verify interface compliance
var _ repositories.AuthCodeRepository = (*MockAuthCodeRepository)(nil)
var _ repositories.RefreshTokenRepository = (*MockRefreshTokenRepository)(nil)
var _ repositories.ClientRepository = (*MockClientRepositoryForToken)(nil)
