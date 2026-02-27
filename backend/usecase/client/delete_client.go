package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
)

// DeleteClientRequest contains the data for deleting an OAuth client
type DeleteClientRequest struct {
	ClientID  string
	TenantID  string
	IPAddress string
	UserAgent string
}

// DeleteClientUseCase handles deletion of OAuth clients
type DeleteClientUseCase struct {
	clientRepo         repositories.ClientRepository
	auditRepo          repositories.AuditRepository
	refreshTokenRepo   repositories.RefreshTokenRepository
	revokedClientCache services.RevokedClientCache
	clientCountCache   services.ClientCountCache
}

// NewDeleteClientUseCase creates a new DeleteClientUseCase
func NewDeleteClientUseCase(
	clientRepo repositories.ClientRepository,
	auditRepo repositories.AuditRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
	revokedClientCache services.RevokedClientCache,
	clientCountCache services.ClientCountCache,
) *DeleteClientUseCase {
	return &DeleteClientUseCase{
		clientRepo:         clientRepo,
		auditRepo:          auditRepo,
		refreshTokenRepo:   refreshTokenRepo,
		revokedClientCache: revokedClientCache,
		clientCountCache:   clientCountCache,
	}
}

// Execute deletes (soft-deletes) an OAuth client and revokes all tokens
func (uc *DeleteClientUseCase) Execute(ctx context.Context, req *DeleteClientRequest) error {
	// Validate request
	if req.ClientID == "" {
		return errors.New("client_id is required")
	}
	if req.TenantID == "" {
		return errors.New("tenant_id is required")
	}

	// Verify client exists and belongs to tenant
	client, err := uc.clientRepo.GetByClientID(ctx, req.ClientID, req.TenantID)
	if err != nil {
		return err
	}

	// Soft-delete the client
	if err := uc.clientRepo.Delete(ctx, req.ClientID); err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	// Revoke all refresh tokens for this client
	if err := uc.refreshTokenRepo.RevokeByClientID(ctx, req.ClientID); err != nil {
		// Log but don't fail - client is already deleted
		fmt.Printf("warning: failed to revoke refresh tokens for client %s: %v\n", req.ClientID, err)
	}

	// Mark client as revoked in Redis cache (blocks access token usage)
	_ = uc.revokedClientCache.Revoke(ctx, req.ClientID)

	// Invalidate client count cache
	_ = uc.clientCountCache.Invalidate(ctx, req.TenantID)

	// Create audit log
	auditLog := entities.NewAuditLog(entities.EventClientDeleted, req.IPAddress, req.UserAgent).
		WithData("client_id", req.ClientID).
		WithData("client_name", client.ClientName).
		WithData("client_type", client.ClientType)

	if tenantUUID, err := uuid.Parse(req.TenantID); err == nil {
		auditLog.WithTenant(tenantUUID)
	}

	_ = uc.auditRepo.Create(ctx, auditLog)

	return nil
}
