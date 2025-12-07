package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
)

// OTPRepository defines the interface for OTP persistence operations (Redis-based)
type OTPRepository interface {
	// Store stores an OTP verification record with TTL
	Store(ctx context.Context, otp *entities.OTPVerification) error

	// FindByTenantID retrieves the active OTP for a tenant
	FindByTenantID(ctx context.Context, tenantID uuid.UUID) (*entities.OTPVerification, error)

	// Update updates an OTP verification record (e.g., increment attempts)
	Update(ctx context.Context, otp *entities.OTPVerification) error

	// Delete removes an OTP verification record
	Delete(ctx context.Context, tenantID uuid.UUID) error

	// IncrementRateLimitCounter increments the OTP request counter for an email
	// Returns the current count
	IncrementRateLimitCounter(ctx context.Context, email string, window time.Duration) (int, error)

	// GetRateLimitCounter gets the current OTP request count for an email
	GetRateLimitCounter(ctx context.Context, email string) (int, error)
}
