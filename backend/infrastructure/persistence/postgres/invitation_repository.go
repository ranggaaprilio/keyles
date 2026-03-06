package postgres

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// InvitationModel is the GORM model for the invitations table
type InvitationModel struct {
	ID          string     `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()"`
	TenantID    string     `gorm:"column:tenant_id;type:uuid;not null;index"`
	Email       string     `gorm:"column:email;type:varchar(255);not null"`
	DisplayName string     `gorm:"column:display_name;type:varchar(255)"`
	TokenHash   string     `gorm:"column:token_hash;type:varchar(255);not null"`
	Status      string     `gorm:"column:status;type:varchar(20);not null;default:'pending'"`
	InvitedBy   string     `gorm:"column:invited_by;type:uuid;not null"`
	ExpiresAt   time.Time  `gorm:"column:expires_at;not null"`
	AcceptedAt  *time.Time `gorm:"column:accepted_at"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null;default:now()"`
}

func (InvitationModel) TableName() string {
	return "invitations"
}

func (m *InvitationModel) toEntity() *entities.Invitation {
	return &entities.Invitation{
		ID:          m.ID,
		TenantID:    m.TenantID,
		Email:       m.Email,
		DisplayName: m.DisplayName,
		TokenHash:   m.TokenHash,
		Status:      entities.InvitationStatus(m.Status),
		InvitedBy:   m.InvitedBy,
		ExpiresAt:   m.ExpiresAt,
		AcceptedAt:  m.AcceptedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func fromInvitationEntity(e *entities.Invitation) *InvitationModel {
	return &InvitationModel{
		ID:          e.ID,
		TenantID:    e.TenantID,
		Email:       e.Email,
		DisplayName: e.DisplayName,
		TokenHash:   e.TokenHash,
		Status:      string(e.Status),
		InvitedBy:   e.InvitedBy,
		ExpiresAt:   e.ExpiresAt,
		AcceptedAt:  e.AcceptedAt,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

// PostgresInvitationRepository implements InvitationRepository using GORM
type PostgresInvitationRepository struct {
	db *gorm.DB
}

// NewPostgresInvitationRepository creates a new PostgreSQL invitation repository
func NewPostgresInvitationRepository(db *gorm.DB) repositories.InvitationRepository {
	return &PostgresInvitationRepository{db: db}
}

func (r *PostgresInvitationRepository) Create(ctx context.Context, inv *entities.Invitation) error {
	model := fromInvitationEntity(inv)
	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return result.Error
	}
	inv.ID = model.ID
	return nil
}

// GetByToken finds a pending invitation by comparing the plain token against stored bcrypt hashes.
func (r *PostgresInvitationRepository) GetByToken(ctx context.Context, plainToken string) (*entities.Invitation, error) {
	var models []InvitationModel
	err := r.db.WithContext(ctx).
		Where("status = ? AND expires_at > ?", string(entities.InvitationStatusPending), time.Now()).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	for _, m := range models {
		if bcrypt.CompareHashAndPassword([]byte(m.TokenHash), []byte(plainToken)) == nil {
			return m.toEntity(), nil
		}
	}

	return nil, gorm.ErrRecordNotFound
}

func (r *PostgresInvitationRepository) GetPendingByEmail(ctx context.Context, tenantID, email string) (*entities.Invitation, error) {
	var model InvitationModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND LOWER(email) = LOWER(?) AND status = ?", tenantID, email, string(entities.InvitationStatusPending)).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return model.toEntity(), nil
}

func (r *PostgresInvitationRepository) UpdateStatus(ctx context.Context, invitationID string, status entities.InvitationStatus, acceptedAt *time.Time) error {
	updates := map[string]interface{}{
		"status":     string(status),
		"updated_at": time.Now(),
	}
	if acceptedAt != nil {
		updates["accepted_at"] = acceptedAt
	}

	result := r.db.WithContext(ctx).
		Model(&InvitationModel{}).
		Where("id = ?", invitationID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *PostgresInvitationRepository) ListByTenant(ctx context.Context, tenantID string, page, pageSize int) ([]*entities.Invitation, int, error) {
	query := r.db.WithContext(ctx).Model(&InvitationModel{}).Where("tenant_id = ?", tenantID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []InvitationModel
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	invitations := make([]*entities.Invitation, len(models))
	for i, m := range models {
		invitations[i] = m.toEntity()
	}
	return invitations, int(total), nil
}

func (r *PostgresInvitationRepository) ExpireStalePending(ctx context.Context) (int64, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&InvitationModel{}).
		Where("status = ? AND expires_at <= ?", string(entities.InvitationStatusPending), now).
		Updates(map[string]interface{}{
			"status":     string(entities.InvitationStatusExpired),
			"updated_at": now,
		})
	return result.RowsAffected, result.Error
}

// Verify interface compliance
var _ repositories.InvitationRepository = (*PostgresInvitationRepository)(nil)
