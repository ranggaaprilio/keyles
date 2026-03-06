/**
 * Integration tests for client management handler (JWT protected endpoints)
 */

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	domainServices "github.com/ranggaaprilio/keyles/domain/services"
	infraServices "github.com/ranggaaprilio/keyles/infrastructure/services"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/interfaces/http/middleware"
	"github.com/ranggaaprilio/keyles/usecase/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockIntegrationClientRepository for integration tests
type MockIntegrationClientRepository struct {
	clients map[string]*entities.Client
}

func NewMockIntegrationClientRepository() *MockIntegrationClientRepository {
	return &MockIntegrationClientRepository{
		clients: make(map[string]*entities.Client),
	}
}

func (m *MockIntegrationClientRepository) Create(ctx context.Context, c *entities.Client) error {
	m.clients[c.ClientID] = c
	return nil
}

func (m *MockIntegrationClientRepository) GetByID(ctx context.Context, clientID string) (*entities.Client, error) {
	if c, ok := m.clients[clientID]; ok && c.IsActive {
		return c, nil
	}
	return nil, errors.New("client not found")
}

func (m *MockIntegrationClientRepository) GetByClientID(ctx context.Context, clientID string, tenantID string) (*entities.Client, error) {
	if c, ok := m.clients[clientID]; ok && c.TenantID == tenantID && c.IsActive {
		return c, nil
	}
	return nil, errors.New("client not found")
}

func (m *MockIntegrationClientRepository) Update(ctx context.Context, c *entities.Client) error {
	if _, ok := m.clients[c.ClientID]; ok {
		m.clients[c.ClientID] = c
		return nil
	}
	return errors.New("client not found")
}

func (m *MockIntegrationClientRepository) Delete(ctx context.Context, clientID string) error {
	if c, ok := m.clients[clientID]; ok {
		c.IsActive = false
		return nil
	}
	return errors.New("client not found")
}

func (m *MockIntegrationClientRepository) ListByTenant(ctx context.Context, tenantID string) ([]*entities.Client, error) {
	var result []*entities.Client
	for _, c := range m.clients {
		if c.TenantID == tenantID && c.IsActive {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *MockIntegrationClientRepository) ValidateCredentials(ctx context.Context, clientID string, clientSecret string) (*entities.Client, error) {
	return nil, errors.New("not implemented")
}

func (m *MockIntegrationClientRepository) CountByTenant(ctx context.Context, tenantID string) (int, error) {
	count := 0
	for _, c := range m.clients {
		if c.TenantID == tenantID && c.IsActive {
			count++
		}
	}
	return count, nil
}

func (m *MockIntegrationClientRepository) ListByTenantPaginated(ctx context.Context, tenantID string, search string, page int, pageSize int) ([]*entities.Client, int, error) {
	var filtered []*entities.Client
	for _, c := range m.clients {
		if c.TenantID == tenantID && c.IsActive {
			if search == "" || strings.Contains(strings.ToLower(c.ClientName), strings.ToLower(search)) {
				filtered = append(filtered, c)
			}
		}
	}
	// Sort by created_at descending for deterministic results
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].ClientName < filtered[j].ClientName
	})
	total := len(filtered)
	offset := (page - 1) * pageSize
	if offset >= total {
		return []*entities.Client{}, total, nil
	}
	end := offset + pageSize
	if end > total {
		end = total
	}
	return filtered[offset:end], total, nil
}

// Ensure MockIntegrationClientRepository implements repositories.ClientRepository
var _ repositories.ClientRepository = (*MockIntegrationClientRepository)(nil)

// MockIntegrationAuditRepository for integration tests
type MockIntegrationAuditRepository struct {
	logs []*entities.AuditLog
}

func (m *MockIntegrationAuditRepository) Create(ctx context.Context, log *entities.AuditLog) error {
	m.logs = append(m.logs, log)
	return nil
}

func (m *MockIntegrationAuditRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entities.AuditLog, error) {
	return nil, nil
}

func (m *MockIntegrationAuditRepository) FindByEventType(ctx context.Context, eventType entities.EventType, limit, offset int) ([]*entities.AuditLog, error) {
	return nil, nil
}

func (m *MockIntegrationAuditRepository) FindRecent(ctx context.Context, limit int) ([]*entities.AuditLog, error) {
	return nil, nil
}

var _ repositories.AuditRepository = (*MockIntegrationAuditRepository)(nil)

// MockIntegrationRefreshTokenRepository for integration tests
type MockIntegrationRefreshTokenRepository struct {
	revokedClientIDs []string
}

func (m *MockIntegrationRefreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	return nil
}
func (m *MockIntegrationRefreshTokenRepository) GetByToken(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	return nil, errors.New("not found")
}
func (m *MockIntegrationRefreshTokenRepository) Revoke(ctx context.Context, tokenHash string, revokedBy string) error {
	return nil
}
func (m *MockIntegrationRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string, clientID string) error {
	return nil
}
func (m *MockIntegrationRefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}
func (m *MockIntegrationRefreshTokenRepository) IsRevoked(ctx context.Context, tokenHash string) (bool, error) {
	return false, nil
}
func (m *MockIntegrationRefreshTokenRepository) UpdateLastUsed(ctx context.Context, tokenHash string) error {
	return nil
}
func (m *MockIntegrationRefreshTokenRepository) RevokeByClientID(ctx context.Context, clientID string) error {
	m.revokedClientIDs = append(m.revokedClientIDs, clientID)
	return nil
}

func (m *MockIntegrationRefreshTokenRepository) RevokeByUserID(ctx context.Context, userID string) error {
	return nil
}
func (m *MockIntegrationRefreshTokenRepository) ListByUserID(ctx context.Context, userID string) ([]*entities.RefreshToken, error) {
	return []*entities.RefreshToken{}, nil
}
func (m *MockIntegrationRefreshTokenRepository) GetByID(ctx context.Context, id int64) (*entities.RefreshToken, error) {
	return nil, errors.New("not found")
}

var _ repositories.RefreshTokenRepository = (*MockIntegrationRefreshTokenRepository)(nil)

// MockIntegrationClientCountCache for integration tests
type MockIntegrationClientCountCache struct{}

func (m *MockIntegrationClientCountCache) Get(ctx context.Context, tenantID string) (int, error) {
	return -1, nil
}
func (m *MockIntegrationClientCountCache) Set(ctx context.Context, tenantID string, count int) error {
	return nil
}
func (m *MockIntegrationClientCountCache) Invalidate(ctx context.Context, tenantID string) error {
	return nil
}

var _ domainServices.ClientCountCache = (*MockIntegrationClientCountCache)(nil)

// MockIntegrationRevokedClientCache for integration tests
type MockIntegrationRevokedClientCache struct {
	revoked map[string]bool
}

func (m *MockIntegrationRevokedClientCache) Revoke(ctx context.Context, clientID string) error {
	m.revoked[clientID] = true
	return nil
}
func (m *MockIntegrationRevokedClientCache) IsRevoked(ctx context.Context, clientID string) (bool, error) {
	return m.revoked[clientID], nil
}

var _ domainServices.RevokedClientCache = (*MockIntegrationRevokedClientCache)(nil)

// MockPasswordServiceIntegration for integration tests
type MockPasswordServiceIntegration struct{}

func (m *MockPasswordServiceIntegration) Hash(password string) (string, error) {
	return "hashed_" + password, nil
}

func (m *MockPasswordServiceIntegration) Verify(password, hash string) error {
	if hash == "hashed_"+password {
		return nil
	}
	return errors.New("password mismatch")
}

// Ensure MockPasswordServiceIntegration implements services.PasswordService
var _ domainServices.PasswordService = (*MockPasswordServiceIntegration)(nil)

func setupClientRouter(t *testing.T) (*gin.Engine, *infraServices.JWTService, *MockIntegrationClientRepository) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtService := infraServices.NewJWTService("test-secret-key-32-characters!", 24)
	clientRepo := NewMockIntegrationClientRepository()
	passwordService := &MockPasswordServiceIntegration{}
	auditRepo := &MockIntegrationAuditRepository{}
	refreshTokenRepo := &MockIntegrationRefreshTokenRepository{}
	countCache := &MockIntegrationClientCountCache{}
	revokedCache := &MockIntegrationRevokedClientCache{revoked: make(map[string]bool)}

	// Create use cases with updated constructors
	createClientUC := client.NewCreateClientUseCase(clientRepo, passwordService, auditRepo, countCache)
	getClientUC := client.NewGetClientUseCase(clientRepo)
	updateClientUC := client.NewUpdateClientUseCase(clientRepo, auditRepo)
	deleteClientUC := client.NewDeleteClientUseCase(clientRepo, auditRepo, refreshTokenRepo, revokedCache, countCache)
	listClientsUC := client.NewListClientsUseCase(clientRepo)
	rotateSecretUC := client.NewRotateSecretUseCase(clientRepo, passwordService, auditRepo)

	// Create handler
	clientHandler := handlers.NewClientHandler(
		createClientUC,
		getClientUC,
		updateClientUC,
		deleteClientUC,
		listClientsUC,
		rotateSecretUC,
	)

	// Setup routes
	admin := router.Group("/api/v1/admin")
	admin.Use(middleware.AuthMiddleware(jwtService))
	{
		clients := admin.Group("/clients")
		{
			clients.POST("", clientHandler.Create)
			clients.GET("", clientHandler.List)
			clients.GET("/:clientId", clientHandler.Get)
			clients.PUT("/:clientId", clientHandler.Update)
			clients.DELETE("/:clientId", clientHandler.Delete)
			clients.POST("/:clientId/rotate-secret", clientHandler.RotateSecret)
		}
	}

	return router, jwtService, clientRepo
}

func TestClientHandler_CreateClient_Success(t *testing.T) {
	router, jwtService, _ := setupClientRouter(t)

	// Generate a valid JWT token
	token, _ := jwtService.GenerateToken("user-123", "tenant-456", "admin@example.com", "admin")

	// Create request body - confidential client
	reqBody := map[string]interface{}{
		"client_name":   "Test Application",
		"client_type":   "confidential",
		"redirect_uris": []string{"https://app.example.com/callback"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/clients", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.NotEmpty(t, resp["client_id"])
	assert.NotEmpty(t, resp["client_secret"])
	assert.Equal(t, "Test Application", resp["client_name"])
	assert.Equal(t, "confidential", resp["client_type"])
	assert.True(t, resp["is_active"].(bool))
}

func TestClientHandler_CreateClient_InvalidRequest(t *testing.T) {
	router, jwtService, _ := setupClientRouter(t)

	token, _ := jwtService.GenerateToken("user-123", "tenant-456", "admin@example.com", "admin")

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
	}{
		{
			name:       "Missing client_name",
			body:       map[string]interface{}{"client_type": "confidential", "redirect_uris": []string{"https://app.example.com/callback"}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Missing redirect_uris",
			body:       map[string]interface{}{"client_name": "Test App", "client_type": "confidential"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Empty redirect_uris",
			body:       map[string]interface{}{"client_name": "Test App", "client_type": "confidential", "redirect_uris": []string{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Missing client_type",
			body:       map[string]interface{}{"client_name": "Test App", "redirect_uris": []string{"https://app.example.com/callback"}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid client_type",
			body:       map[string]interface{}{"client_name": "Test App", "client_type": "invalid", "redirect_uris": []string{"https://app.example.com/callback"}},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/clients", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestClientHandler_CreateClient_Unauthorized(t *testing.T) {
	router, _, _ := setupClientRouter(t)

	reqBody := map[string]interface{}{
		"client_name":   "Test Application",
		"client_type":   "confidential",
		"redirect_uris": []string{"https://app.example.com/callback"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/clients", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestClientHandler_CreateClient_PublicClient(t *testing.T) {
	router, jwtService, _ := setupClientRouter(t)

	token, _ := jwtService.GenerateToken("user-123", "tenant-456", "admin@example.com", "admin")

	reqBody := map[string]interface{}{
		"client_name":   "Public SPA",
		"client_type":   "public",
		"redirect_uris": []string{"https://spa.example.com/callback"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/clients", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.NotEmpty(t, resp["client_id"])
	assert.Nil(t, resp["client_secret"])
	assert.Equal(t, "public", resp["client_type"])
}

func TestClientHandler_CreateClient_WithDescription(t *testing.T) {
	router, jwtService, _ := setupClientRouter(t)

	token, _ := jwtService.GenerateToken("user-123", "tenant-456", "admin@example.com", "admin")

	reqBody := map[string]interface{}{
		"client_name":   "My App",
		"client_type":   "confidential",
		"description":   "A test application for integration testing",
		"redirect_uris": []string{"https://myapp.example.com/callback"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/clients", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A test application for integration testing", resp["description"])
}

func TestClientHandler_ListClients_Success(t *testing.T) {
	router, jwtService, repo := setupClientRouter(t)

	// Pre-populate with test clients
	repo.clients["client-1"] = &entities.Client{
		ClientID:            "client-1",
		TenantID:            "tenant-456",
		ClientName:          "App One",
		ClientType:          "confidential",
		Description:         "First application",
		AllowedRedirectURIs: []string{"https://app1.example.com/callback"},
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	repo.clients["client-2"] = &entities.Client{
		ClientID:            "client-2",
		TenantID:            "tenant-456",
		ClientName:          "App Two",
		ClientType:          "public",
		AllowedRedirectURIs: []string{"https://app2.example.com/callback"},
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	// Different tenant - should not appear in results
	repo.clients["client-3"] = &entities.Client{
		ClientID:            "client-3",
		TenantID:            "other-tenant",
		ClientName:          "Other App",
		AllowedRedirectURIs: []string{"https://other.example.com/callback"},
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	token, _ := jwtService.GenerateToken("user-123", "tenant-456", "admin@example.com", "admin")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/clients", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	clients := resp["clients"].([]interface{})
	assert.Equal(t, 2, len(clients)) // Only tenant-456's clients
	assert.Equal(t, float64(2), resp["total"])
	assert.NotNil(t, resp["page"])
	assert.NotNil(t, resp["page_size"])
	assert.NotNil(t, resp["total_pages"])
}

func TestClientHandler_ListClients_Search(t *testing.T) {
	router, jwtService, repo := setupClientRouter(t)

	repo.clients["client-1"] = &entities.Client{
		ClientID:   "client-1",
		TenantID:   "tenant-456",
		ClientName: "Backend API",
		ClientType: "confidential",
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	repo.clients["client-2"] = &entities.Client{
		ClientID:   "client-2",
		TenantID:   "tenant-456",
		ClientName: "Frontend SPA",
		ClientType: "public",
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	token, _ := jwtService.GenerateToken("user-123", "tenant-456", "admin@example.com", "admin")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/clients?search=Backend", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	clients := resp["clients"].([]interface{})
	assert.Equal(t, 1, len(clients))
	assert.Equal(t, float64(1), resp["total"])
}

func TestClientHandler_GetClient_Success(t *testing.T) {
	router, jwtService, repo := setupClientRouter(t)

	// Pre-populate with test client
	repo.clients["client-123"] = &entities.Client{
		ClientID:            "client-123",
		TenantID:            "tenant-456",
		ClientName:          "Test App",
		ClientType:          "confidential",
		Description:         "A client for testing",
		AllowedRedirectURIs: []string{"https://app.example.com/callback"},
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	token, _ := jwtService.GenerateToken("user-123", "tenant-456", "admin@example.com", "admin")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/clients/client-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "client-123", resp["client_id"])
	assert.Equal(t, "Test App", resp["client_name"])
	assert.Equal(t, "confidential", resp["client_type"])
	assert.Equal(t, "A client for testing", resp["description"])
}

func TestClientHandler_GetClient_NotFound(t *testing.T) {
	router, jwtService, _ := setupClientRouter(t)

	token, _ := jwtService.GenerateToken("user-123", "tenant-456", "admin@example.com", "admin")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/clients/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestClientHandler_GetClient_WrongTenant(t *testing.T) {
	router, jwtService, repo := setupClientRouter(t)

	// Pre-populate with test client for different tenant
	repo.clients["client-123"] = &entities.Client{
		ClientID:            "client-123",
		TenantID:            "other-tenant",
		ClientName:          "Test App",
		AllowedRedirectURIs: []string{"https://app.example.com/callback"},
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	token, _ := jwtService.GenerateToken("user-123", "tenant-456", "admin@example.com", "admin")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/clients/client-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 404 because client belongs to different tenant
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestClientHandler_UpdateClient_Success(t *testing.T) {
	router, jwtService, repo := setupClientRouter(t)

	// Pre-populate with test client
	repo.clients["client-123"] = &entities.Client{
		ClientID:            "client-123",
		TenantID:            "tenant-456",
		ClientName:          "Original Name",
		AllowedRedirectURIs: []string{"https://old.example.com/callback"},
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	token, _ := jwtService.GenerateToken("user-123", "tenant-456", "admin@example.com", "admin")

	reqBody := map[string]interface{}{
		"client_name": "Updated Name",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/clients/client-123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "Updated Name", resp["client_name"])
}

func TestClientHandler_DeleteClient_Success(t *testing.T) {
	router, jwtService, repo := setupClientRouter(t)

	// Pre-populate with test client
	repo.clients["client-123"] = &entities.Client{
		ClientID:            "client-123",
		TenantID:            "tenant-456",
		ClientName:          "To Be Deleted",
		AllowedRedirectURIs: []string{"https://app.example.com/callback"},
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	token, _ := jwtService.GenerateToken("user-123", "tenant-456", "admin@example.com", "admin")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/clients/client-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify client is soft-deleted
	assert.False(t, repo.clients["client-123"].IsActive)
}

func TestClientHandler_RotateSecret_Success(t *testing.T) {
	router, jwtService, repo := setupClientRouter(t)

	// Pre-populate with confidential test client
	repo.clients["client-123"] = &entities.Client{
		ClientID:            "client-123",
		TenantID:            "tenant-456",
		ClientName:          "Test App",
		ClientType:          "confidential",
		ClientSecretHash:    "old-hashed-secret",
		AllowedRedirectURIs: []string{"https://app.example.com/callback"},
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	token, _ := jwtService.GenerateToken("user-123", "tenant-456", "admin@example.com", "admin")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/clients/client-123/rotate-secret", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "client-123", resp["client_id"])
	assert.NotEmpty(t, resp["client_secret"])
	assert.NotEmpty(t, resp["rotated_at"])

	// Verify secret was actually rotated
	assert.NotEqual(t, "old-hashed-secret", repo.clients["client-123"].ClientSecretHash)
}

func TestClientHandler_RotateSecret_PublicClientRejected(t *testing.T) {
	router, jwtService, repo := setupClientRouter(t)

	// Pre-populate with public client
	repo.clients["client-pub"] = &entities.Client{
		ClientID:            "client-pub",
		TenantID:            "tenant-456",
		ClientName:          "Public SPA",
		ClientType:          "public",
		AllowedRedirectURIs: []string{"https://spa.example.com/callback"},
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	token, _ := jwtService.GenerateToken("user-123", "tenant-456", "admin@example.com", "admin")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/clients/client-pub/rotate-secret", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestClientHandler_RotateSecret_NotFound(t *testing.T) {
	router, jwtService, _ := setupClientRouter(t)

	token, _ := jwtService.GenerateToken("user-123", "tenant-456", "admin@example.com", "admin")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/clients/nonexistent/rotate-secret", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestClientHandler_DeleteClient_NotFoundAfterDelete(t *testing.T) {
	router, jwtService, repo := setupClientRouter(t)

	repo.clients["client-del"] = &entities.Client{
		ClientID:   "client-del",
		TenantID:   "tenant-456",
		ClientName: "Delete Me",
		ClientType: "confidential",
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	token, _ := jwtService.GenerateToken("user-123", "tenant-456", "admin@example.com", "admin")

	// Delete the client
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/clients/client-del", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Get should now return 404
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/admin/clients/client-del", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusNotFound, w2.Code)
}
