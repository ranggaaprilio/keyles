package domain_test

import (
"strings"
"testing"
"time"

"github.com/ranggaaprilio/keyles/domain/entities"
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	user := entities.NewUser("tenant-1", " Alice@Example.com ", "Alice Smith")
	assert.Equal(t, "tenant-1", user.TenantID)
	assert.Equal(t, "alice@example.com", user.Email)
	assert.Equal(t, "Alice Smith", user.DisplayName)
	assert.Equal(t, entities.UserStatusPending, user.Status)
	assert.Empty(t, user.PasswordHash)
	assert.Nil(t, user.LastLoginAt)
	assert.False(t, user.CreatedAt.IsZero())
	assert.False(t, user.UpdatedAt.IsZero())
}

func TestNewUser_SetsTimestamps(t *testing.T) {
	before := time.Now()
	user := entities.NewUser("t1", "a@b.com", "A")
	after := time.Now()
	assert.True(t, !user.CreatedAt.Before(before))
	assert.True(t, !user.CreatedAt.After(after))
}

func TestUserValidate_ValidUser(t *testing.T) {
	user := &entities.User{
		TenantID: "tenant-1",
		Email:    "test@example.com",
		Status:   entities.UserStatusActive,
	}
	err := user.Validate()
	assert.NoError(t, err)
}

func TestUserValidate_EmptyEmail(t *testing.T) {
	user := &entities.User{
		TenantID: "tenant-1",
		Email:    "",
		Status:   entities.UserStatusActive,
	}
	err := user.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email cannot be empty")
}

func TestUserValidate_InvalidEmail(t *testing.T) {
	user := &entities.User{
		TenantID: "tenant-1",
		Email:    "notanemail",
		Status:   entities.UserStatusActive,
	}
	err := user.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "valid email")
}

func TestUserValidate_EmptyTenantID(t *testing.T) {
	user := &entities.User{
		TenantID: "",
		Email:    "test@example.com",
		Status:   entities.UserStatusActive,
	}
	err := user.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tenant_id cannot be empty")
}

func TestUserValidate_InvalidStatus(t *testing.T) {
	user := &entities.User{
		TenantID: "tenant-1",
		Email:    "test@example.com",
		Status:   entities.UserStatus("unknown"),
	}
	err := user.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user status")
}

func TestUserValidate_DisplayNameTooLong(t *testing.T) {
	user := &entities.User{
		TenantID:    "tenant-1",
		Email:       "test@example.com",
		Status:      entities.UserStatusActive,
		DisplayName: strings.Repeat("a", 256),
	}
	err := user.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "255 characters")
}

func TestUserStatusConstants(t *testing.T) {
	assert.Equal(t, entities.UserStatus("pending"), entities.UserStatusPending)
	assert.Equal(t, entities.UserStatus("active"), entities.UserStatusActive)
	assert.Equal(t, entities.UserStatus("disabled"), entities.UserStatusDisabled)
}

func TestMaxUsersPerTenant(t *testing.T) {
	assert.Equal(t, 10_000, entities.MaxUsersPerTenant)
}

func TestUserValidate_AllStatuses(t *testing.T) {
	for _, status := range []entities.UserStatus{
		entities.UserStatusPending,
		entities.UserStatusActive,
		entities.UserStatusDisabled,
	} {
		user := &entities.User{
			TenantID: "tenant-1",
			Email:    "test@example.com",
			Status:   status,
		}
		assert.NoError(t, user.Validate(), "status %s should be valid", status)
	}
}
