package entities

import (
	"errors"
	"time"
)

// SigningKey represents an RSA keypair for JWT signing
type SigningKey struct {
	KeyID               string
	Algorithm           string
	PublicKey           string
	PrivateKeyEncrypted string
	IsActive            bool
	CreatedAt           time.Time
	ExpiresAt           *time.Time
}

// SupportedAlgorithms defines the allowed signing algorithms
var SupportedAlgorithms = []string{"RS256", "RS384", "RS512"}

// Validate performs basic validation on the signing key entity
func (sk *SigningKey) Validate() error {
	if sk.KeyID == "" {
		return errors.New("key_id cannot be empty")
	}
	if sk.Algorithm == "" {
		return errors.New("algorithm cannot be empty")
	}
	if sk.PublicKey == "" {
		return errors.New("public_key cannot be empty")
	}
	if sk.PrivateKeyEncrypted == "" {
		return errors.New("private_key_encrypted cannot be empty")
	}

	// Validate algorithm is supported
	if !sk.IsSupportedAlgorithm() {
		return errors.New("unsupported algorithm: must be one of RS256, RS384, RS512")
	}

	// Validate expiration if set
	if sk.ExpiresAt != nil && sk.ExpiresAt.Before(sk.CreatedAt) {
		return errors.New("expires_at must be after created_at")
	}

	return nil
}

// IsSupportedAlgorithm checks if the algorithm is in the supported list
func (sk *SigningKey) IsSupportedAlgorithm() bool {
	for _, algo := range SupportedAlgorithms {
		if sk.Algorithm == algo {
			return true
		}
	}
	return false
}

// IsExpired checks if the key has expired
func (sk *SigningKey) IsExpired() bool {
	if sk.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*sk.ExpiresAt)
}

// IsUsable checks if the key is active and not expired
func (sk *SigningKey) IsUsable() bool {
	return sk.IsActive && !sk.IsExpired()
}

// CanSign checks if this key can be used for signing new tokens
func (sk *SigningKey) CanSign() bool {
	return sk.IsUsable()
}

// CanVerify checks if this key can be used for verifying tokens
// Keys can verify even after being deactivated, as long as they haven't expired
func (sk *SigningKey) CanVerify() bool {
	return !sk.IsExpired()
}
