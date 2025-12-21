package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTenant(t *testing.T) {
	tests := []struct {
		name             string
		organizationName string
		wantErr          bool
		errContains      string
	}{
		{
			name:             "Valid organization name",
			organizationName: "Acme Corp",
			wantErr:          false,
		},
		{
			name:             "Valid organization name with numbers",
			organizationName: "Tech123 Inc",
			wantErr:          false,
		},
		{
			name:             "Valid organization name minimum length",
			organizationName: "ABC",
			wantErr:          false,
		},
		{
			name:             "Valid organization name maximum length",
			organizationName: "A very long organization name that is exactly one hundred characters long for testing purposes ok",
			wantErr:          false,
		},
		{
			name:             "Empty organization name",
			organizationName: "",
			wantErr:          true,
			errContains:      "organization name must be between 3 and 100 characters",
		},
		{
			name:             "Organization name too short",
			organizationName: "A",
			wantErr:          true,
			errContains:      "organization name must be between 3 and 100 characters",
		},
		{
			name:             "Organization name too long",
			organizationName: "A very long organization name that exceeds the maximum allowed length of one hundred characters limit",
			wantErr:          true,
			errContains:      "organization name must be between 3 and 100 characters",
		},
		{
			name:             "Organization name with special characters",
			organizationName: "Acme Corp <script>",
			wantErr:          true,
			errContains:      "organization name must be between 3 and 100 characters",
		},
		{
			name:             "Organization name with SQL injection attempt",
			organizationName: "Acme'; DROP TABLE tenants--",
			wantErr:          true,
			errContains:      "organization name must be between 3 and 100 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant, err := entities.NewTenant(tt.organizationName)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, tenant)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tenant)
				assert.Equal(t, tt.organizationName, tenant.OrganizationName)
				assert.Equal(t, entities.TenantStatusPendingVerification, tenant.Status)
				assert.NotEqual(t, uuid.Nil, tenant.ID)
				assert.False(t, tenant.CreatedAt.IsZero())
				assert.False(t, tenant.UpdatedAt.IsZero())
			}
		})
	}
}

func TestTenant_Activate(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *entities.Tenant
		wantErr     bool
		errContains string
	}{
		{
			name: "Activate pending tenant",
			setup: func() *entities.Tenant {
				tenant, _ := entities.NewTenant("Acme Corp")
				return tenant
			},
			wantErr: false,
		},
		{
			name: "Cannot activate already active tenant",
			setup: func() *entities.Tenant {
				tenant, _ := entities.NewTenant("Acme Corp")
				_ = tenant.Activate()
				return tenant
			},
			wantErr:     true,
			errContains: "must be in pending_verification status",
		},
		{
			name: "Cannot activate suspended tenant",
			setup: func() *entities.Tenant {
				tenant, _ := entities.NewTenant("Acme Corp")
				_ = tenant.Activate()
				_ = tenant.Suspend()
				return tenant
			},
			wantErr:     true,
			errContains: "must be in pending_verification status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant := tt.setup()
			err := tenant.Activate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, entities.TenantStatusActive, tenant.Status)
			}
		})
	}
}

func TestTenant_Suspend(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *entities.Tenant
		wantErr     bool
		errContains string
	}{
		{
			name: "Suspend active tenant",
			setup: func() *entities.Tenant {
				tenant, _ := entities.NewTenant("Acme Corp")
				_ = tenant.Activate()
				return tenant
			},
			wantErr: false,
		},
		{
			name: "Cannot suspend pending tenant",
			setup: func() *entities.Tenant {
				tenant, _ := entities.NewTenant("Acme Corp")
				return tenant
			},
			wantErr:     true,
			errContains: "only active tenants can be suspended",
		},
		{
			name: "Cannot suspend already suspended tenant",
			setup: func() *entities.Tenant {
				tenant, _ := entities.NewTenant("Acme Corp")
				_ = tenant.Activate()
				_ = tenant.Suspend()
				return tenant
			},
			wantErr:     true,
			errContains: "only active tenants can be suspended",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant := tt.setup()
			err := tenant.Suspend()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, entities.TenantStatusSuspended, tenant.Status)
			}
		})
	}
}

func TestTenant_IsActive(t *testing.T) {
	t.Run("Pending tenant is not active", func(t *testing.T) {
		tenant, _ := entities.NewTenant("Acme Corp")
		assert.False(t, tenant.IsActive())
	})

	t.Run("Active tenant is active", func(t *testing.T) {
		tenant, _ := entities.NewTenant("Acme Corp")
		_ = tenant.Activate()
		assert.True(t, tenant.IsActive())
	})

	t.Run("Suspended tenant is not active", func(t *testing.T) {
		tenant, _ := entities.NewTenant("Acme Corp")
		_ = tenant.Activate()
		_ = tenant.Suspend()
		assert.False(t, tenant.IsActive())
	})
}

func TestTenant_IsPendingVerification(t *testing.T) {
	t.Run("Pending tenant is pending", func(t *testing.T) {
		tenant, _ := entities.NewTenant("Acme Corp")
		assert.True(t, tenant.IsPendingVerification())
	})

	t.Run("Active tenant is not pending", func(t *testing.T) {
		tenant, _ := entities.NewTenant("Acme Corp")
		_ = tenant.Activate()
		assert.False(t, tenant.IsPendingVerification())
	})
}

func TestTenant_IsSuspended(t *testing.T) {
	t.Run("Suspended tenant is suspended", func(t *testing.T) {
		tenant, _ := entities.NewTenant("Acme Corp")
		_ = tenant.Activate()
		_ = tenant.Suspend()
		assert.True(t, tenant.IsSuspended())
	})

	t.Run("Active tenant is not suspended", func(t *testing.T) {
		tenant, _ := entities.NewTenant("Acme Corp")
		_ = tenant.Activate()
		assert.False(t, tenant.IsSuspended())
	})
}
