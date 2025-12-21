package usecase_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVerifyTenant_Execute(t *testing.T) {
	t.Run("Successful tenant verification", func(t *testing.T) {
		// Setup mocks
		mockOTPRepo := new(mocks.MockOTPRepository)
		mockTenantRepo := new(mocks.MockTenantRepository)
		mockAuditRepo := new(mocks.MockAuditRepository)

		tenantID := "tenant-123"
		otpCode := "123456"

		// Create test OTP
		testOTP := &entities.OTPVerification{
			TenantID:  tenantID,
			Code:      otpCode,
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			Verified:  false,
		}

		// Create test tenant
		testTenant := &entities.Tenant{
			ID:               tenantID,
			OrganizationName: "Test Org",
			Status:           entities.TenantStatusPendingVerification,
		}

		// Mock expectations
		mockOTPRepo.On("FindByTenantIDAndPurpose", tenantID, "email_verification").Return(testOTP, nil)
		mockOTPRepo.On("Update", mock.MatchedBy(func(otp *entities.OTPVerification) bool {
			return otp.Verified == true
		})).Return(nil)
		mockTenantRepo.On("FindByID", tenantID).Return(testTenant, nil)
		mockTenantRepo.On("Update", mock.MatchedBy(func(t *entities.Tenant) bool {
			return t.Status == entities.TenantStatusActive
		})).Return(nil)
		mockAuditRepo.On("Create", mock.MatchedBy(func(log *entities.AuditLog) bool {
			return log.EventType == "tenant.verified" && log.TenantID == tenantID
		})).Return(nil)

		// Create use case
		useCase := tenant.NewVerifyTenantUseCase(mockOTPRepo, mockTenantRepo, mockAuditRepo)

		// Execute
		err := useCase.Execute(tenantID, otpCode)

		// Assert
		assert.NoError(t, err)
		mockOTPRepo.AssertExpectations(t)
		mockTenantRepo.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t)
	})

	t.Run("OTP not found", func(t *testing.T) {
		mockOTPRepo := new(mocks.MockOTPRepository)
		mockTenantRepo := new(mocks.MockTenantRepository)
		mockAuditRepo := new(mocks.MockAuditRepository)

		tenantID := "tenant-123"
		otpCode := "123456"

		mockOTPRepo.On("FindByTenantIDAndPurpose", tenantID, "email_verification").
			Return(nil, errors.New("OTP not found"))

		useCase := tenant.NewVerifyTenantUseCase(mockOTPRepo, mockTenantRepo, mockAuditRepo)

		err := useCase.Execute(tenantID, otpCode)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "OTP not found")
		mockOTPRepo.AssertExpectations(t)
	})

	t.Run("OTP already verified", func(t *testing.T) {
		mockOTPRepo := new(mocks.MockOTPRepository)
		mockTenantRepo := new(mocks.MockTenantRepository)
		mockAuditRepo := new(mocks.MockAuditRepository)

		tenantID := "tenant-123"
		otpCode := "123456"

		verifiedAt := time.Now().Add(-1 * time.Hour)
		testOTP := &entities.OTPVerification{
			TenantID:   tenantID,
			Code:       otpCode,
			Purpose:    "email_verification",
			ExpiresAt:  time.Now().Add(10 * time.Minute),
			Verified:   true,
			VerifiedAt: &verifiedAt,
		}

		mockOTPRepo.On("FindByTenantIDAndPurpose", tenantID, "email_verification").Return(testOTP, nil)

		useCase := tenant.NewVerifyTenantUseCase(mockOTPRepo, mockTenantRepo, mockAuditRepo)

		err := useCase.Execute(tenantID, otpCode)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already been used")
		mockOTPRepo.AssertExpectations(t)
	})

	t.Run("OTP expired", func(t *testing.T) {
		mockOTPRepo := new(mocks.MockOTPRepository)
		mockTenantRepo := new(mocks.MockTenantRepository)
		mockAuditRepo := new(mocks.MockAuditRepository)

		tenantID := "tenant-123"
		otpCode := "123456"

		testOTP := &entities.OTPVerification{
			TenantID:  tenantID,
			Code:      otpCode,
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(-1 * time.Minute), // Expired
			Verified:  false,
		}

		mockOTPRepo.On("FindByTenantIDAndPurpose", tenantID, "email_verification").Return(testOTP, nil)

		useCase := tenant.NewVerifyTenantUseCase(mockOTPRepo, mockTenantRepo, mockAuditRepo)

		err := useCase.Execute(tenantID, otpCode)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
		mockOTPRepo.AssertExpectations(t)
	})

	t.Run("Invalid OTP code", func(t *testing.T) {
		mockOTPRepo := new(mocks.MockOTPRepository)
		mockTenantRepo := new(mocks.MockTenantRepository)
		mockAuditRepo := new(mocks.MockAuditRepository)

		tenantID := "tenant-123"
		correctCode := "123456"
		wrongCode := "654321"

		testOTP := &entities.OTPVerification{
			TenantID:  tenantID,
			Code:      correctCode,
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			Verified:  false,
		}

		mockOTPRepo.On("FindByTenantIDAndPurpose", tenantID, "email_verification").Return(testOTP, nil)

		useCase := tenant.NewVerifyTenantUseCase(mockOTPRepo, mockTenantRepo, mockAuditRepo)

		err := useCase.Execute(tenantID, wrongCode)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")
		mockOTPRepo.AssertExpectations(t)
	})

	t.Run("Tenant not found", func(t *testing.T) {
		mockOTPRepo := new(mocks.MockOTPRepository)
		mockTenantRepo := new(mocks.MockTenantRepository)
		mockAuditRepo := new(mocks.MockAuditRepository)

		tenantID := "tenant-123"
		otpCode := "123456"

		testOTP := &entities.OTPVerification{
			TenantID:  tenantID,
			Code:      otpCode,
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			Verified:  false,
		}

		mockOTPRepo.On("FindByTenantIDAndPurpose", tenantID, "email_verification").Return(testOTP, nil)
		mockOTPRepo.On("Update", mock.Anything).Return(nil)
		mockTenantRepo.On("FindByID", tenantID).Return(nil, errors.New("tenant not found"))

		useCase := tenant.NewVerifyTenantUseCase(mockOTPRepo, mockTenantRepo, mockAuditRepo)

		err := useCase.Execute(tenantID, otpCode)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tenant not found")
		mockOTPRepo.AssertExpectations(t)
		mockTenantRepo.AssertExpectations(t)
	})

	t.Run("Tenant already active", func(t *testing.T) {
		mockOTPRepo := new(mocks.MockOTPRepository)
		mockTenantRepo := new(mocks.MockTenantRepository)
		mockAuditRepo := new(mocks.MockAuditRepository)

		tenantID := "tenant-123"
		otpCode := "123456"

		testOTP := &entities.OTPVerification{
			TenantID:  tenantID,
			Code:      otpCode,
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			Verified:  false,
		}

		testTenant := &entities.Tenant{
			ID:               tenantID,
			OrganizationName: "Test Org",
			Status:           entities.TenantStatusActive, // Already active
		}

		mockOTPRepo.On("FindByTenantIDAndPurpose", tenantID, "email_verification").Return(testOTP, nil)
		mockOTPRepo.On("Update", mock.Anything).Return(nil)
		mockTenantRepo.On("FindByID", tenantID).Return(testTenant, nil)

		useCase := tenant.NewVerifyTenantUseCase(mockOTPRepo, mockTenantRepo, mockAuditRepo)

		err := useCase.Execute(tenantID, otpCode)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already active")
		mockOTPRepo.AssertExpectations(t)
		mockTenantRepo.AssertExpectations(t)
	})

	t.Run("Database error when updating OTP", func(t *testing.T) {
		mockOTPRepo := new(mocks.MockOTPRepository)
		mockTenantRepo := new(mocks.MockTenantRepository)
		mockAuditRepo := new(mocks.MockAuditRepository)

		tenantID := "tenant-123"
		otpCode := "123456"

		testOTP := &entities.OTPVerification{
			TenantID:  tenantID,
			Code:      otpCode,
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			Verified:  false,
		}

		mockOTPRepo.On("FindByTenantIDAndPurpose", tenantID, "email_verification").Return(testOTP, nil)
		mockOTPRepo.On("Update", mock.Anything).Return(errors.New("database error"))

		useCase := tenant.NewVerifyTenantUseCase(mockOTPRepo, mockTenantRepo, mockAuditRepo)

		err := useCase.Execute(tenantID, otpCode)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		mockOTPRepo.AssertExpectations(t)
	})

	t.Run("Database error when updating tenant", func(t *testing.T) {
		mockOTPRepo := new(mocks.MockOTPRepository)
		mockTenantRepo := new(mocks.MockTenantRepository)
		mockAuditRepo := new(mocks.MockAuditRepository)

		tenantID := "tenant-123"
		otpCode := "123456"

		testOTP := &entities.OTPVerification{
			TenantID:  tenantID,
			Code:      otpCode,
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			Verified:  false,
		}

		testTenant := &entities.Tenant{
			ID:               tenantID,
			OrganizationName: "Test Org",
			Status:           entities.TenantStatusPendingVerification,
		}

		mockOTPRepo.On("FindByTenantIDAndPurpose", tenantID, "email_verification").Return(testOTP, nil)
		mockOTPRepo.On("Update", mock.Anything).Return(nil)
		mockTenantRepo.On("FindByID", tenantID).Return(testTenant, nil)
		mockTenantRepo.On("Update", mock.Anything).Return(errors.New("database error"))

		useCase := tenant.NewVerifyTenantUseCase(mockOTPRepo, mockTenantRepo, mockAuditRepo)

		err := useCase.Execute(tenantID, otpCode)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		mockOTPRepo.AssertExpectations(t)
		mockTenantRepo.AssertExpectations(t)
	})

	t.Run("Audit log creation fails but verification succeeds", func(t *testing.T) {
		// Audit logging should not block the verification process
		mockOTPRepo := new(mocks.MockOTPRepository)
		mockTenantRepo := new(mocks.MockTenantRepository)
		mockAuditRepo := new(mocks.MockAuditRepository)

		tenantID := "tenant-123"
		otpCode := "123456"

		testOTP := &entities.OTPVerification{
			TenantID:  tenantID,
			Code:      otpCode,
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			Verified:  false,
		}

		testTenant := &entities.Tenant{
			ID:               tenantID,
			OrganizationName: "Test Org",
			Status:           entities.TenantStatusPendingVerification,
		}

		mockOTPRepo.On("FindByTenantIDAndPurpose", tenantID, "email_verification").Return(testOTP, nil)
		mockOTPRepo.On("Update", mock.Anything).Return(nil)
		mockTenantRepo.On("FindByID", tenantID).Return(testTenant, nil)
		mockTenantRepo.On("Update", mock.Anything).Return(nil)
		mockAuditRepo.On("Create", mock.Anything).Return(errors.New("audit log error"))

		useCase := tenant.NewVerifyTenantUseCase(mockOTPRepo, mockTenantRepo, mockAuditRepo)

		err := useCase.Execute(tenantID, otpCode)

		// Should succeed despite audit log error
		assert.NoError(t, err)
		mockOTPRepo.AssertExpectations(t)
		mockTenantRepo.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t)
	})
}
