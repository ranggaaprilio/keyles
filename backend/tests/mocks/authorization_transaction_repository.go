package mocks

import (
	"context"
	"time"

	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/stretchr/testify/mock"
)

// MockAuthorizationTransactionRepository is a mock implementation of AuthorizationTransactionRepository
type MockAuthorizationTransactionRepository struct {
	mock.Mock
}

// Create stores a new authorization transaction with the given TTL
func (m *MockAuthorizationTransactionRepository) Create(ctx context.Context, txn *repositories.AuthorizationTransaction, ttl time.Duration) error {
	args := m.Called(ctx, txn, ttl)
	return args.Error(0)
}

// Get retrieves a transaction by ID. Returns nil, nil if not found.
func (m *MockAuthorizationTransactionRepository) Get(ctx context.Context, transactionID string) (*repositories.AuthorizationTransaction, error) {
	args := m.Called(ctx, transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.AuthorizationTransaction), args.Error(1)
}

// UpdateStage updates the transaction stage and may bind a user or session
func (m *MockAuthorizationTransactionRepository) UpdateStage(ctx context.Context, transactionID string, stage repositories.AuthorizationTransactionStage, userID string, sessionID string) error {
	args := m.Called(ctx, transactionID, stage, userID, sessionID)
	return args.Error(0)
}

// Complete atomically marks a pending_consent transaction as completed and returns the stored transaction data
func (m *MockAuthorizationTransactionRepository) Complete(ctx context.Context, transactionID string) (*repositories.AuthorizationTransaction, error) {
	args := m.Called(ctx, transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.AuthorizationTransaction), args.Error(1)
}