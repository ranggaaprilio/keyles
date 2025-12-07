package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/usecase/tenant"
)

// AvailabilityHandler handles availability check HTTP requests
type AvailabilityHandler struct {
	checkAvailabilityUseCase *tenant.CheckAvailabilityUseCase
}

// NewAvailabilityHandler creates a new availability handler
func NewAvailabilityHandler(checkAvailabilityUseCase *tenant.CheckAvailabilityUseCase) *AvailabilityHandler {
	return &AvailabilityHandler{
		checkAvailabilityUseCase: checkAvailabilityUseCase,
	}
}

// AvailabilityResponse represents the availability check response
type AvailabilityResponse struct {
	OrganizationNameAvailable bool `json:"organization_name_available"`
	EmailAvailable            bool `json:"email_available"`
}

// CheckAvailability handles GET /api/v1/check-availability
func (h *AvailabilityHandler) CheckAvailability(c *gin.Context) {
	// Get query parameters
	organizationName := c.Query("organization_name")
	email := c.Query("email")

	// Validate required parameters
	if organizationName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "organization_name is required",
		})
		return
	}

	if email == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "email is required",
		})
		return
	}

	// Execute use case
	result, err := h.checkAvailabilityUseCase.Execute(
		c.Request.Context(),
		organizationName,
		email,
	)

	if err != nil {
		// Map domain errors to HTTP status codes
		statusCode := http.StatusInternalServerError
		errorMessage := "Internal server error"

		// Check for validation errors
		switch {
		case err.Error() == "organization name must be between 3 and 100 characters" ||
			err.Error() == "email must be a valid email address":
			statusCode = http.StatusBadRequest
			errorMessage = err.Error()
		default:
			// For unexpected errors, log but don't expose details
			errorMessage = "Failed to check availability"
		}

		c.JSON(statusCode, ErrorResponse{
			Error: errorMessage,
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, AvailabilityResponse{
		OrganizationNameAvailable: result.OrganizationNameAvailable,
		EmailAvailable:            result.EmailAvailable,
	})
}
