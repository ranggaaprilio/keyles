package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
)

// AuditRepository defines the interface for audit log persistence operations
type AuditRepository interface {
	// Create creates a new audit log entry
	Create(ctx context.Context, log *entities.AuditLog) error

	// FindByTenantID retrieves audit logs for a tenant
	FindByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entities.AuditLog, error)

	// FindByEventType retrieves audit logs by event type
	FindByEventType(ctx context.Context, eventType entities.EventType, limit, offset int) ([]*entities.AuditLog, error)

	// FindRecent retrieves the most recent audit logs
	FindRecent(ctx context.Context, limit int) ([]*entities.AuditLog, error)
}
