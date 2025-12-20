/**
 * Integration tests for auth handler (login endpoint)
 */

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/infrastructure/services"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginHandler_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create mock repositories
	mockUserRepo := &MockUserRepository{}
	mockTenantRepo := &MockTenantRepository{}
	passwordService := services.NewBcryptPasswordService()
	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)

	// Create test data
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	tenant := &entities.Tenant{
		ID:               "tenant-123",
		OrganizationName: "Test Org",
		Status:           "active",
		CreatedAt:        time.Now(),
		VerifiedAt:       func() *time.Time { t := time.Now(); return &t }(),
	}

	user := &entities.AdminUser{
		ID:           "user-123",
		TenantID:     "tenant-123",
		FullName:     "John Doe",
		Email:        "john@example.com",
		PasswordHash: string(hashedPassword),
		Role:         "admin",
		CreatedAt:    time.Now(),
	}

	mockUserRepo.users = map[string]*entities.AdminUser{
		"john@example.com": user,
	}
	mockTenantRepo.tenants = map[string]*entities.Tenant{
		"tenant-123": tenant,
	}

	// Create use case and handler
	authenticateUseCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, passwordService, jwtService)
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
	assert.Equal(t, "user-123", userInfo["id"])
	assert.Equal(t, "john@example.com", userInfo["email"])
	assert.Equal(t, "John Doe", userInfo["full_name"])
	assert.Equal(t, "admin", userInfo["role"])

	tenantInfo := response["tenant"].(map[string]interface{})
	assert.Equal(t, "tenant-123", tenantInfo["id"])
	assert.Equal(t, "Test Org", tenantInfo["organization_name"])
	assert.Equal(t, "active", tenantInfo["status"])
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockUserRepo := &MockUserRepository{}
	mockTenantRepo := &MockTenantRepository{}
	passwordService := services.NewBcryptPasswordService()
	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &entities.AdminUser{
		ID:           "user-123",
		TenantID:     "tenant-123",
		Email:        "john@example.com",
		PasswordHash: string(hashedPassword),
	}

	mockUserRepo.users = map[string]*entities.AdminUser{
		"john@example.com": user,
	}

	authenticateUseCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, passwordService, jwtService)
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

	mockUserRepo := &MockUserRepository{}
	mockTenantRepo := &MockTenantRepository{}
	passwordService := services.NewBcryptPasswordService()
	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)

	mockUserRepo.users = map[string]*entities.AdminUser{}

	authenticateUseCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, passwordService, jwtService)
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

	mockUserRepo := &MockUserRepository{}
	mockTenantRepo := &MockTenantRepository{}
	passwordService := services.NewBcryptPasswordService()
	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	tenant := &entities.Tenant{
		ID:               "tenant-123",
		OrganizationName: "Test Org",
		Status:           "pending",
		CreatedAt:        time.Now(),
	}

	user := &entities.AdminUser{
		ID:           "user-123",
		TenantID:     "tenant-123",
		Email:        "john@example.com",
		PasswordHash: string(hashedPassword),
	}

	mockUserRepo.users = map[string]*entities.AdminUser{
		"john@example.com": user,
	}
	mockTenantRepo.tenants = map[string]*entities.Tenant{
		"tenant-123": tenant,
	}

	authenticateUseCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, passwordService, jwtService)
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

	mockUserRepo := &MockUserRepository{}
	mockTenantRepo := &MockTenantRepository{}
	passwordService := services.NewBcryptPasswordService()
	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)

	authenticateUseCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, passwordService, jwtService)
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

	mockUserRepo := &MockUserRepository{}
	mockTenantRepo := &MockTenantRepository{}
	passwordService := services.NewBcryptPasswordService()
	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)

	authenticateUseCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, passwordService, jwtService)
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

	mockUserRepo := &MockUserRepository{}
	mockTenantRepo := &MockTenantRepository{}
	passwordService := services.NewBcryptPasswordService()
	jwtService := services.NewJWTService("test-secret-key-32-characters!", 24)

	authenticateUseCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, passwordService, jwtService)
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
