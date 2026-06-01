package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/stretchr/testify/require"
)

func newGetConsentDetails(
	txnRepo *mocks.MockAuthorizationTransactionRepository,
	sessionRepo *mocks.MockSessionRepository,
	clientRepo *mocks.MockClientRepository,
	endUserRepo *mocks.MockEndUserRepository,
) *auth.GetConsentDetails {
	return auth.NewGetConsentDetails(txnRepo, sessionRepo, clientRepo, endUserRepo)
}

func boundConsentSession() *repositories.Session {
	return &repositories.Session{SessionID: "session-1", UserID: "user-1", TenantID: "tenant-1"}
}

func TestGetConsentDetails_ReturnsDisplaySafeFields(t *testing.T) {
	ctx := context.Background()
	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)

	sessionRepo.On("Get", ctx, "session-1").Return(boundConsentSession(), nil)
	txnRepo.On("Get", ctx, "txn-consent").Return(pendingConsentTransaction(), nil)
	clientRepo.On("GetByID", ctx, "client-1").Return(&entities.Client{ClientID: "client-1", ClientName: "Example App"}, nil)
	endUserRepo.On("GetByID", ctx, "user-1").Return(&entities.User{ID: "user-1", DisplayName: "Alice"}, nil)

	result, err := newGetConsentDetails(txnRepo, sessionRepo, clientRepo, endUserRepo).Execute(ctx, "txn-consent", "session-1")

	require.NoError(t, err)
	require.Equal(t, "Example App", result.ClientName)
	require.Equal(t, "Alice", result.UserDisplay)
	require.Equal(t, []string{"openid", "email"}, result.Scopes)
	require.Equal(t, "csrf-1", result.InteractionCSRFToken)
}

func TestGetConsentDetails_RejectsMissingOrMismatchedSession(t *testing.T) {
	ctx := context.Background()

	t.Run("missing", func(t *testing.T) {
		sessionRepo := new(mocks.MockSessionRepository)
		sessionRepo.On("Get", ctx, "missing").Return(nil, nil)
		result, err := newGetConsentDetails(
			new(mocks.MockAuthorizationTransactionRepository),
			sessionRepo,
			new(mocks.MockClientRepository),
			new(mocks.MockEndUserRepository),
		).Execute(ctx, "txn-consent", "missing")
		require.Nil(t, result)
		require.ErrorIs(t, err, auth.ErrSessionMissing)
	})

	t.Run("mismatched session binding", func(t *testing.T) {
		sessionRepo := new(mocks.MockSessionRepository)
		txnRepo := new(mocks.MockAuthorizationTransactionRepository)
		sessionRepo.On("Get", ctx, "session-other").Return(&repositories.Session{SessionID: "session-other", UserID: "user-1"}, nil)
		txnRepo.On("Get", ctx, "txn-consent").Return(pendingConsentTransaction(), nil)
		result, err := newGetConsentDetails(
			txnRepo,
			sessionRepo,
			new(mocks.MockClientRepository),
			new(mocks.MockEndUserRepository),
		).Execute(ctx, "txn-consent", "session-other")
		require.Nil(t, result)
		require.ErrorIs(t, err, auth.ErrSessionUserMismatch)
	})
}

func TestGetConsentDetails_HandlesExpiredAndRedisFailures(t *testing.T) {
	ctx := context.Background()

	t.Run("expired transaction", func(t *testing.T) {
		sessionRepo := new(mocks.MockSessionRepository)
		txnRepo := new(mocks.MockAuthorizationTransactionRepository)
		sessionRepo.On("Get", ctx, "session-1").Return(boundConsentSession(), nil)
		txnRepo.On("Get", ctx, "expired").Return(nil, nil)
		result, err := newGetConsentDetails(txnRepo, sessionRepo, new(mocks.MockClientRepository), new(mocks.MockEndUserRepository)).Execute(ctx, "expired", "session-1")
		require.Nil(t, result)
		require.ErrorIs(t, err, auth.ErrTransactionExpired)
	})

	t.Run("redis unavailable", func(t *testing.T) {
		sessionRepo := new(mocks.MockSessionRepository)
		sessionRepo.On("Get", ctx, "session-1").Return(nil, errors.New("redis down"))
		result, err := newGetConsentDetails(
			new(mocks.MockAuthorizationTransactionRepository),
			sessionRepo,
			new(mocks.MockClientRepository),
			new(mocks.MockEndUserRepository),
		).Execute(ctx, "txn-consent", "session-1")
		require.Nil(t, result)
		require.ErrorIs(t, err, auth.ErrTemporarilyUnavailable)
	})
}
