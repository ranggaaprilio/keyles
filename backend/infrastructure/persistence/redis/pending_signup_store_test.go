package redis

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ranggaaprilio/keyles/domain/repositories"
)

func newTestSignupStore(t *testing.T) (*redis.Client, *RedisPendingSignupStore) {
	t.Helper()
	client, _ := newTestThrottler(t) // shares the isolated, flushed test DB
	return client, &RedisPendingSignupStore{client: client}
}

func pendingSignup() *repositories.PendingSignup {
	return &repositories.PendingSignup{
		TenantID:     "tenant-1",
		Email:        "new@example.com",
		DisplayName:  "New User",
		PasswordHash: "bcrypt-hash",
		CodeHash:     "good-hash",
	}
}

func TestPendingSignupStore_VerifySuccessReturnsAndDeletes(t *testing.T) {
	_, store := newTestSignupStore(t)
	ctx := context.Background()
	require.NoError(t, store.Save(ctx, pendingSignup(), time.Minute))

	got, err := store.Verify(ctx, "tenant-1", "new@example.com", "good-hash", 5)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "bcrypt-hash", got.PasswordHash)
	assert.Equal(t, "New User", got.DisplayName)

	again, err := store.Verify(ctx, "tenant-1", "new@example.com", "good-hash", 5)
	require.NoError(t, err)
	assert.Nil(t, again, "code must be single-use")
}

func TestPendingSignupStore_AttemptsCapDeletesRecord(t *testing.T) {
	_, store := newTestSignupStore(t)
	ctx := context.Background()
	require.NoError(t, store.Save(ctx, pendingSignup(), time.Minute))

	for i := 0; i < 3; i++ {
		got, err := store.Verify(ctx, "tenant-1", "new@example.com", "bad-hash", 3)
		require.NoError(t, err)
		assert.Nil(t, got)
	}

	got, err := store.Verify(ctx, "tenant-1", "new@example.com", "good-hash", 3)
	require.NoError(t, err)
	assert.Nil(t, got, "correct code must fail once attempts are exhausted")
}

func TestPendingSignupStore_KeysDoNotExposeEmail(t *testing.T) {
	client, store := newTestSignupStore(t)
	ctx := context.Background()
	require.NoError(t, store.Save(ctx, pendingSignup(), time.Minute))

	keys, err := client.Keys(ctx, "*").Result()
	require.NoError(t, err)
	for _, k := range keys {
		assert.NotContains(t, k, "new@example.com")
	}
}

func TestPendingSignupStore_AllowSendCooldownAndHourlyCap(t *testing.T) {
	client, store := newTestSignupStore(t)
	ctx := context.Background()

	ok, err := store.AllowSend(ctx, "tenant-1", "new@example.com", time.Minute, 2)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = store.AllowSend(ctx, "tenant-1", "new@example.com", time.Minute, 2)
	require.NoError(t, err)
	assert.False(t, ok, "second send inside the cooldown must be refused")

	for i := 0; i < 2; i++ {
		require.NoError(t, client.Del(ctx, store.cooldownKey("tenant-1", "new@example.com")).Err())
		ok, err = store.AllowSend(ctx, "tenant-1", "new@example.com", time.Minute, 2)
		require.NoError(t, err)
		assert.Equal(t, i == 0, ok, "send %d", i+2)
	}
}
