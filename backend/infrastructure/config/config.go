package config

import (
	"fmt"
	"os"
	"strconv"
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
	BrevoAPIKey     string
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
	RateLimitOTPRequestsPerHour   int
	RateLimitOTPAttemptsPerOTP    int
	RateLimitLoginAttemptsPer15Min int

	// OAuth
	OAuthIssuer          string
	OAuthAccessTokenTTL  int
	OAuthRefreshTokenTTL int

	// OTP
	OTPExpirationMinutes int
	OTPLength            int

	// Application
	AppEnv   string
	LogLevel string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvAsInt("DB_PORT", 5432),
		DBName:     getEnv("DB_NAME", "keyles"),
		DBUser:     getEnv("DB_USER", "keyles"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		// Redis
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnvAsInt("REDIS_PORT", 6379),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		// Brevo
		BrevoAPIKey:      getEnv("BREVO_API_KEY", ""),
		BrevoSenderEmail: getEnv("BREVO_SENDER_EMAIL", "noreply@keyles.com"),
		BrevoSenderName:  getEnv("BREVO_SENDER_NAME", "Keyles SSO"),

		// JWT
		JWTSecret:          getEnv("JWT_SECRET", "dev_jwt_secret_change_in_production"),
		JWTExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),

		// Server
		ServerPort: getEnv("SERVER_PORT", "8080"),
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		GinMode:    getEnv("GIN_MODE", "debug"),

		// CORS
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),
		CORSAllowedMethods: getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"),
		CORSAllowedHeaders: getEnv("CORS_ALLOWED_HEADERS", "Origin,Content-Type,Accept,Authorization"),

		// Rate Limiting
		RateLimitOTPRequestsPerHour:   getEnvAsInt("RATE_LIMIT_OTP_REQUESTS_PER_HOUR", 3),
		RateLimitOTPAttemptsPerOTP:    getEnvAsInt("RATE_LIMIT_OTP_ATTEMPTS_PER_OTP", 5),
		RateLimitLoginAttemptsPer15Min: getEnvAsInt("RATE_LIMIT_LOGIN_ATTEMPTS_PER_15MIN", 5),

		// OAuth
		OAuthIssuer:          getEnv("OAUTH_ISSUER", "https://sso.keyles.com"),
		OAuthAccessTokenTTL:  getEnvAsInt("OAUTH_ACCESS_TOKEN_TTL", 900),   // 15 minutes
		OAuthRefreshTokenTTL: getEnvAsInt("OAUTH_REFRESH_TOKEN_TTL", 604800), // 7 days

		// OTP
		OTPExpirationMinutes: getEnvAsInt("OTP_EXPIRATION_MINUTES", 10),
		OTPLength:            getEnvAsInt("OTP_LENGTH", 6),

		// Application
		AppEnv:   getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "debug"),
	}

	// Validate required fields
	if cfg.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}

	if cfg.BrevoAPIKey == "" && cfg.AppEnv == "production" {
		return nil, fmt.Errorf("BREVO_API_KEY is required in production")
	}

	if cfg.JWTSecret == "dev_jwt_secret_change_in_production" && cfg.AppEnv == "production" {
		return nil, fmt.Errorf("JWT_SECRET must be changed in production")
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

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
