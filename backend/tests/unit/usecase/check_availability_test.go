package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ranggaaprilio/keyles/usecase/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Test CheckAvailability use case
func TestCheckAvailability(t *testing.T) {
	tests := []struct {
		name           string
		orgName        string
		email          string
		setupMocks     func(*MockTenantRepository, *MockUserRepository)
		wantErr        bool
		errContains    string
		validateResult func(*testing.T, *tenant.CheckAvailabilityResult)
	}{
		{
			name:    "Both organization name and email available",
			orgName: "Acme Corp",
			email:   "admin@acme.com",
			setupMocks: func(tenantRepo *MockTenantRepository, userRepo *MockUserRepository) {
				tenantRepo.On("OrganizationNameExists", mock.Anything, "Acme Corp").Return(false, nil)
				userRepo.On("EmailExists", mock.Anything, "admin@acme.com").Return(false, nil)
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *tenant.CheckAvailabilityResult) {
				require.NotNil(t, result)
				assert.True(t, result.OrganizationNameAvailable)
				assert.True(t, result.EmailAvailable)
			},
		},
		{
			name:    "Organization name taken",
			orgName: "Acme Corp",
			email:   "admin@acme.com",
			setupMocks: func(tenantRepo *MockTenantRepository, userRepo *MockUserRepository) {
				tenantRepo.On("OrganizationNameExists", mock.Anything, "Acme Corp").Return(true, nil)
				userRepo.On("EmailExists", mock.Anything, "admin@acme.com").Return(false, nil)
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *tenant.CheckAvailabilityResult) {
				require.NotNil(t, result)
				assert.False(t, result.OrganizationNameAvailable)
				assert.True(t, result.EmailAvailable)
			},
		},
		{
			name:    "Email taken",
			orgName: "Acme Corp",
			email:   "admin@acme.com",
			setupMocks: func(tenantRepo *MockTenantRepository, userRepo *MockUserRepository) {
				tenantRepo.On("OrganizationNameExists", mock.Anything, "Acme Corp").Return(false, nil)
				userRepo.On("EmailExists", mock.Anything, "admin@acme.com").Return(true, nil)
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *tenant.CheckAvailabilityResult) {
				require.NotNil(t, result)
				assert.True(t, result.OrganizationNameAvailable)
				assert.False(t, result.EmailAvailable)
			},
		},
		{
			name:    "Both taken",
			orgName: "Acme Corp",
			email:   "admin@acme.com",
			setupMocks: func(tenantRepo *MockTenantRepository, userRepo *MockUserRepository) {
				tenantRepo.On("OrganizationNameExists", mock.Anything, "Acme Corp").Return(true, nil)
				userRepo.On("EmailExists", mock.Anything, "admin@acme.com").Return(true, nil)
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *tenant.CheckAvailabilityResult) {
				require.NotNil(t, result)
				assert.False(t, result.OrganizationNameAvailable)
				assert.False(t, result.EmailAvailable)
			},
		},
		{
			name:    "Database error checking organization name",
			orgName: "Acme Corp",
			email:   "admin@acme.com",
			setupMocks: func(tenantRepo *MockTenantRepository, userRepo *MockUserRepository) {
				tenantRepo.On("OrganizationNameExists", mock.Anything, "Acme Corp").Return(false, errors.New("database error"))
			},
			wantErr:     true,
			errContains: "database error",
		},
		{
			name:    "Database error checking email",
			orgName: "Acme Corp",
			email:   "admin@acme.com",
			setupMocks: func(tenantRepo *MockTenantRepository, userRepo *MockUserRepository) {
				tenantRepo.On("OrganizationNameExists", mock.Anything, "Acme Corp").Return(false, nil)
				userRepo.On("EmailExists", mock.Anything, "admin@acme.com").Return(false, errors.New("database error"))
			},
			wantErr:     true,
			errContains: "database error",
		},
		{
			name:    "Invalid organization name",
			orgName: "A",
			email:   "admin@acme.com",
			setupMocks: func(tenantRepo *MockTenantRepository, userRepo *MockUserRepository) {
				// No mocks needed - validation happens before database calls
			},
			wantErr:     true,
			errContains: "organization name must be between 3 and 100 characters",
		},
		{
			name:    "Invalid email",
			orgName: "Acme Corp",
			email:   "invalid-email",
			setupMocks: func(tenantRepo *MockTenantRepository, userRepo *MockUserRepository) {
				// No mocks needed - validation happens before database calls
			},
			wantErr:     true,
			errContains: "email must be a valid email address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			tenantRepo := new(MockTenantRepository)
			userRepo := new(MockUserRepository)

			tt.setupMocks(tenantRepo, userRepo)

			// Create use case
			useCase := tenant.NewCheckAvailabilityUseCase(tenantRepo, userRepo)

			// Execute
			ctx := context.Background()
			result, err := useCase.Execute(ctx, tt.orgName, tt.email)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}

			// Verify mocks
			tenantRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}
