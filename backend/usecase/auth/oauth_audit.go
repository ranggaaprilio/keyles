package auth

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// OAuthAuditHelper creates sanitized OAuth audit events and falls back to
// structured logging when persistence fails.  Audit persistence failure
// MUST NOT become an OAuth dependency — the helper never returns an error.
type OAuthAuditHelper struct {
	auditRepo repositories.AuditRepository
	logger    *slog.Logger
}

// NewOAuthAuditHelper creates a new OAuthAuditHelper.
// If logger is nil, slog.Default() is used.
func NewOAuthAuditHelper(auditRepo repositories.AuditRepository, logger ...*slog.Logger) *OAuthAuditHelper {
	var l *slog.Logger
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	} else {
		l = slog.Default()
	}
	return &OAuthAuditHelper{
		auditRepo: auditRepo,
		logger:    l,
	}
}

// persist attempts to write the audit log to the repository.
// On failure it logs the sanitized event via structured logging.
func (h *OAuthAuditHelper) persist(ctx context.Context, entry *entities.AuditLog) {
	if err := h.auditRepo.Create(ctx, entry); err != nil {
		h.logFallback(entry, err)
	}
}

// logFallback emits a structured log entry when repository persistence fails.
func (h *OAuthAuditHelper) logFallback(entry *entities.AuditLog, repoErr error) {
	attrs := []slog.Attr{
		slog.String("event_type", string(entry.EventType)),
		slog.String("ip_address", entry.IPAddress),
	}
	if entry.TenantID != nil {
		attrs = append(attrs, slog.String("tenant_id", entry.TenantID.String()))
	}
	if entry.UserID != nil {
		attrs = append(attrs, slog.String("user_id", entry.UserID.String()))
	}
	for k, v := range entry.EventData {
		attrs = append(attrs, slog.Any(k, v))
	}

	h.logger.LogAttrs(nil, slog.LevelError, "audit_persistence_failed",
		append(attrs, slog.String("repository_error", repoErr.Error()))...,
	)
}

// LogLoginSucceeded records a successful OAuth login.
func (h *OAuthAuditHelper) LogLoginSucceeded(ctx context.Context, tenantID, clientID, userID, sourceIP string) {
	entry := entities.NewAuditLog(entities.EventOAuthLoginSucceeded, sourceIP, "")
	if tenantID != "" {
		if id, err := uuid.Parse(tenantID); err == nil {
			entry.WithTenant(id)
		}
	}
	if userID != "" {
		if id, err := uuid.Parse(userID); err == nil {
			entry.WithUser(id)
		}
	}
	if clientID != "" {
		entry.WithData("client_id", clientID)
	}
	h.persist(ctx, entry)
}

// LogLoginFailed records a failed OAuth login.
func (h *OAuthAuditHelper) LogLoginFailed(ctx context.Context, tenantID, clientID, sourceIP string) {
	entry := entities.NewAuditLog(entities.EventOAuthLoginFailed, sourceIP, "")
	if tenantID != "" {
		if id, err := uuid.Parse(tenantID); err == nil {
			entry.WithTenant(id)
		}
	}
	if clientID != "" {
		entry.WithData("client_id", clientID)
	}
	h.persist(ctx, entry)
}

// LogLoginThrottled records a throttled OAuth login attempt.
func (h *OAuthAuditHelper) LogLoginThrottled(ctx context.Context, sourceIP string) {
	entry := entities.NewAuditLog(entities.EventOAuthLoginThrottled, sourceIP, "")
	h.persist(ctx, entry)
}

// LogConsentApproved records user consent approval for an OAuth transaction.
func (h *OAuthAuditHelper) LogConsentApproved(ctx context.Context, tenantID, clientID, userID, transactionID string) {
	entry := entities.NewAuditLog(entities.EventOAuthConsentApproved, "", "")
	if tenantID != "" {
		if id, err := uuid.Parse(tenantID); err == nil {
			entry.WithTenant(id)
		}
	}
	if userID != "" {
		if id, err := uuid.Parse(userID); err == nil {
			entry.WithUser(id)
		}
	}
	if clientID != "" {
		entry.WithData("client_id", clientID)
	}
	if transactionID != "" {
		entry.WithData("transaction_id", transactionID)
	}
	h.persist(ctx, entry)
}

// LogConsentDenied records user consent denial for an OAuth transaction.
func (h *OAuthAuditHelper) LogConsentDenied(ctx context.Context, tenantID, clientID, userID, transactionID string) {
	entry := entities.NewAuditLog(entities.EventOAuthConsentDenied, "", "")
	if tenantID != "" {
		if id, err := uuid.Parse(tenantID); err == nil {
			entry.WithTenant(id)
		}
	}
	if userID != "" {
		if id, err := uuid.Parse(userID); err == nil {
			entry.WithUser(id)
		}
	}
	if clientID != "" {
		entry.WithData("client_id", clientID)
	}
	if transactionID != "" {
		entry.WithData("transaction_id", transactionID)
	}
	h.persist(ctx, entry)
}

// LogLogout records an OAuth logout event.
func (h *OAuthAuditHelper) LogLogout(ctx context.Context, userID, sourceIP string) {
	entry := entities.NewAuditLog(entities.EventOAuthLogout, sourceIP, "")
	if userID != "" {
		if id, err := uuid.Parse(userID); err == nil {
			entry.WithUser(id)
		}
	}
	h.persist(ctx, entry)
}

// LogInvalidCallback records an invalid OAuth callback attempt.
func (h *OAuthAuditHelper) LogInvalidCallback(ctx context.Context, clientID, sourceIP string) {
	entry := entities.NewAuditLog(entities.EventOAuthInvalidCallback, sourceIP, "")
	if clientID != "" {
		entry.WithData("client_id", clientID)
	}
	h.persist(ctx, entry)
}