package role

import (
"context"
"errors"
"fmt"
"time"

"github.com/ranggaaprilio/keyles/domain/entities"
"github.com/ranggaaprilio/keyles/domain/repositories"
)

// RevokeRoleRequest represents a request to revoke a role from a user
type RevokeRoleRequest struct {
	AssignmentID int64
	UserID       string
	ClientID     string
	TenantID     string
	RevokedBy    string
}

// RevokeRole handles role revocation from users (FR-006b)
type RevokeRole struct {
	roleRepo         repositories.RoleRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	eventRepo        repositories.UserEventRepository
	auditRepo        repositories.AuditRepository
}

// NewRevokeRole creates a new RevokeRole use case
func NewRevokeRole(
roleRepo repositories.RoleRepository,
refreshTokenRepo repositories.RefreshTokenRepository,
eventRepo repositories.UserEventRepository,
auditRepo repositories.AuditRepository,
) *RevokeRole {
	return &RevokeRole{
		roleRepo:         roleRepo,
		refreshTokenRepo: refreshTokenRepo,
		eventRepo:        eventRepo,
		auditRepo:        auditRepo,
	}
}

// Execute revokes a role from a user for a client
func (uc *RevokeRole) Execute(ctx context.Context, req RevokeRoleRequest) error {
	// Validate required fields
	if req.AssignmentID == 0 {
		return errors.New("assignment_id is required")
	}
	if req.UserID == "" {
		return errors.New("user_id is required")
	}
	if req.RevokedBy == "" {
		return errors.New("revoked_by is required")
	}

	// Revoke the role (soft delete with metadata)
	if err := uc.roleRepo.Revoke(ctx, req.AssignmentID, req.RevokedBy); err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	// Cascade revocation of refresh tokens if clientID is specified (FR-006e)
	if req.ClientID != "" {
		_ = uc.refreshTokenRepo.RevokeAllForUser(ctx, req.UserID, req.ClientID)
	}

	// Record role_revoked event
	event := &entities.UserEvent{
		TenantID:   req.TenantID,
		UserID:     req.UserID,
		EventType:  entities.EventTypeRoleRevoked,
		Details:    map[string]any{"assignment_id": req.AssignmentID, "revoked_by": req.RevokedBy},
		OccurredAt: time.Now(),
	}
	_ = uc.eventRepo.Record(ctx, event)

	// Audit log
	auditLog := entities.NewAuditLog("role_revoked", "", "")
	auditLog.WithData("assignment_id", req.AssignmentID)
	auditLog.WithData("user_id", req.UserID)
	auditLog.WithData("revoked_by", req.RevokedBy)
	_ = uc.auditRepo.Create(ctx, auditLog)

	return nil
}
