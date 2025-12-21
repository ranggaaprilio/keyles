/**
 * Authentication HTTP handlers
 */

package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/usecase/auth"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authenticateUseCase *auth.AuthenticateAdminUseCase
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authenticateUseCase *auth.AuthenticateAdminUseCase) *AuthHandler {
	return &AuthHandler{
		authenticateUseCase: authenticateUseCase,
	}
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token     string     `json:"token"`
	ExpiresIn int        `json:"expires_in"`
	User      UserInfo   `json:"user"`
	Tenant    TenantInfo `json:"tenant"`
}

// UserInfo represents user information in the response
type UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

// TenantInfo represents tenant information in the response
type TenantInfo struct {
	ID               string `json:"id"`
	OrganizationName string `json:"organization_name"`
	Status           string `json:"status"`
}

// Login handles POST /api/v1/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// Execute authentication use case
	result, err := h.authenticateUseCase.Execute(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		// Map errors to appropriate HTTP status codes
		statusCode := http.StatusInternalServerError
		errorMessage := "Failed to authenticate"

		errStr := err.Error()
		if strings.Contains(errStr, "user not found") || strings.Contains(errStr, "invalid credentials") {
			statusCode = http.StatusUnauthorized
			errorMessage = "Invalid email or password"
		} else if strings.Contains(errStr, "tenant not active") {
			statusCode = http.StatusForbidden
			errorMessage = "Please verify your email before logging in"
		} else if strings.Contains(errStr, "tenant not found") {
			statusCode = http.StatusNotFound
			errorMessage = "Tenant not found"
		}

		c.JSON(statusCode, gin.H{
			"error": errorMessage,
		})
		return
	}

	// Return successful login response
	c.JSON(http.StatusOK, LoginResponse{
		Token:     result.Token,
		ExpiresIn: result.ExpiresIn,
		User: UserInfo{
			ID:       result.User.ID,
			Email:    result.User.Email,
			FullName: result.User.FullName,
			Role:     result.User.Role,
		},
		Tenant: TenantInfo{
			ID:               result.Tenant.ID,
			OrganizationName: result.Tenant.OrganizationName,
			Status:           result.Tenant.Status,
		},
	})
}
