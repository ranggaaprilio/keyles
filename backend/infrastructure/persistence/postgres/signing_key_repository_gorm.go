package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// SigningKeyModel is the GORM model for signing_keys table
type SigningKeyModel struct {
	KeyID               string     `gorm:"column:kid;primaryKey"`
	Algorithm           string     `gorm:"column:algorithm;not null"`
	PrivateKeyEncrypted string     `gorm:"column:private_key;not null"`
	PublicKey           string     `gorm:"column:public_key;not null"`
	IsActive            bool       `gorm:"column:is_active;not null;default:true"`
	CreatedAt           time.Time  `gorm:"column:created_at;not null"`
	ExpiresAt           *time.Time `gorm:"column:expires_at"`
}

func (SigningKeyModel) TableName() string {
	return "signing_keys"
}

// toEntity converts GORM model to domain entity
func (m *SigningKeyModel) toEntity() *entities.SigningKey {
	return &entities.SigningKey{
		KeyID:               m.KeyID,
		Algorithm:           m.Algorithm,
		PrivateKeyEncrypted: m.PrivateKeyEncrypted,
		PublicKey:           m.PublicKey,
		IsActive:            m.IsActive,
		CreatedAt:           m.CreatedAt,
		ExpiresAt:           m.ExpiresAt,
	}
}

// PostgresSigningKeyRepositoryGorm implements SigningKeyRepository using GORM
type PostgresSigningKeyRepositoryGorm struct {
	db *gorm.DB
}

// NewPostgresSigningKeyRepositoryGorm creates a new GORM-based signing key repository
func NewPostgresSigningKeyRepositoryGorm(db *gorm.DB) repositories.SigningKeyRepository {
	return &PostgresSigningKeyRepositoryGorm{db: db}
}

// Create stores a new signing key
func (r *PostgresSigningKeyRepositoryGorm) Create(ctx context.Context, key *entities.SigningKey) error {
	model := &SigningKeyModel{
		KeyID:               key.KeyID,
		Algorithm:           key.Algorithm,
		PrivateKeyEncrypted: key.PrivateKeyEncrypted,
		PublicKey:           key.PublicKey,
		IsActive:            key.IsActive,
		CreatedAt:           key.CreatedAt,
		ExpiresAt:           key.ExpiresAt,
	}
	return r.db.WithContext(ctx).Create(model).Error
}

// GetActive retrieves the currently active signing key for a specific algorithm
func (r *PostgresSigningKeyRepositoryGorm) GetActive(ctx context.Context, algorithm string) (*entities.SigningKey, error) {
	var model SigningKeyModel
	result := r.db.WithContext(ctx).
		Where("is_active = ? AND algorithm = ?", true, algorithm).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("created_at DESC").
		First(&model)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return model.toEntity(), nil
}

// GetByKeyID retrieves a signing key by its key ID
func (r *PostgresSigningKeyRepositoryGorm) GetByKeyID(ctx context.Context, keyID string) (*entities.SigningKey, error) {
	var model SigningKeyModel
	result := r.db.WithContext(ctx).
		Where("kid = ?", keyID).
		First(&model)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return model.toEntity(), nil
}

// GetAllActive retrieves all active signing keys (for JWKS endpoint)
func (r *PostgresSigningKeyRepositoryGorm) ListActive(ctx context.Context) ([]*entities.SigningKey, error) {
	var models []SigningKeyModel
	result := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("created_at DESC").
		Find(&models)

	if result.Error != nil {
		return nil, result.Error
	}

	keys := make([]*entities.SigningKey, len(models))
	for i, m := range models {
		keys[i] = m.toEntity()
	}

	return keys, nil
}

// Deactivate deactivates a signing key
func (r *PostgresSigningKeyRepositoryGorm) Deactivate(ctx context.Context, keyID string) error {
	return r.db.WithContext(ctx).Model(&SigningKeyModel{}).
		Where("kid = ?", keyID).
		Update("is_active", false).Error
}

// DeleteExpired deletes expired signing keys
func (r *PostgresSigningKeyRepositoryGorm) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Delete(&SigningKeyModel{})
	return result.RowsAffected, result.Error
}

// Verify interface compliance
var _ repositories.SigningKeyRepository = (*PostgresSigningKeyRepositoryGorm)(nil)
