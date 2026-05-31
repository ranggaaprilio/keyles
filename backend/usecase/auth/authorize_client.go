package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// ErrAccountDisabled is returned when a disabled user tries to authorize
const ErrAccountDisabled = "Your account has been disabled. Please contact your administrator."

// AuthorizeRequest represents the OAuth authorization request parameters
type AuthorizeRequest struct {
	ClientID            string
	RedirectURI         string
	ResponseType        string
	Scope               string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
	UserID              string
	// TenantID is retained for compatibility but ignored. Tenant context is
	// derived from the registered client.
	TenantID string
}

// AuthorizeResponse represents the OAuth authorization response
type AuthorizeResponse struct {
	Code        string
	State       string
	RedirectURI string
}

// OAuthError represents an OAuth 2.0 error
type OAuthError struct {
	Code        string `json:"error"`
	Description string `json:"error_description,omitempty"`
}

func (e *OAuthError) Error() string {
	if e.Description != "" {
		return e.Code + ": " + e.Description
	}
	return e.Code
}

// OAuth error codes
const (
	ErrInvalidRequest          = "invalid_request"
	ErrUnauthorizedClient      = "unauthorized_client"
	ErrAccessDenied            = "access_denied"
	ErrUnsupportedResponseType = "unsupported_response_type"
	ErrInvalidScope            = "invalid_scope"
	ErrServerError             = "server_error"
	ErrInvalidClient           = "invalid_client"
	ErrInvalidGrant            = "invalid_grant"
)

// Authorization code TTL (5 minutes per RFC 6749)
const AuthCodeTTL = 5 * time.Minute

// AuthorizeClient handles the OAuth authorization flow
type AuthorizeClient struct {
	clientRepo   repositories.ClientRepository
	roleRepo     repositories.RoleRepository
	authCodeRepo repositories.AuthCodeRepository
	endUserRepo  repositories.EndUserRepository
}

// NewAuthorizeClient creates a new AuthorizeClient use case
func NewAuthorizeClient(
	clientRepo repositories.ClientRepository,
	roleRepo repositories.RoleRepository,
	authCodeRepo repositories.AuthCodeRepository,
	endUserRepo repositories.EndUserRepository,
) *AuthorizeClient {
	return &AuthorizeClient{
		clientRepo:   clientRepo,
		roleRepo:     roleRepo,
		authCodeRepo: authCodeRepo,
		endUserRepo:  endUserRepo,
	}
}

// Execute processes the OAuth authorization request
func (uc *AuthorizeClient) Execute(ctx context.Context, req AuthorizeRequest) (*AuthorizeResponse, error) {
	// Validate request parameters
	if err := ValidateAuthorizeRequest(req); err != nil {
		return nil, err
	}

	// Validate PKCE parameters (mandatory per FR-008)
	if err := ValidatePKCE(req.CodeChallenge, req.CodeChallengeMethod); err != nil {
		return nil, err
	}

	// Validate scope includes "openid" for OIDC
	if err := ValidateScope(req.Scope); err != nil {
		return nil, err
	}

	// Validate client_id and get client (FR-017)
	client, err := uc.clientRepo.GetByID(ctx, req.ClientID)
	if err != nil {
		return nil, &OAuthError{
			Code:        ErrInvalidClient,
			Description: "client_id is invalid or unknown",
		}
	}

	// Check client is active (FR-017)
	if !client.IsEnabled() {
		return nil, &OAuthError{
			Code:        ErrUnauthorizedClient,
			Description: "client is not active",
		}
	}

	// Validate redirect_uri matches registered URIs (FR-010, FR-018)
	if !client.IsURIAllowed(req.RedirectURI) {
		return nil, &OAuthError{
			Code:        ErrInvalidRequest,
			Description: "redirect_uri does not match any registered redirect URI",
		}
	}

	// Check user is not disabled (FR-028)
	// This prevents disabled users from initiating new OAuth flows
	endUser, err := uc.endUserRepo.GetByID(ctx, req.UserID)
	if err != nil || endUser == nil || endUser.TenantID != client.TenantID {
		return nil, &OAuthError{
			Code:        ErrAccessDenied,
			Description: "user is not authorized for this tenant",
		}
	}
	if endUser.Status == entities.UserStatusDisabled {
		return nil, &OAuthError{
			Code:        ErrAccessDenied,
			Description: ErrAccountDisabled,
		}
	}

	// Check user has role for this client (FR-006d, FR-012)
	hasRole, err := uc.roleRepo.HasAnyRole(ctx, req.UserID, req.ClientID)
	if err != nil {
		return nil, &OAuthError{
			Code:        ErrServerError,
			Description: "failed to check user roles",
		}
	}
	if !hasRole {
		return nil, &OAuthError{
			Code:        ErrAccessDenied,
			Description: "user does not have permission to access this client application",
		}
	}

	// Generate authorization code (FR-014)
	code, err := generateAuthorizationCode()
	if err != nil {
		return nil, &OAuthError{
			Code:        ErrServerError,
			Description: "failed to generate authorization code",
		}
	}

	// Create authorization code entity
	authCode := &entities.AuthorizationCode{
		Code:                code,
		ClientID:            req.ClientID,
		TenantID:            client.TenantID,
		UserID:              req.UserID,
		RedirectURI:         req.RedirectURI,
		Scope:               req.Scope,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		ExpiresAt:           time.Now().Add(AuthCodeTTL),
		CreatedAt:           time.Now(),
		UsedFlag:            false,
	}

	// Store authorization code in Redis (FR-015)
	if err := uc.authCodeRepo.Store(ctx, authCode, AuthCodeTTL); err != nil {
		return nil, &OAuthError{
			Code:        ErrServerError,
			Description: "failed to store authorization code",
		}
	}

	// Return response with code and state (FR-016)
	return &AuthorizeResponse{
		Code:        code,
		State:       req.State,
		RedirectURI: req.RedirectURI,
	}, nil
}

// ValidateAuthorizeRequest validates the authorization request parameters
func ValidateAuthorizeRequest(req AuthorizeRequest) error {
	if req.ClientID == "" {
		return &OAuthError{
			Code:        ErrInvalidRequest,
			Description: "client_id is required",
		}
	}
	if req.RedirectURI == "" {
		return &OAuthError{
			Code:        ErrInvalidRequest,
			Description: "redirect_uri is required",
		}
	}
	if req.ResponseType != "code" {
		return &OAuthError{
			Code:        ErrUnsupportedResponseType,
			Description: "only response_type=code is supported",
		}
	}
	if req.State == "" {
		return &OAuthError{
			Code:        ErrInvalidRequest,
			Description: "state is required for CSRF protection",
		}
	}
	return nil
}

// ValidatePKCE validates PKCE parameters (mandatory per FR-008)
func ValidatePKCE(codeChallenge, codeChallengeMethod string) error {
	if codeChallenge == "" {
		return &OAuthError{
			Code:        ErrInvalidRequest,
			Description: "code_challenge is required (PKCE is mandatory)",
		}
	}
	if codeChallengeMethod == "" {
		return &OAuthError{
			Code:        ErrInvalidRequest,
			Description: "code_challenge_method is required",
		}
	}
	if codeChallengeMethod != "S256" {
		return &OAuthError{
			Code:        ErrInvalidRequest,
			Description: "only code_challenge_method=S256 is supported",
		}
	}
	return nil
}

// ValidateScope validates the requested scopes
func ValidateScope(scope string) error {
	scopes := strings.Fields(scope)
	hasOpenID := false
	for _, s := range scopes {
		if s == "openid" {
			hasOpenID = true
			break
		}
	}
	if !hasOpenID {
		return &OAuthError{
			Code:        ErrInvalidScope,
			Description: "scope must include 'openid' for OIDC",
		}
	}
	return nil
}

// generateAuthorizationCode generates a cryptographically secure authorization code
func generateAuthorizationCode() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
