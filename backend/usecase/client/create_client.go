package client

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
)

// CreateClientRequest contains the data for creating a new OAuth client
type CreateClientRequest struct {
	TenantID     string
	ClientName   string
	RedirectURIs []string
}

// CreateClientResponse contains the created client data including the plain-text secret
type CreateClientResponse struct {
	ClientID     string
	ClientSecret string // Only returned once at creation time
	ClientName   string
	RedirectURIs []string
	IsActive     bool
	CreatedAt    time.Time
}

// CreateClientUseCase handles the creation of OAuth clients
type CreateClientUseCase struct {
	clientRepo      repositories.ClientRepository
	passwordService services.PasswordService
}

// NewCreateClientUseCase creates a new CreateClientUseCase
func NewCreateClientUseCase(
	clientRepo repositories.ClientRepository,
	passwordService services.PasswordService,
) *CreateClientUseCase {
	return &CreateClientUseCase{
		clientRepo:      clientRepo,
		passwordService: passwordService,
	}
}

// Execute creates a new OAuth client
func (uc *CreateClientUseCase) Execute(ctx context.Context, req *CreateClientRequest) (*CreateClientResponse, error) {
	// Validate request
	if err := uc.validateRequest(req); err != nil {
		return nil, err
	}

	// Validate redirect URIs first (before generating credentials)
	tempClient := &entities.Client{}
	for _, uri := range req.RedirectURIs {
		if err := tempClient.ValidateRedirectURI(uri); err != nil {
			return nil, err
		}
	}

	// Generate client credentials
	clientID, err := generateClientID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client_id: %w", err)
	}

	clientSecret, err := generateClientSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client_secret: %w", err)
	}

	// Hash the client secret
	secretHash, err := uc.passwordService.Hash(clientSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to hash client secret: %w", err)
	}

	// Create client entity
	now := time.Now()
	client := &entities.Client{
		ClientID:            clientID,
		TenantID:            req.TenantID,
		ClientName:          req.ClientName,
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

	return &CreateClientResponse{
		ClientID:     clientID,
		ClientSecret: clientSecret, // Only returned once
		ClientName:   req.ClientName,
		RedirectURIs: req.RedirectURIs,
		IsActive:     true,
		CreatedAt:    now,
	}, nil
}

func (uc *CreateClientUseCase) validateRequest(req *CreateClientRequest) error {
	if req.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if req.ClientName == "" {
		return errors.New("client_name is required")
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
