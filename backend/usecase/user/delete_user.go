package user

import (
"context"
"errors"
"fmt"
"time"

"github.com/ranggaaprilio/keyles/domain/entities"
"github.com/ranggaaprilio/keyles/domain/repositories"
"github.com/ranggaaprilio/keyles/domain/services"
)

var (
ErrCannotDeleteSelf = errors.New("cannot delete your own account")
)

// DeleteUserInput represents the request to delete a user
type DeleteUserInput struct {
	TargetUserID string
	AdminUserID  string
	TenantID     string
}

// DeleteUser handles permanent user deletion with cascade
type DeleteUser struct {
	endUserRepo      repositories.EndUserRepository
	roleRepo         repositories.RoleRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	userBlacklist    services.UserBlacklist
	auditRepo        repositories.AuditRepository
}

// NewDeleteUser creates a new DeleteUser use case
func NewDeleteUser(
endUserRepo repositories.EndUserRepository,
roleRepo repositories.RoleRepository,
refreshTokenRepo repositories.RefreshTokenRepository,
userBlacklist services.UserBlacklist,
auditRepo repositories.AuditRepository,
) *DeleteUser {
	return &DeleteUser{
		endUserRepo:      endUserRepo,
		roleRepo:         roleRepo,
		refreshTokenRepo: refreshTokenRepo,
		userBlacklist:    userBlacklist,
		auditRepo:        auditRepo,
	}
}

// Execute deletes a user with full cascade
func (uc *DeleteUser) Execute(ctx context.Context, input DeleteUserInput) error {
	if input.TargetUserID == "" {
		return errors.New("target_user_id is required")
	}
	if input.AdminUserID == "" {
		return errors.New("admin_user_id is required")
	}
	if input.TenantID == "" {
		return errors.New("tenant_id is required")
	}

	// Cannot delete self
	if input.TargetUserID == input.AdminUserID {
		return ErrCannotDeleteSelf
	}

	// Verify user exists and belongs to tenant
	user, err := uc.endUserRepo.GetByID(ctx, input.TargetUserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if user.TenantID != input.TenantID {
		return errors.New("user not found")
	}

	// Cascade in order:
	// 1. Revoke all role assignments
	_ = uc.roleRepo.RevokeAllForUser(ctx, input.TargetUserID, input.AdminUserID)

	// 2. Revoke all refresh tokens
	_ = uc.refreshTokenRepo.RevokeByUserID(ctx, input.TargetUserID)

	// 3. Add to blacklist (15 min TTL)
	_ = uc.userBlacklist.Add(ctx, input.TargetUserID, 900*time.Second)

	// 4. Hard delete user (DB cascades handle invitations)
	if err := uc.endUserRepo.Delete(ctx, input.TargetUserID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Write audit log entry with deleted user's metadata
auditLog := entities.NewAuditLog("user_deleted", "", "")
auditLog.WithData("user_id", input.TargetUserID)
auditLog.WithData("email", user.Email)
auditLog.WithData("deleted_by", input.AdminUserID)
_ = uc.auditRepo.Create(ctx, auditLog)

return nil
}
