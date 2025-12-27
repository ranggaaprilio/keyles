/**
 * ConsentDecision use case
 * Handles user consent approval/denial for OAuth authorization
 */

package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// ConsentRequest represents a user's consent decision
type ConsentRequest struct {
	// SessionID is the authorization session identifier
	SessionID string
	// UserID is the authenticated user's ID
	UserID string
	// TenantID is the tenant context
	TenantID string
	// ClientID is the OAuth client requesting access
	ClientID string
	// RedirectURI is the callback URL
	RedirectURI string
	// Scope is the requested scope
	Scope string
	// State is the CSRF protection state
	State string
	// CodeChallenge is the PKCE code challenge
	CodeChallenge string
	// CodeChallengeMethod is the PKCE method (S256)
	CodeChallengeMethod string
	// Approved indicates if the user approved the consent
	Approved bool
}

// ConsentResponse represents the result of consent decision
type ConsentResponse struct {
	// RedirectURL is the URL to redirect the user to
	RedirectURL string
	// Code is the authorization code (if approved)
	Code string
	// State is the CSRF state parameter
	State string
	// Error is set if consent was denied or an error occurred
	Error string
	// ErrorDescription provides more details about the error
	ErrorDescription string
}

// ConsentDecisionUseCase handles user consent decisions
type ConsentDecisionUseCase struct {
	clientRepo   repositories.ClientRepository
	authCodeRepo repositories.AuthCodeRepository
}

// NewConsentDecision creates a new ConsentDecision use case
func NewConsentDecision(
	clientRepo repositories.ClientRepository,
	authCodeRepo repositories.AuthCodeRepository,
) *ConsentDecisionUseCase {
	return &ConsentDecisionUseCase{
		clientRepo:   clientRepo,
		authCodeRepo: authCodeRepo,
	}
}

// Execute processes the consent decision
func (uc *ConsentDecisionUseCase) Execute(ctx context.Context, req *ConsentRequest) (*ConsentResponse, error) {
	// Validate request
	if err := uc.validateRequest(req); err != nil {
		return nil, err
	}

	// Validate client exists and redirect URI matches
	client, err := uc.clientRepo.GetByID(ctx, req.ClientID)
	if err != nil || client == nil {
		return &ConsentResponse{
			RedirectURL:      req.RedirectURI,
			State:            req.State,
			Error:            "invalid_client",
			ErrorDescription: "Client not found",
		}, nil
	}

	// Verify redirect URI
	validRedirect := false
	for _, uri := range client.AllowedRedirectURIs {
		if uri == req.RedirectURI {
			validRedirect = true
			break
		}
	}
	if !validRedirect {
		return nil, errors.New("invalid redirect_uri")
	}

	// If consent denied, return error redirect
	if !req.Approved {
		return &ConsentResponse{
			RedirectURL:      req.RedirectURI,
			State:            req.State,
			Error:            "access_denied",
			ErrorDescription: "The user denied the authorization request",
		}, nil
	}

	// Generate authorization code
	code, err := generateSecureCode(32)
	if err != nil {
		return nil, errors.New("failed to generate authorization code")
	}

	// Create authorization code entity
	authCode := &entities.AuthorizationCode{
		Code:                code,
		ClientID:            req.ClientID,
		UserID:              req.UserID,
		TenantID:            req.TenantID,
		RedirectURI:         req.RedirectURI,
		Scope:               req.Scope,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		ExpiresAt:           time.Now().Add(5 * time.Minute),
		CreatedAt:           time.Now(),
	}

	// Store authorization code with 5-minute TTL
	if err := uc.authCodeRepo.Store(ctx, authCode, 5*time.Minute); err != nil {
		return nil, errors.New("failed to store authorization code")
	}

	return &ConsentResponse{
		RedirectURL: req.RedirectURI,
		Code:        code,
		State:       req.State,
	}, nil
}

// validateRequest validates the consent request
func (uc *ConsentDecisionUseCase) validateRequest(req *ConsentRequest) error {
	if req.UserID == "" {
		return errors.New("user_id is required")
	}
	if req.ClientID == "" {
		return errors.New("client_id is required")
	}
	if req.RedirectURI == "" {
		return errors.New("redirect_uri is required")
	}
	if req.CodeChallenge == "" {
		return errors.New("code_challenge is required for PKCE")
	}
	if req.CodeChallengeMethod != "S256" {
		return errors.New("only S256 code_challenge_method is supported")
	}
	return nil
}

// generateSecureCode generates a cryptographically secure random code
func generateSecureCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
