package user

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// ActivityOutput represents paginated user activity events
type ActivityOutput struct {
	Events      []*entities.UserEvent
	TotalCount  int
	CurrentPage int
	TotalPages  int
}

// ListUserActivity handles listing activity events for a user (US4)
type ListUserActivity struct {
	eventRepo repositories.UserEventRepository
}

// NewListUserActivity creates a new ListUserActivity use case
func NewListUserActivity(eventRepo repositories.UserEventRepository) *ListUserActivity {
	return &ListUserActivity{
		eventRepo: eventRepo,
	}
}

// Execute retrieves paginated activity events for a user
func (uc *ListUserActivity) Execute(ctx context.Context, userID string, page int, pageSize int) (*ActivityOutput, error) {
	if userID == "" {
		return nil, errors.New("user_id is required")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 25
	}
	if pageSize > 100 {
		pageSize = 100
	}

	events, total, err := uc.eventRepo.ListByUser(ctx, userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list user activity: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages < 1 {
		totalPages = 1
	}

	return &ActivityOutput{
		Events:      events,
		TotalCount:  total,
		CurrentPage: page,
		TotalPages:  totalPages,
	}, nil
}
