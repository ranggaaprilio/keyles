package auth

import (
	"context"
	"errors"
	"time"

	"github.com/ranggaaprilio/keyles/domain/services"
)

// ValidateTokenInput represents input for token validation
type ValidateTokenInput struct {
	Token            string
	ExpectedTenantID string
	ExpectedAudience string
}

// ValidateTokenOutput represents the output of token validation
type ValidateTokenOutput struct {
	Valid  bool
	Claims *services.TokenClaims
	Error  error
}

// ValidateToken is the use case for validating JWT tokens
type ValidateToken struct {
	tokenService services.TokenService
}

// NewValidateToken creates a new ValidateToken use case
func NewValidateToken(tokenService services.TokenService) *ValidateToken {
	return &ValidateToken{
		tokenService: tokenService,
	}
}

// Execute validates a JWT token
// - Validates signature using TokenService (FR-039)
// - Validates expiration
// - Validates tenant_id if provided
// - Validates audience if provided
func (uc *ValidateToken) Execute(ctx context.Context, input ValidateTokenInput) (*ValidateTokenOutput, error) {
	// Validate signature and parse claims
	claims, err := uc.tokenService.ValidateTokenSignature(ctx, input.Token)
	if err != nil {
		return &ValidateTokenOutput{
			Valid: false,
			Error: errors.New("invalid token signature"),
		}, nil
	}

	// Check expiration
	if time.Now().After(claims.ExpiresAt) {
		return &ValidateTokenOutput{
			Valid: false,
			Error: errors.New("token expired"),
		}, nil
	}

	// Validate tenant_id if provided
	if input.ExpectedTenantID != "" && claims.TenantID != input.ExpectedTenantID {
		return &ValidateTokenOutput{
			Valid: false,
			Error: errors.New("tenant_id mismatch"),
		}, nil
	}

	// Validate audience if provided
	if input.ExpectedAudience != "" {
		found := false
		for _, aud := range claims.Audience {
			if aud == input.ExpectedAudience {
				found = true
				break
			}
		}
		if !found {
			return &ValidateTokenOutput{
				Valid: false,
				Error: errors.New("audience mismatch"),
			}, nil
		}
	}

	return &ValidateTokenOutput{
		Valid:  true,
		Claims: claims,
		Error:  nil,
	}, nil
}
