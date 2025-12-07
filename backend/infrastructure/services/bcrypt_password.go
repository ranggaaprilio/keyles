package services

import (
	"fmt"

	"github.com/ranggaaprilio/keyles/domain/services"
	"golang.org/x/crypto/bcrypt"
)

// BcryptPasswordService implements PasswordService using bcrypt
type BcryptPasswordService struct {
	cost int
}

// NewBcryptPasswordService creates a new bcrypt password service with cost factor 12
func NewBcryptPasswordService() services.PasswordService {
	return &BcryptPasswordService{
		cost: 12, // Cost factor 12 as per FR-018 and constitution
	}
}

// Hash generates a bcrypt hash of the password
func (s *BcryptPasswordService) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// Verify checks if the provided password matches the hash
func (s *BcryptPasswordService) Verify(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("password verification failed: %w", err)
	}
	return nil
}
