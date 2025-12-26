package entities

import (
	"errors"
	"time"
)

// AuthorizationCode represents an OAuth authorization code
type AuthorizationCode struct {
	Code                string
	ClientID            string
	TenantID            string
	UserID              string
	RedirectURI         string
	Scope               string
	CodeChallenge       string
	CodeChallengeMethod string
	ExpiresAt           time.Time
	CreatedAt           time.Time
	UsedFlag            bool
	UsedAt              *time.Time
}

// Validate performs basic validation on the authorization code
func (ac *AuthorizationCode) Validate() error {
	if ac.Code == "" {
		return errors.New("code cannot be empty")
	}
	if ac.ClientID == "" {
		return errors.New("client_id cannot be empty")
	}
	if ac.TenantID == "" {
		return errors.New("tenant_id cannot be empty")
	}
	if ac.UserID == "" {
		return errors.New("user_id cannot be empty")
	}
	if ac.RedirectURI == "" {
		return errors.New("redirect_uri cannot be empty")
	}
	if ac.ExpiresAt.IsZero() {
		return errors.New("expires_at cannot be zero")
	}

	// PKCE validation: code_challenge_method must be S256 if code_challenge is present
	if ac.CodeChallenge != "" && ac.CodeChallengeMethod != "S256" {
		return errors.New("code_challenge_method must be S256 when code_challenge is provided")
	}

	return nil
}

// IsExpired checks if the authorization code has expired
func (ac *AuthorizationCode) IsExpired() bool {
	return time.Now().After(ac.ExpiresAt)
}

// IsUsed checks if the authorization code has already been used
func (ac *AuthorizationCode) IsUsed() bool {
	return ac.UsedFlag
}

// IsValid checks if the authorization code is valid (not expired and not used)
func (ac *AuthorizationCode) IsValid() bool {
	return !ac.IsExpired() && !ac.IsUsed()
}

// MarkAsUsed marks the authorization code as used
func (ac *AuthorizationCode) MarkAsUsed() {
	ac.UsedFlag = true
	now := time.Now()
	ac.UsedAt = &now
}

// ValidatePKCE validates the PKCE code verifier
func (ac *AuthorizationCode) ValidatePKCE(codeVerifier string) bool {
	if ac.CodeChallenge == "" {
		return false
	}
	// This should use SHA256 hash and base64url encoding
	// Implementation will be in the service layer
	return true
}
