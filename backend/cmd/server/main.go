package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ranggaaprilio/keyles/infrastructure/config"
	postgresRepo "github.com/ranggaaprilio/keyles/infrastructure/persistence/postgres"
	redisRepo "github.com/ranggaaprilio/keyles/infrastructure/persistence/redis"
	infraServices "github.com/ranggaaprilio/keyles/infrastructure/services"
	httpServer "github.com/ranggaaprilio/keyles/interfaces/http"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/interfaces/http/middleware"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/ranggaaprilio/keyles/usecase/client"
	"github.com/ranggaaprilio/keyles/usecase/tenant"
)

func main() {
	// Load .env file (ignore error in production where env vars are set directly)
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	log.Printf("Starting Keyles SSO service in %s mode...", cfg.GinMode)

	// Connect to PostgreSQL
	db, err := initPostgreSQL(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// Connect to Redis
	redisClient, err := initRedis(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("Connected to Redis")

	// Initialize repositories
	tenantRepo := postgresRepo.NewPostgresTenantRepository(db)
	userRepo := postgresRepo.NewPostgresUserRepository(db)
	auditRepo := postgresRepo.NewPostgresAuditRepository(db)
	otpRepo := redisRepo.NewRedisOTPRepository(redisClient)
	clientRepo := postgresRepo.NewPostgresClientRepository(db)
	roleRepo := postgresRepo.NewPostgresRoleRepository(db)
	authCodeRepo := redisRepo.NewRedisAuthCodeRepository(redisClient)
	refreshTokenRepo := postgresRepo.NewPostgresRefreshTokenRepository(db)
	signingKeyRepo := postgresRepo.NewPostgresSigningKeyRepositoryGorm(db)

	// Initialize services
	emailService := infraServices.NewBrevoEmailService(cfg.BrevoAPIKey, cfg.BrevoSenderEmail, cfg.BrevoSenderName)
	otpService := infraServices.NewCryptoOTPService()
	passwordService := infraServices.NewBcryptPasswordService() // Returns concrete type now
	jwtService := infraServices.NewJWTService(cfg.JWTSecret, cfg.JWTExpirationHours)
	authJWTService := infraServices.NewAuthJWTServiceAdapter(jwtService)

	// Initialize OAuth token service
	tokenService := infraServices.NewRSATokenService(signingKeyRepo)

	// Initialize use cases
	registerTenantUseCase := tenant.NewRegisterTenantUseCase(
		tenantRepo, userRepo, otpRepo, auditRepo,
		emailService, otpService, passwordService,
	)
	checkAvailabilityUseCase := tenant.NewCheckAvailabilityUseCase(tenantRepo, userRepo)
	verifyTenantUseCase := tenant.NewVerifyTenantUseCase(otpRepo, tenantRepo, auditRepo)
	resendOTPUseCase := tenant.NewResendOTPUseCase(otpRepo, tenantRepo, userRepo, emailService, otpService, auditRepo)
	authenticateAdminUseCase := auth.NewAuthenticateAdminUseCase(userRepo, tenantRepo, passwordService, authJWTService)

	// Initialize client management use cases
	createClientUseCase := client.NewCreateClientUseCase(clientRepo, passwordService)
	getClientUseCase := client.NewGetClientUseCase(clientRepo)
	updateClientUseCase := client.NewUpdateClientUseCase(clientRepo)
	deleteClientUseCase := client.NewDeleteClientUseCase(clientRepo)
	listClientsUseCase := client.NewListClientsUseCase(clientRepo)
	rotateSecretUseCase := client.NewRotateSecretUseCase(clientRepo, passwordService)

	// Initialize OAuth use cases
	authorizeClientUseCase := auth.NewAuthorizeClient(clientRepo, roleRepo, authCodeRepo)
	issueTokenUseCase := auth.NewIssueToken(authCodeRepo, clientRepo, refreshTokenRepo, tokenService, cfg.OAuthIssuer)

	// Initialize handlers
	registrationHandler := handlers.NewRegistrationHandler(registerTenantUseCase)
	availabilityHandler := handlers.NewAvailabilityHandler(checkAvailabilityUseCase)
	verificationHandler := handlers.NewVerificationHandler(verifyTenantUseCase)
	resendOTPHandler := handlers.NewResendOTPHandler(resendOTPUseCase)
	authHandler := handlers.NewAuthHandler(authenticateAdminUseCase)
	dashboardHandler := handlers.NewDashboardHandler(userRepo, tenantRepo)
	healthHandler := handlers.NewHealthHandler(tenantRepo, redisClient)
	clientHandler := handlers.NewClientHandler(
		createClientUseCase,
		getClientUseCase,
		updateClientUseCase,
		deleteClientUseCase,
		listClientsUseCase,
		rotateSecretUseCase,
	)
	oauthHandler := handlers.NewOAuthHandler(authorizeClientUseCase, issueTokenUseCase, clientRepo)
	discoveryHandler := handlers.NewDiscoveryHandler(tokenService, cfg.OAuthIssuer)

	// Initialize middleware
	rateLimiter := middleware.NewRateLimiter(redisClient)

	// Initialize router
	router := httpServer.NewRouter(
		rateLimiter,
		registrationHandler,
		availabilityHandler,
		verificationHandler,
		resendOTPHandler,
		authHandler,
		dashboardHandler,
		healthHandler,
		clientHandler,
		oauthHandler,
		discoveryHandler,
		jwtService,
		cfg.CORSAllowedOrigins,
		cfg.CORSAllowedMethods,
		cfg.CORSAllowedHeaders,
	)
	router.Setup()

	// Log dependency injection status
	log.Printf("Dependency injection complete:")
	log.Printf("  - TenantRepository: %T", tenantRepo)
	log.Printf("  - UserRepository: %T", userRepo)
	log.Printf("  - AuditRepository: %T", auditRepo)
	log.Printf("  - OTPRepository: %T", otpRepo)
	log.Printf("  - EmailService: %T", emailService)
	log.Printf("  - OTPService: %T", otpService)
	log.Printf("  - PasswordService: %T", passwordService)

	// Start HTTP server
	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Starting HTTP server on %s", serverAddr)

	// Graceful shutdown
	go func() {
		if err := router.Run(serverAddr); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Cleanup
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	log.Println("Server stopped")
}

// initPostgreSQL connects to PostgreSQL and configures GORM
func initPostgreSQL(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	// Configure logger based on environment
	logLevel := logger.Silent
	if cfg.GinMode == "debug" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// initRedis connects to Redis
func initRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.GetRedisAddr(),
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return client, nil
}
