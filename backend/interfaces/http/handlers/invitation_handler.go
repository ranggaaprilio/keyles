package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/usecase/user"
)

// InvitationHandler handles public invitation endpoints
type InvitationHandler struct {
	acceptInvitationUC *user.AcceptInvitation
}

// NewInvitationHandler creates a new invitation handler
func NewInvitationHandler(acceptInvitationUC *user.AcceptInvitation) *InvitationHandler {
	return &InvitationHandler{
		acceptInvitationUC: acceptInvitationUC,
	}
}

// AcceptInvitationRequest represents the request body for accepting an invitation
type AcceptInvitationRequest struct {
	Password string `json:"password" binding:"required"`
}

// AcceptInvitation handles POST /api/v1/invitations/:token/accept
func (h *InvitationHandler) AcceptInvitation(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "token is required",
		})
		return
	}

	var req AcceptInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "password is required",
		})
		return
	}

	err := h.acceptInvitationUC.Execute(c.Request.Context(), user.AcceptInvitationRequest{
		PlainToken: token,
		Password:   req.Password,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "server_error"
		if contains(err.Error(), "expired") {
			statusCode = http.StatusGone
			errorCode = "invitation_expired"
		} else if contains(err.Error(), "already been accepted") || contains(err.Error(), "no longer valid") {
			statusCode = http.StatusGone
			errorCode = "invitation_expired"
		} else if contains(err.Error(), "not found") {
			statusCode = http.StatusGone
			errorCode = "invitation_expired"
		} else if contains(err.Error(), "invalid password") {
			statusCode = http.StatusBadRequest
			errorCode = "invalid_request"
		}
		c.JSON(statusCode, gin.H{
			"error":             errorCode,
			"error_description": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account activated successfully",
	})
}
