package entities

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role of a user within a tenant
type UserRole string

const (
	UserRoleAdmin  UserRole = "admin"
	UserRoleOwner  UserRole = "owner"
	UserRoleMember UserRole = "member"
	UserRoleViewer UserRole = "viewer"
)

// AdminUser represents the primary administrator for a tenant
type AdminUser struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	FullName     string
	Email        string
	PasswordHash string
	Role         UserRole
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName specifies the table name for GORM
func (AdminUser) TableName() string {
	return "users"
}

var (
	// Email validation (RFC 5322 simplified)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// Password complexity validation
	passwordMinLength       = 8
	passwordUppercaseRegex  = regexp.MustCompile(`[A-Z]`)
	passwordLowercaseRegex  = regexp.MustCompile(`[a-z]`)
	passwordNumberRegex     = regexp.MustCompile(`[0-9]`)
	passwordSpecialRegex    = regexp.MustCompile(`[@$!%*?&]`)

	// Validation errors
	ErrInvalidEmail           = errors.New("email must be a valid email address")
	ErrInvalidFullName        = errors.New("full name must be between 2 and 100 characters")
	ErrPasswordTooShort       = errors.New("password must be at least 8 characters long")
	ErrPasswordMissingUpper   = errors.New("password must contain at least one uppercase letter")
	ErrPasswordMissingLower   = errors.New("password must contain at least one lowercase letter")
	ErrPasswordMissingNumber  = errors.New("password must contain at least one number")
	ErrPasswordMissingSpecial = errors.New("password must contain at least one special character (@$!%*?&)")
)

// NewAdminUser creates a new admin user (password will be hashed by infrastructure layer)
func NewAdminUser(tenantID uuid.UUID, fullName, email, passwordHash string) (*AdminUser, error) {
	if err := ValidateFullName(fullName); err != nil {
		return nil, err
	}

	if err := ValidateEmail(email); err != nil {
		return nil, err
	}

	now := time.Now()
	return &AdminUser{
		ID:           uuid.New(),
		TenantID:     tenantID,
		FullName:     strings.TrimSpace(fullName),
		Email:        NormalizeEmail(email),
		PasswordHash: passwordHash,
		Role:         UserRoleAdmin,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	trimmed := strings.TrimSpace(email)
	if !emailRegex.MatchString(trimmed) {
		return ErrInvalidEmail
	}
	return nil
}

// NormalizeEmail normalizes email (lowercase, trimmed)
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ValidateFullName validates full name length
func ValidateFullName(name string) error {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) < 2 || len(trimmed) > 100 {
		return ErrInvalidFullName
	}
	return nil
}

// ValidatePassword validates password complexity requirements
// NOTE: This validates the plain text password before hashing
func ValidatePassword(password string) error {
	if len(password) < passwordMinLength {
		return ErrPasswordTooShort
	}

	if !passwordUppercaseRegex.MatchString(password) {
		return ErrPasswordMissingUpper
	}

	if !passwordLowercaseRegex.MatchString(password) {
		return ErrPasswordMissingLower
	}

	if !passwordNumberRegex.MatchString(password) {
		return ErrPasswordMissingNumber
	}

	if !passwordSpecialRegex.MatchString(password) {
		return ErrPasswordMissingSpecial
	}

	return nil
}
