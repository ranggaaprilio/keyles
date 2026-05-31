/**
 * Integration tests for OAuth token refresh (POST /oauth2/token with grant_type=refresh_token)
 * Tests FR-043 through FR-047 from spec.md
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

// MockClientRepositoryForRefreshTest implements ClientRepository for refresh token tests
type MockClientRepositoryForRefreshTest struct {
	clients map[string]*entities.Client
}

func NewMockClientRepositoryForRefreshTest() *MockClientRepositoryForRefreshTest {
	return &MockClientRepositoryForRefreshTest{
		clients: make(map[string]*entities.Client),
	}
}

func (m *MockClientRepositoryForRefreshTest) Create(ctx context.Context, client *entities.Client) error {
	m.clients[client.ClientID] = client
	return nil
}

func (m *MockClientRepositoryForRefreshTest) GetByID(ctx context.Context, id string) (*entities.Client, error) {
	for _, c := range m.clients {
		if c.ClientID == id {
			return c, nil
		}
	}
	return nil, errors.New("client not found")
}

func (m *MockClientRepositoryForRefreshTest) GetByClientID(ctx context.Context, clientID string, tenantID string) (*entities.Client, error) {
	if c, ok := m.clients[clientID]; ok {
		if c.TenantID == tenantID {
			return c, nil
		}
	}
	return nil, errors.New("client not found")
}

func (m *MockClientRepositoryForRefreshTest) Update(ctx context.Context, client *entities.Client) error {
	m.clients[client.ClientID] = client
	return nil
}

func (m *MockClientRepositoryForRefreshTest) Delete(ctx context.Context, clientID string) error {
	delete(m.clients, clientID)
	return nil
}

func (m *MockClientRepositoryForRefreshTest) ListByTenant(ctx context.Context, tenantID string) ([]*entities.Client, error) {
	return nil, nil
}

func (m *MockClientRepositoryForRefreshTest) ValidateCredentials(ctx context.Context, clientID string, clientSecret string) (*entities.Client, error) {
	if c, ok := m.clients[clientID]; ok {
		// For testing, we just check if client_secret matches the raw value stored
		if c.ClientSecretHash == clientSecret {
			return c, nil
		}
		return nil, errors.New("invalid client credentials")
	}
	return nil, errors.New("client not found")
}

func (m *MockClientRepositoryForRefreshTest) CountByTenant(ctx context.Context, tenantID string) (int, error) {
	return 0, nil
}

func (m *MockClientRepositoryForRefreshTest) ListByTenantPaginated(ctx context.Context, tenantID string, search string, page int, pageSize int) ([]*entities.Client, int, error) {
	return nil, 0, nil
}

var _ repositories.ClientRepository = (*MockClientRepositoryForRefreshTest)(nil)

// MockRefreshTokenRepositoryForTest implements RefreshTokenRepository for tests
type MockRefreshTokenRepositoryForTest struct {
	tokens map[string]*entities.RefreshToken
}

func NewMockRefreshTokenRepositoryForTest() *MockRefreshTokenRepositoryForTest {
	return &MockRefreshTokenRepositoryForTest{
		tokens: make(map[string]*entities.RefreshToken),
	}
}

func (m *MockRefreshTokenRepositoryForTest) Create(ctx context.Context, token *entities.RefreshToken) error {
	m.tokens[token.Token] = token
	return nil
}

func (m *MockRefreshTokenRepositoryForTest) GetByToken(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	if t, ok := m.tokens[tokenHash]; ok {
		return t, nil
	}
	return nil, errors.New("token not found")
}

func (m *MockRefreshTokenRepositoryForTest) Revoke(ctx context.Context, tokenHash string, revokedBy string) error {
	if t, ok := m.tokens[tokenHash]; ok {
		t.Revoke(revokedBy)
		return nil
	}
	return errors.New("token not found")
}

func (m *MockRefreshTokenRepositoryForTest) RevokeAllForUser(ctx context.Context, userID string, clientID string) error {
	for _, t := range m.tokens {
		if t.UserID == userID && t.ClientID == clientID {
			t.Revoke("user revocation")
		}
	}
	return nil
}

func (m *MockRefreshTokenRepositoryForTest) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *MockRefreshTokenRepositoryForTest) IsRevoked(ctx context.Context, tokenHash string) (bool, error) {
	if t, ok := m.tokens[tokenHash]; ok {
		return t.IsRevoked(), nil
	}
	return false, errors.New("token not found")
}

func (m *MockRefreshTokenRepositoryForTest) UpdateLastUsed(ctx context.Context, tokenHash string) error {
	if t, ok := m.tokens[tokenHash]; ok {
		t.MarkUsed()
		return nil
	}
	return nil
}

func (m *MockRefreshTokenRepositoryForTest) RevokeByClientID(ctx context.Context, clientID string) error {
	return nil
}

func (m *MockRefreshTokenRepositoryForTest) RevokeByUserID(ctx context.Context, userID string) error {
	return nil
}

func (m *MockRefreshTokenRepositoryForTest) ListByUserID(ctx context.Context, userID string) ([]*entities.RefreshToken, error) {
	tokens := make([]*entities.RefreshToken, 0)
	for _, t := range m.tokens {
		if t.UserID == userID {
			tokens = append(tokens, t)
		}
	}
	return tokens, nil
}

func (m *MockRefreshTokenRepositoryForTest) GetByID(ctx context.Context, id int64) (*entities.RefreshToken, error) {
	for _, t := range m.tokens {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, errors.New("token not found")
}

var _ repositories.RefreshTokenRepository = (*MockRefreshTokenRepositoryForTest)(nil)

// MockTokenServiceForRefreshTest implements TokenService for refresh tests
type MockTokenServiceForRefreshTest struct {
	privateKey *rsa.PrivateKey
	keyID      string
}

func NewMockTokenServiceForRefreshTest() (*MockTokenServiceForRefreshTest, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return &MockTokenServiceForRefreshTest{
		privateKey: privateKey,
		keyID:      "test-key-id",
	}, nil
}

func (m *MockTokenServiceForRefreshTest) SignIDToken(ctx context.Context, claims *services.TokenClaims) (string, error) {
	return "mock-id-token-" + claims.Subject, nil
}

func (m *MockTokenServiceForRefreshTest) SignAccessToken(ctx context.Context, claims *services.TokenClaims) (string, error) {
	return "mock-access-token-" + claims.Subject, nil
}

func (m *MockTokenServiceForRefreshTest) ValidateTokenSignature(ctx context.Context, token string) (*services.TokenClaims, error) {
	return nil, nil
}

func (m *MockTokenServiceForRefreshTest) GetPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	return &m.privateKey.PublicKey, nil
}

func (m *MockTokenServiceForRefreshTest) GetJWKS(ctx context.Context) (*services.JWKS, error) {
	return &services.JWKS{
		Keys: []services.JWK{
			{
				KeyType:   "RSA",
				KeyID:     m.keyID,
				Use:       "sig",
				Algorithm: "RS256",
				N:         base64.RawURLEncoding.EncodeToString(m.privateKey.PublicKey.N.Bytes()),
				E:         "AQAB",
			},
		},
	}, nil
}

func (m *MockTokenServiceForRefreshTest) GetActiveKeyID(ctx context.Context) (string, error) {
	return m.keyID, nil
}

var _ services.TokenService = (*MockTokenServiceForRefreshTest)(nil)

// hashRefreshTokenForTest hashes a refresh token for testing
func hashRefreshTokenForTest(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// setupRefreshTokenTestRouter creates a test router for refresh token tests
func setupRefreshTokenTestRouter(
	clientRepo *MockClientRepositoryForRefreshTest,
	refreshTokenRepo *MockRefreshTokenRepositoryForTest,
	tokenService *MockTokenServiceForRefreshTest,
) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	issuer := "https://auth.example.com"

	// Create use cases
	refreshTokenUC := auth.NewRefreshToken(refreshTokenRepo, clientRepo, tokenService, issuer)

	// Create handler with refresh token support
	oauthHandler := handlers.NewOAuthHandlerWithRefresh(nil, nil, clientRepo, refreshTokenUC)

	// Setup routes
	router.POST("/oauth2/token", oauthHandler.Token)

	return router
}

// TestRefreshTokenEndpoint_Success tests successful refresh token exchange (FR-043)
func TestRefreshTokenEndpoint_Success(t *testing.T) {
	// Setup
	clientRepo := NewMockClientRepositoryForRefreshTest()
	refreshTokenRepo := NewMockRefreshTokenRepositoryForTest()
	tokenService, err := NewMockTokenServiceForRefreshTest()
	require.NoError(t, err)

	// Create test client
	clientID := "test-client-id"
	clientSecret := "test-client-secret"
	tenantID := "test-tenant-id"
	userID := "test-user-id"

	client := &entities.Client{
		ClientID:         clientID,
		TenantID:         tenantID,
		ClientName:       "Test Client",
		ClientSecretHash: clientSecret, // In tests, we use raw secret
		IsActive:         true,
	}
	clientRepo.clients[clientID] = client

	// Create a valid refresh token
	rawRefreshToken := "valid-refresh-token-12345"
	tokenHash := hashRefreshTokenForTest(rawRefreshToken)
	storedToken := &entities.RefreshToken{
		ID:        1,
		Token:     tokenHash,
		UserID:    userID,
		ClientID:  clientID,
		TenantID:  tenantID,
		Scope:     "openid profile email",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}
	refreshTokenRepo.tokens[tokenHash] = storedToken

	router := setupRefreshTokenTestRouter(clientRepo, refreshTokenRepo, tokenService)

	// Create request
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", rawRefreshToken)
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)

	req, _ := http.NewRequest("POST", "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response["access_token"])
	assert.Equal(t, "Bearer", response["token_type"])
	assert.Equal(t, float64(900), response["expires_in"]) // 15 minutes
	assert.Equal(t, "openid profile email", response["scope"])
}

// TestRefreshTokenEndpoint_InvalidRefreshToken tests invalid refresh token (FR-046)
func TestRefreshTokenEndpoint_InvalidRefreshToken(t *testing.T) {
	// Setup
	clientRepo := NewMockClientRepositoryForRefreshTest()
	refreshTokenRepo := NewMockRefreshTokenRepositoryForTest()
	tokenService, err := NewMockTokenServiceForRefreshTest()
	require.NoError(t, err)

	// Create test client
	clientID := "test-client-id"
	clientSecret := "test-client-secret"

	client := &entities.Client{
		ClientID:         clientID,
		TenantID:         "test-tenant-id",
		ClientSecretHash: clientSecret,
		IsActive:         true,
	}
	clientRepo.clients[clientID] = client

	router := setupRefreshTokenTestRouter(clientRepo, refreshTokenRepo, tokenService)

	// Create request with non-existent refresh token
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", "invalid-token")
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)

	req, _ := http.NewRequest("POST", "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "invalid_grant", response["error"])
}

// TestRefreshTokenEndpoint_RevokedToken tests revoked refresh token rejection (FR-046)
func TestRefreshTokenEndpoint_RevokedToken(t *testing.T) {
	// Setup
	clientRepo := NewMockClientRepositoryForRefreshTest()
	refreshTokenRepo := NewMockRefreshTokenRepositoryForTest()
	tokenService, err := NewMockTokenServiceForRefreshTest()
	require.NoError(t, err)

	// Create test client
	clientID := "test-client-id"
	clientSecret := "test-client-secret"
	tenantID := "test-tenant-id"

	client := &entities.Client{
		ClientID:         clientID,
		TenantID:         tenantID,
		ClientSecretHash: clientSecret,
		IsActive:         true,
	}
	clientRepo.clients[clientID] = client

	// Create a revoked refresh token
	rawRefreshToken := "revoked-refresh-token"
	tokenHash := hashRefreshTokenForTest(rawRefreshToken)
	revokedAt := time.Now().Add(-1 * time.Hour)
	storedToken := &entities.RefreshToken{
		ID:          1,
		Token:       tokenHash,
		UserID:      "test-user-id",
		ClientID:    clientID,
		TenantID:    tenantID,
		Scope:       "openid",
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		RevokedFlag: true,
		RevokedAt:   &revokedAt,
	}
	refreshTokenRepo.tokens[tokenHash] = storedToken

	router := setupRefreshTokenTestRouter(clientRepo, refreshTokenRepo, tokenService)

	// Create request
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", rawRefreshToken)
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)

	req, _ := http.NewRequest("POST", "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - revoked token should fail with invalid_grant (FR-050)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "invalid_grant", response["error"])
	assert.Contains(t, response["error_description"], "revoked")
}

// TestRefreshTokenEndpoint_ClientIDMismatch tests client_id validation (FR-047)
func TestRefreshTokenEndpoint_ClientIDMismatch(t *testing.T) {
	// Setup
	clientRepo := NewMockClientRepositoryForRefreshTest()
	refreshTokenRepo := NewMockRefreshTokenRepositoryForTest()
	tokenService, err := NewMockTokenServiceForRefreshTest()
	require.NoError(t, err)

	// Create test clients - two different clients
	originalClientID := "original-client"
	attackerClientID := "attacker-client"
	clientSecret := "test-client-secret"
	tenantID := "test-tenant-id"

	originalClient := &entities.Client{
		ClientID:         originalClientID,
		TenantID:         tenantID,
		ClientSecretHash: clientSecret,
		IsActive:         true,
	}
	attackerClient := &entities.Client{
		ClientID:         attackerClientID,
		TenantID:         tenantID,
		ClientSecretHash: clientSecret,
		IsActive:         true,
	}
	clientRepo.clients[originalClientID] = originalClient
	clientRepo.clients[attackerClientID] = attackerClient

	// Create refresh token for original client
	rawRefreshToken := "valid-refresh-token"
	tokenHash := hashRefreshTokenForTest(rawRefreshToken)
	storedToken := &entities.RefreshToken{
		ID:        1,
		Token:     tokenHash,
		UserID:    "test-user-id",
		ClientID:  originalClientID, // Token belongs to original client
		TenantID:  tenantID,
		Scope:     "openid",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}
	refreshTokenRepo.tokens[tokenHash] = storedToken

	router := setupRefreshTokenTestRouter(clientRepo, refreshTokenRepo, tokenService)

	// Create request using attacker client's credentials with original client's refresh token
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", rawRefreshToken)
	form.Set("client_id", attackerClientID) // Attacker trying to use stolen refresh token
	form.Set("client_secret", clientSecret)

	req, _ := http.NewRequest("POST", "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - should fail because client_id doesn't match (FR-047)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "invalid_grant", response["error"])
	assert.Contains(t, response["error_description"], "client_id")
}

// TestRefreshTokenEndpoint_InvalidClientCredentials tests client authentication failure
func TestRefreshTokenEndpoint_InvalidClientCredentials(t *testing.T) {
	// Setup
	clientRepo := NewMockClientRepositoryForRefreshTest()
	refreshTokenRepo := NewMockRefreshTokenRepositoryForTest()
	tokenService, err := NewMockTokenServiceForRefreshTest()
	require.NoError(t, err)

	// Create test client
	clientID := "test-client-id"
	clientSecret := "correct-secret"
	tenantID := "test-tenant-id"

	client := &entities.Client{
		ClientID:         clientID,
		TenantID:         tenantID,
		ClientSecretHash: clientSecret,
		IsActive:         true,
	}
	clientRepo.clients[clientID] = client

	// Create a valid refresh token
	rawRefreshToken := "valid-refresh-token"
	tokenHash := hashRefreshTokenForTest(rawRefreshToken)
	storedToken := &entities.RefreshToken{
		ID:        1,
		Token:     tokenHash,
		UserID:    "test-user-id",
		ClientID:  clientID,
		TenantID:  tenantID,
		Scope:     "openid",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}
	refreshTokenRepo.tokens[tokenHash] = storedToken

	router := setupRefreshTokenTestRouter(clientRepo, refreshTokenRepo, tokenService)

	// Create request with wrong client secret
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", rawRefreshToken)
	form.Set("client_id", clientID)
	form.Set("client_secret", "wrong-secret") // Wrong secret!

	req, _ := http.NewRequest("POST", "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "invalid_client", response["error"])
}

// TestRefreshTokenEndpoint_MissingParameters tests required parameter validation
func TestRefreshTokenEndpoint_MissingParameters(t *testing.T) {
	// Setup
	clientRepo := NewMockClientRepositoryForRefreshTest()
	refreshTokenRepo := NewMockRefreshTokenRepositoryForTest()
	tokenService, err := NewMockTokenServiceForRefreshTest()
	require.NoError(t, err)

	router := setupRefreshTokenTestRouter(clientRepo, refreshTokenRepo, tokenService)

	testCases := []struct {
		name     string
		formData url.Values
		expected string
	}{
		{
			name: "missing refresh_token",
			formData: url.Values{
				"grant_type":    {"refresh_token"},
				"client_id":     {"client"},
				"client_secret": {"secret"},
			},
			expected: "refresh_token",
		},
		{
			name: "missing client_id",
			formData: url.Values{
				"grant_type":    {"refresh_token"},
				"refresh_token": {"token"},
				"client_secret": {"secret"},
			},
			expected: "client_id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/oauth2/token", strings.NewReader(tc.formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "invalid_request", response["error"])
			assert.Contains(t, response["error_description"], tc.expected)
		})
	}
}

// TestRefreshTokenEndpoint_BasicAuth tests client authentication via Basic Auth header
func TestRefreshTokenEndpoint_BasicAuth(t *testing.T) {
	// Setup
	clientRepo := NewMockClientRepositoryForRefreshTest()
	refreshTokenRepo := NewMockRefreshTokenRepositoryForTest()
	tokenService, err := NewMockTokenServiceForRefreshTest()
	require.NoError(t, err)

	// Create test client
	clientID := "test-client-id"
	clientSecret := "test-client-secret"
	tenantID := "test-tenant-id"
	userID := "test-user-id"

	client := &entities.Client{
		ClientID:         clientID,
		TenantID:         tenantID,
		ClientName:       "Test Client",
		ClientSecretHash: clientSecret,
		IsActive:         true,
	}
	clientRepo.clients[clientID] = client

	// Create a valid refresh token
	rawRefreshToken := "valid-refresh-token"
	tokenHash := hashRefreshTokenForTest(rawRefreshToken)
	storedToken := &entities.RefreshToken{
		ID:        1,
		Token:     tokenHash,
		UserID:    userID,
		ClientID:  clientID,
		TenantID:  tenantID,
		Scope:     "openid profile",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}
	refreshTokenRepo.tokens[tokenHash] = storedToken

	router := setupRefreshTokenTestRouter(clientRepo, refreshTokenRepo, tokenService)

	// Create request with Basic Auth (no client_id/client_secret in body)
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", rawRefreshToken)

	req, _ := http.NewRequest("POST", "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response["access_token"])
	assert.Equal(t, "Bearer", response["token_type"])
}

// TestRefreshTokenEndpoint_SubsequentRefreshAfterRevocation tests that subsequent
// refresh attempts fail after token revocation (FR-050)
func TestRefreshTokenEndpoint_SubsequentRefreshAfterRevocation(t *testing.T) {
	// Setup
	clientRepo := NewMockClientRepositoryForRefreshTest()
	refreshTokenRepo := NewMockRefreshTokenRepositoryForTest()
	tokenService, err := NewMockTokenServiceForRefreshTest()
	require.NoError(t, err)

	// Create test client
	clientID := "test-client-id"
	clientSecret := "test-client-secret"
	tenantID := "test-tenant-id"

	client := &entities.Client{
		ClientID:         clientID,
		TenantID:         tenantID,
		ClientSecretHash: clientSecret,
		IsActive:         true,
	}
	clientRepo.clients[clientID] = client

	// Create a refresh token that will be revoked
	rawRefreshToken := "token-to-revoke"
	tokenHash := hashRefreshTokenForTest(rawRefreshToken)
	storedToken := &entities.RefreshToken{
		ID:        1,
		Token:     tokenHash,
		UserID:    "test-user-id",
		ClientID:  clientID,
		TenantID:  tenantID,
		Scope:     "openid",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}
	refreshTokenRepo.tokens[tokenHash] = storedToken

	router := setupRefreshTokenTestRouter(clientRepo, refreshTokenRepo, tokenService)

	// First refresh should succeed
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", rawRefreshToken)
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)

	req, _ := http.NewRequest("POST", "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Now revoke the token
	refreshTokenRepo.Revoke(context.Background(), tokenHash, "admin")

	// Second refresh should fail
	req2, _ := http.NewRequest("POST", "/oauth2/token", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "invalid_grant", response["error"])
}
