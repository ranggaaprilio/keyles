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

// TenantModel is the GORM model for tenants table
type TenantModel struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	OrganizationName string    `gorm:"type:varchar(100);not null"`
	Status           string    `gorm:"type:varchar(20);not null;default:'pending_verification'"`
	CreatedAt        int64     `gorm:"autoCreateTime:milli"`
	VerifiedAt       *int64    `gorm:"default:null"`
	UpdatedAt        int64     `gorm:"autoUpdateTime:milli"`
}

func (TenantModel) TableName() string {
	return "tenants"
}

// PostgresTenantRepository implements TenantRepository using PostgreSQL
type PostgresTenantRepository struct {
	db *gorm.DB
}

// NewPostgresTenantRepository creates a new PostgreSQL tenant repository
func NewPostgresTenantRepository(db *gorm.DB) repositories.TenantRepository {
	return &PostgresTenantRepository{db: db}
}

// Create creates a new tenant
func (r *PostgresTenantRepository) Create(ctx context.Context, tenant *entities.Tenant) error {
	model := &TenantModel{
		ID:               tenant.ID,
		OrganizationName: tenant.OrganizationName,
		Status:           string(tenant.Status),
		CreatedAt:        tenant.CreatedAt.UnixMilli(),
		UpdatedAt:        tenant.UpdatedAt.UnixMilli(),
	}
	
	if tenant.VerifiedAt != nil {
		verifiedAt := tenant.VerifiedAt.UnixMilli()
		model.VerifiedAt = &verifiedAt
	}

	return r.db.WithContext(ctx).Create(model).Error
}

// FindByID retrieves a tenant by ID
func (r *PostgresTenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Tenant, error) {
	var model TenantModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.toEntity(&model), nil
}

// FindByOrganizationName retrieves a tenant by organization name (case-insensitive)
func (r *PostgresTenantRepository) FindByOrganizationName(ctx context.Context, name string) (*entities.Tenant, error) {
	var model TenantModel
	err := r.db.WithContext(ctx).
		Where("LOWER(organization_name) = LOWER(?)", name).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.toEntity(&model), nil
}

// Update updates an existing tenant
func (r *PostgresTenantRepository) Update(ctx context.Context, tenant *entities.Tenant) error {
	model := &TenantModel{
		ID:               tenant.ID,
		OrganizationName: tenant.OrganizationName,
		Status:           string(tenant.Status),
		UpdatedAt:        tenant.UpdatedAt.UnixMilli(),
	}

	if tenant.VerifiedAt != nil {
		verifiedAt := tenant.VerifiedAt.UnixMilli()
		model.VerifiedAt = &verifiedAt
	}

	return r.db.WithContext(ctx).
		Model(&TenantModel{}).
		Where("id = ?", tenant.ID).
		Updates(model).Error
}

// OrganizationNameExists checks if an organization name already exists
func (r *PostgresTenantRepository) OrganizationNameExists(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&TenantModel{}).
		Where("LOWER(organization_name) = LOWER(?)", strings.TrimSpace(name)).
		Count(&count).Error
	
	return count > 0, err
}

// Delete soft deletes a tenant
func (r *PostgresTenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&TenantModel{}).
		Where("id = ?", id).
		Update("status", "deleted").Error
}

// toEntity converts TenantModel to domain entity
func (r *PostgresTenantRepository) toEntity(model *TenantModel) *entities.Tenant {
	tenant := &entities.Tenant{
		ID:               model.ID,
		OrganizationName: model.OrganizationName,
		Status:           entities.TenantStatus(model.Status),
		CreatedAt:        timeFromUnixMilli(model.CreatedAt),
		UpdatedAt:        timeFromUnixMilli(model.UpdatedAt),
	}

	if model.VerifiedAt != nil {
		verifiedAt := timeFromUnixMilli(*model.VerifiedAt)
		tenant.VerifiedAt = &verifiedAt
	}

	return tenant
}
