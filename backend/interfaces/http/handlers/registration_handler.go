package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/usecase/tenant"
)

// RegistrationHandler handles tenant registration HTTP requests
type RegistrationHandler struct {
	registerUseCase *tenant.RegisterTenantUseCase
}

// NewRegistrationHandler creates a new registration handler
func NewRegistrationHandler(registerUseCase *tenant.RegisterTenantUseCase) *RegistrationHandler {
	return &RegistrationHandler{
		registerUseCase: registerUseCase,
	}
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	OrganizationName string `json:"organization_name" binding:"required"`
	Email            string `json:"email" binding:"required,email"`
	Password         string `json:"password" binding:"required"`
	FullName         string `json:"full_name" binding:"required"`
}

// RegisterResponse represents the successful registration response
type RegisterResponse struct {
	TenantID         string `json:"tenant_id"`
	OrganizationName string `json:"organization_name"`
	Status           string `json:"status"`
	Message          string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// Register handles POST /api/v1/register
func (h *RegistrationHandler) Register(c *gin.Context) {
	var req RegisterRequest

	// Bind and validate JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Get client IP and user agent for audit logging (future enhancement)
	// ipAddress := c.GetHeader("X-Real-IP")
	// if ipAddress == "" {
	// 	ipAddress = c.ClientIP()
	// }
	// userAgent := c.GetHeader("User-Agent")

	// Execute use case
	result, err := h.registerUseCase.Execute(
		c.Request.Context(),
		req.OrganizationName,
		req.Email,
		req.Password,
		req.FullName,
	)

	if err != nil {
		// Map domain errors to HTTP status codes
		statusCode := http.StatusInternalServerError
		errorMessage := "Internal server error"

		// Check for specific error types
		switch {
		case err.Error() == "organization name already exists":
			statusCode = http.StatusConflict
			errorMessage = err.Error()
		case err.Error() == "email already exists":
			statusCode = http.StatusConflict
			errorMessage = err.Error()
		case err.Error() == "organization name must be between 3 and 100 characters" ||
			err.Error() == "email must be a valid email address" ||
			err.Error() == "password must be at least 8 characters long" ||
			err.Error() == "full name must be between 2 and 100 characters":
			statusCode = http.StatusBadRequest
			errorMessage = err.Error()
		default:
			// For unexpected errors, log but don't expose details
			errorMessage = "Failed to register tenant"
		}

		c.JSON(statusCode, ErrorResponse{
			Error: errorMessage,
		})
		return
	}

	// Success response
	c.JSON(http.StatusCreated, RegisterResponse{
		TenantID:         result.TenantID.String(),
		OrganizationName: result.OrganizationName,
		Status:           string(result.Status),
		Message:          "Registration successful. Please check your email for verification code.",
	})
}
