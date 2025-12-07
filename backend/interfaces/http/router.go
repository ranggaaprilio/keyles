package http

import (
	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/interfaces/http/middleware"
)

// Router sets up all HTTP routes
type Router struct {
	engine               *gin.Engine
	rateLimiter          *middleware.RateLimiter
	registrationHandler  *handlers.RegistrationHandler
	availabilityHandler  *handlers.AvailabilityHandler
}

// NewRouter creates a new HTTP router
func NewRouter(
	rateLimiter *middleware.RateLimiter,
	registrationHandler *handlers.RegistrationHandler,
	availabilityHandler *handlers.AvailabilityHandler,
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
	}
}

// Setup configures all routes
func (r *Router) Setup() {
	// API v1 routes
	v1 := r.engine.Group("/api/v1")
	{
		// Health check (public)
		v1.GET("/health", r.healthCheck)

		// Registration routes (public)
		registration := v1.Group("/")
		{
			registration.POST("/register", r.registrationHandler.Register)
			registration.GET("/check-availability", r.availabilityHandler.CheckAvailability)
			// TODO: Phase 4 - OTP verification
			// registration.POST("/verify-otp", handlers.VerifyOTP)
			// registration.POST("/resend-otp", handlers.ResendOTP)
		}

		// Auth routes (public)
		// TODO: Wire up auth handlers in Phase 5
		// auth := v1.Group("/")
		// {
		// 	auth.POST("/login", handlers.Login)
		// }

		// Protected routes (require JWT)
		// TODO: Wire up protected routes in Phase 5
		// protected := v1.Group("/")
		// protected.Use(middleware.Auth())
		// {
		// 	protected.GET("/dashboard", handlers.Dashboard)
		// }
	}
}

// healthCheck is a simple health check endpoint
func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "healthy",
		"service": "keyles-sso",
		"version": "1.0.0",
	})
}

// GetEngine returns the Gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// Run starts the HTTP server
func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
