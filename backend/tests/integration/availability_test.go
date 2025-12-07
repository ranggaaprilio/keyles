package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/infrastructure/persistence/postgres"
	httphandlers "github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/usecase/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAvailabilityTestDB(t *testing.T) *gorm.DB {
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

func TestAvailabilityHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupAvailabilityTestDB(t)

	// Seed test data
	passwordService := &MockPasswordService{}
	hashedPassword, _ := passwordService.Hash("TestPassword123!")

	existingTenant, err := entities.NewTenant("ExistingCorp")
	require.NoError(t, err)
	err = db.Create(existingTenant).Error
	require.NoError(t, err)

	existingUser, err := entities.NewAdminUser(existingTenant.ID, "Existing User", "existing@example.com", hashedPassword)
	require.NoError(t, err)
	err = db.Create(existingUser).Error
	require.NoError(t, err)

	// Initialize repositories
	tenantRepo := postgres.NewPostgresTenantRepository(db)
	userRepo := postgres.NewPostgresUserRepository(db)

	// Initialize use case
	checkAvailabilityUseCase := tenant.NewCheckAvailabilityUseCase(tenantRepo, userRepo)

	// Create handler
	handler := httphandlers.NewAvailabilityHandler(checkAvailabilityUseCase)

	// Setup router
	router := gin.New()
	router.GET("/api/v1/check-availability", handler.CheckAvailability)

	tests := []struct {
		name                string
		queryParams         map[string]string
		expectedStatus      int
		expectedOrgAvail    *bool
		expectedEmailAvail  *bool
		expectedError       string
	}{
		{
			name: "Both available",
			queryParams: map[string]string{
				"organization_name": "NewCorp",
				"email":             "new@example.com",
			},
			expectedStatus:     http.StatusOK,
			expectedOrgAvail:   boolPtr(true),
			expectedEmailAvail: boolPtr(true),
		},
		{
			name: "Organization name taken",
			queryParams: map[string]string{
				"organization_name": "ExistingCorp",
				"email":             "new@example.com",
			},
			expectedStatus:     http.StatusOK,
			expectedOrgAvail:   boolPtr(false),
			expectedEmailAvail: boolPtr(true),
		},
		{
			name: "Email taken",
			queryParams: map[string]string{
				"organization_name": "NewCorp",
				"email":             "existing@example.com",
			},
			expectedStatus:     http.StatusOK,
			expectedOrgAvail:   boolPtr(true),
			expectedEmailAvail: boolPtr(false),
		},
		{
			name: "Both taken",
			queryParams: map[string]string{
				"organization_name": "ExistingCorp",
				"email":             "existing@example.com",
			},
			expectedStatus:     http.StatusOK,
			expectedOrgAvail:   boolPtr(false),
			expectedEmailAvail: boolPtr(false),
		},
		{
			name: "Invalid organization name (too short)",
			queryParams: map[string]string{
				"organization_name": "AB",
				"email":             "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "organization name must be between 3 and 100 characters",
		},
		{
			name: "Invalid email format",
			queryParams: map[string]string{
				"organization_name": "ValidCorp",
				"email":             "invalid-email",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email must be a valid email address",
		},
		{
			name: "Missing organization name",
			queryParams: map[string]string{
				"email": "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "organization_name is required",
		},
		{
			name: "Missing email",
			queryParams: map[string]string{
				"organization_name": "TestCorp",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email is required",
		},
		{
			name:           "Missing both parameters",
			queryParams:    map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "organization_name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build query string
			req := httptest.NewRequest(http.MethodGet, "/api/v1/check-availability", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code, "Response body: %s", w.Body.String())

			// Parse response
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check error message if expected
			if tt.expectedError != "" {
				assert.Contains(t, response["error"].(string), tt.expectedError)
			}

			// Check availability results if successful
			if tt.expectedStatus == http.StatusOK {
				if tt.expectedOrgAvail != nil {
					assert.Equal(t, *tt.expectedOrgAvail, response["organization_name_available"].(bool))
				}
				if tt.expectedEmailAvail != nil {
					assert.Equal(t, *tt.expectedEmailAvail, response["email_available"].(bool))
				}
			}
		})
	}
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}

// MockPasswordService for testing
type MockPasswordService struct{}

func (m *MockPasswordService) Hash(password string) (string, error) {
	return "hashed_" + password, nil
}

func (m *MockPasswordService) Verify(hashedPassword, password string) bool {
	return hashedPassword == "hashed_"+password
}
