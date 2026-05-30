package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

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

// Assign inserts or upserts a role assignment with revoked_at=NULL, is_active=true
func (r *PostgresRoleRepository) Assign(ctx context.Context, assignment *entities.UserRoleAssignment) error {
	if assignment.GrantedAt.IsZero() {
		assignment.GrantedAt = time.Now()
	}

	model := fromUserRoleEntity(assignment)
	model.IsActive = true
	model.RevokedAt = nil
	model.RevokedBy = nil

	result := r.db.WithContext(ctx).
		Where("user_id = ? AND client_id = ? AND role = ?", model.UserID, model.ClientID, model.Role).
		Assign(map[string]interface{}{
			"is_active":  model.IsActive,
			"granted_at": model.GrantedAt,
			"granted_by": model.GrantedBy,
			"revoked_at": nil,
			"revoked_by": nil,
		}).
		FirstOrCreate(model)

	if result.Error != nil {
		return result.Error
	}

	assignment.ID = model.ID
	return nil
}

// Revoke updates a specific role assignment to set is_active=false, revoked_at=NOW(), revoked_by=revokedByUserID
func (r *PostgresRoleRepository) Revoke(ctx context.Context, assignmentID int64, revokedByUserID string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&UserRoleAssignmentModel{}).
		Where("id = ?", assignmentID).
		Updates(map[string]interface{}{
			"is_active":  false,
			"revoked_at": now,
			"revoked_by": revokedByUserID,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// ListByUser retrieves all role assignments for a user
func (r *PostgresRoleRepository) ListByUser(ctx context.Context, userID string) ([]*entities.UserRoleAssignment, error) {
	var models []UserRoleAssignmentModel

	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
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

// ListByClient retrieves paginated role assignments for a client
func (r *PostgresRoleRepository) ListByClient(ctx context.Context, clientID string, page, pageSize int) ([]*entities.UserRoleAssignment, int, error) {
	var models []UserRoleAssignmentModel
	var total int64

	query := r.db.WithContext(ctx).Model(&UserRoleAssignmentModel{}).Where("client_id = ?", clientID)

	result := query.Count(&total)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	offset := (page - 1) * pageSize
	result = query.Order("granted_at DESC").Offset(offset).Limit(pageSize).Find(&models)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	assignments := make([]*entities.UserRoleAssignment, len(models))
	for i, m := range models {
		assignments[i] = m.toEntity()
	}

	return assignments, int(total), nil
}

// RevokeAllForUser revokes all active role assignments for a user
func (r *PostgresRoleRepository) RevokeAllForUser(ctx context.Context, userID, revokedByUserID string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&UserRoleAssignmentModel{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Updates(map[string]interface{}{
			"is_active":  false,
			"revoked_at": now,
			"revoked_by": revokedByUserID,
		})
	return result.Error
}

// GetActiveRoles retrieves active role names for a user in a client
func (r *PostgresRoleRepository) GetActiveRoles(ctx context.Context, userID, clientID string) ([]string, error) {
	var roles []string
	result := r.db.WithContext(ctx).Model(&UserRoleAssignmentModel{}).
		Where("user_id = ? AND client_id = ? AND is_active = ?", userID, clientID, true).
		Pluck("role", &roles)
	if result.Error != nil {
		return nil, result.Error
	}
	return roles, nil
}
