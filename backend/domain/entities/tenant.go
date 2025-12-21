package entities

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TenantStatus represents the lifecycle status of a tenant
type TenantStatus string

const (
	TenantStatusPendingVerification TenantStatus = "pending_verification"
	TenantStatusActive              TenantStatus = "active"
	TenantStatusSuspended           TenantStatus = "suspended"
	TenantStatusDeleted             TenantStatus = "deleted"
)

// Tenant represents an organization in the multi-tenant SSO platform
type Tenant struct {
	ID               uuid.UUID
	OrganizationName string
	Status           TenantStatus
	CreatedAt        time.Time
	VerifiedAt       *time.Time
	UpdatedAt        time.Time
}

var (
	// Organization name validation: 3-100 characters, alphanumeric + spaces + hyphens
	orgNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-]{3,100}$`)

	// Validation errors
	ErrInvalidOrganizationName = errors.New("organization name must be between 3 and 100 characters")
	ErrInvalidTenantStatus     = errors.New("invalid tenant status")
)

// NewTenant creates a new tenant with pending verification status
func NewTenant(organizationName string) (*Tenant, error) {
	if err := ValidateOrganizationName(organizationName); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Tenant{
		ID:               uuid.New(),
		OrganizationName: NormalizeOrganizationName(organizationName),
		Status:           TenantStatusPendingVerification,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

// ValidateOrganizationName validates the organization name format
func ValidateOrganizationName(name string) error {
	trimmed := strings.TrimSpace(name)
	if !orgNameRegex.MatchString(trimmed) {
		return ErrInvalidOrganizationName
	}
	return nil
}

// NormalizeOrganizationName normalizes the organization name (trim spaces, consistent casing)
func NormalizeOrganizationName(name string) string {
	// Trim leading/trailing spaces and collapse multiple spaces
	normalized := strings.Join(strings.Fields(name), " ")
	return normalized
}

// Activate marks the tenant as active after successful verification
func (t *Tenant) Activate() error {
	if t.Status != TenantStatusPendingVerification {
		return errors.New("tenant must be in pending_verification status to activate")
	}

	now := time.Now()
	t.Status = TenantStatusActive
	t.VerifiedAt = &now
	t.UpdatedAt = now
	return nil
}

// Suspend suspends an active tenant
func (t *Tenant) Suspend() error {
	if t.Status != TenantStatusActive {
		return errors.New("only active tenants can be suspended")
	}

	t.Status = TenantStatusSuspended
	t.UpdatedAt = time.Now()
	return nil
}

// IsActive returns true if the tenant is active
func (t *Tenant) IsActive() bool {
	return t.Status == TenantStatusActive
}

// IsPendingVerification returns true if the tenant is pending verification
func (t *Tenant) IsPendingVerification() bool {
	return t.Status == TenantStatusPendingVerification
}

// IsSuspended returns true if the tenant is suspended
func (t *Tenant) IsSuspended() bool {
	return t.Status == TenantStatusSuspended
}
