package entities

import (
	"time"

	"github.com/google/uuid"
)

// InvitationStatus represents the status of an invitation
type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusExpired  InvitationStatus = "expired"
)

// InvitationTTL is the time-to-live for an invitation
const InvitationTTL = 72 * time.Hour

// Invitation represents an invitation sent to a user to join a tenant
type Invitation struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Email       string
	DisplayName *string
	TokenHash   string
	Status      InvitationStatus
	InvitedBy   uuid.UUID
	ExpiresAt   time.Time
	AcceptedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName specifies the table name for GORM
func (Invitation) TableName() string {
	return "invitations"
}

// IsExpired checks if the invitation has expired
func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// IsAccepted checks if the invitation has been accepted
func (i *Invitation) IsAccepted() bool {
	return i.Status == InvitationStatusAccepted
}
