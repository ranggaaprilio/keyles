package repositories

import (
	"context"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
)

// UserEventRepository defines the interface for user event persistence operations
type UserEventRepository interface {
	// Record stores a new user event
	Record(ctx context.Context, event *entities.UserEvent) error

	// ListByUser retrieves paginated events for a specific user
	ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*entities.UserEvent, int, error)

	// DeleteOlderThan removes events older than the specified time
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}
