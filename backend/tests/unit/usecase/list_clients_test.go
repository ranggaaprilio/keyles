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

func TestListClients_Execute(t *testing.T) {
	now := time.Now()
	clientList := []*entities.Client{
		{
			ClientID:            "client-1",
			TenantID:            "tenant-123",
			ClientName:          "App One",
			Description:         "First app",
			ClientType:          "confidential",
			AllowedRedirectURIs: []string{"https://one.example.com/callback"},
			IsActive:            true,
			CreatedAt:           now,
			UpdatedAt:           now,
		},
		{
			ClientID:            "client-2",
			TenantID:            "tenant-123",
			ClientName:          "App Two",
			Description:         "Second app",
			ClientType:          "public",
			AllowedRedirectURIs: []string{"https://two.example.com/callback"},
			IsActive:            true,
			CreatedAt:           now,
			UpdatedAt:           now,
		},
	}

	tests := []struct {
		name         string
		request      *client.ListClientsRequest
		setupMocks   func(*mocks.MockClientRepository)
		wantErr      bool
		errContains  string
		validateResp func(*testing.T, *client.ListClientsResponse)
	}{
		{
			name: "Successful listing",
			request: &client.ListClientsRequest{
				TenantID: "tenant-123",
				Page:     1,
				PageSize: 10,
			},
			setupMocks: func(repo *mocks.MockClientRepository) {
				repo.On("ListByTenantPaginated", mock.Anything, "tenant-123", "", 1, 10).
					Return(clientList, 2, nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *client.ListClientsResponse) {
				assert.Equal(t, 2, len(resp.Clients))
				assert.Equal(t, 2, resp.Total)
				assert.Equal(t, 1, resp.Page)
				assert.Equal(t, 10, resp.PageSize)
				assert.Equal(t, 1, resp.TotalPages)
				assert.Equal(t, "confidential", resp.Clients[0].ClientType)
				assert.Equal(t, "public", resp.Clients[1].ClientType)
			},
		},
		{
			name: "With search filter",
			request: &client.ListClientsRequest{
				TenantID: "tenant-123",
				Search:   "One",
				Page:     1,
				PageSize: 10,
			},
			setupMocks: func(repo *mocks.MockClientRepository) {
				repo.On("ListByTenantPaginated", mock.Anything, "tenant-123", "One", 1, 10).
					Return(clientList[:1], 1, nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *client.ListClientsResponse) {
				assert.Equal(t, 1, len(resp.Clients))
				assert.Equal(t, "App One", resp.Clients[0].ClientName)
			},
		},
		{
			name: "Default page and page_size when zero",
			request: &client.ListClientsRequest{
				TenantID: "tenant-123",
				Page:     0,
				PageSize: 0,
			},
			setupMocks: func(repo *mocks.MockClientRepository) {
				repo.On("ListByTenantPaginated", mock.Anything, "tenant-123", "", 1, 10).
					Return([]*entities.Client{}, 0, nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *client.ListClientsResponse) {
				assert.Equal(t, 1, resp.Page)
				assert.Equal(t, 10, resp.PageSize)
			},
		},
		{
			name: "Caps page size at 25",
			request: &client.ListClientsRequest{
				TenantID: "tenant-123",
				Page:     1,
				PageSize: 100,
			},
			setupMocks: func(repo *mocks.MockClientRepository) {
				repo.On("ListByTenantPaginated", mock.Anything, "tenant-123", "", 1, 25).
					Return([]*entities.Client{}, 0, nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *client.ListClientsResponse) {
				assert.Equal(t, 25, resp.PageSize)
			},
		},
		{
			name: "Empty tenant ID",
			request: &client.ListClientsRequest{
				TenantID: "",
			},
			setupMocks:  func(repo *mocks.MockClientRepository) {},
			wantErr:     true,
			errContains: "tenant_id is required",
		},
		{
			name: "Repository error",
			request: &client.ListClientsRequest{
				TenantID: "tenant-123",
				Page:     1,
				PageSize: 10,
			},
			setupMocks: func(repo *mocks.MockClientRepository) {
				repo.On("ListByTenantPaginated", mock.Anything, "tenant-123", "", 1, 10).
					Return(nil, 0, errors.New("database error"))
			},
			wantErr:     true,
			errContains: "database error",
		},
		{
			name: "Empty result set",
			request: &client.ListClientsRequest{
				TenantID: "tenant-123",
				Page:     1,
				PageSize: 10,
			},
			setupMocks: func(repo *mocks.MockClientRepository) {
				repo.On("ListByTenantPaginated", mock.Anything, "tenant-123", "", 1, 10).
					Return([]*entities.Client{}, 0, nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *client.ListClientsResponse) {
				assert.Equal(t, 0, len(resp.Clients))
				assert.Equal(t, 0, resp.Total)
				assert.Equal(t, 0, resp.TotalPages)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.MockClientRepository)
			tt.setupMocks(repo)

			uc := client.NewListClientsUseCase(repo)

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

func TestListClients_CalculatesTotalPages(t *testing.T) {
	repo := new(mocks.MockClientRepository)

	// 23 total with page_size 10 = 3 pages
	repo.On("ListByTenantPaginated", mock.Anything, "tenant-123", "", 1, 10).
		Return([]*entities.Client{}, 23, nil)

	uc := client.NewListClientsUseCase(repo)

	resp, err := uc.Execute(context.Background(), &client.ListClientsRequest{
		TenantID: "tenant-123",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, 3, resp.TotalPages)
}
