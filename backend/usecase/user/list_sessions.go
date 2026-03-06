package user

import (
"context"
"errors"
"fmt"

"github.com/ranggaaprilio/keyles/domain/entities"
"github.com/ranggaaprilio/keyles/domain/repositories"
)

// SessionOutput represents a user's active session
type SessionOutput struct {
ID         int64
ClientID   string
CreatedAt  string
LastUsedAt string
ExpiresAt  string
}

// ListSessions handles listing active sessions for a user
type ListSessions struct {
endUserRepo      repositories.EndUserRepository
refreshTokenRepo repositories.RefreshTokenRepository
}

// NewListSessions creates a new ListSessions use case
func NewListSessions(
endUserRepo repositories.EndUserRepository,
refreshTokenRepo repositories.RefreshTokenRepository,
) *ListSessions {
return &ListSessions{
endUserRepo:      endUserRepo,
refreshTokenRepo: refreshTokenRepo,
}
}

// ListSessionsInput represents the request to list sessions
type ListSessionsInput struct {
UserID   string
TenantID string
}

// Execute returns active sessions for the user
func (uc *ListSessions) Execute(ctx context.Context, input ListSessionsInput) ([]*entities.RefreshToken, error) {
if input.UserID == "" {
return nil, errors.New("user_id is required")
}
if input.TenantID == "" {
return nil, errors.New("tenant_id is required")
}

// Verify user exists and belongs to tenant
user, err := uc.endUserRepo.GetByID(ctx, input.UserID)
if err != nil {
return nil, fmt.Errorf("user not found: %w", err)
}
if user.TenantID != input.TenantID {
return nil, errors.New("user not found")
}

tokens, err := uc.refreshTokenRepo.ListByUserID(ctx, input.UserID)
if err != nil {
return nil, fmt.Errorf("failed to list sessions: %w", err)
}

// Filter to only active (non-revoked, non-expired) tokens
var activeSessions []*entities.RefreshToken
for _, t := range tokens {
if t.IsValid() {
activeSessions = append(activeSessions, t)
}
}

// Return empty slice instead of nil
if activeSessions == nil {
activeSessions = []*entities.RefreshToken{}
}

return activeSessions, nil
}
