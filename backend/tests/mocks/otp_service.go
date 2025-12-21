package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockOTPService is a mock implementation of OTPService interface
type MockOTPService struct {
	mock.Mock
}

// Generate generates a 6-digit numeric OTP code
func (m *MockOTPService) Generate() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// Validate validates an OTP code format (6 digits)
func (m *MockOTPService) Validate(otp string) bool {
	args := m.Called(otp)
	return args.Bool(0)
}
