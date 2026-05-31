package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
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
	roleRepo         repositories.RoleRepository
	endUserRepo      repositories.EndUserRepository
	issuer           string
}

// NewRefreshToken creates a new RefreshToken use case
func NewRefreshToken(
	refreshTokenRepo repositories.RefreshTokenRepository,
	clientRepo repositories.ClientRepository,
	tokenService services.TokenService,
	issuer string,
	dependencies ...interface{},
) *RefreshToken {
	var roleRepo repositories.RoleRepository
	var endUserRepo repositories.EndUserRepository
	for _, dependency := range dependencies {
		switch typed := dependency.(type) {
		case repositories.RoleRepository:
			roleRepo = typed
		case repositories.EndUserRepository:
			endUserRepo = typed
		}
	}
	return &RefreshToken{
		refreshTokenRepo: refreshTokenRepo,
		clientRepo:       clientRepo,
		tokenService:     tokenService,
		roleRepo:         roleRepo,
		endUserRepo:      endUserRepo,
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
		if rotator, ok := uc.refreshTokenRepo.(repositories.RefreshTokenRotationRepository); ok && storedToken.FamilyID != "" {
			_ = rotator.RevokeFamily(ctx, storedToken.FamilyID, "refresh token replay detected")
		}
		return nil, &OAuthError{
			Code:        ErrInvalidGrant,
			Description: "refresh_token has been revoked or replayed",
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
	_, err = authenticateOAuthClient(ctx, uc.clientRepo, req.ClientID, req.ClientSecret, true)
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
		if !isScopeSubset(req.Scope, storedToken.Scope) {
			return nil, &OAuthError{Code: ErrInvalidScope, Description: "requested scope exceeds the originally granted scope"}
		}
		scope = req.Scope
	}

	if uc.endUserRepo != nil {
		user, err := uc.endUserRepo.GetByID(ctx, storedToken.UserID)
		if err != nil || user == nil || user.TenantID != storedToken.TenantID || user.Status == entities.UserStatusDisabled {
			return nil, &OAuthError{Code: ErrInvalidGrant, Description: "user account is not eligible for token refresh"}
		}
	}

	roles := []string{}
	if uc.roleRepo != nil {
		roles, err = uc.roleRepo.GetActiveRoles(ctx, storedToken.UserID, storedToken.ClientID)
		if err != nil {
			return nil, &OAuthError{Code: ErrServerError, Description: "failed to load user roles"}
		}
		if len(roles) == 0 {
			return nil, &OAuthError{Code: ErrInvalidGrant, Description: "user no longer has access to this client"}
		}
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
		Roles:     roles,
	}

	accessToken, err := uc.tokenService.SignAccessToken(ctx, accessTokenClaims)
	if err != nil {
		return nil, &OAuthError{
			Code:        ErrServerError,
			Description: "failed to generate access token",
		}
	}

	replacementValue, err := generateRefreshToken()
	if err != nil {
		return nil, &OAuthError{Code: ErrServerError, Description: "failed to generate refresh token"}
	}
	replacementHash := hashRefreshToken(replacementValue)
	replacement := &entities.RefreshToken{
		Token:           replacementHash,
		UserID:          storedToken.UserID,
		ClientID:        storedToken.ClientID,
		TenantID:        storedToken.TenantID,
		Scope:           scope,
		ExpiresAt:       time.Now().Add(RefreshTokenTTL),
		CreatedAt:       time.Now(),
		FamilyID:        storedToken.FamilyID,
		ParentTokenHash: tokenHash,
	}
	if replacement.FamilyID == "" {
		replacement.FamilyID = tokenHash
	}

	if rotator, ok := uc.refreshTokenRepo.(repositories.RefreshTokenRotationRepository); ok {
		if err := rotator.Rotate(ctx, tokenHash, replacement); err != nil {
			if errors.Is(err, repositories.ErrRefreshTokenReplay) {
				return nil, &OAuthError{Code: ErrInvalidGrant, Description: "refresh_token replay detected"}
			}
			return nil, &OAuthError{Code: ErrServerError, Description: "failed to rotate refresh_token"}
		}
	} else {
		// Compatibility path for test doubles and adapters that have not opted
		// into transactional rotation. Production PostgreSQL uses Rotate above.
		_ = uc.refreshTokenRepo.UpdateLastUsed(ctx, tokenHash)
		replacementValue = ""
	}

	return &RefreshTokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(AccessTokenTTL.Seconds()),
		RefreshToken: replacementValue,
		Scope:        scope,
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
	return nil
}

// hashRefreshToken hashes a refresh token for storage/lookup
// This function is duplicated from issue_token.go - in production
// this would be in a shared utility package
func hashRefreshTokenForLookup(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
