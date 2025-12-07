package tenant

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
)

var (
	ErrOrganizationNameExists = errors.New("organization name already exists")
	ErrEmailExists            = errors.New("email already exists")
)

// RegisterTenantResult contains the result of tenant registration
type RegisterTenantResult struct {
	TenantID         uuid.UUID
	AdminUserID      uuid.UUID
	OrganizationName string
	Status           entities.TenantStatus
	Message          string
}

// RegisterTenantUseCase handles tenant registration
type RegisterTenantUseCase struct {
	tenantRepo repositories.TenantRepository
	userRepo   repositories.UserRepository
	otpRepo    repositories.OTPRepository
	auditRepo  repositories.AuditRepository
	emailSvc   services.EmailService
	otpSvc     services.OTPService
	pwdSvc     services.PasswordService
}

// NewRegisterTenantUseCase creates a new RegisterTenantUseCase
func NewRegisterTenantUseCase(
	tenantRepo repositories.TenantRepository,
	userRepo repositories.UserRepository,
	otpRepo repositories.OTPRepository,
	auditRepo repositories.AuditRepository,
	emailSvc services.EmailService,
	otpSvc services.OTPService,
	pwdSvc services.PasswordService,
) *RegisterTenantUseCase {
	return &RegisterTenantUseCase{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
		otpRepo:    otpRepo,
		auditRepo:  auditRepo,
		emailSvc:   emailSvc,
		otpSvc:     otpSvc,
		pwdSvc:     pwdSvc,
	}
}

// Execute registers a new tenant with admin user
func (uc *RegisterTenantUseCase) Execute(
	ctx context.Context,
	organizationName string,
	adminEmail string,
	adminPassword string,
	adminFullName string,
) (*RegisterTenantResult, error) {
	// Step 1: Validate inputs
	if err := entities.ValidateOrganizationName(organizationName); err != nil {
		return nil, err
	}

	if err := entities.ValidateEmail(adminEmail); err != nil {
		return nil, err
	}

	if err := entities.ValidatePassword(adminPassword); err != nil {
		return nil, err
	}

	if err := entities.ValidateFullName(adminFullName); err != nil {
		return nil, err
	}

	// Step 2: Check uniqueness
	orgExists, err := uc.tenantRepo.OrganizationNameExists(ctx, organizationName)
	if err != nil {
		return nil, fmt.Errorf("failed to check organization name availability: %w", err)
	}
	if orgExists {
		return nil, ErrOrganizationNameExists
	}

	emailExists, err := uc.userRepo.EmailExists(ctx, adminEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to check email availability: %w", err)
	}
	if emailExists {
		return nil, ErrEmailExists
	}

	// Step 3: Hash password
	passwordHash, err := uc.pwdSvc.Hash(adminPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Step 4: Create tenant entity
	tenant, err := entities.NewTenant(organizationName)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant entity: %w", err)
	}

	// Step 5: Persist tenant
	if err := uc.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Step 6: Create admin user entity
	adminUser, err := entities.NewAdminUser(tenant.ID, adminFullName, adminEmail, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin user entity: %w", err)
	}

	// Step 7: Persist admin user
	if err := uc.userRepo.Create(ctx, adminUser); err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	// Step 8: Generate OTP
	otpCode, err := uc.otpSvc.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Step 9: Create OTP verification entity
	otpVerification := entities.NewOTPVerification(tenant.ID, otpCode, "")

	// Step 10: Store OTP
	if err := uc.otpRepo.Store(ctx, otpVerification); err != nil {
		return nil, fmt.Errorf("failed to store OTP: %w", err)
	}

	// Step 11: Send OTP email
	if err := uc.emailSvc.SendOTPEmail(ctx, adminEmail, adminFullName, otpCode, organizationName); err != nil {
		return nil, fmt.Errorf("failed to send OTP email: %w", err)
	}

	// Step 12: Create audit log
	auditLog := entities.NewAuditLog(entities.EventRegistrationSuccess, "", "").
		WithTenant(tenant.ID).
		WithUser(adminUser.ID).
		WithData("organization_name", organizationName).
		WithData("admin_email", adminEmail)

	if err := uc.auditRepo.Create(ctx, auditLog); err != nil {
		// Log error but don't fail the registration
		// In production, use proper logging framework
		_ = err
	}

	// Step 13: Return result
	return &RegisterTenantResult{
		TenantID:         tenant.ID,
		AdminUserID:      adminUser.ID,
		OrganizationName: tenant.OrganizationName,
		Status:           tenant.Status,
		Message:          "Registration successful. Please check your email for verification code.",
	}, nil
}
