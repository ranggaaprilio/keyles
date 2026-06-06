package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadRejectsInvalidIntegerEnv(t *testing.T) {
	t.Setenv("DB_PASSWORD", "dev_password")
	t.Setenv("DB_PORT", "not-a-number")

	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "DB_PORT must be a valid integer")
}

func TestLoadProductionValidation(t *testing.T) {
	tests := []struct {
		name        string
		mutateEnv   func(t *testing.T)
		errContains string
	}{
		{
			name: "requires DB password",
			mutateEnv: func(t *testing.T) {
				t.Setenv("DB_PASSWORD", "")
			},
			errContains: "DB_PASSWORD is required",
		},
		{
			name: "requires Brevo key",
			mutateEnv: func(t *testing.T) {
				t.Setenv("BREVO_API_KEY", "")
			},
			errContains: "BREVO_API_KEY is required in production",
		},
		{
			name: "rejects default JWT secret",
			mutateEnv: func(t *testing.T) {
				t.Setenv("JWT_SECRET", "dev_jwt_secret_change_in_production")
			},
			errContains: "JWT_SECRET must be changed in production",
		},
		{
			name: "requires minimum JWT secret length",
			mutateEnv: func(t *testing.T) {
				t.Setenv("JWT_SECRET", "too-short")
			},
			errContains: "JWT_SECRET must be at least 32 characters in production",
		},
		{
			name: "requires HTTPS OAuth issuer",
			mutateEnv: func(t *testing.T) {
				t.Setenv("OAUTH_ISSUER", "http://sso.example.com")
			},
			errContains: "OAUTH_ISSUER must be an HTTPS URL in production",
		},
		{
			name: "rejects SKIP_EMAIL_VERIFICATION in production",
			mutateEnv: func(t *testing.T) {
				t.Setenv("SKIP_EMAIL_VERIFICATION", "true")
			},
			errContains: "SKIP_EMAIL_VERIFICATION cannot be enabled in production",
		},
		{
			name: "requires SECURITY_COOKIE_SECURE in production",
			mutateEnv: func(t *testing.T) {
				t.Setenv("SECURITY_COOKIE_SECURE", "false")
			},
			errContains: "SECURITY_COOKIE_SECURE must be true in production",
		},
		{
			name: "rejects non-HTTPS FRONTEND_URL in production",
			mutateEnv: func(t *testing.T) {
				t.Setenv("FRONTEND_URL", "http://frontend.example.com")
			},
			errContains: "FRONTEND_URL must use HTTPS in production",
		},
		{
			name: "rejects invalid FRONTEND_URL in production",
			mutateEnv: func(t *testing.T) {
				t.Setenv("FRONTEND_URL", "not-a-url")
			},
			errContains: "FRONTEND_URL must be a valid URL in production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setValidProductionEnv(t)
			tt.mutateEnv(t)

			_, err := Load()
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestLoadProductionValidConfig(t *testing.T) {
	setValidProductionEnv(t)

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "production", cfg.AppEnv)
}

func TestLoadBrowserFlowDefaults(t *testing.T) {
	t.Setenv("DB_PASSWORD", "test_password")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "http://localhost:3000", cfg.FrontendURL, "FrontendURL should default to http://localhost:3000")
	require.False(t, cfg.SecurityCookieSecure, "SecurityCookieSecure should default to false")
	require.Equal(t, 28800, cfg.SecuritySessionTTL, "SecuritySessionTTL should default to 28800 (8 hours)")
	require.Equal(t, 600, cfg.OAuthAuthTransactionTTL, "OAuthAuthTransactionTTL should default to 600 (10 minutes)")
	require.Equal(t, 5, cfg.RateLimitOAuthLoginFailures, "RateLimitOAuthLoginFailures should default to 5")
	require.Equal(t, 900, cfg.RateLimitOAuthLoginWindowSeconds, "RateLimitOAuthLoginWindowSeconds should default to 900 (15 minutes)")
}

func TestLoadBrowserFlowCustomValues(t *testing.T) {
	t.Setenv("DB_PASSWORD", "test_password")
	t.Setenv("FRONTEND_URL", "https://sso.example.com")
	t.Setenv("SECURITY_COOKIE_SECURE", "true")
	t.Setenv("SECURITY_SESSION_TTL", "7200")
	t.Setenv("OAUTH_AUTH_TRANSACTION_TTL", "300")
	t.Setenv("RATE_LIMIT_OAUTH_LOGIN_FAILURES", "10")
	t.Setenv("RATE_LIMIT_OAUTH_LOGIN_WINDOW_SECONDS", "1800")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "https://sso.example.com", cfg.FrontendURL)
	require.True(t, cfg.SecurityCookieSecure)
	require.Equal(t, 7200, cfg.SecuritySessionTTL)
	require.Equal(t, 300, cfg.OAuthAuthTransactionTTL)
	require.Equal(t, 10, cfg.RateLimitOAuthLoginFailures)
	require.Equal(t, 1800, cfg.RateLimitOAuthLoginWindowSeconds)
}

func TestLoadBrowserFlowInvalidIntegerEnv(t *testing.T) {
	t.Setenv("DB_PASSWORD", "test_password")

	tests := []struct {
		name        string
		envKey      string
		envValue    string
		errContains string
	}{
		{
			name:        "invalid SECURITY_SESSION_TTL",
			envKey:      "SECURITY_SESSION_TTL",
			envValue:    "not-a-number",
			errContains: "SECURITY_SESSION_TTL must be a valid integer",
		},
		{
			name:        "invalid OAUTH_AUTH_TRANSACTION_TTL",
			envKey:      "OAUTH_AUTH_TRANSACTION_TTL",
			envValue:    "abc",
			errContains: "OAUTH_AUTH_TRANSACTION_TTL must be a valid integer",
		},
		{
			name:        "invalid RATE_LIMIT_OAUTH_LOGIN_FAILURES",
			envKey:      "RATE_LIMIT_OAUTH_LOGIN_FAILURES",
			envValue:    "xyz",
			errContains: "RATE_LIMIT_OAUTH_LOGIN_FAILURES must be a valid integer",
		},
		{
			name:        "invalid RATE_LIMIT_OAUTH_LOGIN_WINDOW_SECONDS",
			envKey:      "RATE_LIMIT_OAUTH_LOGIN_WINDOW_SECONDS",
			envValue:    "nope",
			errContains: "RATE_LIMIT_OAUTH_LOGIN_WINDOW_SECONDS must be a valid integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envKey, tt.envValue)

			_, err := Load()
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestLoadBrowserFlowPositiveIntegerValidation(t *testing.T) {
	t.Setenv("DB_PASSWORD", "test_password")

	tests := []struct {
		name        string
		envKey      string
		envValue    string
		errContains string
	}{
		{
			name:        "zero SECURITY_SESSION_TTL",
			envKey:      "SECURITY_SESSION_TTL",
			envValue:    "0",
			errContains: "SECURITY_SESSION_TTL must be a positive integer",
		},
		{
			name:        "negative SECURITY_SESSION_TTL",
			envKey:      "SECURITY_SESSION_TTL",
			envValue:    "-1",
			errContains: "SECURITY_SESSION_TTL must be a positive integer",
		},
		{
			name:        "zero OAUTH_AUTH_TRANSACTION_TTL",
			envKey:      "OAUTH_AUTH_TRANSACTION_TTL",
			envValue:    "0",
			errContains: "OAUTH_AUTH_TRANSACTION_TTL must be a positive integer",
		},
		{
			name:        "zero RATE_LIMIT_OAUTH_LOGIN_FAILURES",
			envKey:      "RATE_LIMIT_OAUTH_LOGIN_FAILURES",
			envValue:    "0",
			errContains: "RATE_LIMIT_OAUTH_LOGIN_FAILURES must be a positive integer",
		},
		{
			name:        "zero RATE_LIMIT_OAUTH_LOGIN_WINDOW_SECONDS",
			envKey:      "RATE_LIMIT_OAUTH_LOGIN_WINDOW_SECONDS",
			envValue:    "0",
			errContains: "RATE_LIMIT_OAUTH_LOGIN_WINDOW_SECONDS must be a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envKey, tt.envValue)

			_, err := Load()
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestLoadRejectsInvalidCookieSecureBool(t *testing.T) {
	t.Setenv("DB_PASSWORD", "test_password")
	t.Setenv("SECURITY_COOKIE_SECURE", "not-a-bool")

	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "SECURITY_COOKIE_SECURE must be a valid boolean")
}

func TestLoadSkipEmailVerificationDefault(t *testing.T) {
	t.Setenv("DB_PASSWORD", "test_password")
	cfg, err := Load()
	require.NoError(t, err)
	require.False(t, cfg.SkipEmailVerification, "SkipEmailVerification should default to false")
}

func TestLoadSkipEmailVerificationEnabled(t *testing.T) {
	t.Setenv("DB_PASSWORD", "test_password")
	t.Setenv("SKIP_EMAIL_VERIFICATION", "true")
	cfg, err := Load()
	require.NoError(t, err)
	require.True(t, cfg.SkipEmailVerification, "SkipEmailVerification should be true when env is set")
}

func setValidProductionEnv(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", "production")
	t.Setenv("DB_PASSWORD", "prod_db_password")
	t.Setenv("BREVO_API_KEY", "brevo_api_key")
	t.Setenv("JWT_SECRET", strings.Repeat("a", 32))
	t.Setenv("OAUTH_ISSUER", "https://sso.example.com")
	t.Setenv("SECURITY_COOKIE_SECURE", "true")
	t.Setenv("FRONTEND_URL", "https://sso.example.com")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("DB_SSL_MODE", "require")
}

func TestLoadProductionRejectsDBSSLModeDisable(t *testing.T) {
	setValidProductionEnv(t)
	t.Setenv("DB_SSL_MODE", "disable")

	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "DB_SSL_MODE must be require, verify-ca, or verify-full in production")
}

func TestLoadProductionRejectsDBSSLModeAllow(t *testing.T) {
	setValidProductionEnv(t)
	t.Setenv("DB_SSL_MODE", "allow")

	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "DB_SSL_MODE must be require, verify-ca, or verify-full in production")
}

func TestLoadProductionRejectsDebugLogLevel(t *testing.T) {
	setValidProductionEnv(t)
	t.Setenv("LOG_LEVEL", "debug")

	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "LOG_LEVEL must not be debug in production")
}

func TestLoadSecurityHardeningDefaults(t *testing.T) {
	t.Setenv("DB_PASSWORD", "test_password")

	cfg, err := Load()
	require.NoError(t, err)
	require.True(t, cfg.CSRFEnabled, "CSRFEnabled should default to true")
	require.Equal(t, "keyles_csrf", cfg.CSRFCookieName)
	require.Equal(t, "X-CSRF-Token", cfg.CSRFHeaderName)
	require.Equal(t, 32, cfg.CSRFTokenLength)
	require.Equal(t, 1048576, cfg.RequestMaxBodySize)
	require.Equal(t, "15s", cfg.RequestReadTimeout)
	require.Equal(t, "15s", cfg.RequestWriteTimeout)
	require.Equal(t, "60s", cfg.RequestIdleTimeout)
	require.True(t, cfg.MetricsEnabled)
	require.Equal(t, "/metrics", cfg.MetricsPath)
	require.Equal(t, 25, cfg.DBMaxOpenConns)
	require.Equal(t, 10, cfg.DBMaxIdleConns)
	require.Equal(t, "5m", cfg.DBConnMaxLifetime)
	require.Equal(t, "30s", cfg.DBStatementTimeout)
	require.Equal(t, 3, cfg.RateLimitRegisterPerHour)
	require.Equal(t, 5, cfg.RateLimitVerifyOTPPer10Min)
	require.Equal(t, 3, cfg.RateLimitResendOTPPerHour)
}