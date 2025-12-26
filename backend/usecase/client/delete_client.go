package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// DeleteClientRequest contains the data for deleting an OAuth client
type DeleteClientRequest struct {
	ClientID string
	TenantID string
}

// DeleteClientUseCase handles deletion of OAuth clients
type DeleteClientUseCase struct {
	clientRepo repositories.ClientRepository
}

// NewDeleteClientUseCase creates a new DeleteClientUseCase
func NewDeleteClientUseCase(clientRepo repositories.ClientRepository) *DeleteClientUseCase {
	return &DeleteClientUseCase{
		clientRepo: clientRepo,
	}
}

// Execute deletes (soft-deletes) an OAuth client
func (uc *DeleteClientUseCase) Execute(ctx context.Context, req *DeleteClientRequest) error {
	// Validate request
	if req.ClientID == "" {
		return errors.New("client_id is required")
	}
	if req.TenantID == "" {
		return errors.New("tenant_id is required")
	}

	// Verify client exists and belongs to tenant
	_, err := uc.clientRepo.GetByClientID(ctx, req.ClientID, req.TenantID)
	if err != nil {
		return err
	}

	// Soft-delete the client
	if err := uc.clientRepo.Delete(ctx, req.ClientID); err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	return nil
}
