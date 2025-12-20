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

// MockTenantRepository for dashboard tests
type MockDashboardTenantRepository struct {
	tenants map[string]*entities.Tenant
}

func (m *MockDashboardTenantRepository) FindByID(ctx context.Context, id string) (*entities.Tenant, error) {
	if tenant, ok := m.tenants[id]; ok {
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

func TestDashboardHandler_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)

	// Create test data
	verifiedAt := time.Now()
	tenant := &entities.Tenant{
		ID:               "tenant-123",
		OrganizationName: "Test Org",
		Status:           "active",
		CreatedAt:        time.Now(),
		VerifiedAt:       &verifiedAt,
	}

	user := &entities.AdminUser{
		ID:        "user-123",
		TenantID:  "tenant-123",
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
			"tenant-123": tenant,
		},
	}

	// Create handler
	dashboardHandler := handlers.NewDashboardHandler(mockUserRepo, mockTenantRepo)

	// Setup route with auth middleware
	router.GET("/api/v1/dashboard", middleware.AuthMiddleware(jwtService), dashboardHandler.GetDashboard)

	// Generate valid JWT token
	token, _ := jwtService.GenerateToken("user-123", "tenant-123", "john@example.com", "admin")

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
	assert.Equal(t, "tenant-123", tenantInfo["id"])
	assert.Equal(t, "Test Org", tenantInfo["organization_name"])
	assert.Equal(t, "active", tenantInfo["status"])
	assert.NotNil(t, tenantInfo["created_at"])
	assert.NotNil(t, tenantInfo["verified_at"])

	userInfo := response["user"].(map[string]interface{})
	assert.Equal(t, "user-123", userInfo["id"])
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
	token, _ := jwtService.GenerateToken("user-123", "tenant-123", "john@example.com", "admin")

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
	token, _ := jwtService.GenerateToken("user-999", "tenant-999", "notfound@example.com", "admin")

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

	user := &entities.AdminUser{
		ID:        "user-123",
		TenantID:  "tenant-123",
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
	token, _ := jwtService.GenerateToken("user-123", "tenant-123", "john@example.com", "admin")

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
