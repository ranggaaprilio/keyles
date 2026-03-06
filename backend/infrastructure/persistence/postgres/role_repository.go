package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// ErrDuplicateRole is returned when trying to assign a role that is already active
var ErrDuplicateRole = errors.New("duplicate active role assignment")

// UserRoleAssignmentModel is the GORM model for user_role_assignments table
type UserRoleAssignmentModel struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    string     `gorm:"column:user_id;not null;index"`
	ClientID  string     `gorm:"column:client_id;not null;index"`
	TenantID  string     `gorm:"column:tenant_id;not null;index"`
	Role      string     `gorm:"column:role;not null"`
	IsActive  bool       `gorm:"column:is_active;not null;default:true"`
	GrantedAt time.Time  `gorm:"column:granted_at;not null"`
	GrantedBy string     `gorm:"column:granted_by;not null"`
	RevokedAt *time.Time `gorm:"column:revoked_at"`
	RevokedBy *string    `gorm:"column:revoked_by"`
}

func (UserRoleAssignmentModel) TableName() string {
	return "user_role_assignments"
}

// toEntity converts GORM model to domain entity
func (m *UserRoleAssignmentModel) toEntity() *entities.UserRoleAssignment {
	return &entities.UserRoleAssignment{
		ID:        m.ID,
		UserID:    m.UserID,
		ClientID:  m.ClientID,
		TenantID:  m.TenantID,
		Role:      m.Role,
		IsActive:  m.IsActive,
		GrantedAt: m.GrantedAt,
		GrantedBy: m.GrantedBy,
		RevokedAt: m.RevokedAt,
		RevokedBy: m.RevokedBy,
	}
}

// fromEntity converts domain entity to GORM model
func fromUserRoleEntity(e *entities.UserRoleAssignment) *UserRoleAssignmentModel {
	return &UserRoleAssignmentModel{
		ID:        e.ID,
		UserID:    e.UserID,
		ClientID:  e.ClientID,
		TenantID:  e.TenantID,
		Role:      e.Role,
		IsActive:  e.IsActive,
		GrantedAt: e.GrantedAt,
		GrantedBy: e.GrantedBy,
		RevokedAt: e.RevokedAt,
		RevokedBy: e.RevokedBy,
	}
}

// PostgresRoleRepository implements RoleRepository using GORM
type PostgresRoleRepository struct {
	db *gorm.DB
}

// NewPostgresRoleRepository creates a new PostgreSQL role repository
func NewPostgresRoleRepository(db *gorm.DB) repositories.RoleRepository {
	return &PostgresRoleRepository{db: db}
}

func (r *PostgresRoleRepository) AssignRole(ctx context.Context, assignment *entities.UserRoleAssignment) error {
	if assignment.GrantedAt.IsZero() {
		assignment.GrantedAt = time.Now()
	}

	model := fromUserRoleEntity(assignment)

	// Use upsert: update on conflict
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND client_id = ? AND role = ?", model.UserID, model.ClientID, model.Role).
		Assign(map[string]interface{}{
			"is_active":  model.IsActive,
			"granted_at": model.GrantedAt,
			"granted_by": model.GrantedBy,
		}).
		FirstOrCreate(model)

	if result.Error != nil {
		return result.Error
	}

	assignment.ID = model.ID
	return nil
}

func (r *PostgresRoleRepository) RevokeRole(ctx context.Context, userID, clientID, role string) error {
	result := r.db.WithContext(ctx).
		Model(&UserRoleAssignmentModel{}).
		Where("user_id = ? AND client_id = ? AND role = ?", userID, clientID, role).
		Update("is_active", false)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *PostgresRoleRepository) GetUserRoles(ctx context.Context, userID, clientID string) ([]*entities.UserRoleAssignment, error) {
	var models []UserRoleAssignmentModel

	result := r.db.WithContext(ctx).
		Where("user_id = ? AND client_id = ? AND is_active = ?", userID, clientID, true).
		Find(&models)

	if result.Error != nil {
		return nil, result.Error
	}

	assignments := make([]*entities.UserRoleAssignment, len(models))
	for i, m := range models {
		assignments[i] = m.toEntity()
	}

	return assignments, nil
}

func (r *PostgresRoleRepository) HasRole(ctx context.Context, userID, clientID, role string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&UserRoleAssignmentModel{}).
		Where("user_id = ? AND client_id = ? AND role = ? AND is_active = ?", userID, clientID, role, true).
		Count(&count)

	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}

func (r *PostgresRoleRepository) HasAnyRole(ctx context.Context, userID, clientID string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&UserRoleAssignmentModel{}).
		Where("user_id = ? AND client_id = ? AND is_active = ?", userID, clientID, true).
		Count(&count)

	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}

func (r *PostgresRoleRepository) ListRolesByClient(ctx context.Context, clientID string) ([]*entities.UserRoleAssignment, error) {
	var models []UserRoleAssignmentModel

	result := r.db.WithContext(ctx).
		Where("client_id = ? AND is_active = ?", clientID, true).
		Order("granted_at DESC").
		Find(&models)

	if result.Error != nil {
		return nil, result.Error
	}

	assignments := make([]*entities.UserRoleAssignment, len(models))
	for i, m := range models {
		assignments[i] = m.toEntity()
	}

	return assignments, nil
}

func (r *PostgresRoleRepository) ListRolesByUser(ctx context.Context, userID string) ([]*entities.UserRoleAssignment, error) {
	var models []UserRoleAssignmentModel

	result := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Order("granted_at DESC").
		Find(&models)

	if result.Error != nil {
		return nil, result.Error
	}

	assignments := make([]*entities.UserRoleAssignment, len(models))
	for i, m := range models {
		assignments[i] = m.toEntity()
	}

	return assignments, nil
}

// --- New methods for feature 005 ---

// Assign creates a new role assignment. Returns ErrDuplicateRole if the same
// user+client+role combination is already active.
func (r *PostgresRoleRepository) Assign(ctx context.Context, assignment *entities.UserRoleAssignment) error {
	if assignment.GrantedAt.IsZero() {
		assignment.GrantedAt = time.Now()
	}

	// Check for existing active assignment
	var count int64
	err := r.db.WithContext(ctx).
		Model(&UserRoleAssignmentModel{}).
		Where("user_id = ? AND client_id = ? AND role = ? AND is_active = ?",
			assignment.UserID, assignment.ClientID, assignment.Role, true).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrDuplicateRole
	}

	model := fromUserRoleEntity(assignment)
	model.IsActive = true
	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return result.Error
	}
	assignment.ID = model.ID
	return nil
}

// Revoke soft-deletes a role assignment by setting is_active=false and recording revocation info.
func (r *PostgresRoleRepository) Revoke(ctx context.Context, assignmentID int64, revokedByUserID string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&UserRoleAssignmentModel{}).
		Where("id = ? AND is_active = ?", assignmentID, true).
		Updates(map[string]interface{}{
			"is_active":  false,
			"revoked_at": now,
			"revoked_by": revokedByUserID,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("role assignment %d not found or already revoked", assignmentID)
	}
	return nil
}

// ListByUser returns all role assignments for a user (including revoked ones).
func (r *PostgresRoleRepository) ListByUser(ctx context.Context, userID string) ([]*entities.UserRoleAssignment, error) {
	var models []UserRoleAssignmentModel
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("granted_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	assignments := make([]*entities.UserRoleAssignment, len(models))
	for i, m := range models {
		assignments[i] = m.toEntity()
	}
	return assignments, nil
}

// ListByClient returns paginated active role assignments for a client.
func (r *PostgresRoleRepository) ListByClient(ctx context.Context, clientID string, page, pageSize int) ([]*entities.UserRoleAssignment, int, error) {
	query := r.db.WithContext(ctx).
		Model(&UserRoleAssignmentModel{}).
		Where("client_id = ? AND is_active = ?", clientID, true)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []UserRoleAssignmentModel
	offset := (page - 1) * pageSize
	err := query.Order("granted_at DESC").Offset(offset).Limit(pageSize).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	assignments := make([]*entities.UserRoleAssignment, len(models))
	for i, m := range models {
		assignments[i] = m.toEntity()
	}
	return assignments, int(total), nil
}

// RevokeAllForUser revokes all active role assignments for a user.
func (r *PostgresRoleRepository) RevokeAllForUser(ctx context.Context, userID, revokedByUserID string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&UserRoleAssignmentModel{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Updates(map[string]interface{}{
			"is_active":  false,
			"revoked_at": now,
			"revoked_by": revokedByUserID,
		})
	return result.Error
}

// GetActiveRoles returns only active role name strings for a user-client pair.
func (r *PostgresRoleRepository) GetActiveRoles(ctx context.Context, userID, clientID string) ([]string, error) {
	var roles []string
	err := r.db.WithContext(ctx).
		Model(&UserRoleAssignmentModel{}).
		Where("user_id = ? AND client_id = ? AND is_active = ?", userID, clientID, true).
		Pluck("role", &roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}
