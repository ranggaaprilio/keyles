/**
 * Integration tests for auth handler (login endpoint)
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
	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/infrastructure/services"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository for auth integration tests
type MockUserRepository struct {
	users map[string]*entities.AdminUser // keyed by email
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.AdminUser) error {
	if m.users == nil {
		m.users = make(map[string]*entities.AdminUser)
	}
	m.users[user.Email] = user
	return nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.AdminUser, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.AdminUser, error) {
	if user, ok := m.users[email]; ok {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entities.AdminUser, error) {
	var result []*entities.AdminUser
	for _, u := range m.users {
		if u.TenantID == tenantID {
			result = append(result, u)
		}
	}
	return result, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.AdminUser) error {
	m.users[user.Email] = user
	return nil
}

func (m *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	_, ok := m.users[email]
	return ok, nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	for email, u := range m.users {
		if u.ID == id {
			delete(m.users, email)
			return nil
		}
	}
	return nil
}

// MockTenantRepository for auth integration tests
type MockTenantRepository struct {
	tenants map[uuid.UUID]*entities.Tenant
}

func (m *MockTenantRepository) Create(ctx context.Context, tenant *entities.Tenant) error {
	if m.tenants == nil {
		m.tenants = make(map[uuid.UUID]*entities.Tenant)
	}
	m.tenants[tenant.ID] = tenant
	return nil
}

func (m *MockTenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Tenant, error) {
	if tenant, ok := m.tenants[id]; ok {
		return tenant, nil
	}
	return nil, errors.New("tenant not found")
}

func (m *MockTenantRepository) FindByOrganizationName(ctx context.Context, name string) (*entities.Tenant, error) {
	for _, t := range m.tenants {
		if t.OrganizationName == name {
			return t, nil
		}
	}
	return nil, errors.New("tenant not found")
}

func (m *MockTenantRepository) Update(ctx context.Context, tenant *entities.Tenant) error {
	m.tenants[tenant.ID] = tenant
	return nil
}

func (m *MockTenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.tenants, id)
	return nil
}

func (m *MockTenantRepository) OrganizationNameExists(ctx context.Context, name string) (bool, error) {
	for _, t := range m.tenants {
		if t.OrganizationName == name {
			return true, nil
		}
	}
	return false, nil
}

func TestLoginHandler_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create mock repositories
	mockUserRepo := &MockUserRepository{users: make(map[string]*entities.AdminUser)}
	mockTenantRepo := &MockTenantRepository{tenants: make(map[uuid.UUID]*entities.Tenant)}
	passwordService := services.NewBcryptPasswordService()
	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)
	jwtAdapter := services.NewAuthJWTServiceAdapter(jwtService)

	// Create test data
	tenantID := uuid.New()
	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	tenant := &entities.Tenant{
		ID:               tenantID,
		OrganizationName: "Test Org",
		Status:           entities.TenantStatusActive,
		CreatedAt:        time.Now(),
		VerifiedAt:       func() *time.Time { t := time.Now(); return &t }(),
	}

	user := &entities.AdminUser{
		ID:           userID,
		TenantID:     tenantID,
		FullName:     "John Doe",
		Email:        "john@example.com",
		PasswordHash: string(hashedPassword),
		Role:         entities.UserRoleAdmin,
		CreatedAt:    time.Now(),
	}

	mockUserRepo.users["john@example.com"] = user
	mockTenantRepo.tenants[tenantID] = tenant

	// Create use case and handler
	authenticateUseCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, passwordService, jwtAdapter)
	authHandler := handlers.NewAuthHandler(authenticateUseCase)

	// Setup route
	router.POST("/api/v1/login", authHandler.Login)

	// Create request
	loginReq := map[string]string{
		"email":    "john@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.NotEmpty(t, response["token"])
	assert.Equal(t, float64(86400), response["expires_in"])

	userInfo := response["user"].(map[string]interface{})
	assert.NotEmpty(t, userInfo["id"])
	assert.Equal(t, "john@example.com", userInfo["email"])
	assert.Equal(t, "John Doe", userInfo["full_name"])
	assert.Equal(t, "admin", userInfo["role"])

	tenantInfo := response["tenant"].(map[string]interface{})
	assert.NotEmpty(t, tenantInfo["id"])
	assert.Equal(t, "Test Org", tenantInfo["organization_name"])
	assert.Equal(t, "active", tenantInfo["status"])
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockUserRepo := &MockUserRepository{users: make(map[string]*entities.AdminUser)}
	mockTenantRepo := &MockTenantRepository{tenants: make(map[uuid.UUID]*entities.Tenant)}
	passwordService := services.NewBcryptPasswordService()
	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)
	jwtAdapter := services.NewAuthJWTServiceAdapter(jwtService)

	userID := uuid.New()
	tenantID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &entities.AdminUser{
		ID:           userID,
		TenantID:     tenantID,
		Email:        "john@example.com",
		PasswordHash: string(hashedPassword),
	}

	mockUserRepo.users["john@example.com"] = user

	authenticateUseCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, passwordService, jwtAdapter)
	authHandler := handlers.NewAuthHandler(authenticateUseCase)

	router.POST("/api/v1/login", authHandler.Login)

	// Create request with wrong password
	loginReq := map[string]string{
		"email":    "john@example.com",
		"password": "wrongpassword",
	}
	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid email or password", response["error"])
}

func TestLoginHandler_UserNotFound(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockUserRepo := &MockUserRepository{users: make(map[string]*entities.AdminUser)}
	mockTenantRepo := &MockTenantRepository{tenants: make(map[uuid.UUID]*entities.Tenant)}
	passwordService := services.NewBcryptPasswordService()
	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)
	jwtAdapter := services.NewAuthJWTServiceAdapter(jwtService)

	mockUserRepo.users = map[string]*entities.AdminUser{}

	authenticateUseCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, passwordService, jwtAdapter)
	authHandler := handlers.NewAuthHandler(authenticateUseCase)

	router.POST("/api/v1/login", authHandler.Login)

	// Create request
	loginReq := map[string]string{
		"email":    "notfound@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid email or password", response["error"])
}

func TestLoginHandler_TenantNotActive(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockUserRepo := &MockUserRepository{users: make(map[string]*entities.AdminUser)}
	mockTenantRepo := &MockTenantRepository{tenants: make(map[uuid.UUID]*entities.Tenant)}
	passwordService := services.NewBcryptPasswordService()
	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)
	jwtAdapter := services.NewAuthJWTServiceAdapter(jwtService)

	tenantID := uuid.New()
	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	tenant := &entities.Tenant{
		ID:               tenantID,
		OrganizationName: "Test Org",
		Status:           entities.TenantStatusPendingVerification,
		CreatedAt:        time.Now(),
	}

	user := &entities.AdminUser{
		ID:           userID,
		TenantID:     tenantID,
		Email:        "john@example.com",
		PasswordHash: string(hashedPassword),
	}

	mockUserRepo.users["john@example.com"] = user
	mockTenantRepo.tenants[tenantID] = tenant

	authenticateUseCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, passwordService, jwtAdapter)
	authHandler := handlers.NewAuthHandler(authenticateUseCase)

	router.POST("/api/v1/login", authHandler.Login)

	// Create request
	loginReq := map[string]string{
		"email":    "john@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Please verify your email before logging in", response["error"])
}

func TestLoginHandler_MissingEmail(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockUserRepo := &MockUserRepository{users: make(map[string]*entities.AdminUser)}
	mockTenantRepo := &MockTenantRepository{tenants: make(map[uuid.UUID]*entities.Tenant)}
	passwordService := services.NewBcryptPasswordService()
	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)
	jwtAdapter := services.NewAuthJWTServiceAdapter(jwtService)

	authenticateUseCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, passwordService, jwtAdapter)
	authHandler := handlers.NewAuthHandler(authenticateUseCase)

	router.POST("/api/v1/login", authHandler.Login)

	// Create request without email
	loginReq := map[string]string{
		"password": "password123",
	}
	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid request data", response["error"])
}

func TestLoginHandler_MissingPassword(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockUserRepo := &MockUserRepository{users: make(map[string]*entities.AdminUser)}
	mockTenantRepo := &MockTenantRepository{tenants: make(map[uuid.UUID]*entities.Tenant)}
	passwordService := services.NewBcryptPasswordService()
	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)
	jwtAdapter := services.NewAuthJWTServiceAdapter(jwtService)

	authenticateUseCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, passwordService, jwtAdapter)
	authHandler := handlers.NewAuthHandler(authenticateUseCase)

	router.POST("/api/v1/login", authHandler.Login)

	// Create request without password
	loginReq := map[string]string{
		"email": "john@example.com",
	}
	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid request data", response["error"])
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockUserRepo := &MockUserRepository{users: make(map[string]*entities.AdminUser)}
	mockTenantRepo := &MockTenantRepository{tenants: make(map[uuid.UUID]*entities.Tenant)}
	passwordService := services.NewBcryptPasswordService()
	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)
	jwtAdapter := services.NewAuthJWTServiceAdapter(jwtService)

	authenticateUseCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, passwordService, jwtAdapter)
	authHandler := handlers.NewAuthHandler(authenticateUseCase)

	router.POST("/api/v1/login", authHandler.Login)

	// Create request with invalid JSON
	req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
