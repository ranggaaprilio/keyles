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
		Description:         "Original desc",
		ClientType:          "confidential",
		ClientSecretHash:    "hashed-secret",
		AllowedRedirectURIs: []string{"https://app.example.com/callback"},
		IsActive:            true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	tests := []struct {
		name         string
		request      *client.UpdateClientRequest
		setupMocks   func(*mocks.MockClientRepository, *mocks.MockAuditRepository)
		wantErr      bool
		errContains  string
		validateResp func(*testing.T, *client.UpdateClientResponse)
	}{
		{
			name: "Update client name",
			request: &client.UpdateClientRequest{
				ClientID:   "client-123",
				TenantID:   "tenant-456",
				ClientName: strPtr("Updated Application"),
				IPAddress:  "127.0.0.1",
				UserAgent:  "test",
			},
			setupMocks: func(repo *mocks.MockClientRepository, audit *mocks.MockAuditRepository) {
				repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").
					Return(existingClient, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Client")).
					Return(nil)
				audit.On("Create", mock.Anything, mock.AnythingOfType("*entities.AuditLog")).Return(nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *client.UpdateClientResponse) {
				assert.Equal(t, "Updated Application", resp.ClientName)
			},
		},
		{
			name: "Update redirect URIs",
			request: &client.UpdateClientRequest{
				ClientID:     "client-123",
				TenantID:     "tenant-456",
				RedirectURIs: []string{"https://new.example.com/callback", "http://localhost:3000/callback"},
				IPAddress:    "127.0.0.1",
				UserAgent:    "test",
			},
			setupMocks: func(repo *mocks.MockClientRepository, audit *mocks.MockAuditRepository) {
				repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").
					Return(existingClient, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Client")).
					Return(nil)
				audit.On("Create", mock.Anything, mock.AnythingOfType("*entities.AuditLog")).Return(nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *client.UpdateClientResponse) {
				assert.Equal(t, 2, len(resp.RedirectURIs))
				assert.Contains(t, resp.RedirectURIs, "https://new.example.com/callback")
			},
		},
		{
			name: "Deactivate client",
			request: &client.UpdateClientRequest{
				ClientID:  "client-123",
				TenantID:  "tenant-456",
				IsActive:  boolPtr(false),
				IPAddress: "127.0.0.1",
				UserAgent: "test",
			},
			setupMocks: func(repo *mocks.MockClientRepository, audit *mocks.MockAuditRepository) {
				repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").
					Return(existingClient, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Client")).
					Return(nil)
				audit.On("Create", mock.Anything, mock.AnythingOfType("*entities.AuditLog")).Return(nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *client.UpdateClientResponse) {
				assert.False(t, resp.IsActive)
			},
		},
		{
			name: "Update description",
			request: &client.UpdateClientRequest{
				ClientID:    "client-123",
				TenantID:    "tenant-456",
				Description: strPtr("New description"),
				IPAddress:   "127.0.0.1",
				UserAgent:   "test",
			},
			setupMocks: func(repo *mocks.MockClientRepository, audit *mocks.MockAuditRepository) {
				repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").
					Return(existingClient, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Client")).
					Return(nil)
				audit.On("Create", mock.Anything, mock.AnythingOfType("*entities.AuditLog")).Return(nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *client.UpdateClientResponse) {
				assert.Equal(t, "New description", resp.Description)
			},
		},
		{
			name: "Client not found",
			request: &client.UpdateClientRequest{
				ClientID:   "nonexistent",
				TenantID:   "tenant-456",
				ClientName: strPtr("New Name"),
			},
			setupMocks: func(repo *mocks.MockClientRepository, audit *mocks.MockAuditRepository) {
				repo.On("GetByClientID", mock.Anything, "nonexistent", "tenant-456").
					Return(nil, errors.New("client not found"))
			},
			wantErr:     true,
			errContains: "client not found",
		},
		{
			name: "Invalid redirect URI",
			request: &client.UpdateClientRequest{
				ClientID:     "client-123",
				TenantID:     "tenant-456",
				RedirectURIs: []string{"invalid-uri-without-scheme"},
			},
			setupMocks: func(repo *mocks.MockClientRepository, audit *mocks.MockAuditRepository) {
				repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").
					Return(existingClient, nil)
			},
			wantErr:     true,
			errContains: "scheme",
		},
		{
			name: "Empty client ID",
			request: &client.UpdateClientRequest{
				ClientID:   "",
				TenantID:   "tenant-456",
				ClientName: strPtr("New Name"),
			},
			setupMocks:  func(repo *mocks.MockClientRepository, audit *mocks.MockAuditRepository) {},
			wantErr:     true,
			errContains: "client_id is required",
		},
		{
			name: "Empty tenant ID",
			request: &client.UpdateClientRequest{
				ClientID:   "client-123",
				TenantID:   "",
				ClientName: strPtr("New Name"),
			},
			setupMocks:  func(repo *mocks.MockClientRepository, audit *mocks.MockAuditRepository) {},
			wantErr:     true,
			errContains: "tenant_id is required",
		},
		{
			name: "Repository update failure",
			request: &client.UpdateClientRequest{
				ClientID:   "client-123",
				TenantID:   "tenant-456",
				ClientName: strPtr("New Name"),
				IPAddress:  "127.0.0.1",
				UserAgent:  "test",
			},
			setupMocks: func(repo *mocks.MockClientRepository, audit *mocks.MockAuditRepository) {
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
			audit := new(mocks.MockAuditRepository)
			tt.setupMocks(repo, audit)

			uc := client.NewUpdateClientUseCase(repo, audit)

			resp, err := uc.Execute(context.Background(), tt.request)

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
	audit := new(mocks.MockAuditRepository)

	var updatedClient *entities.Client
	repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").
		Return(existingClient, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Client")).
		Run(func(args mock.Arguments) {
			updatedClient = args.Get(1).(*entities.Client)
		}).
		Return(nil)
	audit.On("Create", mock.Anything, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	uc := client.NewUpdateClientUseCase(repo, audit)

	req := &client.UpdateClientRequest{
		ClientID:   "client-123",
		TenantID:   "tenant-456",
		ClientName: strPtr("Updated Name"),
		IPAddress:  "127.0.0.1",
		UserAgent:  "test",
	}

	_, err := uc.Execute(context.Background(), req)
	require.NoError(t, err)

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
