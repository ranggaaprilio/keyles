package user

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// SessionOutput represents a single user session
type SessionOutput struct {
	TokenID    string
	ClientID   string
	ClientName string
	CreatedAt  time.Time
	LastUsedAt *time.Time
	ExpiresAt  time.Time
}

// ListSessions handles listing active sessions for a user (US4)
type ListSessions struct {
	refreshTokenRepo repositories.RefreshTokenRepository
}

// NewListSessions creates a new ListSessions use case
func NewListSessions(refreshTokenRepo repositories.RefreshTokenRepository) *ListSessions {
	return &ListSessions{
		refreshTokenRepo: refreshTokenRepo,
	}
}

// Execute lists all non-revoked, non-expired refresh tokens for a user
func (uc *ListSessions) Execute(ctx context.Context, userID string, tenantID string) ([]SessionOutput, error) {
	if userID == "" {
		return nil, errors.New("user_id is required")
	}
	if tenantID == "" {
		return nil, errors.New("tenant_id is required")
	}

	tokens, err := uc.refreshTokenRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	now := time.Now()
	outputs := make([]SessionOutput, 0)
	for _, token := range tokens {
		if token.IsRevoked() || token.IsExpired() || now.After(token.ExpiresAt) {
			continue
		}

		outputs = append(outputs, SessionOutput{
			TokenID:    strconv.FormatInt(token.ID, 10),
			ClientID:   token.ClientID,
			ClientName: "",
			CreatedAt:  token.CreatedAt,
			LastUsedAt: token.LastUsedAt,
			ExpiresAt:  token.ExpiresAt,
		})
	}

	return outputs, nil
}
