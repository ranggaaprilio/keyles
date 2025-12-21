package services

// OTPService defines the interface for OTP generation and validation
type OTPService interface {
	// Generate generates a cryptographically random OTP code
	Generate() (string, error)

	// Validate validates an OTP code format
	Validate(otp string) bool
}
