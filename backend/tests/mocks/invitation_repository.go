package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/mock"
)

// MockInvitationRepository is a mock implementation of InvitationRepository
type MockInvitationRepository struct {
	mock.Mock
}

// Create stores a new invitation
func (m *MockInvitationRepository) Create(ctx context.Context, inv *entities.Invitation) error {
	args := m.Called(ctx, inv)
	return args.Error(0)
}

// GetByToken retrieves an invitation by its plain token
func (m *MockInvitationRepository) GetByToken(ctx context.Context, plainToken string) (*entities.Invitation, error) {
	args := m.Called(ctx, plainToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Invitation), args.Error(1)
}

// GetPendingByEmail retrieves a pending invitation by email within a tenant
func (m *MockInvitationRepository) GetPendingByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entities.Invitation, error) {
	args := m.Called(ctx, tenantID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Invitation), args.Error(1)
}

// UpdateStatus updates the status of an invitation
func (m *MockInvitationRepository) UpdateStatus(ctx context.Context, invitationID uuid.UUID, status entities.InvitationStatus, acceptedAt *time.Time) error {
	args := m.Called(ctx, invitationID, status, acceptedAt)
	return args.Error(0)
}

// ListByTenant retrieves paginated invitations for a tenant
func (m *MockInvitationRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*entities.Invitation, int, error) {
	args := m.Called(ctx, tenantID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*entities.Invitation), args.Int(1), args.Error(2)
}

// ExpireStalePending expires invitations that have been pending too long
func (m *MockInvitationRepository) ExpireStalePending(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
