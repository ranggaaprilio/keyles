package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/usecase/tenant"
)

// VerificationHandler handles OTP verification HTTP requests
type VerificationHandler struct {
	verifyUseCase *tenant.VerifyTenantUseCase
}

// NewVerificationHandler creates a new verification handler
func NewVerificationHandler(verifyUseCase *tenant.VerifyTenantUseCase) *VerificationHandler {
	return &VerificationHandler{
		verifyUseCase: verifyUseCase,
	}
}

// VerifyOTPRequest represents the OTP verification request payload
type VerifyOTPRequest struct {
	TenantID string `json:"tenant_id" binding:"required"`
	OTPCode  string `json:"otp_code" binding:"required"`
}

// VerifyOTPResponse represents the successful verification response
type VerifyOTPResponse struct {
	TenantID string `json:"tenant_id"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

// VerifyOTP handles POST /api/v1/verify-otp
func (h *VerificationHandler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest

	// Bind and validate JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Additional validation for OTP format (6 digits)
	if len(req.OTPCode) != 6 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "OTP code must be 6 digits",
		})
		return
	}

	// Execute use case
	err := h.verifyUseCase.Execute(req.TenantID, req.OTPCode)

	if err != nil {
		// Map domain errors to HTTP status codes
		statusCode := http.StatusInternalServerError
		errorMessage := "Internal server error"

		errorMsg := err.Error()

		// Check for specific error types
		switch {
		case strings.Contains(errorMsg, "OTP not found"):
			statusCode = http.StatusNotFound
			errorMessage = "OTP not found. Please request a new code."
		case strings.Contains(errorMsg, "already been used"):
			statusCode = http.StatusBadRequest
			errorMessage = "This OTP has already been used. Please request a new code."
		case strings.Contains(errorMsg, "expired"):
			statusCode = http.StatusUnauthorized
			errorMessage = "OTP has expired. Please request a new code."
		case strings.Contains(errorMsg, "invalid"):
			statusCode = http.StatusUnauthorized
			errorMessage = "Invalid OTP code. Please try again."
		case strings.Contains(errorMsg, "tenant not found"):
			statusCode = http.StatusNotFound
			errorMessage = "Tenant not found."
		case strings.Contains(errorMsg, "already active"):
			statusCode = http.StatusBadRequest
			errorMessage = "Tenant is already verified and active."
		case strings.Contains(errorMsg, "invalid tenant ID"):
			statusCode = http.StatusBadRequest
			errorMessage = "Invalid tenant ID format."
		default:
			// For unexpected errors, log but don't expose details
			errorMessage = "Failed to verify OTP"
		}

		c.JSON(statusCode, ErrorResponse{
			Error: errorMessage,
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, VerifyOTPResponse{
		TenantID: req.TenantID,
		Status:   "active",
		Message:  "Email verified successfully. Your organization account is now active.",
	})
}
