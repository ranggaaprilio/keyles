package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// RedisOTPRepository implements OTPRepository using Redis
type RedisOTPRepository struct {
	client *redis.Client
}

// NewRedisOTPRepository creates a new Redis OTP repository
func NewRedisOTPRepository(client *redis.Client) repositories.OTPRepository {
	return &RedisOTPRepository{client: client}
}

// Create stores an OTP verification record with TTL
func (r *RedisOTPRepository) Create(ctx context.Context, otp *entities.OTPVerification) error {
	key := r.otpKey(otp.TenantID)
	
	data, err := json.Marshal(otp)
	if err != nil {
		return fmt.Errorf("failed to marshal OTP: %w", err)
	}

	ttl := time.Until(otp.ExpiresAt)
	if ttl <= 0 {
		ttl = entities.OTPExpirationMins * time.Minute
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

// FindByTenantID retrieves the active OTP for a tenant
func (r *RedisOTPRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID) (*entities.OTPVerification, error) {
	key := r.otpKey(tenantID.String())
	
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // OTP not found or expired
		}
		return nil, fmt.Errorf("failed to get OTP: %w", err)
	}

	var otp entities.OTPVerification
	if err := json.Unmarshal(data, &otp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OTP: %w", err)
	}

	// Check if expired
	if otp.IsExpired() {
		r.Delete(ctx, tenantID) // Clean up expired OTP
		return nil, nil
	}

	return &otp, nil
}

// FindByTenantIDAndPurpose retrieves the OTP for a tenant and purpose
func (r *RedisOTPRepository) FindByTenantIDAndPurpose(ctx context.Context, tenantID, purpose string) (*entities.OTPVerification, error) {
	key := r.otpKey(tenantID)
	
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // OTP not found or expired
		}
		return nil, fmt.Errorf("failed to get OTP: %w", err)
	}

	var otp entities.OTPVerification
	if err := json.Unmarshal(data, &otp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OTP: %w", err)
	}

	// Check if purpose matches
	if otp.Purpose != purpose {
		return nil, nil // Wrong purpose
	}

	// Check if expired
	if otp.IsExpired() {
		r.Delete(ctx, otp.ID) // Clean up expired OTP
		return nil, nil
	}

	return &otp, nil
}

// DeleteExpired removes all expired OTP records (no-op for Redis with TTL)
func (r *RedisOTPRepository) DeleteExpired(ctx context.Context) error {
	// Redis automatically expires keys with TTL, so this is a no-op
	return nil
}

// Update updates an OTP verification record
func (r *RedisOTPRepository) Update(ctx context.Context, otp *entities.OTPVerification) error {
	// For Redis, update is the same as create (overwrites with same key)
	return r.Create(ctx, otp)
}

// Delete removes an OTP verification record
func (r *RedisOTPRepository) Delete(ctx context.Context, id uuid.UUID) error {
	key := r.otpKey(id.String())
	return r.client.Del(ctx, key).Err()
}

// IncrementRateLimitCounter increments the OTP request counter for an email
func (r *RedisOTPRepository) IncrementRateLimitCounter(ctx context.Context, email string, window time.Duration) (int, error) {
	key := r.rateLimitKey(email)
	
	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("failed to increment rate limit: %w", err)
	}

	return int(incr.Val()), nil
}

// GetRateLimitCounter gets the current OTP request count for an email
func (r *RedisOTPRepository) GetRateLimitCounter(ctx context.Context, email string) (int, error) {
	key := r.rateLimitKey(email)
	
	count, err := r.client.Get(ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // No requests yet
		}
		return 0, fmt.Errorf("failed to get rate limit: %w", err)
	}

	return count, nil
}

// otpKey generates the Redis key for OTP storage
func (r *RedisOTPRepository) otpKey(tenantID string) string {
	return fmt.Sprintf("otp:tenant:%s", tenantID)
}

// rateLimitKey generates the Redis key for rate limiting
func (r *RedisOTPRepository) rateLimitKey(email string) string {
	return fmt.Sprintf("ratelimit:otp:%s", email)
}
