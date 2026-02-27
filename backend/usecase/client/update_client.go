package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// UpdateClientRequest contains the data for updating an OAuth client
type UpdateClientRequest struct {
	ClientID     string
	TenantID     string
	ClientName   *string  // Optional - only update if provided
	Description  *string  // Optional - only update if provided
	RedirectURIs []string // Optional - only update if provided (non-nil)
	IsActive     *bool    // Optional - only update if provided
	IPAddress    string
	UserAgent    string
}

// UpdateClientResponse contains the updated client data
type UpdateClientResponse struct {
	ClientID     string
	ClientName   string
	Description  string
	ClientType   string
	RedirectURIs []string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UpdateClientUseCase handles updating OAuth clients
type UpdateClientUseCase struct {
	clientRepo repositories.ClientRepository
	auditRepo  repositories.AuditRepository
}

// NewUpdateClientUseCase creates a new UpdateClientUseCase
func NewUpdateClientUseCase(
	clientRepo repositories.ClientRepository,
	auditRepo repositories.AuditRepository,
) *UpdateClientUseCase {
	return &UpdateClientUseCase{
		clientRepo: clientRepo,
		auditRepo:  auditRepo,
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

	// Validate new redirect URIs if provided (strict HTTPS)
	if req.RedirectURIs != nil {
		tempClient := &entities.Client{}
		for _, uri := range req.RedirectURIs {
			if err := tempClient.ValidateRedirectURIStrict(uri, true); err != nil {
				return nil, err
			}
		}
	}

	// Track changed fields for audit
	changedFields := make(map[string]interface{})

	// Apply updates (only for provided fields)
	if req.ClientName != nil && *req.ClientName != client.ClientName {
		if len(*req.ClientName) < 3 || len(*req.ClientName) > 100 {
			return nil, errors.New("client_name must be between 3 and 100 characters")
		}
		changedFields["client_name"] = map[string]string{
			"old": client.ClientName,
			"new": *req.ClientName,
		}
		client.ClientName = *req.ClientName
	}
	if req.Description != nil && *req.Description != client.Description {
		if len(*req.Description) > 500 {
			return nil, errors.New("description must not exceed 500 characters")
		}
		changedFields["description"] = map[string]string{
			"old": client.Description,
			"new": *req.Description,
		}
		client.Description = *req.Description
	}
	if req.RedirectURIs != nil {
		changedFields["redirect_uris"] = "updated"
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

	// Create audit log
	if len(changedFields) > 0 {
		auditLog := entities.NewAuditLog(entities.EventClientUpdated, req.IPAddress, req.UserAgent).
			WithData("client_id", req.ClientID).
			WithData("changed_fields", changedFields)

		if tenantUUID, err := uuid.Parse(req.TenantID); err == nil {
			auditLog.WithTenant(tenantUUID)
		}

		_ = uc.auditRepo.Create(ctx, auditLog)
	}

	return &UpdateClientResponse{
		ClientID:     client.ClientID,
		ClientName:   client.ClientName,
		Description:  client.Description,
		ClientType:   client.ClientType,
		RedirectURIs: client.AllowedRedirectURIs,
		IsActive:     client.IsActive,
		CreatedAt:    client.CreatedAt,
		UpdatedAt:    client.UpdatedAt,
	}, nil
}
