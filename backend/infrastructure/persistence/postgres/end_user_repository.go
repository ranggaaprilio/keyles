package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// EndUserModel is the GORM model for end-users in the users table.
// It maps extra columns added by migration 000009.
type EndUserModel struct {
	ID           string     `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()"`
	TenantID     string     `gorm:"column:tenant_id;type:uuid;not null;index"`
	Email        string     `gorm:"column:email;type:varchar(255);not null"`
	FullName     string     `gorm:"column:full_name;type:varchar(100)"`
	DisplayName  string     `gorm:"column:display_name;type:varchar(255)"`
	PasswordHash string     `gorm:"column:password_hash;type:varchar(255);not null"`
	Status       string     `gorm:"column:status;type:varchar(20);not null;default:'pending'"`
	Role         string     `gorm:"column:role;type:varchar(20);not null;default:'user'"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;not null;default:now()"`
}

func (EndUserModel) TableName() string {
	return "users"
}

// toEndUserEntity converts GORM model to domain User entity
func (m *EndUserModel) toEndUserEntity() *entities.User {
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

// fromEndUserEntity converts domain User entity to GORM model
func fromEndUserEntity(e *entities.User) *EndUserModel {
	return &EndUserModel{
		ID:           e.ID,
		TenantID:     e.TenantID,
		Email:        e.Email,
		FullName:     e.DisplayName, // map DisplayName → full_name column as well
		DisplayName:  e.DisplayName,
		PasswordHash: e.PasswordHash,
		Status:       string(e.Status),
		Role:         "user", // end-users always get the 'user' base role
		LastLoginAt:  e.LastLoginAt,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}

// PostgresEndUserRepository implements EndUserRepository using GORM
type PostgresEndUserRepository struct {
	db *gorm.DB
}

// NewPostgresEndUserRepository creates a new PostgreSQL end-user repository
func NewPostgresEndUserRepository(db *gorm.DB) repositories.EndUserRepository {
	return &PostgresEndUserRepository{db: db}
}

func (r *PostgresEndUserRepository) GetByID(ctx context.Context, userID string) (*entities.User, error) {
	var model EndUserModel
	err := r.db.WithContext(ctx).Where("id = ?", userID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return model.toEndUserEntity(), nil
}

func (r *PostgresEndUserRepository) GetByEmail(ctx context.Context, tenantID, email string) (*entities.User, error) {
	var model EndUserModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND LOWER(email) = LOWER(?)", tenantID, strings.TrimSpace(email)).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return model.toEndUserEntity(), nil
}

func (r *PostgresEndUserRepository) Create(ctx context.Context, user *entities.User) error {
	model := fromEndUserEntity(user)
	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return result.Error
	}
	user.ID = model.ID
	return nil
}

func (r *PostgresEndUserRepository) Update(ctx context.Context, user *entities.User) error {
	now := time.Now()
	user.UpdatedAt = now

	result := r.db.WithContext(ctx).
		Model(&EndUserModel{}).
		Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"email":         user.Email,
			"full_name":     user.DisplayName,
			"display_name":  user.DisplayName,
			"password_hash": user.PasswordHash,
			"status":        string(user.Status),
			"updated_at":    now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *PostgresEndUserRepository) ListByTenant(ctx context.Context, tenantID, search string, status entities.UserStatus, page, pageSize int) ([]*entities.User, int, error) {
	query := r.db.WithContext(ctx).Model(&EndUserModel{}).Where("tenant_id = ?", tenantID)

	if status != "" {
		query = query.Where("status = ?", string(status))
	}

	if search = strings.TrimSpace(search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where("(LOWER(display_name) LIKE ? OR LOWER(email) LIKE ?)", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []EndUserModel
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	users := make([]*entities.User, len(models))
	for i, m := range models {
		users[i] = m.toEndUserEntity()
	}
	return users, int(total), nil
}

func (r *PostgresEndUserRepository) CountByTenant(ctx context.Context, tenantID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&EndUserModel{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error
	return int(count), err
}

func (r *PostgresEndUserRepository) UpdateStatus(ctx context.Context, userID string, status entities.UserStatus) error {
	result := r.db.WithContext(ctx).
		Model(&EndUserModel{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"status":     string(status),
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *PostgresEndUserRepository) UpdateLastLogin(ctx context.Context, userID string, at time.Time) error {
	return r.db.WithContext(ctx).
		Model(&EndUserModel{}).
		Where("id = ?", userID).
		Update("last_login_at", at).Error
}

func (r *PostgresEndUserRepository) Delete(ctx context.Context, userID string) error {
	result := r.db.WithContext(ctx).Where("id = ?", userID).Delete(&EndUserModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Verify interface compliance
var _ repositories.EndUserRepository = (*PostgresEndUserRepository)(nil)
