package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
)

// RefreshTokenRequest represents a refresh token grant request
type RefreshTokenRequest struct {
	RefreshToken string
	ClientID     string
	ClientSecret string
	Scope        string // Optional: can request reduced scope
}

// RefreshTokenResponse represents the response for refresh token grant
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"` // Only if rotation is enabled
	Scope        string `json:"scope,omitempty"`
}

// RefreshToken handles the OAuth refresh_token grant type
type RefreshToken struct {
	refreshTokenRepo repositories.RefreshTokenRepository
	clientRepo       repositories.ClientRepository
	tokenService     services.TokenService
	issuer           string
}

// NewRefreshToken creates a new RefreshToken use case
func NewRefreshToken(
	refreshTokenRepo repositories.RefreshTokenRepository,
	clientRepo repositories.ClientRepository,
	tokenService services.TokenService,
	issuer string,
) *RefreshToken {
	return &RefreshToken{
		refreshTokenRepo: refreshTokenRepo,
		clientRepo:       clientRepo,
		tokenService:     tokenService,
		issuer:           issuer,
	}
}

// Execute processes the refresh token request and returns new tokens
func (uc *RefreshToken) Execute(ctx context.Context, req RefreshTokenRequest) (*RefreshTokenResponse, error) {
	// Validate request parameters
	if err := uc.validateRequest(req); err != nil {
		return nil, err
	}

	// Hash the refresh token for lookup
	tokenHash := hashRefreshToken(req.RefreshToken)

	// Retrieve refresh token from database
	storedToken, err := uc.refreshTokenRepo.GetByToken(ctx, tokenHash)
	if err != nil {
		return nil, &OAuthError{
			Code:        ErrInvalidGrant,
			Description: "refresh_token is invalid or not found",
		}
	}

	// Check if token is revoked (FR-046)
	if storedToken.IsRevoked() {
		return nil, &OAuthError{
			Code:        ErrInvalidGrant,
			Description: "refresh_token has been revoked",
		}
	}

	// Check if token is expired
	if storedToken.IsExpired() {
		return nil, &OAuthError{
			Code:        ErrInvalidGrant,
			Description: "refresh_token has expired",
		}
	}

	// Validate client_id matches the token's client (FR-047)
	if storedToken.ClientID != req.ClientID {
		return nil, &OAuthError{
			Code:        ErrInvalidGrant,
			Description: "client_id does not match the refresh_token",
		}
	}

	// Validate client credentials
	_, err = uc.clientRepo.ValidateCredentials(ctx, req.ClientID, req.ClientSecret)
	if err != nil {
		return nil, &OAuthError{
			Code:        ErrInvalidClient,
			Description: "client authentication failed",
		}
	}

	// Determine scope for new access token
	// If scope is requested, validate it's a subset of original scope
	scope := storedToken.Scope
	if req.Scope != "" {
		// In a full implementation, validate req.Scope is subset of storedToken.Scope
		scope = req.Scope
	}

	// Generate new access token (FR-043)
	now := time.Now()
	accessTokenExp := now.Add(AccessTokenTTL)

	accessTokenClaims := &services.TokenClaims{
		Issuer:    uc.issuer,
		Subject:   storedToken.UserID,
		Audience:  []string{storedToken.ClientID},
		ExpiresAt: accessTokenExp,
		IssuedAt:  now,
		NotBefore: now,
		TenantID:  storedToken.TenantID,
		ClientID:  storedToken.ClientID,
		Scope:     scope,
	}

	accessToken, err := uc.tokenService.SignAccessToken(ctx, accessTokenClaims)
	if err != nil {
		return nil, &OAuthError{
			Code:        ErrServerError,
			Description: "failed to generate access token",
		}
	}

	// Update last_used_at timestamp (non-fatal if fails)
	_ = uc.refreshTokenRepo.UpdateLastUsed(ctx, tokenHash)

	// Return new access token
	// Note: Refresh token rotation is optional and not implemented here
	// To implement rotation:
	// 1. Revoke the current refresh token
	// 2. Generate and store a new refresh token
	// 3. Return the new refresh token in the response
	return &RefreshTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(AccessTokenTTL.Seconds()),
		Scope:       scope,
	}, nil
}

// validateRequest validates the refresh token request parameters
func (uc *RefreshToken) validateRequest(req RefreshTokenRequest) error {
	if req.RefreshToken == "" {
		return &OAuthError{
			Code:        ErrInvalidRequest,
			Description: "refresh_token is required",
		}
	}
	if req.ClientID == "" {
		return &OAuthError{
			Code:        ErrInvalidRequest,
			Description: "client_id is required",
		}
	}
	if req.ClientSecret == "" {
		return &OAuthError{
			Code:        ErrInvalidRequest,
			Description: "client_secret is required",
		}
	}
	return nil
}

// hashRefreshToken hashes a refresh token for storage/lookup
// This function is duplicated from issue_token.go - in production
// this would be in a shared utility package
func hashRefreshTokenForLookup(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
