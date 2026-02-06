package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/usecase/auth"
)

// SessionHandler handles admin session management endpoints
type SessionHandler struct {
	revokeTokenUC *auth.RevokeToken
}

// NewSessionHandler creates a new session management handler
func NewSessionHandler(revokeTokenUC *auth.RevokeToken) *SessionHandler {
	return &SessionHandler{
		revokeTokenUC: revokeTokenUC,
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
		"message": "User sessions revoked successfully",
		"user_id": req.UserID,
		"client_id": req.ClientID,
	})
}
