package services

import "context"

// EmailService defines the interface for email operations
type EmailService interface {
	// SendOTPEmail sends an OTP verification email to the specified address
	SendOTPEmail(ctx context.Context, toEmail, toName, otpCode, organizationName string) error

	// SendWelcomeEmail sends a welcome email after successful verification
	SendWelcomeEmail(ctx context.Context, toEmail, toName, organizationName string) error

	// SendInvitationEmail sends an invitation email with a link for the user to activate their account
	SendInvitationEmail(ctx context.Context, toEmail, toName, inviteURL, orgName string) error

	// SendSignupOTPEmail sends the verification code for an end-user self sign-up
	SendSignupOTPEmail(ctx context.Context, toEmail, toName, otpCode, clientName string) error
}
