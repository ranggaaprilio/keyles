package services

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
)

// RSATokenService implements TokenService using RS256 algorithm
type RSATokenService struct {
	keyRepo repositories.SigningKeyRepository
}

// NewRSATokenService creates a new RSA token service
func NewRSATokenService(keyRepo repositories.SigningKeyRepository) services.TokenService {
	return &RSATokenService{keyRepo: keyRepo}
}

// SignIDToken creates a signed OIDC ID token (JWT)
func (s *RSATokenService) SignIDToken(ctx context.Context, claims *services.TokenClaims) (string, error) {
	return s.signToken(ctx, claims)
}

// SignAccessToken creates a signed OAuth 2.0 access token (JWT)
func (s *RSATokenService) SignAccessToken(ctx context.Context, claims *services.TokenClaims) (string, error) {
	return s.signToken(ctx, claims)
}

// signToken signs a JWT token using the active RSA key
func (s *RSATokenService) signToken(ctx context.Context, claims *services.TokenClaims) (string, error) {
	// Get the active signing key
	key, err := s.keyRepo.GetActive(ctx, "RS256")
	if err != nil {
		return "", fmt.Errorf("failed to get active signing key: %w", err)
	}

	if key == nil {
		return "", errors.New("no active signing key found")
	}

	// Parse the private key from PEM format
	privateKey, err := s.parsePrivateKey(key.PrivateKeyEncrypted)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":            claims.Issuer,
		"sub":            claims.Subject,
		"aud":            claims.Audience,
		"exp":            claims.ExpiresAt.Unix(),
		"iat":            claims.IssuedAt.Unix(),
		"nbf":            claims.NotBefore.Unix(),
		"jti":            claims.JWTID,
		"email":          claims.Email,
		"email_verified": claims.EmailVerified,
		"name":           claims.Name,
		"given_name":     claims.GivenName,
		"family_name":    claims.FamilyName,
		"tenant_id":      claims.TenantID,
		"client_id":      claims.ClientID,
		"scope":          claims.Scope,
		"roles":          claims.Roles,
	})
	// Set OIDC claims conditionally on non-zero values to avoid cluttering
	// access tokens with unused fields.
	if claims.Nonce != "" {
		token.Claims.(jwt.MapClaims)["nonce"] = claims.Nonce
	}
	if claims.AuthTime != 0 {
		token.Claims.(jwt.MapClaims)["auth_time"] = claims.AuthTime
	}

	// Set the key ID in the header
	token.Header["kid"] = key.KeyID

	// Sign the token
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateTokenSignature validates a JWT signature and returns claims
func (s *RSATokenService) ValidateTokenSignature(ctx context.Context, tokenString string) (*services.TokenClaims, error) {
	// Parse the token to get the key ID
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if token.Method != jwt.SigningMethodRS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the key ID from the header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing key ID in token header")
		}

		// Get the public key for this key ID
		publicKey, err := s.GetPublicKey(ctx, kid)
		if err != nil {
			return nil, err
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("failed to extract claims")
	}
	// Convert to TokenClaims
	tokenClaims := &services.TokenClaims{
		Issuer:        getString(claims, "iss"),
		Subject:       getString(claims, "sub"),
		Audience:      getStringArray(claims, "aud"),
		ExpiresAt:     getTime(claims, "exp"),
		IssuedAt:      getTime(claims, "iat"),
		NotBefore:     getTime(claims, "nbf"),
		JWTID:         getString(claims, "jti"),
		Email:         getString(claims, "email"),
		EmailVerified: getBool(claims, "email_verified"),
		Name:          getString(claims, "name"),
		GivenName:     getString(claims, "given_name"),
		FamilyName:    getString(claims, "family_name"),
		TenantID:      getString(claims, "tenant_id"),
		ClientID:      getString(claims, "client_id"),
		Scope:         getString(claims, "scope"),
		Roles:         getStringArray(claims, "roles"),
		Nonce:         getString(claims, "nonce"),
		AuthTime:      getInt64(claims, "auth_time"),
	}

	return tokenClaims, nil
}

// GetPublicKey retrieves the public key for a specific key ID
func (s *RSATokenService) GetPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	key, err := s.keyRepo.GetByKeyID(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get signing key: %w", err)
	}

	if key == nil {
		return nil, errors.New("signing key not found")
	}

	return s.parsePublicKey(key.PublicKey)
}

// GetJWKS generates the JWKS (JSON Web Key Set) for the /.well-known/jwks.json endpoint
func (s *RSATokenService) GetJWKS(ctx context.Context) (*services.JWKS, error) {
	// Get all active keys
	keys, err := s.keyRepo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list active keys: %w", err)
	}

	jwks := &services.JWKS{
		Keys: make([]services.JWK, 0, len(keys)),
	}

	for _, key := range keys {
		// Parse public key
		publicKey, err := s.parsePublicKey(key.PublicKey)
		if err != nil {
			continue // Skip invalid keys
		}

		// Convert to JWK format
		jwk := services.JWK{
			KeyType:   "RSA",
			Use:       "sig",
			KeyID:     key.KeyID,
			Algorithm: key.Algorithm,
			N:         base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
			E:         base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes()),
		}

		jwks.Keys = append(jwks.Keys, jwk)
	}

	return jwks, nil
}

// GetActiveKeyID returns the key ID of the currently active signing key
func (s *RSATokenService) GetActiveKeyID(ctx context.Context) (string, error) {
	key, err := s.keyRepo.GetActive(ctx, "RS256")
	if err != nil {
		return "", fmt.Errorf("failed to get active signing key: %w", err)
	}

	if key == nil {
		return "", errors.New("no active signing key found")
	}

	return key.KeyID, nil
}

// parsePrivateKey parses a PEM-encoded RSA private key
func (s *RSATokenService) parsePrivateKey(pemKey string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		privateKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not an RSA private key")
		}
		return privateKey, nil
	}

	return privateKey, nil
}

// parsePublicKey parses a PEM-encoded RSA public key
func (s *RSATokenService) parsePublicKey(pemKey string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		// Try PKIX format
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		publicKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("not an RSA public key")
		}
		return publicKey, nil
	}

	return publicKey, nil
}

// Helper functions for claim extraction
func getString(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}
	return ""
}

func getStringArray(claims jwt.MapClaims, key string) []string {
	if val, ok := claims[key].([]interface{}); ok {
		result := make([]string, 0, len(val))
		for _, v := range val {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	if val, ok := claims[key].(string); ok {
		return []string{val}
	}
	return nil
}

func getBool(claims jwt.MapClaims, key string) bool {
	if val, ok := claims[key].(bool); ok {
		return val
	}
	return false
}

func getTime(claims jwt.MapClaims, key string) time.Time {
	if val, ok := claims[key].(float64); ok {
		return time.Unix(int64(val), 0)
	}
	return time.Time{}
}

func getInt64(claims jwt.MapClaims, key string) int64 {
	if val, ok := claims[key].(float64); ok {
		return int64(val)
	}
	return 0
}
