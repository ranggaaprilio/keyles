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
		ClientType:          "confidential",
		ClientSecretHash:    "hashed-secret",
		AllowedRedirectURIs: []string{"https://app.example.com/callback"},
		IsActive:            true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	tests := []struct {
		name        string
		request     *client.DeleteClientRequest
		setupMocks  func(*mocks.MockClientRepository, *mocks.MockAuditRepository, *mocks.MockRefreshTokenRepository, *mocks.MockRevokedClientCache, *mocks.MockClientCountCache)
		wantErr     bool
		errContains string
	}{
		{
			name: "Successful deletion",
			request: &client.DeleteClientRequest{
				ClientID:  "client-123",
				TenantID:  "tenant-456",
				IPAddress: "127.0.0.1",
				UserAgent: "test",
			},
			setupMocks: func(repo *mocks.MockClientRepository, audit *mocks.MockAuditRepository, rt *mocks.MockRefreshTokenRepository, revoked *mocks.MockRevokedClientCache, countCache *mocks.MockClientCountCache) {
				repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").
					Return(existingClient, nil)
				repo.On("Delete", mock.Anything, "client-123").Return(nil)
				rt.On("RevokeByClientID", mock.Anything, "client-123").Return(nil)
				revoked.On("Revoke", mock.Anything, "client-123").Return(nil)
				countCache.On("Invalidate", mock.Anything, "tenant-456").Return(nil)
				audit.On("Create", mock.Anything, mock.AnythingOfType("*entities.AuditLog")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Client not found",
			request: &client.DeleteClientRequest{
				ClientID: "nonexistent",
				TenantID: "tenant-456",
			},
			setupMocks: func(repo *mocks.MockClientRepository, audit *mocks.MockAuditRepository, rt *mocks.MockRefreshTokenRepository, revoked *mocks.MockRevokedClientCache, countCache *mocks.MockClientCountCache) {
				repo.On("GetByClientID", mock.Anything, "nonexistent", "tenant-456").
					Return(nil, errors.New("client not found"))
			},
			wantErr:     true,
			errContains: "client not found",
		},
		{
			name: "Empty client ID",
			request: &client.DeleteClientRequest{
				ClientID: "",
				TenantID: "tenant-456",
			},
			setupMocks:  func(repo *mocks.MockClientRepository, audit *mocks.MockAuditRepository, rt *mocks.MockRefreshTokenRepository, revoked *mocks.MockRevokedClientCache, countCache *mocks.MockClientCountCache) {},
			wantErr:     true,
			errContains: "client_id is required",
		},
		{
			name: "Empty tenant ID",
			request: &client.DeleteClientRequest{
				ClientID: "client-123",
				TenantID: "",
			},
			setupMocks:  func(repo *mocks.MockClientRepository, audit *mocks.MockAuditRepository, rt *mocks.MockRefreshTokenRepository, revoked *mocks.MockRevokedClientCache, countCache *mocks.MockClientCountCache) {},
			wantErr:     true,
			errContains: "tenant_id is required",
		},
		{
			name: "Repository delete failure",
			request: &client.DeleteClientRequest{
				ClientID:  "client-123",
				TenantID:  "tenant-456",
				IPAddress: "127.0.0.1",
				UserAgent: "test",
			},
			setupMocks: func(repo *mocks.MockClientRepository, audit *mocks.MockAuditRepository, rt *mocks.MockRefreshTokenRepository, revoked *mocks.MockRevokedClientCache, countCache *mocks.MockClientCountCache) {
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
			audit := new(mocks.MockAuditRepository)
			rt := new(mocks.MockRefreshTokenRepository)
			revoked := new(mocks.MockRevokedClientCache)
			countCache := new(mocks.MockClientCountCache)
			tt.setupMocks(repo, audit, rt, revoked, countCache)

			uc := client.NewDeleteClientUseCase(repo, audit, rt, revoked, countCache)

			err := uc.Execute(context.Background(), tt.request)

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

func TestDeleteClient_RevokesTokensAndCache(t *testing.T) {
	now := time.Now()
	existingClient := &entities.Client{
		ClientID:            "client-123",
		TenantID:            "tenant-456",
		ClientName:          "Test Application",
		ClientType:          "confidential",
		ClientSecretHash:    "hashed-secret",
		AllowedRedirectURIs: []string{"https://app.example.com/callback"},
		IsActive:            true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	repo := new(mocks.MockClientRepository)
	audit := new(mocks.MockAuditRepository)
	rt := new(mocks.MockRefreshTokenRepository)
	revoked := new(mocks.MockRevokedClientCache)
	countCache := new(mocks.MockClientCountCache)

	repo.On("GetByClientID", mock.Anything, "client-123", "tenant-456").Return(existingClient, nil)
	repo.On("Delete", mock.Anything, "client-123").Return(nil)
	rt.On("RevokeByClientID", mock.Anything, "client-123").Return(nil)
	revoked.On("Revoke", mock.Anything, "client-123").Return(nil)
	countCache.On("Invalidate", mock.Anything, "tenant-456").Return(nil)
	audit.On("Create", mock.Anything, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	uc := client.NewDeleteClientUseCase(repo, audit, rt, revoked, countCache)

	err := uc.Execute(context.Background(), &client.DeleteClientRequest{
		ClientID:  "client-123",
		TenantID:  "tenant-456",
		IPAddress: "127.0.0.1",
		UserAgent: "test",
	})
	require.NoError(t, err)

	// Verify all cascade operations were called
	rt.AssertCalled(t, "RevokeByClientID", mock.Anything, "client-123")
	revoked.AssertCalled(t, "Revoke", mock.Anything, "client-123")
	countCache.AssertCalled(t, "Invalidate", mock.Anything, "tenant-456")
	audit.AssertCalled(t, "Create", mock.Anything, mock.AnythingOfType("*entities.AuditLog"))
}
