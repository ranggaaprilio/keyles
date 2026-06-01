package redis

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ranggaaprilio/keyles/domain/repositories"
)

func newSessionTestClient(t *testing.T) *redis.Client {
	t.Helper()
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		t.Skip("Redis not available")
	}
	t.Cleanup(func() { client.Close() })
	return client
}

func uniqueSessionID() string {
	return "sess_" + uuid.New().String()
}

// --- marshaling unit tests (no Redis needed) ---

func TestMarshalSession_ClientAgnosticShape(t *testing.T) {
	s := &repositories.Session{
		SessionID:       uniqueSessionID(),
		UserID:          "user-1",
		TenantID:        "tenant-1",
		AuthenticatedAt: time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC),
		CreatedAt:       time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC),
		ExpiresAt:       time.Date(2025, 6, 1, 20, 0, 0, 0, time.UTC),
	}

	data, err := marshalSession(s)
	require.NoError(t, err)

	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))

	// Must NOT contain client_id — sessions are client-agnostic.
	_, hasClientID := raw["client_id"]
	assert.False(t, hasClientID, "serialized session must not contain client_id field")

	// Must contain the expected client-agnostic fields.
	_, hasSessionID := raw["session_id"]
	assert.True(t, hasSessionID, "serialized session must contain session_id")

	_, hasUserID := raw["user_id"]
	assert.True(t, hasUserID, "serialized session must contain user_id")

	_, hasTenantID := raw["tenant_id"]
	assert.True(t, hasTenantID, "serialized session must contain tenant_id")

	_, hasAuthAt := raw["authenticated_at"]
	assert.True(t, hasAuthAt, "serialized session must contain authenticated_at")
}

func TestMarshalSession_AuthenticatedAtRoundTrip(t *testing.T) {
	// Use a time with non-zero nanoseconds to verify precision is preserved.
	ts := time.Date(2025, 6, 1, 14, 30, 45, 123456789, time.UTC)

	s := &repositories.Session{
		SessionID:       uniqueSessionID(),
		UserID:          "user-1",
		TenantID:        "tenant-1",
		AuthenticatedAt: ts,
		CreatedAt:       time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		ExpiresAt:       time.Date(2025, 6, 1, 8, 0, 0, 0, time.UTC),
	}

	data, err := marshalSession(s)
	require.NoError(t, err)

	got, err := unmarshalSession(data)
	require.NoError(t, err)

	assert.True(t, got.AuthenticatedAt.Equal(ts),
		"AuthenticatedAt must round-trip: got %v, want %v", got.AuthenticatedAt, ts)

	// Also verify the RFC3339Nano string representation in the wire format.
	var d sessionData
	require.NoError(t, json.Unmarshal(data, &d))
	assert.Equal(t, ts.Format(time.RFC3339Nano), d.AuthenticatedAt)
}

func TestMarshalSession_AuthenticatedAtZeroNanosRoundTrip(t *testing.T) {
	// Zero nanoseconds must also round-trip correctly.
	ts := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	s := &repositories.Session{
		SessionID:       uniqueSessionID(),
		UserID:          "user-2",
		TenantID:        "tenant-2",
		AuthenticatedAt: ts,
		CreatedAt:       ts,
		ExpiresAt:       ts.Add(8 * time.Hour),
	}

	data, err := marshalSession(s)
	require.NoError(t, err)

	got, err := unmarshalSession(data)
	require.NoError(t, err)

	assert.True(t, got.AuthenticatedAt.Equal(ts),
		"AuthenticatedAt zero-nanos must round-trip: got %v, want %v", got.AuthenticatedAt, ts)
}

// --- integration tests (require Redis) ---

func TestRedisSessionRepository_CreateAndGet(t *testing.T) {
	client := newSessionTestClient(t)
	repo := NewRedisSessionRepository(client)
	ctx := context.Background()

	sid := uniqueSessionID()
	now := time.Now().UTC().Truncate(time.Nanosecond)

	s := &repositories.Session{
		SessionID:       sid,
		UserID:          "user-10",
		TenantID:        "tenant-10",
		AuthenticatedAt: now,
		CreatedAt:       now,
		ExpiresAt:       now.Add(8 * time.Hour),
		Metadata:        map[string]interface{}{"ip": "127.0.0.1"},
	}

	err := repo.Create(ctx, s, 8*time.Hour)
	require.NoError(t, err)

	got, err := repo.Get(ctx, sid)
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, s.SessionID, got.SessionID)
	assert.Equal(t, s.UserID, got.UserID)
	assert.Equal(t, s.TenantID, got.TenantID)
	assert.True(t, got.AuthenticatedAt.Equal(s.AuthenticatedAt),
		"AuthenticatedAt mismatch: got %v, want %v", got.AuthenticatedAt, s.AuthenticatedAt)
	assert.True(t, got.CreatedAt.Equal(s.CreatedAt))
	assert.True(t, got.ExpiresAt.Equal(s.ExpiresAt))

	// Cleanup
	require.NoError(t, repo.Delete(ctx, sid))
}

func TestRedisSessionRepository_Get_NotFound(t *testing.T) {
	client := newSessionTestClient(t)
	repo := NewRedisSessionRepository(client)
	ctx := context.Background()

	got, err := repo.Get(ctx, "nonexistent_session")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestRedisSessionRepository_Delete(t *testing.T) {
	client := newSessionTestClient(t)
	repo := NewRedisSessionRepository(client)
	ctx := context.Background()

	sid := uniqueSessionID()
	now := time.Now().UTC().Truncate(time.Nanosecond)

	s := &repositories.Session{
		SessionID:       sid,
		UserID:          "user-20",
		TenantID:        "tenant-20",
		AuthenticatedAt: now,
		CreatedAt:       now,
		ExpiresAt:       now.Add(8 * time.Hour),
	}

	require.NoError(t, repo.Create(ctx, s, 30*time.Minute))
	require.NoError(t, repo.Delete(ctx, sid))

	got, err := repo.Get(ctx, sid)
	require.NoError(t, err)
	assert.Nil(t, got, "session must be nil after deletion")
}

func TestRedisSessionRepository_Exists(t *testing.T) {
	client := newSessionTestClient(t)
	repo := NewRedisSessionRepository(client)
	ctx := context.Background()

	sid := uniqueSessionID()
	now := time.Now().UTC().Truncate(time.Nanosecond)

	exists, err := repo.Exists(ctx, sid)
	require.NoError(t, err)
	assert.False(t, exists, "session should not exist before creation")

	s := &repositories.Session{
		SessionID:       sid,
		UserID:          "user-30",
		TenantID:        "tenant-30",
		AuthenticatedAt: now,
		CreatedAt:       now,
		ExpiresAt:       now.Add(8 * time.Hour),
	}

	require.NoError(t, repo.Create(ctx, s, 30*time.Minute))

	exists, err = repo.Exists(ctx, sid)
	require.NoError(t, err)
	assert.True(t, exists, "session should exist after creation")

	require.NoError(t, repo.Delete(ctx, sid))
}

func TestRedisSessionRepository_AuthenticatedAt_IntegrationRoundTrip(t *testing.T) {
	client := newSessionTestClient(t)
	repo := NewRedisSessionRepository(client)
	ctx := context.Background()

	sid := uniqueSessionID()

	// Use a time with full nanosecond precision to exercise the wire format.
	ts := time.Date(2025, 12, 31, 23, 59, 59, 999999999, time.UTC)

	s := &repositories.Session{
		SessionID:       sid,
		UserID:          "user-40",
		TenantID:        "tenant-40",
		AuthenticatedAt: ts,
		CreatedAt:       ts.Add(-time.Hour),
		ExpiresAt:       ts.Add(7 * time.Hour),
	}

	require.NoError(t, repo.Create(ctx, s, 8*time.Hour))

	got, err := repo.Get(ctx, sid)
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.True(t, got.AuthenticatedAt.Equal(ts),
		"AuthenticatedAt must survive Redis round-trip: got %v, want %v",
		got.AuthenticatedAt, ts)

	require.NoError(t, repo.Delete(ctx, sid))
}