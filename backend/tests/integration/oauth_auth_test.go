/**
 * Integration tests for OAuth authorization flow
 */

package integration

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockIntegrationAuthCodeRepository for integration tests
type MockIntegrationAuthCodeRepository struct {
	codes map[string]*entities.AuthorizationCode
}

func NewMockIntegrationAuthCodeRepository() *MockIntegrationAuthCodeRepository {
	return &MockIntegrationAuthCodeRepository{
		codes: make(map[string]*entities.AuthorizationCode),
	}
}

func (m *MockIntegrationAuthCodeRepository) Store(ctx context.Context, code *entities.AuthorizationCode, ttl time.Duration) error {
	m.codes[code.Code] = code
	return nil
}

func (m *MockIntegrationAuthCodeRepository) Get(ctx context.Context, code string) (*entities.AuthorizationCode, error) {
	if c, ok := m.codes[code]; ok {
		return c, nil
	}
	return nil, errors.New("authorization code not found")
}

func (m *MockIntegrationAuthCodeRepository) MarkAsUsed(ctx context.Context, code string) error {
	if c, ok := m.codes[code]; ok {
		c.MarkAsUsed()
		return nil
	}
	return errors.New("authorization code not found")
}

func (m *MockIntegrationAuthCodeRepository) Delete(ctx context.Context, code string) error {
	delete(m.codes, code)
	return nil
}

func (m *MockIntegrationAuthCodeRepository) Exists(ctx context.Context, code string) (bool, error) {
	_, ok := m.codes[code]
	return ok, nil
}

var _ repositories.AuthCodeRepository = (*MockIntegrationAuthCodeRepository)(nil)

// MockIntegrationRoleRepository for integration tests
type MockIntegrationRoleRepository struct {
	roles map[string][]*entities.UserRoleAssignment
}

func NewMockIntegrationRoleRepository() *MockIntegrationRoleRepository {
	return &MockIntegrationRoleRepository{
		roles: make(map[string][]*entities.UserRoleAssignment),
	}
}

func (m *MockIntegrationRoleRepository) AssignRole(ctx context.Context, assignment *entities.UserRoleAssignment) error {
	key := assignment.UserID + ":" + assignment.ClientID
	m.roles[key] = append(m.roles[key], assignment)
	return nil
}

func (m *MockIntegrationRoleRepository) RevokeRole(ctx context.Context, userID, clientID, role string) error {
	return nil
}

func (m *MockIntegrationRoleRepository) GetUserRoles(ctx context.Context, userID, clientID string) ([]*entities.UserRoleAssignment, error) {
	key := userID + ":" + clientID
	return m.roles[key], nil
}

func (m *MockIntegrationRoleRepository) HasRole(ctx context.Context, userID, clientID, role string) (bool, error) {
	key := userID + ":" + clientID
	for _, r := range m.roles[key] {
		if r.Role == role && r.IsActive {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockIntegrationRoleRepository) HasAnyRole(ctx context.Context, userID, clientID string) (bool, error) {
	key := userID + ":" + clientID
	for _, r := range m.roles[key] {
		if r.IsActive {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockIntegrationRoleRepository) ListRolesByClient(ctx context.Context, clientID string) ([]*entities.UserRoleAssignment, error) {
	return nil, nil
}

func (m *MockIntegrationRoleRepository) ListRolesByUser(ctx context.Context, userID string) ([]*entities.UserRoleAssignment, error) {
	return nil, nil
}

var _ repositories.RoleRepository = (*MockIntegrationRoleRepository)(nil)

func setupOAuthRouter(t *testing.T) (*gin.Engine, *MockIntegrationClientRepository, *MockIntegrationRoleRepository, *MockIntegrationAuthCodeRepository) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	clientRepo := NewMockIntegrationClientRepository()
	roleRepo := NewMockIntegrationRoleRepository()
	authCodeRepo := NewMockIntegrationAuthCodeRepository()

	testClient := &entities.Client{
		ClientID:            "test_client_123",
		TenantID:            "tenant_xyz",
		ClientName:          "Test OAuth App",
		ClientSecretHash:    "hashed_secret",
		AllowedRedirectURIs: []string{"https://app.example.com/callback", "https://app.example.com/oauth"},
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	clientRepo.clients[testClient.ClientID] = testClient

	roleRepo.roles["user_123:test_client_123"] = []*entities.UserRoleAssignment{
		{
			UserID:   "user_123",
			ClientID: "test_client_123",
			Role:     "user",
			IsActive: true,
		},
	}

	authorizeClientUC := auth.NewAuthorizeClient(clientRepo, roleRepo, authCodeRepo)
	oauthHandler := handlers.NewOAuthHandler(authorizeClientUC, clientRepo)

	router.GET("/oauth2/auth", oauthHandler.Authorize)

	return router, clientRepo, roleRepo, authCodeRepo
}

func TestOAuthHandler_Authorize_Success(t *testing.T) {
	router, _, _, authCodeRepo := setupOAuthRouter(t)

	params := url.Values{}
	params.Set("client_id", "test_client_123")
	params.Set("redirect_uri", "https://app.example.com/callback")
	params.Set("response_type", "code")
	params.Set("scope", "openid profile email")
	params.Set("state", "csrf_token_abc123")
	params.Set("code_challenge", "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM")
	params.Set("code_challenge_method", "S256")

	req := httptest.NewRequest(http.MethodGet, "/oauth2/auth?"+params.Encode(), nil)
	req.Header.Set("X-User-ID", "user_123")
	req.Header.Set("X-Tenant-ID", "tenant_xyz")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)

	location := w.Header().Get("Location")
	require.NotEmpty(t, location)

	redirectURL, err := url.Parse(location)
	require.NoError(t, err)

	assert.NotEmpty(t, redirectURL.Query().Get("code"))
	assert.Equal(t, "csrf_token_abc123", redirectURL.Query().Get("state"))

	code := redirectURL.Query().Get("code")
	assert.True(t, len(authCodeRepo.codes) > 0)
	storedCode, err := authCodeRepo.Get(context.Background(), code)
	require.NoError(t, err)
	assert.Equal(t, "test_client_123", storedCode.ClientID)
	assert.Equal(t, "user_123", storedCode.UserID)
	assert.Equal(t, "openid profile email", storedCode.Scope)
}

func TestOAuthHandler_Authorize_InvalidClientID(t *testing.T) {
	router, _, _, _ := setupOAuthRouter(t)

	params := url.Values{}
	params.Set("client_id", "invalid_client")
	params.Set("redirect_uri", "https://app.example.com/callback")
	params.Set("response_type", "code")
	params.Set("scope", "openid")
	params.Set("state", "csrf_token")
	params.Set("code_challenge", "challenge")
	params.Set("code_challenge_method", "S256")

	req := httptest.NewRequest(http.MethodGet, "/oauth2/auth?"+params.Encode(), nil)
	req.Header.Set("X-User-ID", "user_123")
	req.Header.Set("X-Tenant-ID", "tenant_xyz")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_client")
}

func TestOAuthHandler_Authorize_InvalidRedirectURI(t *testing.T) {
	router, _, _, _ := setupOAuthRouter(t)

	params := url.Values{}
	params.Set("client_id", "test_client_123")
	params.Set("redirect_uri", "https://malicious.com/callback")
	params.Set("response_type", "code")
	params.Set("scope", "openid")
	params.Set("state", "csrf_token")
	params.Set("code_challenge", "challenge")
	params.Set("code_challenge_method", "S256")

	req := httptest.NewRequest(http.MethodGet, "/oauth2/auth?"+params.Encode(), nil)
	req.Header.Set("X-User-ID", "user_123")
	req.Header.Set("X-Tenant-ID", "tenant_xyz")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "redirect_uri")
}

func TestOAuthHandler_Authorize_UserWithoutRole(t *testing.T) {
	router, _, _, _ := setupOAuthRouter(t)

	params := url.Values{}
	params.Set("client_id", "test_client_123")
	params.Set("redirect_uri", "https://app.example.com/callback")
	params.Set("response_type", "code")
	params.Set("scope", "openid")
	params.Set("state", "csrf_token")
	params.Set("code_challenge", "challenge")
	params.Set("code_challenge_method", "S256")

	req := httptest.NewRequest(http.MethodGet, "/oauth2/auth?"+params.Encode(), nil)
	req.Header.Set("X-User-ID", "user_no_role")
	req.Header.Set("X-Tenant-ID", "tenant_xyz")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "access_denied")
}

func TestOAuthHandler_Authorize_MissingPKCE(t *testing.T) {
	router, _, _, _ := setupOAuthRouter(t)

	params := url.Values{}
	params.Set("client_id", "test_client_123")
	params.Set("redirect_uri", "https://app.example.com/callback")
	params.Set("response_type", "code")
	params.Set("scope", "openid")
	params.Set("state", "csrf_token")

	req := httptest.NewRequest(http.MethodGet, "/oauth2/auth?"+params.Encode(), nil)
	req.Header.Set("X-User-ID", "user_123")
	req.Header.Set("X-Tenant-ID", "tenant_xyz")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_request")
}

func TestOAuthHandler_Authorize_InvalidResponseType(t *testing.T) {
	router, _, _, _ := setupOAuthRouter(t)

	params := url.Values{}
	params.Set("client_id", "test_client_123")
	params.Set("redirect_uri", "https://app.example.com/callback")
	params.Set("response_type", "token")
	params.Set("scope", "openid")
	params.Set("state", "csrf_token")
	params.Set("code_challenge", "challenge")
	params.Set("code_challenge_method", "S256")

	req := httptest.NewRequest(http.MethodGet, "/oauth2/auth?"+params.Encode(), nil)
	req.Header.Set("X-User-ID", "user_123")
	req.Header.Set("X-Tenant-ID", "tenant_xyz")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "unsupported_response_type")
}

func TestOAuthHandler_Authorize_InvalidScope(t *testing.T) {
	router, _, _, _ := setupOAuthRouter(t)

	params := url.Values{}
	params.Set("client_id", "test_client_123")
	params.Set("redirect_uri", "https://app.example.com/callback")
	params.Set("response_type", "code")
	params.Set("scope", "profile email")
	params.Set("state", "csrf_token")
	params.Set("code_challenge", "challenge")
	params.Set("code_challenge_method", "S256")

	req := httptest.NewRequest(http.MethodGet, "/oauth2/auth?"+params.Encode(), nil)
	req.Header.Set("X-User-ID", "user_123")
	req.Header.Set("X-Tenant-ID", "tenant_xyz")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_scope")
}

func TestOAuthHandler_Authorize_MissingState(t *testing.T) {
	router, _, _, _ := setupOAuthRouter(t)

	params := url.Values{}
	params.Set("client_id", "test_client_123")
	params.Set("redirect_uri", "https://app.example.com/callback")
	params.Set("response_type", "code")
	params.Set("scope", "openid")
	params.Set("code_challenge", "challenge")
	params.Set("code_challenge_method", "S256")

	req := httptest.NewRequest(http.MethodGet, "/oauth2/auth?"+params.Encode(), nil)
	req.Header.Set("X-User-ID", "user_123")
	req.Header.Set("X-Tenant-ID", "tenant_xyz")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_request")
}

func TestOAuthHandler_Authorize_InactiveClient(t *testing.T) {
	router, clientRepo, _, _ := setupOAuthRouter(t)

	clientRepo.clients["test_client_123"].IsActive = false

	params := url.Values{}
	params.Set("client_id", "test_client_123")
	params.Set("redirect_uri", "https://app.example.com/callback")
	params.Set("response_type", "code")
	params.Set("scope", "openid")
	params.Set("state", "csrf_token")
	params.Set("code_challenge", "challenge")
	params.Set("code_challenge_method", "S256")

	req := httptest.NewRequest(http.MethodGet, "/oauth2/auth?"+params.Encode(), nil)
	req.Header.Set("X-User-ID", "user_123")
	req.Header.Set("X-Tenant-ID", "tenant_xyz")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized)
}

func TestOAuthHandler_Authorize_PlainPKCENotAllowed(t *testing.T) {
	router, _, _, _ := setupOAuthRouter(t)

	params := url.Values{}
	params.Set("client_id", "test_client_123")
	params.Set("redirect_uri", "https://app.example.com/callback")
	params.Set("response_type", "code")
	params.Set("scope", "openid")
	params.Set("state", "csrf_token")
	params.Set("code_challenge", "plain_verifier")
	params.Set("code_challenge_method", "plain")

	req := httptest.NewRequest(http.MethodGet, "/oauth2/auth?"+params.Encode(), nil)
	req.Header.Set("X-User-ID", "user_123")
	req.Header.Set("X-Tenant-ID", "tenant_xyz")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_request")
}
