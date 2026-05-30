package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/usecase/user"
)

// UserHandler handles admin user management endpoints
type UserHandler struct {
	inviteUserUC       *user.InviteUser
	getUserUC          *user.GetUser
	listUsersUC        *user.ListUsers
	updateUserUC       *user.UpdateUser
	enableUserUC       *user.EnableUser
	disableUserUC      *user.DisableUser
	deleteUserUC       *user.DeleteUser
	resendInvitationUC *user.ResendInvitation
	listSessionsUC     *user.ListSessions
}

// NewUserHandler creates a new user management handler
func NewUserHandler(
	inviteUserUC *user.InviteUser,
	getUserUC *user.GetUser,
	listUsersUC *user.ListUsers,
	updateUserUC *user.UpdateUser,
	enableUserUC *user.EnableUser,
	disableUserUC *user.DisableUser,
	deleteUserUC *user.DeleteUser,
	resendInvitationUC *user.ResendInvitation,
	listSessionsUC *user.ListSessions,
) *UserHandler {
	return &UserHandler{
		inviteUserUC:       inviteUserUC,
		getUserUC:          getUserUC,
		listUsersUC:        listUsersUC,
		updateUserUC:       updateUserUC,
		enableUserUC:       enableUserUC,
		disableUserUC:      disableUserUC,
		deleteUserUC:       deleteUserUC,
		resendInvitationUC: resendInvitationUC,
		listSessionsUC:     listSessionsUC,
	}
}

// InviteUserRequest represents the request body for inviting a user
type InviteUserRequest struct {
	Email       string `json:"email" binding:"required,email"`
	DisplayName string `json:"display_name"`
}

// InviteUser handles POST /api/v1/admin/users/invite
func (h *UserHandler) InviteUser(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tenant not found in token",
		})
		return
	}

	adminUserID, _ := c.Get("user_id")
	invitedBy, _ := adminUserID.(string)

	var req InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	resp, err := h.inviteUserUC.Execute(c.Request.Context(), user.InviteUserRequest{
		TenantID:    tenantID.(string),
		Email:       req.Email,
		DisplayName: req.DisplayName,
		InvitedBy:   invitedBy,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if contains(err.Error(), "already exists") {
			statusCode = http.StatusConflict
		} else if contains(err.Error(), "invalid") {
			statusCode = http.StatusBadRequest
		} else if contains(err.Error(), "quota exceeded") || contains(err.Error(), "limit reached") {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":           resp.Invitation.ID,
		"email":        resp.Invitation.Email,
		"display_name": resp.Invitation.DisplayName,
		"status":       resp.Invitation.Status,
		"expires_at":   resp.Invitation.ExpiresAt,
	})
}

// GetUser handles GET /api/v1/admin/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tenant not found in token",
		})
		return
	}

	resp, err := h.getUserUC.Execute(c.Request.Context(), user.GetUserRequest{
		UserID:   userID,
		TenantID: tenantID.(string),
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if contains(err.Error(), "tenant mismatch") {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Count active sessions
	sessions, _ := h.listSessionsUC.Execute(c.Request.Context(), userID, tenantID.(string))

	c.JSON(http.StatusOK, gin.H{
		"id":                resp.ID,
		"email":             resp.Email,
		"display_name":      resp.DisplayName,
		"status":            resp.Status,
		"last_login_at":     resp.LastLoginAt,
		"created_at":        resp.CreatedAt,
		"updated_at":        resp.UpdatedAt,
		"roles_by_client":   resp.RolesByClient,
		"active_sessions":   len(sessions),
	})
}

// ListUsers handles GET /api/v1/admin/users
func (h *UserHandler) ListUsers(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tenant not found in token",
		})
		return
	}

	search := c.Query("search")
	statusFilter := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "25"))

	var status entities.UserStatus
	if statusFilter != "" {
		switch statusFilter {
		case "pending":
			status = entities.UserStatusPending
		case "active":
			status = entities.UserStatusActive
		case "disabled":
			status = entities.UserStatusDisabled
		}
	}

	resp, err := h.listUsersUC.Execute(c.Request.Context(), user.ListUsersRequest{
		TenantID:     tenantID.(string),
		Search:       search,
		StatusFilter: status,
		Page:         page,
		PageSize:     pageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":       resp.Users,
		"total_count": resp.TotalCount,
		"page":        resp.Page,
		"page_size":   resp.PageSize,
		"total_pages": resp.TotalPages,
	})
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	DisplayName string `json:"display_name"`
}

// UpdateUser handles PATCH /api/v1/admin/users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tenant not found in token",
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	resp, err := h.updateUserUC.Execute(c.Request.Context(), user.UpdateUserRequest{
		UserID:      userID,
		TenantID:    tenantID.(string),
		DisplayName: req.DisplayName,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if contains(err.Error(), "tenant mismatch") {
			statusCode = http.StatusForbidden
		} else if contains(err.Error(), "invalid") {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           resp.ID,
		"email":        resp.Email,
		"display_name": resp.DisplayName,
		"status":       resp.Status,
		"last_login_at": resp.LastLoginAt,
		"created_at":   resp.CreatedAt,
		"updated_at":   resp.UpdatedAt,
	})
}

// UpdateUserStatusRequest represents the request body for updating user status
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active disabled"`
}

// UpdateUserStatus handles PATCH /api/v1/admin/users/:id/status
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tenant not found in token",
		})
		return
	}

	adminUserID, _ := c.Get("user_id")
	adminID, _ := adminUserID.(string)

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	var err error
	if req.Status == "active" {
		err = h.enableUserUC.Execute(c.Request.Context(), userID, tenantID.(string))
	} else if req.Status == "disabled" {
		err = h.disableUserUC.Execute(c.Request.Context(), user.DisableUserRequest{
			TargetUserID: userID,
			AdminUserID:  adminID,
			TenantID:     tenantID.(string),
		})
	}

	if err != nil {
		statusCode := http.StatusInternalServerError
		if contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if contains(err.Error(), "tenant mismatch") {
			statusCode = http.StatusForbidden
		} else if contains(err.Error(), "cannot disable your own account") || contains(err.Error(), "cannot delete your own account") {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     userID,
		"status": req.Status,
	})
}

// DeleteUser handles DELETE /api/v1/admin/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tenant not found in token",
		})
		return
	}

	adminUserID, _ := c.Get("user_id")
	adminID, _ := adminUserID.(string)

	err := h.deleteUserUC.Execute(c.Request.Context(), user.DeleteUserRequest{
		TargetUserID: userID,
		AdminUserID:  adminID,
		TenantID:     tenantID.(string),
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if contains(err.Error(), "tenant mismatch") {
			statusCode = http.StatusForbidden
		} else if contains(err.Error(), "cannot delete your own account") {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ResendInvitation handles POST /api/v1/admin/users/:id/resend-invitation
func (h *UserHandler) ResendInvitation(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tenant not found in token",
		})
		return
	}

	err := h.resendInvitationUC.Execute(c.Request.Context(), user.ResendInvitationRequest{
		UserID:   userID,
		TenantID: tenantID.(string),
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if contains(err.Error(), "tenant mismatch") {
			statusCode = http.StatusForbidden
		} else if contains(err.Error(), "not in pending status") {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation resent successfully",
		"user_id": userID,
	})
}
