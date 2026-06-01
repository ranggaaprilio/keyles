package http

import (
	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/domain/services"
	redisRepo "github.com/ranggaaprilio/keyles/infrastructure/persistence/redis"
	infraServices "github.com/ranggaaprilio/keyles/infrastructure/services"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/interfaces/http/middleware"
	"github.com/ranggaaprilio/keyles/usecase/auth"
)

// Router sets up all HTTP routes
type Router struct {
	engine              *gin.Engine
	rateLimiter         *redisRepo.RedisRateLimiter
	registrationHandler *handlers.RegistrationHandler
	availabilityHandler *handlers.AvailabilityHandler
	verificationHandler *handlers.VerificationHandler
	resendOTPHandler    *handlers.ResendOTPHandler
	authHandler         *handlers.AuthHandler
	dashboardHandler    *handlers.DashboardHandler
	healthHandler       *handlers.HealthHandler
	clientHandler       *handlers.ClientHandler
	oauthHandler        *handlers.OAuthHandler
	discoveryHandler    *handlers.DiscoveryHandler
	roleHandler         *handlers.RoleHandler
	sessionHandler      *handlers.SessionHandler
	userHandler         *handlers.UserHandler
	invitationHandler   *handlers.InvitationHandler
	activityHandler     *handlers.ActivityHandler
	userinfoHandler     *handlers.UserinfoHandler
	jwtService          *infraServices.JWTService
	userBlacklist       services.UserBlacklist
	oauthTokenValidator *auth.AccessTokenValidator
}

// NewRouter creates a new HTTP router
func NewRouter(
	rateLimiter *redisRepo.RedisRateLimiter,
	registrationHandler *handlers.RegistrationHandler,
	availabilityHandler *handlers.AvailabilityHandler,
	verificationHandler *handlers.VerificationHandler,
	resendOTPHandler *handlers.ResendOTPHandler,
	authHandler *handlers.AuthHandler,
	dashboardHandler *handlers.DashboardHandler,
	healthHandler *handlers.HealthHandler,
	clientHandler *handlers.ClientHandler,
	oauthHandler *handlers.OAuthHandler,
	discoveryHandler *handlers.DiscoveryHandler,
	roleHandler *handlers.RoleHandler,
	sessionHandler *handlers.SessionHandler,
	userHandler *handlers.UserHandler,
	invitationHandler *handlers.InvitationHandler,
	activityHandler *handlers.ActivityHandler,
	userinfoHandler *handlers.UserinfoHandler,
	jwtService *infraServices.JWTService,
	userBlacklist services.UserBlacklist,
	oauthTokenValidator *auth.AccessTokenValidator,
	corsOrigins, corsMethods, corsHeaders string,
) *Router {
	engine := gin.New()

	// Global middleware
	engine.Use(gin.Logger())
	engine.Use(middleware.RecoveryHandler())
	engine.Use(middleware.CORS(corsOrigins, corsMethods, corsHeaders))
	engine.Use(middleware.ErrorHandler())

	return &Router{
		engine:              engine,
		rateLimiter:         rateLimiter,
		registrationHandler: registrationHandler,
		availabilityHandler: availabilityHandler,
		verificationHandler: verificationHandler,
		resendOTPHandler:    resendOTPHandler,
		authHandler:         authHandler,
		dashboardHandler:    dashboardHandler,
		healthHandler:       healthHandler,
		clientHandler:       clientHandler,
		oauthHandler:        oauthHandler,
		discoveryHandler:    discoveryHandler,
		roleHandler:         roleHandler,
		sessionHandler:      sessionHandler,
		userHandler:         userHandler,
		invitationHandler:   invitationHandler,
		activityHandler:     activityHandler,
		userinfoHandler:     userinfoHandler,
		jwtService:          jwtService,
		userBlacklist:       userBlacklist,
		oauthTokenValidator: oauthTokenValidator,
	}
}

// Setup configures all routes
func (r *Router) Setup() {
	blacklistCheck := func(c *gin.Context) { c.Next() }
	if r.userBlacklist != nil {
		blacklistCheck = middleware.BlacklistCheckMiddleware(r.userBlacklist)
	}

	// Root health check
	r.engine.GET("/health", r.healthHandler.Health)
	r.engine.GET("/health/db", r.healthHandler.HealthDB)
	r.engine.GET("/health/redis", r.healthHandler.HealthRedis)

	// OIDC Discovery endpoints (public - no auth required)
	r.engine.GET("/.well-known/openid-configuration", r.discoveryHandler.OpenIDConfiguration)
	r.engine.GET("/.well-known/jwks.json", r.discoveryHandler.JWKS)

	// OAuth 2.0 routes (public - handles authentication)
	oauth2 := r.engine.Group("/oauth2")
	{
		oauth2.GET("/auth", r.oauthHandler.AuthorizeBrowser)
		oauth2.POST("/auth", r.oauthHandler.Authorize)
		// Browser-facing OAuth flow (Phase 3)
		oauth2.POST("/login", r.oauthHandler.Login)
		oauth2.GET("/consent/:transactionId", r.oauthHandler.ConsentDetail)
		oauth2.POST("/consent", r.oauthHandler.ConsentDecision)
		oauth2.POST("/logout", r.oauthHandler.Logout)
		// Token endpoint with rate limiting (FR-057: 10 req/min per client_id)
		if r.rateLimiter != nil {
			oauth2.POST("/token", middleware.RateLimiterMiddleware(r.rateLimiter), r.oauthHandler.Token)
		} else {
			oauth2.POST("/token", r.oauthHandler.Token)
		}
		// Revocation endpoint per RFC 7009 (FR-051)
		oauth2.POST("/revoke", r.oauthHandler.Revoke)
		oauth2.POST("/introspect", r.oauthHandler.Introspect)
		// UserInfo endpoint per OIDC spec (FR-052) - requires valid access token
		oauth2.GET("/userinfo", middleware.OAuthAccessTokenMiddleware(r.oauthTokenValidator), r.userinfoHandler.UserInfo)
	}

	// API v1 routes
	v1 := r.engine.Group("/api/v1")
	{
		v1.GET("/health", r.healthHandler.Health)

		// Registration routes (public)
		registration := v1.Group("/")
		{
			registration.POST("/register", r.registrationHandler.Register)
			registration.GET("/check-availability", r.availabilityHandler.CheckAvailability)

			// OTP verification routes (Phase 4)
			registration.POST("/verify-otp", r.verificationHandler.VerifyOTP)
			registration.POST("/resend-otp", r.resendOTPHandler.ResendOTP)
		}

		// Auth routes (public) - Phase 5
		v1.POST("/login", r.authHandler.Login)
		if r.invitationHandler != nil {
			v1.GET("/invitations/:token/validate", r.invitationHandler.ValidateInvitation)
			v1.POST("/invitations/:token/accept", r.invitationHandler.AcceptInvitation)
		}

		// Protected routes (require JWT) - Phase 5
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(r.jwtService), blacklistCheck)
		{
			protected.GET("/dashboard", r.dashboardHandler.GetDashboard)
		}

		// Admin routes (require JWT) - Client Management
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(r.jwtService), blacklistCheck)
		{
			// Client management routes
			clients := admin.Group("/clients")
			{
				clients.POST("", r.clientHandler.Create)
				clients.GET("", r.clientHandler.List)
				clients.GET("/:clientId", r.clientHandler.Get)
				clients.PUT("/:clientId", r.clientHandler.Update)
				clients.DELETE("/:clientId", r.clientHandler.Delete)
				clients.POST("/:clientId/rotate-secret", r.clientHandler.RotateSecret)
			}

			// Role management routes (FR-006a, FR-006b)
			roles := admin.Group("/roles")
			{
				roles.POST("/assign", r.roleHandler.AssignRole)
				roles.POST("/revoke", r.roleHandler.RevokeRole)
				roles.GET("/users/:userId", r.roleHandler.ListUserRoles)
				roles.GET("/clients/:clientId", r.roleHandler.ListClientRoles)
			}

			// User management routes
			if r.userHandler != nil {
				users := admin.Group("/users")
				{
					users.GET("", r.userHandler.ListUsers)
					users.POST("/invite", r.userHandler.InviteUser)
					users.GET("/:id", r.userHandler.GetUser)
					users.PATCH("/:id", r.userHandler.UpdateUser)
					users.PATCH("/:id/status", r.userHandler.UpdateUserStatus)
					users.DELETE("/:id", r.userHandler.DeleteUser)
					users.POST("/:id/resend-invitation", r.userHandler.ResendInvitation)

					if r.roleHandler != nil {
						users.GET("/:id/roles", r.roleHandler.ListUserRolesForUser)
						users.POST("/:id/roles", r.roleHandler.AssignUserRole)
						users.DELETE("/:id/roles/:assignmentId", r.roleHandler.RevokeUserRole)
					}

					if r.sessionHandler != nil {
						users.GET("/:id/sessions", r.sessionHandler.ListUserSessions)
						users.DELETE("/:id/sessions/:sessionId", r.sessionHandler.RevokeUserSession)
					}

					if r.activityHandler != nil {
						users.GET("/:id/activity", r.activityHandler.ListUserActivity)
					}
				}
			}
		}
	}
}

// GetEngine returns the Gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// Run starts the HTTP server
func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
