package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// EnsureSigningKey generates and stores an RS256 signing key when no active one exists.
// It returns true when a new key was created.
func EnsureSigningKey(ctx context.Context, keyRepo repositories.SigningKeyRepository) (bool, error) {
	existing, err := keyRepo.GetActive(ctx, "RS256")
	if err != nil {
		return false, fmt.Errorf("failed to look up active signing key: %w", err)
	}
	if existing != nil {
		return false, nil
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return false, fmt.Errorf("failed to generate signing key: %w", err)
	}

	suffix := make([]byte, 4)
	if _, err := rand.Read(suffix); err != nil {
		return false, fmt.Errorf("failed to generate key id: %w", err)
	}
	now := time.Now().UTC()

	key := &entities.SigningKey{
		KeyID:     fmt.Sprintf("key_%s_%s", now.Format("20060102"), hex.EncodeToString(suffix)),
		Algorithm: "RS256",
		PublicKey: string(pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
		})),
		PrivateKeyEncrypted: string(pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		})),
		IsActive:  true,
		CreatedAt: now,
	}
	if err := keyRepo.Create(ctx, key); err != nil {
		return false, fmt.Errorf("failed to store signing key: %w", err)
	}
	return true, nil
}
