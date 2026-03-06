package postgres

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// UserEventModel is the GORM model for the user_events table
type UserEventModel struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID    string    `gorm:"column:tenant_id;type:uuid;not null;index"`
	UserID      string    `gorm:"column:user_id;type:uuid;not null;index"`
	ClientID    *string   `gorm:"column:client_id;type:uuid"`
	EventType   string    `gorm:"column:event_type;type:varchar(50);not null"`
	IPAddress   *string   `gorm:"column:ip_address;type:inet"`
	CountryCode *string   `gorm:"column:country_code;type:char(2)"`
	Details     []byte    `gorm:"column:details;type:jsonb"`
	OccurredAt  time.Time `gorm:"column:occurred_at;not null;default:now()"`
}

func (UserEventModel) TableName() string {
	return "user_events"
}

func (m *UserEventModel) toEntity() *entities.UserEvent {
	ev := &entities.UserEvent{
		ID:          m.ID,
		TenantID:    m.TenantID,
		UserID:      m.UserID,
		ClientID:    m.ClientID,
		EventType:   entities.UserEventType(m.EventType),
		IPAddress:   m.IPAddress,
		CountryCode: m.CountryCode,
		OccurredAt:  m.OccurredAt,
	}
	if len(m.Details) > 0 {
		_ = json.Unmarshal(m.Details, &ev.Details)
	}
	return ev
}

func fromUserEventEntity(e *entities.UserEvent) *UserEventModel {
	m := &UserEventModel{
		ID:          e.ID,
		TenantID:    e.TenantID,
		UserID:      e.UserID,
		ClientID:    e.ClientID,
		EventType:   string(e.EventType),
		IPAddress:   e.IPAddress,
		CountryCode: e.CountryCode,
		OccurredAt:  e.OccurredAt,
	}
	if e.Details != nil {
		m.Details, _ = json.Marshal(e.Details)
	}
	return m
}

// PostgresUserEventRepository implements UserEventRepository using GORM
type PostgresUserEventRepository struct {
	db *gorm.DB
}

// NewPostgresUserEventRepository creates a new PostgreSQL user event repository
func NewPostgresUserEventRepository(db *gorm.DB) repositories.UserEventRepository {
	return &PostgresUserEventRepository{db: db}
}

func (r *PostgresUserEventRepository) Record(ctx context.Context, event *entities.UserEvent) error {
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now()
	}
	model := fromUserEventEntity(event)
	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return result.Error
	}
	event.ID = model.ID
	return nil
}

func (r *PostgresUserEventRepository) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*entities.UserEvent, int, error) {
	query := r.db.WithContext(ctx).Model(&UserEventModel{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []UserEventModel
	offset := (page - 1) * pageSize
	err := query.Order("occurred_at DESC").Offset(offset).Limit(pageSize).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	events := make([]*entities.UserEvent, len(models))
	for i, m := range models {
		events[i] = m.toEntity()
	}
	return events, int(total), nil
}

func (r *PostgresUserEventRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("occurred_at < ?", before).
		Delete(&UserEventModel{})
	return result.RowsAffected, result.Error
}

// Verify interface compliance
var _ repositories.UserEventRepository = (*PostgresUserEventRepository)(nil)
