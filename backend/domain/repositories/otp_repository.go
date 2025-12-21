package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
)

// OTPRepository defines the interface for OTP persistence operations (Redis-based)
type OTPRepository interface {
	// Create stores an OTP verification record
	Create(ctx context.Context, otp *entities.OTPVerification) error

	// FindByTenantIDAndPurpose retrieves the OTP for a tenant and purpose
	FindByTenantIDAndPurpose(ctx context.Context, tenantID, purpose string) (*entities.OTPVerification, error)

	// Update updates an OTP verification record (e.g., mark as verified)
	Update(ctx context.Context, otp *entities.OTPVerification) error

	// Delete removes an OTP verification record
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteExpired removes all expired OTP records
	DeleteExpired(ctx context.Context) error

	// IncrementRateLimitCounter increments the OTP request counter for an email
	// Returns the current count
	IncrementRateLimitCounter(ctx context.Context, email string, window time.Duration) (int, error)

	// GetRateLimitCounter gets the current OTP request count for an email
	GetRateLimitCounter(ctx context.Context, email string) (int, error)
}
