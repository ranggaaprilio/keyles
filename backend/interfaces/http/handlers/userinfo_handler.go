package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/usecase/auth"
)

// UserinfoHandler handles the OIDC UserInfo endpoint
type UserinfoHandler struct {
	getUserInfoUC *auth.GetUserInfo
}

// NewUserinfoHandler creates a new UserInfo handler
func NewUserinfoHandler(getUserInfoUC *auth.GetUserInfo) *UserinfoHandler {
	return &UserinfoHandler{
		getUserInfoUC: getUserInfoUC,
	}
}

// UserInfo handles GET /oauth2/userinfo (FR-052)
// Returns user profile information from a valid access token
// Requires Bearer token authentication via middleware
func (h *UserinfoHandler) UserInfo(c *gin.Context) {
	// Extract user ID from JWT claims (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_token",
			"error_description": "access token is invalid or expired",
		})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_token",
			"error_description": "user_id claim is missing or invalid",
		})
		return
	}

	// Extract client ID from JWT claims (set by auth middleware)
	clientID, _ := c.Get("client_id")
	clientIDStr, _ := clientID.(string)

	// Execute use case
	userInfo, err := h.getUserInfoUC.Execute(c.Request.Context(), userIDStr, clientIDStr)
	if err != nil {
		if err.Error() == "user not found" || err.Error() == "invalid user_id format" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":             "not_found",
				"error_description": "user not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "failed to retrieve user information",
		})
		return
	}

	// Return UserInfo claims as JSON
	c.JSON(http.StatusOK, userInfo)
}
