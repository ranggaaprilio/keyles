/**
 * Health Check Handler
 * Provides endpoints for monitoring application health
 */

package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/redis/go-redis/v9"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	tenantRepo repositories.TenantRepository
	redisClient *redis.Client
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(tenantRepo repositories.TenantRepository, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{
		tenantRepo: tenantRepo,
		redisClient: redisClient,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string            `json:"status"`
	Service string            `json:"service"`
	Version string            `json:"version"`
	Checks  map[string]string `json:"checks,omitempty"`
}

// Health returns basic health status
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "healthy",
		Service: "keyles-api",
		Version: "1.0.0",
	})
}

// HealthDB checks database connectivity
func (h *HealthHandler) HealthDB(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	checks := make(map[string]string)
	overallStatus := "healthy"

	// Test database connection by attempting a simple query
	_, err := h.tenantRepo.FindByID(ctx, uuid.Nil)
	if err != nil {
		// Expected to not find this ID, but connection should work
		if err.Error() == "tenant not found" {
			checks["database"] = "healthy"
		} else {
			checks["database"] = "unhealthy: " + err.Error()
			overallStatus = "unhealthy"
		}
	} else {
		checks["database"] = "healthy"
	}

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, HealthResponse{
		Status:  overallStatus,
		Service: "keyles-api",
		Version: "1.0.0",
		Checks:  checks,
	})
}

// HealthRedis checks Redis connectivity
func (h *HealthHandler) HealthRedis(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	checks := make(map[string]string)
	overallStatus := "healthy"

	// Test Redis connection with PING
	if h.redisClient != nil {
		err := h.redisClient.Ping(ctx).Err()
		if err != nil {
			checks["redis"] = "unhealthy: " + err.Error()
			overallStatus = "unhealthy"
		} else {
			checks["redis"] = "healthy"
		}
	} else {
		checks["redis"] = "not configured"
	}

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, HealthResponse{
		Status:  overallStatus,
		Service: "keyles-api",
		Version: "1.0.0",
		Checks:  checks,
	})
}
