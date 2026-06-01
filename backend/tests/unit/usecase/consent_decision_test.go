package usecase_test

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func pendingConsentTransaction() *repositories.AuthorizationTransaction {
	return &repositories.AuthorizationTransaction{
		TransactionID:        "txn-consent",
		ClientID:             "client-1",
		TenantID:             "tenant-1",
		UserID:               "user-1",
		SessionID:            "session-1",
		RedirectURI:          "https://app.example/callback?existing=value",
		State:                "state-1",
		Scope:                "openid email",
		CodeChallenge:        "challenge",
		CodeChallengeMethod:  "S256",
		InteractionCSRFToken: "csrf-1",
		Stage:                repositories.TransactionStagePendingConsent,
	}
}

func newConsentDecision(
	txnRepo *mocks.MockAuthorizationTransactionRepository,
	sessionRepo *mocks.MockSessionRepository,
	endUserRepo *mocks.MockEndUserRepository,
	roleRepo *mocks.MockRoleRepository,
	authCodeRepo *mocks.MockAuthCodeRepository,
	auditRepo *mocks.MockAuditRepository,
) *auth.ConsentDecision {
	return auth.NewConsentDecision(
		txnRepo,
		sessionRepo,
		new(mocks.MockClientRepository),
		endUserRepo,
		roleRepo,
		authCodeRepo,
		auth.NewOAuthAuditHelper(auditRepo),
	)
}

func TestConsentDecision_InvalidCSRFDoesNotConsumeTransaction(t *testing.T) {
	ctx := context.Background()
	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Get", ctx, "txn-consent").Return(pendingConsentTransaction(), nil)
	uc := newConsentDecision(
		txnRepo,
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockRoleRepository),
		new(mocks.MockAuthCodeRepository),
		new(mocks.MockAuditRepository),
	)

	result, err := uc.Execute(ctx, auth.ConsentDecisionRequest{
		TransactionID:        "txn-consent",
		InteractionCSRFToken: "wrong",
		SessionID:            "session-1",
		Approved:             true,
	})

	require.Nil(t, result)
	require.ErrorIs(t, err, auth.ErrInvalidCSRFToken)
	txnRepo.AssertNotCalled(t, "Complete", mock.Anything, mock.Anything)
}

func TestConsentDecision_ApprovalPreservesCallbackQuery(t *testing.T) {
	ctx := context.Background()
	txn := pendingConsentTransaction()
	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	authCodeRepo := new(mocks.MockAuthCodeRepository)
	auditRepo := new(mocks.MockAuditRepository)
	uc := newConsentDecision(txnRepo, sessionRepo, endUserRepo, roleRepo, authCodeRepo, auditRepo)

	txnRepo.On("Get", ctx, "txn-consent").Return(txn, nil)
	sessionRepo.On("Get", ctx, "session-1").Return(&repositories.Session{
		SessionID:       "session-1",
		UserID:          "user-1",
		TenantID:        "tenant-1",
		AuthenticatedAt: time.Unix(1700000000, 0),
	}, nil)
	endUserRepo.On("GetByID", ctx, "user-1").Return(&entities.User{ID: "user-1", Status: entities.UserStatusActive}, nil)
	roleRepo.On("HasAnyRole", ctx, "user-1", "client-1").Return(true, nil)
	txnRepo.On("Complete", ctx, "txn-consent").Return(txn, nil)
	authCodeRepo.On("Store", ctx, mock.AnythingOfType("*entities.AuthorizationCode"), auth.AuthCodeTTL).Return(nil)
	auditRepo.On("Create", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	result, err := uc.Execute(ctx, auth.ConsentDecisionRequest{
		TransactionID:        "txn-consent",
		InteractionCSRFToken: "csrf-1",
		SessionID:            "session-1",
		Approved:             true,
	})

	require.NoError(t, err)
	callback, err := url.Parse(result.RedirectURL)
	require.NoError(t, err)
	require.Equal(t, "value", callback.Query().Get("existing"))
	require.Equal(t, "state-1", callback.Query().Get("state"))
	require.NotEmpty(t, callback.Query().Get("code"))
}

func TestGetConsentDetails_RedisFailureIsTemporaryUnavailable(t *testing.T) {
	ctx := context.Background()
	sessionRepo := new(mocks.MockSessionRepository)
	sessionRepo.On("Get", ctx, "session-1").Return(nil, errors.New("redis down"))
	uc := auth.NewGetConsentDetails(
		new(mocks.MockAuthorizationTransactionRepository),
		sessionRepo,
		new(mocks.MockClientRepository),
		new(mocks.MockEndUserRepository),
	)

	result, err := uc.Execute(ctx, "txn-consent", "session-1")

	require.Nil(t, result)
	require.ErrorIs(t, err, auth.ErrTemporarilyUnavailable)
}

func TestConsentDecision_DenialCompletesTransactionAndPreservesState(t *testing.T) {
	ctx := context.Background()
	txn := pendingConsentTransaction()
	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	auditRepo := new(mocks.MockAuditRepository)

	txnRepo.On("Get", ctx, "txn-consent").Return(txn, nil)
	sessionRepo.On("Get", ctx, "session-1").Return(boundConsentSession(), nil)
	endUserRepo.On("GetByID", ctx, "user-1").Return(&entities.User{ID: "user-1", Status: entities.UserStatusActive}, nil)
	roleRepo.On("HasAnyRole", ctx, "user-1", "client-1").Return(true, nil)
	txnRepo.On("Complete", ctx, "txn-consent").Return(txn, nil)
	auditRepo.On("Create", ctx, mock.MatchedBy(func(log *entities.AuditLog) bool {
		return log.EventType == entities.EventOAuthConsentDenied
	})).Return(nil)

	result, err := newConsentDecision(
		txnRepo,
		sessionRepo,
		endUserRepo,
		roleRepo,
		new(mocks.MockAuthCodeRepository),
		auditRepo,
	).Execute(ctx, auth.ConsentDecisionRequest{
		TransactionID:        "txn-consent",
		InteractionCSRFToken: "csrf-1",
		SessionID:            "session-1",
		Approved:             false,
	})

	require.NoError(t, err)
	require.True(t, result.IsDenied)
	callback, err := url.Parse(result.RedirectURL)
	require.NoError(t, err)
	require.Equal(t, "access_denied", callback.Query().Get("error"))
	require.Equal(t, "state-1", callback.Query().Get("state"))
	require.Equal(t, "value", callback.Query().Get("existing"))
}

func TestConsentDecision_ReplayAndRedisFailures(t *testing.T) {
	ctx := context.Background()

	t.Run("replay rejected", func(t *testing.T) {
		txn := pendingConsentTransaction()
		txnRepo := new(mocks.MockAuthorizationTransactionRepository)
		sessionRepo := new(mocks.MockSessionRepository)
		endUserRepo := new(mocks.MockEndUserRepository)
		roleRepo := new(mocks.MockRoleRepository)
		txnRepo.On("Get", ctx, "txn-consent").Return(txn, nil)
		sessionRepo.On("Get", ctx, "session-1").Return(boundConsentSession(), nil)
		endUserRepo.On("GetByID", ctx, "user-1").Return(&entities.User{ID: "user-1", Status: entities.UserStatusActive}, nil)
		roleRepo.On("HasAnyRole", ctx, "user-1", "client-1").Return(true, nil)
		txnRepo.On("Complete", ctx, "txn-consent").Return(nil, errors.New(repositories.ErrTransactionAlreadyCompleted))

		result, err := newConsentDecision(txnRepo, sessionRepo, endUserRepo, roleRepo, new(mocks.MockAuthCodeRepository), new(mocks.MockAuditRepository)).Execute(ctx, auth.ConsentDecisionRequest{
			TransactionID: "txn-consent", InteractionCSRFToken: "csrf-1", SessionID: "session-1", Approved: true,
		})
		require.Nil(t, result)
		require.ErrorIs(t, err, auth.ErrTransactionExpired)
	})

	t.Run("redis completion failure fails closed", func(t *testing.T) {
		txn := pendingConsentTransaction()
		txnRepo := new(mocks.MockAuthorizationTransactionRepository)
		sessionRepo := new(mocks.MockSessionRepository)
		endUserRepo := new(mocks.MockEndUserRepository)
		roleRepo := new(mocks.MockRoleRepository)
		txnRepo.On("Get", ctx, "txn-consent").Return(txn, nil)
		sessionRepo.On("Get", ctx, "session-1").Return(boundConsentSession(), nil)
		endUserRepo.On("GetByID", ctx, "user-1").Return(&entities.User{ID: "user-1", Status: entities.UserStatusActive}, nil)
		roleRepo.On("HasAnyRole", ctx, "user-1", "client-1").Return(true, nil)
		txnRepo.On("Complete", ctx, "txn-consent").Return(nil, errors.New("redis down"))

		result, err := newConsentDecision(txnRepo, sessionRepo, endUserRepo, roleRepo, new(mocks.MockAuthCodeRepository), new(mocks.MockAuditRepository)).Execute(ctx, auth.ConsentDecisionRequest{
			TransactionID: "txn-consent", InteractionCSRFToken: "csrf-1", SessionID: "session-1", Approved: true,
		})
		require.Nil(t, result)
		require.ErrorIs(t, err, auth.ErrTemporarilyUnavailable)
	})
}
