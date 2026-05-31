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

func setValidProductionEnv(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", "production")
	t.Setenv("DB_PASSWORD", "prod_db_password")
	t.Setenv("BREVO_API_KEY", "brevo_api_key")
	t.Setenv("JWT_SECRET", strings.Repeat("a", 32))
	t.Setenv("OAUTH_ISSUER", "https://sso.example.com")
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
