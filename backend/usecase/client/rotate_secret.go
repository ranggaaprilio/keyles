package client

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
)

// RotateSecretRequest contains the data for rotating a client secret
type RotateSecretRequest struct {
	ClientID string
	TenantID string
}

// RotateSecretResponse contains the new client secret
type RotateSecretResponse struct {
	ClientID        string
	ClientSecret    string // New secret - only returned once
	RotatedAt       time.Time
}

// RotateSecretUseCase handles rotation of client secrets
type RotateSecretUseCase struct {
	clientRepo      repositories.ClientRepository
	passwordService services.PasswordService
}

// NewRotateSecretUseCase creates a new RotateSecretUseCase
func NewRotateSecretUseCase(
	clientRepo repositories.ClientRepository,
	passwordService services.PasswordService,
) *RotateSecretUseCase {
	return &RotateSecretUseCase{
		clientRepo:      clientRepo,
		passwordService: passwordService,
	}
}

// Execute rotates the client secret
func (uc *RotateSecretUseCase) Execute(ctx context.Context, req *RotateSecretRequest) (*RotateSecretResponse, error) {
	// Validate request
	if req.ClientID == "" {
		return nil, errors.New("client_id is required")
	}
	if req.TenantID == "" {
		return nil, errors.New("tenant_id is required")
	}

	// Get existing client (verify ownership)
	client, err := uc.clientRepo.GetByClientID(ctx, req.ClientID, req.TenantID)
	if err != nil {
		return nil, err
	}

	// Generate new secret
	newSecret, err := generateNewSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new secret: %w", err)
	}

	// Hash the new secret
	secretHash, err := uc.passwordService.Hash(newSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to hash new secret: %w", err)
	}

	// Update client with new secret hash
	now := time.Now()
	client.ClientSecretHash = secretHash
	client.UpdatedAt = now

	if err := uc.clientRepo.Update(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to update client secret: %w", err)
	}

	return &RotateSecretResponse{
		ClientID:     req.ClientID,
		ClientSecret: newSecret, // Only returned once
		RotatedAt:    now,
	}, nil
}

// generateNewSecret generates a cryptographically secure new secret
func generateNewSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
