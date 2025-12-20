package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"gorm.io/gorm"
)

// AuditLogModel is the GORM model for audit_logs table
type AuditLogModel struct {
	ID        uuid.UUID              `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TenantID  *uuid.UUID             `gorm:"type:uuid;index"`
	UserID    *uuid.UUID             `gorm:"type:uuid;index"`
	EventType string                 `gorm:"type:varchar(50);not null;index"`
	EventData map[string]interface{} `gorm:"type:jsonb"`
	IPAddress string                 `gorm:"type:varchar(45)"`
	UserAgent string                 `gorm:"type:text"`
	CreatedAt time.Time              `gorm:"type:timestamp;not null;default:now();index"`
}

func (AuditLogModel) TableName() string {
	return "audit_logs"
}

// PostgresAuditRepository implements AuditRepository using PostgreSQL
type PostgresAuditRepository struct {
	db *gorm.DB
}

// NewPostgresAuditRepository creates a new PostgreSQL audit repository
func NewPostgresAuditRepository(db *gorm.DB) repositories.AuditRepository {
	return &PostgresAuditRepository{db: db}
}

// Create creates a new audit log entry
func (r *PostgresAuditRepository) Create(ctx context.Context, log *entities.AuditLog) error {
	model := &AuditLogModel{
		ID:        log.ID,
		TenantID:  log.TenantID,
		UserID:    log.UserID,
		EventType: string(log.EventType),
		EventData: log.EventData,
		IPAddress: log.IPAddress,
		UserAgent: log.UserAgent,
		CreatedAt: log.CreatedAt,
	}

	return r.db.WithContext(ctx).Create(model).Error
}

// FindByTenantID retrieves audit logs for a tenant
func (r *PostgresAuditRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entities.AuditLog, error) {
	var models []AuditLogModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	return r.toEntities(models), nil
}

// FindByEventType retrieves audit logs by event type
func (r *PostgresAuditRepository) FindByEventType(ctx context.Context, eventType entities.EventType, limit, offset int) ([]*entities.AuditLog, error) {
	var models []AuditLogModel
	err := r.db.WithContext(ctx).
		Where("event_type = ?", string(eventType)).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	return r.toEntities(models), nil
}

// FindRecent retrieves the most recent audit logs
func (r *PostgresAuditRepository) FindRecent(ctx context.Context, limit int) ([]*entities.AuditLog, error) {
	var models []AuditLogModel
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	return r.toEntities(models), nil
}

// toEntities converts slice of AuditLogModel to slice of entities
func (r *PostgresAuditRepository) toEntities(models []AuditLogModel) []*entities.AuditLog {
	logs := make([]*entities.AuditLog, len(models))
	for i, model := range models {
		logs[i] = r.toEntity(&model)
	}
	return logs
}

// toEntity converts AuditLogModel to domain entity
func (r *PostgresAuditRepository) toEntity(model *AuditLogModel) *entities.AuditLog {
	return &entities.AuditLog{
		ID:        model.ID,
		TenantID:  model.TenantID,
		UserID:    model.UserID,
		EventType: entities.EventType(model.EventType),
		EventData: model.EventData,
		IPAddress: model.IPAddress,
		UserAgent: model.UserAgent,
		CreatedAt: model.CreatedAt,
	}
}
