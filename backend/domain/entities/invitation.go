package entities

import (
	"time"
)

// InvitationStatus represents the lifecycle state of an invitation
type InvitationStatus string

const (
	// InvitationStatusPending indicates an invitation that hasn't been accepted yet
	InvitationStatusPending InvitationStatus = "pending"
	// InvitationStatusAccepted indicates an invitation that has been successfully accepted
	InvitationStatusAccepted InvitationStatus = "accepted"
	// InvitationStatusExpired indicates an invitation that has passed its expiry time
	InvitationStatusExpired InvitationStatus = "expired"
)

// InvitationTTL is the duration an invitation token remains valid after issuance
const InvitationTTL = 72 * time.Hour

// Invitation represents a pending user activation request sent by a tenant administrator
type Invitation struct {
	ID          string
	TenantID    string
	Email       string
	DisplayName string
	TokenHash   string           // bcrypt hash of the one-time token
	Status      InvitationStatus // pending | accepted | expired
	InvitedBy   string           // user ID of the sending administrator
	ExpiresAt   time.Time
	AcceptedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsExpired returns true if the invitation has passed its expiry time
func (inv *Invitation) IsExpired() bool {
	return time.Now().After(inv.ExpiresAt)
}

// IsAccepted returns true if the invitation has been accepted
func (inv *Invitation) IsAccepted() bool {
	return inv.Status == InvitationStatusAccepted
}
