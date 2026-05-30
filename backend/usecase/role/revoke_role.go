package role

import (
	"context"
	"errors"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// RevokeRole handles role revocation from users (FR-006b)
type RevokeRole struct {
	roleRepo         repositories.RoleRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	eventRepo        repositories.UserEventRepository
}

// NewRevokeRole creates a new RevokeRole use case
func NewRevokeRole(
	roleRepo repositories.RoleRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
	eventRepo repositories.UserEventRepository,
) *RevokeRole {
	return &RevokeRole{
		roleRepo:         roleRepo,
		refreshTokenRepo: refreshTokenRepo,
		eventRepo:        eventRepo,
	}
}

// Execute revokes a role assignment by its ID
func (uc *RevokeRole) Execute(ctx context.Context, assignmentID int64, revokedBy string) error {
	if assignmentID <= 0 {
		return errors.New("assignment_id is required")
	}

	if err := uc.roleRepo.Revoke(ctx, assignmentID, revokedBy); err != nil {
		return errors.New("failed to revoke role: " + err.Error())
	}

	event := &entities.UserEvent{
		UserID:     "",
		EventType:  entities.EventTypeRoleRevoked,
		Details:    map[string]any{"assignment_id": assignmentID, "revoked_by": revokedBy},
		OccurredAt: time.Now(),
	}
	if err := uc.eventRepo.Record(ctx, event); err != nil {
		return errors.New("failed to record role revoked event: " + err.Error())
	}

	return nil
}
