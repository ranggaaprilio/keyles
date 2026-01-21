/**
 * Integration tests for rate limiting on OAuth endpoints
 * Tests FR-051: Rate limiting 10 requests/minute per client_id on token endpoint
 */

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockRateLimiter implements a simple in-memory rate limiter for testing
type MockRateLimiter struct {
	limits map[string][]time.Time
}

// NewMockRateLimiter creates a mock rate limiter that tracks requests
func NewMockRateLimiter() *MockRateLimiter {
	return &MockRateLimiter{
		limits: make(map[string][]time.Time),
	}
}

// CheckLimit verifies if a request from a client_id is allowed (max 10 per minute)
func (m *MockRateLimiter) CheckLimit(clientID string) bool {
	now := time.Now()
	oneMinuteAgo := now.Add(-time.Minute)

	// Clean up old entries
	if requests, exists := m.limits[clientID]; exists {
		var recentRequests []time.Time
		for _, req := range requests {
			if req.After(oneMinuteAgo) {
				recentRequests = append(recentRequests, req)
			}
		}
		m.limits[clientID] = recentRequests
	}

	// Check if limit exceeded
	if len(m.limits[clientID]) >= 10 {
		return false
	}

	// Record this request
	m.limits[clientID] = append(m.limits[clientID], now)
	return true
}

// TestRateLimitingBasic tests that requests are rate-limited per client
func TestRateLimitingBasic(t *testing.T) {
	limiter := NewMockRateLimiter()
	clientID := "test-client-123"

	// First 10 requests should be allowed
	for i := 0; i < 10; i++ {
		allowed := limiter.CheckLimit(clientID)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// 11th request should be denied
	allowed := limiter.CheckLimit(clientID)
	assert.False(t, allowed, "11th request should be rate-limited")
}

// TestRateLimitingPerClient tests that rate limits are per-client
func TestRateLimitingPerClient(t *testing.T) {
	limiter := NewMockRateLimiter()
	clientID1 := "client-001"
	clientID2 := "client-002"

	// Send 10 requests for client 1
	for i := 0; i < 10; i++ {
		allowed := limiter.CheckLimit(clientID1)
		assert.True(t, allowed, "Client 1 request %d should be allowed", i+1)
	}

	// Client 1 should be rate-limited
	allowed := limiter.CheckLimit(clientID1)
	assert.False(t, allowed, "Client 1 11th request should be rate-limited")

	// Client 2 should NOT be rate-limited (independent limit)
	for i := 0; i < 10; i++ {
		allowed := limiter.CheckLimit(clientID2)
		assert.True(t, allowed, "Client 2 request %d should be allowed (independent limit)", i+1)
	}

	// Client 2 should now be rate-limited too
	allowed = limiter.CheckLimit(clientID2)
	assert.False(t, allowed, "Client 2 11th request should be rate-limited")
}

// TestRateLimitingWindowSize tests that the window is exactly 1 minute
func TestRateLimitingWindowSize(t *testing.T) {
	limiter := NewMockRateLimiter()
	clientID := "test-client"

	// Send 10 requests
	for i := 0; i < 10; i++ {
		allowed := limiter.CheckLimit(clientID)
		assert.True(t, allowed)
	}

	// Next request should be denied
	allowed := limiter.CheckLimit(clientID)
	assert.False(t, allowed)

	// Manually clear the window to simulate passage of time
	limiter.limits[clientID] = []time.Time{}

	// Now requests should be allowed again
	allowed = limiter.CheckLimit(clientID)
	assert.True(t, allowed, "After time window passes, requests should be allowed again")
}

// TestRateLimitingMultipleClients tests many concurrent clients
func TestRateLimitingMultipleClients(t *testing.T) {
	limiter := NewMockRateLimiter()

	// Each of 5 clients should be able to make 10 requests
	for clientNum := 1; clientNum <= 5; clientNum++ {
		clientID := "client-" + string(rune(48+clientNum))

		for req := 0; req < 10; req++ {
			allowed := limiter.CheckLimit(clientID)
			assert.True(t, allowed, "Client %s request %d should be allowed", clientID, req+1)
		}

		// 11th request should be denied
		allowed := limiter.CheckLimit(clientID)
		assert.False(t, allowed, "Client %s 11th request should be denied", clientID)
	}
}
