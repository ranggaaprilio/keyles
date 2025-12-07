package entities

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// OTPVerificationStatus represents the status of an OTP verification
type OTPVerificationStatus string

const (
	OTPStatusPending  OTPVerificationStatus = "pending"
	OTPStatusVerified OTPVerificationStatus = "verified"
	OTPStatusExpired  OTPVerificationStatus = "expired"
)

// OTPVerification represents an email verification code for tenant activation
type OTPVerification struct {
	ID           uuid.UUID
	TenantID     string    // Changed to string for simpler validation
	Code         string    // Renamed from OTPCode
	Purpose      string    // email_verification or password_reset
	CreatedAt    time.Time
	ExpiresAt    time.Time
	VerifiedAt   *time.Time
	Verified     bool      // Added for simpler checks
	AttemptCount int
	IPAddress    string
}

const (
	// OTP configuration
	OTPLength          = 6
	OTPExpirationMins  = 10
	MaxOTPAttempts     = 5
)

var (
	// Validation errors
	ErrOTPExpired     = errors.New("OTP has expired")
	ErrOTPInvalid     = errors.New("invalid OTP code")
	ErrOTPMaxAttempts = errors.New("maximum OTP verification attempts exceeded")
	ErrOTPAlreadyUsed = errors.New("OTP has already been used")
	otpCodeRegex      = regexp.MustCompile(`^\d{6}$`)
)

// Validate checks if the OTP verification has valid data
func (o *OTPVerification) Validate() error {
	if o.TenantID == "" {
		return errors.New("tenant_id is required")
	}

	if o.Code == "" {
		return errors.New("code is required")
	}

	if len(o.Code) != 6 {
		return errors.New("code must be exactly 6 digits")
	}

	if !otpCodeRegex.MatchString(o.Code) {
		return errors.New("code must contain only digits")
	}

	if o.Purpose == "" {
		return errors.New("purpose is required")
	}

	if o.Purpose != "email_verification" && o.Purpose != "password_reset" {
		return errors.New("purpose must be 'email_verification' or 'password_reset'")
	}

	if o.ExpiresAt.IsZero() {
		return errors.New("expires_at is required")
	}

	return nil
}

// IsExpired checks if the OTP has expired
func (o *OTPVerification) IsExpired() bool {
	return time.Now().After(o.ExpiresAt) || time.Now().Equal(o.ExpiresAt)
}

// IsVerified checks if the OTP has been successfully verified
func (o *OTPVerification) IsVerified() bool {
	return o.Verified
}

// MarkAsVerified marks the OTP as verified
func (o *OTPVerification) MarkAsVerified() {
	now := time.Now()
	o.Verified = true
	o.VerifiedAt = &now
}

// CanBeVerified checks if the OTP can be verified (not expired and not already verified)
func (o *OTPVerification) CanBeVerified() bool {
	return !o.IsExpired() && !o.IsVerified()
}

// Invalidate invalidates the OTP by setting its expiration to the past
func (o *OTPVerification) Invalidate() {
	o.ExpiresAt = time.Now().Add(-1 * time.Second)
}

// NewOTPVerification creates a new OTP verification record
func NewOTPVerification(tenantID, otpCode, purpose, ipAddress string) *OTPVerification {
	now := time.Now()
	return &OTPVerification{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Code:         otpCode,
		Purpose:      purpose,
		CreatedAt:    now,
		ExpiresAt:    now.Add(OTPExpirationMins * time.Minute),
		Verified:     false,
		AttemptCount: 0,
		IPAddress:    ipAddress,
	}
}

// VerifyCode attempts to verify the provided OTP code
func (o *OTPVerification) VerifyCode(providedOTP string) error {
	// Check if already verified
	if o.IsVerified() {
		return ErrOTPAlreadyUsed
	}

	// Check if expired
	if o.IsExpired() {
		return ErrOTPExpired
	}

	// Increment attempt count
	o.AttemptCount++

	// Check max attempts
	if o.AttemptCount > MaxOTPAttempts {
		o.Invalidate()
		return ErrOTPMaxAttempts
	}

	// Verify OTP code
	if o.Code != providedOTP {
		return ErrOTPInvalid
	}

	// Success - mark as verified
	o.MarkAsVerified()
	return nil
}

// CanRetry checks if the user can still attempt verification
func (o *OTPVerification) CanRetry() bool {
	return o.CanBeVerified() && o.AttemptCount < MaxOTPAttempts
}

// RemainingAttempts returns the number of remaining verification attempts
func (o *OTPVerification) RemainingAttempts() int {
	remaining := MaxOTPAttempts - o.AttemptCount
	if remaining < 0 {
		return 0
	}
	return remaining
}

// TimeUntilExpiration returns the duration until the OTP expires
func (o *OTPVerification) TimeUntilExpiration() time.Duration {
	return time.Until(o.ExpiresAt)
}

// String returns a string representation of the OTP (for logging - masked)
func (o *OTPVerification) String() string {
	return fmt.Sprintf("OTP{TenantID: %s, Purpose: %s, Verified: %v, Expired: %v}", 
		o.TenantID, o.Purpose, o.Verified, o.IsExpired())
}
