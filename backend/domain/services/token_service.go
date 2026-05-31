package services

import (
	"context"
	"crypto/rsa"
	"time"
)

// JWK represents a JSON Web Key (for JWKS endpoint)
type JWK struct {
	KeyType   string `json:"kty"`
	Use       string `json:"use"`
	KeyID     string `json:"kid"`
	Algorithm string `json:"alg"`
	N         string `json:"n"` // RSA modulus (base64url)
	E         string `json:"e"` // RSA exponent (base64url)
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// TokenClaims represents JWT claims shared by OAuth access tokens and ID tokens.
type TokenClaims struct {
	Issuer    string    `json:"iss"`
	Subject   string    `json:"sub"`
	Audience  []string  `json:"aud"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
	NotBefore time.Time `json:"nbf,omitempty"`
	JWTID     string    `json:"jti,omitempty"`

	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	Name          string `json:"name,omitempty"`
	GivenName     string `json:"given_name,omitempty"`
	FamilyName    string `json:"family_name,omitempty"`

	TenantID string   `json:"tenant_id,omitempty"`
	ClientID string   `json:"client_id,omitempty"`
	Scope    string   `json:"scope,omitempty"`
	Roles    []string `json:"roles"`
}

// TokenService defines the interface for JWT token operations
// Handles RS256 signing and verification
type TokenService interface {
	// SignIDToken creates a signed OIDC ID token (JWT)
	SignIDToken(ctx context.Context, claims *TokenClaims) (string, error)

	// SignAccessToken creates a signed OAuth 2.0 access token (JWT)
	SignAccessToken(ctx context.Context, claims *TokenClaims) (string, error)

	// ValidateTokenSignature validates a JWT signature and returns claims
	ValidateTokenSignature(ctx context.Context, token string) (*TokenClaims, error)

	// GetPublicKey retrieves the public key for a specific key ID
	GetPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error)

	// GetJWKS generates the JWKS (JSON Web Key Set) for the /.well-known/jwks.json endpoint
	GetJWKS(ctx context.Context) (*JWKS, error)

	// GetActiveKeyID returns the key ID of the currently active signing key
	GetActiveKeyID(ctx context.Context) (string, error)
}
