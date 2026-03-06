package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/ranggaaprilio/keyles/usecase/user"
)

// SessionHandler handles admin session management endpoints
type SessionHandler struct {
	revokeTokenUC    *auth.RevokeToken
	listSessionsUC   *user.ListSessions
	revokeSessionUC  *user.RevokeSession
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

	sessions, err := h.listSessionsUC.Execute(c.Request.Context(), user.ListSessionsInput{
		UserID:   userID,
		TenantID: tenantID,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{
			"error":   "list_sessions_failed",
			"message": err.Error(),
		})
		return
	}

	sessionList := make([]gin.H, len(sessions))
	for i, s := range sessions {
		item := gin.H{
			"id":         s.ID,
			"client_id":  s.ClientID,
			"created_at": s.CreatedAt,
			"expires_at": s.ExpiresAt,
		}
		if s.LastUsedAt != nil {
			item["last_used_at"] = s.LastUsedAt
		}
		sessionList[i] = item
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"sessions": sessionList,
	})
}

// RevokeUserSession handles DELETE /api/v1/admin/users/:id/sessions/:sessionId
func (h *SessionHandler) RevokeUserSession(c *gin.Context) {
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

	sessionIDStr := c.Param("sessionId")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil || sessionID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "A valid session ID is required.",
		})
		return
	}

	err = h.revokeSessionUC.Execute(c.Request.Context(), user.RevokeSessionInput{
		UserID:    userID,
		TenantID:  tenantID,
		TokenID:   sessionID,
		RevokedBy: adminID,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := err.Error()
		if contains(errMsg, "not found") {
			statusCode = http.StatusNotFound
		} else if contains(errMsg, "already revoked") {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{
			"error":   "revoke_session_failed",
			"message": errMsg,
		})
		return
	}

	c.Status(http.StatusNoContent)
}
