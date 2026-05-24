package services

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

// FositeOAuthProvider implements OAuthProvider using custom OAuth logic
// TODO: Replace with full fosite integration when storage adapters are complete
type FositeOAuthProvider struct {
	clientRepo      repositories.ClientRepository
	authCodeRepo    repositories.AuthCodeRepository
	refreshRepo     repositories.RefreshTokenRepository
	tokenService    services.TokenService
	issuer          string
	authCodeTTL     time.Duration
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// FositeConfig holds configuration for the OAuth provider
type FositeConfig struct {
	Issuer           string
	AuthCodeTTL      time.Duration
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	ClientRepo       repositories.ClientRepository
	AuthCodeRepo     repositories.AuthCodeRepository
	RefreshTokenRepo repositories.RefreshTokenRepository
	TokenService     services.TokenService
}

// NewFositeOAuthProvider creates a new OAuth provider
func NewFositeOAuthProvider(cfg *FositeConfig) (services.OAuthProvider, error) {
	return &FositeOAuthProvider{
		clientRepo:      cfg.ClientRepo,
		authCodeRepo:    cfg.AuthCodeRepo,
		refreshRepo:     cfg.RefreshTokenRepo,
		tokenService:    cfg.TokenService,
		issuer:          cfg.Issuer,
		authCodeTTL:     cfg.AuthCodeTTL,
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
	}, nil
}

// GenerateAuthCode creates a new authorization code
func (p *FositeOAuthProvider) GenerateAuthCode(ctx context.Context, req *services.AuthorizationRequest) (string, error) {
	// Validate client
	client, err := p.clientRepo.GetByClientID(ctx, req.ClientID, req.TenantID)
	if err != nil {
		return "", fmt.Errorf("invalid client: %w", err)
	}

	// Validate redirect URI
	if !client.IsURIAllowed(req.RedirectURI) {
		return "", fmt.Errorf("invalid redirect_uri")
	}

	// Validate PKCE challenge
	if req.CodeChallenge == "" || req.CodeChallengeMethod != "S256" {
		return "", fmt.Errorf("PKCE required: code_challenge and code_challenge_method=S256 must be provided")
	}

	// Generate authorization code
	code, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate code: %w", err)
	}

	// Store authorization code
	authCode := &entities.AuthorizationCode{
		Code:                code,
		ClientID:            req.ClientID,
		TenantID:            req.TenantID,
		UserID:              req.UserID,
		RedirectURI:         req.RedirectURI,
		Scope:               req.Scope,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		ExpiresAt:           time.Now().Add(p.authCodeTTL),
		CreatedAt:           time.Now(),
		UsedFlag:            false,
	}

	if err := p.authCodeRepo.Store(ctx, authCode, p.authCodeTTL); err != nil {
		return "", fmt.Errorf("failed to store authorization code: %w", err)
	}

	return code, nil
}

// ExchangeCodeForTokens exchanges an authorization code for tokens
func (p *FositeOAuthProvider) ExchangeCodeForTokens(ctx context.Context, req *services.TokenRequest) (*services.TokenResponse, error) {
	// Get and validate authorization code
	authCode, err := p.authCodeRepo.Get(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization code")
	}

	if !authCode.IsValid() {
		return nil, fmt.Errorf("authorization code expired or already used")
	}

	// Validate PKCE verifier
	if !validatePKCE(authCode.CodeChallenge, authCode.CodeChallengeMethod, req.CodeVerifier) {
		return nil, fmt.Errorf("invalid code_verifier")
	}

	// Mark code as used
	if err := p.authCodeRepo.MarkAsUsed(ctx, req.Code); err != nil {
		return nil, fmt.Errorf("failed to mark code as used: %w", err)
	}

	// Generate access token
	accessToken, err := p.tokenService.SignAccessToken(ctx, &services.TokenClaims{
		Subject:   authCode.UserID,
		ClientID:  authCode.ClientID,
		TenantID:  authCode.TenantID,
		Scope:     authCode.Scope,
		Issuer:    p.issuer,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(p.accessTokenTTL),
		Audience:  []string{authCode.ClientID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshTokenValue, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	refreshToken := &entities.RefreshToken{
		Token:       refreshTokenValue,
		UserID:      authCode.UserID,
		ClientID:    authCode.ClientID,
		TenantID:    authCode.TenantID,
		Scope:       authCode.Scope,
		ExpiresAt:   time.Now().Add(p.refreshTokenTTL),
		CreatedAt:   time.Now(),
		RevokedFlag: false,
	}

	if err := p.refreshRepo.Create(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &services.TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(p.accessTokenTTL.Seconds()),
		RefreshToken: refreshTokenValue,
		Scope:        authCode.Scope,
	}, nil
}

// RefreshAccessToken generates a new access token using a refresh token
func (p *FositeOAuthProvider) RefreshAccessToken(ctx context.Context, refreshTokenValue string, clientID string, clientSecret string) (*services.TokenResponse, error) {
	// Get refresh token
	refreshToken, err := p.refreshRepo.GetByToken(ctx, refreshTokenValue)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	if !refreshToken.IsValid() {
		return nil, fmt.Errorf("refresh token expired or revoked")
	}

	// Validate client
	_, err = p.clientRepo.ValidateCredentials(ctx, clientID, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid client credentials")
	}

	// Update last used timestamp
	refreshToken.MarkUsed()
	if err := p.refreshRepo.UpdateLastUsed(ctx, refreshTokenValue); err != nil {
		return nil, fmt.Errorf("failed to update refresh token last used: %w", err)
	}

	// Generate new access token
	accessToken, err := p.tokenService.SignAccessToken(ctx, &services.TokenClaims{
		Subject:   refreshToken.UserID,
		ClientID:  refreshToken.ClientID,
		TenantID:  refreshToken.TenantID,
		Scope:     refreshToken.Scope,
		Issuer:    p.issuer,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(p.accessTokenTTL),
		Audience:  []string{refreshToken.ClientID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &services.TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(p.accessTokenTTL.Seconds()),
		RefreshToken: refreshTokenValue,
		Scope:        refreshToken.Scope,
	}, nil
}

// RevokeToken revokes a refresh token
func (p *FositeOAuthProvider) RevokeToken(ctx context.Context, token string, tokenTypeHint string, clientID string, clientSecret string) error {
	refreshToken, err := p.refreshRepo.GetByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("token not found")
	}

	return p.refreshRepo.Revoke(ctx, refreshToken.Token, "user_requested")
}

// ValidateToken validates an access token
func (p *FositeOAuthProvider) ValidateToken(ctx context.Context, token string) (*services.IntrospectionResponse, error) {
	claims, err := p.tokenService.ValidateTokenSignature(ctx, token)
	if err != nil {
		return &services.IntrospectionResponse{
			Active: false,
		}, nil
	}

	return &services.IntrospectionResponse{
		Active:   true,
		ClientID: claims.ClientID,
		Username: claims.Subject,
		Scope:    claims.Scope,
		Exp:      claims.ExpiresAt.Unix(),
		Iat:      claims.IssuedAt.Unix(),
		Sub:      claims.Subject,
		Iss:      claims.Issuer,
	}, nil
}

// GetAuthorizationURL constructs the authorization URL for client redirect
func (p *FositeOAuthProvider) GetAuthorizationURL(req *services.AuthorizationRequest) string {
	// TODO: Implement proper URL construction
	return fmt.Sprintf("%s/oauth/authorize", p.issuer)
}

// generateToken generates a cryptographically secure random token
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// validatePKCE validates the PKCE code verifier against the challenge
func validatePKCE(challenge, method, verifier string) bool {
	if challenge == "" || verifier == "" || method != "S256" {
		return false
	}

	sum := sha256.Sum256([]byte(verifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(sum[:])
	return expectedChallenge == challenge
}
