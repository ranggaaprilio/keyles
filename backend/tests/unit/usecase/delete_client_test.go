package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteClient_Execute(t *testing.T) {
	now := time.Now()
	existingClient := &entities.Client{
		ClientID:            "client-123",
		TenantID:            "tenant-456",
		ClientName:          "Test Application",
		ClientSecretHash:    "hashed-secret",
		AllowedRedirectURIs: []string{"https://app.example.com/callback"},
		IsActive:            true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	tests := []struct {
		name        string
		clientID    string
		tenantID    string
		setupMocks  func(*mocks.MockClientRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:     "Successful deletion",
			clientID: "client-123",
			tenantID: "tenant-456",
			setupMocks: func(repo *mocks.MockClientRepository) {
				repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").
					Return(existingClient, nil)
				repo.On("Delete", mock.Anything, "client-123").
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "Client not found",
			clientID: "nonexistent",
			tenantID: "tenant-456",
			setupMocks: func(repo *mocks.MockClientRepository) {
				repo.On("GetByClientID", mock.Anything, "nonexistent", "tenant-456").
					Return(nil, errors.New("client not found"))
			},
			wantErr:     true,
			errContains: "client not found",
		},
		{
			name:     "Wrong tenant - access denied",
			clientID: "client-123",
			tenantID: "wrong-tenant",
			setupMocks: func(repo *mocks.MockClientRepository) {
				repo.On("GetByClientID", mock.Anything, "client-123", "wrong-tenant").
					Return(nil, errors.New("client not found"))
			},
			wantErr:     true,
			errContains: "client not found",
		},
		{
			name:        "Empty client ID",
			clientID:    "",
			tenantID:    "tenant-456",
			setupMocks:  func(repo *mocks.MockClientRepository) {},
			wantErr:     true,
			errContains: "client_id is required",
		},
		{
			name:        "Empty tenant ID",
			clientID:    "client-123",
			tenantID:    "",
			setupMocks:  func(repo *mocks.MockClientRepository) {},
			wantErr:     true,
			errContains: "tenant_id is required",
		},
		{
			name:     "Repository delete failure",
			clientID: "client-123",
			tenantID: "tenant-456",
			setupMocks: func(repo *mocks.MockClientRepository) {
				repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").
					Return(existingClient, nil)
				repo.On("Delete", mock.Anything, "client-123").
					Return(errors.New("database error"))
			},
			wantErr:     true,
			errContains: "failed to delete client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.MockClientRepository)
			tt.setupMocks(repo)

			uc := client.NewDeleteClientUseCase(repo)

			req := &client.DeleteClientRequest{
				ClientID: tt.clientID,
				TenantID: tt.tenantID,
			}

			err := uc.Execute(context.Background(), req)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestDeleteClient_VerifiesTenantOwnership(t *testing.T) {
	now := time.Now()
	tenantAClient := &entities.Client{
		ClientID:            "client-123",
		TenantID:            "tenant-A",
		ClientName:          "Tenant A's Application",
		ClientSecretHash:    "hashed-secret",
		AllowedRedirectURIs: []string{"https://app.example.com/callback"},
		IsActive:            true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	repo := new(mocks.MockClientRepository)

	// Tenant B tries to delete Tenant A's client
	repo.On("GetByClientID", mock.Anything, "client-123", "tenant-B").
		Return(nil, errors.New("client not found")) // Repository should return not found for wrong tenant

	uc := client.NewDeleteClientUseCase(repo)

	req := &client.DeleteClientRequest{
		ClientID: "client-123",
		TenantID: "tenant-B", // Wrong tenant
	}

	err := uc.Execute(context.Background(), req)

	// Should fail because tenant B doesn't own this client
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Delete should NOT be called because verification failed
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)

	// Cleanup - check that the client still "exists" for Tenant A
	_ = tenantAClient // Unused, just demonstrating the concept
}
