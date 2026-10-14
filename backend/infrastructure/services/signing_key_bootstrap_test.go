package services

import (
	"context"
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/services"
	"github.com/stretchr/testify/require"
)

type recordingSigningKeyRepository struct {
	signingKeyRepositoryStub
}

func (r *recordingSigningKeyRepository) Create(_ context.Context, key *entities.SigningKey) error {
	r.key = key
	return nil
}

func TestEnsureSigningKey_CreatesUsableKeyWhenMissing(t *testing.T) {
	ctx := context.Background()
	repo := &recordingSigningKeyRepository{}

	created, err := EnsureSigningKey(ctx, repo)
	require.NoError(t, err)
	require.True(t, created)
	require.NotNil(t, repo.key)
	require.NoError(t, repo.key.Validate())
	require.True(t, repo.key.CanSign())

	tokenService := NewRSATokenService(repo)
	token, err := tokenService.SignIDToken(ctx, &services.TokenClaims{Issuer: "https://sso.test.com", Subject: "user-1", IssuedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)})
	require.NoError(t, err)
	_, err = tokenService.ValidateTokenSignature(ctx, token)
	require.NoError(t, err)

	jwks, err := tokenService.GetJWKS(ctx)
	require.NoError(t, err)
	require.Len(t, jwks.Keys, 1)
}

func TestEnsureSigningKey_KeepsExistingKey(t *testing.T) {
	existing := &entities.SigningKey{KeyID: "existing", Algorithm: "RS256", IsActive: true}
	repo := &recordingSigningKeyRepository{signingKeyRepositoryStub{key: existing}}

	created, err := EnsureSigningKey(context.Background(), repo)
	require.NoError(t, err)
	require.False(t, created)
	require.Same(t, existing, repo.key)
}
