package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/services"
)

// CryptoOTPService implements OTPService using crypto/rand
type CryptoOTPService struct{}

// NewCryptoOTPService creates a new crypto-based OTP service
func NewCryptoOTPService() services.OTPService {
	return &CryptoOTPService{}
}

// Generate generates a cryptographically random OTP code
func (s *CryptoOTPService) Generate() (string, error) {
	// Generate a number between 0 and 999999
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random OTP: %w", err)
	}

	// Format as 6-digit string with leading zeros
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// Validate validates an OTP code format
func (s *CryptoOTPService) Validate(otp string) bool {
	// OTP must be exactly 6 digits
	matched, _ := regexp.MatchString(`^\d{6}$`, otp)
	return matched && len(otp) == entities.OTPLength
}
