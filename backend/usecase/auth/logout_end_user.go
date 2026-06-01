package auth

import (
	"context"

	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// LogoutEndUser terminates an end-user SSO session.
// It deletes the session from Redis (best-effort) and emits an audit event.
// The use case always succeeds — Redis failures are swallowed because the
// session cookie expiry on the client side is the authoritative logout signal.
type LogoutEndUser struct {
	sessionRepo repositories.SessionRepository
	auditHelper *OAuthAuditHelper
}

// NewLogoutEndUser constructs the use case.
func NewLogoutEndUser(
	sessionRepo repositories.SessionRepository,
	auditHelper *OAuthAuditHelper,
) *LogoutEndUser {
	return &LogoutEndUser{
		sessionRepo: sessionRepo,
		auditHelper: auditHelper,
	}
}

// Execute deletes the session identified by sessionID and records an audit event.
// It never returns an error: Redis deletion failures are ignored since the
// session will expire naturally, and the handler always clears the cookie.
func (uc *LogoutEndUser) Execute(ctx context.Context, sessionID, sourceIP string) error {
	// Attempt to look up the session so we can record the user ID in the
	// audit log. If the session is already gone (expired or concurrent
	// logout), we still emit the event without a user ID.
	var userID string
	if session, err := uc.sessionRepo.Get(ctx, sessionID); err == nil && session != nil {
		userID = session.UserID
	}

	// Best-effort delete. Swallow the error — the session TTL will clean
	// up stale entries, and the client-side cookie is already being
	// cleared by the handler.
	_ = uc.sessionRepo.Delete(ctx, sessionID)

	// Emit audit event. The helper never returns an error (it falls back
	// to structured logging on persistence failure).
	uc.auditHelper.LogLogout(ctx, userID, sourceIP)

	return nil
}