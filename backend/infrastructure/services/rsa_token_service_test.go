package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/services"
	"github.com/stretchr/testify/require"
)

type signingKeyRepositoryStub struct {
	key *entities.SigningKey
}

func (s *signingKeyRepositoryStub) Create(context.Context, *entities.SigningKey) error {
	return nil
}

func (s *signingKeyRepositoryStub) GetActive(context.Context, string) (*entities.SigningKey, error) {
	return s.key, nil
}

func (s *signingKeyRepositoryStub) GetByKeyID(context.Context, string) (*entities.SigningKey, error) {
	return s.key, nil
}

func (s *signingKeyRepositoryStub) ListActive(context.Context) ([]*entities.SigningKey, error) {
	return []*entities.SigningKey{s.key}, nil
}

func (s *signingKeyRepositoryStub) Deactivate(context.Context, string) error {
	return nil
}

func (s *signingKeyRepositoryStub) DeleteExpired(context.Context) (int64, error) {
	return 0, nil
}

func TestRSATokenService_RolesRoundTrip(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	key := &entities.SigningKey{
		KeyID:               "test-key",
		Algorithm:           "RS256",
		PublicKey:           string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)})),
		PrivateKeyEncrypted: string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})),
		IsActive:            true,
	}
	tokenService := NewRSATokenService(&signingKeyRepositoryStub{key: key})
	now := time.Now().Truncate(time.Second)
	claims := &services.TokenClaims{
		Issuer:    "https://sso.test.com",
		Subject:   "user-123",
		Audience:  []string{"client-123"},
		ExpiresAt: now.Add(15 * time.Minute),
		IssuedAt:  now,
		NotBefore: now,
		JWTID:     "jwt-123",
		TenantID:  "tenant-123",
		ClientID:  "client-123",
		Scope:     "openid profile",
		Roles:     []string{"reader", "writer"},
	}

	token, err := tokenService.SignAccessToken(context.Background(), claims)
	require.NoError(t, err)

	parsed, err := tokenService.ValidateTokenSignature(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, claims.Roles, parsed.Roles)
}
