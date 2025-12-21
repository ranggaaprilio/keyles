package http

import (
	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/infrastructure/services"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/interfaces/http/middleware"
)

// Router sets up all HTTP routes
type Router struct {
	engine               *gin.Engine
	rateLimiter          *middleware.RateLimiter
	registrationHandler  *handlers.RegistrationHandler
	availabilityHandler  *handlers.AvailabilityHandler
	verificationHandler  *handlers.VerificationHandler
	resendOTPHandler     *handlers.ResendOTPHandler
	authHandler          *handlers.AuthHandler
	dashboardHandler     *handlers.DashboardHandler
	healthHandler        *handlers.HealthHandler
	jwtService           *services.JWTService
}

// NewRouter creates a new HTTP router
func NewRouter(
	rateLimiter *middleware.RateLimiter,
	registrationHandler *handlers.RegistrationHandler,
	availabilityHandler *handlers.AvailabilityHandler,
	verificationHandler *handlers.VerificationHandler,
	resendOTPHandler *handlers.ResendOTPHandler,
	authHandler *handlers.AuthHandler,
	dashboardHandler *handlers.DashboardHandler,
	healthHandler *handlers.HealthHandler,
	jwtService *services.JWTService,
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
		jwtService:          jwtService,
	}
}
// Setup configures all routes
func (r *Router) Setup() {
	// Root health check
	r.engine.GET("/health", r.healthHandler.Health)
	r.engine.GET("/health/db", r.healthHandler.HealthDB)
	r.engine.GET("/health/redis", r.healthHandler.HealthRedis)

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

		// Protected routes (require JWT) - Phase 5
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(r.jwtService))
		{
			protected.GET("/dashboard", r.dashboardHandler.GetDashboard)
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
