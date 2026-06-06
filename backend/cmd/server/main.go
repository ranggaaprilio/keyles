package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/infrastructure/config"
	"github.com/ranggaaprilio/keyles/infrastructure/logging"
	postgresRepo "github.com/ranggaaprilio/keyles/infrastructure/persistence/postgres"
	redisRepo "github.com/ranggaaprilio/keyles/infrastructure/persistence/redis"
	infraServices "github.com/ranggaaprilio/keyles/infrastructure/services"
	httpServer "github.com/ranggaaprilio/keyles/interfaces/http"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/interfaces/http/middleware"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/ranggaaprilio/keyles/usecase/client"
	"github.com/ranggaaprilio/keyles/usecase/role"
	"github.com/ranggaaprilio/keyles/usecase/tenant"
	"github.com/ranggaaprilio/keyles/usecase/user"
)

func main() {
	// Load .env file (ignore error in production where env vars are set directly)
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize structured logger
	logger := logging.NewLogger(cfg.LogLevel)
	middleware.Logger = logger.With("component", "middleware")

	// Validate security configuration for production
	securityConfig := &entities.SecurityConfig{
		AppEnv:               cfg.AppEnv,
		JWTSecret:            cfg.JWTSecret,
		DBPassword:           cfg.DBPassword,
		DBSSLMode:            cfg.DBSSLMode,
		BrevoAPIKey:          cfg.BrevoAPIKey,
		SecurityCookieSecure: cfg.SecurityCookieSecure,
		OAuthIssuer:          cfg.OAuthIssuer,
		FrontendURL:          cfg.FrontendURL,
		LogLevel:             cfg.LogLevel,
	}
	if err := securityConfig.ValidateForProduction(); err != nil {
		logger.Error("Security validation failed", "error", err)
		os.Exit(1)
	}

	// Initialize Prometheus metrics
	if cfg.MetricsEnabled {
		logger.Info("Metrics enabled", "path", cfg.MetricsPath)
	}

	// Initialize logger
	logger.Info("Starting Keyles SSO service", "mode", cfg.GinMode)

	// Connect to PostgreSQL
	db, err := initPostgreSQL(cfg)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	logger.Info("Connected to PostgreSQL")

	// Connect to Redis
	redisClient, err := initRedis(cfg)
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	logger.Info("Connected to Redis")

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
	endUserRepo := postgresRepo.NewPostgresEndUserRepository(db)
	invitationRepo := postgresRepo.NewPostgresInvitationRepository(db)
	userEventRepo := postgresRepo.NewPostgresUserEventRepository(db)
	txnRepo := redisRepo.NewRedisAuthorizationTransactionRepository(redisClient)
	sessionRepo := redisRepo.NewRedisSessionRepository(redisClient)
	loginThrottler := redisRepo.NewRedisLoginThrottler(redisClient, cfg.RateLimitOAuthLoginFailures, cfg.RateLimitOAuthLoginWindowSeconds)

	// Initialize Redis caches
	clientCountCache := redisRepo.NewClientCountCache(redisClient)
	revokedClientCache := redisRepo.NewRevokedClientCache(redisClient)
	userBlacklist := redisRepo.NewUserBlacklist(redisClient)
	userCountCache := redisRepo.NewUserCountCache(redisClient)

	// Initialize services
	emailService := infraServices.NewBrevoEmailService(cfg.BrevoAPIKey, cfg.BrevoSenderEmail, cfg.BrevoSenderName)
	otpService := infraServices.NewCryptoOTPService()
	passwordService := infraServices.NewBcryptPasswordService() // Returns concrete type now
	jwtService := infraServices.NewJWTService(cfg.JWTSecret, cfg.JWTExpirationHours)
	authJWTService := infraServices.NewAuthJWTServiceAdapter(jwtService)

	// Initialize OAuth token service
	tokenService := infraServices.NewRSATokenService(signingKeyRepo)
	// Initialize OAuth audit helper
	oauthAuditHelper := auth.NewOAuthAuditHelper(auditRepo)

	registerTenantUseCase := tenant.NewRegisterTenantUseCase(
		tenantRepo, userRepo, otpRepo, auditRepo,
		emailService, otpService, passwordService,
		cfg.SkipEmailVerification,
	)
	checkAvailabilityUseCase := tenant.NewCheckAvailabilityUseCase(tenantRepo, userRepo)
	verifyTenantUseCase := tenant.NewVerifyTenantUseCase(otpRepo, tenantRepo, auditRepo)
	resendOTPUseCase := tenant.NewResendOTPUseCase(otpRepo, tenantRepo, userRepo, emailService, otpService, auditRepo)
	authenticateAdminUseCase := auth.NewAuthenticateAdminUseCase(userRepo, tenantRepo, passwordService, authJWTService)

	// Initialize client management use cases
	createClientUseCase := client.NewCreateClientUseCase(clientRepo, passwordService, auditRepo, clientCountCache)
	getClientUseCase := client.NewGetClientUseCase(clientRepo)
	updateClientUseCase := client.NewUpdateClientUseCase(clientRepo, auditRepo)
	deleteClientUseCase := client.NewDeleteClientUseCase(clientRepo, auditRepo, refreshTokenRepo, revokedClientCache, clientCountCache)
	listClientsUseCase := client.NewListClientsUseCase(clientRepo)
	rotateSecretUseCase := client.NewRotateSecretUseCase(clientRepo, passwordService, auditRepo)

	// Initialize role management use cases
	assignRoleUseCase := role.NewAssignRole(roleRepo, userRepo, clientRepo, userEventRepo, auditRepo)
	revokeRoleUseCase := role.NewRevokeRole(roleRepo, refreshTokenRepo, userEventRepo, auditRepo)
	listUserRolesUseCase := role.NewListUserRoles(roleRepo)

	// Initialize OAuth use cases
	authorizeClientUseCase := auth.NewAuthorizeClient(clientRepo, roleRepo, authCodeRepo, endUserRepo)
	issueTokenUseCase := auth.NewIssueToken(authCodeRepo, clientRepo, refreshTokenRepo, roleRepo, tokenService, cfg.OAuthIssuer, endUserRepo)
	refreshTokenUseCase := auth.NewRefreshToken(refreshTokenRepo, clientRepo, tokenService, cfg.OAuthIssuer, roleRepo, endUserRepo)
	getUserInfoUseCase := auth.NewGetUserInfo(endUserRepo, roleRepo)
	revokeTokenUseCase := auth.NewRevokeToken(refreshTokenRepo)
	oauthTokenValidator := auth.NewAccessTokenValidator(tokenService, clientRepo, endUserRepo, revokedClientCache, userBlacklist, cfg.OAuthIssuer)
	introspectTokenUseCase := auth.NewIntrospectToken(clientRepo, oauthTokenValidator)
	transactionTTL := time.Duration(cfg.OAuthAuthTransactionTTL) * time.Second
	oauthInteraction := auth.NewOAuthInteraction(clientRepo, roleRepo, txnRepo, sessionRepo, endUserRepo, oauthAuditHelper, cfg.FrontendURL, transactionTTL)
	sessionTTL := time.Duration(cfg.SecuritySessionTTL) * time.Second
	authenticateEndUser := auth.NewAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, loginThrottler, passwordService, oauthAuditHelper, sessionTTL, cfg.FrontendURL)
	getConsentDetails := auth.NewGetConsentDetails(txnRepo, sessionRepo, clientRepo, endUserRepo)
	consentDecision := auth.NewConsentDecision(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, authCodeRepo, oauthAuditHelper)
	logoutEndUser := auth.NewLogoutEndUser(sessionRepo, oauthAuditHelper)

	// Initialize user management use cases
	inviteUserUseCase := user.NewInviteUser(endUserRepo, invitationRepo, userEventRepo, emailService, userCountCache)
	acceptInvitationUseCase := user.NewAcceptInvitation(endUserRepo, invitationRepo, userEventRepo, passwordService)
	resendUserInvitationUseCase := user.NewResendInvitation(endUserRepo, invitationRepo, userEventRepo, emailService)
	listUsersUseCase := user.NewListUsers(endUserRepo)
	getUserUseCase := user.NewGetUser(endUserRepo, roleRepo)
	updateUserUseCase := user.NewUpdateUser(endUserRepo, auditRepo)
	disableUserUseCase := user.NewDisableUser(endUserRepo, userRepo, refreshTokenRepo, userBlacklist, userEventRepo, auditRepo)
	enableUserUseCase := user.NewEnableUser(endUserRepo, userEventRepo, auditRepo)
	deleteUserUseCase := user.NewDeleteUser(endUserRepo, roleRepo, refreshTokenRepo, userBlacklist, auditRepo)
	listSessionsUseCase := user.NewListSessions(endUserRepo, refreshTokenRepo)
	revokeSessionUseCase := user.NewRevokeSession(endUserRepo, refreshTokenRepo, userEventRepo)
	listUserActivityUseCase := user.NewListUserActivity(endUserRepo, userEventRepo)

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
	oauthHandler := handlers.NewOAuthHandlerFullBrowser(authorizeClientUseCase, issueTokenUseCase, clientRepo, refreshTokenUseCase, revokeTokenUseCase, introspectTokenUseCase, oauthInteraction, authenticateEndUser, getConsentDetails, consentDecision, logoutEndUser, sessionRepo, loginThrottler, oauthAuditHelper, cfg)
	discoveryHandler := handlers.NewDiscoveryHandler(tokenService, cfg.OAuthIssuer)
	roleHandler := handlers.NewRoleHandler(assignRoleUseCase, revokeRoleUseCase, listUserRolesUseCase)
	userHandler := handlers.NewUserHandler(
		inviteUserUseCase,
		getUserUseCase,
		listUsersUseCase,
		updateUserUseCase,
		disableUserUseCase,
		enableUserUseCase,
		deleteUserUseCase,
		resendUserInvitationUseCase,
	)
	invitationHandler := handlers.NewInvitationHandler(acceptInvitationUseCase, invitationRepo)
	activityHandler := handlers.NewActivityHandler(listUserActivityUseCase)
	sessionHandler := handlers.NewSessionHandler(revokeTokenUseCase, listSessionsUseCase, revokeSessionUseCase)
	userinfoHandler := handlers.NewUserinfoHandler(getUserInfoUseCase)

	// Initialize middleware
	rateLimiter, err := redisRepo.NewRedisRateLimiter(redisClient, limiter.Rate{Period: time.Minute, Limit: 10})
	if err != nil {
		logger.Error("Failed to initialize OAuth rate limiter", "error", err)
		os.Exit(1)
	}

	// Initialize sliding window rate limiter for public endpoints
	slidingRateLimiter := middleware.NewRateLimiter(redisClient)

	// Initialize router
	router := httpServer.NewRouter(
		cfg,
		rateLimiter,
		slidingRateLimiter,
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
		roleHandler,
		sessionHandler,
		userHandler,
		invitationHandler,
		activityHandler,
		userinfoHandler,
		jwtService,
		userBlacklist,
		oauthTokenValidator,
		cfg.CORSAllowedOrigins,
		cfg.CORSAllowedMethods,
		cfg.CORSAllowedHeaders,
	)
	router.Setup()

	// Log dependency injection status
	logger.Info("Dependency injection complete",
		"tenantRepository", fmt.Sprintf("%T", tenantRepo),
		"userRepository", fmt.Sprintf("%T", userRepo),
		"auditRepository", fmt.Sprintf("%T", auditRepo),
		"otpRepository", fmt.Sprintf("%T", otpRepo),
		"emailService", fmt.Sprintf("%T", emailService),
		"otpService", fmt.Sprintf("%T", otpService),
		"passwordService", fmt.Sprintf("%T", passwordService),
		"skipEmailVerification", cfg.SkipEmailVerification,
	)

	// Start HTTP server with timeouts
	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	readTimeout, _ := time.ParseDuration(cfg.RequestReadTimeout)
	writeTimeout, _ := time.ParseDuration(cfg.RequestWriteTimeout)
	idleTimeout, _ := time.ParseDuration(cfg.RequestIdleTimeout)
	if readTimeout == 0 {
		readTimeout = 15 * time.Second
	}
	if writeTimeout == 0 {
		writeTimeout = 15 * time.Second
	}
	if idleTimeout == 0 {
		idleTimeout = 60 * time.Second
	}

	server := &http.Server{
		Addr:         serverAddr,
		Handler:      router.GetEngine(),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	logger.Info("Starting HTTP server", "address", serverAddr)

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start HTTP server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	// Cleanup
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	logger.Info("Server stopped")
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

	// Configure connection pool from config
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	connMaxLifetime, _ := time.ParseDuration(cfg.DBConnMaxLifetime)
	if connMaxLifetime == 0 {
		connMaxLifetime = 5 * time.Minute
	}

	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	// Set statement timeout on each new connection
	if cfg.DBStatementTimeout != "" {
		if err := db.Exec(fmt.Sprintf("SET statement_timeout = '%s'", cfg.DBStatementTimeout)).Error; err != nil {
			return nil, fmt.Errorf("failed to set statement timeout: %w", err)
		}
	}

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
