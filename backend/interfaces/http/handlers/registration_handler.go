package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/domain/entities"
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

	// Execute use case
	result, err := h.registerUseCase.Execute(
		c.Request.Context(),
		req.OrganizationName,
		req.Email,
		req.Password,
		req.FullName,
	)

	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := "Internal server error"

		switch {
		case errors.Is(err, tenant.ErrOrganizationNameExists),
			errors.Is(err, tenant.ErrEmailExists):
			statusCode = http.StatusConflict
			errorMessage = err.Error()
		case errors.Is(err, entities.ErrInvalidOrganizationName),
			errors.Is(err, entities.ErrInvalidEmail),
			errors.Is(err, entities.ErrPasswordTooShort),
			errors.Is(err, entities.ErrPasswordMissingUpper),
			errors.Is(err, entities.ErrPasswordMissingLower),
			errors.Is(err, entities.ErrPasswordMissingNumber),
			errors.Is(err, entities.ErrPasswordMissingSpecial),
			errors.Is(err, entities.ErrInvalidFullName):
			statusCode = http.StatusBadRequest
			errorMessage = err.Error()
		default:
			slog.Error("registration failed", "error", err)
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
		Message:          result.Message,
	})
}
