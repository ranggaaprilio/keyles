package usecase_test

import (
	"context"
	"testing"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateClient_Execute(t *testing.T) {
	tests := []struct {
		name         string
		request      *client.CreateClientRequest
		setupMocks   func(*mocks.MockClientRepository, *mocks.MockPasswordService, *mocks.MockAuditRepository, *mocks.MockClientCountCache)
		wantErr      bool
		errContains  string
		validateResp func(*testing.T, *client.CreateClientResponse)
	}{
		{
			name: "Successful confidential client creation",
			request: &client.CreateClientRequest{
				TenantID:     "tenant-123",
				ClientName:   "Test Application",
				Description:  "A test app",
				ClientType:   "confidential",
				RedirectURIs: []string{"https://app.example.com/callback"},
				IPAddress:    "127.0.0.1",
				UserAgent:    "test",
			},
			setupMocks: func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService, audit *mocks.MockAuditRepository, cache *mocks.MockClientCountCache) {
				repo.On("CountByTenant", mock.Anything, "tenant-123").Return(0, nil)
				pwd.On("Hash", mock.AnythingOfType("string")).Return("hashed-secret", nil)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Client")).Return(nil)
				cache.On("Invalidate", mock.Anything, "tenant-123").Return(nil)
				audit.On("Create", mock.Anything, mock.AnythingOfType("*entities.AuditLog")).Return(nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *client.CreateClientResponse) {
				assert.NotEmpty(t, resp.ClientID)
				assert.NotEmpty(t, resp.ClientSecret)
				assert.Equal(t, "Test Application", resp.ClientName)
				assert.Equal(t, "confidential", resp.ClientType)
				assert.Equal(t, []string{"https://app.example.com/callback"}, resp.RedirectURIs)
			},
		},
		{
			name: "Successful public client creation - no secret",
			request: &client.CreateClientRequest{
				TenantID:     "tenant-123",
				ClientName:   "Public SPA",
				ClientType:   "public",
				RedirectURIs: []string{"https://spa.example.com/callback"},
				IPAddress:    "127.0.0.1",
				UserAgent:    "test",
			},
			setupMocks: func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService, audit *mocks.MockAuditRepository, cache *mocks.MockClientCountCache) {
				repo.On("CountByTenant", mock.Anything, "tenant-123").Return(0, nil)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Client")).Return(nil)
				cache.On("Invalidate", mock.Anything, "tenant-123").Return(nil)
				audit.On("Create", mock.Anything, mock.AnythingOfType("*entities.AuditLog")).Return(nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *client.CreateClientResponse) {
				assert.NotEmpty(t, resp.ClientID)
				assert.Empty(t, resp.ClientSecret, "public clients should not have a secret")
				assert.Equal(t, "public", resp.ClientType)
			},
		},
		{
			name: "Quota exceeded",
			request: &client.CreateClientRequest{
				TenantID:     "tenant-123",
				ClientName:   "Too Many Clients",
				ClientType:   "confidential",
				RedirectURIs: []string{"https://app.example.com/callback"},
			},
			setupMocks: func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService, audit *mocks.MockAuditRepository, cache *mocks.MockClientCountCache) {
				repo.On("CountByTenant", mock.Anything, "tenant-123").Return(entities.MaxClientsPerTenant, nil)
			},
			wantErr:     true,
			errContains: "quota exceeded",
		},
		{
			name: "Empty tenant ID",
			request: &client.CreateClientRequest{
				TenantID:     "",
				ClientName:   "Test",
				ClientType:   "confidential",
				RedirectURIs: []string{"https://app.example.com/callback"},
			},
			setupMocks:  func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService, audit *mocks.MockAuditRepository, cache *mocks.MockClientCountCache) {},
			wantErr:     true,
			errContains: "tenant_id is required",
		},
		{
			name: "Empty client name",
			request: &client.CreateClientRequest{
				TenantID:     "tenant-123",
				ClientName:   "",
				ClientType:   "confidential",
				RedirectURIs: []string{"https://app.example.com/callback"},
			},
			setupMocks:  func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService, audit *mocks.MockAuditRepository, cache *mocks.MockClientCountCache) {},
			wantErr:     true,
			errContains: "client_name is required",
		},
		{
			name: "No redirect URIs",
			request: &client.CreateClientRequest{
				TenantID:     "tenant-123",
				ClientName:   "Test Application",
				ClientType:   "confidential",
				RedirectURIs: []string{},
			},
			setupMocks:  func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService, audit *mocks.MockAuditRepository, cache *mocks.MockClientCountCache) {},
			wantErr:     true,
			errContains: "at least one redirect_uri is required",
		},
		{
			name: "Invalid redirect URI - with fragment",
			request: &client.CreateClientRequest{
				TenantID:     "tenant-123",
				ClientName:   "Test Application",
				ClientType:   "confidential",
				RedirectURIs: []string{"https://app.example.com/callback#section"},
			},
			setupMocks: func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService, audit *mocks.MockAuditRepository, cache *mocks.MockClientCountCache) {
				repo.On("CountByTenant", mock.Anything, "tenant-123").Return(0, nil)
			},
			wantErr:     true,
			errContains: "fragment",
		},
		{
			name: "Password hashing failure",
			request: &client.CreateClientRequest{
				TenantID:     "tenant-123",
				ClientName:   "Test Application",
				ClientType:   "confidential",
				RedirectURIs: []string{"https://app.example.com/callback"},
			},
			setupMocks: func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService, audit *mocks.MockAuditRepository, cache *mocks.MockClientCountCache) {
				repo.On("CountByTenant", mock.Anything, "tenant-123").Return(0, nil)
				pwd.On("Hash", mock.AnythingOfType("string")).Return("", assert.AnError)
			},
			wantErr:     true,
			errContains: "failed to hash client secret",
		},
		{
			name: "Repository create failure",
			request: &client.CreateClientRequest{
				TenantID:     "tenant-123",
				ClientName:   "Test Application",
				ClientType:   "confidential",
				RedirectURIs: []string{"https://app.example.com/callback"},
			},
			setupMocks: func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService, audit *mocks.MockAuditRepository, cache *mocks.MockClientCountCache) {
				repo.On("CountByTenant", mock.Anything, "tenant-123").Return(0, nil)
				pwd.On("Hash", mock.AnythingOfType("string")).Return("hashed-secret", nil)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Client")).Return(assert.AnError)
			},
			wantErr:     true,
			errContains: "failed to create client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.MockClientRepository)
			pwd := new(mocks.MockPasswordService)
			audit := new(mocks.MockAuditRepository)
			cache := new(mocks.MockClientCountCache)
			tt.setupMocks(repo, pwd, audit, cache)

			uc := client.NewCreateClientUseCase(repo, pwd, audit, cache)

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

func TestCreateClient_GeneratesUniqueCredentials(t *testing.T) {
	repo := new(mocks.MockClientRepository)
	pwd := new(mocks.MockPasswordService)
	audit := new(mocks.MockAuditRepository)
	cache := new(mocks.MockClientCountCache)

	repo.On("CountByTenant", mock.Anything, "tenant-123").Return(0, nil)
	pwd.On("Hash", mock.AnythingOfType("string")).Return("hashed-secret", nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Client")).Return(nil)
	cache.On("Invalidate", mock.Anything, "tenant-123").Return(nil)
	audit.On("Create", mock.Anything, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	uc := client.NewCreateClientUseCase(repo, pwd, audit, cache)

	req := &client.CreateClientRequest{
		TenantID:     "tenant-123",
		ClientName:   "Test Application",
		ClientType:   "confidential",
		RedirectURIs: []string{"https://app.example.com/callback"},
	}

	resp1, err1 := uc.Execute(context.Background(), req)
	require.NoError(t, err1)

	resp2, err2 := uc.Execute(context.Background(), req)
	require.NoError(t, err2)

	assert.NotEqual(t, resp1.ClientID, resp2.ClientID)
	assert.NotEqual(t, resp1.ClientSecret, resp2.ClientSecret)
}

func TestCreateClient_SetsDefaultValues(t *testing.T) {
	repo := new(mocks.MockClientRepository)
	pwd := new(mocks.MockPasswordService)
	audit := new(mocks.MockAuditRepository)
	cache := new(mocks.MockClientCountCache)

	var createdClient *entities.Client
	repo.On("CountByTenant", mock.Anything, "tenant-123").Return(0, nil)
	pwd.On("Hash", mock.AnythingOfType("string")).Return("hashed-secret", nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Client")).
		Run(func(args mock.Arguments) {
			createdClient = args.Get(1).(*entities.Client)
		}).
		Return(nil)
	cache.On("Invalidate", mock.Anything, "tenant-123").Return(nil)
	audit.On("Create", mock.Anything, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	uc := client.NewCreateClientUseCase(repo, pwd, audit, cache)

	req := &client.CreateClientRequest{
		TenantID:     "tenant-123",
		ClientName:   "Test Application",
		ClientType:   "confidential",
		RedirectURIs: []string{"https://app.example.com/callback"},
	}

	_, err := uc.Execute(context.Background(), req)
	require.NoError(t, err)

	assert.True(t, createdClient.IsActive)
	assert.False(t, createdClient.CreatedAt.IsZero())
	assert.False(t, createdClient.UpdatedAt.IsZero())
}
