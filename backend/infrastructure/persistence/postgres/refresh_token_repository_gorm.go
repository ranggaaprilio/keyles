package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// RefreshTokenModel is the GORM model for refresh_tokens table
type RefreshTokenModel struct {
	ID                  int64      `gorm:"column:id;primaryKey;autoIncrement"`
	Token               string     `gorm:"column:token;not null;uniqueIndex"`
	UserID              string     `gorm:"column:user_id;not null;index"`
	ClientID            string     `gorm:"column:client_id;not null;index"`
	TenantID            string     `gorm:"column:tenant_id;not null;index"`
	Scope               string     `gorm:"column:scope"`
	ExpiresAt           time.Time  `gorm:"column:expires_at;not null;index"`
	CreatedAt           time.Time  `gorm:"column:created_at;not null"`
	LastUsedAt          *time.Time `gorm:"column:last_used_at"`
	IsRevoked           bool       `gorm:"column:is_revoked;not null;default:false"`
	RevokedAt           *time.Time `gorm:"column:revoked_at"`
	RevokedReason       string     `gorm:"column:revoked_reason"`
	FamilyID            string     `gorm:"column:family_id;not null;index"`
	ParentTokenHash     string     `gorm:"column:parent_token_hash"`
	ReplacedByTokenHash string     `gorm:"column:replaced_by_token_hash"`
}

func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}

// toEntity converts GORM model to domain entity
func (m *RefreshTokenModel) toEntity() *entities.RefreshToken {
	return &entities.RefreshToken{
		ID:                  m.ID,
		Token:               m.Token,
		UserID:              m.UserID,
		ClientID:            m.ClientID,
		TenantID:            m.TenantID,
		Scope:               m.Scope,
		ExpiresAt:           m.ExpiresAt,
		CreatedAt:           m.CreatedAt,
		LastUsedAt:          m.LastUsedAt,
		RevokedFlag:         m.IsRevoked,
		RevokedAt:           m.RevokedAt,
		RevokedReason:       m.RevokedReason,
		FamilyID:            m.FamilyID,
		ParentTokenHash:     m.ParentTokenHash,
		ReplacedByTokenHash: m.ReplacedByTokenHash,
	}
}

// fromRefreshTokenEntity converts domain entity to GORM model
func fromRefreshTokenEntity(e *entities.RefreshToken) *RefreshTokenModel {
	return &RefreshTokenModel{
		ID:                  e.ID,
		Token:               e.Token,
		UserID:              e.UserID,
		ClientID:            e.ClientID,
		TenantID:            e.TenantID,
		Scope:               e.Scope,
		ExpiresAt:           e.ExpiresAt,
		CreatedAt:           e.CreatedAt,
		LastUsedAt:          e.LastUsedAt,
		IsRevoked:           e.RevokedFlag,
		RevokedAt:           e.RevokedAt,
		RevokedReason:       e.RevokedReason,
		FamilyID:            e.FamilyID,
		ParentTokenHash:     e.ParentTokenHash,
		ReplacedByTokenHash: e.ReplacedByTokenHash,
	}
}

// PostgresRefreshTokenRepository implements RefreshTokenRepository using GORM
type PostgresRefreshTokenRepository struct {
	db *gorm.DB
}

// NewPostgresRefreshTokenRepository creates a new PostgresRefreshTokenRepository
func NewPostgresRefreshTokenRepository(db *gorm.DB) repositories.RefreshTokenRepository {
	return &PostgresRefreshTokenRepository{db: db}
}

// Create stores a new refresh token
func (r *PostgresRefreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	if token.FamilyID == "" {
		token.FamilyID = token.Token
	}
	model := fromRefreshTokenEntity(token)
	if model.CreatedAt.IsZero() {
		model.CreatedAt = time.Now()
	}

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return result.Error
	}

	token.ID = model.ID
	return nil
}

// Rotate revokes a refresh token and inserts its replacement in one
// transaction. A replay revokes the complete token family before returning.
func (r *PostgresRefreshTokenRepository) Rotate(ctx context.Context, currentTokenHash string, replacement *entities.RefreshToken) error {
	replayed := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current RefreshTokenModel
		result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token = ?", currentTokenHash).
			First(&current)
		if result.Error != nil {
			return result.Error
		}

		if current.IsRevoked || current.ReplacedByTokenHash != "" {
			if err := revokeFamily(tx, current.FamilyID, "refresh token replay detected"); err != nil {
				return err
			}
			replayed = true
			return nil
		}

		now := time.Now()
		if err := tx.Model(&RefreshTokenModel{}).
			Where("token = ? AND is_revoked = ?", currentTokenHash, false).
			Updates(map[string]interface{}{
				"is_revoked":             true,
				"revoked_at":             now,
				"revoked_reason":         "rotated",
				"last_used_at":           now,
				"replaced_by_token_hash": replacement.Token,
			}).Error; err != nil {
			return err
		}

		replacement.FamilyID = current.FamilyID
		replacement.ParentTokenHash = currentTokenHash
		model := fromRefreshTokenEntity(replacement)
		if model.CreatedAt.IsZero() {
			model.CreatedAt = now
		}
		if err := tx.Create(model).Error; err != nil {
			return err
		}
		replacement.ID = model.ID
		return nil
	})
	if err != nil {
		return err
	}
	if replayed {
		return repositories.ErrRefreshTokenReplay
	}
	return nil
}

func (r *PostgresRefreshTokenRepository) RevokeFamily(ctx context.Context, familyID string, reason string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return revokeFamily(tx, familyID, reason)
	})
}

func revokeFamily(tx *gorm.DB, familyID string, reason string) error {
	if familyID == "" {
		return errors.New("refresh token family_id is required")
	}
	return tx.Model(&RefreshTokenModel{}).
		Where("family_id = ? AND is_revoked = ?", familyID, false).
		Updates(map[string]interface{}{
			"is_revoked":     true,
			"revoked_at":     time.Now(),
			"revoked_reason": reason,
		}).Error
}

// GetByToken retrieves a refresh token by its token value (hashed)
func (r *PostgresRefreshTokenRepository) GetByToken(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	var model RefreshTokenModel
	result := r.db.WithContext(ctx).Where("token = ?", tokenHash).First(&model)
	if result.Error != nil {
		return nil, result.Error
	}
	return model.toEntity(), nil
}

// Revoke marks a refresh token as revoked
func (r *PostgresRefreshTokenRepository) Revoke(ctx context.Context, tokenHash string, revokedBy string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&RefreshTokenModel{}).
		Where("token = ?", tokenHash).
		Updates(map[string]interface{}{
			"is_revoked":     true,
			"revoked_at":     now,
			"revoked_reason": "revoked by " + revokedBy,
		})
	return result.Error
}

// RevokeAllForUser revokes all refresh tokens for a user-client combination
func (r *PostgresRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string, clientID string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&RefreshTokenModel{}).
		Where("user_id = ? AND client_id = ? AND is_revoked = ?", userID, clientID, false).
		Updates(map[string]interface{}{
			"is_revoked":     true,
			"revoked_at":     now,
			"revoked_reason": "revoked for user-client pair",
		})
	return result.Error
}

// DeleteExpired removes expired tokens from the database
func (r *PostgresRefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&RefreshTokenModel{})
	return result.RowsAffected, result.Error
}

// IsRevoked checks if a token is revoked
func (r *PostgresRefreshTokenRepository) IsRevoked(ctx context.Context, tokenHash string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&RefreshTokenModel{}).
		Where("token = ? AND is_revoked = ?", tokenHash, true).
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// UpdateLastUsed updates the last_used_at timestamp
func (r *PostgresRefreshTokenRepository) UpdateLastUsed(ctx context.Context, tokenHash string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&RefreshTokenModel{}).
		Where("token = ?", tokenHash).
		Update("last_used_at", now)
	return result.Error
}

// RevokeByClientID revokes all refresh tokens issued to a specific client
func (r *PostgresRefreshTokenRepository) RevokeByClientID(ctx context.Context, clientID string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&RefreshTokenModel{}).
		Where("client_id = ? AND is_revoked = ?", clientID, false).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": now,
		})
	return result.Error
}

// RevokeByUserID revokes all active refresh tokens for a given user across all clients
func (r *PostgresRefreshTokenRepository) RevokeByUserID(ctx context.Context, userID string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&RefreshTokenModel{}).
		Where("user_id = ? AND is_revoked = ?", userID, false).
		Updates(map[string]interface{}{
			"is_revoked":     true,
			"revoked_at":     now,
			"revoked_reason": "user account action",
		})
	return result.Error
}

// ListByUserID returns active, unexpired refresh tokens for a user
func (r *PostgresRefreshTokenRepository) ListByUserID(ctx context.Context, userID string) ([]*entities.RefreshToken, error) {
	var models []RefreshTokenModel
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND is_revoked = ? AND expires_at > ?", userID, false, time.Now()).
		Order("created_at DESC").
		Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}

	tokens := make([]*entities.RefreshToken, 0, len(models))
	for _, m := range models {
		tokens = append(tokens, m.toEntity())
	}
	return tokens, nil
}

// GetByID retrieves a refresh token by its database ID
func (r *PostgresRefreshTokenRepository) GetByID(ctx context.Context, id int64) (*entities.RefreshToken, error) {
	var model RefreshTokenModel
	result := r.db.WithContext(ctx).First(&model, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return model.toEntity(), nil
}

// Verify interface compliance
var _ repositories.RefreshTokenRepository = (*PostgresRefreshTokenRepository)(nil)
var _ repositories.RefreshTokenRotationRepository = (*PostgresRefreshTokenRepository)(nil)
