package domain_test

import (
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/assert"
)

func TestOTPVerification_Validation(t *testing.T) {
	t.Run("Valid OTP verification", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			Verified:  false,
		}

		err := otp.Validate()
		assert.NoError(t, err)
	})

	t.Run("Missing tenant ID", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "",
			Code:      "123456",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}

		err := otp.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tenant_id is required")
	})

	t.Run("Missing OTP code", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}

		err := otp.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "code is required")
	})

	t.Run("Invalid OTP code length - too short", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "12345", // 5 digits instead of 6
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}

		err := otp.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "code must be exactly 6 digits")
	})

	t.Run("Invalid OTP code length - too long", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "1234567", // 7 digits instead of 6
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}

		err := otp.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "code must be exactly 6 digits")
	})

	t.Run("Invalid OTP code format - contains letters", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "12A456",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}

		err := otp.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "code must contain only digits")
	})

	t.Run("Invalid OTP code format - contains special characters", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123-56",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}

		err := otp.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "code must contain only digits")
	})

	t.Run("Missing purpose", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}

		err := otp.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "purpose is required")
	})

	t.Run("Invalid purpose value", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "invalid_purpose",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}

		err := otp.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "purpose must be 'email_verification' or 'password_reset'")
	})

	t.Run("Valid purpose - email_verification", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}

		err := otp.Validate()
		assert.NoError(t, err)
	})

	t.Run("Valid purpose - password_reset", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "password_reset",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}

		err := otp.Validate()
		assert.NoError(t, err)
	})

	t.Run("Missing expires_at", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "email_verification",
			ExpiresAt: time.Time{}, // Zero time
		}

		err := otp.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expires_at is required")
	})
}

func TestOTPVerification_IsExpired(t *testing.T) {
	t.Run("OTP is not expired", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(5 * time.Minute),
			Verified:  false,
		}

		assert.False(t, otp.IsExpired())
	})

	t.Run("OTP is expired - past expiration", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(-1 * time.Minute), // Expired 1 minute ago
			Verified:  false,
		}

		assert.True(t, otp.IsExpired())
	})

	t.Run("OTP is expired - exactly at expiration", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "email_verification",
			ExpiresAt: time.Now(),
			Verified:  false,
		}

		// Account for test execution time - should be expired or very close
		time.Sleep(10 * time.Millisecond)
		assert.True(t, otp.IsExpired())
	})

	t.Run("OTP just created - 10 minutes TTL", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			Verified:  false,
		}

		assert.False(t, otp.IsExpired())
	})
}

func TestOTPVerification_IsVerified(t *testing.T) {
	t.Run("OTP is verified", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			Verified:  true,
			VerifiedAt: func() *time.Time {
				t := time.Now()
				return &t
			}(),
		}

		assert.True(t, otp.IsVerified())
	})

	t.Run("OTP is not verified", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			Verified:  false,
			VerifiedAt: nil,
		}

		assert.False(t, otp.IsVerified())
	})
}

func TestOTPVerification_MarkAsVerified(t *testing.T) {
	t.Run("Mark OTP as verified", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			Verified:  false,
			VerifiedAt: nil,
		}

		before := time.Now()
		otp.MarkAsVerified()
		after := time.Now()

		assert.True(t, otp.Verified)
		assert.NotNil(t, otp.VerifiedAt)
		assert.True(t, otp.VerifiedAt.After(before) || otp.VerifiedAt.Equal(before))
		assert.True(t, otp.VerifiedAt.Before(after) || otp.VerifiedAt.Equal(after))
	})

	t.Run("Mark already verified OTP - idempotent", func(t *testing.T) {
		firstVerifiedAt := time.Now().Add(-1 * time.Minute)
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			Verified:  true,
			VerifiedAt: &firstVerifiedAt,
		}

		otp.MarkAsVerified()

		// Should update to new timestamp
		assert.True(t, otp.Verified)
		assert.NotNil(t, otp.VerifiedAt)
		assert.True(t, otp.VerifiedAt.After(firstVerifiedAt))
	})
}

func TestOTPVerification_CanBeVerified(t *testing.T) {
	t.Run("Can be verified - valid and not expired", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			Verified:  false,
		}

		assert.True(t, otp.CanBeVerified())
	})

	t.Run("Cannot be verified - already verified", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			Verified:  true,
			VerifiedAt: func() *time.Time {
				t := time.Now()
				return &t
			}(),
		}

		assert.False(t, otp.CanBeVerified())
	})

	t.Run("Cannot be verified - expired", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(-1 * time.Minute), // Expired
			Verified:  false,
		}

		assert.False(t, otp.CanBeVerified())
	})

	t.Run("Cannot be verified - both expired and verified", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(-1 * time.Minute), // Expired
			Verified:  true,
			VerifiedAt: func() *time.Time {
				t := time.Now()
				return &t
			}(),
		}

		assert.False(t, otp.CanBeVerified())
	})
}

func TestOTPVerification_Invalidate(t *testing.T) {
	t.Run("Invalidate OTP by setting expiration to past", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			Verified:  false,
		}

		assert.False(t, otp.IsExpired())

		otp.Invalidate()

		assert.True(t, otp.IsExpired())
		assert.True(t, otp.ExpiresAt.Before(time.Now()))
	})

	t.Run("Invalidate already expired OTP", func(t *testing.T) {
		otp := &entities.OTPVerification{
			TenantID:  "tenant-123",
			Code:      "123456",
			Purpose:   "email_verification",
			ExpiresAt: time.Now().Add(-5 * time.Minute), // Already expired
			Verified:  false,
		}

		assert.True(t, otp.IsExpired())

		otp.Invalidate()

		// Should still be expired
		assert.True(t, otp.IsExpired())
	})
}
