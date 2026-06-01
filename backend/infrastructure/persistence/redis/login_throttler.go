package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/ranggaaprilio/keyles/domain/services"
)

// RedisLoginThrottler implements services.LoginThrottler using Redis
// fixed-window counters for source-IP and tenant-scoped email.
//
// The TTL is set only when a counter is first created; subsequent
// failures increment without extending the window.
type RedisLoginThrottler struct {
	client        *redis.Client
	maxFailures   int
	windowSeconds int
}

// NewRedisLoginThrottler creates a new Redis-backed login throttler.
func NewRedisLoginThrottler(client *redis.Client, maxFailures int, windowSeconds int) services.LoginThrottler {
	return &RedisLoginThrottler{
		client:        client,
		maxFailures:   maxFailures,
		windowSeconds: windowSeconds,
	}
}

const (
	// recordFailureScript increments a counter. If the key does not exist it
	// is created with TTL windowSeconds; if it already exists only INCRBY is
	// applied (no TTL extension).
	recordFailureScript = `
local current = redis.call("INCRBY", KEYS[1], ARGV[1])
if current == tonumber(ARGV[1]) then
	redis.call("EXPIRE", KEYS[1], ARGV[2])
end
return current
`
)

// ipKey returns the Redis key for source-IP failure counters.
func (t *RedisLoginThrottler) ipKey(sourceIP string) string {
	return fmt.Sprintf("oauth:login-failure:ip:%s", sourceIP)
}

// emailKey returns the Redis key for tenant-scoped email failure counters.
func (t *RedisLoginThrottler) emailKey(tenantID, normalizedEmail string) string {
	digest := sha256.Sum256([]byte(normalizedEmail))
	return fmt.Sprintf("oauth:login-failure:email:%s:%s", tenantID, hex.EncodeToString(digest[:]))
}

// IsThrottled returns true when either the source-IP counter or the
// tenant-email counter has reached or exceeded the configured maximum.
func (t *RedisLoginThrottler) IsThrottled(ctx context.Context, sourceIP, tenantID, normalizedEmail string) (bool, error) {
	ipCount, err := t.getCounter(ctx, t.ipKey(sourceIP))
	if err != nil {
		return false, fmt.Errorf("failed to check IP throttle: %w", err)
	}
	if ipCount >= t.maxFailures {
		return true, nil
	}

	emailCount, err := t.getCounter(ctx, t.emailKey(tenantID, normalizedEmail))
	if err != nil {
		return false, fmt.Errorf("failed to check email throttle: %w", err)
	}
	if emailCount >= t.maxFailures {
		return true, nil
	}

	return false, nil
}

// RecordFailure atomically increments both the source-IP counter and the
// tenant-email counter. Each key is created with the configured window TTL
// only on the first failure; later increments do not extend the TTL.
func (t *RedisLoginThrottler) RecordFailure(ctx context.Context, sourceIP, tenantID, normalizedEmail string) error {
	script := redis.NewScript(recordFailureScript)
	keys := []string{t.ipKey(sourceIP), t.emailKey(tenantID, normalizedEmail)}

	for _, key := range keys {
		_, err := script.Run(ctx, t.client, []string{key}, 1, t.windowSeconds).Int()
		if err != nil {
			return fmt.Errorf("failed to record login failure for key %s: %w", key, err)
		}
	}

	return nil
}

// ClearEmailBucket removes the tenant-email counter so a user is not
// penalised by past failures on a subsequent login.
func (t *RedisLoginThrottler) ClearEmailBucket(ctx context.Context, tenantID, normalizedEmail string) error {
	key := t.emailKey(tenantID, normalizedEmail)
	if err := t.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to clear email bucket: %w", err)
	}
	return nil
}

// getCounter reads an integer counter from Redis. Returns 0 when the key
// does not exist.
func (t *RedisLoginThrottler) getCounter(ctx context.Context, key string) (int, error) {
	val, err := t.client.Get(ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	return val, nil
}
