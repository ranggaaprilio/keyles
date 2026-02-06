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
	"github.com/ranggaaprilio/keyles/infrastructure/services"
	httphandlers "github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/usecase/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate all tables (skip AuditLog due to SQLite map type limitation)
	err = db.AutoMigrate(
		&entities.Tenant{},
		&entities.AdminUser{},
	)
	require.NoError(t, err)

	return db
}

// MockEmailService for testing (doesn't actually send emails)
type MockEmailService struct{}

func (m *MockEmailService) SendOTPEmail(ctx context.Context, toEmail, toName, otpCode, organizationName string) error {
	return nil // Success
}

func (m *MockEmailService) SendWelcomeEmail(ctx context.Context, toEmail, toName, organizationName string) error {
	return nil
}

// MockOTPService for testing (generates predictable OTPs)
type MockOTPService struct{}

func (m *MockOTPService) Generate() (string, error) {
	return "123456", nil // Predictable for testing
}

func (m *MockOTPService) Validate(otp string) bool {
	return len(otp) == 6
}

func TestRegistrationHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	// Initialize repositories
	tenantRepo := postgres.NewPostgresTenantRepository(db)
	userRepo := postgres.NewPostgresUserRepository(db)
	
	// Mock audit repository (SQLite doesn't support JSON fields easily)
	auditRepo := &MockAuditRepository{}

	// Mock Redis OTP repository (in-memory map for testing)
	otpRepo := &MockOTPRepository{
		store: make(map[string]*entities.OTPVerification),
		rateLimit: make(map[string]int),
	}

	// Initialize services
	emailService := &MockEmailService{}
	otpService := &MockOTPService{}
	passwordService := services.NewBcryptPasswordService()

	// Initialize use case
	registerUseCase := tenant.NewRegisterTenantUseCase(
		tenantRepo,
		userRepo,
		otpRepo,
		auditRepo,
		emailService,
		otpService,
		passwordService,
	)

	// Create handler
	handler := httphandlers.NewRegistrationHandler(registerUseCase)

	// Setup router
	router := gin.New()
	router.POST("/api/v1/register", handler.Register)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedError  string
		checkDB        bool
	}{
		{
			name: "Successful registration",
			payload: map[string]interface{}{
				"organization_name": "TechCorp",
				"email":             "admin@techcorp.com",
				"password":          "SecurePassword123!",
				"full_name":         "John Doe",
			},
			expectedStatus: http.StatusCreated,
			checkDB:        true,
		},
		{
			name: "Duplicate organization name",
			payload: map[string]interface{}{
				"organization_name": "TechCorp", // Already exists from first test
				"email":             "another@techcorp.com",
				"password":          "SecurePassword123!",
				"full_name":         "Jane Doe",
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "organization name already exists",
		},
		{
			name: "Duplicate email",
			payload: map[string]interface{}{
				"organization_name": "AnotherCorp",
				"email":             "admin@techcorp.com", // Already exists
				"password":          "SecurePassword123!",
				"full_name":         "Jane Doe",
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "email already exists",
		},
		{
			name: "Invalid organization name (too short)",
			payload: map[string]interface{}{
				"organization_name": "AB",
				"email":             "admin@newcorp.com",
				"password":          "SecurePassword123!",
				"full_name":         "John Doe",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "organization name must be between 3 and 100 characters",
		},
		{
			name: "Invalid email format",
			payload: map[string]interface{}{
				"organization_name": "ValidCorp",
				"email":             "invalid-email",
				"password":          "SecurePassword123!",
				"full_name":         "John Doe",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email", // Gin validation error contains field name
		},
		{
			name: "Invalid password (too short)",
			payload: map[string]interface{}{
				"organization_name": "ValidCorp2",
				"email":             "admin@validcorp2.com",
				"password":          "Weak1!",
				"full_name":         "John Doe",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "password must be at least 8 characters long",
		},
		{
			name: "Missing required fields",
			payload: map[string]interface{}{
				"organization_name": "MissingFields",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Real-IP", "192.168.1.1")
			req.Header.Set("User-Agent", "Test-Agent")

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
				assert.Contains(t, response["error"].(string), tt.expectedError)
			}

			// Check database if successful
			if tt.checkDB && w.Code == http.StatusCreated {
				// Verify tenant was created
				orgName := tt.payload["organization_name"].(string)
				var dbTenant entities.Tenant
				err := db.Where("organization_name = ?", orgName).First(&dbTenant).Error
				require.NoError(t, err)

				assert.Equal(t, orgName, dbTenant.OrganizationName)
				assert.Equal(t, entities.TenantStatusPendingVerification, dbTenant.Status)
				assert.NotEqual(t, uuid.Nil, dbTenant.ID)

				// Verify admin user was created
				email := tt.payload["email"].(string)
				var dbUser entities.AdminUser
				err = db.Where("email = ?", email).First(&dbUser).Error
				require.NoError(t, err)

				assert.Equal(t, email, dbUser.Email)
				assert.Equal(t, dbTenant.ID, dbUser.TenantID)

				// Verify OTP was stored
				otp, exists := otpRepo.store[dbTenant.ID.String()]
				require.True(t, exists, "OTP should be stored")
				assert.Equal(t, "123456", otp.Code)

				// Note: Audit log verification skipped due to SQLite JSON type limitations
			}
		})
	}
}

// MockOTPRepository for testing
type MockOTPRepository struct {
	store     map[string]*entities.OTPVerification // keyed by TenantID (string)
	rateLimit map[string]int
}

func (m *MockOTPRepository) Create(ctx context.Context, otp *entities.OTPVerification) error {
	if m.store == nil {
		m.store = make(map[string]*entities.OTPVerification)
	}
	m.store[otp.TenantID] = otp
	return nil
}

func (m *MockOTPRepository) FindByTenantIDAndPurpose(ctx context.Context, tenantID, purpose string) (*entities.OTPVerification, error) {
	otp, exists := m.store[tenantID]
	if !exists || otp.Purpose != purpose {
		return nil, gorm.ErrRecordNotFound
	}
	return otp, nil
}

func (m *MockOTPRepository) Update(ctx context.Context, otp *entities.OTPVerification) error {
	m.store[otp.TenantID] = otp
	return nil
}

func (m *MockOTPRepository) Delete(ctx context.Context, id uuid.UUID) error {
	for key, otp := range m.store {
		if otp.ID == id {
			delete(m.store, key)
			return nil
		}
	}
	return nil
}

func (m *MockOTPRepository) DeleteExpired(ctx context.Context) error {
	return nil
}

func (m *MockOTPRepository) IncrementRateLimitCounter(ctx context.Context, email string, window time.Duration) (int, error) {
	if m.rateLimit == nil {
		m.rateLimit = make(map[string]int)
	}
	m.rateLimit[email]++
	return m.rateLimit[email], nil
}

func (m *MockOTPRepository) GetRateLimitCounter(ctx context.Context, email string) (int, error) {
	if m.rateLimit == nil {
		return 0, nil
	}
	return m.rateLimit[email], nil
}

// MockAuditRepository for testing (no-op since SQLite doesn't handle JSON fields well)
type MockAuditRepository struct{}

func (m *MockAuditRepository) Create(ctx context.Context, log *entities.AuditLog) error {
	return nil // No-op for integration tests
}

func (m *MockAuditRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entities.AuditLog, error) {
	return []*entities.AuditLog{}, nil
}

func (m *MockAuditRepository) FindByEventType(ctx context.Context, eventType entities.EventType, limit, offset int) ([]*entities.AuditLog, error) {
	return []*entities.AuditLog{}, nil
}

func (m *MockAuditRepository) FindRecent(ctx context.Context, limit int) ([]*entities.AuditLog, error) {
	return []*entities.AuditLog{}, nil
}
