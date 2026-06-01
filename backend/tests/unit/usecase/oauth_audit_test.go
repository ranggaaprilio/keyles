package usecase_test

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// newHelper creates an OAuthAuditHelper with the given mock repo
// and returns both the helper and the mock for assertion setup.
func newHelper(repo *mocks.MockAuditRepository) *auth.OAuthAuditHelper {
	return auth.NewOAuthAuditHelper(repo, slog.Default())
}

// captureAuditLog extracts the AuditLog argument from the mock call.
func captureAuditCreate(repo *mocks.MockAuditRepository) *entities.AuditLog {
	args := repo.Called(mock.Anything, mock.Anything)
	_ = args.Error(0) // may be nil
	return args.Get(1).(*entities.AuditLog)
}

func TestOAuthAuditHelper_LogLoginSucceeded(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	clientID := "client_abc"

	t.Run("persists audit log with all identifiers", func(t *testing.T) {
		repo := new(mocks.MockAuditRepository)
		h := newHelper(repo)

		repo.On("Create", mock.Anything, mock.Anything).Return(nil)

		h.LogLoginSucceeded(ctx, tenantID.String(), clientID, userID.String(), "10.0.0.1")

		require.Len(t, repo.Calls, 1)
		entry := repo.Calls[0].Arguments.Get(1).(*entities.AuditLog)
		require.Equal(t, entities.EventOAuthLoginSucceeded, entry.EventType)
		require.NotNil(t, entry.TenantID)
		require.Equal(t, tenantID, *entry.TenantID)
		require.NotNil(t, entry.UserID)
		require.Equal(t, userID, *entry.UserID)
		require.Equal(t, "10.0.0.1", entry.IPAddress)
		require.Equal(t, clientID, entry.EventData["client_id"])
	})

	t.Run("handles empty identifiers gracefully", func(t *testing.T) {
		repo := new(mocks.MockAuditRepository)
		h := newHelper(repo)

		repo.On("Create", mock.Anything, mock.Anything).Return(nil)

		h.LogLoginSucceeded(ctx, "", "", "", "10.0.0.1")

		require.Len(t, repo.Calls, 1)
		entry := repo.Calls[0].Arguments.Get(1).(*entities.AuditLog)
		require.Equal(t, entities.EventOAuthLoginSucceeded, entry.EventType)
		require.Nil(t, entry.TenantID)
		require.Nil(t, entry.UserID)
		require.Equal(t, "10.0.0.1", entry.IPAddress)
	})
}

func TestOAuthAuditHelper_LogLoginFailed(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	t.Run("persists audit log with tenant and client", func(t *testing.T) {
		repo := new(mocks.MockAuditRepository)
		h := newHelper(repo)

		repo.On("Create", mock.Anything, mock.Anything).Return(nil)

		h.LogLoginFailed(ctx, tenantID.String(), "client_xyz", "192.168.1.1")

		require.Len(t, repo.Calls, 1)
		entry := repo.Calls[0].Arguments.Get(1).(*entities.AuditLog)
		require.Equal(t, entities.EventOAuthLoginFailed, entry.EventType)
		require.NotNil(t, entry.TenantID)
		require.Equal(t, tenantID, *entry.TenantID)
		require.Equal(t, "192.168.1.1", entry.IPAddress)
		require.Equal(t, "client_xyz", entry.EventData["client_id"])
	})
}

func TestOAuthAuditHelper_LogLoginThrottled(t *testing.T) {
	ctx := context.Background()

	t.Run("persists throttled event with source IP only", func(t *testing.T) {
		repo := new(mocks.MockAuditRepository)
		h := newHelper(repo)

		repo.On("Create", mock.Anything, mock.Anything).Return(nil)

		h.LogLoginThrottled(ctx, "10.0.0.5")

		require.Len(t, repo.Calls, 1)
		entry := repo.Calls[0].Arguments.Get(1).(*entities.AuditLog)
		require.Equal(t, entities.EventOAuthLoginThrottled, entry.EventType)
		require.Equal(t, "10.0.0.5", entry.IPAddress)
		require.Nil(t, entry.TenantID)
		require.Nil(t, entry.UserID)
	})
}

func TestOAuthAuditHelper_LogConsentApproved(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	txID := "tx_12345"

	t.Run("persists with all identifiers including transaction", func(t *testing.T) {
		repo := new(mocks.MockAuditRepository)
		h := newHelper(repo)

		repo.On("Create", mock.Anything, mock.Anything).Return(nil)

		h.LogConsentApproved(ctx, tenantID.String(), "client_def", userID.String(), txID)

		require.Len(t, repo.Calls, 1)
		entry := repo.Calls[0].Arguments.Get(1).(*entities.AuditLog)
		require.Equal(t, entities.EventOAuthConsentApproved, entry.EventType)
		require.NotNil(t, entry.TenantID)
		require.Equal(t, tenantID, *entry.TenantID)
		require.NotNil(t, entry.UserID)
		require.Equal(t, userID, *entry.UserID)
		require.Equal(t, "client_def", entry.EventData["client_id"])
		require.Equal(t, txID, entry.EventData["transaction_id"])
	})
}

func TestOAuthAuditHelper_LogConsentDenied(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	txID := "tx_denied_999"

	t.Run("persists denial with all identifiers", func(t *testing.T) {
		repo := new(mocks.MockAuditRepository)
		h := newHelper(repo)

		repo.On("Create", mock.Anything, mock.Anything).Return(nil)

		h.LogConsentDenied(ctx, tenantID.String(), "client_den", userID.String(), txID)

		require.Len(t, repo.Calls, 1)
		entry := repo.Calls[0].Arguments.Get(1).(*entities.AuditLog)
		require.Equal(t, entities.EventOAuthConsentDenied, entry.EventType)
		require.Equal(t, "client_den", entry.EventData["client_id"])
		require.Equal(t, txID, entry.EventData["transaction_id"])
	})
}

func TestOAuthAuditHelper_LogLogout(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("persists logout with user and IP", func(t *testing.T) {
		repo := new(mocks.MockAuditRepository)
		h := newHelper(repo)

		repo.On("Create", mock.Anything, mock.Anything).Return(nil)

		h.LogLogout(ctx, userID.String(), "172.16.0.1")

		require.Len(t, repo.Calls, 1)
		entry := repo.Calls[0].Arguments.Get(1).(*entities.AuditLog)
		require.Equal(t, entities.EventOAuthLogout, entry.EventType)
		require.NotNil(t, entry.UserID)
		require.Equal(t, userID, *entry.UserID)
		require.Equal(t, "172.16.0.1", entry.IPAddress)
	})
}

func TestOAuthAuditHelper_LogInvalidCallback(t *testing.T) {
	ctx := context.Background()

	t.Run("persists invalid callback with client ID and IP", func(t *testing.T) {
		repo := new(mocks.MockAuditRepository)
		h := newHelper(repo)

		repo.On("Create", mock.Anything, mock.Anything).Return(nil)

		h.LogInvalidCallback(ctx, "client_bad", "10.10.10.10")

		require.Len(t, repo.Calls, 1)
		entry := repo.Calls[0].Arguments.Get(1).(*entities.AuditLog)
		require.Equal(t, entities.EventOAuthInvalidCallback, entry.EventType)
		require.Equal(t, "10.10.10.10", entry.IPAddress)
		require.Equal(t, "client_bad", entry.EventData["client_id"])
	})
}

func TestOAuthAuditHelper_SecretExclusion(t *testing.T) {
	// Verify that no OAuth secrets ever appear in persisted audit logs.
	// Secrets include: passwords, cookies, authorization codes, PKCE values.
	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()

	secretKeys := []string{
		"password", "cookie", "authorization_code", "code_verifier",
		"code_challenge", "client_secret", "access_token", "refresh_token",
	}

	t.Run("login succeeded event contains no secrets", func(t *testing.T) {
		repo := new(mocks.MockAuditRepository)
		h := newHelper(repo)

		repo.On("Create", mock.Anything, mock.Anything).Return(nil)

		h.LogLoginSucceeded(ctx, tenantID.String(), "client_id", userID.String(), "1.2.3.4")

		entry := repo.Calls[0].Arguments.Get(1).(*entities.AuditLog)
		for _, key := range secretKeys {
			_, exists := entry.EventData[key]
			require.False(t, exists, "audit event data must not contain secret key %q", key)
		}
	})

	t.Run("consent approved event contains no secrets", func(t *testing.T) {
		repo := new(mocks.MockAuditRepository)
		h := newHelper(repo)

		repo.On("Create", mock.Anything, mock.Anything).Return(nil)

		h.LogConsentApproved(ctx, tenantID.String(), "client_id", userID.String(), "tx_1")

		entry := repo.Calls[0].Arguments.Get(1).(*entities.AuditLog)
		for _, key := range secretKeys {
			_, exists := entry.EventData[key]
			require.False(t, exists, "audit event data must not contain secret key %q", key)
		}
	})
}

func TestOAuthAuditHelper_RepositoryWriteFailure(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()

	t.Run("falls back to slog on repository error", func(t *testing.T) {
		repo := new(mocks.MockAuditRepository)
		var logBuf strings.Builder
		handler := slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelError})
		logger := slog.New(handler)
		h := auth.NewOAuthAuditHelper(repo, logger)

		repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db connection lost"))

		// Must not panic or return error — the call is void
		h.LogLoginSucceeded(ctx, tenantID.String(), "client_1", userID.String(), "10.0.0.1")

		repo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
		// Fallback log was emitted
		output := logBuf.String()
		require.Contains(t, output, "audit_persistence_failed")
		require.Contains(t, output, string(entities.EventOAuthLoginSucceeded))
		require.Contains(t, output, "db connection lost")
	})

	t.Run("falls back on throttled event", func(t *testing.T) {
		repo := new(mocks.MockAuditRepository)
		var logBuf strings.Builder
		handler := slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelError})
		logger := slog.New(handler)
		h := auth.NewOAuthAuditHelper(repo, logger)

		repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("write timeout"))

		h.LogLoginThrottled(ctx, "10.0.0.5")

		repo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
		output := logBuf.String()
		require.Contains(t, output, "audit_persistence_failed")
		require.Contains(t, output, string(entities.EventOAuthLoginThrottled))
	})

	t.Run("falls back on consent approved", func(t *testing.T) {
		repo := new(mocks.MockAuditRepository)
		var logBuf strings.Builder
		handler := slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelError})
		logger := slog.New(handler)
		h := auth.NewOAuthAuditHelper(repo, logger)

		repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("disk full"))

		h.LogConsentApproved(ctx, tenantID.String(), "client_2", userID.String(), "tx_555")

		repo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
		output := logBuf.String()
		require.Contains(t, output, "audit_persistence_failed")
		require.Contains(t, output, string(entities.EventOAuthConsentApproved))
	})

	t.Run("falls back on consent denied", func(t *testing.T) {
		repo := new(mocks.MockAuditRepository)
		var logBuf strings.Builder
		handler := slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelError})
		logger := slog.New(handler)
		h := auth.NewOAuthAuditHelper(repo, logger)

		repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("connection refused"))

		h.LogConsentDenied(ctx, tenantID.String(), "client_3", userID.String(), "tx_666")

		repo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
		output := logBuf.String()
		require.Contains(t, output, "audit_persistence_failed")
	})

	t.Run("falls back on logout", func(t *testing.T) {
		repo := new(mocks.MockAuditRepository)
		var logBuf strings.Builder
		handler := slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelError})
		logger := slog.New(handler)
		h := auth.NewOAuthAuditHelper(repo, logger)

		repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("unavailable"))

		h.LogLogout(ctx, userID.String(), "10.0.0.99")

		repo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
		output := logBuf.String()
		require.Contains(t, output, "audit_persistence_failed")
		require.Contains(t, output, string(entities.EventOAuthLogout))
	})

	t.Run("falls back on invalid callback", func(t *testing.T) {
		repo := new(mocks.MockAuditRepository)
		var logBuf strings.Builder
		handler := slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelError})
		logger := slog.New(handler)
		h := auth.NewOAuthAuditHelper(repo, logger)

		repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("timeout"))

		h.LogInvalidCallback(ctx, "client_bad", "10.10.10.10")

		repo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
		output := logBuf.String()
		require.Contains(t, output, "audit_persistence_failed")
		require.Contains(t, output, string(entities.EventOAuthInvalidCallback))
	})
}

func TestOAuthAuditHelper_InvalidUUID(t *testing.T) {
	ctx := context.Background()

	t.Run("skips unparseable tenant and user IDs", func(t *testing.T) {
		repo := new(mocks.MockAuditRepository)
		h := newHelper(repo)

		repo.On("Create", mock.Anything, mock.Anything).Return(nil)

		h.LogLoginSucceeded(ctx, "not-a-uuid", "client_1", "also-not-uuid", "1.2.3.4")

		require.Len(t, repo.Calls, 1)
		entry := repo.Calls[0].Arguments.Get(1).(*entities.AuditLog)
		require.Nil(t, entry.TenantID, "invalid tenant UUID should be skipped, not stored")
		require.Nil(t, entry.UserID, "invalid user UUID should be skipped, not stored")
		// client_id is a string, so it should still be stored
		require.Equal(t, "client_1", entry.EventData["client_id"])
	})
}

func TestOAuthAuditHelper_DefaultLogger(t *testing.T) {
	t.Run("uses slog.Default when no logger provided", func(t *testing.T) {
		repo := new(mocks.MockAuditRepository)
		repo.On("Create", mock.Anything, mock.Anything).Return(nil)

		h := auth.NewOAuthAuditHelper(repo)
		require.NotNil(t, h)
	})
}