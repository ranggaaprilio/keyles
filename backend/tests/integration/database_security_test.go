package integration

import (
	"os"
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/infrastructure/config"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseConnectionPoolConfig(t *testing.T) {
	// Save and restore env vars
	originalEnv := make(map[string]string)
	for _, key := range []string{
		"APP_ENV", "DB_PASSWORD", "DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS",
		"DB_CONN_MAX_LIFETIME", "DB_STATEMENT_TIMEOUT", "DB_SSL_MODE",
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

	os.Setenv("DB_PASSWORD", "test_password")
	os.Setenv("DB_MAX_OPEN_CONNS", "25")
	os.Setenv("DB_MAX_IDLE_CONNS", "10")
	os.Setenv("DB_CONN_MAX_LIFETIME", "5m")
	os.Setenv("DB_STATEMENT_TIMEOUT", "30s")

	cfg, err := config.Load()
	assert.NoError(t, err)
	assert.Equal(t, 25, cfg.DBMaxOpenConns)
	assert.Equal(t, 10, cfg.DBMaxIdleConns)
	assert.Equal(t, "5m", cfg.DBConnMaxLifetime)
	assert.Equal(t, "30s", cfg.DBStatementTimeout)
}

func TestDatabaseSSLModeProduction(t *testing.T) {
	originalEnv := make(map[string]string)
	for _, key := range []string{
		"APP_ENV", "DB_PASSWORD", "BREVO_API_KEY", "JWT_SECRET",
		"OAUTH_ISSUER", "SECURITY_COOKIE_SECURE", "FRONTEND_URL",
		"LOG_LEVEL", "DB_SSL_MODE",
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

	os.Setenv("APP_ENV", "production")
	os.Setenv("DB_PASSWORD", "strong_password")
	os.Setenv("BREVO_API_KEY", "test_key")
	os.Setenv("JWT_SECRET", "this_is_a_very_long_and_secure_jwt_secret_key_12345")
	os.Setenv("OAUTH_ISSUER", "https://sso.keyles.com")
	os.Setenv("SECURITY_COOKIE_SECURE", "true")
	os.Setenv("FRONTEND_URL", "https://app.keyles.com")
	os.Setenv("LOG_LEVEL", "info")

	// Test require SSL mode passes
	os.Setenv("DB_SSL_MODE", "require")
	cfg, err := config.Load()
	assert.NoError(t, err)
	assert.Equal(t, "require", cfg.DBSSLMode)

	// Test verify-ca SSL mode passes
	os.Setenv("DB_SSL_MODE", "verify-ca")
	cfg, err = config.Load()
	assert.NoError(t, err)
	assert.Equal(t, "verify-ca", cfg.DBSSLMode)

	// Test verify-full SSL mode passes
	os.Setenv("DB_SSL_MODE", "verify-full")
	cfg, err = config.Load()
	assert.NoError(t, err)
	assert.Equal(t, "verify-full", cfg.DBSSLMode)
}

func TestDatabaseConnectionPoolDefaults(t *testing.T) {
	originalEnv := make(map[string]string)
	for _, key := range []string{"DB_PASSWORD"} {
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

	os.Setenv("DB_PASSWORD", "test_password")

	cfg, err := config.Load()
	assert.NoError(t, err)
	assert.Equal(t, 25, cfg.DBMaxOpenConns, "Default max open conns should be 25")
	assert.Equal(t, 10, cfg.DBMaxIdleConns, "Default max idle conns should be 10")
	assert.Equal(t, "5m", cfg.DBConnMaxLifetime, "Default conn max lifetime should be 5m")
	assert.Equal(t, "30s", cfg.DBStatementTimeout, "Default statement timeout should be 30s")
}

func TestDatabaseStatementTimeout(t *testing.T) {
	// Verify statement timeout config is parsed correctly
	originalEnv := make(map[string]string)
	for _, key := range []string{"DB_PASSWORD", "DB_STATEMENT_TIMEOUT"} {
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

	os.Setenv("DB_PASSWORD", "test_password")
	os.Setenv("DB_STATEMENT_TIMEOUT", "60s")

	cfg, err := config.Load()
	assert.NoError(t, err)
	assert.Equal(t, "60s", cfg.DBStatementTimeout)

	// Verify it can be parsed as duration
	duration, err := time.ParseDuration(cfg.DBStatementTimeout)
	assert.NoError(t, err)
	assert.Equal(t, 60*time.Second, duration)
}
