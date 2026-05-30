package postgres

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"gorm.io/gorm"
)

// PostgresUserEvent is the GORM model for user_events table
type PostgresUserEvent struct {
	ID          int64           `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID    string          `gorm:"column:tenant_id;not null;index"`
	UserID      string          `gorm:"column:user_id;not null;index"`
	ClientID    *string         `gorm:"column:client_id;index"`
	EventType   string          `gorm:"column:event_type;not null;index"`
	IPAddress   *string         `gorm:"column:ip_address"`
	CountryCode *string         `gorm:"column:country_code"`
	Details     json.RawMessage `gorm:"column:details;type:jsonb"`
	OccurredAt  time.Time       `gorm:"column:occurred_at;not null;index"`
}

func (PostgresUserEvent) TableName() string {
	return "user_events"
}

// PostgresUserEventRepository implements UserEventRepository using PostgreSQL
type PostgresUserEventRepository struct {
	db *gorm.DB
}

// NewPostgresUserEventRepository creates a new PostgreSQL user event repository
func NewPostgresUserEventRepository(db *gorm.DB) repositories.UserEventRepository {
	return &PostgresUserEventRepository{db: db}
}

// Record stores a new user event (fire-and-forget)
func (r *PostgresUserEventRepository) Record(ctx context.Context, event *entities.UserEvent) error {
	model, err := fromUserEventEntity(event)
	if err != nil {
		log.Printf("failed to marshal user event details: %v", err)
		return nil
	}

	err = r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		log.Printf("failed to record user event: %v", err)
	}
	return nil
}

// ListByUser retrieves paginated events for a specific user
func (r *PostgresUserEventRepository) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*entities.UserEvent, int, error) {
	var models []PostgresUserEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&PostgresUserEvent{}).Where("user_id = ?", userID)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("occurred_at DESC").Offset(offset).Limit(pageSize).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	events := make([]*entities.UserEvent, len(models))
	for i, model := range models {
		events[i], err = model.toEntity()
		if err != nil {
			return nil, 0, err
		}
	}

	return events, int(total), nil
}

// DeleteOlderThan removes events older than the specified time
func (r *PostgresUserEventRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("occurred_at < ?", before).
		Delete(&PostgresUserEvent{})
	return result.RowsAffected, result.Error
}

// toEntity converts PostgresUserEvent to domain entity
func (m *PostgresUserEvent) toEntity() (*entities.UserEvent, error) {
	var details map[string]any
	if len(m.Details) > 0 {
		if err := json.Unmarshal(m.Details, &details); err != nil {
			return nil, err
		}
	}

	return &entities.UserEvent{
		ID:          m.ID,
		TenantID:    m.TenantID,
		UserID:      m.UserID,
		ClientID:    m.ClientID,
		EventType:   entities.UserEventType(m.EventType),
		IPAddress:   m.IPAddress,
		CountryCode: m.CountryCode,
		Details:     details,
		OccurredAt:  m.OccurredAt,
	}, nil
}

// fromUserEventEntity converts domain entity to PostgresUserEvent
func fromUserEventEntity(event *entities.UserEvent) (*PostgresUserEvent, error) {
	details := json.RawMessage("{}")
	if event.Details != nil {
		b, err := json.Marshal(event.Details)
		if err != nil {
			return nil, err
		}
		details = b
	}

	return &PostgresUserEvent{
		ID:          event.ID,
		TenantID:    event.TenantID,
		UserID:      event.UserID,
		ClientID:    event.ClientID,
		EventType:   string(event.EventType),
		IPAddress:   event.IPAddress,
		CountryCode: event.CountryCode,
		Details:     details,
		OccurredAt:  event.OccurredAt,
	}, nil
}

// Verify interface compliance
var _ repositories.UserEventRepository = (*PostgresUserEventRepository)(nil)
