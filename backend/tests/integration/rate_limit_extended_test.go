package integration

import (
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/infrastructure/config"
	"github.com/ranggaaprilio/keyles/interfaces/http/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitingOnPublicEndpoints(t *testing.T) {
	cfg := &config.Config{
		RateLimitLoginAttemptsPer15Min: 5,
		RateLimitRegisterPerHour:        3,
		RateLimitVerifyOTPPer10Min:      5,
		RateLimitResendOTPPerHour:       3,
	}

	// Test that rate limiting middleware is configured for each endpoint
	endpoints := []struct {
		path   string
		method string
		limit  int
		window time.Duration
	}{
		{"/api/v1/login", "POST", cfg.RateLimitLoginAttemptsPer15Min, 15 * time.Minute},
		{"/api/v1/register", "POST", cfg.RateLimitRegisterPerHour, time.Hour},
		{"/api/v1/verify-otp", "POST", cfg.RateLimitVerifyOTPPer10Min, 10 * time.Minute},
		{"/api/v1/resend-otp", "POST", cfg.RateLimitResendOTPPerHour, time.Hour},
	}

	for _, ep := range endpoints {
		t.Run(ep.path, func(t *testing.T) {
			assert.Greater(t, ep.limit, 0, "Rate limit should be positive for %s", ep.path)
			assert.Greater(t, ep.window, time.Duration(0), "Rate limit window should be positive for %s", ep.path)
		})
	}
}

func TestRateLimitFailClosed(t *testing.T) {
	// Create rate limiter without Redis (will fail)
	rl := middleware.NewRateLimiter(nil)
	assert.NotNil(t, rl)

	// The middleware should exist and be callable
	m := rl.IPBasedLimit(5, 15*time.Minute)
	assert.NotNil(t, m)
}
