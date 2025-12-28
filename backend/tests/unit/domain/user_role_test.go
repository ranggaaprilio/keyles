package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ranggaaprilio/keyles/domain/entities"
)

// TestUserRoleAssignment_Validate tests the validation of user role assignments
func TestUserRoleAssignment_Validate(t *testing.T) {
	tests := []struct {
		name    string
		role    *entities.UserRoleAssignment
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid role assignment",
			role: &entities.UserRoleAssignment{
				ID:        1,
				UserID:    "user-123",
				ClientID:  "client-123",
				TenantID:  "tenant-123",
				Role:      "user",
				IsActive:  true,
				GrantedAt: time.Now(),
				GrantedBy: "admin-123",
			},
			wantErr: false,
		},
		{
			name: "Valid admin role",
			role: &entities.UserRoleAssignment{
				ID:        2,
				UserID:    "user-456",
				ClientID:  "client-123",
				TenantID:  "tenant-123",
				Role:      "admin",
				IsActive:  true,
				GrantedAt: time.Now(),
				GrantedBy: "admin-123",
			},
			wantErr: false,
		},
		{
			name: "Valid viewer role",
			role: &entities.UserRoleAssignment{
				ID:        3,
				UserID:    "user-789",
				ClientID:  "client-123",
				TenantID:  "tenant-123",
				Role:      "viewer",
				IsActive:  true,
				GrantedAt: time.Now(),
				GrantedBy: "admin-123",
			},
			wantErr: false,
		},
		{
			name: "Missing user_id",
			role: &entities.UserRoleAssignment{
				ID:        4,
				UserID:    "",
				ClientID:  "client-123",
				TenantID:  "tenant-123",
				Role:      "user",
				IsActive:  true,
				GrantedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "user_id cannot be empty",
		},
		{
			name: "Missing client_id",
			role: &entities.UserRoleAssignment{
				ID:        5,
				UserID:    "user-123",
				ClientID:  "",
				TenantID:  "tenant-123",
				Role:      "user",
				IsActive:  true,
				GrantedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "client_id cannot be empty",
		},
		{
			name: "Missing tenant_id",
			role: &entities.UserRoleAssignment{
				ID:        6,
				UserID:    "user-123",
				ClientID:  "client-123",
				TenantID:  "",
				Role:      "user",
				IsActive:  true,
				GrantedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "tenant_id cannot be empty",
		},
		{
			name: "Missing role",
			role: &entities.UserRoleAssignment{
				ID:        7,
				UserID:    "user-123",
				ClientID:  "client-123",
				TenantID:  "tenant-123",
				Role:      "",
				IsActive:  true,
				GrantedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "role cannot be empty",
		},
		{
			name: "Invalid role value",
			role: &entities.UserRoleAssignment{
				ID:        8,
				UserID:    "user-123",
				ClientID:  "client-123",
				TenantID:  "tenant-123",
				Role:      "superuser",
				IsActive:  true,
				GrantedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "invalid role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.role.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUserRoleAssignment_IsValidRole tests role validation
func TestUserRoleAssignment_IsValidRole(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"admin is valid", "admin", true},
		{"user is valid", "user", true},
		{"viewer is valid", "viewer", true},
		{"superuser is invalid", "superuser", false},
		{"empty is invalid", "", false},
		{"random is invalid", "random", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assignment := &entities.UserRoleAssignment{Role: tt.role}
			assert.Equal(t, tt.expected, assignment.IsValidRole())
		})
	}
}

// TestUserRoleAssignment_IsEnabled tests active status checking
func TestUserRoleAssignment_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		isActive bool
		expected bool
	}{
		{"Active role is enabled", true, true},
		{"Inactive role is not enabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assignment := &entities.UserRoleAssignment{IsActive: tt.isActive}
			assert.Equal(t, tt.expected, assignment.IsEnabled())
		})
	}
}
