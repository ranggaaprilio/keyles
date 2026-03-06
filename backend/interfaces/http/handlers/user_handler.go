package handlers

import (
	"net/http"
	"strconv"
	"strings"

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
	disableUserUC      *user.DisableUser
	enableUserUC       *user.EnableUser
	deleteUserUC       *user.DeleteUser
	resendInvitationUC *user.ResendInvitation
}

// NewUserHandler creates a new user management handler
func NewUserHandler(
	inviteUserUC *user.InviteUser,
	getUserUC *user.GetUser,
	listUsersUC *user.ListUsers,
	updateUserUC *user.UpdateUser,
	disableUserUC *user.DisableUser,
	enableUserUC *user.EnableUser,
	deleteUserUC *user.DeleteUser,
	resendInvitationUC *user.ResendInvitation,
) *UserHandler {
	return &UserHandler{
		inviteUserUC:       inviteUserUC,
		getUserUC:          getUserUC,
		listUsersUC:        listUsersUC,
		updateUserUC:       updateUserUC,
		disableUserUC:      disableUserUC,
		enableUserUC:       enableUserUC,
		deleteUserUC:       deleteUserUC,
		resendInvitationUC: resendInvitationUC,
	}
}

// InviteUserRequest represents the JSON body for inviting a user
type InviteUserRequest struct {
	Email       string `json:"email" binding:"required,email"`
	DisplayName string `json:"display_name"`
}

// InviteUser handles POST /api/v1/admin/users (invite a new user)
func (h *UserHandler) InviteUser(c *gin.Context) {
	var req InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": "A valid email address is required.",
		})
		return
	}

	tenantID, adminID := getAdminContext(c)
	if tenantID == "" {
		return
	}

	// Use the frontend base URL from the X-Invite-Base-URL header, or fall back to origin
	inviteBaseURL := c.GetHeader("X-Invite-Base-URL")
	if inviteBaseURL == "" {
		inviteBaseURL = c.GetHeader("Origin")
	}

	output, err := h.inviteUserUC.Execute(c.Request.Context(), user.InviteUserInput{
		TenantID:      tenantID,
		Email:         req.Email,
		DisplayName:   req.DisplayName,
		InvitedBy:     adminID,
		InviteBaseURL: inviteBaseURL,
	})
	if err != nil {
		mapUserError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"invitation_id": output.InvitationID,
		"user_id":       output.UserID,
		"email":         output.Email,
		"display_name":  output.DisplayName,
		"status":        output.Status,
		"expires_at":    output.ExpiresAt,
	})
}

// ListUsers handles GET /api/v1/admin/users
func (h *UserHandler) ListUsers(c *gin.Context) {
	tenantID, _ := getAdminContext(c)
	if tenantID == "" {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "25"))
	search := c.Query("search")
	statusFilter := entities.UserStatus(c.Query("status"))

	output, err := h.listUsersUC.Execute(c.Request.Context(), user.ListUsersInput{
		TenantID:     tenantID,
		Search:       search,
		StatusFilter: statusFilter,
		Page:         page,
		PageSize:     pageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to list users.",
		})
		return
	}

	users := make([]gin.H, len(output.Users))
	for i, u := range output.Users {
		users[i] = gin.H{
			"id":           u.ID,
			"email":        u.Email,
			"display_name": u.DisplayName,
			"status":       u.Status,
			"last_login_at": u.LastLoginAt,
			"created_at":   u.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":        output.Page,
			"page_size":   output.PageSize,
			"total_count": output.Total,
			"total_pages": output.TotalPages,
		},
	})
}

// GetUser handles GET /api/v1/admin/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
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

	output, err := h.getUserUC.Execute(c.Request.Context(), userID, tenantID)
	if err != nil {
		mapUserError(c, err)
		return
	}

	roles := make([]gin.H, len(output.RoleAssignments))
	for i, r := range output.RoleAssignments {
		role := gin.H{
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
			role["revoked_at"] = r.RevokedAt
		}
		if r.RevokedBy != nil {
			role["revoked_by"] = *r.RevokedBy
		}
		roles[i] = role
	}

	u := output.User
	c.JSON(http.StatusOK, gin.H{
		"id":               u.ID,
		"email":            u.Email,
		"display_name":     u.DisplayName,
		"status":           u.Status,
		"last_login_at":    u.LastLoginAt,
		"created_at":       u.CreatedAt,
		"updated_at":       u.UpdatedAt,
		"role_assignments": roles,
	})
}

// UpdateUserRequest represents the JSON body for updating a user
type UpdateUserRequest struct {
	DisplayName *string `json:"display_name"`
}

// UpdateUser handles PATCH /api/v1/admin/users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
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

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": "Invalid request body.",
		})
		return
	}

	updatedUser, err := h.updateUserUC.Execute(c.Request.Context(), user.UpdateUserInput{
		UserID:      userID,
		TenantID:    tenantID,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		mapUserError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           updatedUser.ID,
		"email":        updatedUser.Email,
		"display_name": updatedUser.DisplayName,
		"status":       updatedUser.Status,
		"last_login_at": updatedUser.LastLoginAt,
		"created_at":   updatedUser.CreatedAt,
		"updated_at":   updatedUser.UpdatedAt,
	})
}

// UpdateUserStatusRequest represents the JSON body for enabling/disabling a user
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateUserStatus handles PUT /api/v1/admin/users/:id/status
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
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

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": "A valid status (active or disabled) is required.",
		})
		return
	}

	switch req.Status {
	case "disabled":
		err := h.disableUserUC.Execute(c.Request.Context(), user.DisableUserInput{
			TargetUserID: userID,
			AdminUserID:  adminID,
			TenantID:     tenantID,
		})
		if err != nil {
			mapUserError(c, err)
			return
		}
	case "active":
		err := h.enableUserUC.Execute(c.Request.Context(), user.EnableUserInput{
			TargetUserID: userID,
			AdminUserID:  adminID,
			TenantID:     tenantID,
		})
		if err != nil {
			mapUserError(c, err)
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": "Status must be 'active' or 'disabled'.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":    userID,
		"status":     req.Status,
		"updated_at": nil, // let client refetch for precise timestamp
	})
}

// DeleteUser handles DELETE /api/v1/admin/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
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

	err := h.deleteUserUC.Execute(c.Request.Context(), user.DeleteUserInput{
		TargetUserID: userID,
		AdminUserID:  adminID,
		TenantID:     tenantID,
	})
	if err != nil {
		mapUserError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ResendInvitation handles POST /api/v1/admin/users/:id/resend-invitation
func (h *UserHandler) ResendInvitation(c *gin.Context) {
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

	inviteBaseURL := c.GetHeader("X-Invite-Base-URL")
	if inviteBaseURL == "" {
		inviteBaseURL = c.GetHeader("Origin")
	}

	err := h.resendInvitationUC.Execute(c.Request.Context(), user.ResendInvitationInput{
		UserID:        userID,
		TenantID:      tenantID,
		InviteBaseURL: inviteBaseURL,
	})
	if err != nil {
		mapUserError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation resent successfully.",
	})
}

// getAdminContext extracts tenant_id and user_id from the JWT-authenticated context.
// Returns empty string for tenantID and sends 401 if not authenticated.
func getAdminContext(c *gin.Context) (tenantID, adminID string) {
	tid, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Admin authentication required.",
		})
		return "", ""
	}
	tenantID, _ = tid.(string)

	aid, _ := c.Get("user_id")
	adminID, _ = aid.(string)
	return tenantID, adminID
}

// mapUserError maps use case errors to appropriate HTTP responses
func mapUserError(c *gin.Context, err error) {
	msg := err.Error()

	switch {
	case strings.Contains(msg, "quota exceeded"):
		c.JSON(http.StatusConflict, gin.H{
			"error":   "quota_exceeded",
			"message": msg,
		})
	case strings.Contains(msg, "already exists"):
		c.JSON(http.StatusConflict, gin.H{
			"error":   "conflict",
			"message": msg,
		})
	case strings.Contains(msg, "pending invitation"):
		c.JSON(http.StatusConflict, gin.H{
			"error":   "conflict",
			"message": msg,
		})
	case strings.Contains(msg, "not found"):
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "The requested resource was not found.",
		})
	case strings.Contains(msg, "cannot disable your own"):
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": msg,
		})
	case strings.Contains(msg, "cannot disable an administrator"):
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": msg,
		})
	case strings.Contains(msg, "cannot delete your own"):
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": msg,
		})
	case strings.Contains(msg, "not in pending status"):
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_state",
			"message": "Invitations can only be resent for users in 'pending' status.",
		})
	case strings.Contains(msg, "can only enable disabled"):
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_state",
			"message": msg,
		})
	case strings.Contains(msg, "display name must not exceed"):
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": msg,
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "An internal error occurred.",
		})
	}
}
