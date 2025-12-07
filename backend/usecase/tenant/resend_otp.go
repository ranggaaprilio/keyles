package tenant

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
)

const (
	// MaxResendRequests is the maximum number of resend requests allowed per time window
	MaxResendRequests = 3
	// ResendRateLimitWindow is the time window for rate limiting (10 minutes)
	ResendRateLimitWindow = 10 * time.Minute
)

// ResendOTPUseCase handles the business logic for resending OTP codes
type ResendOTPUseCase struct {
	otpRepo      repositories.OTPRepository
	tenantRepo   repositories.TenantRepository
	userRepo     repositories.UserRepository
	emailService services.EmailService
	otpService   services.OTPService
	auditRepo    repositories.AuditRepository
}

// NewResendOTPUseCase creates a new instance of ResendOTPUseCase
func NewResendOTPUseCase(
	otpRepo repositories.OTPRepository,
	tenantRepo repositories.TenantRepository,
	userRepo repositories.UserRepository,
	emailService services.EmailService,
	otpService services.OTPService,
	auditRepo repositories.AuditRepository,
) *ResendOTPUseCase {
	return &ResendOTPUseCase{
		otpRepo:      otpRepo,
		tenantRepo:   tenantRepo,
		userRepo:     userRepo,
		emailService: emailService,
		otpService:   otpService,
		auditRepo:    auditRepo,
	}
}

// Execute performs the OTP resend operation
func (uc *ResendOTPUseCase) Execute(tenantID string) error {
	ctx := context.Background()

	// Parse tenant ID to UUID
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return fmt.Errorf("invalid tenant ID: %w", err)
	}

	// Step 1: Find tenant
	tenant, err := uc.tenantRepo.FindByID(ctx, tenantUUID)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	// Step 2: Find admin user for this tenant to get email
	users, err := uc.userRepo.FindByTenantID(ctx, tenantUUID)
	if err != nil || len(users) == 0 {
		return fmt.Errorf("admin user not found for tenant: %w", err)
	}
	
	// Use the first user (admin) for email
	adminUser := users[0]

	// Step 3: Check rate limiting - max 3 requests per 10 minutes
	requestCount, err := uc.otpRepo.IncrementRateLimitCounter(ctx, adminUser.Email, ResendRateLimitWindow)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}

	if requestCount > MaxResendRequests {
		return errors.New("rate limit exceeded: too many OTP resend requests. Please wait 10 minutes before trying again")
	}

	// Step 4: Find existing OTP for this tenant
	existingOTP, err := uc.otpRepo.FindByTenantIDAndPurpose(ctx, tenantID, "email_verification")
	if err != nil {
		return fmt.Errorf("OTP not found: %w", err)
	}

	// Step 5: Invalidate the old OTP
	existingOTP.Invalidate()
	if err := uc.otpRepo.Update(ctx, existingOTP); err != nil {
		return fmt.Errorf("failed to invalidate old OTP: %w", err)
	}

	// Step 6: Generate new OTP code
	newOTPCode, err := uc.otpService.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Step 7: Create new OTP verification record
	newOTP := &entities.OTPVerification{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Code:      newOTPCode,
		Purpose:   "email_verification",
		ExpiresAt: time.Now().Add(entities.OTPExpirationMins * time.Minute),
		CreatedAt: time.Now(),
		Verified:  false,
	}

	// Validate the new OTP
	if err := newOTP.Validate(); err != nil {
		return fmt.Errorf("invalid OTP data: %w", err)
	}

	// Step 8: Store the new OTP
	if err := uc.otpRepo.Create(ctx, newOTP); err != nil {
		return fmt.Errorf("failed to create new OTP: %w", err)
	}

	// Step 9: Send OTP via email
	if err := uc.emailService.SendOTPEmail(ctx, adminUser.Email, adminUser.FullName, newOTPCode, tenant.OrganizationName); err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	// Step 10: Create audit log (non-blocking - error doesn't prevent resend success)
	auditLog := &entities.AuditLog{
		ID:        uuid.New(),
		TenantID:  &tenant.ID,
		EventType: entities.EventOTPSent,
		EventData: map[string]interface{}{
			"action": "resend",
			"email":  adminUser.Email,
		},
		CreatedAt: time.Now(),
	}
	_ = uc.auditRepo.Create(ctx, auditLog)

	return nil
}
