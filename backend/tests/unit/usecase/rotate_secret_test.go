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

func TestRotateSecret_Execute(t *testing.T) {
	now := time.Now()

	confidentialClient := &entities.Client{
		ClientID:            "client-123",
		TenantID:            "tenant-456",
		ClientName:          "Test Application",
		ClientType:          "confidential",
		ClientSecretHash:    "old-hashed-secret",
		AllowedRedirectURIs: []string{"https://app.example.com/callback"},
		IsActive:            true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	publicClient := &entities.Client{
		ClientID:            "client-public",
		TenantID:            "tenant-456",
		ClientName:          "Public SPA",
		ClientType:          "public",
		AllowedRedirectURIs: []string{"https://spa.example.com/callback"},
		IsActive:            true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	tests := []struct {
		name         string
		request      *client.RotateSecretRequest
		setupMocks   func(*mocks.MockClientRepository, *mocks.MockPasswordService, *mocks.MockAuditRepository)
		wantErr      bool
		errContains  string
		validateResp func(*testing.T, *client.RotateSecretResponse)
	}{
		{
			name: "Successful secret rotation",
			request: &client.RotateSecretRequest{
				ClientID:  "client-123",
				TenantID:  "tenant-456",
				IPAddress: "127.0.0.1",
				UserAgent: "test",
			},
			setupMocks: func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService, audit *mocks.MockAuditRepository) {
				repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").
					Return(confidentialClient, nil)
				pwd.On("Hash", mock.AnythingOfType("string")).Return("new-hashed-secret", nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Client")).Return(nil)
				audit.On("Create", mock.Anything, mock.AnythingOfType("*entities.AuditLog")).Return(nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *client.RotateSecretResponse) {
				assert.Equal(t, "client-123", resp.ClientID)
				assert.NotEmpty(t, resp.ClientSecret)
				assert.False(t, resp.RotatedAt.IsZero())
			},
		},
		{
			name: "Public client cannot rotate secret",
			request: &client.RotateSecretRequest{
				ClientID:  "client-public",
				TenantID:  "tenant-456",
				IPAddress: "127.0.0.1",
				UserAgent: "test",
			},
			setupMocks: func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService, audit *mocks.MockAuditRepository) {
				repo.On("GetByClientID", mock.Anything, "client-public", "tenant-456").
					Return(publicClient, nil)
			},
			wantErr:     true,
			errContains: "secret rotation is not available for public clients",
		},
		{
			name: "Client not found",
			request: &client.RotateSecretRequest{
				ClientID: "nonexistent",
				TenantID: "tenant-456",
			},
			setupMocks: func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService, audit *mocks.MockAuditRepository) {
				repo.On("GetByClientID", mock.Anything, "nonexistent", "tenant-456").
					Return(nil, errors.New("client not found"))
			},
			wantErr:     true,
			errContains: "client not found",
		},
		{
			name: "Empty client ID",
			request: &client.RotateSecretRequest{
				ClientID: "",
				TenantID: "tenant-456",
			},
			setupMocks:  func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService, audit *mocks.MockAuditRepository) {},
			wantErr:     true,
			errContains: "client_id is required",
		},
		{
			name: "Empty tenant ID",
			request: &client.RotateSecretRequest{
				ClientID: "client-123",
				TenantID: "",
			},
			setupMocks:  func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService, audit *mocks.MockAuditRepository) {},
			wantErr:     true,
			errContains: "tenant_id is required",
		},
		{
			name: "Hash failure",
			request: &client.RotateSecretRequest{
				ClientID:  "client-123",
				TenantID:  "tenant-456",
				IPAddress: "127.0.0.1",
				UserAgent: "test",
			},
			setupMocks: func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService, audit *mocks.MockAuditRepository) {
				repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").
					Return(confidentialClient, nil)
				pwd.On("Hash", mock.AnythingOfType("string")).Return("", errors.New("hash error"))
			},
			wantErr:     true,
			errContains: "failed to hash",
		},
		{
			name: "Repository update failure",
			request: &client.RotateSecretRequest{
				ClientID:  "client-123",
				TenantID:  "tenant-456",
				IPAddress: "127.0.0.1",
				UserAgent: "test",
			},
			setupMocks: func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService, audit *mocks.MockAuditRepository) {
				repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").
					Return(confidentialClient, nil)
				pwd.On("Hash", mock.AnythingOfType("string")).Return("new-hashed-secret", nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Client")).
					Return(errors.New("update failed"))
			},
			wantErr:     true,
			errContains: "failed to update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
repo := new(mocks.MockClientRepository)
pwd := new(mocks.MockPasswordService)
audit := new(mocks.MockAuditRepository)
tt.setupMocks(repo, pwd, audit)

uc := client.NewRotateSecretUseCase(repo, pwd, audit)

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
			pwd.AssertExpectations(t)
		})
	}
}

func TestRotateSecret_GeneratesNewSecret(t *testing.T) {
	now := time.Now()
	confidentialClient := &entities.Client{
		ClientID:         "client-123",
		TenantID:         "tenant-456",
		ClientName:       "Test App",
		ClientType:       "confidential",
		ClientSecretHash: "old-hash",
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	repo := new(mocks.MockClientRepository)
	pwd := new(mocks.MockPasswordService)
	audit := new(mocks.MockAuditRepository)

	repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").Return(confidentialClient, nil)
	pwd.On("Hash", mock.AnythingOfType("string")).Return("new-hash", nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Client")).Return(nil)
	audit.On("Create", mock.Anything, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	uc := client.NewRotateSecretUseCase(repo, pwd, audit)

	resp, err := uc.Execute(context.Background(), &client.RotateSecretRequest{
		ClientID:  "client-123",
		TenantID:  "tenant-456",
		IPAddress: "127.0.0.1",
		UserAgent: "test",
	})
	require.NoError(t, err)
	// New secret should be non-empty and different from old hash
	assert.NotEmpty(t, resp.ClientSecret)
	assert.NotEqual(t, "old-hash", resp.ClientSecret)
}
