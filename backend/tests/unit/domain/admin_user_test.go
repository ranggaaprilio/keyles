package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAdminUser(t *testing.T) {
	tenantID := uuid.New()

	tests := []struct {
		name         string
		tenantID     uuid.UUID
		email        string
		passwordHash string
		fullName     string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "Valid admin user",
			tenantID:     tenantID,
			email:        "admin@acme.com",
			passwordHash: "hashed_password",
			fullName:     "John Doe",
			wantErr:      false,
		},
		{
			name:         "Valid email with subdomain",
			tenantID:     tenantID,
			email:        "admin@mail.acme.com",
			passwordHash: "hashed_password",
			fullName:     "John Doe",
			wantErr:      false,
		},
		{
			name:         "Valid email with plus addressing",
			tenantID:     tenantID,
			email:        "admin+test@acme.com",
			passwordHash: "hashed_password",
			fullName:     "John Doe",
			wantErr:      false,
		},
		{
			name:         "Empty email",
			tenantID:     tenantID,
			email:        "",
			passwordHash: "hashed_password",
			fullName:     "John Doe",
			wantErr:      true,
			errContains:  "email must be a valid email address",
		},
		{
			name:         "Invalid email format - no @",
			tenantID:     tenantID,
			email:        "adminacme.com",
			passwordHash: "hashed_password",
			fullName:     "John Doe",
			wantErr:      true,
			errContains:  "email must be a valid email address",
		},
		{
			name:         "Invalid email format - no domain",
			tenantID:     tenantID,
			email:        "admin@",
			passwordHash: "hashed_password",
			fullName:     "John Doe",
			wantErr:      true,
			errContains:  "email must be a valid email address",
		},
		{
			name:         "Invalid email format - no local part",
			tenantID:     tenantID,
			email:        "@acme.com",
			passwordHash: "hashed_password",
			fullName:     "John Doe",
			wantErr:      true,
			errContains:  "email must be a valid email address",
		},
		{
			name:         "Empty password hash should not fail (hash validation happens elsewhere)",
			tenantID:     tenantID,
			email:        "admin@acme.com",
			passwordHash: "",
			fullName:     "John Doe",
			wantErr:      false,
		},
		{
			name:         "Empty full name",
			tenantID:     tenantID,
			email:        "admin@acme.com",
			passwordHash: "hashed_password",
			fullName:     "",
			wantErr:      true,
			errContains:  "full name must be between 2 and 100 characters",
		},
		{
			name:         "Full name too short",
			tenantID:     tenantID,
			email:        "admin@acme.com",
			passwordHash: "hashed_password",
			fullName:     "A",
			wantErr:      true,
			errContains:  "full name must be between 2 and 100 characters",
		},
		{
			name:         "Full name too long",
			tenantID:     tenantID,
			email:        "admin@acme.com",
			passwordHash: "hashed_password",
			fullName:     "A very long full name that exceeds one hundred characters which is the maximum allowed length for this field",
			wantErr:      true,
			errContains:  "full name must be between 2 and 100 characters",
		},
		{
			name:         "Nil tenant ID should not fail (ID validation happens elsewhere)",
			tenantID:     uuid.Nil,
			email:        "admin@acme.com",
			passwordHash: "hashed_password",
			fullName:     "John Doe",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := entities.NewAdminUser(tt.tenantID, tt.fullName, tt.email, tt.passwordHash)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.tenantID, user.TenantID)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.passwordHash, user.PasswordHash)
				assert.Equal(t, tt.fullName, user.FullName)
				assert.Equal(t, entities.UserRoleAdmin, user.Role)
				assert.NotEqual(t, uuid.Nil, user.ID)
				assert.False(t, user.CreatedAt.IsZero())
				assert.False(t, user.UpdatedAt.IsZero())
			}
		})
	}
}

func TestAdminUser_ValidateEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "Valid email",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "Valid email with subdomain",
			email:   "test@mail.example.com",
			wantErr: false,
		},
		{
			name:    "Valid email with plus",
			email:   "test+tag@example.com",
			wantErr: false,
		},
		{
			name:    "Valid email with numbers",
			email:   "test123@example456.com",
			wantErr: false,
		},
		{
			name:        "Empty email",
			email:       "",
			wantErr:     true,
			errContains: "email must be a valid email address",
		},
		{
			name:        "Email without @",
			email:       "testexample.com",
			wantErr:     true,
			errContains: "email must be a valid email address",
		},
		{
			name:        "Email without domain",
			email:       "test@",
			wantErr:     true,
			errContains: "email must be a valid email address",
		},
		{
			name:        "Email without local part",
			email:       "@example.com",
			wantErr:     true,
			errContains: "email must be a valid email address",
		},
		{
			name:        "Email with spaces",
			email:       "test @example.com",
			wantErr:     true,
			errContains: "email must be a valid email address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := entities.ValidateEmail(tt.email)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAdminUser_ValidatePassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "Valid password",
			password: "Password123!",
			wantErr:  false,
		},
		{
			name:     "Valid password with special characters",
			password: "P@ssw0rd!#$",
			wantErr:  false,
		},
		{
			name:        "Password too short",
			password:    "Pass1!",
			wantErr:     true,
			errContains: "password must be at least 8 characters",
		},
		{
			name:        "Password without uppercase",
			password:    "password123!",
			wantErr:     true,
			errContains: "password must contain at least one uppercase letter",
		},
		{
			name:        "Password without lowercase",
			password:    "PASSWORD123!",
			wantErr:     true,
			errContains: "password must contain at least one lowercase letter",
		},
		{
			name:        "Password without number",
			password:    "Password!",
			wantErr:     true,
			errContains: "password must contain at least one number",
		},
		{
			name:        "Password without special character",
			password:    "Password123",
			wantErr:     true,
			errContains: "password must contain at least one special character",
		},
		{
			name:        "Empty password",
			password:    "",
			wantErr:     true,
			errContains: "password must be at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := entities.ValidatePassword(tt.password)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAdminUser_ValidateFullName(t *testing.T) {
	tests := []struct {
		name        string
		fullName    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "Valid full name",
			fullName: "John Doe",
			wantErr:  false,
		},
		{
			name:     "Valid full name minimum length",
			fullName: "AB",
			wantErr:  false,
		},
		{
			name:     "Valid full name maximum length",
			fullName: "A very long full name that is exactly one hundred characters long for testing purposes okay now",
			wantErr:  false,
		},
		{
			name:        "Empty full name",
			fullName:    "",
			wantErr:     true,
			errContains: "full name must be between 2 and 100 characters",
		},
		{
			name:        "Full name too short",
			fullName:    "A",
			wantErr:     true,
			errContains: "full name must be between 2 and 100 characters",
		},
		{
			name:        "Full name too long",
			fullName:    "A very long full name that exceeds one hundred characters which is the maximum allowed length for this",
			wantErr:     true,
			errContains: "full name must be between 2 and 100 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := entities.ValidateFullName(tt.fullName)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
