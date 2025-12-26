/**
 * Integration tests for dashboard handler (JWT protected endpoint)
 */

package integration

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/infrastructure/services"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/interfaces/http/middleware"
	"github.com/stretchr/testify/assert"
)

// MockUserRepository for dashboard tests
type MockDashboardUserRepository struct {
	users map[string]*entities.AdminUser
}

func (m *MockDashboardUserRepository) FindByEmail(ctx context.Context, email string) (*entities.AdminUser, error) {
	if user, ok := m.users[email]; ok {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (m *MockDashboardUserRepository) Create(ctx context.Context, user *entities.AdminUser) error {
	return nil
}

func (m *MockDashboardUserRepository) Update(ctx context.Context, user *entities.AdminUser) error {
	return nil
}

func (m *MockDashboardUserRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entities.AdminUser, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDashboardUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *MockDashboardUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockDashboardUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.AdminUser, error) {
	return nil, errors.New("not implemented")
}

// MockTenantRepository for dashboard tests
type MockDashboardTenantRepository struct {
	tenants map[string]*entities.Tenant
}

func (m *MockDashboardTenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Tenant, error) {
	if tenant, ok := m.tenants[id.String()]; ok {
		return tenant, nil
	}
	return nil, errors.New("tenant not found")
}

func (m *MockDashboardTenantRepository) Create(ctx context.Context, tenant *entities.Tenant) error {
	return nil
}

func (m *MockDashboardTenantRepository) Update(ctx context.Context, tenant *entities.Tenant) error {
	return nil
}

func (m *MockDashboardTenantRepository) FindByOrganizationName(ctx context.Context, name string) (*entities.Tenant, error) {
	return nil, errors.New("not implemented")
}

func (m *MockDashboardTenantRepository) OrganizationNameExists(ctx context.Context, name string) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *MockDashboardTenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestDashboardHandler_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)

	// Create test UUIDs
	tenantID := uuid.New()
	userID := uuid.New()

	// Create test data
	verifiedAt := time.Now()
	tenant := &entities.Tenant{
		ID:               tenantID,
		OrganizationName: "Test Org",
		Status:           "active",
		CreatedAt:        time.Now(),
		VerifiedAt:       &verifiedAt,
	}

	user := &entities.AdminUser{
		ID:        userID,
		TenantID:  tenantID,
		FullName:  "John Doe",
		Email:     "john@example.com",
		Role:      "admin",
		CreatedAt: time.Now(),
	}

	mockUserRepo := &MockDashboardUserRepository{
		users: map[string]*entities.AdminUser{
			"john@example.com": user,
		},
	}

	mockTenantRepo := &MockDashboardTenantRepository{
		tenants: map[string]*entities.Tenant{
			tenantID.String(): tenant,
		},
	}

	// Create handler
	dashboardHandler := handlers.NewDashboardHandler(mockUserRepo, mockTenantRepo)

	// Setup route with auth middleware
	router.GET("/api/v1/dashboard", middleware.AuthMiddleware(jwtService), dashboardHandler.GetDashboard)

	// Generate valid JWT token
	token, _ := jwtService.GenerateToken(userID.String(), tenantID.String(), "john@example.com", "admin")

	// Create request with authorization header
	req, _ := http.NewRequest("GET", "/api/v1/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	tenantInfo := response["tenant"].(map[string]interface{})
	assert.Equal(t, tenantID.String(), tenantInfo["id"])
	assert.Equal(t, "Test Org", tenantInfo["organization_name"])
	assert.Equal(t, "active", tenantInfo["status"])
	assert.NotNil(t, tenantInfo["created_at"])
	assert.NotNil(t, tenantInfo["verified_at"])

	userInfo := response["user"].(map[string]interface{})
	assert.Equal(t, userID.String(), userInfo["id"])
	assert.Equal(t, "John Doe", userInfo["full_name"])
	assert.Equal(t, "john@example.com", userInfo["email"])
	assert.Equal(t, "admin", userInfo["role"])
}

func TestDashboardHandler_MissingToken(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)
	mockUserRepo := &MockDashboardUserRepository{}
	mockTenantRepo := &MockDashboardTenantRepository{}

	dashboardHandler := handlers.NewDashboardHandler(mockUserRepo, mockTenantRepo)

	router.GET("/api/v1/dashboard", middleware.AuthMiddleware(jwtService), dashboardHandler.GetDashboard)

	// Create request without authorization header
	req, _ := http.NewRequest("GET", "/api/v1/dashboard", nil)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Authorization header is required", response["error"])
}

func TestDashboardHandler_InvalidToken(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)
	mockUserRepo := &MockDashboardUserRepository{}
	mockTenantRepo := &MockDashboardTenantRepository{}

	dashboardHandler := handlers.NewDashboardHandler(mockUserRepo, mockTenantRepo)

	router.GET("/api/v1/dashboard", middleware.AuthMiddleware(jwtService), dashboardHandler.GetDashboard)

	// Create request with invalid token
	req, _ := http.NewRequest("GET", "/api/v1/dashboard", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid or expired token", response["error"])
}

func TestDashboardHandler_ExpiredToken(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create JWT service with -1 hour expiration (already expired)
	jwtService := services.NewJWTService("test-secret-key-32-characters!", -1)
	mockUserRepo := &MockDashboardUserRepository{}
	mockTenantRepo := &MockDashboardTenantRepository{}

	dashboardHandler := handlers.NewDashboardHandler(mockUserRepo, mockTenantRepo)

	router.GET("/api/v1/dashboard", middleware.AuthMiddleware(jwtService), dashboardHandler.GetDashboard)

	// Generate expired token
	token, _ := jwtService.GenerateToken(uuid.New().String(), uuid.New().String(), "john@example.com", "admin")

	// Create request with expired token
	req, _ := http.NewRequest("GET", "/api/v1/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid or expired token", response["error"])
}

func TestDashboardHandler_InvalidAuthorizationFormat(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)
	mockUserRepo := &MockDashboardUserRepository{}
	mockTenantRepo := &MockDashboardTenantRepository{}

	dashboardHandler := handlers.NewDashboardHandler(mockUserRepo, mockTenantRepo)

	router.GET("/api/v1/dashboard", middleware.AuthMiddleware(jwtService), dashboardHandler.GetDashboard)

	// Create request with invalid authorization format (missing "Bearer ")
	req, _ := http.NewRequest("GET", "/api/v1/dashboard", nil)
	req.Header.Set("Authorization", "InvalidToken")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid authorization header format", response["error"])
}

func TestDashboardHandler_UserNotFound(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)

	mockUserRepo := &MockDashboardUserRepository{
		users: map[string]*entities.AdminUser{},
	}

	mockTenantRepo := &MockDashboardTenantRepository{
		tenants: map[string]*entities.Tenant{},
	}

	dashboardHandler := handlers.NewDashboardHandler(mockUserRepo, mockTenantRepo)

	router.GET("/api/v1/dashboard", middleware.AuthMiddleware(jwtService), dashboardHandler.GetDashboard)

	// Generate valid JWT token for non-existent user
	token, _ := jwtService.GenerateToken(uuid.New().String(), uuid.New().String(), "notfound@example.com", "admin")

	// Create request with authorization header
	req, _ := http.NewRequest("GET", "/api/v1/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "User not found", response["error"])
}

func TestDashboardHandler_TenantNotFound(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)

	tenantID := uuid.New()
	userID := uuid.New()

	user := &entities.AdminUser{
		ID:        userID,
		TenantID:  tenantID,
		FullName:  "John Doe",
		Email:     "john@example.com",
		Role:      "admin",
		CreatedAt: time.Now(),
	}

	mockUserRepo := &MockDashboardUserRepository{
		users: map[string]*entities.AdminUser{
			"john@example.com": user,
		},
	}

	mockTenantRepo := &MockDashboardTenantRepository{
		tenants: map[string]*entities.Tenant{},
	}

	dashboardHandler := handlers.NewDashboardHandler(mockUserRepo, mockTenantRepo)

	router.GET("/api/v1/dashboard", middleware.AuthMiddleware(jwtService), dashboardHandler.GetDashboard)

	// Generate valid JWT token but tenant doesn't exist
	token, _ := jwtService.GenerateToken(userID.String(), tenantID.String(), "john@example.com", "admin")

	// Create request with authorization header
	req, _ := http.NewRequest("GET", "/api/v1/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Tenant not found", response["error"])
}
