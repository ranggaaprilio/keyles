package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ranggaaprilio/keyles/infrastructure/config"
	"github.com/ranggaaprilio/keyles/usecase/auth"
)

// OAuthSignupHandler serves end-user self sign-up within the browser OAuth flow.
type OAuthSignupHandler struct {
	selfSignup *auth.SelfSignup
	config     *config.Config
}

// NewOAuthSignupHandler creates a new OAuthSignupHandler.
func NewOAuthSignupHandler(selfSignup *auth.SelfSignup, cfg *config.Config) *OAuthSignupHandler {
	return &OAuthSignupHandler{selfSignup: selfSignup, config: cfg}
}

type signupRequest struct {
	TransactionID string `json:"transaction_id"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	DisplayName   string `json:"display_name"`
}

type signupVerifyRequest struct {
	TransactionID string `json:"transaction_id"`
	Email         string `json:"email"`
	Code          string `json:"code"`
}

type signupResendRequest struct {
	TransactionID string `json:"transaction_id"`
	Email         string `json:"email"`
}

// verificationSent is returned for every accepted sign-up/resend request,
// regardless of whether the email is already registered.
var verificationSent = gin.H{
	"status":               "verification_sent",
	"resend_after_seconds": int(auth.SignupResendCooldown.Seconds()),
}

// Context handles GET /oauth2/signup/context?transaction_id=...
func (h *OAuthSignupHandler) Context(c *gin.Context) {
	result, err := h.selfSignup.GetContext(c.Request.Context(), c.Query("transaction_id"))
	if err != nil {
		writeSignupError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"client_name":    result.ClientName,
		"signup_enabled": result.SignupEnabled,
	})
}

// Signup handles POST /oauth2/signup.
func (h *OAuthSignupHandler) Signup(c *gin.Context) {
	var req signupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeSignupBadRequest(c, "failed to parse request body")
		return
	}
	err := h.selfSignup.Request(c.Request.Context(), auth.SignupRequest{
		TransactionID: req.TransactionID,
		Email:         req.Email,
		Password:      req.Password,
		DisplayName:   req.DisplayName,
		SourceIP:      directPeerIP(c),
	})
	if err != nil {
		writeSignupError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, verificationSent)
}

// Resend handles POST /oauth2/signup/resend.
func (h *OAuthSignupHandler) Resend(c *gin.Context) {
	var req signupResendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeSignupBadRequest(c, "failed to parse request body")
		return
	}
	if err := h.selfSignup.Resend(c.Request.Context(), req.TransactionID, req.Email); err != nil {
		writeSignupError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, verificationSent)
}

// Verify handles POST /oauth2/signup/verify. On success it sets the SSO
// session cookie and returns the consent redirect, like Login.
func (h *OAuthSignupHandler) Verify(c *gin.Context) {
	var req signupVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeSignupBadRequest(c, "failed to parse request body")
		return
	}
	result, err := h.selfSignup.Verify(c.Request.Context(), auth.VerifySignupRequest{
		TransactionID: req.TransactionID,
		Email:         req.Email,
		Code:          req.Code,
		SourceIP:      directPeerIP(c),
	})
	if err != nil {
		writeSignupError(c, err)
		return
	}

	setSSOCookie(c, h.config, result.SessionID)
	c.JSON(http.StatusOK, gin.H{"redirect_url": result.RedirectURL})
}

func writeSignupBadRequest(c *gin.Context, description string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error":             "invalid_request",
		"error_description": description,
	})
}

// writeSignupError maps self sign-up errors to the OAuth-style JSON error shape.
func writeSignupError(c *gin.Context, err error) {
	var validationErr *auth.SignupValidationError
	switch {
	case errors.As(err, &validationErr):
		writeSignupBadRequest(c, validationErr.Message)
	case errors.Is(err, auth.ErrSignupInvalidTransaction):
		c.JSON(http.StatusNotFound, gin.H{
			"error":             "invalid_transaction",
			"error_description": "the sign-in request has expired; please start again",
		})
	case errors.Is(err, auth.ErrSignupDisabled):
		c.JSON(http.StatusForbidden, gin.H{
			"error":             "signup_disabled",
			"error_description": "self sign-up is not enabled for this application",
		})
	case errors.Is(err, auth.ErrSignupUnavailable):
		c.JSON(http.StatusForbidden, gin.H{
			"error":             "signup_unavailable",
			"error_description": "new accounts cannot be created at this time",
		})
	case errors.Is(err, auth.ErrInvalidSignupCode):
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_code",
			"error_description": "invalid or expired verification code",
		})
	case errors.Is(err, auth.ErrThrottled):
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":             "throttled",
			"error_description": "too many attempts; please try again later",
		})
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "temporarily_unavailable"})
	}
}
