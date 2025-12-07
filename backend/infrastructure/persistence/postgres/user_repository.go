package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"gorm.io/gorm"
)

// UserModel is the GORM model for users table
type UserModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TenantID     uuid.UUID `gorm:"type:uuid;not null;index"`
	FullName     string    `gorm:"type:varchar(100);not null"`
	Email        string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_users_email"`
	PasswordHash string    `gorm:"type:varchar(255);not null"`
	Role         string    `gorm:"type:varchar(20);not null;default:'admin'"`
	CreatedAt    int64     `gorm:"autoCreateTime:milli"`
	UpdatedAt    int64     `gorm:"autoUpdateTime:milli"`
}

func (UserModel) TableName() string {
	return "users"
}

// PostgresUserRepository implements UserRepository using PostgreSQL
type PostgresUserRepository struct {
	db *gorm.DB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *gorm.DB) repositories.UserRepository {
	return &PostgresUserRepository{db: db}
}

// Create creates a new user
func (r *PostgresUserRepository) Create(ctx context.Context, user *entities.AdminUser) error {
	model := &UserModel{
		ID:           user.ID,
		TenantID:     user.TenantID,
		FullName:     user.FullName,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Role:         string(user.Role),
		CreatedAt:    user.CreatedAt.UnixMilli(),
		UpdatedAt:    user.UpdatedAt.UnixMilli(),
	}

	return r.db.WithContext(ctx).Create(model).Error
}

// FindByID retrieves a user by ID
func (r *PostgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.AdminUser, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.toEntity(&model), nil
}

// FindByEmail retrieves a user by email (case-insensitive)
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*entities.AdminUser, error) {
	var model UserModel
	err := r.db.WithContext(ctx).
		Where("LOWER(email) = LOWER(?)", email).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.toEntity(&model), nil
}

// FindByTenantID retrieves all users for a tenant
func (r *PostgresUserRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entities.AdminUser, error) {
	var models []UserModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	users := make([]*entities.AdminUser, len(models))
	for i, model := range models {
		users[i] = r.toEntity(&model)
	}

	return users, nil
}

// Update updates an existing user
func (r *PostgresUserRepository) Update(ctx context.Context, user *entities.AdminUser) error {
	model := &UserModel{
		ID:           user.ID,
		TenantID:     user.TenantID,
		FullName:     user.FullName,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Role:         string(user.Role),
		UpdatedAt:    user.UpdatedAt.UnixMilli(),
	}

	return r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("id = ?", user.ID).
		Updates(model).Error
}

// EmailExists checks if an email already exists
func (r *PostgresUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("LOWER(email) = LOWER(?)", strings.TrimSpace(email)).
		Count(&count).Error
	
	return count > 0, err
}

// Delete deletes a user
func (r *PostgresUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&UserModel{}).Error
}

// toEntity converts UserModel to domain entity
func (r *PostgresUserRepository) toEntity(model *UserModel) *entities.AdminUser {
	return &entities.AdminUser{
		ID:           model.ID,
		TenantID:     model.TenantID,
		FullName:     model.FullName,
		Email:        model.Email,
		PasswordHash: model.PasswordHash,
		Role:         entities.UserRole(model.Role),
		CreatedAt:    timeFromUnixMilli(model.CreatedAt),
		UpdatedAt:    timeFromUnixMilli(model.UpdatedAt),
	}
}
