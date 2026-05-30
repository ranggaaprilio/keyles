package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"gorm.io/gorm"
)

// PostgresInvitation is the GORM model for invitations table
type PostgresInvitation struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	Email       string     `gorm:"type:varchar(255);not null"`
	DisplayName *string    `gorm:"type:varchar(255)"`
	TokenHash   string     `gorm:"type:varchar(255);not null;index"`
	Status      string     `gorm:"type:varchar(20);not null;default:'pending'"`
	InvitedBy   uuid.UUID  `gorm:"type:uuid;not null"`
	ExpiresAt   time.Time  `gorm:"type:timestamp;not null"`
	AcceptedAt  *time.Time `gorm:"type:timestamp"`
	CreatedAt   time.Time  `gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt   time.Time  `gorm:"type:timestamp;not null;default:now()"`
}

func (PostgresInvitation) TableName() string {
	return "invitations"
}

// PostgresInvitationRepository implements InvitationRepository using PostgreSQL
type PostgresInvitationRepository struct {
	db *gorm.DB
}

// NewPostgresInvitationRepository creates a new PostgreSQL invitation repository
func NewPostgresInvitationRepository(db *gorm.DB) repositories.InvitationRepository {
	return &PostgresInvitationRepository{db: db}
}

// Create stores a new invitation
func (r *PostgresInvitationRepository) Create(ctx context.Context, inv *entities.Invitation) error {
	model := fromInvitationEntity(inv)
	return r.db.WithContext(ctx).Create(model).Error
}

// GetByToken retrieves an invitation by its token hash
func (r *PostgresInvitationRepository) GetByToken(ctx context.Context, plainToken string) (*entities.Invitation, error) {
	var model PostgresInvitation
	err := r.db.WithContext(ctx).Where("token_hash = ?", plainToken).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return model.toEntity(), nil
}

// GetPendingByEmail retrieves a pending invitation by email within a tenant
func (r *PostgresInvitationRepository) GetPendingByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entities.Invitation, error) {
	var model PostgresInvitation
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

// UpdateStatus updates the status of an invitation
func (r *PostgresInvitationRepository) UpdateStatus(ctx context.Context, invitationID uuid.UUID, status entities.InvitationStatus, acceptedAt *time.Time) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if acceptedAt != nil {
		updates["accepted_at"] = *acceptedAt
	}
	return r.db.WithContext(ctx).Model(&PostgresInvitation{}).
		Where("id = ?", invitationID).
		Updates(updates).Error
}

// ListByTenant retrieves paginated invitations for a tenant
func (r *PostgresInvitationRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*entities.Invitation, int, error) {
	var models []PostgresInvitation
	var total int64

	query := r.db.WithContext(ctx).Model(&PostgresInvitation{}).Where("tenant_id = ?", tenantID)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	invitations := make([]*entities.Invitation, len(models))
	for i, model := range models {
		invitations[i] = model.toEntity()
	}

	return invitations, int(total), nil
}

// ExpireStalePending expires invitations that have been pending too long
func (r *PostgresInvitationRepository) ExpireStalePending(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).Model(&PostgresInvitation{}).
		Where("status = ? AND expires_at < ?", string(entities.InvitationStatusPending), time.Now()).
		Update("status", string(entities.InvitationStatusExpired))
	return result.RowsAffected, result.Error
}

// toEntity converts PostgresInvitation to domain entity
func (m *PostgresInvitation) toEntity() *entities.Invitation {
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

// fromInvitationEntity converts domain entity to PostgresInvitation
func fromInvitationEntity(inv *entities.Invitation) *PostgresInvitation {
	return &PostgresInvitation{
		ID:          inv.ID,
		TenantID:    inv.TenantID,
		Email:       inv.Email,
		DisplayName: inv.DisplayName,
		TokenHash:   inv.TokenHash,
		Status:      string(inv.Status),
		InvitedBy:   inv.InvitedBy,
		ExpiresAt:   inv.ExpiresAt,
		AcceptedAt:  inv.AcceptedAt,
		CreatedAt:   inv.CreatedAt,
		UpdatedAt:   inv.UpdatedAt,
	}
}

// Verify interface compliance
var _ repositories.InvitationRepository = (*PostgresInvitationRepository)(nil)
