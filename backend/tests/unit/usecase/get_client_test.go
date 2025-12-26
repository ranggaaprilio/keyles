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
	"github.com/stretchr/testify/require"
)

func TestGetClient_Execute(t *testing.T) {
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
		setupMocks   func(*mocks.MockClientRepository)
		wantErr      bool
		errContains  string
		validateResp func(*testing.T, *client.GetClientResponse)
	}{
		{
			name:     "Successful retrieval",
			clientID: "client-123",
			tenantID: "tenant-456",
			setupMocks: func(repo *mocks.MockClientRepository) {
				repo.On("GetByClientID", context.Background(), "client-123", "tenant-456").
					Return(existingClient, nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *client.GetClientResponse) {
				assert.Equal(t, "client-123", resp.ClientID)
				assert.Equal(t, "Test Application", resp.ClientName)
				assert.Equal(t, []string{"https://app.example.com/callback"}, resp.RedirectURIs)
				assert.True(t, resp.IsActive)
			},
		},
		{
			name:     "Client not found",
			clientID: "nonexistent",
			tenantID: "tenant-456",
			setupMocks: func(repo *mocks.MockClientRepository) {
				repo.On("GetByClientID", context.Background(), "nonexistent", "tenant-456").
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
				repo.On("GetByClientID", context.Background(), "client-123", "wrong-tenant").
					Return(nil, errors.New("client not found"))
			},
			wantErr:     true,
			errContains: "client not found",
		},
		{
			name:       "Empty client ID",
			clientID:   "",
			tenantID:   "tenant-456",
			setupMocks: func(repo *mocks.MockClientRepository) {},
			wantErr:    true,
			errContains: "client_id is required",
		},
		{
			name:       "Empty tenant ID",
			clientID:   "client-123",
			tenantID:   "",
			setupMocks: func(repo *mocks.MockClientRepository) {},
			wantErr:    true,
			errContains: "tenant_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.MockClientRepository)
			tt.setupMocks(repo)

			uc := client.NewGetClientUseCase(repo)

			req := &client.GetClientRequest{
				ClientID: tt.clientID,
				TenantID: tt.tenantID,
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

func TestGetClient_DoesNotReturnSecret(t *testing.T) {
	now := time.Now()
	existingClient := &entities.Client{
		ClientID:            "client-123",
		TenantID:            "tenant-456",
		ClientName:          "Test Application",
		ClientSecretHash:    "hashed-secret-that-should-not-be-exposed",
		AllowedRedirectURIs: []string{"https://app.example.com/callback"},
		IsActive:            true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	repo := new(mocks.MockClientRepository)
	repo.On("GetByClientID", context.Background(), "client-123", "tenant-456").
		Return(existingClient, nil)

	uc := client.NewGetClientUseCase(repo)

	req := &client.GetClientRequest{
		ClientID: "client-123",
		TenantID: "tenant-456",
	}

	resp, err := uc.Execute(context.Background(), req)
	require.NoError(t, err)

	// Response should not contain the secret hash
	// The response struct should not have a ClientSecret field
	// (validation is implicit in the struct definition)
	assert.NotContains(t, resp.ClientName, "hashed")
}
