package usecase_test

import (
	"context"
	"errors"
	"testing"

	//"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Test RegisterTenant use case
func TestRegisterTenant(t *testing.T) {
	tests := []struct {
		name           string
		orgName        string
		adminEmail     string
		adminPassword  string
		adminFullName  string
		setupMocks     func(*mocks.MockTenantRepository, *mocks.MockUserRepository, *mocks.MockOTPRepository, *mocks.MockAuditRepository, *mocks.MockEmailService, *mocks.MockOTPService, *mocks.MockPasswordService)
		wantErr        bool
		errContains    string
		validateResult func(*testing.T, *tenant.RegisterTenantResult)
	}{
		{
			name:          "Successful registration",
			orgName:       "Acme Corp",
			adminEmail:    "admin@acme.com",
			adminPassword: "Password123!",
			adminFullName: "John Doe",
			setupMocks: func(tenantRepo *mocks.MockTenantRepository, userRepo *mocks.MockUserRepository, otpRepo *mocks.MockOTPRepository, auditRepo *mocks.MockAuditRepository, emailSvc *mocks.MockEmailService, otpSvc *mocks.MockOTPService, pwdSvc *mocks.MockPasswordService) {
				// Organization name doesn't exist
				tenantRepo.On("OrganizationNameExists", mock.Anything, "Acme Corp").Return(false, nil)
				// Email doesn't exist
				userRepo.On("EmailExists", mock.Anything, "admin@acme.com").Return(false, nil)
				// Hash password
				pwdSvc.On("Hash", "Password123!").Return("hashed_password", nil)
				// Create tenant
				tenantRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Tenant")).Return(nil)
				// Create admin user
				userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.AdminUser")).Return(nil)
				// Generate OTP
				otpSvc.On("Generate").Return("123456", nil)
				// Store OTP
				otpRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.OTPVerification")).Return(nil)
				// Send OTP email
				emailSvc.On("SendOTPEmail", mock.Anything, "admin@acme.com", "John Doe", "123456", "Acme Corp").Return(nil)
				// Create audit log
				auditRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.AuditLog")).Return(nil)
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *tenant.RegisterTenantResult) {
				require.NotNil(t, result)
				assert.NotEqual(t, uuid.Nil, result.TenantID)
				assert.NotEqual(t, uuid.Nil, result.AdminUserID)
				assert.Equal(t, "Acme Corp", result.OrganizationName)
				assert.Equal(t, entities.TenantStatusPendingVerification, result.Status)
			},
		},
		{
			name:          "Organization name already exists",
			orgName:       "Acme Corp",
			adminEmail:    "admin@acme.com",
			adminPassword: "Password123!",
			adminFullName: "John Doe",
			setupMocks: func(tenantRepo *mocks.MockTenantRepository, userRepo *mocks.MockUserRepository, otpRepo *mocks.MockOTPRepository, auditRepo *mocks.MockAuditRepository, emailSvc *mocks.MockEmailService, otpSvc *mocks.MockOTPService, pwdSvc *mocks.MockPasswordService) {
				tenantRepo.On("OrganizationNameExists", mock.Anything, "Acme Corp").Return(true, nil)
			},
			wantErr:     true,
			errContains: "organization name already exists",
		},
		{
			name:          "Email already exists",
			orgName:       "Acme Corp",
			adminEmail:    "admin@acme.com",
			adminPassword: "Password123!",
			adminFullName: "John Doe",
			setupMocks: func(tenantRepo *mocks.MockTenantRepository, userRepo *mocks.MockUserRepository, otpRepo *mocks.MockOTPRepository, auditRepo *mocks.MockAuditRepository, emailSvc *mocks.MockEmailService, otpSvc *mocks.MockOTPService, pwdSvc *mocks.MockPasswordService) {
				tenantRepo.On("OrganizationNameExists", mock.Anything, "Acme Corp").Return(false, nil)
				userRepo.On("EmailExists", mock.Anything, "admin@acme.com").Return(true, nil)
			},
			wantErr:     true,
			errContains: "email already exists",
		},
		{
		name:          "Invalid organization name",
		orgName:       "A",
		adminEmail:    "admin@acme.com",
		adminPassword: "Password123!",
		adminFullName: "John Doe",
		setupMocks:    func(*mocks.MockTenantRepository, *mocks.MockUserRepository, *mocks.MockOTPRepository, *mocks.MockAuditRepository, *mocks.MockEmailService, *mocks.MockOTPService, *mocks.MockPasswordService) {},
		wantErr:       true,
		errContains:   "organization name must be between 3 and 100 characters",
		},
		{
		name:          "Invalid email",
		orgName:       "Acme Corp",
		adminEmail:    "invalid-email",
		adminPassword: "Password123!",
		adminFullName: "John Doe",
		setupMocks:    func(*mocks.MockTenantRepository, *mocks.MockUserRepository, *mocks.MockOTPRepository, *mocks.MockAuditRepository, *mocks.MockEmailService, *mocks.MockOTPService, *mocks.MockPasswordService) {},
		wantErr:       true,
		errContains:   "email must be a valid email address",
		},
		{
			name:          "Invalid password",
			orgName:       "Acme Corp",
			adminEmail:    "admin@acme.com",
			adminPassword: "weak",
			adminFullName: "John Doe",
			setupMocks:    func(*mocks.MockTenantRepository, *mocks.MockUserRepository, *mocks.MockOTPRepository, *mocks.MockAuditRepository, *mocks.MockEmailService, *mocks.MockOTPService, *mocks.MockPasswordService) {},
			wantErr:       true,
			errContains:   "password must be at least 8 characters",
		},
		{
			name:          "Database error when creating tenant",
			orgName:       "Acme Corp",
			adminEmail:    "admin@acme.com",
			adminPassword: "Password123!",
			adminFullName: "John Doe",
			setupMocks: func(tenantRepo *mocks.MockTenantRepository, userRepo *mocks.MockUserRepository, otpRepo *mocks.MockOTPRepository, auditRepo *mocks.MockAuditRepository, emailSvc *mocks.MockEmailService, otpSvc *mocks.MockOTPService, pwdSvc *mocks.MockPasswordService) {
				tenantRepo.On("OrganizationNameExists", mock.Anything, "Acme Corp").Return(false, nil)
				userRepo.On("EmailExists", mock.Anything, "admin@acme.com").Return(false, nil)
				pwdSvc.On("Hash", "Password123!").Return("hashed_password", nil)
				tenantRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Tenant")).Return(errors.New("database error"))
			},
			wantErr:     true,
			errContains: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			tenantRepo := new(mocks.MockTenantRepository)
			userRepo := new(mocks.MockUserRepository)
			otpRepo := new(mocks.MockOTPRepository)
			auditRepo := new(mocks.MockAuditRepository)
			emailSvc := new(mocks.MockEmailService)
			otpSvc := new(mocks.MockOTPService)
			pwdSvc := new(mocks.MockPasswordService)

			tt.setupMocks(tenantRepo, userRepo, otpRepo, auditRepo, emailSvc, otpSvc, pwdSvc)

			// Create use case
			useCase := tenant.NewRegisterTenantUseCase(
				tenantRepo,
				userRepo,
				otpRepo,
				auditRepo,
				emailSvc,
				otpSvc,
				pwdSvc,
				false, // skipEmailVerification
			)

			// Execute
			ctx := context.Background()
			result, err := useCase.Execute(ctx, tt.orgName, tt.adminEmail, tt.adminPassword, tt.adminFullName)

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
			otpRepo.AssertExpectations(t)
			auditRepo.AssertExpectations(t)
			emailSvc.AssertExpectations(t)
			otpSvc.AssertExpectations(t)
			pwdSvc.AssertExpectations(t)
		})
	}
}
// TestRegisterTenant_SkipEmailVerification tests that when skipEmailVerification is true,
// the tenant is auto-activated without OTP or email
func TestRegisterTenant_SkipEmailVerification(t *testing.T) {
	tenantRepo := new(mocks.MockTenantRepository)
	userRepo := new(mocks.MockUserRepository)
	otpRepo := new(mocks.MockOTPRepository)
	auditRepo := new(mocks.MockAuditRepository)
	emailSvc := new(mocks.MockEmailService)
	otpSvc := new(mocks.MockOTPService)
	pwdSvc := new(mocks.MockPasswordService)
	tenantRepo.On("OrganizationNameExists", mock.Anything, "SkipCorp").Return(false, nil)
	userRepo.On("EmailExists", mock.Anything, "admin@skip.com").Return(false, nil)
	pwdSvc.On("Hash", "Password123!").Return("hashed_password", nil)
	tenantRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Tenant")).Return(nil)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.AdminUser")).Return(nil)
	// Tenant.Update is called to persist auto-activation
	tenantRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Tenant")).Return(nil)
	auditRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.AuditLog")).Return(nil)
	// OTP and email should NOT be called
	useCase := tenant.NewRegisterTenantUseCase(
		tenantRepo,
		userRepo,
		otpRepo,
		auditRepo,
		emailSvc,
		otpSvc,
		pwdSvc,
		true, // skipEmailVerification
	)
	ctx := context.Background()
	result, err := useCase.Execute(ctx, "SkipCorp", "admin@skip.com", "Password123!", "Admin User")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, entities.TenantStatusActive, result.Status,
		"tenant should be active when email verification is skipped")
	assert.Contains(t, result.Message, "auto-activated")
	// Verify OTP and email were never called
	otpRepo.AssertNotCalled(t, "Create")
	emailSvc.AssertNotCalled(t, "SendOTPEmail")
	otpSvc.AssertNotCalled(t, "Generate")
	tenantRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}
