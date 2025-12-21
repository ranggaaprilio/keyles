package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/infrastructure/persistence/postgres"
	httphandlers "github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/usecase/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestResendOTPHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupResendOTPTestDB(t)

	// Initialize repositories
	tenantRepo := postgres.NewPostgresTenantRepository(db)
	userRepo := postgres.NewPostgresUserRepository(db)
	auditRepo := &MockAuditRepository{}

	// Mock OTP repository
	otpRepo := &MockOTPRepositoryV2{
		store:     make(map[string]*entities.OTPVerification),
		rateLimit: make(map[string]int),
	}

	// Mock services
	emailService := &MockEmailService{}
	otpService := &MockOTPService{}

	// Initialize use case
	resendUseCase := tenant.NewResendOTPUseCase(
		otpRepo,
		tenantRepo,
		userRepo,
		emailService,
		otpService,
		auditRepo,
	)

	// Create handler
	handler := httphandlers.NewResendOTPHandler(resendUseCase)

	// Setup router
	router := gin.New()
	router.POST("/api/v1/resend-otp", handler.ResendOTP)

	// Create test tenant and user
	testTenant, testUser := createResendTestTenantAndUser(t, db)

	// Create initial OTP
	initialOTP := &entities.OTPVerification{
		ID:        uuid.New(),
		TenantID:  testTenant.ID.String(),
		Code:      "123456",
		Purpose:   "email_verification",
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now().Add(-5 * time.Minute),
		Verified:  false,
	}
	err := otpRepo.Create(context.Background(), initialOTP)
	require.NoError(t, err)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedError  string
		setupFunc      func()
		checkState     func(*testing.T)
	}{
		{
			name: "Successful OTP resend",
			payload: map[string]interface{}{
				"tenant_id": testTenant.ID.String(),
			},
			expectedStatus: http.StatusOK,
			checkState: func(t *testing.T) {
				// Verify new OTP was created
				newOTP, err := otpRepo.FindByTenantIDAndPurpose(context.Background(), testTenant.ID.String(), "email_verification")
				require.NoError(t, err)
				assert.Equal(t, "123456", newOTP.Code) // MockOTPService always returns "123456"
				
				// Verify rate limit was incremented
				count, err := otpRepo.GetRateLimitCounter(context.Background(), testUser.Email)
				require.NoError(t, err)
				assert.Equal(t, 1, count)
			},
		},
		{
			name: "Rate limit exceeded",
			payload: map[string]interface{}{
				"tenant_id": testTenant.ID.String(),
			},
			expectedStatus: http.StatusTooManyRequests,
			expectedError:  "rate limit exceeded",
			setupFunc: func() {
				// Simulate 4 previous requests (exceeds max of 3)
				otpRepo.rateLimit[testUser.Email] = 3
			},
		},
		{
			name: "OTP not found for tenant",
			payload: map[string]interface{}{
				"tenant_id": uuid.New().String(), // Random tenant with no OTP
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "OTP not found",
			setupFunc: func() {
				// Create tenant without OTP
				newTenant := &entities.Tenant{
					ID:               uuid.New(),
					OrganizationName: "NoOTPOrg",
					Status:           entities.TenantStatusPendingVerification,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				}
				db.Create(newTenant)
				
				newUser := &entities.AdminUser{
					ID:           uuid.New(),
					TenantID:     newTenant.ID,
					FullName:     "No OTP User",
					Email:        "nootp@test.com",
					PasswordHash: "$2a$10$dummyhash",
					Role:         entities.UserRoleAdmin,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				db.Create(newUser)
			},
		},
		{
			name: "Tenant not found",
			payload: map[string]interface{}{
				"tenant_id": uuid.New().String(), // Non-existent tenant
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "tenant not found",
		},
		{
			name: "Invalid tenant ID format",
			payload: map[string]interface{}{
				"tenant_id": "invalid-uuid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid tenant ID",
		},
		{
			name: "Missing tenant ID",
			payload: map[string]interface{}{
				// Empty payload
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Multiple rapid requests within window",
			payload: map[string]interface{}{
				"tenant_id": testTenant.ID.String(),
			},
			expectedStatus: http.StatusOK,
			setupFunc: func() {
				// Reset rate limit to allow 2 more requests
				otpRepo.rateLimit[testUser.Email] = 1
			},
			checkState: func(t *testing.T) {
				// Verify rate limit incremented
				count, err := otpRepo.GetRateLimitCounter(context.Background(), testUser.Email)
				require.NoError(t, err)
				assert.Equal(t, 2, count)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run setup function if provided
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			// Create request
			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/resend-otp", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code, "Response body: %s", w.Body.String())

			// Parse response
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check error message if expected
			if tt.expectedError != "" {
				errorMsg, ok := response["error"].(string)
				require.True(t, ok, "Expected error field in response")
				assert.Contains(t, errorMsg, tt.expectedError)
			}

			// Check success message if status is OK
			if w.Code == http.StatusOK {
				message, ok := response["message"].(string)
				require.True(t, ok, "Expected message field in response")
				assert.Contains(t, message, "OTP resent")
			}

			// Run state checks if provided
			if tt.checkState != nil {
				tt.checkState(t)
			}
		})
	}
}

// setupResendOTPTestDB creates an in-memory SQLite database for resend OTP tests
func setupResendOTPTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate tables
	err = db.AutoMigrate(
		&entities.Tenant{},
		&entities.AdminUser{},
	)
	require.NoError(t, err)

	return db
}

// createResendTestTenantAndUser creates a test tenant and admin user for resend tests
func createResendTestTenantAndUser(t *testing.T, db *gorm.DB) (*entities.Tenant, *entities.AdminUser) {
	tenant := &entities.Tenant{
		ID:               uuid.New(),
		OrganizationName: "ResendTestOrg",
		Status:           entities.TenantStatusPendingVerification,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	err := db.Create(tenant).Error
	require.NoError(t, err)

	user := &entities.AdminUser{
		ID:           uuid.New(),
		TenantID:     tenant.ID,
		FullName:     "Resend Test Admin",
		Email:        "admin@resendtest.com",
		PasswordHash: "$2a$10$dummyhash",
		Role:         entities.UserRoleAdmin,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	return tenant, user
}
