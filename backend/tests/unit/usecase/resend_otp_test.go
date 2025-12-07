package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestResendOTP_Success tests successful OTP resend flow
func TestResendOTP_Success(t *testing.T) {
	// Arrange
	otpRepo := new(mocks.MockOTPRepository)
	tenantRepo := new(mocks.MockTenantRepository)
	userRepo := new(mocks.MockUserRepository)
	emailService := new(mocks.MockEmailService)
	otpService := new(mocks.MockOTPService)
	auditRepo := new(mocks.MockAuditRepository)

	tenantID := uuid.New()
	email := "admin@example.com"
	fullName := "Admin User"
	orgName := "Example Org"
	oldOTPCode := "123456"
	newOTPCode := "654321"

	// Existing OTP that will be invalidated
	oldOTP := &entities.OTPVerification{
		ID:        uuid.New(),
		TenantID:  tenantID.String(),
		Code:      oldOTPCode,
		Purpose:   "email_verification",
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now().Add(-5 * time.Minute),
	}

	// Tenant
	tenantObj := &entities.Tenant{
		ID:               tenantID,
		OrganizationName: orgName,
		Status:           entities.TenantStatusPendingVerification,
	}

	// Admin user
	adminUser := &entities.AdminUser{
		ID:       uuid.New(),
		TenantID: tenantID,
		Email:    email,
		FullName: fullName,
		Role:     entities.UserRoleAdmin,
	}

	// Mock expectations
	tenantRepo.On("FindByID", mock.Anything, tenantID).Return(tenantObj, nil)
	userRepo.On("FindByTenantID", mock.Anything, tenantID).Return([]*entities.AdminUser{adminUser}, nil)
	otpRepo.On("IncrementRateLimitCounter", mock.Anything, email, 10*time.Minute).Return(1, nil)
	otpRepo.On("FindByTenantIDAndPurpose", mock.Anything, tenantID.String(), "email_verification").Return(oldOTP, nil)
	otpRepo.On("Update", mock.Anything, mock.MatchedBy(func(otp *entities.OTPVerification) bool {
		return otp.ID == oldOTP.ID && otp.ExpiresAt.Before(time.Now())
	})).Return(nil)
	
	otpService.On("Generate").Return(newOTPCode, nil)
	
	otpRepo.On("Create", mock.Anything, mock.MatchedBy(func(otp *entities.OTPVerification) bool {
		return otp.TenantID == tenantID.String() &&
			otp.Code == newOTPCode &&
			otp.Purpose == "email_verification" &&
			otp.ExpiresAt.After(time.Now().Add(9*time.Minute))
	})).Return(nil)
	
	emailService.On("SendOTPEmail", mock.Anything, email, fullName, newOTPCode, orgName).Return(nil)
	auditRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *entities.AuditLog) bool {
		return log.TenantID != nil && *log.TenantID == tenantID && log.EventType == entities.EventOTPSent
	})).Return(nil)

	// Act
	useCase := tenant.NewResendOTPUseCase(otpRepo, tenantRepo, userRepo, emailService, otpService, auditRepo)
	err := useCase.Execute(tenantID.String())

	// Assert
	assert.NoError(t, err)
	otpRepo.AssertExpectations(t)
	tenantRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	emailService.AssertExpectations(t)
	otpService.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

// TestResendOTP_RateLimitExceeded tests rate limiting enforcement
func TestResendOTP_RateLimitExceeded(t *testing.T) {
	// Arrange
	otpRepo := new(mocks.MockOTPRepository)
	tenantRepo := new(mocks.MockTenantRepository)
	userRepo := new(mocks.MockUserRepository)
	emailService := new(mocks.MockEmailService)
	otpService := new(mocks.MockOTPService)
	auditRepo := new(mocks.MockAuditRepository)

	tenantID := uuid.New()
	email := "admin@example.com"

	tenantObj := &entities.Tenant{
		ID:               tenantID,
		OrganizationName: "Example Org",
		Status:           entities.TenantStatusPendingVerification,
	}

	adminUser := &entities.AdminUser{
		ID:       uuid.New(),
		TenantID: tenantID,
		Email:    email,
		FullName: "Admin User",
		Role:     entities.UserRoleAdmin,
	}

	// Mock expectations - rate limit exceeded (4 requests, max is 3)
	tenantRepo.On("FindByID", mock.Anything, tenantID).Return(tenantObj, nil)
	userRepo.On("FindByTenantID", mock.Anything, tenantID).Return([]*entities.AdminUser{adminUser}, nil)
	otpRepo.On("IncrementRateLimitCounter", mock.Anything, email, 10*time.Minute).Return(4, nil)

	// Act
	useCase := tenant.NewResendOTPUseCase(otpRepo, tenantRepo, userRepo, emailService, otpService, auditRepo)
	err := useCase.Execute(tenantID.String())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded")
	tenantRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	otpRepo.AssertExpectations(t)
	otpRepo.AssertNotCalled(t, "Update")
	otpRepo.AssertNotCalled(t, "Create")
	emailService.AssertNotCalled(t, "SendOTPEmail")
}

// TestResendOTP_OTPNotFound tests when no existing OTP is found
func TestResendOTP_OTPNotFound(t *testing.T) {
	// Arrange
	otpRepo := new(mocks.MockOTPRepository)
	tenantRepo := new(mocks.MockTenantRepository)
	userRepo := new(mocks.MockUserRepository)
	emailService := new(mocks.MockEmailService)
	otpService := new(mocks.MockOTPService)
	auditRepo := new(mocks.MockAuditRepository)

	tenantID := uuid.New()
	email := "admin@example.com"

	tenantObj := &entities.Tenant{
		ID:               tenantID,
		OrganizationName: "Example Org",
		Status:           entities.TenantStatusPendingVerification,
	}

	adminUser := &entities.AdminUser{
		ID:       uuid.New(),
		TenantID: tenantID,
		Email:    email,
		FullName: "Admin User",
		Role:     entities.UserRoleAdmin,
	}

	// Mock expectations
	tenantRepo.On("FindByID", mock.Anything, tenantID).Return(tenantObj, nil)
	userRepo.On("FindByTenantID", mock.Anything, tenantID).Return([]*entities.AdminUser{adminUser}, nil)
	otpRepo.On("IncrementRateLimitCounter", mock.Anything, email, 10*time.Minute).Return(1, nil)
	otpRepo.On("FindByTenantIDAndPurpose", mock.Anything, tenantID.String(), "email_verification").Return(nil, errors.New("OTP not found"))

	// Act
	useCase := tenant.NewResendOTPUseCase(otpRepo, tenantRepo, userRepo, emailService, otpService, auditRepo)
	err := useCase.Execute(tenantID.String())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OTP not found")
	otpRepo.AssertExpectations(t)
	tenantRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	otpRepo.AssertNotCalled(t, "Update")
	otpRepo.AssertNotCalled(t, "Create")
	emailService.AssertNotCalled(t, "SendOTPEmail")
}

// TestResendOTP_TenantNotFound tests when tenant doesn't exist
func TestResendOTP_TenantNotFound(t *testing.T) {
	// Arrange
	otpRepo := new(mocks.MockOTPRepository)
	tenantRepo := new(mocks.MockTenantRepository)
	userRepo := new(mocks.MockUserRepository)
	emailService := new(mocks.MockEmailService)
	otpService := new(mocks.MockOTPService)
	auditRepo := new(mocks.MockAuditRepository)

	tenantID := uuid.New()

	// Mock expectations
	tenantRepo.On("FindByID", mock.Anything, tenantID).Return(nil, errors.New("tenant not found"))

	// Act
	useCase := tenant.NewResendOTPUseCase(otpRepo, tenantRepo, userRepo, emailService, otpService, auditRepo)
	err := useCase.Execute(tenantID.String())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant not found")
	tenantRepo.AssertExpectations(t)
	userRepo.AssertNotCalled(t, "FindByTenantID")
	otpRepo.AssertNotCalled(t, "IncrementRateLimitCounter")
	otpRepo.AssertNotCalled(t, "FindByTenantIDAndPurpose")
	emailService.AssertNotCalled(t, "SendOTPEmail")
}

// TestResendOTP_AdminUserNotFound tests when admin user doesn't exist
func TestResendOTP_AdminUserNotFound(t *testing.T) {
	// Arrange
	otpRepo := new(mocks.MockOTPRepository)
	tenantRepo := new(mocks.MockTenantRepository)
	userRepo := new(mocks.MockUserRepository)
	emailService := new(mocks.MockEmailService)
	otpService := new(mocks.MockOTPService)
	auditRepo := new(mocks.MockAuditRepository)

	tenantID := uuid.New()

	tenantObj := &entities.Tenant{
		ID:               tenantID,
		OrganizationName: "Example Org",
		Status:           entities.TenantStatusPendingVerification,
	}

	// Mock expectations
	tenantRepo.On("FindByID", mock.Anything, tenantID).Return(tenantObj, nil)
	userRepo.On("FindByTenantID", mock.Anything, tenantID).Return([]*entities.AdminUser{}, errors.New("user not found"))

	// Act
	useCase := tenant.NewResendOTPUseCase(otpRepo, tenantRepo, userRepo, emailService, otpService, auditRepo)
	err := useCase.Execute(tenantID.String())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "admin user not found")
	tenantRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	otpRepo.AssertNotCalled(t, "IncrementRateLimitCounter")
	emailService.AssertNotCalled(t, "SendOTPEmail")
}

// TestResendOTP_EmailServiceFailure tests email sending failure
func TestResendOTP_EmailServiceFailure(t *testing.T) {
	// Arrange
	otpRepo := new(mocks.MockOTPRepository)
	tenantRepo := new(mocks.MockTenantRepository)
	userRepo := new(mocks.MockUserRepository)
	emailService := new(mocks.MockEmailService)
	otpService := new(mocks.MockOTPService)
	auditRepo := new(mocks.MockAuditRepository)

	tenantID := uuid.New()
	email := "admin@example.com"
	fullName := "Admin User"
	orgName := "Example Org"
	oldOTPCode := "123456"
	newOTPCode := "654321"

	oldOTP := &entities.OTPVerification{
		ID:        uuid.New(),
		TenantID:  tenantID.String(),
		Code:      oldOTPCode,
		Purpose:   "email_verification",
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now().Add(-5 * time.Minute),
	}

	tenantObj := &entities.Tenant{
		ID:               tenantID,
		OrganizationName: orgName,
		Status:           entities.TenantStatusPendingVerification,
	}

	adminUser := &entities.AdminUser{
		ID:       uuid.New(),
		TenantID: tenantID,
		Email:    email,
		FullName: fullName,
		Role:     entities.UserRoleAdmin,
	}

	// Mock expectations
	tenantRepo.On("FindByID", mock.Anything, tenantID).Return(tenantObj, nil)
	userRepo.On("FindByTenantID", mock.Anything, tenantID).Return([]*entities.AdminUser{adminUser}, nil)
	otpRepo.On("IncrementRateLimitCounter", mock.Anything, email, 10*time.Minute).Return(1, nil)
	otpRepo.On("FindByTenantIDAndPurpose", mock.Anything, tenantID.String(), "email_verification").Return(oldOTP, nil)
	otpRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
	otpService.On("Generate").Return(newOTPCode, nil)
	otpRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	emailService.On("SendOTPEmail", mock.Anything, email, fullName, newOTPCode, orgName).Return(errors.New("email service unavailable"))

	// Act
	useCase := tenant.NewResendOTPUseCase(otpRepo, tenantRepo, userRepo, emailService, otpService, auditRepo)
	err := useCase.Execute(tenantID.String())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email service unavailable")
	emailService.AssertExpectations(t)
}

// TestResendOTP_AuditLogFailureDoesNotBlockResend tests that audit log failures don't prevent resend
func TestResendOTP_AuditLogFailureDoesNotBlockResend(t *testing.T) {
	// Arrange
	otpRepo := new(mocks.MockOTPRepository)
	tenantRepo := new(mocks.MockTenantRepository)
	userRepo := new(mocks.MockUserRepository)
	emailService := new(mocks.MockEmailService)
	otpService := new(mocks.MockOTPService)
	auditRepo := new(mocks.MockAuditRepository)

	tenantID := uuid.New()
	email := "admin@example.com"
	fullName := "Admin User"
	orgName := "Example Org"
	oldOTPCode := "123456"
	newOTPCode := "654321"

	oldOTP := &entities.OTPVerification{
		ID:        uuid.New(),
		TenantID:  tenantID.String(),
		Code:      oldOTPCode,
		Purpose:   "email_verification",
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now().Add(-5 * time.Minute),
	}

	tenantObj := &entities.Tenant{
		ID:               tenantID,
		OrganizationName: orgName,
		Status:           entities.TenantStatusPendingVerification,
	}

	adminUser := &entities.AdminUser{
		ID:       uuid.New(),
		TenantID: tenantID,
		Email:    email,
		FullName: fullName,
		Role:     entities.UserRoleAdmin,
	}

	// Mock expectations
	tenantRepo.On("FindByID", mock.Anything, tenantID).Return(tenantObj, nil)
	userRepo.On("FindByTenantID", mock.Anything, tenantID).Return([]*entities.AdminUser{adminUser}, nil)
	otpRepo.On("IncrementRateLimitCounter", mock.Anything, email, 10*time.Minute).Return(1, nil)
	otpRepo.On("FindByTenantIDAndPurpose", mock.Anything, tenantID.String(), "email_verification").Return(oldOTP, nil)
	otpRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
	otpService.On("Generate").Return(newOTPCode, nil)
	otpRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	emailService.On("SendOTPEmail", mock.Anything, email, fullName, newOTPCode, orgName).Return(nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("audit database unavailable"))

	// Act
	useCase := tenant.NewResendOTPUseCase(otpRepo, tenantRepo, userRepo, emailService, otpService, auditRepo)
	err := useCase.Execute(tenantID.String())

	// Assert - should still succeed despite audit log failure
	assert.NoError(t, err)
	auditRepo.AssertExpectations(t)
}
