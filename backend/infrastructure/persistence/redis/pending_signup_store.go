package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// RedisPendingSignupStore implements repositories.PendingSignupStore.
// Keys embed a SHA-256 of the email so addresses never appear in key names.
type RedisPendingSignupStore struct {
	client *redis.Client
}

// NewRedisPendingSignupStore creates a new Redis-backed pending sign-up store.
func NewRedisPendingSignupStore(client *redis.Client) repositories.PendingSignupStore {
	return &RedisPendingSignupStore{client: client}
}

// verifySignupScript compares the code hash. On match it returns the record
// and deletes it; on mismatch it increments attempts and deletes the record
// once ARGV[2] attempts are reached.
var verifySignupScript = redis.NewScript(`
local key = KEYS[1]
local stored = redis.call("HGET", key, "code_hash")
if not stored then
	return false
end
if stored == ARGV[1] then
	local all = redis.call("HGETALL", key)
	redis.call("DEL", key)
	return all
end
local attempts = redis.call("HINCRBY", key, "attempts", 1)
if attempts >= tonumber(ARGV[2]) then
	redis.call("DEL", key)
end
return false
`)

func emailDigest(email string) string {
	digest := sha256.Sum256([]byte(email))
	return hex.EncodeToString(digest[:])
}

func (s *RedisPendingSignupStore) signupKey(tenantID, email string) string {
	return fmt.Sprintf("oauth:signup:%s:%s", tenantID, emailDigest(email))
}

func (s *RedisPendingSignupStore) cooldownKey(tenantID, email string) string {
	return fmt.Sprintf("oauth:signup-cooldown:%s:%s", tenantID, emailDigest(email))
}

func (s *RedisPendingSignupStore) sendCountKey(tenantID, email string) string {
	return fmt.Sprintf("oauth:signup-sends:%s:%s", tenantID, emailDigest(email))
}

// AllowSend enforces a per-email cooldown and an hourly send cap.
func (s *RedisPendingSignupStore) AllowSend(ctx context.Context, tenantID, email string, cooldown time.Duration, maxPerHour int) (bool, error) {
	ok, err := s.client.SetNX(ctx, s.cooldownKey(tenantID, email), 1, cooldown).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check signup cooldown: %w", err)
	}
	if !ok {
		return false, nil
	}

	count, err := redis.NewScript(recordFailureScript).Run(ctx, s.client,
		[]string{s.sendCountKey(tenantID, email)}, 1, int(time.Hour.Seconds())).Int()
	if err != nil {
		return false, fmt.Errorf("failed to count signup sends: %w", err)
	}
	return count <= maxPerHour, nil
}

// Save stores (or replaces) a pending sign-up and resets its attempt counter.
func (s *RedisPendingSignupStore) Save(ctx context.Context, signup *repositories.PendingSignup, ttl time.Duration) error {
	key := s.signupKey(signup.TenantID, signup.Email)
	_, err := s.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, key)
		pipe.HSet(ctx, key, map[string]interface{}{
			"tenant_id":      signup.TenantID,
			"transaction_id": signup.TransactionID,
			"email":          signup.Email,
			"display_name":   signup.DisplayName,
			"password_hash":  signup.PasswordHash,
			"code_hash":      signup.CodeHash,
			"attempts":       0,
		})
		pipe.Expire(ctx, key, ttl)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to store pending signup: %w", err)
	}
	return nil
}

// Get returns the pending sign-up, or nil, nil if none exists.
func (s *RedisPendingSignupStore) Get(ctx context.Context, tenantID, email string) (*repositories.PendingSignup, error) {
	fields, err := s.client.HGetAll(ctx, s.signupKey(tenantID, email)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get pending signup: %w", err)
	}
	if len(fields) == 0 {
		return nil, nil
	}
	return fieldsToPendingSignup(fields), nil
}

// Verify atomically checks the code hash; see repositories.PendingSignupStore.
func (s *RedisPendingSignupStore) Verify(ctx context.Context, tenantID, email, codeHash string, maxAttempts int) (*repositories.PendingSignup, error) {
	result, err := verifySignupScript.Run(ctx, s.client,
		[]string{s.signupKey(tenantID, email)}, codeHash, maxAttempts).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to verify pending signup: %w", err)
	}
	flat, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected verify result type: %T", result)
	}
	return fieldsToPendingSignup(sliceToMap(flat)), nil
}

// Delete removes a pending sign-up.
func (s *RedisPendingSignupStore) Delete(ctx context.Context, tenantID, email string) error {
	if err := s.client.Del(ctx, s.signupKey(tenantID, email)).Err(); err != nil {
		return fmt.Errorf("failed to delete pending signup: %w", err)
	}
	return nil
}

func fieldsToPendingSignup(fields map[string]string) *repositories.PendingSignup {
	return &repositories.PendingSignup{
		TenantID:      fields["tenant_id"],
		TransactionID: fields["transaction_id"],
		Email:         fields["email"],
		DisplayName:   fields["display_name"],
		PasswordHash:  fields["password_hash"],
		CodeHash:      fields["code_hash"],
	}
}
