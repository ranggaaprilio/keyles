package entities

import (
	"errors"
	"time"
)

// UserRoleAssignment represents a role assignment for a user to access a client
type UserRoleAssignment struct {
	ID         int64
	UserID     string
	ClientID   string
	TenantID   string
	Role       string
	IsActive   bool
	GrantedAt  time.Time
	GrantedBy  string
}

// ValidRoles defines the allowed role values
var ValidRoles = []string{"admin", "user", "viewer"}

// Validate performs basic validation on the user role entity
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
	if ur.Role == "" {
		return errors.New("role cannot be empty")
	}

	// Validate role is in the allowed list
	if !ur.IsValidRole() {
		return errors.New("invalid role: must be one of admin, user, viewer")
	}

	return nil
}

// IsValidRole checks if the role is in the valid roles list
func (ur *UserRoleAssignment) IsValidRole() bool {
	for _, validRole := range ValidRoles {
		if ur.Role == validRole {
			return true
		}
	}
	return false
}

// IsEnabled checks if the role assignment is active
func (ur *UserRoleAssignment) IsEnabled() bool {
	return ur.IsActive
}
