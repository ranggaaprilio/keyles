package user

import (
"context"
"errors"
"math"

"github.com/ranggaaprilio/keyles/domain/entities"
"github.com/ranggaaprilio/keyles/domain/repositories"
)

// ListUsersInput represents the request to list users
type ListUsersInput struct {
	TenantID     string
	Search       string
	StatusFilter entities.UserStatus
	Page         int
	PageSize     int
}

// ListUsersOutput represents the paginated list response
type ListUsersOutput struct {
	Users      []*entities.User
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

// ListUsers handles paginated user listing
type ListUsers struct {
	endUserRepo repositories.EndUserRepository
}

// NewListUsers creates a new ListUsers use case
func NewListUsers(endUserRepo repositories.EndUserRepository) *ListUsers {
	return &ListUsers{endUserRepo: endUserRepo}
}

// Execute returns a paginated list of users for the tenant
func (uc *ListUsers) Execute(ctx context.Context, input ListUsersInput) (*ListUsersOutput, error) {
	if input.TenantID == "" {
		return nil, errors.New("tenant_id is required")
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

	users, total, err := uc.endUserRepo.ListByTenant(ctx, input.TenantID, input.Search, input.StatusFilter, page, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &ListUsersOutput{
		Users:      users,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
