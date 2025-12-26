package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// UpdateClientRequest contains the data for updating an OAuth client
type UpdateClientRequest struct {
	ClientID     string
	TenantID     string
	ClientName   *string  // Optional - only update if provided
	RedirectURIs []string // Optional - only update if provided (non-nil)
	IsActive     *bool    // Optional - only update if provided
}

// UpdateClientResponse contains the updated client data
type UpdateClientResponse struct {
	ClientID     string
	ClientName   string
	RedirectURIs []string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UpdateClientUseCase handles updating OAuth clients
type UpdateClientUseCase struct {
	clientRepo repositories.ClientRepository
}

// NewUpdateClientUseCase creates a new UpdateClientUseCase
func NewUpdateClientUseCase(clientRepo repositories.ClientRepository) *UpdateClientUseCase {
	return &UpdateClientUseCase{
		clientRepo: clientRepo,
	}
}

// Execute updates an OAuth client
func (uc *UpdateClientUseCase) Execute(ctx context.Context, req *UpdateClientRequest) (*UpdateClientResponse, error) {
	// Validate request
	if req.ClientID == "" {
		return nil, errors.New("client_id is required")
	}
	if req.TenantID == "" {
		return nil, errors.New("tenant_id is required")
	}

	// Get existing client (scoped by tenant)
	client, err := uc.clientRepo.GetByClientID(ctx, req.ClientID, req.TenantID)
	if err != nil {
		return nil, err
	}

	// Validate new redirect URIs if provided
	if req.RedirectURIs != nil {
		tempClient := &entities.Client{}
		for _, uri := range req.RedirectURIs {
			if err := tempClient.ValidateRedirectURI(uri); err != nil {
				return nil, err
			}
		}
	}

	// Apply updates (only for provided fields)
	if req.ClientName != nil {
		client.ClientName = *req.ClientName
	}
	if req.RedirectURIs != nil {
		client.AllowedRedirectURIs = req.RedirectURIs
	}
	if req.IsActive != nil {
		client.IsActive = *req.IsActive
	}

	// Update timestamp
	client.UpdatedAt = time.Now()

	// Save changes
	if err := uc.clientRepo.Update(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to update client: %w", err)
	}

	return &UpdateClientResponse{
		ClientID:     client.ClientID,
		ClientName:   client.ClientName,
		RedirectURIs: client.AllowedRedirectURIs,
		IsActive:     client.IsActive,
		CreatedAt:    client.CreatedAt,
		UpdatedAt:    client.UpdatedAt,
	}, nil
}
