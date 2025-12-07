package tenant

import (
	"context"
	"fmt"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// CheckAvailabilityResult contains the availability check results
type CheckAvailabilityResult struct {
	OrganizationNameAvailable bool
	EmailAvailable            bool
}

// CheckAvailabilityUseCase handles checking organization name and email availability
type CheckAvailabilityUseCase struct {
	tenantRepo repositories.TenantRepository
	userRepo   repositories.UserRepository
}

// NewCheckAvailabilityUseCase creates a new CheckAvailabilityUseCase
func NewCheckAvailabilityUseCase(
	tenantRepo repositories.TenantRepository,
	userRepo repositories.UserRepository,
) *CheckAvailabilityUseCase {
	return &CheckAvailabilityUseCase{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
	}
}

// Execute checks if organization name and email are available
func (uc *CheckAvailabilityUseCase) Execute(
	ctx context.Context,
	organizationName string,
	email string,
) (*CheckAvailabilityResult, error) {
	// Validate inputs
	if err := entities.ValidateOrganizationName(organizationName); err != nil {
		return nil, err
	}

	if err := entities.ValidateEmail(email); err != nil {
		return nil, err
	}

	// Check organization name availability
	orgExists, err := uc.tenantRepo.OrganizationNameExists(ctx, organizationName)
	if err != nil {
		return nil, fmt.Errorf("failed to check organization name availability: %w", err)
	}

	// Check email availability
	emailExists, err := uc.userRepo.EmailExists(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email availability: %w", err)
	}

	return &CheckAvailabilityResult{
		OrganizationNameAvailable: !orgExists,
		EmailAvailable:            !emailExists,
	}, nil
}
