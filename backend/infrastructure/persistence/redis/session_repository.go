package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// RedisSessionRepository implements SessionRepository using Redis
type RedisSessionRepository struct {
	client *redis.Client
}

// NewRedisSessionRepository creates a new Redis session repository
func NewRedisSessionRepository(client *redis.Client) repositories.SessionRepository {
	return &RedisSessionRepository{client: client}
}

// sessionKey generates the Redis key for a session
func (r *RedisSessionRepository) sessionKey(sessionID string) string {
	return fmt.Sprintf("oauth:session:%s", sessionID)
}

// Create stores a new session with automatic expiration
func (r *RedisSessionRepository) Create(ctx context.Context, session *repositories.Session, ttl time.Duration) error {
	key := r.sessionKey(session.SessionID)

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

// Get retrieves a session by session ID
func (r *RedisSessionRepository) Get(ctx context.Context, sessionID string) (*repositories.Session, error) {
	key := r.sessionKey(sessionID)

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Session not found or expired
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session repositories.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// Delete removes a session (logout)
func (r *RedisSessionRepository) Delete(ctx context.Context, sessionID string) error {
	key := r.sessionKey(sessionID)
	return r.client.Del(ctx, key).Err()
}

// Exists checks if a session exists and is valid
func (r *RedisSessionRepository) Exists(ctx context.Context, sessionID string) (bool, error) {
	key := r.sessionKey(sessionID)
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Extend extends the TTL of an existing session
func (r *RedisSessionRepository) Extend(ctx context.Context, sessionID string, ttl time.Duration) error {
	key := r.sessionKey(sessionID)
	return r.client.Expire(ctx, key, ttl).Err()
}
