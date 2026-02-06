package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockPasswordService is a mock implementation of PasswordService
type MockPasswordService struct {
	mock.Mock
}

func (m *MockPasswordService) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

// Compare mocks password comparison (for auth.PasswordService)
func (m *MockPasswordService) Compare(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

// Verify mocks password verification (for domain/services.PasswordService)
func (m *MockPasswordService) Verify(password, hash string) error {
	args := m.Called(password, hash)
	return args.Error(0)
}
