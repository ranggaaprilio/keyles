package user

import (
"context"
"errors"
"fmt"
"math"

"github.com/ranggaaprilio/keyles/domain/entities"
"github.com/ranggaaprilio/keyles/domain/repositories"
)

// ListUserActivityInput represents the request to list user activity
type ListUserActivityInput struct {
	UserID   string
	TenantID string
	Page     int
	PageSize int
}

// ListUserActivityOutput represents the paginated activity response
type ListUserActivityOutput struct {
	Events     []*entities.UserEvent
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

// ListUserActivity handles listing paginated activity events for a user
type ListUserActivity struct {
	endUserRepo repositories.EndUserRepository
	eventRepo   repositories.UserEventRepository
}

// NewListUserActivity creates a new ListUserActivity use case
func NewListUserActivity(
endUserRepo repositories.EndUserRepository,
eventRepo repositories.UserEventRepository,
) *ListUserActivity {
	return &ListUserActivity{
		endUserRepo: endUserRepo,
		eventRepo:   eventRepo,
	}
}

// Execute returns paginated activity events for the user
func (uc *ListUserActivity) Execute(ctx context.Context, input ListUserActivityInput) (*ListUserActivityOutput, error) {
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

	// Default and clamp pagination
	page := input.Page
	if page < 1 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize < 1 {
		pageSize = 25
	}
	if pageSize > 100 {
		pageSize = 100
	}

	events, total, err := uc.eventRepo.ListByUser(ctx, input.UserID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list activity: %w", err)
	}

	if events == nil {
		events = []*entities.UserEvent{}
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &ListUserActivityOutput{
		Events:     events,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
