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
		name          string
		tenantID      string
		clientName    string
		redirectURIs  []string
		setupMocks    func(*mocks.MockClientRepository, *mocks.MockPasswordService)
		wantErr       bool
		errContains   string
		validateResp  func(*testing.T, *client.CreateClientResponse)
	}{
		{
			name:         "Successful client creation",
			tenantID:     "tenant-123",
			clientName:   "Test Application",
			redirectURIs: []string{"https://app.example.com/callback"},
			setupMocks: func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService) {
				pwd.On("Hash", mock.AnythingOfType("string")).Return("hashed-secret", nil)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Client")).Return(nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *client.CreateClientResponse) {
				assert.NotEmpty(t, resp.ClientID)
				assert.NotEmpty(t, resp.ClientSecret)
				assert.Equal(t, "Test Application", resp.ClientName)
				assert.Equal(t, []string{"https://app.example.com/callback"}, resp.RedirectURIs)
			},
		},
		{
			name:         "Successful client creation with multiple redirect URIs",
			tenantID:     "tenant-123",
			clientName:   "Multi-Redirect App",
			redirectURIs: []string{
				"https://app.example.com/callback",
				"https://staging.example.com/callback",
				"http://localhost:3000/callback",
			},
			setupMocks: func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService) {
				pwd.On("Hash", mock.AnythingOfType("string")).Return("hashed-secret", nil)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Client")).Return(nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *client.CreateClientResponse) {
				assert.NotEmpty(t, resp.ClientID)
				assert.Equal(t, 3, len(resp.RedirectURIs))
			},
		},
		{
			name:         "Empty tenant ID",
			tenantID:     "",
			clientName:   "Test Application",
			redirectURIs: []string{"https://app.example.com/callback"},
			setupMocks:   func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService) {},
			wantErr:      true,
			errContains:  "tenant_id is required",
		},
		{
			name:         "Empty client name",
			tenantID:     "tenant-123",
			clientName:   "",
			redirectURIs: []string{"https://app.example.com/callback"},
			setupMocks:   func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService) {},
			wantErr:      true,
			errContains:  "client_name is required",
		},
		{
			name:         "No redirect URIs",
			tenantID:     "tenant-123",
			clientName:   "Test Application",
			redirectURIs: []string{},
			setupMocks:   func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService) {},
			wantErr:      true,
			errContains:  "at least one redirect_uri is required",
		},
		{
			name:         "Invalid redirect URI - missing scheme",
			tenantID:     "tenant-123",
			clientName:   "Test Application",
			redirectURIs: []string{"app.example.com/callback"},
			setupMocks:   func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService) {},
			wantErr:      true,
			errContains:  "must have a scheme",
		},
		{
			name:         "Invalid redirect URI - with fragment",
			tenantID:     "tenant-123",
			clientName:   "Test Application",
			redirectURIs: []string{"https://app.example.com/callback#section"},
			setupMocks:   func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService) {},
			wantErr:      true,
			errContains:  "must not contain a fragment",
		},
		{
			name:         "Password hashing failure",
			tenantID:     "tenant-123",
			clientName:   "Test Application",
			redirectURIs: []string{"https://app.example.com/callback"},
			setupMocks: func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService) {
				pwd.On("Hash", mock.AnythingOfType("string")).Return("", assert.AnError)
			},
			wantErr:     true,
			errContains: "failed to hash client secret",
		},
		{
			name:         "Repository create failure",
			tenantID:     "tenant-123",
			clientName:   "Test Application",
			redirectURIs: []string{"https://app.example.com/callback"},
			setupMocks: func(repo *mocks.MockClientRepository, pwd *mocks.MockPasswordService) {
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
			tt.setupMocks(repo, pwd)

			uc := client.NewCreateClientUseCase(repo, pwd)

			req := &client.CreateClientRequest{
				TenantID:     tt.tenantID,
				ClientName:   tt.clientName,
				RedirectURIs: tt.redirectURIs,
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
			pwd.AssertExpectations(t)
		})
	}
}

func TestCreateClient_GeneratesUniqueCredentials(t *testing.T) {
	repo := new(mocks.MockClientRepository)
	pwd := new(mocks.MockPasswordService)

	pwd.On("Hash", mock.AnythingOfType("string")).Return("hashed-secret", nil).Twice()
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Client")).Return(nil).Twice()

	uc := client.NewCreateClientUseCase(repo, pwd)

	req := &client.CreateClientRequest{
		TenantID:     "tenant-123",
		ClientName:   "Test Application",
		RedirectURIs: []string{"https://app.example.com/callback"},
	}

	resp1, err1 := uc.Execute(context.Background(), req)
	require.NoError(t, err1)

	resp2, err2 := uc.Execute(context.Background(), req)
	require.NoError(t, err2)

	// Client IDs should be unique
	assert.NotEqual(t, resp1.ClientID, resp2.ClientID)

	// Client secrets should be unique
	assert.NotEqual(t, resp1.ClientSecret, resp2.ClientSecret)
}

func TestCreateClient_SetsDefaultValues(t *testing.T) {
	repo := new(mocks.MockClientRepository)
	pwd := new(mocks.MockPasswordService)

	var createdClient *entities.Client
	pwd.On("Hash", mock.AnythingOfType("string")).Return("hashed-secret", nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Client")).
		Run(func(args mock.Arguments) {
			createdClient = args.Get(1).(*entities.Client)
		}).
		Return(nil)

	uc := client.NewCreateClientUseCase(repo, pwd)

	req := &client.CreateClientRequest{
		TenantID:     "tenant-123",
		ClientName:   "Test Application",
		RedirectURIs: []string{"https://app.example.com/callback"},
	}

	_, err := uc.Execute(context.Background(), req)
	require.NoError(t, err)

	// Should be active by default
	assert.True(t, createdClient.IsActive)

	// Should have creation timestamp
	assert.False(t, createdClient.CreatedAt.IsZero())
	assert.False(t, createdClient.UpdatedAt.IsZero())
}
