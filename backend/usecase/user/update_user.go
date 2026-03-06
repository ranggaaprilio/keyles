package user

import (
"context"
"errors"
"fmt"
"time"

"github.com/ranggaaprilio/keyles/domain/entities"
"github.com/ranggaaprilio/keyles/domain/repositories"
)

// UpdateUserInput represents the request to update a user
type UpdateUserInput struct {
	UserID      string
	TenantID    string
	DisplayName *string // nil means no change
}

// UpdateUser handles updating user details (display_name only)
type UpdateUser struct {
	endUserRepo repositories.EndUserRepository
	auditRepo   repositories.AuditRepository
}

// NewUpdateUser creates a new UpdateUser use case
func NewUpdateUser(endUserRepo repositories.EndUserRepository, auditRepo repositories.AuditRepository) *UpdateUser {
	return &UpdateUser{
		endUserRepo: endUserRepo,
		auditRepo:   auditRepo,
	}
}

// Execute updates a user's display name
func (uc *UpdateUser) Execute(ctx context.Context, input UpdateUserInput) (*entities.User, error) {
if input.UserID == "" {
return nil, errors.New("user_id is required")
}
if input.TenantID == "" {
return nil, errors.New("tenant_id is required")
}

// Validate display name length if provided
if input.DisplayName != nil && len(*input.DisplayName) > 255 {
return nil, errors.New("display name must not exceed 255 characters")
}

user, err := uc.endUserRepo.GetByID(ctx, input.UserID)
if err != nil {
return nil, fmt.Errorf("user not found: %w", err)
}

// Tenant isolation
if user.TenantID != input.TenantID {
return nil, errors.New("user not found")
}

// Apply changes
if input.DisplayName != nil {
user.DisplayName = *input.DisplayName
}
user.UpdatedAt = time.Now()

if err := uc.endUserRepo.Update(ctx, user); err != nil {
return nil, fmt.Errorf("failed to update user: %w", err)
}

// Audit log (fire-and-forget)
auditLog := entities.NewAuditLog("user_updated", "", "")
auditLog.WithData("user_id", input.UserID)
auditLog.WithData("tenant_id", input.TenantID)
if input.DisplayName != nil {
auditLog.WithData("display_name", *input.DisplayName)
}
_ = uc.auditRepo.Create(ctx, auditLog)

return user, nil
}
