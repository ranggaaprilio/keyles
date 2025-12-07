package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/usecase/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock repositories and services
type MockTenantRepository struct {
	mock.Mock
}

func (m *MockTenantRepository) Create(ctx context.Context, tenant *entities.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

func (m *MockTenantRepository) FindByOrganizationName(ctx context.Context, orgName string) (*entities.Tenant, error) {
	args := m.Called(ctx, orgName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

func (m *MockTenantRepository) Update(ctx context.Context, tenant *entities.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) OrganizationNameExists(ctx context.Context, orgName string) (bool, error) {
	args := m.Called(ctx, orgName)
	return args.Bool(0), args.Error(1)
}

func (m *MockTenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.AdminUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.AdminUser, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.AdminUser), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.AdminUser, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.AdminUser), args.Error(1)
}

func (m *MockUserRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entities.AdminUser, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.AdminUser), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.AdminUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockOTPRepository struct {
	mock.Mock
}

func (m *MockOTPRepository) Store(ctx context.Context, otp *entities.OTPVerification) error {
	args := m.Called(ctx, otp)
	return args.Error(0)
}

func (m *MockOTPRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID) (*entities.OTPVerification, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.OTPVerification), args.Error(1)
}

func (m *MockOTPRepository) Update(ctx context.Context, otp *entities.OTPVerification) error {
	args := m.Called(ctx, otp)
	return args.Error(0)
}

func (m *MockOTPRepository) Delete(ctx context.Context, tenantID uuid.UUID) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *MockOTPRepository) IncrementRateLimitCounter(ctx context.Context, email string, window time.Duration) (int, error) {
	args := m.Called(ctx, email, window)
	return args.Int(0), args.Error(1)
}

func (m *MockOTPRepository) GetRateLimitCounter(ctx context.Context, email string) (int, error) {
	args := m.Called(ctx, email)
	return args.Int(0), args.Error(1)
}

type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) Create(ctx context.Context, log *entities.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAuditRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entities.AuditLog, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) FindByEventType(ctx context.Context, eventType entities.EventType, limit int, offset int) ([]*entities.AuditLog, error) {
	args := m.Called(ctx, eventType, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) FindRecent(ctx context.Context, limit int) ([]*entities.AuditLog, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.AuditLog), args.Error(1)
}

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendOTPEmail(ctx context.Context, toEmail, toName, otpCode, organizationName string) error {
	args := m.Called(ctx, toEmail, toName, otpCode, organizationName)
	return args.Error(0)
}

func (m *MockEmailService) SendWelcomeEmail(ctx context.Context, toEmail, toName, organizationName string) error {
	args := m.Called(ctx, toEmail, toName, organizationName)
	return args.Error(0)
}

type MockOTPService struct {
	mock.Mock
}

func (m *MockOTPService) Generate() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockOTPService) Validate(code string) bool {
	args := m.Called(code)
	return args.Bool(0)
}

type MockPasswordService struct {
	mock.Mock
}

func (m *MockPasswordService) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordService) Verify(password, hash string) error {
	args := m.Called(password, hash)
	return args.Error(0)
}

// Test RegisterTenant use case
func TestRegisterTenant(t *testing.T) {
	tests := []struct {
		name           string
		orgName        string
		adminEmail     string
		adminPassword  string
		adminFullName  string
		setupMocks     func(*MockTenantRepository, *MockUserRepository, *MockOTPRepository, *MockAuditRepository, *MockEmailService, *MockOTPService, *MockPasswordService)
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
			setupMocks: func(tenantRepo *MockTenantRepository, userRepo *MockUserRepository, otpRepo *MockOTPRepository, auditRepo *MockAuditRepository, emailSvc *MockEmailService, otpSvc *MockOTPService, pwdSvc *MockPasswordService) {
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
				otpRepo.On("Store", mock.Anything, mock.AnythingOfType("*entities.OTPVerification")).Return(nil)
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
			setupMocks: func(tenantRepo *MockTenantRepository, userRepo *MockUserRepository, otpRepo *MockOTPRepository, auditRepo *MockAuditRepository, emailSvc *MockEmailService, otpSvc *MockOTPService, pwdSvc *MockPasswordService) {
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
			setupMocks: func(tenantRepo *MockTenantRepository, userRepo *MockUserRepository, otpRepo *MockOTPRepository, auditRepo *MockAuditRepository, emailSvc *MockEmailService, otpSvc *MockOTPService, pwdSvc *MockPasswordService) {
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
		setupMocks:    func(*MockTenantRepository, *MockUserRepository, *MockOTPRepository, *MockAuditRepository, *MockEmailService, *MockOTPService, *MockPasswordService) {},
		wantErr:       true,
		errContains:   "organization name must be between 3 and 100 characters",
		},
		{
		name:          "Invalid email",
		orgName:       "Acme Corp",
		adminEmail:    "invalid-email",
		adminPassword: "Password123!",
		adminFullName: "John Doe",
		setupMocks:    func(*MockTenantRepository, *MockUserRepository, *MockOTPRepository, *MockAuditRepository, *MockEmailService, *MockOTPService, *MockPasswordService) {},
		wantErr:       true,
		errContains:   "email must be a valid email address",
		},
		{
			name:          "Invalid password",
			orgName:       "Acme Corp",
			adminEmail:    "admin@acme.com",
			adminPassword: "weak",
			adminFullName: "John Doe",
			setupMocks:    func(*MockTenantRepository, *MockUserRepository, *MockOTPRepository, *MockAuditRepository, *MockEmailService, *MockOTPService, *MockPasswordService) {},
			wantErr:       true,
			errContains:   "password must be at least 8 characters",
		},
		{
			name:          "Database error when creating tenant",
			orgName:       "Acme Corp",
			adminEmail:    "admin@acme.com",
			adminPassword: "Password123!",
			adminFullName: "John Doe",
			setupMocks: func(tenantRepo *MockTenantRepository, userRepo *MockUserRepository, otpRepo *MockOTPRepository, auditRepo *MockAuditRepository, emailSvc *MockEmailService, otpSvc *MockOTPService, pwdSvc *MockPasswordService) {
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
			tenantRepo := new(MockTenantRepository)
			userRepo := new(MockUserRepository)
			otpRepo := new(MockOTPRepository)
			auditRepo := new(MockAuditRepository)
			emailSvc := new(MockEmailService)
			otpSvc := new(MockOTPService)
			pwdSvc := new(MockPasswordService)

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
