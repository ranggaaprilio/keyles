package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/usecase/role"
)

// RoleHandler handles admin role management endpoints
type RoleHandler struct {
	assignRoleUC    *role.AssignRole
	revokeRoleUC    *role.RevokeRole
	listUserRolesUC *role.ListUserRoles
}

// NewRoleHandler creates a new role management handler
func NewRoleHandler(
	assignRoleUC *role.AssignRole,
	revokeRoleUC *role.RevokeRole,
	listUserRolesUC *role.ListUserRoles,
) *RoleHandler {
	return &RoleHandler{
		assignRoleUC:    assignRoleUC,
		revokeRoleUC:    revokeRoleUC,
		listUserRolesUC: listUserRolesUC,
	}
}

// AssignRoleRequest represents the request body for assigning a role
type AssignRoleRequest struct {
	UserID   string `json:"user_id"`
	ClientID string `json:"client_id" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

// RevokeRoleRequest represents the request body for revoking a role
type RevokeRoleRequest struct {
	AssignmentID int64 `json:"assignment_id"`
}

// AssignRole handles POST /api/admin/roles/assign (FR-006a) and POST /api/v1/admin/users/:id/roles
func (h *RoleHandler) AssignRole(c *gin.Context) {
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "client_id and role are required",
		})
		return
	}

	// Support both legacy body param and new path param
	userID := c.Param("id")
	if userID == "" {
		userID = req.UserID
	}
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "user_id is required",
		})
		return
	}

	// Get admin context from JWT middleware
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Admin authentication required",
		})
		return
	}

	adminID, _ := c.Get("user_id")
	grantedBy, _ := adminID.(string)

	err := h.assignRoleUC.Execute(c.Request.Context(), role.AssignRoleRequest{
		UserID:    userID,
		ClientID:  req.ClientID,
		TenantID:  tenantID.(string),
		Role:      req.Role,
		GrantedBy: grantedBy,
	})

	if err != nil {
		// Determine appropriate status code based on error
		statusCode := http.StatusInternalServerError
		if contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if contains(err.Error(), "already has role") {
			statusCode = http.StatusConflict
		} else if contains(err.Error(), "invalid") {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, gin.H{
			"error":   "role_assignment_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Role assigned successfully",
		"user_id":   userID,
		"client_id": req.ClientID,
		"role":      req.Role,
	})
}

// RevokeRole handles POST /api/admin/roles/revoke (FR-006b) and DELETE /api/v1/admin/users/:id/roles/:assignmentId
func (h *RoleHandler) RevokeRole(c *gin.Context) {
	var req RevokeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// For DELETE with path param, body may be empty; ignore binding error
	}

	// Support both legacy body param and new path param
	assignmentID := req.AssignmentID
	if assignmentID == 0 {
		assignmentIDStr := c.Param("assignmentId")
		if assignmentIDStr != "" {
			var err error
			assignmentID, err = strconv.ParseInt(assignmentIDStr, 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "invalid_request",
					"message": "assignmentId must be a valid integer",
				})
				return
			}
		}
	}
	if assignmentID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "assignment_id is required",
		})
		return
	}

	_, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Admin authentication required",
		})
		return
	}

	adminID, _ := c.Get("user_id")
	revokedBy, _ := adminID.(string)

	err := h.revokeRoleUC.Execute(c.Request.Context(), assignmentID, revokedBy)

	if err != nil {
		statusCode := http.StatusInternalServerError
		if contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "role_revocation_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Role revoked successfully",
		"assignment_id": assignmentID,
	})
}

// ListUserRoles handles GET /api/admin/roles/users/:userId and GET /api/v1/admin/users/:id/roles
func (h *RoleHandler) ListUserRoles(c *gin.Context) {
	// Support both legacy userId param and new id param
	userID := c.Param("userId")
	if userID == "" {
		userID = c.Param("id")
	}
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "userId parameter is required",
		})
		return
	}

	// Optional client filter
	clientID := c.Query("client_id")

	// Get admin context from JWT middleware
	_, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Admin authentication required",
		})
		return
	}

	resp, err := h.listUserRolesUC.Execute(c.Request.Context(), role.ListUserRolesRequest{
		UserID:   userID,
		ClientID: clientID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "list_roles_failed",
			"message": err.Error(),
		})
		return
	}

	// Convert to flat response format
	rolesResponse := make([]gin.H, len(resp.Roles))
	// Group by clientID
	rolesByClient := make(map[string][]gin.H)
	for i, r := range resp.Roles {
		roleItem := gin.H{
			"id":         r.ID,
			"user_id":    r.UserID,
			"client_id":  r.ClientID,
			"tenant_id":  r.TenantID,
			"role":       r.Role,
			"is_active":  r.IsActive,
			"granted_at": r.GrantedAt,
			"granted_by": r.GrantedBy,
		}
		rolesResponse[i] = roleItem
		if r.IsActive {
			rolesByClient[r.ClientID] = append(rolesByClient[r.ClientID], roleItem)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":         userID,
		"roles":           rolesResponse,
		"roles_by_client": rolesByClient,
	})
}

// ListClientRoles handles GET /api/admin/roles/clients/:clientId
func (h *RoleHandler) ListClientRoles(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "clientId parameter is required",
		})
		return
	}

	// Get admin context from JWT middleware
	_, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Admin authentication required",
		})
		return
	}

	// This would need a new use case, but for now we can use the repository directly
	// In a full implementation, create a ListClientRoles use case
	c.JSON(http.StatusOK, gin.H{
		"client_id": clientID,
		"roles":     []gin.H{},
		"message":   "List client roles endpoint - implementation pending",
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
