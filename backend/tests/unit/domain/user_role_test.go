package domain_test

import (
	"strings"
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
			name: "Custom role value is now valid",
			role: &entities.UserRoleAssignment{
				ID:        8,
				UserID:    "user-123",
				ClientID:  "client-123",
				TenantID:  "tenant-123",
				Role:      "superuser",
				IsActive:  true,
				GrantedAt: time.Now(),
			},
			wantErr: false,
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

// TestUserRoleAssignment_RoleValidation tests role validation via Validate()
func TestUserRoleAssignment_RoleValidation(t *testing.T) {
	tests := []struct {
		name      string
		role      string
		expectErr bool
	}{
		{"admin is valid", "admin", false},
		{"custom role is valid", "billing_manager", false},
		{"empty role is invalid", "", true},
		{"whitespace-only role is invalid", "   ", true},
		{"role exceeding max length is invalid", strings.Repeat("a", entities.MaxRoleNameLength+1), true},
		{"role at max length is valid", strings.Repeat("a", entities.MaxRoleNameLength), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assignment := &entities.UserRoleAssignment{
				UserID:   "user-1",
				ClientID: "client-1",
				TenantID: "tenant-1",
				Role:     tt.role,
			}
			err := assignment.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
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
