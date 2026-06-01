package redis

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestThrottler(t *testing.T) (*redis.Client, *RedisLoginThrottler) {
	t.Helper()

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   9, // Use a separate DB for tests to avoid collisions
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		t.Skip("Redis not available, skipping integration test")
	}

	// Flush test DB before each test
	if err := client.FlushDB(ctx).Err(); err != nil {
		client.Close()
		t.Fatalf("flush db: %v", err)
	}

	throttler := &RedisLoginThrottler{
		client:        client,
		maxFailures:   5,
		windowSeconds: 900, // 15 minutes
	}

	t.Cleanup(func() {
		client.FlushDB(context.Background())
		client.Close()
	})

	return client, throttler
}

func TestIsThrottled_NotThrottledInitially(t *testing.T) {
	_, throttler := newTestThrottler(t)
	ctx := context.Background()

	throttled, err := throttler.IsThrottled(ctx, "192.0.2.1", "tenant-1", "user@example.com")
	require.NoError(t, err)
	assert.False(t, throttled, "should not be throttled with no failures")
}

func TestEmailKey_DoesNotExposeNormalizedEmail(t *testing.T) {
	throttler := &RedisLoginThrottler{}
	key := throttler.emailKey("tenant-1", "alice@example.com")

	assert.NotContains(t, key, "alice@example.com")
	assert.Contains(t, key, "oauth:login-failure:email:tenant-1:")
}

func TestIsThrottled_SourceIPCounter(t *testing.T) {
	_, throttler := newTestThrottler(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		require.NoError(t, throttler.RecordFailure(ctx, "192.0.2.1", "tenant-1", "different@example.com"))
	}

	throttled, err := throttler.IsThrottled(ctx, "192.0.2.1", "tenant-1", "user@example.com")
	require.NoError(t, err)
	assert.True(t, throttled, "should be throttled when IP counter >= maxFailures even with different email")
}

func TestIsThrottled_TenantEmailCounter(t *testing.T) {
	_, throttler := newTestThrottler(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		require.NoError(t, throttler.RecordFailure(ctx, "10.0.0.1", "tenant-1", "user@example.com"))
	}

	throttled, err := throttler.IsThrottled(ctx, "10.0.0.2", "tenant-1", "user@example.com")
	require.NoError(t, err)
	assert.True(t, throttled, "should be throttled when email counter >= maxFailures even from different IP")
}

func TestIsThrottled_FiveFailuresBlocks(t *testing.T) {
	_, throttler := newTestThrottler(t)
	ctx := context.Background()

	// Four failures should not block
	for i := 0; i < 4; i++ {
		require.NoError(t, throttler.RecordFailure(ctx, "192.0.2.1", "tenant-1", "user@example.com"))
	}

	throttled, err := throttler.IsThrottled(ctx, "192.0.2.1", "tenant-1", "user@example.com")
	require.NoError(t, err)
	assert.False(t, throttled, "should not be throttled after 4 failures (maxFailures=5)")

	// Fifth failure triggers throttling
	require.NoError(t, throttler.RecordFailure(ctx, "192.0.2.1", "tenant-1", "user@example.com"))

	throttled, err = throttler.IsThrottled(ctx, "192.0.2.1", "tenant-1", "user@example.com")
	require.NoError(t, err)
	assert.True(t, throttled, "should be throttled when both counters reach maxFailures")
}

func TestRecordFailure_TTLCreatedOnlyOnFirstFailure(t *testing.T) {
	client, throttler := newTestThrottler(t)
	ctx := context.Background()

	ipKey := throttler.ipKey("192.0.2.99")
	emailKey := throttler.emailKey("tenant-1", "ttl-test@example.com")

	// First failure: creates key with TTL
	require.NoError(t, throttler.RecordFailure(ctx, "192.0.2.99", "tenant-1", "ttl-test@example.com"))

	ipTTL1, err := client.TTL(ctx, ipKey).Result()
	require.NoError(t, err)
	assert.Greater(t, ipTTL1.Seconds(), float64(890), "IP key TTL should be close to 900s after first failure")

	emailTTL1, err := client.TTL(ctx, emailKey).Result()
	require.NoError(t, err)
	assert.Greater(t, emailTTL1.Seconds(), float64(890), "email key TTL should be close to 900s after first failure")

	// Wait a moment to let TTL tick down
	time.Sleep(2 * time.Second)

	// Second failure: should NOT extend TTL
	require.NoError(t, throttler.RecordFailure(ctx, "192.0.2.99", "tenant-1", "ttl-test@example.com"))

	ipTTL2, err := client.TTL(ctx, ipKey).Result()
	require.NoError(t, err)
	assert.Less(t, ipTTL2.Seconds(), ipTTL1.Seconds()+1,
		"IP TTL should NOT be extended on subsequent failure")

	emailTTL2, err := client.TTL(ctx, emailKey).Result()
	require.NoError(t, err)
	assert.Less(t, emailTTL2.Seconds(), emailTTL1.Seconds()+1,
		"email TTL should NOT be extended on subsequent failure")

	// Verify increments are atomic: count should be 2
	ipCount, err := client.Get(ctx, ipKey).Int()
	require.NoError(t, err)
	assert.Equal(t, 2, ipCount)

	emailCount, err := client.Get(ctx, emailKey).Int()
	require.NoError(t, err)
	assert.Equal(t, 2, emailCount)
}

func TestRecordFailure_AtomicIncrementsWithoutTTLExtension(t *testing.T) {
	client, throttler := newTestThrottler(t)
	ctx := context.Background()

	// Record 3 failures
	for i := 0; i < 3; i++ {
		require.NoError(t, throttler.RecordFailure(ctx, "192.0.2.50", "tenant-x", "atomic@example.com"))
	}

	ipKey := throttler.ipKey("192.0.2.50")
	emailKey := throttler.emailKey("tenant-x", "atomic@example.com")

	ipCount, err := client.Get(ctx, ipKey).Int()
	require.NoError(t, err)
	assert.Equal(t, 3, ipCount, "IP counter should be exactly 3")

	emailCount, err := client.Get(ctx, emailKey).Int()
	require.NoError(t, err)
	assert.Equal(t, 3, emailCount, "email counter should be exactly 3")
}

func TestClearEmailBucket(t *testing.T) {
	client, throttler := newTestThrottler(t)
	ctx := context.Background()

	// Record failures
	require.NoError(t, throttler.RecordFailure(ctx, "192.0.2.10", "tenant-1", "clear@example.com"))
	require.NoError(t, throttler.RecordFailure(ctx, "192.0.2.10", "tenant-1", "clear@example.com"))

	emailKey := throttler.emailKey("tenant-1", "clear@example.com")
	countBefore, err := client.Get(ctx, emailKey).Int()
	require.NoError(t, err)
	assert.Equal(t, 2, countBefore, "email counter should be 2 before clearing")

	// Clear the email bucket
	require.NoError(t, throttler.ClearEmailBucket(ctx, "tenant-1", "clear@example.com"))

	_, err = client.Get(ctx, emailKey).Int()
	assert.ErrorIs(t, err, redis.Nil, "email counter should be removed after clearing")

	// IP counter should still exist
	ipKey := throttler.ipKey("192.0.2.10")
	ipCount, err := client.Get(ctx, ipKey).Int()
	require.NoError(t, err)
	assert.Equal(t, 2, ipCount, "IP counter should not be affected by clearing email bucket")
}

func TestClearEmailBucket_EnablesLogin(t *testing.T) {
	_, throttler := newTestThrottler(t)
	ctx := context.Background()

	// Reach max failures on email
	for i := 0; i < 5; i++ {
		require.NoError(t, throttler.RecordFailure(ctx, "10.0.0.1", "tenant-1", "unblock@example.com"))
	}

	// Verify throttled
	throttled, err := throttler.IsThrottled(ctx, "10.0.0.1", "tenant-1", "unblock@example.com")
	require.NoError(t, err)
	assert.True(t, throttled, "should be throttled after reaching max failures")

	// Clear email bucket but IP is still blocked
	require.NoError(t, throttler.ClearEmailBucket(ctx, "tenant-1", "unblock@example.com"))

	// From same IP still blocked
	throttled, err = throttler.IsThrottled(ctx, "10.0.0.1", "tenant-1", "unblock@example.com")
	require.NoError(t, err)
	assert.True(t, throttled, "should still be throttled from IP counter even after clearing email bucket")

	// From different IP, different email bucket started fresh
	throttled, err = throttler.IsThrottled(ctx, "10.0.0.2", "tenant-1", "unblock@example.com")
	require.NoError(t, err)
	assert.False(t, throttled, "should not be throttled from new IP after clearing email bucket")
}

func TestRedisError(t *testing.T) {
	// Create a throttler pointing to a non-existent Redis to trigger errors
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:16379", // unlikely to have Redis here
	})
	defer client.Close()

	throttler := &RedisLoginThrottler{
		client:        client,
		maxFailures:   5,
		windowSeconds: 900,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := throttler.IsThrottled(ctx, "192.0.2.1", "tenant-1", "error@example.com")
	assert.Error(t, err, "IsThrottled should return error when Redis is unreachable")

	err = throttler.RecordFailure(ctx, "192.0.2.1", "tenant-1", "error@example.com")
	assert.Error(t, err, "RecordFailure should return error when Redis is unreachable")

	err = throttler.ClearEmailBucket(ctx, "tenant-1", "error@example.com")
	assert.Error(t, err, "ClearEmailBucket should return error when Redis is unreachable")
}
