package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/redis/go-redis/v9"

	"github.com/ranggaaprilio/keyles/domain/repositories"
)

func newTxTestClient(t *testing.T) *redis.Client {
	t.Helper()
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available on localhost:6379, skipping integration test")
	}
	return client
}

func uniqueID() string {
	return fmt.Sprintf("txn-%d", time.Now().UnixNano())
}

func sampleTransaction(id string) *repositories.AuthorizationTransaction {
	maxAge := 300
	return &repositories.AuthorizationTransaction{
		TransactionID:        id,
		ClientID:             "client-123",
		TenantID:             "tenant-456",
		RedirectURI:          "https://example.com/callback",
		ResponseType:         "code",
		Scope:                "openid profile",
		State:                "statetoken",
		CodeChallenge:        "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		CodeChallengeMethod:  "S256",
		Nonce:                "nonce-abc",
		Prompt:               []string{"consent"},
		MaxAgeSeconds:        &maxAge,
		UserID:               "",
		SessionID:            "",
		InteractionCSRFToken: "csrf-token-xyz",
		Stage:                repositories.TransactionStagePendingLogin,
		CreatedAt:            time.Now().UTC().Truncate(time.Millisecond),
		ExpiresAt:            time.Now().UTC().Add(10 * time.Minute).Truncate(time.Millisecond),
	}
}

func cleanupKey(t *testing.T, client *redis.Client, id string) {
	t.Helper()
	key := fmt.Sprintf("oauth:transaction:%s", id)
	client.Del(context.Background(), key)
}

func TestCreate_StoresWithTTL(t *testing.T) {
	client := newTxTestClient(t)
	defer client.Close()

	id := uniqueID()
	txn := sampleTransaction(id)
	defer cleanupKey(t, client, id)

	repo := NewRedisAuthorizationTransactionRepository(client)
	ctx := context.Background()

	ttl := 10 * time.Minute
	err := repo.Create(ctx, txn, ttl)
	require.NoError(t, err)

	key := fmt.Sprintf("oauth:transaction:%s", id)
	duration, err := client.TTL(ctx, key).Result()
	require.NoError(t, err)

	assert.GreaterOrEqual(t, duration, 9*time.Minute)
	assert.LessOrEqual(t, duration, 10*time.Minute)
}

func TestGet_RetrievesTransaction(t *testing.T) {
	client := newTxTestClient(t)
	defer client.Close()

	id := uniqueID()
	txn := sampleTransaction(id)
	defer cleanupKey(t, client, id)

	repo := NewRedisAuthorizationTransactionRepository(client)
	ctx := context.Background()

	ttl := 10 * time.Minute
	err := repo.Create(ctx, txn, ttl)
	require.NoError(t, err)

	got, err := repo.Get(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, txn.TransactionID, got.TransactionID)
	assert.Equal(t, txn.ClientID, got.ClientID)
	assert.Equal(t, txn.TenantID, got.TenantID)
	assert.Equal(t, txn.RedirectURI, got.RedirectURI)
	assert.Equal(t, txn.ResponseType, got.ResponseType)
	assert.Equal(t, txn.Scope, got.Scope)
	assert.Equal(t, txn.State, got.State)
	assert.Equal(t, txn.CodeChallenge, got.CodeChallenge)
	assert.Equal(t, txn.CodeChallengeMethod, got.CodeChallengeMethod)
	assert.Equal(t, txn.Nonce, got.Nonce)
	assert.Equal(t, txn.Prompt, got.Prompt)
	require.NotNil(t, got.MaxAgeSeconds)
	assert.Equal(t, *txn.MaxAgeSeconds, *got.MaxAgeSeconds)
	assert.Equal(t, txn.InteractionCSRFToken, got.InteractionCSRFToken)
	assert.Equal(t, txn.Stage, got.Stage)
	assert.WithinDuration(t, txn.CreatedAt, got.CreatedAt, time.Second)
	assert.WithinDuration(t, txn.ExpiresAt, got.ExpiresAt, time.Second)
}

func TestGet_NotFound(t *testing.T) {
	client := newTxTestClient(t)
	defer client.Close()

	repo := NewRedisAuthorizationTransactionRepository(client)
	ctx := context.Background()

	got, err := repo.Get(ctx, "nonexistent-id-12345")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestUpdateStage_BindsUserAndSession(t *testing.T) {
	client := newTxTestClient(t)
	defer client.Close()

	id := uniqueID()
	txn := sampleTransaction(id)
	defer cleanupKey(t, client, id)

	repo := NewRedisAuthorizationTransactionRepository(client)
	ctx := context.Background()

	ttl := 10 * time.Minute
	err := repo.Create(ctx, txn, ttl)
	require.NoError(t, err)

	err = repo.UpdateStage(ctx, id, repositories.TransactionStagePendingConsent, "user-789", "session-101")
	require.NoError(t, err)

	got, err := repo.Get(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, repositories.TransactionStagePendingConsent, got.Stage)
	assert.Equal(t, "user-789", got.UserID)
	assert.Equal(t, "session-101", got.SessionID)
}

func TestUpdateStage_NotFound(t *testing.T) {
	client := newTxTestClient(t)
	defer client.Close()

	repo := NewRedisAuthorizationTransactionRepository(client)
	ctx := context.Background()

	err := repo.UpdateStage(ctx, "nonexistent-id-12345", repositories.TransactionStagePendingConsent, "user-1", "session-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), repositories.ErrTransactionNotFound)
}

func TestComplete_AtomicOneTimeConsumption(t *testing.T) {
	client := newTxTestClient(t)
	defer client.Close()

	id := uniqueID()
	txn := sampleTransaction(id)
	defer cleanupKey(t, client, id)

	repo := NewRedisAuthorizationTransactionRepository(client)
	ctx := context.Background()

	ttl := 10 * time.Minute
	err := repo.Create(ctx, txn, ttl)
	require.NoError(t, err)

	err = repo.UpdateStage(ctx, id, repositories.TransactionStagePendingConsent, "user-789", "session-101")
	require.NoError(t, err)

	// First completion should succeed
	completed, err := repo.Complete(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, completed)

	assert.Equal(t, repositories.TransactionStageCompleted, completed.Stage)
	assert.Equal(t, "user-789", completed.UserID)
	assert.Equal(t, "session-101", completed.SessionID)
	assert.Equal(t, txn.ClientID, completed.ClientID)
	assert.Equal(t, txn.RedirectURI, completed.RedirectURI)
	assert.Equal(t, txn.Scope, completed.Scope)
	assert.Equal(t, txn.State, completed.State)
	assert.Equal(t, txn.CodeChallenge, completed.CodeChallenge)
	require.NotNil(t, completed.CompletedAt)

	// Second completion should fail (replay rejection)
	_, err = repo.Complete(ctx, id)
	require.Error(t, err)
	assert.Contains(t, err.Error(), repositories.ErrTransactionAlreadyCompleted)
}

func TestComplete_NotFound(t *testing.T) {
	client := newTxTestClient(t)
	defer client.Close()

	repo := NewRedisAuthorizationTransactionRepository(client)
	ctx := context.Background()

	_, err := repo.Complete(ctx, "nonexistent-id-12345")
	require.Error(t, err)
	assert.Contains(t, err.Error(), repositories.ErrTransactionNotFound)
}

func TestComplete_WrongStage(t *testing.T) {
	client := newTxTestClient(t)
	defer client.Close()

	id := uniqueID()
	txn := sampleTransaction(id)
	defer cleanupKey(t, client, id)

	repo := NewRedisAuthorizationTransactionRepository(client)
	ctx := context.Background()

	ttl := 10 * time.Minute
	err := repo.Create(ctx, txn, ttl)
	require.NoError(t, err)

	// Transaction is at pending_login stage, not pending_consent.
	_, err = repo.Complete(ctx, id)
	require.Error(t, err)
	assert.Contains(t, err.Error(), repositories.ErrTransactionWrongStage)
}

func TestExpiry_ShortTTL(t *testing.T) {
	client := newTxTestClient(t)
	defer client.Close()

	id := uniqueID()
	txn := sampleTransaction(id)
	defer cleanupKey(t, client, id)

	repo := NewRedisAuthorizationTransactionRepository(client)
	ctx := context.Background()

	// Create with a very short TTL
	ttl := 1 * time.Second
	err := repo.Create(ctx, txn, ttl)
	require.NoError(t, err)

	// Verify it exists immediately
	got, err := repo.Get(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, got)

	// Wait for expiry
	time.Sleep(2 * time.Second)

	// Should be gone after expiry
	got, err = repo.Get(ctx, id)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestReplayRejection_SecondCompleteFails(t *testing.T) {
	client := newTxTestClient(t)
	defer client.Close()

	id := uniqueID()
	txn := sampleTransaction(id)
	defer cleanupKey(t, client, id)

	repo := NewRedisAuthorizationTransactionRepository(client)
	ctx := context.Background()

	ttl := 10 * time.Minute
	err := repo.Create(ctx, txn, ttl)
	require.NoError(t, err)

	err = repo.UpdateStage(ctx, id, repositories.TransactionStagePendingConsent, "user-1", "sess-1")
	require.NoError(t, err)

	// Complete first time - succeeds
	_, err = repo.Complete(ctx, id)
	require.NoError(t, err)

	// Complete second time - must fail with already_completed
	_, err = repo.Complete(ctx, id)
	require.Error(t, err)
	assert.Contains(t, err.Error(), repositories.ErrTransactionAlreadyCompleted)

	// Get should still return the completed transaction
	got, err := repo.Get(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, repositories.TransactionStageCompleted, got.Stage)
}
