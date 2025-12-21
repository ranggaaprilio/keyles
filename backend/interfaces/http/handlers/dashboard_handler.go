/**
 * Dashboard HTTP handler
 */

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// DashboardHandler handles dashboard requests
type DashboardHandler struct {
	userRepo   repositories.UserRepository
	tenantRepo repositories.TenantRepository
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(userRepo repositories.UserRepository, tenantRepo repositories.TenantRepository) *DashboardHandler {
	return &DashboardHandler{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
	}
}

// DashboardResponse represents the dashboard response
type DashboardResponse struct {
	Tenant DashboardTenant `json:"tenant"`
	User   DashboardUser   `json:"user"`
}

// DashboardTenant represents tenant info in dashboard
type DashboardTenant struct {
	ID               string  `json:"id"`
	OrganizationName string  `json:"organization_name"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"created_at"`
	VerifiedAt       *string `json:"verified_at"`
}

// DashboardUser represents user info in dashboard
type DashboardUser struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// GetDashboard handles GET /api/v1/dashboard
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	// Get user info from JWT claims (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tenant not found",
		})
		return
	}

	// Load user details
	user, err := h.userRepo.FindByEmail(c.Request.Context(), c.GetString("email"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// Load tenant details
	tenantUUID, err := uuid.Parse(tenantID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}
	tenant, err := h.tenantRepo.FindByID(c.Request.Context(), tenantUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tenant not found",
		})
		return
	}

	// Format verified_at timestamp
	var verifiedAt *string
	if tenant.VerifiedAt != nil {
		formatted := tenant.VerifiedAt.Format("2006-01-02T15:04:05Z07:00")
		verifiedAt = &formatted
	}

	// Return dashboard data
	c.JSON(http.StatusOK, DashboardResponse{
		Tenant: DashboardTenant{
			ID:               tenant.ID.String(),
			OrganizationName: tenant.OrganizationName,
			Status:           string(tenant.Status),
			CreatedAt:        tenant.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			VerifiedAt:       verifiedAt,
		},
		User: DashboardUser{
			ID:       userID.(string),
			FullName: user.FullName,
			Email:    user.Email,
			Role:     string(user.Role),
		},
	})
}
