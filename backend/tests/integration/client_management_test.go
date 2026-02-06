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
	"testing"
	"time"

	"github.com/gin-gonic/gin"
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

// Ensure MockIntegrationClientRepository implements repositories.ClientRepository
var _ repositories.ClientRepository = (*MockIntegrationClientRepository)(nil)

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

	// Create use cases
	createClientUC := client.NewCreateClientUseCase(clientRepo, passwordService)
	getClientUC := client.NewGetClientUseCase(clientRepo)
	updateClientUC := client.NewUpdateClientUseCase(clientRepo)
	deleteClientUC := client.NewDeleteClientUseCase(clientRepo)
	listClientsUC := client.NewListClientsUseCase(clientRepo)
	rotateSecretUC := client.NewRotateSecretUseCase(clientRepo, passwordService)

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

	// Create request body
	reqBody := map[string]interface{}{
		"client_name":   "Test Application",
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
			body:       map[string]interface{}{"redirect_uris": []string{"https://app.example.com/callback"}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Missing redirect_uris",
			body:       map[string]interface{}{"client_name": "Test App"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Empty redirect_uris",
			body:       map[string]interface{}{"client_name": "Test App", "redirect_uris": []string{}},
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

func TestClientHandler_ListClients_Success(t *testing.T) {
	router, jwtService, repo := setupClientRouter(t)

	// Pre-populate with test clients
	repo.clients["client-1"] = &entities.Client{
		ClientID:            "client-1",
		TenantID:            "tenant-456",
		ClientName:          "App One",
		AllowedRedirectURIs: []string{"https://app1.example.com/callback"},
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	repo.clients["client-2"] = &entities.Client{
		ClientID:            "client-2",
		TenantID:            "tenant-456",
		ClientName:          "App Two",
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
}

func TestClientHandler_GetClient_Success(t *testing.T) {
	router, jwtService, repo := setupClientRouter(t)

	// Pre-populate with test client
	repo.clients["client-123"] = &entities.Client{
		ClientID:            "client-123",
		TenantID:            "tenant-456",
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

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "client-123", resp["client_id"])
	assert.Equal(t, "Test App", resp["client_name"])
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

	// Pre-populate with test client
	repo.clients["client-123"] = &entities.Client{
		ClientID:            "client-123",
		TenantID:            "tenant-456",
		ClientName:          "Test App",
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
