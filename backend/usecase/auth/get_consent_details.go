package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// Sentinel errors for GetConsentDetails.
// ErrTransactionExpired is defined in authenticate_end_user.go (same package).
var (
	ErrSessionMissing        = errors.New("session not found or expired")
	ErrTransactionWrongStage = errors.New("transaction is not in pending_consent stage")
	ErrSessionUserMismatch   = errors.New("session user does not match transaction user")
	ErrConsentDenied         = errors.New("consent denied: user is disabled or has no role for this client")
)

// ConsentDetailsResponse is returned by GetConsentDetails on success.
type ConsentDetailsResponse struct {
	TransactionID        string
	InteractionCSRFToken string
	ClientID             string
	ClientName           string
	ClientLogoURI        string // optional; empty if client has no logo
	Scopes               []string
	UserDisplay          string
}

// GetConsentDetails retrieves the information a consent screen needs:
// client name/logo, requested scopes, and user display name.
type GetConsentDetails struct {
	txnRepo     repositories.AuthorizationTransactionRepository
	sessionRepo repositories.SessionRepository
	clientRepo  repositories.ClientRepository
	endUserRepo repositories.EndUserRepository
}

// NewGetConsentDetails constructs the use case.
func NewGetConsentDetails(
	txnRepo repositories.AuthorizationTransactionRepository,
	sessionRepo repositories.SessionRepository,
	clientRepo repositories.ClientRepository,
	endUserRepo repositories.EndUserRepository,
) *GetConsentDetails {
	return &GetConsentDetails{
		txnRepo:     txnRepo,
		sessionRepo: sessionRepo,
		clientRepo:  clientRepo,
		endUserRepo: endUserRepo,
	}
}

// Execute fetches consent details for the given transaction + session cookie.
func (uc *GetConsentDetails) Execute(ctx context.Context, transactionID, sessionID string) (*ConsentDetailsResponse, error) {
	// 1. Get session by cookie session ID
	session, err := uc.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return nil, ErrTemporarilyUnavailable
	}
	if session == nil {
		return nil, ErrSessionMissing
	}

	// 2. Get transaction by transaction ID
	txn, err := uc.txnRepo.Get(ctx, transactionID)
	if err != nil {
		return nil, ErrTemporarilyUnavailable
	}
	if txn == nil {
		return nil, ErrTransactionExpired
	}

	// 3. Validate transaction is in pending_consent stage
	if txn.Stage != repositories.TransactionStagePendingConsent {
		return nil, ErrTransactionWrongStage
	}

	// 4. Validate session.UserID matches transaction.UserID
	if session.UserID != txn.UserID || session.SessionID != txn.SessionID {
		return nil, ErrSessionUserMismatch
	}

	// 5. Load client details (name, logo)
	client, err := uc.clientRepo.GetByID(ctx, txn.ClientID)
	if err != nil {
		return nil, ErrTemporarilyUnavailable
	}
	if client == nil {
		return nil, ErrConsentDenied
	}

	// 6. Load user display name
	endUser, err := uc.endUserRepo.GetByID(ctx, txn.UserID)
	if err != nil {
		return nil, ErrTemporarilyUnavailable
	}
	if endUser == nil {
		return nil, ErrConsentDenied
	}

	// 7. Return ConsentDetailsResponse
	scopes := strings.Fields(txn.Scope)
	clientLogoURI := ""
	// Client doesn't have a LogoURI field in the current model,
	// but we keep the field for future use.

	_ = clientLogoURI // suppress unused warning until LogoURI is added to Client entity

	return &ConsentDetailsResponse{
		TransactionID:        txn.TransactionID,
		InteractionCSRFToken: txn.InteractionCSRFToken,
		ClientID:             txn.ClientID,
		ClientName:           client.ClientName,
		ClientLogoURI:        "",
		Scopes:               scopes,
		UserDisplay:          endUser.DisplayName,
	}, nil
}
