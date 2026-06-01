package usecase_test

import (
	"context"
	"testing"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newLogoutEndUser(
	sessionRepo *mocks.MockSessionRepository,
	auditRepo *mocks.MockAuditRepository,
) *auth.LogoutEndUser {
	auditHelper := auth.NewOAuthAuditHelper(auditRepo)
	return auth.NewLogoutEndUser(sessionRepo, auditHelper)
}

func TestLogoutEndUser_IdempotentDeletion(t *testing.T) {
	ctx := context.Background()
	sessionRepo := new(mocks.MockSessionRepository)
	auditRepo := new(mocks.MockAuditRepository)
	uc := newLogoutEndUser(sessionRepo, auditRepo)

	session := mocks.NewMockSession(func(s *repositories.Session) {
		s.SessionID = "sess_logout_001"
		s.UserID = "user_001"
	})

	sessionRepo.On("Get", ctx, "sess_logout_001").Return(session, nil)
	sessionRepo.On("Delete", ctx, "sess_logout_001").Return(nil)
	auditRepo.On("Create", ctx, mockAuditLogMatcher(entities.EventOAuthLogout)).Return(nil)

	err := uc.Execute(ctx, "sess_logout_001", "192.0.2.1")
	require.NoError(t, err)
	sessionRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestLogoutEndUser_SessionNotFound(t *testing.T) {
	ctx := context.Background()
	sessionRepo := new(mocks.MockSessionRepository)
	auditRepo := new(mocks.MockAuditRepository)
	uc := newLogoutEndUser(sessionRepo, auditRepo)

	// Session already gone — Get returns nil
	sessionRepo.On("Get", ctx, "sess_gone").Return(nil, nil)
	sessionRepo.On("Delete", ctx, "sess_gone").Return(nil)
	auditRepo.On("Create", ctx, mockAuditLogMatcher(entities.EventOAuthLogout)).Return(nil)

	err := uc.Execute(ctx, "sess_gone", "192.0.2.2")
	require.NoError(t, err)
	sessionRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestLogoutEndUser_RedisFailureTolerance(t *testing.T) {
	ctx := context.Background()

	t.Run("Get fails", func(t *testing.T) {
		sessionRepo := new(mocks.MockSessionRepository)
		auditRepo := new(mocks.MockAuditRepository)
		uc := newLogoutEndUser(sessionRepo, auditRepo)

		// Get returns an error (Redis down)
		sessionRepo.On("Get", ctx, "sess_redis_fail").Return(nil, errRedisTransient)
		sessionRepo.On("Delete", ctx, "sess_redis_fail").Return(nil)
		// Audit still fires, but without user ID
		auditRepo.On("Create", ctx, mockAuditLogMatcher(entities.EventOAuthLogout)).Return(nil)

		err := uc.Execute(ctx, "sess_redis_fail", "192.0.2.3")
		require.NoError(t, err)
		sessionRepo.AssertExpectations(t)
		auditRepo.AssertExpectations(t)
	})

	t.Run("Delete fails", func(t *testing.T) {
		sessionRepo := new(mocks.MockSessionRepository)
		auditRepo := new(mocks.MockAuditRepository)
		uc := newLogoutEndUser(sessionRepo, auditRepo)

		session := mocks.NewMockSession(func(s *repositories.Session) {
			s.SessionID = "sess_del_fail"
			s.UserID = "user_del_fail"
		})

		sessionRepo.On("Get", ctx, "sess_del_fail").Return(session, nil)
		// Delete fails — but Execute still succeeds (best-effort)
		sessionRepo.On("Delete", ctx, "sess_del_fail").Return(errRedisTransient)
		auditRepo.On("Create", ctx, mockAuditLogMatcher(entities.EventOAuthLogout)).Return(nil)

		err := uc.Execute(ctx, "sess_del_fail", "192.0.2.4")
		require.NoError(t, err)
		sessionRepo.AssertExpectations(t)
		auditRepo.AssertExpectations(t)
	})

	t.Run("Both Get and Delete fail", func(t *testing.T) {
		sessionRepo := new(mocks.MockSessionRepository)
		auditRepo := new(mocks.MockAuditRepository)
		uc := newLogoutEndUser(sessionRepo, auditRepo)

		sessionRepo.On("Get", ctx, "sess_total_fail").Return(nil, errRedisTransient)
		sessionRepo.On("Delete", ctx, "sess_total_fail").Return(errRedisTransient)
		auditRepo.On("Create", ctx, mockAuditLogMatcher(entities.EventOAuthLogout)).Return(nil)

		err := uc.Execute(ctx, "sess_total_fail", "192.0.2.5")
		require.NoError(t, err)
		sessionRepo.AssertExpectations(t)
		auditRepo.AssertExpectations(t)
	})
}

func TestLogoutEndUser_AuditEmission(t *testing.T) {
	ctx := context.Background()
	sessionRepo := new(mocks.MockSessionRepository)
	auditRepo := new(mocks.MockAuditRepository)
	uc := newLogoutEndUser(sessionRepo, auditRepo)

	session := mocks.NewMockSession(func(s *repositories.Session) {
		s.SessionID = "sess_audit_001"
		s.UserID = "user_audit_001"
	})

	sessionRepo.On("Get", ctx, "sess_audit_001").Return(session, nil)
	sessionRepo.On("Delete", ctx, "sess_audit_001").Return(nil)

	// Verify audit event is created with the correct event type
	auditRepo.On("Create", ctx, mockAuditLogMatcher(entities.EventOAuthLogout)).Return(nil)

	err := uc.Execute(ctx, "sess_audit_001", "10.0.0.1")
	require.NoError(t, err)
	sessionRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

// mockAuditLogMatcher returns a mock.Matcher that verifies the audit log has
// the expected event type. This avoids coupling tests to specific UUID or
// timestamp values.
func mockAuditLogMatcher(expectedType entities.EventType) interface{} {
	return mock.MatchedBy(func(log *entities.AuditLog) bool {
		return log.EventType == expectedType
	})
}

// errRedisTransient simulates a transient Redis error in tests.
var errRedisTransient = &redisTransientError{}

type redisTransientError struct{}

func (e *redisTransientError) Error() string { return "redis: transient error" }
