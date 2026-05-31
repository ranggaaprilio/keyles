package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// RedisAuthCodeRepository implements AuthCodeRepository using Redis
type RedisAuthCodeRepository struct {
	client *redis.Client
}

// NewRedisAuthCodeRepository creates a new Redis authorization code repository
func NewRedisAuthCodeRepository(client *redis.Client) repositories.AuthCodeRepository {
	return &RedisAuthCodeRepository{client: client}
}

// authCodeKey generates the Redis key for an authorization code
func (r *RedisAuthCodeRepository) authCodeKey(code string) string {
	return fmt.Sprintf("oauth:authcode:%s", code)
}

// Store saves an authorization code with automatic expiration
func (r *RedisAuthCodeRepository) Store(ctx context.Context, code *entities.AuthorizationCode, ttl time.Duration) error {
	key := r.authCodeKey(code.Code)

	data, err := json.Marshal(code)
	if err != nil {
		return fmt.Errorf("failed to marshal authorization code: %w", err)
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

// Get retrieves an authorization code
func (r *RedisAuthCodeRepository) Get(ctx context.Context, code string) (*entities.AuthorizationCode, error) {
	key := r.authCodeKey(code)

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Code not found or expired
		}
		return nil, fmt.Errorf("failed to get authorization code: %w", err)
	}

	var authCode entities.AuthorizationCode
	if err := json.Unmarshal(data, &authCode); err != nil {
		return nil, fmt.Errorf("failed to unmarshal authorization code: %w", err)
	}

	return &authCode, nil
}

// MarkAsUsed marks an authorization code as used
func (r *RedisAuthCodeRepository) MarkAsUsed(ctx context.Context, code string) error {
	authCode, err := r.Get(ctx, code)
	if err != nil {
		return err
	}

	if authCode == nil {
		return fmt.Errorf("authorization code not found")
	}

	authCode.MarkAsUsed()

	// Update the stored code
	key := r.authCodeKey(code)
	data, err := json.Marshal(authCode)
	if err != nil {
		return fmt.Errorf("failed to marshal authorization code: %w", err)
	}

	// Keep TTL from original code
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

// Consume atomically marks an authorization code as used and returns its
// previous value. The Lua script prevents concurrent exchanges from both
// succeeding.
func (r *RedisAuthCodeRepository) Consume(ctx context.Context, code string) (*entities.AuthorizationCode, error) {
	key := r.authCodeKey(code)
	script := redis.NewScript(`
local value = redis.call("GET", KEYS[1])
if not value then
	return nil
end

local parsed = cjson.decode(value)
if parsed.UsedFlag == true then
	return nil
end

parsed.UsedFlag = true
parsed.UsedAt = ARGV[1]
redis.call("SET", KEYS[1], cjson.encode(parsed), "KEEPTTL")
return value
`)

	value, err := script.Run(ctx, r.client, []string{key}, time.Now().Format(time.RFC3339Nano)).Text()
	if err != nil {
		if err == redis.Nil {
			return nil, repositories.ErrAuthorizationCodeUnavailable
		}
		return nil, fmt.Errorf("failed to consume authorization code: %w", err)
	}

	var authCode entities.AuthorizationCode
	if err := json.Unmarshal([]byte(value), &authCode); err != nil {
		return nil, fmt.Errorf("failed to unmarshal authorization code: %w", err)
	}
	if authCode.IsExpired() {
		return nil, repositories.ErrAuthorizationCodeUnavailable
	}
	return &authCode, nil
}

// Delete removes an authorization code
func (r *RedisAuthCodeRepository) Delete(ctx context.Context, code string) error {
	key := r.authCodeKey(code)
	return r.client.Del(ctx, key).Err()
}

// Exists checks if an authorization code exists
func (r *RedisAuthCodeRepository) Exists(ctx context.Context, code string) (bool, error) {
	key := r.authCodeKey(code)
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
