/**
 * Integration tests for OAuth token exchange (POST /oauth2/token)
 * Tests FR-019 through FR-036 from spec.md
 */

package integration

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockClientRepositoryForTokenTest implements ClientRepository for token tests
type MockClientRepositoryForTokenTest struct {
	clients map[string]*entities.Client
}

func NewMockClientRepositoryForTokenTest() *MockClientRepositoryForTokenTest {
	return &MockClientRepositoryForTokenTest{
		clients: make(map[string]*entities.Client),
	}
}

func (m *MockClientRepositoryForTokenTest) Create(ctx context.Context, client *entities.Client) error {
	m.clients[client.ClientID] = client
	return nil
}

func (m *MockClientRepositoryForTokenTest) GetByID(ctx context.Context, id string) (*entities.Client, error) {
	for _, c := range m.clients {
		if c.ClientID == id {
			return c, nil
		}
	}
	return nil, errors.New("client not found")
}

func (m *MockClientRepositoryForTokenTest) GetByClientID(ctx context.Context, clientID string, tenantID string) (*entities.Client, error) {
	if c, ok := m.clients[clientID]; ok {
		if c.TenantID == tenantID {
			return c, nil
		}
	}
	return nil, errors.New("client not found")
}

func (m *MockClientRepositoryForTokenTest) Update(ctx context.Context, client *entities.Client) error {
	m.clients[client.ClientID] = client
	return nil
}

func (m *MockClientRepositoryForTokenTest) Delete(ctx context.Context, clientID string) error {
	delete(m.clients, clientID)
	return nil
}

func (m *MockClientRepositoryForTokenTest) ListByTenant(ctx context.Context, tenantID string) ([]*entities.Client, error) {
	return nil, nil
}

func (m *MockClientRepositoryForTokenTest) ValidateCredentials(ctx context.Context, clientID string, clientSecret string) (*entities.Client, error) {
	if c, ok := m.clients[clientID]; ok {
		// For testing, we just check if client_secret matches the raw value stored
		if c.ClientSecretHash == clientSecret {
			return c, nil
		}
		return nil, errors.New("invalid client credentials")
	}
	return nil, errors.New("client not found")
}

func (m *MockClientRepositoryForTokenTest) CountByTenant(ctx context.Context, tenantID string) (int, error) {
	return 0, nil
}

func (m *MockClientRepositoryForTokenTest) ListByTenantPaginated(ctx context.Context, tenantID string, search string, page int, pageSize int) ([]*entities.Client, int, error) {
	return nil, 0, nil
}

var _ repositories.ClientRepository = (*MockClientRepositoryForTokenTest)(nil)

// MockRefreshTokenRepository for token tests
type MockRefreshTokenRepository struct {
	tokens map[string]*entities.RefreshToken
}

func NewMockRefreshTokenRepository() *MockRefreshTokenRepository {
	return &MockRefreshTokenRepository{
		tokens: make(map[string]*entities.RefreshToken),
	}
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	m.tokens[token.Token] = token
	return nil
}

func (m *MockRefreshTokenRepository) GetByToken(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	if t, ok := m.tokens[tokenHash]; ok {
		return t, nil
	}
	return nil, errors.New("token not found")
}

func (m *MockRefreshTokenRepository) Revoke(ctx context.Context, tokenHash string, revokedBy string) error {
	if t, ok := m.tokens[tokenHash]; ok {
		t.Revoke(revokedBy)
		return nil
	}
	return errors.New("token not found")
}

func (m *MockRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string, clientID string) error {
	for _, t := range m.tokens {
		if t.UserID == userID && t.ClientID == clientID {
			t.Revoke("user revocation")
		}
	}
	return nil
}

func (m *MockRefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *MockRefreshTokenRepository) IsRevoked(ctx context.Context, tokenHash string) (bool, error) {
	if t, ok := m.tokens[tokenHash]; ok {
		return t.IsRevoked(), nil
	}
	return false, errors.New("token not found")
}

func (m *MockRefreshTokenRepository) UpdateLastUsed(ctx context.Context, tokenHash string) error {
	return nil
}

func (m *MockRefreshTokenRepository) RevokeByClientID(ctx context.Context, clientID string) error {
	return nil
}

func (m *MockRefreshTokenRepository) RevokeByUserID(ctx context.Context, userID string) error {
	return nil
}

func (m *MockRefreshTokenRepository) ListByUserID(ctx context.Context, userID string) ([]*entities.RefreshToken, error) {
	return []*entities.RefreshToken{}, nil
}

func (m *MockRefreshTokenRepository) GetByID(ctx context.Context, id int64) (*entities.RefreshToken, error) {
	return nil, errors.New("not found")
}

var _ repositories.RefreshTokenRepository = (*MockRefreshTokenRepository)(nil)

// MockTokenService for token tests
type MockTokenService struct {
	keyID string
}

func NewMockTokenService() *MockTokenService {
	return &MockTokenService{
		keyID: "test-key-1",
	}
}

func (m *MockTokenService) SignIDToken(ctx context.Context, claims *services.TokenClaims) (string, error) {
	// Return a mock JWT token (simplified for testing)
	return "mock-id-token-jwt.payload.signature", nil
}

func (m *MockTokenService) SignAccessToken(ctx context.Context, claims *services.TokenClaims) (string, error) {
	// Return a mock JWT token (simplified for testing)
	return "mock-access-token-jwt.payload.signature", nil
}

func (m *MockTokenService) ValidateTokenSignature(ctx context.Context, token string) (*services.TokenClaims, error) {
	return nil, nil
}

func (m *MockTokenService) GetPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	return nil, nil
}

func (m *MockTokenService) GetJWKS(ctx context.Context) (*services.JWKS, error) {
	return &services.JWKS{
		Keys: []services.JWK{
			{
				KeyType:   "RSA",
				Use:       "sig",
				KeyID:     m.keyID,
				Algorithm: "RS256",
				N:         "test-n",
				E:         "AQAB",
			},
		},
	}, nil
}

func (m *MockTokenService) GetActiveKeyID(ctx context.Context) (string, error) {
	return m.keyID, nil
}

var _ services.TokenService = (*MockTokenService)(nil)

// Helper function to generate PKCE pair
func generatePKCEPair() (verifier string, challenge string) {
	// Generate a random verifier
	bytes := make([]byte, 32)
	rand.Read(bytes)
	verifier = base64.RawURLEncoding.EncodeToString(bytes)

	// Generate challenge using S256
	hash := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(hash[:])
	return
}

func setupTokenRouter(t *testing.T) (*gin.Engine, *MockClientRepositoryForTokenTest, *MockIntegrationAuthCodeRepository, *MockRefreshTokenRepository) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	clientRepo := NewMockClientRepositoryForTokenTest()
	authCodeRepo := NewMockIntegrationAuthCodeRepository()
	refreshTokenRepo := NewMockRefreshTokenRepository()
	tokenService := NewMockTokenService()
	roleRepo := NewMockIntegrationRoleRepository()
	roleRepo.roles["user_123:test_client_123"] = []*entities.UserRoleAssignment{{
		UserID: "user_123", ClientID: "test_client_123", Role: "user", IsActive: true,
	}}

	// Create test client with "test-secret" as the raw secret
	testClient := &entities.Client{
		ClientID:            "test_client_123",
		TenantID:            "tenant_xyz",
		ClientName:          "Test OAuth App",
		ClientSecretHash:    "test-secret", // For testing, we use raw value
		AllowedRedirectURIs: []string{"https://app.example.com/callback"},
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	clientRepo.clients[testClient.ClientID] = testClient

	issueTokenUC := auth.NewIssueToken(authCodeRepo, clientRepo, refreshTokenRepo, roleRepo, tokenService, "https://sso.test.com")
	oauthHandler := handlers.NewOAuthHandler(nil, issueTokenUC, clientRepo)

	router.POST("/oauth2/token", oauthHandler.Token)

	return router, clientRepo, authCodeRepo, refreshTokenRepo
}

func TestOAuthHandler_Token_Success(t *testing.T) {
	router, _, authCodeRepo, _ := setupTokenRouter(t)

	// Generate PKCE pair
	verifier, challenge := generatePKCEPair()

	// Create authorization code
	authCode := &entities.AuthorizationCode{
		Code:                "test-auth-code-123",
		ClientID:            "test_client_123",
		TenantID:            "tenant_xyz",
		UserID:              "user_123",
		RedirectURI:         "https://app.example.com/callback",
		Scope:               "openid profile email",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		ExpiresAt:           time.Now().Add(5 * time.Minute),
		CreatedAt:           time.Now(),
		UsedFlag:            false,
	}
	authCodeRepo.codes[authCode.Code] = authCode

	// Build token request
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", "test-auth-code-123")
	form.Set("redirect_uri", "https://app.example.com/callback")
	form.Set("client_id", "test_client_123")
	form.Set("client_secret", "test-secret")
	form.Set("code_verifier", verifier)

	req := httptest.NewRequest(http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var tokenResp auth.TokenResponse
	err := json.Unmarshal(w.Body.Bytes(), &tokenResp)
	require.NoError(t, err)

	// Verify token response structure (FR-027)
	assert.NotEmpty(t, tokenResp.AccessToken)
	assert.NotEmpty(t, tokenResp.IDToken)
	assert.NotEmpty(t, tokenResp.RefreshToken)
	assert.Equal(t, "Bearer", tokenResp.TokenType)
	assert.Equal(t, 900, tokenResp.ExpiresIn) // 15 minutes
	assert.Equal(t, "openid profile email", tokenResp.Scope)
}

func TestOAuthHandler_Token_InvalidGrantType(t *testing.T) {
	router, _, _, _ := setupTokenRouter(t)

	form := url.Values{}
	form.Set("grant_type", "invalid_grant")
	form.Set("code", "test-code")
	form.Set("client_id", "test_client_123")
	form.Set("client_secret", "test-secret")

	req := httptest.NewRequest(http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Equal(t, "unsupported_grant_type", errResp["error"])
}

func TestOAuthHandler_Token_ExpiredCode(t *testing.T) {
	router, _, authCodeRepo, _ := setupTokenRouter(t)

	verifier, challenge := generatePKCEPair()

	// Create expired authorization code (FR-023)
	authCode := &entities.AuthorizationCode{
		Code:                "expired-code",
		ClientID:            "test_client_123",
		TenantID:            "tenant_xyz",
		UserID:              "user_123",
		RedirectURI:         "https://app.example.com/callback",
		Scope:               "openid profile email",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		ExpiresAt:           time.Now().Add(-1 * time.Minute), // Expired
		CreatedAt:           time.Now().Add(-6 * time.Minute),
		UsedFlag:            false,
	}
	authCodeRepo.codes[authCode.Code] = authCode

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", "expired-code")
	form.Set("redirect_uri", "https://app.example.com/callback")
	form.Set("client_id", "test_client_123")
	form.Set("client_secret", "test-secret")
	form.Set("code_verifier", verifier)

	req := httptest.NewRequest(http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Equal(t, "invalid_grant", errResp["error"])
	assert.Contains(t, errResp["error_description"], "expired")
}

func TestOAuthHandler_Token_CodeReuse(t *testing.T) {
	router, _, authCodeRepo, _ := setupTokenRouter(t)

	verifier, challenge := generatePKCEPair()

	// Create already used authorization code (FR-024)
	usedAt := time.Now().Add(-1 * time.Minute)
	authCode := &entities.AuthorizationCode{
		Code:                "used-code",
		ClientID:            "test_client_123",
		TenantID:            "tenant_xyz",
		UserID:              "user_123",
		RedirectURI:         "https://app.example.com/callback",
		Scope:               "openid profile email",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		ExpiresAt:           time.Now().Add(5 * time.Minute),
		CreatedAt:           time.Now().Add(-2 * time.Minute),
		UsedFlag:            true, // Already used
		UsedAt:              &usedAt,
	}
	authCodeRepo.codes[authCode.Code] = authCode

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", "used-code")
	form.Set("redirect_uri", "https://app.example.com/callback")
	form.Set("client_id", "test_client_123")
	form.Set("client_secret", "test-secret")
	form.Set("code_verifier", verifier)

	req := httptest.NewRequest(http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Equal(t, "invalid_grant", errResp["error"])
	assert.Contains(t, errResp["error_description"], "already been used")
}

func TestOAuthHandler_Token_InvalidPKCE(t *testing.T) {
	router, _, authCodeRepo, _ := setupTokenRouter(t)

	_, challenge := generatePKCEPair()

	authCode := &entities.AuthorizationCode{
		Code:                "test-code-pkce",
		ClientID:            "test_client_123",
		TenantID:            "tenant_xyz",
		UserID:              "user_123",
		RedirectURI:         "https://app.example.com/callback",
		Scope:               "openid profile email",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		ExpiresAt:           time.Now().Add(5 * time.Minute),
		CreatedAt:           time.Now(),
		UsedFlag:            false,
	}
	authCodeRepo.codes[authCode.Code] = authCode

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", "test-code-pkce")
	form.Set("redirect_uri", "https://app.example.com/callback")
	form.Set("client_id", "test_client_123")
	form.Set("client_secret", "test-secret")
	form.Set("code_verifier", "wrong-verifier") // Wrong verifier (FR-021)

	req := httptest.NewRequest(http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Equal(t, "invalid_grant", errResp["error"])
	assert.Contains(t, errResp["error_description"], "code_verifier")
}

func TestOAuthHandler_Token_MissingPKCE(t *testing.T) {
	router, _, _, _ := setupTokenRouter(t)

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", "test-code")
	form.Set("redirect_uri", "https://app.example.com/callback")
	form.Set("client_id", "test_client_123")
	form.Set("client_secret", "test-secret")
	// Missing code_verifier (FR-008)

	req := httptest.NewRequest(http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Equal(t, "invalid_request", errResp["error"])
	assert.Contains(t, errResp["error_description"], "code_verifier")
}

func TestOAuthHandler_Token_InvalidClientCredentials(t *testing.T) {
	router, _, authCodeRepo, _ := setupTokenRouter(t)

	verifier, challenge := generatePKCEPair()

	authCode := &entities.AuthorizationCode{
		Code:                "test-code-client",
		ClientID:            "test_client_123",
		TenantID:            "tenant_xyz",
		UserID:              "user_123",
		RedirectURI:         "https://app.example.com/callback",
		Scope:               "openid profile email",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		ExpiresAt:           time.Now().Add(5 * time.Minute),
		CreatedAt:           time.Now(),
		UsedFlag:            false,
	}
	authCodeRepo.codes[authCode.Code] = authCode

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", "test-code-client")
	form.Set("redirect_uri", "https://app.example.com/callback")
	form.Set("client_id", "test_client_123")
	form.Set("client_secret", "wrong-secret") // Wrong secret (FR-025)
	form.Set("code_verifier", verifier)

	req := httptest.NewRequest(http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Equal(t, "invalid_client", errResp["error"])
}

func TestOAuthHandler_Token_RedirectURIMismatch(t *testing.T) {
	router, _, authCodeRepo, _ := setupTokenRouter(t)

	verifier, challenge := generatePKCEPair()

	authCode := &entities.AuthorizationCode{
		Code:                "test-code-redirect",
		ClientID:            "test_client_123",
		TenantID:            "tenant_xyz",
		UserID:              "user_123",
		RedirectURI:         "https://app.example.com/callback",
		Scope:               "openid profile email",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		ExpiresAt:           time.Now().Add(5 * time.Minute),
		CreatedAt:           time.Now(),
		UsedFlag:            false,
	}
	authCodeRepo.codes[authCode.Code] = authCode

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", "test-code-redirect")
	form.Set("redirect_uri", "https://different.example.com/callback") // Mismatch (FR-022)
	form.Set("client_id", "test_client_123")
	form.Set("client_secret", "test-secret")
	form.Set("code_verifier", verifier)

	req := httptest.NewRequest(http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Equal(t, "invalid_grant", errResp["error"])
	assert.Contains(t, errResp["error_description"], "redirect_uri")
}

func TestOAuthHandler_Token_ClientIDMismatch(t *testing.T) {
	router, clientRepo, authCodeRepo, _ := setupTokenRouter(t)

	verifier, challenge := generatePKCEPair()

	// Add another client
	clientRepo.clients["other_client"] = &entities.Client{
		ClientID:            "other_client",
		TenantID:            "tenant_xyz",
		ClientName:          "Other Client",
		ClientSecretHash:    "other-secret",
		AllowedRedirectURIs: []string{"https://other.example.com/callback"},
		IsActive:            true,
	}

	// Auth code was issued to test_client_123
	authCode := &entities.AuthorizationCode{
		Code:                "test-code-client-mismatch",
		ClientID:            "test_client_123",
		TenantID:            "tenant_xyz",
		UserID:              "user_123",
		RedirectURI:         "https://app.example.com/callback",
		Scope:               "openid profile email",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		ExpiresAt:           time.Now().Add(5 * time.Minute),
		CreatedAt:           time.Now(),
		UsedFlag:            false,
	}
	authCodeRepo.codes[authCode.Code] = authCode

	// Try to exchange with different client (FR-025)
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", "test-code-client-mismatch")
	form.Set("redirect_uri", "https://app.example.com/callback")
	form.Set("client_id", "other_client") // Different client
	form.Set("client_secret", "other-secret")
	form.Set("code_verifier", verifier)

	req := httptest.NewRequest(http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Equal(t, "invalid_grant", errResp["error"])
	assert.Contains(t, errResp["error_description"], "client_id")
}

func TestOAuthHandler_Token_BasicAuth(t *testing.T) {
	router, _, authCodeRepo, _ := setupTokenRouter(t)

	verifier, challenge := generatePKCEPair()

	authCode := &entities.AuthorizationCode{
		Code:                "test-code-basic-auth",
		ClientID:            "test_client_123",
		TenantID:            "tenant_xyz",
		UserID:              "user_123",
		RedirectURI:         "https://app.example.com/callback",
		Scope:               "openid profile email",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		ExpiresAt:           time.Now().Add(5 * time.Minute),
		CreatedAt:           time.Now(),
		UsedFlag:            false,
	}
	authCodeRepo.codes[authCode.Code] = authCode

	// Use Basic Auth instead of form parameters
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", "test-code-basic-auth")
	form.Set("redirect_uri", "https://app.example.com/callback")
	form.Set("code_verifier", verifier)
	// client_id and client_secret via Basic Auth

	req := httptest.NewRequest(http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("test_client_123", "test-secret")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var tokenResp auth.TokenResponse
	err := json.Unmarshal(w.Body.Bytes(), &tokenResp)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenResp.AccessToken)
}

func TestOAuthHandler_Token_CodeMarkedAsUsed(t *testing.T) {
	router, _, authCodeRepo, _ := setupTokenRouter(t)

	verifier, challenge := generatePKCEPair()

	authCode := &entities.AuthorizationCode{
		Code:                "test-code-mark-used",
		ClientID:            "test_client_123",
		TenantID:            "tenant_xyz",
		UserID:              "user_123",
		RedirectURI:         "https://app.example.com/callback",
		Scope:               "openid profile email",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		ExpiresAt:           time.Now().Add(5 * time.Minute),
		CreatedAt:           time.Now(),
		UsedFlag:            false,
	}
	authCodeRepo.codes[authCode.Code] = authCode

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", "test-code-mark-used")
	form.Set("redirect_uri", "https://app.example.com/callback")
	form.Set("client_id", "test_client_123")
	form.Set("client_secret", "test-secret")
	form.Set("code_verifier", verifier)

	// First request should succeed
	req := httptest.NewRequest(http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify code is marked as used (FR-026)
	storedCode := authCodeRepo.codes["test-code-mark-used"]
	assert.True(t, storedCode.UsedFlag)
	assert.NotNil(t, storedCode.UsedAt)

	// Second request with same code should fail (FR-024)
	req2 := httptest.NewRequest(http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)

	var errResp map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &errResp)
	assert.Equal(t, "invalid_grant", errResp["error"])
}

func TestOAuthHandler_Token_PublicClientWithoutSecret(t *testing.T) {
	router, clientRepo, authCodeRepo, _ := setupTokenRouter(t)
	clientRepo.clients["test_client_123"].ClientType = entities.ClientTypePublic
	clientRepo.clients["test_client_123"].ClientSecretHash = ""
	verifier, challenge := generatePKCEPair()
	authCodeRepo.codes["public-code"] = &entities.AuthorizationCode{
		Code: "public-code", ClientID: "test_client_123", TenantID: "tenant_xyz", UserID: "user_123",
		RedirectURI: "https://app.example.com/callback", Scope: "openid", CodeChallenge: challenge,
		CodeChallengeMethod: "S256", ExpiresAt: time.Now().Add(time.Minute),
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", "public-code")
	form.Set("redirect_uri", "https://app.example.com/callback")
	form.Set("client_id", "test_client_123")
	form.Set("code_verifier", verifier)

	req := httptest.NewRequest(http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOAuthHandler_Token_BrowserFlowPKCE(t *testing.T) {
	t.Run("matching verifier succeeds once", func(t *testing.T) {
		harness := newOAuthBrowserHarness(t)
		verifier, challenge := generatePKCEPair()
		code := issueBrowserAuthorizationCode(t, harness, challenge)

		response := exchangeBrowserAuthorizationCode(t, harness, code, verifier)
		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, http.StatusBadRequest, exchangeBrowserAuthorizationCode(t, harness, code, verifier).Code)
	})

	t.Run("wrong verifier is rejected", func(t *testing.T) {
		harness := newOAuthBrowserHarness(t)
		_, challenge := generatePKCEPair()
		wrongVerifier, _ := generatePKCEPair()
		code := issueBrowserAuthorizationCode(t, harness, challenge)

		assert.Equal(t, http.StatusBadRequest, exchangeBrowserAuthorizationCode(t, harness, code, wrongVerifier).Code)
	})

	t.Run("expired code is rejected", func(t *testing.T) {
		harness := newOAuthBrowserHarness(t)
		verifier, challenge := generatePKCEPair()
		code := issueBrowserAuthorizationCode(t, harness, challenge)
		harness.codes.codes[code].ExpiresAt = time.Now().Add(-time.Second)

		assert.Equal(t, http.StatusBadRequest, exchangeBrowserAuthorizationCode(t, harness, code, verifier).Code)
	})
}

func issueBrowserAuthorizationCode(t *testing.T, harness *oauthBrowserHarness, challenge string) string {
	t.Helper()
	_, transactionID := harness.beginWithChallenge(t, browserClientA, browserCallbackA, "", challenge, nil)
	loginResponse := harness.login(t, transactionID, "correct-password")
	require.Equal(t, http.StatusOK, loginResponse.Code)
	cookie := responseCookie(t, loginResponse)
	csrfToken := harness.transactions.transactions[transactionID].InteractionCSRFToken
	consentResponse := harness.consent(t, transactionID, csrfToken, true, cookie)
	require.Equal(t, http.StatusOK, consentResponse.Code)
	code := redirectQuery(t, consentResponse).Get("code")
	require.NotEmpty(t, code)
	return code
}

func exchangeBrowserAuthorizationCode(t *testing.T, harness *oauthBrowserHarness, code, verifier string) *httptest.ResponseRecorder {
	t.Helper()
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", browserCallbackA)
	form.Set("client_id", browserClientA)
	form.Set("code_verifier", verifier)
	request := httptest.NewRequest(http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, request)
	return recorder
}
