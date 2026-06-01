package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
)

// Sentinel errors returned by AuthenticateEndUser.
// Callers MUST NOT branch on which specific check failed — NFR-003
// requires that all credential failures surface the same generic message.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTransactionExpired = errors.New("authorization transaction expired or not found")
	ErrThrottled          = errors.New("too many login attempts; please try again later")
)

// AuthenticateEndUserRequest is the input for the end-user authentication
// use case. TransactionID comes from the /authorize redirect; Email and
// Password are collected on the login form; SourceIP is extracted from the
// request by the transport layer.
type AuthenticateEndUserRequest struct {
	TransactionID string
	Email         string
	Password      string
	SourceIP      string
}

// AuthenticateEndUserResponse is returned on successful authentication.
// RedirectURL points to the frontend consent screen so the browser can
// continue the OAuth flow without a second round-trip.
type AuthenticateEndUserResponse struct {
	RedirectURL string
	SessionID   string // opaque session token; handler sets this as cookie
}

// AuthenticateEndUser handles end-user authentication in the browser-facing
// OAuth consent flow. It validates the pending transaction, checks the
// throttler, verifies credentials, creates a client-agnostic session, and
// binds the transaction to the authenticated user.
type AuthenticateEndUser struct {
	txnRepo     repositories.AuthorizationTransactionRepository
	sessionRepo repositories.SessionRepository
	clientRepo  repositories.ClientRepository
	endUserRepo repositories.EndUserRepository
	roleRepo    repositories.RoleRepository
	throttler   services.LoginThrottler
	passwordSvc services.PasswordService
	auditHelper *OAuthAuditHelper
	sessionTTL  time.Duration
	frontendURL string
}

// NewAuthenticateEndUser constructs the use case with all required
// dependencies. sessionTTL defaults to 8 hours if zero. frontendURL
// is the base URL of the frontend (e.g. https://sso.example.com).
func NewAuthenticateEndUser(
	txnRepo repositories.AuthorizationTransactionRepository,
	sessionRepo repositories.SessionRepository,
	clientRepo repositories.ClientRepository,
	endUserRepo repositories.EndUserRepository,
	roleRepo repositories.RoleRepository,
	throttler services.LoginThrottler,
	passwordSvc services.PasswordService,
	auditHelper *OAuthAuditHelper,
	sessionTTL time.Duration,
	frontendURL string,
) *AuthenticateEndUser {
	if sessionTTL == 0 {
		sessionTTL = 8 * time.Hour
	}
	return &AuthenticateEndUser{
		txnRepo:     txnRepo,
		sessionRepo: sessionRepo,
		clientRepo:  clientRepo,
		endUserRepo: endUserRepo,
		roleRepo:    roleRepo,
		throttler:   throttler,
		passwordSvc: passwordSvc,
		auditHelper: auditHelper,
		sessionTTL:  sessionTTL,
		frontendURL: strings.TrimRight(frontendURL, "/"),
	}
}

// Execute authenticates an end-user and, on success, transitions the
// authorization transaction to pending_consent. On any failure a generic
// ErrInvalidCredentials is returned so that callers cannot enumerate
// which check failed (NFR-003).
func (uc *AuthenticateEndUser) Execute(ctx context.Context, req AuthenticateEndUserRequest) (*AuthenticateEndUserResponse, error) {
	// ── 1. Load and validate the authorization transaction ────────────────
	txn, err := uc.txnRepo.Get(ctx, req.TransactionID)
	if err != nil {
		return nil, ErrTemporarilyUnavailable
	}
	if txn == nil || txn.Stage != repositories.TransactionStagePendingLogin {
		return nil, ErrTransactionExpired
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))
	tenantID := txn.TenantID
	clientID := txn.ClientID

	// ── 2. Throttle check BEFORE credential verification ──────────────────
	throttled, err := uc.throttler.IsThrottled(ctx, req.SourceIP, tenantID, normalizedEmail)
	if err != nil {
		return nil, ErrTemporarilyUnavailable
	}
	if throttled {
		uc.auditHelper.LogLoginThrottled(ctx, req.SourceIP)
		return nil, ErrThrottled
	}

	// Helper: record a failed attempt and return generic credentials error.
	// Never reveals which check failed.
	credentialFailure := func() (*AuthenticateEndUserResponse, error) {
		if err := uc.throttler.RecordFailure(ctx, req.SourceIP, tenantID, normalizedEmail); err != nil {
			return nil, ErrTemporarilyUnavailable
		}
		uc.auditHelper.LogLoginFailed(ctx, tenantID, clientID, req.SourceIP)
		return nil, ErrInvalidCredentials
	}

	// ── 3. Load the client from the transaction ───────────────────────────
	client, err := uc.clientRepo.GetByID(ctx, clientID)
	if err != nil || client == nil || !client.IsEnabled() {
		return credentialFailure()
	}

	// ── 4. Fetch tenant-scoped end user by email ──────────────────────────
	user, err := uc.endUserRepo.GetByEmail(ctx, tenantID, normalizedEmail)
	if err != nil || user == nil {
		return credentialFailure()
	}

	// ── 5. Check active status ────────────────────────────────────────────
	if user.Status != entities.UserStatusActive {
		return credentialFailure()
	}

	// ── 6. Check client role ──────────────────────────────────────────────
	hasRole, err := uc.roleRepo.HasAnyRole(ctx, user.ID, clientID)
	if err != nil || !hasRole {
		return credentialFailure()
	}

	// ── 7. Verify bcrypt password ──────────────────────────────────────────
	if err := uc.passwordSvc.Verify(req.Password, user.PasswordHash); err != nil {
		return credentialFailure()
	}

	// ── 8. Create client-agnostic Session with crypto/rand SessionID ──────
	sessionID, err := GenerateSessionID()
	if err != nil {
		return credentialFailure()
	}

	now := time.Now()
	session := &repositories.Session{
		SessionID:       sessionID,
		UserID:          user.ID,
		TenantID:        tenantID,
		AuthenticatedAt: now,
		CreatedAt:       now,
		ExpiresAt:       now.Add(uc.sessionTTL),
		Metadata:        map[string]interface{}{},
	}

	if err := uc.sessionRepo.Create(ctx, session, uc.sessionTTL); err != nil {
		return nil, ErrTemporarilyUnavailable
	}

	// ── 9. Clear email throttle bucket on success ─────────────────────────
	if err := uc.throttler.ClearEmailBucket(ctx, tenantID, normalizedEmail); err != nil {
		_ = uc.sessionRepo.Delete(ctx, sessionID)
		return nil, ErrTemporarilyUnavailable
	}

	// ── 10. Bind transaction (UpdateStage → pending_consent) ──────────────
	if err := uc.txnRepo.UpdateStage(ctx, req.TransactionID, repositories.TransactionStagePendingConsent, user.ID, sessionID); err != nil {
		_ = uc.sessionRepo.Delete(ctx, sessionID)
		return nil, ErrTemporarilyUnavailable
	}

	// ── 11. Record last login ─────────────────────────────────────────────
	_ = uc.endUserRepo.UpdateLastLogin(ctx, user.ID, now)

	// ── 12. Emit success audit event ─────────────────────────────────────
	uc.auditHelper.LogLoginSucceeded(ctx, tenantID, clientID, user.ID, req.SourceIP)

	// ── Build response ────────────────────────────────────────────────────
	redirectURL := fmt.Sprintf("%s/oauth2/consent?transaction_id=%s", uc.frontendURL, req.TransactionID)
	return &AuthenticateEndUserResponse{
		RedirectURL: redirectURL,
		SessionID:   sessionID,
	}, nil
}

// GenerateSessionID creates a cryptographically random session identifier
// using at least 32 bytes of entropy from crypto/rand, encoded as
// base64url (no padding). Exported for testing.
func GenerateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
