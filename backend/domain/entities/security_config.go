package entities

import (
	"fmt"
	"net/url"
	"strings"
)

// SecurityConfig holds production security validation rules
type SecurityConfig struct {
	AppEnv               string
	JWTSecret            string
	DBPassword           string
	DBSSLMode            string
	BrevoAPIKey          string
	SecurityCookieSecure bool
	OAuthIssuer          string
	FrontendURL          string
	LogLevel             string
}

// ValidateForProduction checks that all security requirements are met for production deployment
func (sc *SecurityConfig) ValidateForProduction() error {
	if sc.AppEnv != "production" {
		return nil
	}

	if sc.JWTSecret == "dev_jwt_secret_change_in_production" {
		return fmt.Errorf("JWT_SECRET must be changed from default value in production")
	}
	if len(sc.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters in production")
	}
	if sc.DBPassword == "" {
		return fmt.Errorf("DB_PASSWORD is required in production")
	}
	if sc.DBSSLMode == "disable" || sc.DBSSLMode == "allow" {
		return fmt.Errorf("DB_SSL_MODE must be require, verify-ca, or verify-full in production")
	}
	if sc.BrevoAPIKey == "" {
		return fmt.Errorf("BREVO_API_KEY is required in production")
	}
	if !sc.SecurityCookieSecure {
		return fmt.Errorf("SECURITY_COOKIE_SECURE must be true in production")
	}
	issuer, err := url.Parse(sc.OAuthIssuer)
	if err != nil || issuer.Scheme != "https" || issuer.Host == "" {
		return fmt.Errorf("OAUTH_ISSUER must be an HTTPS URL in production")
	}
	frontendURL, err := url.Parse(sc.FrontendURL)
	if err != nil || frontendURL.Scheme == "" || frontendURL.Host == "" {
		return fmt.Errorf("FRONTEND_URL must be a valid URL in production")
	}
	if frontendURL.Scheme != "https" {
		return fmt.Errorf("FRONTEND_URL must use HTTPS in production")
	}
	if strings.ToLower(sc.LogLevel) == "debug" {
		return fmt.Errorf("LOG_LEVEL must not be debug in production")
	}

	return nil
}
