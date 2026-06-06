package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration
type Config struct {
	// Database
	DBHost     string
	DBPort     int
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	// Redis
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int

	// Brevo (Email)
	BrevoAPIKey      string
	BrevoSenderEmail string
	BrevoSenderName  string

	// JWT
	JWTSecret          string
	JWTExpirationHours int

	// Server
	ServerPort string
	ServerHost string
	GinMode    string

	// CORS
	CORSAllowedOrigins string
	CORSAllowedMethods string
	CORSAllowedHeaders string

	// Rate Limiting
	RateLimitOTPRequestsPerHour    int
	RateLimitOTPAttemptsPerOTP     int
	RateLimitLoginAttemptsPer15Min int

	// OAuth
	OAuthIssuer          string
	OAuthAccessTokenTTL  int
	OAuthRefreshTokenTTL int

	// OTP
	OTPExpirationMinutes int
	OTPLength            int
	SkipEmailVerification bool

	// Browser-flow OAuth configuration
	FrontendURL                     string
	SecurityCookieSecure            bool
	SecuritySessionTTL              int // seconds; default 28800 (8 hours)
	OAuthAuthTransactionTTL         int // seconds; default 600 (10 minutes)
	RateLimitOAuthLoginFailures     int // max failures per 15-min window; default 5
	RateLimitOAuthLoginWindowSeconds int // fixed window duration in seconds; default 900

	// Application
	AppEnv   string
	LogLevel string

	// CSRF Protection
	CSRFEnabled     bool
	CSRFCookieName  string
	CSRFHeaderName  string
	CSRFTokenLength int

	// Security Headers
	SecurityHeadersCSP  string
	SecurityHeadersHSTS string

	// Request Limits & Timeouts
	RequestMaxBodySize  int
	RequestReadTimeout  string
	RequestWriteTimeout string
	RequestIdleTimeout  string

	// Metrics
	MetricsEnabled bool
	MetricsPath    string

	// Database Connection Pool
	DBMaxOpenConns     int
	DBMaxIdleConns     int
	DBConnMaxLifetime  string
	DBStatementTimeout string

	// Per-Endpoint Rate Limiting
	RateLimitRegisterPerHour   int
	RateLimitVerifyOTPPer10Min int
	RateLimitResendOTPPerHour  int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	dbPort, err := getEnvAsInt("DB_PORT", 5432)
	if err != nil {
		return nil, err
	}
	redisPort, err := getEnvAsInt("REDIS_PORT", 6379)
	if err != nil {
		return nil, err
	}
	redisDB, err := getEnvAsInt("REDIS_DB", 0)
	if err != nil {
		return nil, err
	}
	jwtExpirationHours, err := getEnvAsInt("JWT_EXPIRATION_HOURS", 24)
	if err != nil {
		return nil, err
	}
	rateLimitOTPRequestsPerHour, err := getEnvAsInt("RATE_LIMIT_OTP_REQUESTS_PER_HOUR", 3)
	if err != nil {
		return nil, err
	}
	rateLimitOTPAttemptsPerOTP, err := getEnvAsInt("RATE_LIMIT_OTP_ATTEMPTS_PER_OTP", 5)
	if err != nil {
		return nil, err
	}
	rateLimitLoginAttemptsPer15Min, err := getEnvAsInt("RATE_LIMIT_LOGIN_ATTEMPTS_PER_15MIN", 5)
	if err != nil {
		return nil, err
	}
	oauthAccessTokenTTL, err := getEnvAsInt("OAUTH_ACCESS_TOKEN_TTL", 900)
	if err != nil {
		return nil, err
	}
	oauthRefreshTokenTTL, err := getEnvAsInt("OAUTH_REFRESH_TOKEN_TTL", 604800)
	if err != nil {
		return nil, err
	}
	otpExpirationMinutes, err := getEnvAsInt("OTP_EXPIRATION_MINUTES", 10)
	if err != nil {
		return nil, err
	}
	otpLength, err := getEnvAsInt("OTP_LENGTH", 6)
	if err != nil {
		return nil, err
	}
	skipEmailVerification, err := getEnvAsBool("SKIP_EMAIL_VERIFICATION", false)
	if err != nil {
		return nil, err
	}
	securityCookieSecure, err := getEnvAsBool("SECURITY_COOKIE_SECURE", false)
	if err != nil {
		return nil, err
	}
	securitySessionTTL, err := getEnvAsInt("SECURITY_SESSION_TTL", 28800)
	if err != nil {
		return nil, err
	}
	oauthAuthTransactionTTL, err := getEnvAsInt("OAUTH_AUTH_TRANSACTION_TTL", 600)
	if err != nil {
		return nil, err
	}
	rateLimitOAuthLoginFailures, err := getEnvAsInt("RATE_LIMIT_OAUTH_LOGIN_FAILURES", 5)
	if err != nil {
		return nil, err
	}
	rateLimitOAuthLoginWindowSeconds, err := getEnvAsInt("RATE_LIMIT_OAUTH_LOGIN_WINDOW_SECONDS", 900)
	if err != nil {
		return nil, err
	}

	// New security hardening config fields
	csrfEnabled, err := getEnvAsBool("CSRF_ENABLED", true)
	if err != nil {
		return nil, err
	}
	csrfTokenLength, err := getEnvAsInt("CSRF_TOKEN_LENGTH", 32)
	if err != nil {
		return nil, err
	}
	requestMaxBodySize, err := getEnvAsInt("REQUEST_MAX_BODY_SIZE", 1048576)
	if err != nil {
		return nil, err
	}
	metricsEnabled, err := getEnvAsBool("METRICS_ENABLED", true)
	if err != nil {
		return nil, err
	}
	dbMaxOpenConns, err := getEnvAsInt("DB_MAX_OPEN_CONNS", 25)
	if err != nil {
		return nil, err
	}
	dbMaxIdleConns, err := getEnvAsInt("DB_MAX_IDLE_CONNS", 10)
	if err != nil {
		return nil, err
	}
	rateLimitRegisterPerHour, err := getEnvAsInt("RATE_LIMIT_REGISTER_PER_HOUR", 3)
	if err != nil {
		return nil, err
	}
	rateLimitVerifyOTPPer10Min, err := getEnvAsInt("RATE_LIMIT_VERIFY_OTP_PER_10MIN", 5)
	if err != nil {
		return nil, err
	}
	rateLimitResendOTPPerHour, err := getEnvAsInt("RATE_LIMIT_RESEND_OTP_PER_HOUR", 3)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     dbPort,
		DBName:     getEnv("DB_NAME", "keyles"),
		DBUser:     getEnv("DB_USER", "keyles"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		// Redis
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     redisPort,
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,

		// Brevo
		BrevoAPIKey:      getEnv("BREVO_API_KEY", ""),
		BrevoSenderEmail: getEnv("BREVO_SENDER_EMAIL", "noreply@keyles.com"),
		BrevoSenderName:  getEnv("BREVO_SENDER_NAME", "Keyles SSO"),

		// JWT
		JWTSecret:          getEnv("JWT_SECRET", "dev_jwt_secret_change_in_production"),
		JWTExpirationHours: jwtExpirationHours,

		// Server
		ServerPort: getEnv("SERVER_PORT", "8080"),
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		GinMode:    getEnv("GIN_MODE", "debug"),

		// CORS
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),
		CORSAllowedMethods: getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"),
		CORSAllowedHeaders: getEnv("CORS_ALLOWED_HEADERS", "Origin,Content-Type,Accept,Authorization"),

		// Rate Limiting
		RateLimitOTPRequestsPerHour:    rateLimitOTPRequestsPerHour,
		RateLimitOTPAttemptsPerOTP:     rateLimitOTPAttemptsPerOTP,
		RateLimitLoginAttemptsPer15Min: rateLimitLoginAttemptsPer15Min,

		// OAuth
		OAuthIssuer:          getEnv("OAUTH_ISSUER", "https://sso.keyles.com"),
		OAuthAccessTokenTTL:  oauthAccessTokenTTL,  // 15 minutes
		OAuthRefreshTokenTTL: oauthRefreshTokenTTL, // 7 days

		// OTP
		OTPExpirationMinutes:     otpExpirationMinutes,
		OTPLength:                otpLength,
		SkipEmailVerification:    skipEmailVerification,

		// Browser-flow OAuth configuration
		FrontendURL:                      getEnv("FRONTEND_URL", "http://localhost:3000"),
		SecurityCookieSecure:              securityCookieSecure,
		SecuritySessionTTL:                securitySessionTTL,
		OAuthAuthTransactionTTL:            oauthAuthTransactionTTL,
		RateLimitOAuthLoginFailures:        rateLimitOAuthLoginFailures,
		RateLimitOAuthLoginWindowSeconds:    rateLimitOAuthLoginWindowSeconds,

		// Application
		AppEnv:   getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "debug"),

		// CSRF Protection
		CSRFEnabled:     csrfEnabled,
		CSRFCookieName:  getEnv("CSRF_COOKIE_NAME", "keyles_csrf"),
		CSRFHeaderName:  getEnv("CSRF_HEADER_NAME", "X-CSRF-Token"),
		CSRFTokenLength: csrfTokenLength,

		// Security Headers
		SecurityHeadersCSP:  getEnv("SECURITY_HEADERS_CSP", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"),
		SecurityHeadersHSTS: getEnv("SECURITY_HEADERS_HSTS", "max-age=31536000; includeSubDomains"),

		// Request Limits & Timeouts
		RequestMaxBodySize:  requestMaxBodySize,
		RequestReadTimeout:  getEnv("REQUEST_READ_TIMEOUT", "15s"),
		RequestWriteTimeout: getEnv("REQUEST_WRITE_TIMEOUT", "15s"),
		RequestIdleTimeout:  getEnv("REQUEST_IDLE_TIMEOUT", "60s"),

		// Metrics
		MetricsEnabled: metricsEnabled,
		MetricsPath:    getEnv("METRICS_PATH", "/metrics"),

		// Database Connection Pool
		DBMaxOpenConns:     dbMaxOpenConns,
		DBMaxIdleConns:     dbMaxIdleConns,
		DBConnMaxLifetime:  getEnv("DB_CONN_MAX_LIFETIME", "5m"),
		DBStatementTimeout: getEnv("DB_STATEMENT_TIMEOUT", "30s"),

		// Per-Endpoint Rate Limiting
		RateLimitRegisterPerHour:   rateLimitRegisterPerHour,
		RateLimitVerifyOTPPer10Min: rateLimitVerifyOTPPer10Min,
		RateLimitResendOTPPerHour:  rateLimitResendOTPPerHour,
	}

	// Validate required fields
	if cfg.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}

	if cfg.SecuritySessionTTL <= 0 {
		return nil, fmt.Errorf("SECURITY_SESSION_TTL must be a positive integer")
	}
	if cfg.OAuthAuthTransactionTTL <= 0 {
		return nil, fmt.Errorf("OAUTH_AUTH_TRANSACTION_TTL must be a positive integer")
	}
	if cfg.RateLimitOAuthLoginFailures <= 0 {
		return nil, fmt.Errorf("RATE_LIMIT_OAUTH_LOGIN_FAILURES must be a positive integer")
	}
	if cfg.RateLimitOAuthLoginWindowSeconds <= 0 {
		return nil, fmt.Errorf("RATE_LIMIT_OAUTH_LOGIN_WINDOW_SECONDS must be a positive integer")
	}

	if cfg.AppEnv == "production" {
		if cfg.DBPassword == "" {
			return nil, fmt.Errorf("DB_PASSWORD is required in production")
		}
		if cfg.BrevoAPIKey == "" {
			return nil, fmt.Errorf("BREVO_API_KEY is required in production")
		}
		if cfg.JWTSecret == "dev_jwt_secret_change_in_production" {
			return nil, fmt.Errorf("JWT_SECRET must be changed in production")
		}
		if len(cfg.JWTSecret) < 32 {
			return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters in production")
		}
		if cfg.SkipEmailVerification {
			return nil, fmt.Errorf("SKIP_EMAIL_VERIFICATION cannot be enabled in production")
		}
		if !cfg.SecurityCookieSecure {
			return nil, fmt.Errorf("SECURITY_COOKIE_SECURE must be true in production")
		}
		issuer, err := url.Parse(cfg.OAuthIssuer)
		if err != nil || issuer.Scheme != "https" || issuer.Host == "" {
			return nil, fmt.Errorf("OAUTH_ISSUER must be an HTTPS URL in production")
		}
		frontendURL, err := url.Parse(cfg.FrontendURL)
		if err != nil || frontendURL.Scheme == "" || frontendURL.Host == "" {
			return nil, fmt.Errorf("FRONTEND_URL must be a valid URL in production")
		}
		if frontendURL.Scheme != "https" {
			return nil, fmt.Errorf("FRONTEND_URL must use HTTPS in production")
		}
		if cfg.LogLevel == "debug" {
			return nil, fmt.Errorf("LOG_LEVEL must not be debug in production")
		}
		if cfg.DBSSLMode == "disable" || cfg.DBSSLMode == "allow" {
			return nil, fmt.Errorf("DB_SSL_MODE must be require, verify-ca, or verify-full in production")
		}
	}

	return cfg, nil
}

// GetDSN returns the PostgreSQL connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// GetRedisAddr returns the Redis address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.RedisHost, c.RedisPort)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) (int, error) {
	valueStr := os.Getenv(key)
	if strings.TrimSpace(valueStr) == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", key, err)
	}

	return value, nil
}

func getEnvAsBool(key string, defaultValue bool) (bool, error) {
	valueStr := strings.TrimSpace(os.Getenv(key))
	if valueStr == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.ParseBool(valueStr)
	if err != nil {
		return false, fmt.Errorf("%s must be a valid boolean (true/false): %w", key, err)
	}
	return parsed, nil
}