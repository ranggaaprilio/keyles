package integration

import (
	"os"
	"testing"

	"github.com/ranggaaprilio/keyles/infrastructure/config"
	"github.com/stretchr/testify/assert"
)

// TestConfigRejectsDefaultSecretsInProduction verifies that the backend
// refuses to start with default/weak secrets when APP_ENV=production
func TestConfigRejectsDefaultSecretsInProduction(t *testing.T) {
	// Save and restore original env vars
	originalEnv := make(map[string]string)
	for _, key := range []string{
		"APP_ENV", "JWT_SECRET", "DB_PASSWORD", "DB_SSL_MODE",
		"BREVO_API_KEY", "SECURITY_COOKIE_SECURE", "OAUTH_ISSUER",
		"FRONTEND_URL", "LOG_LEVEL",
	} {
		originalEnv[key] = os.Getenv(key)
	}
	defer func() {
		for key, val := range originalEnv {
			if val == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, val)
			}
		}
	}()

	// Set production environment
	os.Setenv("APP_ENV", "production")
	os.Setenv("DB_PASSWORD", "strong_password_123")
	os.Setenv("BREVO_API_KEY", "test-api-key")
	os.Setenv("SECURITY_COOKIE_SECURE", "true")
	os.Setenv("OAUTH_ISSUER", "https://sso.keyles.com")
	os.Setenv("FRONTEND_URL", "https://app.keyles.com")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("DB_SSL_MODE", "require")

	// Test 1: Default JWT secret should be rejected
	os.Setenv("JWT_SECRET", "dev_jwt_secret_change_in_production")
	_, err := config.Load()
	assert.Error(t, err, "Should reject default JWT_SECRET in production")
	assert.Contains(t, err.Error(), "JWT_SECRET")

	// Test 2: Short JWT secret should be rejected
	os.Setenv("JWT_SECRET", "short")
	_, err = config.Load()
	assert.Error(t, err, "Should reject short JWT_SECRET in production")
	assert.Contains(t, err.Error(), "JWT_SECRET")

	// Test 3: Valid JWT secret should pass (with other valid values)
	os.Setenv("JWT_SECRET", "this_is_a_very_long_and_secure_jwt_secret_key_12345")
	cfg, err := config.Load()
	assert.NoError(t, err, "Should accept valid configuration in production")
	assert.NotNil(t, cfg)

	// Test 4: DB_SSL_MODE=disable should be rejected
	os.Setenv("JWT_SECRET", "this_is_a_very_long_and_secure_jwt_secret_key_12345")
	os.Setenv("DB_SSL_MODE", "disable")
	_, err = config.Load()
	assert.Error(t, err, "Should reject DB_SSL_MODE=disable in production")
	assert.Contains(t, err.Error(), "DB_SSL_MODE")

	// Test 5: DB_SSL_MODE=allow should be rejected
	os.Setenv("DB_SSL_MODE", "allow")
	_, err = config.Load()
	assert.Error(t, err, "Should reject DB_SSL_MODE=allow in production")
	assert.Contains(t, err.Error(), "DB_SSL_MODE")

	// Test 6: LOG_LEVEL=debug should be rejected
	os.Setenv("DB_SSL_MODE", "require")
	os.Setenv("LOG_LEVEL", "debug")
	_, err = config.Load()
	assert.Error(t, err, "Should reject LOG_LEVEL=debug in production")
	assert.Contains(t, err.Error(), "LOG_LEVEL")
}
