package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockEmailService is a mock implementation of EmailService interface
type MockEmailService struct {
	mock.Mock
}

// SendOTPEmail sends an OTP verification email
func (m *MockEmailService) SendOTPEmail(ctx context.Context, toEmail, toName, otpCode, organizationName string) error {
	args := m.Called(ctx, toEmail, toName, otpCode, organizationName)
	return args.Error(0)
}

// SendWelcomeEmail sends a welcome email after verification
func (m *MockEmailService) SendWelcomeEmail(ctx context.Context, toEmail, toName, organizationName string) error {
	args := m.Called(ctx, toEmail, toName, organizationName)
	return args.Error(0)
}

// SendInvitationEmail sends an invitation email to a prospective user
func (m *MockEmailService) SendInvitationEmail(ctx context.Context, toEmail, toName, inviteURL, orgName string) error {
	args := m.Called(ctx, toEmail, toName, inviteURL, orgName)
	return args.Error(0)
}
