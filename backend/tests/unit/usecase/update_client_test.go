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

func TestUpdateClient_Execute(t *testing.T) {
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
		name         string
		clientID     string
		tenantID     string
		clientName   *string
		redirectURIs []string
		isActive     *bool
		setupMocks   func(*mocks.MockClientRepository)
		wantErr      bool
		errContains  string
		validateResp func(*testing.T, *client.UpdateClientResponse)
	}{
		{
			name:         "Update client name",
			clientID:     "client-123",
			tenantID:     "tenant-456",
			clientName:   strPtr("Updated Application"),
			redirectURIs: nil,
			isActive:     nil,
			setupMocks: func(repo *mocks.MockClientRepository) {
				repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").
					Return(existingClient, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Client")).
					Return(nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *client.UpdateClientResponse) {
				assert.Equal(t, "Updated Application", resp.ClientName)
			},
		},
		{
			name:         "Update redirect URIs",
			clientID:     "client-123",
			tenantID:     "tenant-456",
			clientName:   nil,
			redirectURIs: []string{"https://new.example.com/callback", "http://localhost:3000/callback"},
			isActive:     nil,
			setupMocks: func(repo *mocks.MockClientRepository) {
				repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").
					Return(existingClient, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Client")).
					Return(nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *client.UpdateClientResponse) {
				assert.Equal(t, 2, len(resp.RedirectURIs))
				assert.Contains(t, resp.RedirectURIs, "https://new.example.com/callback")
			},
		},
		{
			name:         "Deactivate client",
			clientID:     "client-123",
			tenantID:     "tenant-456",
			clientName:   nil,
			redirectURIs: nil,
			isActive:     boolPtr(false),
			setupMocks: func(repo *mocks.MockClientRepository) {
				repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").
					Return(existingClient, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Client")).
					Return(nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *client.UpdateClientResponse) {
				assert.False(t, resp.IsActive)
			},
		},
		{
			name:         "Client not found",
			clientID:     "nonexistent",
			tenantID:     "tenant-456",
			clientName:   strPtr("New Name"),
			redirectURIs: nil,
			isActive:     nil,
			setupMocks: func(repo *mocks.MockClientRepository) {
				repo.On("GetByClientID", mock.Anything, "nonexistent", "tenant-456").
					Return(nil, errors.New("client not found"))
			},
			wantErr:     true,
			errContains: "client not found",
		},
		{
			name:         "Invalid redirect URI",
			clientID:     "client-123",
			tenantID:     "tenant-456",
			clientName:   nil,
			redirectURIs: []string{"invalid-uri-without-scheme"},
			isActive:     nil,
			setupMocks: func(repo *mocks.MockClientRepository) {
				repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").
					Return(existingClient, nil)
			},
			wantErr:     true,
			errContains: "must have a scheme",
		},
		{
			name:         "Empty client ID",
			clientID:     "",
			tenantID:     "tenant-456",
			clientName:   strPtr("New Name"),
			redirectURIs: nil,
			isActive:     nil,
			setupMocks:   func(repo *mocks.MockClientRepository) {},
			wantErr:      true,
			errContains:  "client_id is required",
		},
		{
			name:         "Empty tenant ID",
			clientID:     "client-123",
			tenantID:     "",
			clientName:   strPtr("New Name"),
			redirectURIs: nil,
			isActive:     nil,
			setupMocks:   func(repo *mocks.MockClientRepository) {},
			wantErr:      true,
			errContains:  "tenant_id is required",
		},
		{
			name:         "Repository update failure",
			clientID:     "client-123",
			tenantID:     "tenant-456",
			clientName:   strPtr("New Name"),
			redirectURIs: nil,
			isActive:     nil,
			setupMocks: func(repo *mocks.MockClientRepository) {
				repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").
					Return(existingClient, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Client")).
					Return(errors.New("database error"))
			},
			wantErr:     true,
			errContains: "failed to update client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.MockClientRepository)
			tt.setupMocks(repo)

			uc := client.NewUpdateClientUseCase(repo)

			req := &client.UpdateClientRequest{
				ClientID:     tt.clientID,
				TenantID:     tt.tenantID,
				ClientName:   tt.clientName,
				RedirectURIs: tt.redirectURIs,
				IsActive:     tt.isActive,
			}

			resp, err := uc.Execute(context.Background(), req)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				if tt.validateResp != nil {
					tt.validateResp(t, resp)
				}
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestUpdateClient_PreservesUnchangedFields(t *testing.T) {
	now := time.Now()
	existingClient := &entities.Client{
		ClientID:            "client-123",
		TenantID:            "tenant-456",
		ClientName:          "Original Name",
		ClientSecretHash:    "original-secret-hash",
		AllowedRedirectURIs: []string{"https://original.example.com/callback"},
		IsActive:            true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	repo := new(mocks.MockClientRepository)

	var updatedClient *entities.Client
	repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").
		Return(existingClient, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Client")).
		Run(func(args mock.Arguments) {
			updatedClient = args.Get(1).(*entities.Client)
		}).
		Return(nil)

	uc := client.NewUpdateClientUseCase(repo)

	// Only update the name
	req := &client.UpdateClientRequest{
		ClientID:   "client-123",
		TenantID:   "tenant-456",
		ClientName: strPtr("Updated Name"),
	}

	_, err := uc.Execute(context.Background(), req)
	require.NoError(t, err)

	// Verify unchanged fields are preserved
	assert.Equal(t, "Updated Name", updatedClient.ClientName)
	assert.Equal(t, "original-secret-hash", updatedClient.ClientSecretHash)
	assert.Equal(t, existingClient.AllowedRedirectURIs, updatedClient.AllowedRedirectURIs)
	assert.Equal(t, existingClient.IsActive, updatedClient.IsActive)
	assert.Equal(t, existingClient.CreatedAt, updatedClient.CreatedAt)
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
