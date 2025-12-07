package tenant

import (
	"context"
	"errors"
	"fmt"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// VerifyTenantUseCase handles the OTP verification and tenant activation
type VerifyTenantUseCase struct {
	otpRepo    repositories.OTPRepository
	tenantRepo repositories.TenantRepository
	auditRepo  repositories.AuditRepository
}

// NewVerifyTenantUseCase creates a new verify tenant use case
func NewVerifyTenantUseCase(
	otpRepo repositories.OTPRepository,
	tenantRepo repositories.TenantRepository,
	auditRepo repositories.AuditRepository,
) *VerifyTenantUseCase {
	return &VerifyTenantUseCase{
		otpRepo:    otpRepo,
		tenantRepo: tenantRepo,
		auditRepo:  auditRepo,
	}
}

// Execute verifies the OTP and activates the tenant
func (uc *VerifyTenantUseCase) Execute(tenantID, otpCode string) error {
	ctx := context.Background()

	// Step 1: Find the OTP for email verification
	otp, err := uc.otpRepo.FindByTenantIDAndPurpose(ctx, tenantID, "email_verification")
	if err != nil {
		return fmt.Errorf("OTP not found: %w", err)
	}

	// Step 2: Check if OTP is already verified
	if otp.IsVerified() {
		return errors.New("OTP has already been used")
	}

	// Step 3: Check if OTP is expired
	if otp.IsExpired() {
		return errors.New("OTP has expired")
	}

	// Step 4: Verify the OTP code matches
	if otp.Code != otpCode {
		return errors.New("invalid OTP code")
	}

	// Step 5: Mark OTP as verified
	otp.MarkAsVerified()
	if err := uc.otpRepo.Update(ctx, otp); err != nil {
		return fmt.Errorf("failed to update OTP: %w", err)
	}

	// Step 6: Find the tenant
	tenant, err := uc.tenantRepo.FindByID(tenantID)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	// Step 7: Check if tenant is already active
	if tenant.IsActive() {
		return errors.New("tenant is already active")
	}

	// Step 8: Activate the tenant
	if err := tenant.Activate(); err != nil {
		return fmt.Errorf("failed to activate tenant: %w", err)
	}

	// Step 9: Update tenant in database
	if err := uc.tenantRepo.Update(tenant); err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	// Step 10: Create audit log (non-blocking - failure doesn't affect verification)
	auditLog := &entities.AuditLog{
		TenantID:  tenantID,
		EventType: "tenant.verified",
		Details:   map[string]interface{}{"organization_name": tenant.OrganizationName},
	}
	_ = uc.auditRepo.Create(ctx, auditLog) // Ignore error to not block verification

	return nil
}
