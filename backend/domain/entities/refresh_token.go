package entities

import (
	"errors"
	"time"
)

// RefreshToken represents a refresh token in the OAuth flow
type RefreshToken struct {
	ID                  int64
	Token               string
	UserID              string
	ClientID            string
	TenantID            string
	Scope               string
	ExpiresAt           time.Time
	CreatedAt           time.Time
	LastUsedAt          *time.Time
	RevokedFlag         bool
	RevokedAt           *time.Time
	RevokedReason       string
	FamilyID            string
	ParentTokenHash     string
	ReplacedByTokenHash string
}

// Validate performs basic validation on the refresh token
func (rt *RefreshToken) Validate() error {
	if rt.Token == "" {
		return errors.New("token cannot be empty")
	}
	if rt.UserID == "" {
		return errors.New("user_id cannot be empty")
	}
	if rt.ClientID == "" {
		return errors.New("client_id cannot be empty")
	}
	if rt.TenantID == "" {
		return errors.New("tenant_id cannot be empty")
	}
	if rt.ExpiresAt.IsZero() {
		return errors.New("expires_at cannot be zero")
	}
	return nil
}

// IsExpired checks if the token has expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsRevoked checks if the token has been revoked
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedFlag
}

// IsValid checks if the token is valid (not expired and not revoked)
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsExpired() && !rt.IsRevoked()
}

// Revoke marks the token as revoked
func (rt *RefreshToken) Revoke(reason string) {
	rt.RevokedFlag = true
	now := time.Now()
	rt.RevokedAt = &now
	rt.RevokedReason = reason
}

// MarkUsed updates the last used timestamp
func (rt *RefreshToken) MarkUsed() {
	now := time.Now()
	rt.LastUsedAt = &now
}
