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

// sessionData is the on-wire shape stored in Redis.
// All time.Time fields are represented as RFC3339Nano strings so that
// nanosecond precision survives JSON round-trips deterministically.
// No ClientID field — sessions are client-agnostic; a single session
// may service multiple OAuth clients within the same tenant.
type sessionData struct {
	SessionID       string                 `json:"session_id"`
	UserID          string                 `json:"user_id"`
	TenantID        string                 `json:"tenant_id"`
	AuthenticatedAt string                 `json:"authenticated_at"`
	CreatedAt       string                 `json:"created_at"`
	ExpiresAt       string                 `json:"expires_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

func marshalSession(s *repositories.Session) ([]byte, error) {
	d := sessionData{
		SessionID:       s.SessionID,
		UserID:          s.UserID,
		TenantID:        s.TenantID,
		AuthenticatedAt: s.AuthenticatedAt.Format(time.RFC3339Nano),
		CreatedAt:       s.CreatedAt.Format(time.RFC3339Nano),
		ExpiresAt:       s.ExpiresAt.Format(time.RFC3339Nano),
		Metadata:        s.Metadata,
	}
	data, err := json.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}
	return data, nil
}

func unmarshalSession(data []byte) (*repositories.Session, error) {
	var d sessionData
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	authenticatedAt, err := time.Parse(time.RFC3339Nano, d.AuthenticatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse authenticated_at: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, d.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	expiresAt, err := time.Parse(time.RFC3339Nano, d.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expires_at: %w", err)
	}

	return &repositories.Session{
		SessionID:       d.SessionID,
		UserID:          d.UserID,
		TenantID:        d.TenantID,
		AuthenticatedAt: authenticatedAt,
		CreatedAt:       createdAt,
		ExpiresAt:       expiresAt,
		Metadata:        d.Metadata,
	}, nil
}

// sessionKey generates the Redis key for a session
func (r *RedisSessionRepository) sessionKey(sessionID string) string {
	return fmt.Sprintf("oauth:session:%s", sessionID)
}

// Create stores a new session with automatic expiration
func (r *RedisSessionRepository) Create(ctx context.Context, session *repositories.Session, ttl time.Duration) error {
	key := r.sessionKey(session.SessionID)

	data, err := marshalSession(session)
	if err != nil {
		return err
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

	return unmarshalSession(data)
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