package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
)

// Additional OAuth error codes for token endpoint
const (
	ErrUnsupportedGrantType = "unsupported_grant_type"
)

// Token TTLs
const (
	AccessTokenTTL  = 15 * time.Minute // FR-033: 15 minute access token
	RefreshTokenTTL = 7 * 24 * time.Hour // FR-035: 7 day refresh token
)

// TokenRequest represents an OAuth token exchange request
type TokenRequest struct {
	GrantType    string
	Code         string
	RedirectURI  string
	ClientID     string
	ClientSecret string
	CodeVerifier string
	RefreshToken string
	Scope        string
}

// TokenResponse represents an OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// IssueToken handles the OAuth token exchange (authorization_code grant type)
type IssueToken struct {
	authCodeRepo     repositories.AuthCodeRepository
	clientRepo       repositories.ClientRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	roleRepo         repositories.RoleRepository
	tokenService     services.TokenService
	issuer           string
}

// NewIssueToken creates a new IssueToken use case
func NewIssueToken(
	authCodeRepo repositories.AuthCodeRepository,
	clientRepo repositories.ClientRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
	roleRepo repositories.RoleRepository,
	tokenService services.TokenService,
	issuer string,
) *IssueToken {
	return &IssueToken{
		authCodeRepo:     authCodeRepo,
		clientRepo:       clientRepo,
		refreshTokenRepo: refreshTokenRepo,
		roleRepo:         roleRepo,
		tokenService:     tokenService,
		issuer:           issuer,
	}
}

// Execute processes the token exchange request
func (uc *IssueToken) Execute(ctx context.Context, req TokenRequest) (*TokenResponse, error) {
	// Validate grant_type (FR-020)
	if req.GrantType != "authorization_code" {
		return nil, &OAuthError{
			Code:        ErrUnsupportedGrantType,
			Description: "only grant_type=authorization_code is supported",
		}
	}

	// Validate request parameters (FR-020)
	if err := ValidateTokenRequest(req); err != nil {
		return nil, err
	}

	// Retrieve authorization code (FR-023)
	authCode, err := uc.authCodeRepo.Get(ctx, req.Code)
	if err != nil {
		return nil, &OAuthError{
			Code:        ErrInvalidGrant,
			Description: "authorization code is invalid or not found",
		}
	}

	// Validate authorization code is not expired (FR-023)
	if authCode.IsExpired() {
		return nil, &OAuthError{
			Code:        ErrInvalidGrant,
			Description: "authorization code has expired",
		}
	}

	// Validate authorization code is not used (FR-024)
	if authCode.IsUsed() {
		return nil, &OAuthError{
			Code:        ErrInvalidGrant,
			Description: "authorization code has already been used",
		}
	}

	// Validate client credentials (FR-025)
	_, err = uc.clientRepo.ValidateCredentials(ctx, req.ClientID, req.ClientSecret)
	if err != nil {
		return nil, &OAuthError{
			Code:        ErrInvalidClient,
			Description: "client authentication failed",
		}
	}

	// Validate client_id matches authorization code (FR-025)
	if authCode.ClientID != req.ClientID {
		return nil, &OAuthError{
			Code:        ErrInvalidGrant,
			Description: "client_id does not match the authorization code",
		}
	}

	// Validate redirect_uri matches (FR-022)
	if authCode.RedirectURI != req.RedirectURI {
		return nil, &OAuthError{
			Code:        ErrInvalidGrant,
			Description: "redirect_uri does not match the authorization request",
		}
	}

	// Validate PKCE code_verifier (FR-021)
	if err := VerifyPKCE(req.CodeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod); err != nil {
		return nil, err
	}

	// Mark authorization code as used (FR-026)
	if err := uc.authCodeRepo.MarkAsUsed(ctx, req.Code); err != nil {
		return nil, &OAuthError{
			Code:        ErrServerError,
			Description: "failed to revoke authorization code",
		}
	}

	// Fetch active roles for JWT claims (FR-022, FR-024)
	roles, err := uc.roleRepo.GetActiveRoles(ctx, authCode.UserID, authCode.ClientID)
	if err != nil {
		// Non-fatal: include empty array if roles cannot be fetched
		roles = []string{}
	}
	if roles == nil {
		roles = []string{}
	}

	// Generate tokens
	now := time.Now()
	accessTokenExp := now.Add(AccessTokenTTL)
	refreshTokenExp := now.Add(RefreshTokenTTL)

	// Create ID token claims (FR-030, FR-031)
	idTokenClaims := &services.TokenClaims{
		Issuer:    uc.issuer,
		Subject:   authCode.UserID,
		Audience:  []string{authCode.ClientID},
		ExpiresAt: accessTokenExp,
		IssuedAt:  now,
		NotBefore: now,
		TenantID:  authCode.TenantID,
		ClientID:  authCode.ClientID,
		Scope:     authCode.Scope,
		Roles:     roles,
	}

	// Sign ID token (FR-036)
	idToken, err := uc.tokenService.SignIDToken(ctx, idTokenClaims)
	if err != nil {
		return nil, &OAuthError{
			Code:        ErrServerError,
			Description: "failed to generate ID token",
		}
	}

	// Create access token claims (FR-032)
	accessTokenClaims := &services.TokenClaims{
		Issuer:    uc.issuer,
		Subject:   authCode.UserID,
		Audience:  []string{authCode.ClientID},
		ExpiresAt: accessTokenExp,
		IssuedAt:  now,
		NotBefore: now,
		TenantID:  authCode.TenantID,
		ClientID:  authCode.ClientID,
		Scope:     authCode.Scope,
		Roles:     roles,
	}

	// Sign access token (FR-036)
	accessToken, err := uc.tokenService.SignAccessToken(ctx, accessTokenClaims)
	if err != nil {
		return nil, &OAuthError{
			Code:        ErrServerError,
			Description: "failed to generate access token",
		}
	}

	// Generate refresh token (FR-034)
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, &OAuthError{
			Code:        ErrServerError,
			Description: "failed to generate refresh token",
		}
	}

	// Hash refresh token for storage
	refreshTokenHash := hashRefreshToken(refreshToken)

	// Store refresh token (FR-035)
	refreshTokenEntity := &entities.RefreshToken{
		Token:     refreshTokenHash,
		UserID:    authCode.UserID,
		ClientID:  authCode.ClientID,
		TenantID:  authCode.TenantID,
		Scope:     authCode.Scope,
		ExpiresAt: refreshTokenExp,
		CreatedAt: now,
	}

	if err := uc.refreshTokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, &OAuthError{
			Code:        ErrServerError,
			Description: "failed to store refresh token",
		}
	}

	// Return token response (FR-027)
	return &TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(AccessTokenTTL.Seconds()),
		RefreshToken: refreshToken,
		IDToken:      idToken,
		Scope:        authCode.Scope,
	}, nil
}

// ValidateTokenRequest validates the token exchange request parameters
func ValidateTokenRequest(req TokenRequest) error {
	if req.Code == "" {
		return &OAuthError{
			Code:        ErrInvalidRequest,
			Description: "code is required",
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
	if req.RedirectURI == "" {
		return &OAuthError{
			Code:        ErrInvalidRequest,
			Description: "redirect_uri is required",
		}
	}
	if req.CodeVerifier == "" {
		return &OAuthError{
			Code:        ErrInvalidRequest,
			Description: "code_verifier is required (PKCE is mandatory)",
		}
	}
	return nil
}

// VerifyPKCE validates the PKCE code verifier against the stored code challenge
// Implements S256 verification per RFC 7636
func VerifyPKCE(codeVerifier, codeChallenge, codeChallengeMethod string) error {
	if codeChallengeMethod != "S256" {
		return &OAuthError{
			Code:        ErrInvalidRequest,
			Description: "only code_challenge_method=S256 is supported",
		}
	}

	// S256: code_challenge = BASE64URL(SHA256(code_verifier))
	hash := sha256.Sum256([]byte(codeVerifier))
	calculatedChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	if calculatedChallenge != codeChallenge {
		return &OAuthError{
			Code:        ErrInvalidGrant,
			Description: "code_verifier does not match code_challenge",
		}
	}

	return nil
}

// generateRefreshToken generates a cryptographically secure refresh token
func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// hashRefreshToken hashes a refresh token for storage
func hashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
