package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
)

// InvitationRepository defines the interface for invitation persistence operations
type InvitationRepository interface {
	// Create stores a new invitation
	Create(ctx context.Context, inv *entities.Invitation) error

	// GetByToken retrieves an invitation by its plain token
	GetByToken(ctx context.Context, plainToken string) (*entities.Invitation, error)

	// GetPendingByEmail retrieves a pending invitation by email within a tenant
	GetPendingByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entities.Invitation, error)

	// UpdateStatus updates the status of an invitation
	UpdateStatus(ctx context.Context, invitationID uuid.UUID, status entities.InvitationStatus, acceptedAt *time.Time) error

	// ListByTenant retrieves paginated invitations for a tenant
	ListByTenant(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*entities.Invitation, int, error)

	// ExpireStalePending expires invitations that have been pending too long
	ExpireStalePending(ctx context.Context) (int64, error)
}
