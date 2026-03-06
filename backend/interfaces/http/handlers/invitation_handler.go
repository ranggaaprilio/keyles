package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/usecase/user"
)

// InvitationHandler handles public invitation endpoints (no auth required)
type InvitationHandler struct {
	acceptInvitationUC *user.AcceptInvitation
	invitationRepo     repositories.InvitationRepository
}

// NewInvitationHandler creates a new invitation handler
func NewInvitationHandler(
	acceptInvitationUC *user.AcceptInvitation,
	invitationRepo repositories.InvitationRepository,
) *InvitationHandler {
	return &InvitationHandler{
		acceptInvitationUC: acceptInvitationUC,
		invitationRepo:     invitationRepo,
	}
}

// ValidateInvitation handles GET /api/v1/invitations/:token/validate
// Returns invitation details without consuming the token.
func (h *InvitationHandler) ValidateInvitation(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invitation token is required.",
		})
		return
	}

	invitation, err := h.invitationRepo.GetByToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusGone, gin.H{
			"error":   "invitation_invalid",
			"message": "This invitation link is invalid, expired, or has already been used.",
		})
		return
	}

	if invitation.IsExpired() {
		c.JSON(http.StatusGone, gin.H{
			"error":   "invitation_expired",
			"message": "This invitation link has expired. Please contact your administrator to request a new invitation.",
		})
		return
	}

	if invitation.IsAccepted() {
		c.JSON(http.StatusGone, gin.H{
			"error":   "invitation_used",
			"message": "This invitation link has already been used.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email":        invitation.Email,
		"display_name": invitation.DisplayName,
		"expires_at":   invitation.ExpiresAt,
	})
}

// AcceptInvitationRequest represents the JSON body for accepting an invitation
type AcceptInvitationRequest struct {
	Password string `json:"password" binding:"required"`
}

// AcceptInvitation handles POST /api/v1/invitations/:token/accept
func (h *InvitationHandler) AcceptInvitation(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invitation token is required.",
		})
		return
	}

	var req AcceptInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": "Password is required.",
		})
		return
	}

	err := h.acceptInvitationUC.Execute(c.Request.Context(), user.AcceptInvitationInput{
		Token:    token,
		Password: req.Password,
	})
	if err != nil {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "expired"):
			c.JSON(http.StatusGone, gin.H{
				"error":   "invitation_expired",
				"message": "This invitation link has expired. Please contact your administrator to request a new invitation.",
			})
		case strings.Contains(msg, "already been accepted"):
			c.JSON(http.StatusGone, gin.H{
				"error":   "invitation_used",
				"message": "This invitation link has already been used.",
			})
		case strings.Contains(msg, "password") || strings.Contains(msg, "Password"):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_error",
				"message": "Password does not meet minimum strength requirements.",
			})
		case strings.Contains(msg, "invalid invitation token"):
			c.JSON(http.StatusGone, gin.H{
				"error":   "invitation_invalid",
				"message": "This invitation link is invalid, expired, or has already been used.",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "server_error",
				"message": "An internal error occurred.",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account activated successfully. You may now sign in.",
	})
}
