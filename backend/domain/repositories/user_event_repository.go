package repositories

import (
	"context"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
)

// UserEventRepository defines the interface for user activity event persistence
type UserEventRepository interface {
	// Record inserts a new event. Fire-and-forget; errors are logged but do not block the caller.
	Record(ctx context.Context, event *entities.UserEvent) error

	// ListByUser returns paginated activity events for a user in reverse chronological order
	ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*entities.UserEvent, int, error)

	// DeleteOlderThan deletes events older than the given timestamp (used by retention job)
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}
