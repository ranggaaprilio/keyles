package client

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
)

// CreateClientRequest contains the data for creating a new OAuth client
type CreateClientRequest struct {
	TenantID     string
	ClientName   string
	Description  string
	ClientType   string // "confidential" or "public"
	RedirectURIs []string
	IPAddress    string
	UserAgent    string
}

// CreateClientResponse contains the created client data including the plain-text secret
type CreateClientResponse struct {
	ClientID     string
	ClientSecret string // Only returned once at creation time; empty for public clients
	ClientName   string
	Description  string
	ClientType   string
	RedirectURIs []string
	IsActive     bool
	CreatedAt    time.Time
}

// CreateClientUseCase handles the creation of OAuth clients
type CreateClientUseCase struct {
	clientRepo       repositories.ClientRepository
	passwordService  services.PasswordService
	auditRepo        repositories.AuditRepository
	clientCountCache services.ClientCountCache
}

// NewCreateClientUseCase creates a new CreateClientUseCase
func NewCreateClientUseCase(
	clientRepo repositories.ClientRepository,
	passwordService services.PasswordService,
	auditRepo repositories.AuditRepository,
	clientCountCache services.ClientCountCache,
) *CreateClientUseCase {
	return &CreateClientUseCase{
		clientRepo:       clientRepo,
		passwordService:  passwordService,
		auditRepo:        auditRepo,
		clientCountCache: clientCountCache,
	}
}

// Execute creates a new OAuth client
func (uc *CreateClientUseCase) Execute(ctx context.Context, req *CreateClientRequest) (*CreateClientResponse, error) {
	// Validate request
	if err := uc.validateRequest(req); err != nil {
		return nil, err
	}

	// Check quota
	count, err := uc.clientRepo.CountByTenant(ctx, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to check client quota: %w", err)
	}
	if count >= entities.MaxClientsPerTenant {
		return nil, fmt.Errorf("quota exceeded: maximum %d clients per tenant", entities.MaxClientsPerTenant)
	}

	// Validate redirect URIs (strict HTTPS enforcement)
	tempClient := &entities.Client{}
	for _, uri := range req.RedirectURIs {
		if err := tempClient.ValidateRedirectURIStrict(uri, true); err != nil {
			return nil, err
		}
	}

	// Generate client ID
	clientID, err := generateClientID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client_id: %w", err)
	}

	var clientSecret string
	var secretHash string

	// Generate secret only for confidential clients
	if req.ClientType == entities.ClientTypeConfidential {
		clientSecret, err = generateClientSecret()
		if err != nil {
			return nil, fmt.Errorf("failed to generate client_secret: %w", err)
		}

		secretHash, err = uc.passwordService.Hash(clientSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to hash client secret: %w", err)
		}
	}

	// Create client entity
	now := time.Now()
	client := &entities.Client{
		ClientID:            clientID,
		TenantID:            req.TenantID,
		ClientName:          req.ClientName,
		Description:         req.Description,
		ClientType:          req.ClientType,
		ClientSecretHash:    secretHash,
		AllowedRedirectURIs: req.RedirectURIs,
		IsActive:            true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	// Save to repository
	if err := uc.clientRepo.Create(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// Invalidate count cache
	_ = uc.clientCountCache.Invalidate(ctx, req.TenantID)

	// Create audit log
	uc.createAuditLog(ctx, req, clientID)

	return &CreateClientResponse{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		ClientName:   req.ClientName,
		Description:  req.Description,
		ClientType:   req.ClientType,
		RedirectURIs: req.RedirectURIs,
		IsActive:     true,
		CreatedAt:    now,
	}, nil
}

func (uc *CreateClientUseCase) createAuditLog(ctx context.Context, req *CreateClientRequest, clientID string) {
	auditLog := entities.NewAuditLog(entities.EventClientCreated, req.IPAddress, req.UserAgent).
		WithData("client_id", clientID).
		WithData("client_name", req.ClientName).
		WithData("client_type", req.ClientType)

	if tenantUUID, err := uuid.Parse(req.TenantID); err == nil {
		auditLog.WithTenant(tenantUUID)
	}

	_ = uc.auditRepo.Create(ctx, auditLog)
}

func (uc *CreateClientUseCase) validateRequest(req *CreateClientRequest) error {
	if req.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if req.ClientName == "" {
		return errors.New("client_name is required")
	}
	if len(req.ClientName) < 3 || len(req.ClientName) > 100 {
		return errors.New("client_name must be between 3 and 100 characters")
	}
	if len(req.Description) > 500 {
		return errors.New("description must not exceed 500 characters")
	}
	if req.ClientType != entities.ClientTypeConfidential && req.ClientType != entities.ClientTypePublic {
		return fmt.Errorf("invalid client_type: must be '%s' or '%s'", entities.ClientTypeConfidential, entities.ClientTypePublic)
	}
	if len(req.RedirectURIs) == 0 {
		return errors.New("at least one redirect_uri is required")
	}
	return nil
}

// generateClientID generates a URL-safe client ID
func generateClientID() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// generateClientSecret generates a cryptographically secure client secret
func generateClientSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
