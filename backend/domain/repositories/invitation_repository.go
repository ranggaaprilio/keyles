package repositories

import (
	"context"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
)

// InvitationRepository defines the interface for invitation persistence operations
type InvitationRepository interface {
	// Create inserts a new invitation record
	Create(ctx context.Context, inv *entities.Invitation) error

	// GetByToken fetches a pending invitation by validating the provided plain token
	// against stored hashes. Returns ErrNotFound if no match or token is expired/used.
	GetByToken(ctx context.Context, plainToken string) (*entities.Invitation, error)

	// GetPendingByEmail returns the active pending invitation for an email in a tenant, if any
	GetPendingByEmail(ctx context.Context, tenantID, email string) (*entities.Invitation, error)

	// UpdateStatus transitions an invitation's status (e.g., pending → accepted or expired)
	UpdateStatus(ctx context.Context, invitationID string, status entities.InvitationStatus, acceptedAt *time.Time) error

	// ListByTenant returns all invitations for a tenant (admin view)
	ListByTenant(ctx context.Context, tenantID string, page, pageSize int) ([]*entities.Invitation, int, error)

	// ExpireStalePending marks all pending invitations past their expires_at as expired.
	// Intended to be called by a scheduled background job.
	ExpireStalePending(ctx context.Context) (int64, error)
}
