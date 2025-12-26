package mocks

import (
	"context"

	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/mock"
)

// MockOTPRepository is a mock implementation of OTPRepository
type MockOTPRepository struct {
	mock.Mock
}

func (m *MockOTPRepository) Create(ctx context.Context, otp *entities.OTPVerification) error {
	args := m.Called(ctx, otp)
	return args.Error(0)
}

func (m *MockOTPRepository) FindByTenantIDAndPurpose(ctx context.Context, tenantID, purpose string) (*entities.OTPVerification, error) {
	args := m.Called(ctx, tenantID, purpose)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.OTPVerification), args.Error(1)
}

func (m *MockOTPRepository) Update(ctx context.Context, otp *entities.OTPVerification) error {
	args := m.Called(ctx, otp)
	return args.Error(0)
}

func (m *MockOTPRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOTPRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockOTPRepository) IncrementRateLimitCounter(ctx context.Context, email string, window time.Duration) (int, error) {
	args := m.Called(ctx, email, window)
	return args.Int(0), args.Error(1)
}

func (m *MockOTPRepository) GetRateLimitCounter(ctx context.Context, email string) (int, error) {
	args := m.Called(ctx, email)
	return args.Int(0), args.Error(1)
}
