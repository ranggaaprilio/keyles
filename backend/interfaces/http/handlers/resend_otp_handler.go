package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/usecase/tenant"
)

// ResendOTPHandler handles OTP resend HTTP requests
type ResendOTPHandler struct {
	resendUseCase *tenant.ResendOTPUseCase
}

// NewResendOTPHandler creates a new resend OTP handler
func NewResendOTPHandler(resendUseCase *tenant.ResendOTPUseCase) *ResendOTPHandler {
	return &ResendOTPHandler{
		resendUseCase: resendUseCase,
	}
}

// ResendOTPRequest represents the resend OTP request payload
type ResendOTPRequest struct {
	TenantID string `json:"tenant_id" binding:"required"`
}

// ResendOTPResponse represents the successful resend response
type ResendOTPResponse struct {
	TenantID string `json:"tenant_id"`
	Message  string `json:"message"`
}

// ResendOTP handles POST /api/v1/resend-otp
func (h *ResendOTPHandler) ResendOTP(c *gin.Context) {
	var req ResendOTPRequest

	// Bind and validate JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Execute use case
	err := h.resendUseCase.Execute(req.TenantID)

	if err != nil {
		// Map domain errors to HTTP status codes
		statusCode := http.StatusInternalServerError
		errorMessage := "Internal server error"

		errorMsg := err.Error()

		// Check for specific error types
		switch {
		case strings.Contains(errorMsg, "rate limit exceeded"):
			statusCode = http.StatusTooManyRequests
			errorMessage = "Too many OTP requests. Please wait 10 minutes before trying again."
		case strings.Contains(errorMsg, "OTP not found"):
			statusCode = http.StatusNotFound
			errorMessage = "No OTP found for this tenant. Please register first."
		case strings.Contains(errorMsg, "tenant not found"):
			statusCode = http.StatusNotFound
			errorMessage = "Tenant not found."
		case strings.Contains(errorMsg, "admin user not found"):
			statusCode = http.StatusNotFound
			errorMessage = "Admin user not found for this tenant."
		case strings.Contains(errorMsg, "invalid tenant ID"):
			statusCode = http.StatusBadRequest
			errorMessage = "Invalid tenant ID format."
		case strings.Contains(errorMsg, "email service unavailable"):
			statusCode = http.StatusServiceUnavailable
			errorMessage = "Email service is temporarily unavailable. Please try again later."
		default:
			// For unexpected errors, log but don't expose details
			errorMessage = "Failed to resend OTP"
		}

		c.JSON(statusCode, ErrorResponse{
			Error: errorMessage,
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, ResendOTPResponse{
		TenantID: req.TenantID,
		Message:  "OTP resent successfully. Please check your email.",
	})
}
