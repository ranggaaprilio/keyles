package repositories

import (
	"context"
	"time"
)

// AuthorizationTransactionStage represents the current stage of an OAuth
// authorization transaction.
type AuthorizationTransactionStage string

const (
	TransactionStagePendingLogin   AuthorizationTransactionStage = "pending_login"
	TransactionStagePendingConsent AuthorizationTransactionStage = "pending_consent"
	TransactionStageCompleted      AuthorizationTransactionStage = "completed"
)

// AuthorizationTransaction holds server-side OAuth request state during
// the browser-facing login and consent flow.
type AuthorizationTransaction struct {
	TransactionID        string
	ClientID             string
	TenantID             string
	RedirectURI          string
	ResponseType         string
	Scope                string
	State                string
	CodeChallenge        string
	CodeChallengeMethod  string
	Nonce                string
	Prompt               []string
	MaxAgeSeconds        *int
	UserID               string
	SessionID            string
	InteractionCSRFToken string
	Stage                AuthorizationTransactionStage
	CreatedAt            time.Time
	ExpiresAt            time.Time
	CompletedAt          *time.Time
}

// AuthorizationTransactionRepository manages authorization transactions in Redis.
// Transactions are short-lived (default 10-minute TTL) and must be consumed atomically
// to prevent replay attacks.
type AuthorizationTransactionRepository interface {
	// Create stores a new authorization transaction with the given TTL.
	Create(ctx context.Context, txn *AuthorizationTransaction, ttl time.Duration) error

	// Get retrieves a transaction by ID. Returns nil, nil if not found.
	Get(ctx context.Context, transactionID string) (*AuthorizationTransaction, error)

	// UpdateStage updates the transaction stage and may bind a user or session.
	// Returns ErrTransactionNotFound if the transaction does not exist or has expired.
	UpdateStage(ctx context.Context, transactionID string, stage AuthorizationTransactionStage, userID string, sessionID string) error

	// Complete atomically marks a pending_consent transaction as completed and
	// returns the stored transaction data. Returns ErrTransactionNotFound or
	// ErrTransactionAlreadyCompleted if the transaction is missing, expired, or
	// already consumed. This is the one-time-use gate that prevents replay.
	Complete(ctx context.Context, transactionID string) (*AuthorizationTransaction, error)
}

// Transaction errors
var (
	ErrTransactionNotFound         = "authorization transaction not found or expired"
	ErrTransactionAlreadyCompleted = "authorization transaction already used"
	ErrTransactionWrongStage       = "authorization transaction is not awaiting consent"
)
