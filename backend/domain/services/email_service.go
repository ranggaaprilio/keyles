package services

import "context"

// EmailService defines the interface for email operations
type EmailService interface {
	// SendOTPEmail sends an OTP verification email to the specified address
	SendOTPEmail(ctx context.Context, toEmail, toName, otpCode, organizationName string) error

	// SendWelcomeEmail sends a welcome email after successful verification
	SendWelcomeEmail(ctx context.Context, toEmail, toName, organizationName string) error
}
