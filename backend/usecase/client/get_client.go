package client

import (
	"context"
	"errors"
	"time"

	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// GetClientRequest contains the data for retrieving an OAuth client
type GetClientRequest struct {
	ClientID string
	TenantID string
}

// GetClientResponse contains the client data (without secret)
type GetClientResponse struct {
	ClientID     string
	ClientName   string
	RedirectURIs []string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// GetClientUseCase handles retrieval of OAuth clients
type GetClientUseCase struct {
	clientRepo repositories.ClientRepository
}

// NewGetClientUseCase creates a new GetClientUseCase
func NewGetClientUseCase(clientRepo repositories.ClientRepository) *GetClientUseCase {
	return &GetClientUseCase{
		clientRepo: clientRepo,
	}
}

// Execute retrieves an OAuth client by ID (scoped to tenant)
func (uc *GetClientUseCase) Execute(ctx context.Context, req *GetClientRequest) (*GetClientResponse, error) {
	// Validate request
	if req.ClientID == "" {
		return nil, errors.New("client_id is required")
	}
	if req.TenantID == "" {
		return nil, errors.New("tenant_id is required")
	}

	// Get client from repository (scoped by tenant)
	client, err := uc.clientRepo.GetByClientID(ctx, req.ClientID, req.TenantID)
	if err != nil {
		return nil, err
	}

	return &GetClientResponse{
		ClientID:     client.ClientID,
		ClientName:   client.ClientName,
		RedirectURIs: client.AllowedRedirectURIs,
		IsActive:     client.IsActive,
		CreatedAt:    client.CreatedAt,
		UpdatedAt:    client.UpdatedAt,
	}, nil
}
