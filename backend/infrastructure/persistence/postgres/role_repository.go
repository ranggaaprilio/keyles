package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

type roleRepository struct {
	pool *pgxpool.Pool
}

func NewRoleRepository(pool *pgxpool.Pool) repositories.RoleRepository {
	return &roleRepository{pool: pool}
}

func (r *roleRepository) AssignRole(ctx context.Context, assignment *entities.UserRoleAssignment) error {
	query := `
		INSERT INTO user_role_assignments (user_id, client_id, tenant_id, role, is_active, granted_at, granted_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, client_id, role)
		DO UPDATE SET is_active = $5, granted_at = $6, granted_by = $7
		RETURNING id
	`

	if assignment.GrantedAt.IsZero() {
		assignment.GrantedAt = time.Now()
	}

	err := r.pool.QueryRow(
ctx,
query,
assignment.UserID,
assignment.ClientID,
assignment.TenantID,
assignment.Role,
assignment.IsActive,
assignment.GrantedAt,
assignment.GrantedBy,
).Scan(&assignment.ID)

	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

func (r *roleRepository) RevokeRole(ctx context.Context, userID, clientID, role string) error {
	query := `
		UPDATE user_role_assignments
		SET is_active = false
		WHERE user_id = $1 AND client_id = $2 AND role = $3
	`

	commandTag, err := r.pool.Exec(ctx, query, userID, clientID, role)
	if err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("role assignment not found")
	}

	return nil
}

func (r *roleRepository) GetUserRoles(ctx context.Context, userID, clientID string) ([]*entities.UserRoleAssignment, error) {
	query := `
		SELECT id, user_id, client_id, tenant_id, role, is_active, granted_at, granted_by
		FROM user_role_assignments
		WHERE user_id = $1 AND client_id = $2 AND is_active = true
	`

	rows, err := r.pool.Query(ctx, query, userID, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()

	var assignments []*entities.UserRoleAssignment
	for rows.Next() {
		var a entities.UserRoleAssignment
		err := rows.Scan(
&a.ID,
			&a.UserID,
			&a.ClientID,
			&a.TenantID,
			&a.Role,
			&a.IsActive,
			&a.GrantedAt,
			&a.GrantedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role assignment: %w", err)
		}
		assignments = append(assignments, &a)
	}

	return assignments, nil
}

func (r *roleRepository) HasRole(ctx context.Context, userID, clientID, role string) (bool, error) {
	query := `
		SELECT EXISTS(
SELECT 1 FROM user_role_assignments
WHERE user_id = $1 AND client_id = $2 AND role = $3 AND is_active = true
)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, clientID, role).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check role: %w", err)
	}

	return exists, nil
}

func (r *roleRepository) ListRolesByClient(ctx context.Context, clientID string) ([]*entities.UserRoleAssignment, error) {
	query := `
		SELECT id, user_id, client_id, tenant_id, role, is_active, granted_at, granted_by
		FROM user_role_assignments
		WHERE client_id = $1 AND is_active = true
		ORDER BY granted_at DESC
	`

	rows, err := r.pool.Query(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles by client: %w", err)
	}
	defer rows.Close()

	var assignments []*entities.UserRoleAssignment
	for rows.Next() {
		var a entities.UserRoleAssignment
		err := rows.Scan(
&a.ID,
			&a.UserID,
			&a.ClientID,
			&a.TenantID,
			&a.Role,
			&a.IsActive,
			&a.GrantedAt,
			&a.GrantedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role assignment: %w", err)
		}
		assignments = append(assignments, &a)
	}

	return assignments, nil
}

func (r *roleRepository) ListRolesByUser(ctx context.Context, userID string) ([]*entities.UserRoleAssignment, error) {
	query := `
		SELECT id, user_id, client_id, tenant_id, role, is_active, granted_at, granted_by
		FROM user_role_assignments
		WHERE user_id = $1 AND is_active = true
		ORDER BY granted_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles by user: %w", err)
	}
	defer rows.Close()

	var assignments []*entities.UserRoleAssignment
	for rows.Next() {
		var a entities.UserRoleAssignment
		err := rows.Scan(
&a.ID,
			&a.UserID,
			&a.ClientID,
			&a.TenantID,
			&a.Role,
			&a.IsActive,
			&a.GrantedAt,
			&a.GrantedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role assignment: %w", err)
		}
		assignments = append(assignments, &a)
	}

	return assignments, nil
}
