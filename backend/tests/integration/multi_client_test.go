/**
 * Integration tests for multi-client scenarios
 * Tests that multiple clients for one tenant have independent configurations
 * and cannot use each other's tokens/credentials
 */

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMultiClientIndependentCredentials tests that multiple clients have independent credentials
func TestMultiClientIndependentCredentials(t *testing.T) {
	repo := NewMockIntegrationClientRepository()
	ctx := context.Background()

	// Create first client
	client1 := &entities.Client{
		ClientID:             "client-a-123",
		TenantID:             "tenant-1",
		ClientName:           "Client App A",
		ClientSecretHash:     "hash_secret_a",
		AllowedRedirectURIs:  []string{"https://app-a.example.com/callback"},
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err := repo.Create(ctx, client1)
	require.NoError(t, err)

	// Create second client
	client2 := &entities.Client{
		ClientID:             "client-b-456",
		TenantID:             "tenant-1",
		ClientName:           "Client App B",
		ClientSecretHash:     "hash_secret_b",
		AllowedRedirectURIs:  []string{"https://app-b.example.com/callback"},
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err = repo.Create(ctx, client2)
	require.NoError(t, err)

	// Verify both clients exist with different credentials
	retrieved1, err := repo.GetByID(ctx, client1.ClientID)
	require.NoError(t, err)
	assert.Equal(t, "Client App A", retrieved1.ClientName)

	retrieved2, err := repo.GetByID(ctx, client2.ClientID)
	require.NoError(t, err)
	assert.Equal(t, "Client App B", retrieved2.ClientName)

	// Verify credentials are different
	assert.NotEqual(t, retrieved1.ClientSecretHash, retrieved2.ClientSecretHash)
}

// TestMultiClientRedirectURIIndependence tests that redirect URI updates affect only one client
func TestMultiClientRedirectURIIndependence(t *testing.T) {
	repo := NewMockIntegrationClientRepository()
	ctx := context.Background()

	// Create two clients
	client1 := &entities.Client{
		ClientID:             "client-a-123",
		TenantID:             "tenant-1",
		ClientName:           "Client A",
		ClientSecretHash:     "secret_a",
		AllowedRedirectURIs:  []string{"https://app-a.example.com/callback"},
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	client2 := &entities.Client{
		ClientID:             "client-b-456",
		TenantID:             "tenant-1",
		ClientName:           "Client B",
		ClientSecretHash:     "secret_b",
		AllowedRedirectURIs:  []string{"https://app-b.example.com/callback"},
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	repo.Create(ctx, client1)
	repo.Create(ctx, client2)

	// Update client 1's redirect URI
	client1.AllowedRedirectURIs = []string{"https://app-a-new.example.com/callback"}
	client1.ClientName = "Client A Updated"
	client1.UpdatedAt = time.Now()

	err := repo.Update(ctx, client1)
	require.NoError(t, err)

	// Verify client 1 was updated
	retrieved1, _ := repo.GetByID(ctx, client1.ClientID)
	assert.Equal(t, "Client A Updated", retrieved1.ClientName)
	assert.Contains(t, retrieved1.AllowedRedirectURIs, "https://app-a-new.example.com/callback")
	assert.NotContains(t, retrieved1.AllowedRedirectURIs, "https://app-a.example.com/callback")

	// Verify client 2 was NOT changed
	retrieved2, _ := repo.GetByID(ctx, client2.ClientID)
	assert.Equal(t, "Client B", retrieved2.ClientName)
	assert.Contains(t, retrieved2.AllowedRedirectURIs, "https://app-b.example.com/callback")
}

// TestMultiClientTenantIsolation tests that clients from different tenants are isolated
func TestMultiClientTenantIsolation(t *testing.T) {
	repo := NewMockIntegrationClientRepository()
	ctx := context.Background()

	// Create clients for different tenants
	clientT1 := &entities.Client{
		ClientID:             "client-t1-123",
		TenantID:             "tenant-1",
		ClientName:           "Tenant 1 Client",
		ClientSecretHash:     "secret_t1",
		AllowedRedirectURIs:  []string{"https://tenant1.example.com/callback"},
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	clientT2 := &entities.Client{
		ClientID:             "client-t2-456",
		TenantID:             "tenant-2",
		ClientName:           "Tenant 2 Client",
		ClientSecretHash:     "secret_t2",
		AllowedRedirectURIs:  []string{"https://tenant2.example.com/callback"},
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	repo.Create(ctx, clientT1)
	repo.Create(ctx, clientT2)

	// List clients for tenant 1
	clientsT1, err := repo.ListByTenant(ctx, "tenant-1")
	require.NoError(t, err)
	assert.Greater(t, len(clientsT1), 0)

	// List clients for tenant 2
	clientsT2, err := repo.ListByTenant(ctx, "tenant-2")
	require.NoError(t, err)
	assert.Greater(t, len(clientsT2), 0)

	// Verify no overlapping clients between tenants
	for _, c1 := range clientsT1 {
		for _, c2 := range clientsT2 {
			assert.NotEqual(t, c1.ClientID, c2.ClientID)
		}
	}

	// Verify GetByClientID respects tenant boundaries
	retrievedT1, err := repo.GetByClientID(ctx, clientT1.ClientID, "tenant-1")
	require.NoError(t, err)
	assert.Equal(t, "Tenant 1 Client", retrievedT1.ClientName)

	// Try to get tenant 1's client with tenant 2's ID - should fail
	retrievedWrongTenant, err := repo.GetByClientID(ctx, clientT1.ClientID, "tenant-2")
	assert.Error(t, err)
	assert.Nil(t, retrievedWrongTenant)
}

// TestMultiClientListingPerTenant tests that listing clients returns only tenant's clients
func TestMultiClientListingPerTenant(t *testing.T) {
	repo := NewMockIntegrationClientRepository()
	ctx := context.Background()

	// Create 3 clients for tenant 1
	for i := 1; i <= 3; i++ {
		client := &entities.Client{
			ClientID:             "client-t1-" + string(rune(48+i)),
			TenantID:             "tenant-1",
			ClientName:           "Tenant 1 Client " + string(rune(48+i)),
			ClientSecretHash:     "secret_t1_" + string(rune(48+i)),
			AllowedRedirectURIs:  []string{"https://tenant1-app" + string(rune(48+i)) + ".example.com/callback"},
			IsActive:             true,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}
		repo.Create(ctx, client)
	}

	// Create 2 clients for tenant 2
	for i := 1; i <= 2; i++ {
		client := &entities.Client{
			ClientID:             "client-t2-" + string(rune(48+i)),
			TenantID:             "tenant-2",
			ClientName:           "Tenant 2 Client " + string(rune(48+i)),
			ClientSecretHash:     "secret_t2_" + string(rune(48+i)),
			AllowedRedirectURIs:  []string{"https://tenant2-app" + string(rune(48+i)) + ".example.com/callback"},
			IsActive:             true,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}
		repo.Create(ctx, client)
	}

	// List clients for tenant 1
	clientsT1, err := repo.ListByTenant(ctx, "tenant-1")
	require.NoError(t, err)
	assert.Equal(t, 3, len(clientsT1), "Tenant 1 should have 3 clients")

	// List clients for tenant 2
	clientsT2, err := repo.ListByTenant(ctx, "tenant-2")
	require.NoError(t, err)
	assert.Equal(t, 2, len(clientsT2), "Tenant 2 should have 2 clients")

	// Verify all clients in tenant 1 list have tenant-1 ID
	for _, c := range clientsT1 {
		assert.Equal(t, "tenant-1", c.TenantID)
	}

	// Verify all clients in tenant 2 list have tenant-2 ID
	for _, c := range clientsT2 {
		assert.Equal(t, "tenant-2", c.TenantID)
	}
}

// TestMultiClientDeletion tests that deleting one client doesn't affect others
func TestMultiClientDeletion(t *testing.T) {
	repo := NewMockIntegrationClientRepository()
	ctx := context.Background()

	// Create two clients
	client1 := &entities.Client{
		ClientID:             "client-a-123",
		TenantID:             "tenant-1",
		ClientName:           "Client A",
		ClientSecretHash:     "secret_a",
		AllowedRedirectURIs:  []string{"https://app-a.example.com/callback"},
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	client2 := &entities.Client{
		ClientID:             "client-b-456",
		TenantID:             "tenant-1",
		ClientName:           "Client B",
		ClientSecretHash:     "secret_b",
		AllowedRedirectURIs:  []string{"https://app-b.example.com/callback"},
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	repo.Create(ctx, client1)
	repo.Create(ctx, client2)

	// Delete first client
	err := repo.Delete(ctx, client1.ClientID)
	require.NoError(t, err)

	// Verify first client is soft-deleted (GetByID won't find it because it filters on IsActive)
	_, err = repo.GetByID(ctx, client1.ClientID)
	assert.Error(t, err, "Soft-deleted client should not be retrievable by GetByID")

	// Verify second client still exists and is active
	retrievedClient2, err := repo.GetByID(ctx, client2.ClientID)
	require.NoError(t, err)
	assert.True(t, retrievedClient2.IsActive)
	assert.Equal(t, "Client B", retrievedClient2.ClientName)
}

// TestMultiClientCredentialIsolation tests that credentials are stored independently
func TestMultiClientCredentialIsolation(t *testing.T) {
	repo := NewMockIntegrationClientRepository()
	ctx := context.Background()

	client1 := &entities.Client{
		ClientID:             "client-a",
		TenantID:             "tenant-1",
		ClientName:           "Client A",
		ClientSecretHash:     "secret_a",
		AllowedRedirectURIs:  []string{"https://app-a.example.com/callback"},
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	client2 := &entities.Client{
		ClientID:             "client-b",
		TenantID:             "tenant-1",
		ClientName:           "Client B",
		ClientSecretHash:     "secret_b",
		AllowedRedirectURIs:  []string{"https://app-b.example.com/callback"},
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	repo.Create(ctx, client1)
	repo.Create(ctx, client2)

	// Verify both clients have different credentials
	retrieved1, _ := repo.GetByID(ctx, client1.ClientID)
	retrieved2, _ := repo.GetByID(ctx, client2.ClientID)

	assert.NotEqual(t, retrieved1.ClientSecretHash, retrieved2.ClientSecretHash)
	assert.Equal(t, "secret_a", retrieved1.ClientSecretHash)
	assert.Equal(t, "secret_b", retrieved2.ClientSecretHash)
}
