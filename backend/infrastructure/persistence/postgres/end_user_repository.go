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

// PostgresUser is the GORM model for users table (end-user view)
type PostgresUser struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TenantID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	Email        string     `gorm:"type:varchar(255);not null"`
	DisplayName  *string    `gorm:"type:varchar(255)"`
	PasswordHash string     `gorm:"type:varchar(255);not null"`
	Status       string     `gorm:"type:varchar(20);not null;default:'active'"`
	LastLoginAt  *time.Time `gorm:"type:timestamp"`
	IsActive     bool       `gorm:"type:bool;not null;default:true"`
	CreatedAt    time.Time  `gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt    time.Time  `gorm:"type:timestamp;not null;default:now()"`
}

func (PostgresUser) TableName() string {
	return "users"
}

// PostgresEndUserRepository implements EndUserRepository using PostgreSQL
type PostgresEndUserRepository struct {
	db *gorm.DB
}

// NewPostgresEndUserRepository creates a new PostgreSQL end-user repository
func NewPostgresEndUserRepository(db *gorm.DB) repositories.EndUserRepository {
	return &PostgresEndUserRepository{db: db}
}

// Create creates a new end user
func (r *PostgresEndUserRepository) Create(ctx context.Context, user *entities.User) error {
	model := fromUserEntity(user)
	return r.db.WithContext(ctx).Create(model).Error
}

// GetByID retrieves an end user by ID
func (r *PostgresEndUserRepository) GetByID(ctx context.Context, userID uuid.UUID) (*entities.User, error) {
	var model PostgresUser
	err := r.db.WithContext(ctx).Where("id = ?", userID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return model.toEntity(), nil
}

// GetByEmail retrieves an end user by email within a tenant (case-insensitive)
func (r *PostgresEndUserRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entities.User, error) {
	var model PostgresUser
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND LOWER(email) = LOWER(?)", tenantID, email).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return model.toEntity(), nil
}

// ListByTenant retrieves paginated end users for a tenant with optional filtering
func (r *PostgresEndUserRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, search string, status entities.UserStatus, page, pageSize int) ([]*entities.User, int, error) {
	var models []PostgresUser
	var total int64

	query := r.db.WithContext(ctx).Model(&PostgresUser{}).Where("tenant_id = ?", tenantID)

	if search != "" {
		query = query.Where("display_name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if status != "" {
		query = query.Where("status = ?", string(status))
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	users := make([]*entities.User, len(models))
	for i, model := range models {
		users[i] = model.toEntity()
	}

	return users, int(total), nil
}

// CountByTenant returns the total number of end users in a tenant
func (r *PostgresEndUserRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&PostgresUser{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return int(count), err
}

// UpdateStatus updates the status of an end user
func (r *PostgresEndUserRepository) UpdateStatus(ctx context.Context, userID uuid.UUID, status entities.UserStatus) error {
	return r.db.WithContext(ctx).Model(&PostgresUser{}).
		Where("id = ?", userID).
		Update("status", string(status)).Error
}

// UpdateLastLogin updates the last login timestamp for an end user
func (r *PostgresEndUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, at time.Time) error {
	return r.db.WithContext(ctx).Model(&PostgresUser{}).
		Where("id = ?", userID).
		Update("last_login_at", at).Error
}

// Update updates an existing end user
func (r *PostgresEndUserRepository) Update(ctx context.Context, user *entities.User) error {
	model := fromUserEntity(user)
	return r.db.WithContext(ctx).Model(&PostgresUser{}).
		Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"display_name": model.DisplayName,
			"updated_at":   model.UpdatedAt,
		}).Error
}

// Delete removes an end user
func (r *PostgresEndUserRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", userID).
		Delete(&PostgresUser{}).Error
}

// toEntity converts PostgresUser to domain entity
func (m *PostgresUser) toEntity() *entities.User {
	return &entities.User{
		ID:           m.ID,
		TenantID:     m.TenantID,
		Email:        m.Email,
		DisplayName:  m.DisplayName,
		PasswordHash: m.PasswordHash,
		Status:       entities.UserStatus(m.Status),
		LastLoginAt:  m.LastLoginAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// fromUserEntity converts domain entity to PostgresUser
func fromUserEntity(user *entities.User) *PostgresUser {
	return &PostgresUser{
		ID:           user.ID,
		TenantID:     user.TenantID,
		Email:        user.Email,
		DisplayName:  user.DisplayName,
		PasswordHash: user.PasswordHash,
		Status:       string(user.Status),
		LastLoginAt:  user.LastLoginAt,
		IsActive:     user.Status != entities.UserStatusDisabled,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

// Verify interface compliance
var _ repositories.EndUserRepository = (*PostgresEndUserRepository)(nil)
