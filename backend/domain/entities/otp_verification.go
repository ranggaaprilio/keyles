package entities

import (
	"errors"
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
	TenantID     uuid.UUID
	OTPCode      string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	VerifiedAt   *time.Time
	Status       OTPVerificationStatus
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
)

// NewOTPVerification creates a new OTP verification record
func NewOTPVerification(tenantID uuid.UUID, otpCode, ipAddress string) *OTPVerification {
	now := time.Now()
	return &OTPVerification{
		ID:           uuid.New(),
		TenantID:     tenantID,
		OTPCode:      otpCode,
		CreatedAt:    now,
		ExpiresAt:    now.Add(OTPExpirationMins * time.Minute),
		Status:       OTPStatusPending,
		AttemptCount: 0,
		IPAddress:    ipAddress,
	}
}

// Verify attempts to verify the OTP code
func (o *OTPVerification) Verify(providedOTP string) error {
	// Check if already verified
	if o.Status == OTPStatusVerified {
		return ErrOTPAlreadyUsed
	}

	// Check if expired
	if time.Now().After(o.ExpiresAt) {
		o.Status = OTPStatusExpired
		return ErrOTPExpired
	}

	// Increment attempt count
	o.AttemptCount++

	// Check max attempts
	if o.AttemptCount > MaxOTPAttempts {
		o.Status = OTPStatusExpired
		return ErrOTPMaxAttempts
	}

	// Verify OTP code
	if o.OTPCode != providedOTP {
		return ErrOTPInvalid
	}

	// Success - mark as verified
	now := time.Now()
	o.Status = OTPStatusVerified
	o.VerifiedAt = &now
	return nil
}

// IsExpired checks if the OTP has expired
func (o *OTPVerification) IsExpired() bool {
	return time.Now().After(o.ExpiresAt) || o.Status == OTPStatusExpired
}

// IsVerified checks if the OTP has been successfully verified
func (o *OTPVerification) IsVerified() bool {
	return o.Status == OTPStatusVerified
}

// CanRetry checks if the user can still attempt verification
func (o *OTPVerification) CanRetry() bool {
	return !o.IsExpired() && !o.IsVerified() && o.AttemptCount < MaxOTPAttempts
}
