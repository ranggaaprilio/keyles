package auth

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// ConsentDecisionRequest is the input for the consent decision use case.
type ConsentDecisionRequest struct {
	TransactionID        string
	InteractionCSRFToken string
	Approved             bool
	SessionID            string // from cookie
}

// ConsentDecisionResponse is returned by the consent decision use case.
type ConsentDecisionResponse struct {
	RedirectURL string // callback URL with code+state or error+state
	IsDenied    bool   // true when user denied
}

// ConsentDecision handles user consent approval/denial for the browser-facing
// OAuth flow. It atomically completes the transaction, validates CSRF + session,
// and either issues an authorization code (approved) or returns an error
// redirect (denied / validation failure).
type ConsentDecision struct {
	txnRepo      repositories.AuthorizationTransactionRepository
	sessionRepo  repositories.SessionRepository
	clientRepo   repositories.ClientRepository
	endUserRepo  repositories.EndUserRepository
	roleRepo     repositories.RoleRepository
	authCodeRepo repositories.AuthCodeRepository
	auditHelper  *OAuthAuditHelper
}

// NewConsentDecision constructs the use case.
func NewConsentDecision(
	txnRepo repositories.AuthorizationTransactionRepository,
	sessionRepo repositories.SessionRepository,
	clientRepo repositories.ClientRepository,
	endUserRepo repositories.EndUserRepository,
	roleRepo repositories.RoleRepository,
	authCodeRepo repositories.AuthCodeRepository,
	auditHelper *OAuthAuditHelper,
) *ConsentDecision {
	return &ConsentDecision{
		txnRepo:      txnRepo,
		sessionRepo:  sessionRepo,
		clientRepo:   clientRepo,
		endUserRepo:  endUserRepo,
		roleRepo:     roleRepo,
		authCodeRepo: authCodeRepo,
		auditHelper:  auditHelper,
	}
}

// ErrInvalidCSRFToken is returned when the CSRF token doesn't match the transaction.
var ErrInvalidCSRFToken = errors.New("invalid CSRF token")

// ErrTemporarilyUnavailable is returned on infrastructure (Redis) failures.
var ErrTemporarilyUnavailable = errors.New("temporarily unavailable")

// Execute processes the consent decision.
func (uc *ConsentDecision) Execute(ctx context.Context, req ConsentDecisionRequest) (*ConsentDecisionResponse, error) {
	// Validate the interaction before atomically consuming it so malformed
	// submissions cannot invalidate another user's pending consent.
	txn, err := uc.txnRepo.Get(ctx, req.TransactionID)
	if err != nil {
		return nil, ErrTemporarilyUnavailable
	}
	if txn == nil {
		return nil, ErrTransactionExpired
	}
	if txn.Stage != repositories.TransactionStagePendingConsent {
		return nil, ErrTransactionExpired
	}

	if txn.InteractionCSRFToken != req.InteractionCSRFToken {
		return nil, ErrInvalidCSRFToken
	}

	session, err := uc.sessionRepo.Get(ctx, req.SessionID)
	if err != nil {
		return nil, ErrTemporarilyUnavailable
	}
	if session == nil {
		return nil, ErrSessionUserMismatch
	}
	if session.UserID != txn.UserID || session.SessionID != txn.SessionID {
		return nil, ErrSessionUserMismatch
	}

	endUser, userErr := uc.endUserRepo.GetByID(ctx, txn.UserID)
	if userErr != nil {
		return nil, ErrTemporarilyUnavailable
	}
	if endUser == nil || endUser.Status != entities.UserStatusActive {
		return uc.consumeDenied(ctx, txn)
	}
	hasRole, roleErr := uc.roleRepo.HasAnyRole(ctx, txn.UserID, txn.ClientID)
	if roleErr != nil {
		return nil, ErrTemporarilyUnavailable
	}
	if !hasRole {
		return uc.consumeDenied(ctx, txn)
	}

	if !req.Approved {
		return uc.consumeDenied(ctx, txn)
	}

	completedTxn, err := uc.complete(ctx, req.TransactionID)
	if err != nil {
		return nil, err
	}

	code, codeErr := generateConsentAuthCode()
	if codeErr != nil {
		return nil, ErrTemporarilyUnavailable
	}

	authCode := &entities.AuthorizationCode{
		Code:                code,
		ClientID:            completedTxn.ClientID,
		TenantID:            completedTxn.TenantID,
		UserID:              completedTxn.UserID,
		RedirectURI:         completedTxn.RedirectURI,
		Scope:               completedTxn.Scope,
		CodeChallenge:       completedTxn.CodeChallenge,
		CodeChallengeMethod: completedTxn.CodeChallengeMethod,
		Nonce:               completedTxn.Nonce,
		AuthenticatedAt:     &session.AuthenticatedAt,
		ExpiresAt:           time.Now().Add(AuthCodeTTL),
		CreatedAt:           time.Now(),
	}

	if storeErr := uc.authCodeRepo.Store(ctx, authCode, AuthCodeTTL); storeErr != nil {
		return nil, ErrTemporarilyUnavailable
	}

	// Build callback: {RedirectURI}?code={code}&state={State}
	redirectURL := buildCodeRedirect(completedTxn.RedirectURI, code, completedTxn.State)

	uc.auditHelper.LogConsentApproved(ctx, completedTxn.TenantID, completedTxn.ClientID, completedTxn.UserID, completedTxn.TransactionID)

	return &ConsentDecisionResponse{
		RedirectURL: redirectURL,
		IsDenied:    false,
	}, nil
}

func (uc *ConsentDecision) consumeDenied(ctx context.Context, txn *repositories.AuthorizationTransaction) (*ConsentDecisionResponse, error) {
	completedTxn, err := uc.complete(ctx, txn.TransactionID)
	if err != nil {
		return nil, err
	}
	uc.auditHelper.LogConsentDenied(ctx, completedTxn.TenantID, completedTxn.ClientID, completedTxn.UserID, completedTxn.TransactionID)
	return &ConsentDecisionResponse{
		RedirectURL: buildErrorRedirect(completedTxn.RedirectURI, "access_denied", "The user denied the request", completedTxn.State),
		IsDenied:    true,
	}, nil
}

func (uc *ConsentDecision) complete(ctx context.Context, transactionID string) (*repositories.AuthorizationTransaction, error) {
	txn, err := uc.txnRepo.Complete(ctx, transactionID)
	if err != nil {
		if strings.Contains(err.Error(), repositories.ErrTransactionNotFound) ||
			strings.Contains(err.Error(), repositories.ErrTransactionAlreadyCompleted) ||
			strings.Contains(err.Error(), repositories.ErrTransactionWrongStage) {
			return nil, ErrTransactionExpired
		}
		return nil, ErrTemporarilyUnavailable
	}
	return txn, nil
}

// buildErrorRedirect constructs an error redirect URL per RFC 6749 §4.1.2.1.
func buildErrorRedirect(redirectURI, errorCode, description, state string) string {
	return buildRedirect(redirectURI, map[string]string{
		"error":             errorCode,
		"error_description": description,
		"state":             state,
	})
}

func buildCodeRedirect(redirectURI, code, state string) string {
	return buildRedirect(redirectURI, map[string]string{
		"code":  code,
		"state": state,
	})
}

func buildRedirect(redirectURI string, values map[string]string) string {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return redirectURI
	}
	v := url.Values{}
	for key, value := range values {
		if value != "" {
			v.Set(key, value)
		}
	}
	query := u.Query()
	for key, value := range v {
		query[key] = value
	}
	u.RawQuery = query.Encode()
	return u.String()
}

// generateConsentAuthCode generates a cryptographically secure authorization code.
func generateConsentAuthCode() (string, error) {
	return generateAuthorizationCode()
}
