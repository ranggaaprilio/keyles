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
	UserID   string `json:"user_id" binding:"required"`
	ClientID string `json:"client_id" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

// RevokeRoleRequest represents the request body for revoking a role
type RevokeRoleRequest struct {
	AssignmentID int64  `json:"assignment_id" binding:"required"`
	UserID       string `json:"user_id" binding:"required"`
	ClientID     string `json:"client_id"`
}

// AssignRole handles POST /api/admin/roles/assign (FR-006a)
func (h *RoleHandler) AssignRole(c *gin.Context) {
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "user_id, client_id, and role are required",
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
		UserID:    req.UserID,
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
		"user_id":   req.UserID,
		"client_id": req.ClientID,
		"role":      req.Role,
	})
}

// RevokeRole handles POST /api/admin/roles/revoke (FR-006b)
func (h *RoleHandler) RevokeRole(c *gin.Context) {
	var req RevokeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "assignment_id and user_id are required",
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
	revokedBy, _ := adminID.(string)

	err := h.revokeRoleUC.Execute(c.Request.Context(), role.RevokeRoleRequest{
		AssignmentID: req.AssignmentID,
		UserID:       req.UserID,
		ClientID:     req.ClientID,
		TenantID:     tenantID.(string),
		RevokedBy:    revokedBy,
	})

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
		"assignment_id": req.AssignmentID,
		"user_id":       req.UserID,
	})
}

// ListUserRoles handles GET /api/admin/roles/users/:userId
func (h *RoleHandler) ListUserRoles(c *gin.Context) {
	userID := c.Param("userId")
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

	// Convert to response format
	rolesResponse := make([]gin.H, len(resp.Roles))
	for i, r := range resp.Roles {
		rolesResponse[i] = gin.H{
			"id":         r.ID,
			"user_id":    r.UserID,
			"client_id":  r.ClientID,
			"tenant_id":  r.TenantID,
			"role":       r.Role,
			"is_active":  r.IsActive,
			"granted_at": r.GrantedAt,
			"granted_by": r.GrantedBy,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"roles":   rolesResponse,
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

// --- New handlers for /api/v1/admin/users/:id/roles pattern ---

// AssignUserRoleRequest represents the JSON body for assigning a role via user-scoped endpoint
type AssignUserRoleRequest struct {
	ClientID string `json:"client_id" binding:"required"`
	RoleName string `json:"role_name" binding:"required"`
}

// AssignUserRole handles POST /api/v1/admin/users/:id/roles
func (h *RoleHandler) AssignUserRole(c *gin.Context) {
	tenantID, adminID := getAdminContext(c)
	if tenantID == "" {
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "User ID is required.",
		})
		return
	}

	var req AssignUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": "client_id and role_name are required.",
		})
		return
	}

	err := h.assignRoleUC.Execute(c.Request.Context(), role.AssignRoleRequest{
		UserID:    userID,
		ClientID:  req.ClientID,
		TenantID:  tenantID,
		Role:      req.RoleName,
		GrantedBy: adminID,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := err.Error()
		if contains(errMsg, "not found") {
			statusCode = http.StatusNotFound
		} else if contains(errMsg, "already has role") {
			statusCode = http.StatusConflict
		} else if contains(errMsg, "invalid") {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{
			"error":   "role_assignment_failed",
			"message": errMsg,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Role assigned successfully",
		"user_id":   userID,
		"client_id": req.ClientID,
		"role_name": req.RoleName,
	})
}

// RevokeUserRole handles DELETE /api/v1/admin/users/:id/roles/:assignmentId
func (h *RoleHandler) RevokeUserRole(c *gin.Context) {
	tenantID, adminID := getAdminContext(c)
	if tenantID == "" {
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "User ID is required.",
		})
		return
	}

	assignmentIDStr := c.Param("assignmentId")
	assignmentID, err := strconv.ParseInt(assignmentIDStr, 10, 64)
	if err != nil || assignmentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "A valid assignment ID is required.",
		})
		return
	}

	err = h.revokeRoleUC.Execute(c.Request.Context(), role.RevokeRoleRequest{
		AssignmentID: assignmentID,
		UserID:       userID,
		TenantID:     tenantID,
		RevokedBy:    adminID,
	})
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

	c.Status(http.StatusNoContent)
}

// ListUserRolesForUser handles GET /api/v1/admin/users/:id/roles
func (h *RoleHandler) ListUserRolesForUser(c *gin.Context) {
	tenantID, _ := getAdminContext(c)
	if tenantID == "" {
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "User ID is required.",
		})
		return
	}

	clientID := c.Query("client_id")

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

	rolesResponse := make([]gin.H, len(resp.Roles))
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
		if r.RevokedAt != nil {
			roleItem["revoked_at"] = r.RevokedAt
		}
		if r.RevokedBy != nil {
			roleItem["revoked_by"] = *r.RevokedBy
		}
		rolesResponse[i] = roleItem
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"roles":   rolesResponse,
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
