package user

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// ListUsersRequest represents a request to list users for a tenant
type ListUsersRequest struct {
	TenantID     string
	Search       string
	StatusFilter entities.UserStatus
	Page         int
	PageSize     int
}

// UserListItem represents a single user in the list
type UserListItem struct {
	ID          string
	Email       string
	DisplayName *string
	Status      entities.UserStatus
	LastLoginAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ListUsersResponse contains the paginated list of users
type ListUsersResponse struct {
	Users      []UserListItem
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
}

// ListUsers handles listing users for a tenant (T040)
type ListUsers struct {
	userRepo repositories.EndUserRepository
}

// NewListUsers creates a new ListUsers use case
func NewListUsers(userRepo repositories.EndUserRepository) *ListUsers {
	return &ListUsers{
		userRepo: userRepo,
	}
}

// Execute lists users for a tenant with optional filtering and pagination
func (uc *ListUsers) Execute(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error) {
	if req.TenantID == "" {
		return nil, errors.New("tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, errors.New("invalid tenant_id format")
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 25
	}
	if pageSize > 100 {
		pageSize = 100
	}

	users, total, err := uc.userRepo.ListByTenant(ctx, tenantUUID, req.Search, req.StatusFilter, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	items := make([]UserListItem, len(users))
	for i, u := range users {
		items[i] = UserListItem{
			ID:          u.ID.String(),
			Email:       u.Email,
			DisplayName: u.DisplayName,
			Status:      u.Status,
			LastLoginAt: u.LastLoginAt,
			CreatedAt:   u.CreatedAt,
			UpdatedAt:   u.UpdatedAt,
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages < 1 {
		totalPages = 1
	}

	return &ListUsersResponse{
		Users:      items,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
