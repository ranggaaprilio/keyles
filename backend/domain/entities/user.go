package entities

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// UserStatus represents the lifecycle state of a user account
type UserStatus string

const (
	// UserStatusPending indicates an invited user who hasn't set their password yet
	UserStatusPending UserStatus = "pending"
	// UserStatusActive indicates a fully active user
	UserStatusActive UserStatus = "active"
	// UserStatusDisabled indicates an administratively disabled user
	UserStatusDisabled UserStatus = "disabled"
)

// MaxUsersPerTenant defines the tenant user quota
const MaxUsersPerTenant = 10_000

// User represents an end-user who can authenticate via the tenant's SSO.
// This is distinct from AdminUser which represents tenant administrators.
// Both entities map to the same "users" PostgreSQL table.
type User struct {
	ID           string
	TenantID     string
	Email        string
	DisplayName  string
	PasswordHash string
	Status       UserStatus
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// userEmailRegex is a simplified RFC 5322 email validation pattern
var userEmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// NewUser creates a new end-user entity with defaults
func NewUser(tenantID, email, displayName string) *User {
	now := time.Now()
	return &User{
		TenantID:    tenantID,
		Email:       strings.ToLower(strings.TrimSpace(email)),
		DisplayName: strings.TrimSpace(displayName),
		Status:      UserStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Validate performs basic validation on the user entity
func (u *User) Validate() error {
	if u.TenantID == "" {
		return errors.New("tenant_id cannot be empty")
	}
	if u.Email == "" {
		return errors.New("email cannot be empty")
	}
	trimmedEmail := strings.TrimSpace(u.Email)
	if !userEmailRegex.MatchString(trimmedEmail) {
		return errors.New("email must be a valid email address")
	}
	if !isValidUserStatus(u.Status) {
		return errors.New("invalid user status: must be pending, active, or disabled")
	}
	if len(u.DisplayName) > 255 {
		return errors.New("display name must not exceed 255 characters")
	}
	return nil
}

// isValidUserStatus checks if the status is one of the allowed values
func isValidUserStatus(s UserStatus) bool {
	switch s {
	case UserStatusPending, UserStatusActive, UserStatusDisabled:
		return true
	default:
		return false
	}
}
