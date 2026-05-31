package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRefreshTokenRotateAndReplayRevokesFamily(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&RefreshTokenModel{}))

	repo := NewPostgresRefreshTokenRepository(db)
	ctx := context.Background()
	original := &entities.RefreshToken{
		Token: "original-hash", UserID: "user", ClientID: "client", TenantID: "tenant",
		Scope: "openid", ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, original))

	rotator := repo.(repositories.RefreshTokenRotationRepository)
	replacement := &entities.RefreshToken{
		Token: "replacement-hash", UserID: "user", ClientID: "client", TenantID: "tenant",
		Scope: "openid", ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now(),
	}
	require.NoError(t, rotator.Rotate(ctx, original.Token, replacement))

	storedReplacement, err := repo.GetByToken(ctx, replacement.Token)
	require.NoError(t, err)
	require.Equal(t, original.Token, storedReplacement.FamilyID)
	require.Equal(t, original.Token, storedReplacement.ParentTokenHash)

	replayedReplacement := &entities.RefreshToken{
		Token: "should-not-be-inserted", UserID: "user", ClientID: "client", TenantID: "tenant",
		Scope: "openid", ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now(),
	}
	err = rotator.Rotate(ctx, original.Token, replayedReplacement)
	require.True(t, errors.Is(err, repositories.ErrRefreshTokenReplay))

	storedReplacement, err = repo.GetByToken(ctx, replacement.Token)
	require.NoError(t, err)
	require.True(t, storedReplacement.IsRevoked())
	require.Equal(t, "refresh token replay detected", storedReplacement.RevokedReason)
}

func TestRefreshTokenListByUserIDOnlyReturnsActiveSessions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&RefreshTokenModel{}))

	repo := NewPostgresRefreshTokenRepository(db)
	ctx := context.Background()
	for _, token := range []*entities.RefreshToken{
		{Token: "active", UserID: "user", ClientID: "client", TenantID: "tenant", ExpiresAt: time.Now().Add(time.Hour)},
		{Token: "revoked", UserID: "user", ClientID: "client", TenantID: "tenant", ExpiresAt: time.Now().Add(time.Hour), RevokedFlag: true},
		{Token: "expired", UserID: "user", ClientID: "client", TenantID: "tenant", ExpiresAt: time.Now().Add(-time.Hour)},
	} {
		require.NoError(t, repo.Create(ctx, token))
	}

	tokens, err := repo.ListByUserID(ctx, "user")
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	require.Equal(t, "active", tokens[0].Token)
}
