package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockLoginThrottler is a mock implementation of LoginThrottler
type MockLoginThrottler struct {
	mock.Mock
}

// IsThrottled returns true when either the source-IP counter or the tenant-email counter has reached or exceeded the configured maximum
func (m *MockLoginThrottler) IsThrottled(ctx context.Context, sourceIP, tenantID, normalizedEmail string) (bool, error) {
	args := m.Called(ctx, sourceIP, tenantID, normalizedEmail)
	return args.Bool(0), args.Error(1)
}

// RecordFailure atomically increments both the source-IP counter and the tenant-email counter
func (m *MockLoginThrottler) RecordFailure(ctx context.Context, sourceIP, tenantID, normalizedEmail string) error {
	args := m.Called(ctx, sourceIP, tenantID, normalizedEmail)
	return args.Error(0)
}

// ClearEmailBucket removes the tenant-email counter after a successful login
func (m *MockLoginThrottler) ClearEmailBucket(ctx context.Context, tenantID, normalizedEmail string) error {
	args := m.Called(ctx, tenantID, normalizedEmail)
	return args.Error(0)
}