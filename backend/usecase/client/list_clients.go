package client

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// ListClientsRequest contains the data for listing OAuth clients
type ListClientsRequest struct {
	TenantID string
	Search   string
	Page     int
	PageSize int
}

// ClientListItem represents a single client in the list
type ClientListItem struct {
	ClientID     string
	ClientName   string
	Description  string
	ClientType   string
	RedirectURIs []string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ListClientsResponse contains the paginated list of clients
type ListClientsResponse struct {
	Clients    []ClientListItem
	Total      int
	Page       int
	PageSize   int
	TotalPages int
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

// Execute lists all OAuth clients for a tenant with pagination and search
func (uc *ListClientsUseCase) Execute(ctx context.Context, req *ListClientsRequest) (*ListClientsResponse, error) {
	// Validate request
	if req.TenantID == "" {
		return nil, errors.New("tenant_id is required")
	}

	// Apply defaults
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 25 {
		pageSize = 25
	}

	// Get clients from repository with pagination
	clients, total, err := uc.clientRepo.ListByTenantPaginated(ctx, req.TenantID, req.Search, page, pageSize)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	items := make([]ClientListItem, len(clients))
	for i, c := range clients {
		items[i] = ClientListItem{
			ClientID:     c.ClientID,
			ClientName:   c.ClientName,
			Description:  c.Description,
			ClientType:   c.ClientType,
			RedirectURIs: c.AllowedRedirectURIs,
			IsActive:     c.IsActive,
			CreatedAt:    c.CreatedAt,
			UpdatedAt:    c.UpdatedAt,
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &ListClientsResponse{
		Clients:    items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
