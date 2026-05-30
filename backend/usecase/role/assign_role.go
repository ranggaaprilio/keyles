package role

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// AssignRoleRequest represents a request to assign a role to a user
type AssignRoleRequest struct {
	UserID    string
	ClientID  string
	TenantID  string
	Role      string
	GrantedBy string
}

// AssignRole handles role assignment to users (FR-006a)
type AssignRole struct {
	roleRepo   repositories.RoleRepository
	userRepo   repositories.UserRepository
	clientRepo repositories.ClientRepository
	eventRepo  repositories.UserEventRepository
}

// NewAssignRole creates a new AssignRole use case
func NewAssignRole(
	roleRepo repositories.RoleRepository,
	userRepo repositories.UserRepository,
	clientRepo repositories.ClientRepository,
	eventRepo repositories.UserEventRepository,
) *AssignRole {
	return &AssignRole{
		roleRepo:   roleRepo,
		userRepo:   userRepo,
		clientRepo: clientRepo,
		eventRepo:  eventRepo,
	}
}

// Execute assigns a role to a user for a client
func (uc *AssignRole) Execute(ctx context.Context, req AssignRoleRequest) error {
	// Validate required fields
	if req.UserID == "" {
		return errors.New("user_id is required")
	}
	if req.ClientID == "" {
		return errors.New("client_id is required")
	}
	if req.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if req.Role == "" {
		return errors.New("role is required")
	}

	// Validate role name length (free-form, 1–100 chars per FR-015)
	if len(req.Role) > entities.MaxRoleNameLength {
		return errors.New("role must be at most 100 characters")
	}

	// Parse user ID as UUID
	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return errors.New("invalid user_id format")
	}

	// Verify user exists
	user, err := uc.userRepo.FindByID(ctx, userUUID)
	if err != nil {
		return errors.New("user not found: " + err.Error())
	}

	// Verify client exists
	client, err := uc.clientRepo.GetByID(ctx, req.ClientID)
	if err != nil {
		return errors.New("client not found: " + err.Error())
	}

	// Verify user and client belong to the same tenant
	if user.TenantID.String() != client.TenantID {
		return errors.New("user and client must belong to the same tenant")
	}

	// Verify the tenant ID matches
	if user.TenantID.String() != req.TenantID {
		return errors.New("tenant mismatch")
	}

	// Check for duplicate role assignment
	hasRole, err := uc.roleRepo.HasRole(ctx, req.UserID, req.ClientID, req.Role)
	if err != nil {
		return errors.New("failed to check existing role: " + err.Error())
	}
	if hasRole {
		return errors.New("user already has role: " + req.Role)
	}

	// Create role assignment
	assignment := &entities.UserRoleAssignment{
		UserID:    req.UserID,
		ClientID:  req.ClientID,
		TenantID:  req.TenantID,
		Role:      req.Role,
		IsActive:  true,
		GrantedAt: time.Now(),
		GrantedBy: req.GrantedBy,
	}

	// Save to database
	if err := uc.roleRepo.Assign(ctx, assignment); err != nil {
		return errors.New("failed to assign role: " + err.Error())
	}

	event := &entities.UserEvent{
		TenantID:   req.TenantID,
		UserID:     req.UserID,
		EventType:  entities.EventTypeRoleAssigned,
		Details:    map[string]any{"role": req.Role, "client_id": req.ClientID, "granted_by": req.GrantedBy},
		OccurredAt: time.Now(),
	}
	if err := uc.eventRepo.Record(ctx, event); err != nil {
		return errors.New("failed to record role assigned event: " + err.Error())
	}

	return nil
}
