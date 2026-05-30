package entities

import (
	"errors"
	"time"
)

const MaxRoleNameLength = 100

type UserRoleAssignment struct {
	ID         int64
	UserID     string
	ClientID   string
	TenantID   string
	Role       string
	IsActive   bool
	GrantedAt  time.Time
	GrantedBy  string
	RevokedAt  *time.Time
	RevokedBy  *string
}

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
	if len(ur.Role) > MaxRoleNameLength {
		return errors.New("role must be at most 100 characters")
	}
	return nil
}

func (ur *UserRoleAssignment) IsEnabled() bool {
	return ur.IsActive
}
