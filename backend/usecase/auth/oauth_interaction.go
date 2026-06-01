package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// TransactionTTL is the default TTL for authorization transactions (10 minutes).
const TransactionTTL = 10 * time.Minute

// Interaction errors — these are used internally and when building redirect URLs.
const (
	ErrInvalidRedirectURI         = "invalid_redirect_uri"
	ErrLoginRequired              = "login_required"
	ErrConsentRequired            = "consent_required"
	ErrTemporarilyUnavailableCode = "temporarily_unavailable"
)

// InitializeAuthRequest holds the parameters for the initial /oauth2/auth request.
type InitializeAuthRequest struct {
	ClientID            string
	RedirectURI         string
	ResponseType        string
	Scope               string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
	Nonce               string
	Prompt              string
	MaxAge              *int
	MaxAgeInvalid       bool
	SessionCookie       string // empty if no session cookie
	SourceIP            string
}

// InitializeAuthResponse holds the result of the interaction initialization.
type InitializeAuthResponse struct {
	RedirectURL  string
	IsLocalError bool // true when redirecting to frontend error page
}

// OAuthInteraction handles the initial OAuth /oauth2/auth request:
// validating parameters, creating an authorization transaction, and
// determining the redirect target.
type OAuthInteraction struct {
	clientRepo     repositories.ClientRepository
	roleRepo       repositories.RoleRepository
	txnRepo        repositories.AuthorizationTransactionRepository
	sessionRepo    repositories.SessionRepository
	endUserRepo    repositories.EndUserRepository
	auditHelper    *OAuthAuditHelper
	frontendURL    string
	transactionTTL time.Duration
}

// NewOAuthInteraction creates a new OAuthInteraction use case.
func NewOAuthInteraction(
	clientRepo repositories.ClientRepository,
	roleRepo repositories.RoleRepository,
	txnRepo repositories.AuthorizationTransactionRepository,
	sessionRepo repositories.SessionRepository,
	endUserRepo repositories.EndUserRepository,
	auditHelper *OAuthAuditHelper,
	frontendURL string,
	transactionTTL ...time.Duration,
) *OAuthInteraction {
	ttl := TransactionTTL
	if len(transactionTTL) > 0 && transactionTTL[0] > 0 {
		ttl = transactionTTL[0]
	}
	return &OAuthInteraction{
		clientRepo:     clientRepo,
		roleRepo:       roleRepo,
		txnRepo:        txnRepo,
		sessionRepo:    sessionRepo,
		endUserRepo:    endUserRepo,
		auditHelper:    auditHelper,
		frontendURL:    frontendURL,
		transactionTTL: ttl,
	}
}

// InitializeAuth validates the OAuth request, creates an authorization transaction,
// and returns a redirect URL for the frontend.
func (uc *OAuthInteraction) InitializeAuth(ctx context.Context, req InitializeAuthRequest) (*InitializeAuthResponse, error) {
	// 1. Validate required fields.
	if err := uc.validateRequest(req); err != nil {
		return uc.localErrorRedirect(err.Code, err.Description), nil
	}

	// 2. Look up client by client_id.
	client, err := uc.clientRepo.GetByID(ctx, req.ClientID)
	if err != nil || client == nil {
		return uc.localErrorRedirect(ErrInvalidClient, "client_id is invalid or unknown"), nil
	}
	if !client.IsEnabled() {
		return uc.localErrorRedirect(ErrInvalidClient, "client is not active"), nil
	}

	// 3. Validate redirect_uri against registered URIs.
	if !client.IsURIAllowed(req.RedirectURI) {
		uc.auditHelper.LogInvalidCallback(ctx, req.ClientID, req.SourceIP)
		return uc.localErrorRedirect(ErrInvalidRedirectURI, "redirect_uri does not match any registered redirect URI"), nil
	}
	if req.MaxAgeInvalid || (req.MaxAge != nil && *req.MaxAge < 0) {
		return uc.callbackErrorRedirect(req, ErrInvalidRequest, "max_age must be a non-negative integer"), nil
	}

	// 4. Parse prompt values.
	prompts := ParsePrompt(req.Prompt)
	hasNone := ContainsString(prompts, "none")
	hasLogin := ContainsString(prompts, "login")
	hasConsent := ContainsString(prompts, "consent")
	hasSelectAccount := ContainsString(prompts, "select_account")
	for _, prompt := range prompts {
		if prompt != "none" && prompt != "login" && prompt != "consent" && prompt != "select_account" {
			return uc.callbackErrorRedirect(req, ErrInvalidRequest, "unsupported prompt value"), nil
		}
	}

	// prompt=none combined with other values → error to validated callback.
	if hasNone && (len(prompts) > 1) {
		return uc.callbackErrorRedirect(req, ErrInvalidRequest, "prompt=none must be the only prompt value"), nil
	}

	// select_account is not supported.
	if hasSelectAccount {
		return uc.callbackErrorRedirect(req, ErrUnsupportedResponseType, "prompt=select_account is not supported"), nil
	}

	// 5. Resolve tenant from client.
	tenantID := client.TenantID

	// 6. Create authorization transaction with crypto/rand-backed identifiers.
	txnID, err := GenerateSecureToken()
	if err != nil {
		return uc.localErrorRedirect(ErrServerError, "failed to generate transaction ID"), nil
	}
	csrfToken, err := GenerateSecureToken()
	if err != nil {
		return uc.localErrorRedirect(ErrServerError, "failed to generate CSRF token"), nil
	}

	now := time.Now()
	txn := &repositories.AuthorizationTransaction{
		TransactionID:        txnID,
		ClientID:             req.ClientID,
		TenantID:             tenantID,
		RedirectURI:          req.RedirectURI,
		ResponseType:         req.ResponseType,
		Scope:                req.Scope,
		State:                req.State,
		CodeChallenge:        req.CodeChallenge,
		CodeChallengeMethod:  req.CodeChallengeMethod,
		Nonce:                req.Nonce,
		Prompt:               prompts,
		MaxAgeSeconds:        req.MaxAge,
		InteractionCSRFToken: csrfToken,
		Stage:                repositories.TransactionStagePendingLogin,
		CreatedAt:            now,
		ExpiresAt:            now.Add(uc.transactionTTL),
	}

	if err := uc.txnRepo.Create(ctx, txn, uc.transactionTTL); err != nil {
		// Redis fail-closed: any persistence error → local error redirect.
		return uc.localErrorRedirect(ErrTemporarilyUnavailableCode, "failed to create authorization transaction"), nil
	}

	// 7. Try to reuse an existing session.
	if req.SessionCookie != "" {
		session, sessErr := uc.sessionRepo.Get(ctx, req.SessionCookie)
		if sessErr != nil {
			return uc.localErrorRedirect(ErrTemporarilyUnavailableCode, "failed to load browser session"), nil
		}
		if session != nil {
			redirect, handled := uc.trySessionRedirect(ctx, req, txnID, session, client, tenantID, hasNone, hasLogin, hasConsent)
			if handled {
				return redirect, nil
			}
		}
	}
	// 8. prompt=none without any session → callback error per OIDC spec.
	if hasNone {
		return uc.callbackErrorRedirect(req, ErrLoginRequired, "login_required: no existing session"), nil
	}
	// 9. No eligible session → redirect to frontend login.
	return uc.loginRedirect(txnID), nil
}

// trySessionRedirect attempts to use an existing session to short-circuit
// the login or consent step. Returns (response, true) if a redirect was
// determined, or (nil, false) if the session should not be reused.
func (uc *OAuthInteraction) trySessionRedirect(
	ctx context.Context,
	req InitializeAuthRequest,
	txnID string,
	session *repositories.Session,
	client *entities.Client,
	tenantID string,
	hasNone bool,
	hasLogin bool,
	hasConsent bool,
) (*InitializeAuthResponse, bool) {
	// Session must belong to the same tenant.
	if session.TenantID != tenantID {
		return nil, false
	}

	// Check user is still eligible.
	user, err := uc.endUserRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return uc.localErrorRedirect(ErrTemporarilyUnavailableCode, "failed to validate browser session"), true
	}
	if user == nil {
		return nil, false
	}
	if user.Status != entities.UserStatusActive {
		return nil, false
	}

	// Verify user still has a role for this client.
	hasRole, err := uc.roleRepo.HasAnyRole(ctx, session.UserID, req.ClientID)
	if err != nil {
		return uc.localErrorRedirect(ErrTemporarilyUnavailableCode, "failed to validate browser session"), true
	}
	if !hasRole {
		return nil, false
	}

	// prompt=none: user must re-authenticate to give consent (no persisted grants
	// in this feature). Return error to validated callback per OIDC spec.
	if hasNone {
		return uc.callbackErrorRedirect(req, ErrConsentRequired, "consent_required: no persisted consent for this client"), true
	}

	// prompt=login or max_age exceeded: force re-authentication, do NOT reuse session.
	if hasLogin {
		return uc.loginRedirect(txnID), true
	}
	if req.MaxAge != nil {
		maxAgeDuration := time.Duration(*req.MaxAge) * time.Second
		if time.Since(session.AuthenticatedAt) > maxAgeDuration {
			return uc.loginRedirect(txnID), true
		}
	}

	// Session is eligible. Bind transaction to session and redirect to consent.
	// Update stage to pending_consent with the session user.
	if err := uc.txnRepo.UpdateStage(ctx, txnID, repositories.TransactionStagePendingConsent, session.UserID, session.SessionID); err != nil {
		return uc.localErrorRedirect(ErrTemporarilyUnavailableCode, "failed to continue authorization request"), true
	}

	return uc.consentRedirect(txnID), true
}

// validateRequest checks all required OAuth/OIDC parameters.
func (uc *OAuthInteraction) validateRequest(req InitializeAuthRequest) *OAuthError {
	if req.ClientID == "" {
		return &OAuthError{Code: ErrInvalidRequest, Description: "client_id is required"}
	}
	if req.RedirectURI == "" {
		return &OAuthError{Code: ErrInvalidRequest, Description: "redirect_uri is required"}
	}
	if req.ResponseType != "code" {
		return &OAuthError{Code: ErrUnsupportedResponseType, Description: "only response_type=code is supported"}
	}
	if req.State == "" {
		return &OAuthError{Code: ErrInvalidRequest, Description: "state is required for CSRF protection"}
	}
	if req.CodeChallenge == "" {
		return &OAuthError{Code: ErrInvalidRequest, Description: "code_challenge is required (PKCE is mandatory)"}
	}
	if req.CodeChallengeMethod != "S256" {
		return &OAuthError{Code: ErrInvalidRequest, Description: "only code_challenge_method=S256 is supported"}
	}
	if err := ValidateScope(req.Scope); err != nil {
		return err.(*OAuthError)
	}
	return nil
}

// localErrorRedirect constructs a redirect to the frontend error page.
func (uc *OAuthInteraction) localErrorRedirect(code, description string) *InitializeAuthResponse {
	return &InitializeAuthResponse{
		RedirectURL:  fmt.Sprintf("%s/oauth2/error?error=%s&error_description=%s", uc.frontendURL, url.QueryEscape(code), url.QueryEscape(description)),
		IsLocalError: true,
	}
}

// loginRedirect constructs a redirect to the frontend login page.
func (uc *OAuthInteraction) loginRedirect(txnID string) *InitializeAuthResponse {
	return &InitializeAuthResponse{
		RedirectURL:  fmt.Sprintf("%s/oauth2/login?transaction_id=%s", uc.frontendURL, url.QueryEscape(txnID)),
		IsLocalError: false,
	}
}

// consentRedirect constructs a redirect to the frontend consent page.
func (uc *OAuthInteraction) consentRedirect(txnID string) *InitializeAuthResponse {
	return &InitializeAuthResponse{
		RedirectURL:  fmt.Sprintf("%s/oauth2/consent?transaction_id=%s", uc.frontendURL, url.QueryEscape(txnID)),
		IsLocalError: false,
	}
}

// callbackErrorRedirect constructs a redirect to the validated callback URI
// with error parameters, preserving the state parameter.
func (uc *OAuthInteraction) callbackErrorRedirect(req InitializeAuthRequest, code, description string) *InitializeAuthResponse {
	u, _ := url.Parse(req.RedirectURI)
	q := u.Query()
	q.Set("error", code)
	q.Set("error_description", description)
	if req.State != "" {
		q.Set("state", req.State)
	}
	u.RawQuery = q.Encode()
	return &InitializeAuthResponse{
		RedirectURL:  u.String(),
		IsLocalError: false,
	}
}

// GenerateSecureToken creates a cryptographically random opaque identifier.
func GenerateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// ParsePrompt splits the space-delimited prompt parameter per OIDC spec.
func ParsePrompt(prompt string) []string {
	if prompt == "" {
		return nil
	}
	parts := strings.Fields(prompt)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// ContainsString checks if a string slice contains a given value.
func ContainsString(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
