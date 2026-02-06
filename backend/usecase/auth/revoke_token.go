package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"

	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// RevokeTokenRequest represents a token revocation request per RFC 7009
type RevokeTokenRequest struct {
	Token         string // The token to revoke
	TokenTypeHint string // Optional hint: "refresh_token" or "access_token"
}

// RevokeToken handles the OAuth token revocation endpoint (RFC 7009)
type RevokeToken struct {
	refreshTokenRepo repositories.RefreshTokenRepository
}

// NewRevokeToken creates a new RevokeToken use case
func NewRevokeToken(
	refreshTokenRepo repositories.RefreshTokenRepository,
) *RevokeToken {
	return &RevokeToken{
		refreshTokenRepo: refreshTokenRepo,
	}
}

// Execute processes the token revocation request
// Per RFC 7009, the server responds with HTTP 200 OK regardless of whether
// the token was valid or not (to prevent token enumeration attacks)
func (uc *RevokeToken) Execute(ctx context.Context, req RevokeTokenRequest) error {
	// Validate request parameters
	if req.Token == "" {
		return &OAuthError{
			Code:        ErrInvalidRequest,
			Description: "token parameter is required",
		}
	}

	// If hint is access_token, we can't revoke JWTs (they're stateless)
	// Per RFC 7009, we should return 200 OK anyway
	if req.TokenTypeHint == "access_token" {
		// Access tokens are JWTs and not stored in database
		// They will expire naturally based on their exp claim
		// Return success without doing anything
		return nil
	}

	// Try to revoke as refresh token (default behavior)
	tokenHash := hashTokenForRevocation(req.Token)

	// Try to find the token
	token, err := uc.refreshTokenRepo.GetByToken(ctx, tokenHash)
	if err != nil {
		// Token not found - per RFC 7009, return success
		// This prevents token enumeration attacks
		return nil
	}

	// If already revoked, return success
	if token.IsRevoked() {
		return nil
	}

	// Revoke the token (FR-048)
	if err := uc.refreshTokenRepo.Revoke(ctx, tokenHash, "client_request"); err != nil {
		return &OAuthError{
			Code:        ErrServerError,
			Description: "failed to revoke token",
		}
	}

	return nil
}

// RevokeAllForUser revokes all refresh tokens for a user-client pair
// This is used when:
// - Admin explicitly revokes user's access (FR-049)
// - User's role is revoked from a client (FR-006e)
// - User account is suspended/deleted
func (uc *RevokeToken) RevokeAllForUser(ctx context.Context, userID string, clientID string) error {
	return uc.refreshTokenRepo.RevokeAllForUser(ctx, userID, clientID)
}

// hashTokenForRevocation hashes a token for database lookup
func hashTokenForRevocation(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
