package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// UserStatus represents the status of a user within a tenant
type UserStatus string

const (
	UserStatusPending  UserStatus = "pending"
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

// MaxUsersPerTenant is the maximum number of users allowed per tenant
const MaxUsersPerTenant = 10_000

// User represents an end-user invited by administrators within a tenant
type User struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Email        string
	DisplayName  *string
	PasswordHash string
	Status       UserStatus
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}

// NewUser creates a new end-user (password will be set during invitation acceptance)
func NewUser(tenantID uuid.UUID, email, passwordHash string) (*User, error) {
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}
	if tenantID == uuid.Nil {
		return nil, errors.New("tenant_id is required")
	}
	now := time.Now()
	return &User{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Email:        NormalizeEmail(email),
		PasswordHash: passwordHash,
		Status:       UserStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// Validate validates the user fields
func (u *User) Validate() error {
	if u.ID == uuid.Nil {
		return errors.New("id is required")
	}
	if u.TenantID == uuid.Nil {
		return errors.New("tenant_id is required")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	if err := ValidateEmail(u.Email); err != nil {
		return err
	}
	validStatuses := map[UserStatus]bool{
		UserStatusPending:  true,
		UserStatusActive:   true,
		UserStatusDisabled: true,
	}
	if !validStatuses[u.Status] {
		return errors.New("status must be one of: pending, active, disabled")
	}
	if u.DisplayName != nil && len(*u.DisplayName) > 255 {
		return errors.New("display_name must be at most 255 characters")
	}
	return nil
}
