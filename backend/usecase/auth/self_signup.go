package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
)

// Self sign-up tuning. The OTP expiry matches the copy in the email template.
const (
	SignupOTPTTL                 = 10 * time.Minute
	SignupOTPMaxAttempts         = 5
	SignupResendCooldown         = 60 * time.Second
	SignupMaxSendsPerHour        = 5
	SignupTransactionTTL         = 15 * time.Minute
	SignupTransactionMaxLifetime = 30 * time.Minute
)

// Sentinel errors returned by SelfSignup. Request/Resend never reveal whether
// an email is already registered: they succeed identically in every case.
var (
	ErrSignupInvalidTransaction = errors.New("authorization transaction expired or not found")
	ErrSignupDisabled           = errors.New("self sign-up is not enabled for this application")
	ErrSignupUnavailable        = errors.New("self sign-up is currently unavailable")
	ErrInvalidSignupCode        = errors.New("invalid or expired verification code")
)

// SignupValidationError carries a user-facing input validation message.
type SignupValidationError struct{ Message string }

func (e *SignupValidationError) Error() string { return e.Message }

// SignupContext describes the sign-up page for an authorization transaction.
type SignupContext struct {
	ClientName    string
	SignupEnabled bool
}

// SignupRequest is the input for starting a self sign-up.
type SignupRequest struct {
	TransactionID string
	Email         string
	Password      string
	DisplayName   string
	SourceIP      string
}

// VerifySignupRequest is the input for completing a self sign-up.
type VerifySignupRequest struct {
	TransactionID string
	Email         string
	Code          string
	SourceIP      string
}

// SelfSignup lets end-users register themselves into the tenant that owns the
// OAuth client of an in-flight authorization transaction. The user record is
// created only after the emailed code is verified; until then the sign-up
// lives in the PendingSignupStore.
type SelfSignup struct {
	txnRepo      repositories.AuthorizationTransactionRepository
	clientRepo   repositories.ClientRepository
	endUserRepo  repositories.EndUserRepository
	userRepo     repositories.UserRepository // global email existence check
	roleRepo     repositories.RoleRepository
	signupStore  repositories.PendingSignupStore
	passwordSvc  services.PasswordService
	otpSvc       services.OTPService
	emailService services.EmailService
	countCache   services.UserCountCache
	auditHelper  *OAuthAuditHelper
	completer    *LoginCompleter
}

// NewSelfSignup constructs the self sign-up use case.
func NewSelfSignup(
	txnRepo repositories.AuthorizationTransactionRepository,
	clientRepo repositories.ClientRepository,
	endUserRepo repositories.EndUserRepository,
	userRepo repositories.UserRepository,
	roleRepo repositories.RoleRepository,
	signupStore repositories.PendingSignupStore,
	passwordSvc services.PasswordService,
	otpSvc services.OTPService,
	emailService services.EmailService,
	countCache services.UserCountCache,
	auditHelper *OAuthAuditHelper,
	completer *LoginCompleter,
) *SelfSignup {
	return &SelfSignup{
		txnRepo:      txnRepo,
		clientRepo:   clientRepo,
		endUserRepo:  endUserRepo,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		signupStore:  signupStore,
		passwordSvc:  passwordSvc,
		otpSvc:       otpSvc,
		emailService: emailService,
		countCache:   countCache,
		auditHelper:  auditHelper,
		completer:    completer,
	}
}

// loadTransaction returns the pending_login transaction and its enabled client.
func (uc *SelfSignup) loadTransaction(ctx context.Context, transactionID string) (*repositories.AuthorizationTransaction, *entities.Client, error) {
	txn, err := uc.txnRepo.Get(ctx, transactionID)
	if err != nil {
		return nil, nil, ErrTemporarilyUnavailable
	}
	if txn == nil || txn.Stage != repositories.TransactionStagePendingLogin {
		return nil, nil, ErrSignupInvalidTransaction
	}
	client, err := uc.clientRepo.GetByID(ctx, txn.ClientID)
	if err != nil {
		return nil, nil, ErrTemporarilyUnavailable
	}
	if client == nil || !client.IsEnabled() || client.TenantID != txn.TenantID {
		return nil, nil, ErrSignupInvalidTransaction
	}
	return txn, client, nil
}

// signupEnabled reports whether the client accepts self sign-ups.
func signupEnabled(client *entities.Client) bool {
	return client.AllowSelfSignup && strings.TrimSpace(client.SelfSignupRole) != ""
}

// GetContext returns the client name and whether self sign-up is enabled.
func (uc *SelfSignup) GetContext(ctx context.Context, transactionID string) (*SignupContext, error) {
	_, client, err := uc.loadTransaction(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	return &SignupContext{ClientName: client.ClientName, SignupEnabled: signupEnabled(client)}, nil
}

// Request validates the sign-up, stores it as pending, and emails a code.
// It returns nil for new and already-registered emails alike.
func (uc *SelfSignup) Request(ctx context.Context, req SignupRequest) error {
	email := entities.NormalizeEmail(req.Email)
	displayName := strings.TrimSpace(req.DisplayName)
	if err := entities.ValidateEmail(email); err != nil {
		return &SignupValidationError{Message: err.Error()}
	}
	if err := entities.ValidateFullName(displayName); err != nil {
		return &SignupValidationError{Message: err.Error()}
	}
	if err := entities.ValidatePassword(req.Password); err != nil {
		return &SignupValidationError{Message: err.Error()}
	}

	txn, client, err := uc.loadTransaction(ctx, req.TransactionID)
	if err != nil {
		return err
	}
	if !signupEnabled(client) {
		return ErrSignupDisabled
	}

	// Hash before any email lookup so response timing does not depend on it.
	passwordHash, err := uc.passwordSvc.Hash(req.Password)
	if err != nil {
		return ErrTemporarilyUnavailable
	}

	if err := uc.checkQuota(ctx, txn.TenantID); err != nil {
		return err
	}

	exists, err := uc.userRepo.EmailExists(ctx, email)
	if err != nil {
		return ErrTemporarilyUnavailable
	}
	if !exists {
		allowed, err := uc.signupStore.AllowSend(ctx, txn.TenantID, email, SignupResendCooldown, SignupMaxSendsPerHour)
		if err != nil {
			return ErrTemporarilyUnavailable
		}
		if allowed {
			pending := &repositories.PendingSignup{
				TenantID:      txn.TenantID,
				TransactionID: req.TransactionID,
				Email:         email,
				DisplayName:   displayName,
				PasswordHash:  passwordHash,
			}
			if err := uc.sendCode(ctx, pending, client.ClientName); err != nil {
				return err
			}
		}
	}

	_ = uc.txnRepo.ExtendTTL(ctx, req.TransactionID, SignupTransactionTTL, SignupTransactionMaxLifetime)
	uc.auditHelper.LogSignupRequested(ctx, txn.TenantID, txn.ClientID, req.SourceIP)
	return nil
}

// Resend emails a fresh code for an existing pending sign-up. Like Request,
// it succeeds identically whether or not a pending sign-up exists.
func (uc *SelfSignup) Resend(ctx context.Context, transactionID, rawEmail string) error {
	email := entities.NormalizeEmail(rawEmail)
	txn, client, err := uc.loadTransaction(ctx, transactionID)
	if err != nil {
		return err
	}
	if !signupEnabled(client) {
		return ErrSignupDisabled
	}

	pending, err := uc.signupStore.Get(ctx, txn.TenantID, email)
	if err != nil {
		return ErrTemporarilyUnavailable
	}
	// Only the browser that started the sign-up may request new codes.
	if pending != nil && pending.TransactionID == transactionID {
		allowed, err := uc.signupStore.AllowSend(ctx, txn.TenantID, email, SignupResendCooldown, SignupMaxSendsPerHour)
		if err != nil {
			return ErrTemporarilyUnavailable
		}
		if allowed {
			if err := uc.sendCode(ctx, pending, client.ClientName); err != nil {
				return err
			}
		}
	}

	_ = uc.txnRepo.ExtendTTL(ctx, transactionID, SignupTransactionTTL, SignupTransactionMaxLifetime)
	return nil
}

// sendCode generates a new code, stores its transaction-bound hash with the
// pending sign-up, and emails the plaintext code. Email delivery failures are
// logged but not returned: surfacing them would reveal that the address is
// not yet registered.
func (uc *SelfSignup) sendCode(ctx context.Context, pending *repositories.PendingSignup, clientName string) error {
	code, err := uc.otpSvc.Generate()
	if err != nil {
		return ErrTemporarilyUnavailable
	}
	pending.CodeHash = hashSignupCode(pending.TransactionID, code)
	if err := uc.signupStore.Save(ctx, pending, SignupOTPTTL); err != nil {
		return ErrTemporarilyUnavailable
	}
	if err := uc.emailService.SendSignupOTPEmail(ctx, pending.Email, pending.DisplayName, code, clientName); err != nil {
		slog.ErrorContext(ctx, "signup_otp_email_failed", slog.String("tenant_id", pending.TenantID), slog.String("error", err.Error()))
		_ = uc.signupStore.Delete(ctx, pending.TenantID, pending.Email)
	}
	return nil
}

// Verify checks the emailed code, creates the active user with the client's
// default role, and completes login so the browser can continue to consent.
func (uc *SelfSignup) Verify(ctx context.Context, req VerifySignupRequest) (*AuthenticateEndUserResponse, error) {
	email := entities.NormalizeEmail(req.Email)
	if !uc.otpSvc.Validate(req.Code) {
		return nil, ErrInvalidSignupCode
	}

	txn, client, err := uc.loadTransaction(ctx, req.TransactionID)
	if err != nil {
		return nil, err
	}
	if !signupEnabled(client) {
		return nil, ErrSignupDisabled
	}

	// Attempts are capped per code by the store (and per IP by the route
	// limiter); the login throttler is not used so that failed codes cannot
	// lock the address owner out of password login.
	pending, err := uc.signupStore.Verify(ctx, txn.TenantID, email, hashSignupCode(req.TransactionID, req.Code), SignupOTPMaxAttempts)
	if err != nil {
		return nil, ErrTemporarilyUnavailable
	}
	if pending == nil {
		return nil, ErrInvalidSignupCode
	}

	if err := uc.checkQuota(ctx, txn.TenantID); err != nil {
		return nil, err
	}

	user := entities.NewUser(txn.TenantID, pending.Email, pending.DisplayName)
	user.ID = uuid.New().String()
	user.PasswordHash = pending.PasswordHash
	user.Status = entities.UserStatusActive
	if err := uc.endUserRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repositories.ErrDuplicateEmail) {
			// Registered elsewhere since the code was sent.
			return nil, ErrInvalidSignupCode
		}
		return nil, ErrTemporarilyUnavailable
	}
	_ = uc.countCache.Invalidate(ctx, txn.TenantID)

	assignment := &entities.UserRoleAssignment{
		UserID:    user.ID,
		ClientID:  client.ClientID,
		TenantID:  txn.TenantID,
		Role:      strings.TrimSpace(client.SelfSignupRole),
		IsActive:  true,
		GrantedAt: time.Now(),
		GrantedBy: user.ID,
	}
	if err := uc.roleRepo.Assign(ctx, assignment); err != nil {
		// Without a role the user could never log in; roll back so the email is free again.
		_ = uc.endUserRepo.Delete(ctx, user.ID)
		_ = uc.countCache.Invalidate(ctx, txn.TenantID)
		return nil, ErrTemporarilyUnavailable
	}

	uc.auditHelper.LogSignupCompleted(ctx, txn.TenantID, client.ClientID, user.ID, req.SourceIP)

	_ = uc.txnRepo.ExtendTTL(ctx, req.TransactionID, TransactionTTL, SignupTransactionMaxLifetime)
	resp, err := uc.completer.Complete(ctx, req.TransactionID, user, txn.TenantID, client.ClientID, req.SourceIP)
	if err != nil {
		// The account exists; the user can still sign in with their password.
		return nil, ErrTemporarilyUnavailable
	}
	return resp, nil
}

// checkQuota enforces MaxUsersPerTenant (cache first, DB fallback).
func (uc *SelfSignup) checkQuota(ctx context.Context, tenantID string) error {
	count, hit, _ := uc.countCache.Get(ctx, tenantID)
	if !hit {
		var err error
		count, err = uc.endUserRepo.CountByTenant(ctx, tenantID)
		if err != nil {
			return ErrTemporarilyUnavailable
		}
		_ = uc.countCache.Set(ctx, tenantID, count)
	}
	if count >= entities.MaxUsersPerTenant {
		return ErrSignupUnavailable
	}
	return nil
}

// hashSignupCode binds a verification code to the authorization transaction
// that requested it, so a code is only valid in the browser that started the
// sign-up. This prevents another party from replacing the pending sign-up
// (with their own password) and having the owner verify it.
func hashSignupCode(transactionID, code string) string {
	digest := sha256.Sum256([]byte(transactionID + ":" + code))
	return hex.EncodeToString(digest[:])
}
