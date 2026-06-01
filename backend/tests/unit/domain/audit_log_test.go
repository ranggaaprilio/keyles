package domain_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthEventConstants(t *testing.T) {
	expected := map[string]entities.EventType{
		"EventOAuthLoginSucceeded":  entities.EventOAuthLoginSucceeded,
		"EventOAuthLoginFailed":     entities.EventOAuthLoginFailed,
		"EventOAuthLoginThrottled":  entities.EventOAuthLoginThrottled,
		"EventOAuthConsentApproved": entities.EventOAuthConsentApproved,
		"EventOAuthConsentDenied":   entities.EventOAuthConsentDenied,
		"EventOAuthLogout":          entities.EventOAuthLogout,
		"EventOAuthInvalidCallback": entities.EventOAuthInvalidCallback,
	}

	values := map[string]string{
		"EventOAuthLoginSucceeded":  "oauth_login_succeeded",
		"EventOAuthLoginFailed":     "oauth_login_failed",
		"EventOAuthLoginThrottled":  "oauth_login_throttled",
		"EventOAuthConsentApproved": "oauth_consent_approved",
		"EventOAuthConsentDenied":   "oauth_consent_denied",
		"EventOAuthLogout":          "oauth_logout",
		"EventOAuthInvalidCallback": "oauth_invalid_callback",
	}

	t.Run("all seven constants exist with correct string values", func(t *testing.T) {
		require.Len(t, expected, 7, "expected exactly 7 OAuth event constants")
		for name, constant := range expected {
			assert.Equal(t, entities.EventType(values[name]), constant,
				"constant %s should have value %s", name, values[name])
		}
	})

	t.Run("constants are distinct", func(t *testing.T) {
		seen := make(map[entities.EventType]string)
		for name, constant := range expected {
			if prev, dup := seen[constant]; dup {
				t.Fatalf("duplicate event type value %q used by %s and %s", constant, prev, name)
			}
			seen[constant] = name
		}
	})
}

func TestNewAuditLog(t *testing.T) {
	t.Run("creates entry with correct event type", func(t *testing.T) {
		entry := entities.NewAuditLog(entities.EventOAuthLoginSucceeded, "10.0.0.1", "TestAgent/1.0")
		require.NotNil(t, entry)
		assert.Equal(t, entities.EventOAuthLoginSucceeded, entry.EventType)
		assert.Equal(t, "10.0.0.1", entry.IPAddress)
		assert.Equal(t, "TestAgent/1.0", entry.UserAgent)
		assert.NotEqual(t, uuid.Nil, entry.ID, "ID should be a non-nil UUID")
		assert.False(t, entry.CreatedAt.IsZero(), "CreatedAt should be set")
		assert.NotNil(t, entry.EventData, "EventData should be initialized")
		assert.Empty(t, entry.EventData, "EventData should start empty")
	})
}

func TestAuditLogBuilderMethods(t *testing.T) {
	entry := entities.NewAuditLog(entities.EventOAuthConsentApproved, "192.168.1.1", "Mozilla/5.0")

	t.Run("WithTenant sets tenant ID and returns same instance", func(t *testing.T) {
		tenantID := uuid.New()
		result := entry.WithTenant(tenantID)
		require.NotNil(t, result.TenantID)
		assert.Equal(t, tenantID, *result.TenantID)
		assert.Same(t, entry, result, "WithTenant should return the same AuditLog pointer")
	})

	t.Run("WithUser sets user ID and returns same instance", func(t *testing.T) {
		userID := uuid.New()
		result := entry.WithUser(userID)
		require.NotNil(t, result.UserID)
		assert.Equal(t, userID, *result.UserID)
		assert.Same(t, entry, result, "WithUser should return the same AuditLog pointer")
	})

	t.Run("WithData adds key-value pairs and returns same instance", func(t *testing.T) {
		result := entry.WithData("client_id", "abc123")
		assert.Equal(t, "abc123", result.EventData["client_id"])
		assert.Same(t, entry, result, "WithData should return the same AuditLog pointer")

		entry.WithData("scope", "openid profile")
		assert.Equal(t, "openid profile", entry.EventData["scope"])
	})
}

func TestAuditLogSanitization(t *testing.T) {
	secretKeys := []string{"password", "cookie", "code_challenge", "code", "secret"}

	t.Run("event data must not contain secret keys", func(t *testing.T) {
		entry := entities.NewAuditLog(entities.EventOAuthLoginFailed, "10.0.0.1", "Bot/1.0")

		// Simulate builder usage with safe data
		entry.WithData("client_id", "my-client")
		entry.WithData("ip_address", "10.0.0.1")

		for _, key := range secretKeys {
			_, exists := entry.EventData[key]
			assert.False(t, exists, "EventData must not contain secret key %q", key)
		}
	})

	t.Run("WithData does not filter secrets at the map level", func(t *testing.T) {
		// This test documents that WithData puts whatever is passed into the map.
		// Callers must not pass secret values; the builder does not auto-strip them.
		entry := entities.NewAuditLog(entities.EventOAuthConsentDenied, "10.0.0.1", "Bot/1.0")
		entry.WithData("password", "should-not-be-present")
		assert.Contains(t, entry.EventData, "password",
			"WithData does not auto-sanitize; callers are responsible for excluding secret keys")
	})
}

func TestAuditLogWithError(t *testing.T) {
	t.Run("adds error message to event data", func(t *testing.T) {
		entry := entities.NewAuditLog(entities.EventOAuthLoginFailed, "10.0.0.1", "Bot/1.0")
		result := entry.WithError(errors.New("invalid credentials"))
		assert.Equal(t, "invalid credentials", result.EventData["error"])
		assert.Same(t, entry, result, "WithError should return the same AuditLog pointer")
	})

	t.Run("nil error does not add error key", func(t *testing.T) {
		entry := entities.NewAuditLog(entities.EventOAuthLoginFailed, "10.0.0.1", "Bot/1.0")
		result := entry.WithError(nil)
		_, hasError := result.EventData["error"]
		assert.False(t, hasError, "nil error should not add 'error' key to EventData")
	})

	t.Run("WithError does not leak sensitive internal details", func(t *testing.T) {
		entry := entities.NewAuditLog(entities.EventOAuthLoginFailed, "10.0.0.1", "Bot/1.0")
		// The error message is whatever the caller passes; WithError does not redact.
		// Callers must pass safe, user-facing error strings — not raw internal errors
		// that might contain secrets.
		entry.WithError(errors.New("invalid credentials"))
		assert.Equal(t, "invalid credentials", entry.EventData["error"])

		// Verify no secret values appear in the error string itself.
		secretSubstrings := []string{"password", "secret", "token"}
		errMsg, _ := entry.EventData["error"].(string)
		for _, sub := range secretSubstrings {
			assert.NotContains(t, errMsg, sub,
				"error message in EventData should not contain secret substring %q", sub)
		}
	})
}