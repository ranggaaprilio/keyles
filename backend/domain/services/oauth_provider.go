package services

import (
	"context"
	"time"
)

// AuthorizationRequest represents an OAuth 2.0 authorization request
type AuthorizationRequest struct {
	ClientID            string
	RedirectURI         string
	ResponseType        string
	Scope               string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
	UserID              string
	TenantID            string
}

// TokenRequest represents an OAuth 2.0 token request
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

// TokenResponse represents an OAuth 2.0 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// IntrospectionResponse represents token introspection response
type IntrospectionResponse struct {
	Active    bool   `json:"active"`
	Scope     string `json:"scope,omitempty"`
	ClientID  string `json:"client_id,omitempty"`
	Username  string `json:"username,omitempty"`
	TokenType string `json:"token_type,omitempty"`
	Exp       int64  `json:"exp,omitempty"`
	Iat       int64  `json:"iat,omitempty"`
	Sub       string `json:"sub,omitempty"`
	Aud       string `json:"aud,omitempty"`
	Iss       string `json:"iss,omitempty"`
}

// OAuthProvider defines the interface for OAuth 2.0 / OIDC operations
// This is typically implemented by a wrapper around github.com/ory/fosite
type OAuthProvider interface {
	// GenerateAuthCode generates an authorization code for the authorization flow
	GenerateAuthCode(ctx context.Context, req *AuthorizationRequest) (string, error)

	// ExchangeCodeForTokens exchanges an authorization code for tokens
	ExchangeCodeForTokens(ctx context.Context, req *TokenRequest) (*TokenResponse, error)

	// RefreshAccessToken refreshes an access token using a refresh token
	RefreshAccessToken(ctx context.Context, refreshToken string, clientID string, clientSecret string) (*TokenResponse, error)

	// ValidateToken validates an access token and returns its claims
	ValidateToken(ctx context.Context, token string) (*IntrospectionResponse, error)

	// RevokeToken revokes an access or refresh token
	RevokeToken(ctx context.Context, token string, tokenTypeHint string, clientID string, clientSecret string) error

	// GetAuthorizationURL constructs the authorization URL for client redirect
	GetAuthorizationURL(req *AuthorizationRequest) string
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	// Standard claims
	Issuer    string    `json:"iss"`
	Subject   string    `json:"sub"`
	Audience  []string  `json:"aud"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
	NotBefore time.Time `json:"nbf,omitempty"`
	JWTID     string    `json:"jti,omitempty"`

	// OIDC claims
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	Name          string `json:"name,omitempty"`
	GivenName     string `json:"given_name,omitempty"`
	FamilyName    string `json:"family_name,omitempty"`

	// Custom claims
	TenantID string `json:"tenant_id,omitempty"`
	ClientID string `json:"client_id,omitempty"`
	Scope    string `json:"scope,omitempty"`

	// RBAC claims (feature 005)
	Roles []string `json:"roles,omitempty"`
}
