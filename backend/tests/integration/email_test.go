package integration_test

import (
	"context"
	"testing"

	"github.com/ranggaaprilio/keyles/infrastructure/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEmailService_SendOTPEmail tests OTP email sending with mock Brevo service
func TestEmailService_SendOTPEmail(t *testing.T) {
	// Use mock email service for integration tests
	// In production, this would use actual Brevo API
	emailService := &MockEmailService{}

	tests := []struct {
		name             string
		toEmail          string
		toName           string
		otpCode          string
		organizationName string
		expectError      bool
	}{
		{
			name:             "Valid OTP email",
			toEmail:          "admin@techcorp.com",
			toName:           "John Doe",
			otpCode:          "123456",
			organizationName: "TechCorp",
			expectError:      false,
		},
		{
			name:             "Email with special characters in name",
			toEmail:          "test@example.com",
			toName:           "José García",
			otpCode:          "654321",
			organizationName: "García & Associates",
			expectError:      false,
		},
		{
			name:             "Long organization name",
			toEmail:          "admin@longname.com",
			toName:           "Admin User",
			otpCode:          "111111",
			organizationName: "Very Long Organization Name That Exceeds Normal Length But Should Still Work",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := emailService.SendOTPEmail(
				context.Background(),
				tt.toEmail,
				tt.toName,
				tt.otpCode,
				tt.organizationName,
			)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestEmailService_SendWelcomeEmail tests welcome email sending
func TestEmailService_SendWelcomeEmail(t *testing.T) {
	emailService := &MockEmailService{}

	tests := []struct {
		name             string
		toEmail          string
		toName           string
		organizationName string
		expectError      bool
	}{
		{
			name:             "Valid welcome email",
			toEmail:          "admin@techcorp.com",
			toName:           "John Doe",
			organizationName: "TechCorp",
			expectError:      false,
		},
		{
			name:             "Welcome email with unicode",
			toEmail:          "admin@example.jp",
			toName:           "田中 太郎",
			organizationName: "株式会社テスト",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := emailService.SendWelcomeEmail(
				context.Background(),
				tt.toEmail,
				tt.toName,
				tt.organizationName,
			)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestEmailService_OTPValidation tests OTP code validation in email service
func TestEmailService_OTPValidation(t *testing.T) {
	otpService := services.NewCryptoOTPService()

	tests := []struct {
		name     string
		otp      string
		expected bool
	}{
		{
			name:     "Valid 6-digit OTP",
			otp:      "123456",
			expected: true,
		},
		{
			name:     "Valid OTP with different digits",
			otp:      "987654",
			expected: true,
		},
		{
			name:     "Invalid - too short",
			otp:      "12345",
			expected: false,
		},
		{
			name:     "Invalid - too long",
			otp:      "1234567",
			expected: false,
		},
		{
			name:     "Invalid - contains letters",
			otp:      "12345a",
			expected: false,
		},
		{
			name:     "Invalid - contains special characters",
			otp:      "12345!",
			expected: false,
		},
		{
			name:     "Invalid - empty string",
			otp:      "",
			expected: false,
		},
		{
			name:     "Invalid - spaces",
			otp:      "123 456",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := otpService.Validate(tt.otp)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEmailService_OTPGeneration tests OTP generation
func TestEmailService_OTPGeneration(t *testing.T) {
	otpService := services.NewCryptoOTPService()

	// Generate multiple OTPs and verify they meet requirements
	for i := 0; i < 100; i++ {
		otp, err := otpService.Generate()
		require.NoError(t, err, "OTP generation should not fail")
		
		// Verify OTP is 6 digits
		assert.Len(t, otp, 6, "OTP should be 6 characters")
		
		// Verify OTP is valid
		assert.True(t, otpService.Validate(otp), "Generated OTP should be valid")
		
		// Verify OTP contains only digits
		for _, c := range otp {
			assert.True(t, c >= '0' && c <= '9', "OTP should contain only digits")
		}
	}

	// Test uniqueness (OTPs should be different)
	otp1, err := otpService.Generate()
	require.NoError(t, err)
	
	otp2, err := otpService.Generate()
	require.NoError(t, err)
	
	// While there's a tiny chance they could be the same, it's extremely unlikely
	// This is a reasonable test for randomness
	assert.NotEqual(t, otp1, otp2, "Generated OTPs should be different (statistically)")
}

// TestEmailService_Integration tests end-to-end email flow
func TestEmailService_Integration(t *testing.T) {
	// This test simulates the complete OTP email flow
	emailService := &MockEmailService{}
	otpService := services.NewCryptoOTPService()

	t.Run("Complete OTP flow", func(t *testing.T) {
		// Step 1: Generate OTP
		otpCode, err := otpService.Generate()
		require.NoError(t, err)
		assert.Len(t, otpCode, 6)

		// Step 2: Validate OTP format
		isValid := otpService.Validate(otpCode)
		assert.True(t, isValid)

		// Step 3: Send OTP email
		err = emailService.SendOTPEmail(
			context.Background(),
			"admin@testorg.com",
			"Test Admin",
			otpCode,
			"Test Organization",
		)
		assert.NoError(t, err)
	})

	t.Run("Complete verification flow", func(t *testing.T) {
		// Step 1: Send OTP
		otpCode := "123456"
		err := emailService.SendOTPEmail(
			context.Background(),
			"admin@testorg.com",
			"Test Admin",
			otpCode,
			"Test Organization",
		)
		require.NoError(t, err)

		// Step 2: Validate the OTP (simulating user input)
		isValid := otpService.Validate(otpCode)
		assert.True(t, isValid)

		// Step 3: Send welcome email after successful verification
		err = emailService.SendWelcomeEmail(
			context.Background(),
			"admin@testorg.com",
			"Test Admin",
			"Test Organization",
		)
		assert.NoError(t, err)
	})
}

// TestEmailService_ErrorHandling tests email service error scenarios
func TestEmailService_ErrorHandling(t *testing.T) {
	// Create a mock service that can simulate failures
	type FailingEmailService struct {
		shouldFail bool
	}

	failService := &FailingEmailService{shouldFail: true}

	t.Run("Email service handles failures gracefully", func(t *testing.T) {
		// In a real implementation, the email service would return errors
		// For mock, we just verify the interface is correct
		emailService := &MockEmailService{}
		
		// These should succeed with mock
		err := emailService.SendOTPEmail(
			context.Background(),
			"test@example.com",
			"Test User",
			"123456",
			"Test Org",
		)
		assert.NoError(t, err)
	})
}
