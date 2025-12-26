package domain_test

import (
	"testing"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Validate(t *testing.T) {
	tests := []struct {
		name        string
		client      *entities.Client
		wantErr     bool
		errContains string
	}{
		{
			name: "Valid client with HTTPS redirect URI",
			client: &entities.Client{
				ClientID:            "test-client-id",
				TenantID:            "tenant-123",
				ClientName:          "Test Client",
				ClientSecretHash:    "hashed-secret",
				AllowedRedirectURIs: []string{"https://app.example.com/callback"},
				IsActive:            true,
			},
			wantErr: false,
		},
		{
			name: "Valid client with HTTP localhost redirect URI",
			client: &entities.Client{
				ClientID:            "test-client-id",
				TenantID:            "tenant-123",
				ClientName:          "Test Client",
				ClientSecretHash:    "hashed-secret",
				AllowedRedirectURIs: []string{"http://localhost:3000/callback"},
				IsActive:            true,
			},
			wantErr: false,
		},
		{
			name: "Valid client with multiple redirect URIs",
			client: &entities.Client{
				ClientID:            "test-client-id",
				TenantID:            "tenant-123",
				ClientName:          "Test Client",
				ClientSecretHash:    "hashed-secret",
				AllowedRedirectURIs: []string{
					"https://app.example.com/callback",
					"https://staging.example.com/callback",
					"http://localhost:3000/callback",
				},
				IsActive: true,
			},
			wantErr: false,
		},
		{
			name: "Missing client_id",
			client: &entities.Client{
				ClientID:            "",
				TenantID:            "tenant-123",
				ClientName:          "Test Client",
				ClientSecretHash:    "hashed-secret",
				AllowedRedirectURIs: []string{"https://app.example.com/callback"},
			},
			wantErr:     true,
			errContains: "client_id cannot be empty",
		},
		{
			name: "Missing tenant_id",
			client: &entities.Client{
				ClientID:            "test-client-id",
				TenantID:            "",
				ClientName:          "Test Client",
				ClientSecretHash:    "hashed-secret",
				AllowedRedirectURIs: []string{"https://app.example.com/callback"},
			},
			wantErr:     true,
			errContains: "tenant_id cannot be empty",
		},
		{
			name: "Missing client_name",
			client: &entities.Client{
				ClientID:            "test-client-id",
				TenantID:            "tenant-123",
				ClientName:          "",
				ClientSecretHash:    "hashed-secret",
				AllowedRedirectURIs: []string{"https://app.example.com/callback"},
			},
			wantErr:     true,
			errContains: "client_name cannot be empty",
		},
		{
			name: "Missing client_secret_hash",
			client: &entities.Client{
				ClientID:            "test-client-id",
				TenantID:            "tenant-123",
				ClientName:          "Test Client",
				ClientSecretHash:    "",
				AllowedRedirectURIs: []string{"https://app.example.com/callback"},
			},
			wantErr:     true,
			errContains: "client_secret_hash cannot be empty",
		},
		{
			name: "No redirect URIs",
			client: &entities.Client{
				ClientID:            "test-client-id",
				TenantID:            "tenant-123",
				ClientName:          "Test Client",
				ClientSecretHash:    "hashed-secret",
				AllowedRedirectURIs: []string{},
			},
			wantErr:     true,
			errContains: "at least one redirect URI is required",
		},
		{
			name: "Invalid redirect URI - no scheme",
			client: &entities.Client{
				ClientID:            "test-client-id",
				TenantID:            "tenant-123",
				ClientName:          "Test Client",
				ClientSecretHash:    "hashed-secret",
				AllowedRedirectURIs: []string{"app.example.com/callback"},
			},
			wantErr:     true,
			errContains: "must have a scheme",
		},
		{
			name: "Invalid redirect URI - with fragment",
			client: &entities.Client{
				ClientID:            "test-client-id",
				TenantID:            "tenant-123",
				ClientName:          "Test Client",
				ClientSecretHash:    "hashed-secret",
				AllowedRedirectURIs: []string{"https://app.example.com/callback#fragment"},
			},
			wantErr:     true,
			errContains: "must not contain a fragment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_ValidateRedirectURI(t *testing.T) {
	client := &entities.Client{}

	tests := []struct {
		name        string
		redirectURI string
		wantErr     bool
		errContains string
	}{
		{
			name:        "Valid HTTPS URI",
			redirectURI: "https://app.example.com/callback",
			wantErr:     false,
		},
		{
			name:        "Valid HTTPS URI with port",
			redirectURI: "https://app.example.com:8443/callback",
			wantErr:     false,
		},
		{
			name:        "Valid HTTPS URI with query params",
			redirectURI: "https://app.example.com/callback?param=value",
			wantErr:     false,
		},
		{
			name:        "Valid HTTP localhost URI",
			redirectURI: "http://localhost:3000/callback",
			wantErr:     false,
		},
		{
			name:        "Valid HTTP 127.0.0.1 URI",
			redirectURI: "http://127.0.0.1:3000/callback",
			wantErr:     false,
		},
		{
			name:        "Empty redirect URI",
			redirectURI: "",
			wantErr:     true,
			errContains: "redirect_uri cannot be empty",
		},
		{
			name:        "URI without scheme",
			redirectURI: "app.example.com/callback",
			wantErr:     true,
			errContains: "must have a scheme",
		},
		{
			name:        "URI without host",
			redirectURI: "https:///callback",
			wantErr:     true,
			errContains: "must have a host",
		},
		{
			name:        "URI with fragment",
			redirectURI: "https://app.example.com/callback#section",
			wantErr:     true,
			errContains: "must not contain a fragment",
		},
		{
			name:        "Custom scheme URI (mobile app)",
			redirectURI: "myapp://callback",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.ValidateRedirectURI(tt.redirectURI)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_IsURIAllowed(t *testing.T) {
	client := &entities.Client{
		ClientID:            "test-client",
		TenantID:            "tenant-123",
		ClientName:          "Test Client",
		ClientSecretHash:    "hashed-secret",
		AllowedRedirectURIs: []string{
			"https://app.example.com/callback",
			"https://staging.example.com/callback",
			"http://localhost:3000/callback",
		},
	}

	tests := []struct {
		name        string
		redirectURI string
		want        bool
	}{
		{
			name:        "Exact match - production",
			redirectURI: "https://app.example.com/callback",
			want:        true,
		},
		{
			name:        "Exact match - staging",
			redirectURI: "https://staging.example.com/callback",
			want:        true,
		},
		{
			name:        "Exact match - localhost",
			redirectURI: "http://localhost:3000/callback",
			want:        true,
		},
		{
			name:        "Not allowed - different subdomain",
			redirectURI: "https://dev.example.com/callback",
			want:        false,
		},
		{
			name:        "Not allowed - different path",
			redirectURI: "https://app.example.com/different-path",
			want:        false,
		},
		{
			name:        "Not allowed - different port",
			redirectURI: "http://localhost:4000/callback",
			want:        false,
		},
		{
			name:        "Not allowed - different scheme",
			redirectURI: "http://app.example.com/callback",
			want:        false,
		},
		{
			name:        "Not allowed - with query param (strict match)",
			redirectURI: "https://app.example.com/callback?code=123",
			want:        false,
		},
		{
			name:        "Not allowed - empty URI",
			redirectURI: "",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.IsURIAllowed(tt.redirectURI)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestClient_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		isActive bool
		want     bool
	}{
		{
			name:     "Active client is enabled",
			isActive: true,
			want:     true,
		},
		{
			name:     "Inactive client is not enabled",
			isActive: false,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &entities.Client{
				ClientID:            "test-client",
				TenantID:            "tenant-123",
				ClientName:          "Test Client",
				ClientSecretHash:    "hashed-secret",
				AllowedRedirectURIs: []string{"https://app.example.com/callback"},
				IsActive:            tt.isActive,
			}
			result := client.IsEnabled()
			assert.Equal(t, tt.want, result)
		})
	}
}
