package entities

import (
	"errors"
	"strings"
	"time"
)

// MaxRoleNameLength is the maximum allowed length for a free-form role name
const MaxRoleNameLength = 100

// UserRoleAssignment represents a role assignment for a user to access a client
type UserRoleAssignment struct {
	ID        int64
	UserID    string
	ClientID  string
	TenantID  string
	Role      string
	IsActive  bool
	GrantedAt time.Time
	GrantedBy string
	RevokedAt *time.Time
	RevokedBy *string
}

// Validate performs basic validation on the user role entity.
// Role names are free-form (1–100 characters); whitespace-only names are rejected.
func (ur *UserRoleAssignment) Validate() error {
	if ur.UserID == "" {
		return errors.New("user_id cannot be empty")
	}
	if ur.ClientID == "" {
		return errors.New("client_id cannot be empty")
	}
	if ur.TenantID == "" {
		return errors.New("tenant_id cannot be empty")
	}
	if strings.TrimSpace(ur.Role) == "" {
		return errors.New("role cannot be empty")
	}
	if len(ur.Role) > MaxRoleNameLength {
		return errors.New("role name must not exceed 100 characters")
	}

	return nil
}

// IsEnabled checks if the role assignment is active
func (ur *UserRoleAssignment) IsEnabled() bool {
	return ur.IsActive
}
