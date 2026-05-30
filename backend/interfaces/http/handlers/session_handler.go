package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/ranggaaprilio/keyles/usecase/user"
)

// SessionHandler handles admin session management endpoints
type SessionHandler struct {
	revokeTokenUC  *auth.RevokeToken
	listSessionsUC *user.ListSessions
	revokeSessionUC *user.RevokeSession
}

// NewSessionHandler creates a new session management handler
func NewSessionHandler(
	revokeTokenUC *auth.RevokeToken,
	listSessionsUC *user.ListSessions,
	revokeSessionUC *user.RevokeSession,
) *SessionHandler {
	return &SessionHandler{
		revokeTokenUC:   revokeTokenUC,
		listSessionsUC:  listSessionsUC,
		revokeSessionUC: revokeSessionUC,
	}
}

// RevokeUserSessionRequest represents a request to revoke user sessions
type RevokeUserSessionRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	ClientID string `json:"client_id" binding:"required"`
}

// RevokeUserSessions revokes all refresh tokens for a user-client pair (FR-049)
// This is an admin endpoint that allows administrators to terminate user sessions
func (h *SessionHandler) RevokeUserSessions(c *gin.Context) {
	var req RevokeUserSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "user_id and client_id are required",
		})
		return
	}

	// Get admin user context (from JWT middleware)
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Admin authentication required",
		})
		return
	}

	// Revoke all tokens for the user-client pair
	if err := h.revokeTokenUC.RevokeAllForUser(c.Request.Context(), req.UserID, req.ClientID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": "Failed to revoke user sessions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "User sessions revoked successfully",
		"user_id":   req.UserID,
		"client_id": req.ClientID,
	})
}

// ListUserSessions handles GET /api/v1/admin/users/:id/sessions
func (h *SessionHandler) ListUserSessions(c *gin.Context) {
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

	sessions, err := h.listSessionsUC.Execute(c.Request.Context(), userID, tenantID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"sessions": sessions,
	})
}

// RevokeUserSession handles DELETE /api/v1/admin/users/:id/sessions/:sessionId
func (h *SessionHandler) RevokeUserSession(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Session ID is required",
		})
		return
	}

	adminUserID, _ := c.Get("user_id")
	adminID, _ := adminUserID.(string)

	if err := h.revokeSessionUC.Execute(c.Request.Context(), userID, sessionID, adminID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Session revoked successfully",
		"user_id":    userID,
		"session_id": sessionID,
	})
}
