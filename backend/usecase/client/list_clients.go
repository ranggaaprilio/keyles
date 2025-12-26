package client

import (
	"context"
	"errors"
	"time"

	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// ListClientsRequest contains the data for listing OAuth clients
type ListClientsRequest struct {
	TenantID string
}

// ClientListItem represents a single client in the list
type ClientListItem struct {
	ClientID     string
	ClientName   string
	RedirectURIs []string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ListClientsResponse contains the list of clients
type ListClientsResponse struct {
	Clients []ClientListItem
	Total   int
}

// ListClientsUseCase handles listing OAuth clients for a tenant
type ListClientsUseCase struct {
	clientRepo repositories.ClientRepository
}

// NewListClientsUseCase creates a new ListClientsUseCase
func NewListClientsUseCase(clientRepo repositories.ClientRepository) *ListClientsUseCase {
	return &ListClientsUseCase{
		clientRepo: clientRepo,
	}
}

// Execute lists all OAuth clients for a tenant
func (uc *ListClientsUseCase) Execute(ctx context.Context, req *ListClientsRequest) (*ListClientsResponse, error) {
	// Validate request
	if req.TenantID == "" {
		return nil, errors.New("tenant_id is required")
	}

	// Get clients from repository
	clients, err := uc.clientRepo.ListByTenant(ctx, req.TenantID)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	items := make([]ClientListItem, len(clients))
	for i, c := range clients {
		items[i] = ClientListItem{
			ClientID:     c.ClientID,
			ClientName:   c.ClientName,
			RedirectURIs: c.AllowedRedirectURIs,
			IsActive:     c.IsActive,
			CreatedAt:    c.CreatedAt,
			UpdatedAt:    c.UpdatedAt,
		}
	}

	return &ListClientsResponse{
		Clients: items,
		Total:   len(items),
	}, nil
}
