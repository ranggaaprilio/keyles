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

func TestVerificationHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupVerificationTestDB(t)

	// Initialize repositories
	tenantRepo := postgres.NewPostgresTenantRepository(db)
	userRepo := postgres.NewPostgresUserRepository(db)
	auditRepo := &MockAuditRepository{}

	// Mock OTP repository with new interface
	otpRepo := &MockOTPRepositoryV2{
		store:      make(map[string]*entities.OTPVerification),
		rateLimit:  make(map[string]int),
	}

	// Initialize use case
	verifyUseCase := tenant.NewVerifyTenantUseCase(otpRepo, tenantRepo, auditRepo)

	// Create handler
	handler := httphandlers.NewVerificationHandler(verifyUseCase)

	// Setup router
	router := gin.New()
	router.POST("/api/v1/verify-otp", handler.VerifyOTP)

	// Create test tenant and user
	testTenant, testUser := createTestTenantAndUser(t, db)
	
	// Create valid OTP
	validOTP := &entities.OTPVerification{
		ID:        uuid.New(),
		TenantID:  testTenant.ID.String(),
		Code:      "123456",
		Purpose:   "email_verification",
		ExpiresAt: time.Now().Add(10 * time.Minute),
		CreatedAt: time.Now(),
		Verified:  false,
	}
	err := otpRepo.Create(context.Background(), validOTP)
	require.NoError(t, err)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedError  string
		setupFunc      func()
		checkDB        func(*testing.T)
	}{
		{
			name: "Successful OTP verification",
			payload: map[string]interface{}{
				"tenant_id": testTenant.ID.String(),
				"otp_code":  "123456",
			},
			expectedStatus: http.StatusOK,
			checkDB: func(t *testing.T) {
				// Verify tenant status changed to active
				var updatedTenant entities.Tenant
				err := db.First(&updatedTenant, testTenant.ID).Error
				require.NoError(t, err)
				assert.Equal(t, entities.TenantStatusActive, updatedTenant.Status)
				assert.NotNil(t, updatedTenant.VerifiedAt)

				// Verify OTP was marked as verified
				otp, err := otpRepo.FindByTenantIDAndPurpose(context.Background(), testTenant.ID.String(), "email_verification")
				require.NoError(t, err)
				assert.True(t, otp.Verified)
			},
		},
		{
			name: "Invalid OTP code",
			payload: map[string]interface{}{
				"tenant_id": testTenant.ID.String(),
				"otp_code":  "999999",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid",
			setupFunc: func() {
				// Reset tenant status and recreate OTP
				db.Model(&testTenant).Updates(map[string]interface{}{
					"status":      entities.TenantStatusPendingVerification,
					"verified_at": nil,
				})
				validOTP.Verified = false
				otpRepo.Update(context.Background(), validOTP)
			},
		},
		{
			name: "Expired OTP",
			payload: map[string]interface{}{
				"tenant_id": testTenant.ID.String(),
				"otp_code":  "123456",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "expired",
			setupFunc: func() {
				// Create expired OTP
				expiredOTP := &entities.OTPVerification{
					ID:        uuid.New(),
					TenantID:  testTenant.ID.String(),
					Code:      "123456",
					Purpose:   "email_verification",
					ExpiresAt: time.Now().Add(-1 * time.Minute), // Already expired
					CreatedAt: time.Now().Add(-15 * time.Minute),
					Verified:  false,
				}
				otpRepo.store[testTenant.ID.String()+":email_verification"] = expiredOTP
			},
		},
		{
			name: "Already verified OTP",
			payload: map[string]interface{}{
				"tenant_id": testTenant.ID.String(),
				"otp_code":  "123456",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "already been used",
			setupFunc: func() {
				// Mark OTP as already verified
				validOTP.Verified = true
				validOTP.VerifiedAt = func() *time.Time { t := time.Now(); return &t }()
				otpRepo.Update(context.Background(), validOTP)
			},
		},
		{
			name: "OTP not found",
			payload: map[string]interface{}{
				"tenant_id": uuid.New().String(), // Random tenant ID
				"otp_code":  "123456",
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "OTP not found",
		},
		{
			name: "Invalid tenant ID format",
			payload: map[string]interface{}{
				"tenant_id": "invalid-uuid",
				"otp_code":  "123456",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid tenant ID",
		},
		{
			name: "Missing OTP code",
			payload: map[string]interface{}{
				"tenant_id": testTenant.ID.String(),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid OTP format (not 6 digits)",
			payload: map[string]interface{}{
				"tenant_id": testTenant.ID.String(),
				"otp_code":  "12345", // Only 5 digits
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "OTP code must be 6 digits",
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

			req := httptest.NewRequest(http.MethodPost, "/api/v1/verify-otp", bytes.NewReader(body))
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

			// Run database checks if provided
			if tt.checkDB != nil {
				tt.checkDB(t)
			}
		})
	}
}

// setupVerificationTestDB creates an in-memory SQLite database for verification tests
func setupVerificationTestDB(t *testing.T) *gorm.DB {
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

// createTestTenantAndUser creates a test tenant and admin user
func createTestTenantAndUser(t *testing.T, db *gorm.DB) (*entities.Tenant, *entities.AdminUser) {
	tenant := &entities.Tenant{
		ID:               uuid.New(),
		OrganizationName: "TestOrg",
		Status:           entities.TenantStatusPendingVerification,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	err := db.Create(tenant).Error
	require.NoError(t, err)

	user := &entities.AdminUser{
		ID:           uuid.New(),
		TenantID:     tenant.ID,
		FullName:     "Test Admin",
		Email:        "admin@testorg.com",
		PasswordHash: "$2a$10$dummyhash",
		Role:         entities.UserRoleAdmin,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	return tenant, user
}

// MockOTPRepositoryV2 implements the updated OTP repository interface
type MockOTPRepositoryV2 struct {
	store     map[string]*entities.OTPVerification // Key: "tenantID:purpose"
	rateLimit map[string]int
}

func (m *MockOTPRepositoryV2) Create(ctx context.Context, otp *entities.OTPVerification) error {
	key := otp.TenantID + ":" + otp.Purpose
	m.store[key] = otp
	return nil
}

func (m *MockOTPRepositoryV2) FindByTenantIDAndPurpose(ctx context.Context, tenantID, purpose string) (*entities.OTPVerification, error) {
	key := tenantID + ":" + purpose
	otp, exists := m.store[key]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return otp, nil
}

func (m *MockOTPRepositoryV2) Update(ctx context.Context, otp *entities.OTPVerification) error {
	key := otp.TenantID + ":" + otp.Purpose
	m.store[key] = otp
	return nil
}

func (m *MockOTPRepositoryV2) Delete(ctx context.Context, id uuid.UUID) error {
	// Find and delete by ID
	for key, otp := range m.store {
		if otp.ID == id {
			delete(m.store, key)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (m *MockOTPRepositoryV2) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	for key, otp := range m.store {
		if otp.ExpiresAt.Before(now) {
			delete(m.store, key)
		}
	}
	return nil
}

func (m *MockOTPRepositoryV2) IncrementRateLimitCounter(ctx context.Context, email string, window time.Duration) (int, error) {
	m.rateLimit[email]++
	return m.rateLimit[email], nil
}

func (m *MockOTPRepositoryV2) GetRateLimitCounter(ctx context.Context, email string) (int, error) {
	return m.rateLimit[email], nil
}
