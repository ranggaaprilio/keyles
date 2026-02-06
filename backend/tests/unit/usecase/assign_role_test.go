package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/role"
)

// TestAssignRole_Success tests successful role assignment
func TestAssignRole_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRoleRepo := new(mocks.MockRoleRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockClientRepo := new(mocks.MockClientRepository)

	userID := uuid.New()
	tenantID := uuid.New()

	validUser := &entities.AdminUser{
		ID:       userID,
		TenantID: tenantID,
	}

	validClient := &entities.Client{
		ClientID: "client-123",
		TenantID: tenantID.String(),
		IsActive: true,
	}

	mockUserRepo.On("FindByID", ctx, userID).Return(validUser, nil)
	mockClientRepo.On("GetByID", ctx, "client-123").Return(validClient, nil)
	mockRoleRepo.On("HasRole", ctx, userID.String(), "client-123", "user").Return(false, nil)
	mockRoleRepo.On("AssignRole", ctx, mock.AnythingOfType("*entities.UserRoleAssignment")).Return(nil)

	uc := role.NewAssignRole(mockRoleRepo, mockUserRepo, mockClientRepo)

	// Act
	err := uc.Execute(ctx, role.AssignRoleRequest{
		UserID:    userID.String(),
		ClientID:  "client-123",
		TenantID:  tenantID.String(),
		Role:      "user",
		GrantedBy: "admin-123",
	})

	// Assert
	assert.NoError(t, err)
	mockRoleRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockClientRepo.AssertExpectations(t)
}

// TestAssignRole_UserNotFound tests error when user doesn't exist
func TestAssignRole_UserNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRoleRepo := new(mocks.MockRoleRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockClientRepo := new(mocks.MockClientRepository)

	userID := uuid.New()

	mockUserRepo.On("FindByID", ctx, userID).Return(nil, errors.New("user not found"))

	uc := role.NewAssignRole(mockRoleRepo, mockUserRepo, mockClientRepo)

	// Act
	err := uc.Execute(ctx, role.AssignRoleRequest{
		UserID:    userID.String(),
		ClientID:  "client-123",
		TenantID:  "tenant-123",
		Role:      "user",
		GrantedBy: "admin-123",
	})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

// TestAssignRole_ClientNotFound tests error when client doesn't exist
func TestAssignRole_ClientNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRoleRepo := new(mocks.MockRoleRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockClientRepo := new(mocks.MockClientRepository)

	userID := uuid.New()
	tenantID := uuid.New()

	validUser := &entities.AdminUser{
		ID:       userID,
		TenantID: tenantID,
	}

	mockUserRepo.On("FindByID", ctx, userID).Return(validUser, nil)
	mockClientRepo.On("GetByID", ctx, "client-123").Return(nil, errors.New("client not found"))

	uc := role.NewAssignRole(mockRoleRepo, mockUserRepo, mockClientRepo)

	// Act
	err := uc.Execute(ctx, role.AssignRoleRequest{
		UserID:    userID.String(),
		ClientID:  "client-123",
		TenantID:  tenantID.String(),
		Role:      "user",
		GrantedBy: "admin-123",
	})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client not found")
}

// TestAssignRole_DuplicatePrevention tests that duplicate role assignments are prevented
func TestAssignRole_DuplicatePrevention(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRoleRepo := new(mocks.MockRoleRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockClientRepo := new(mocks.MockClientRepository)

	userID := uuid.New()
	tenantID := uuid.New()

	validUser := &entities.AdminUser{
		ID:       userID,
		TenantID: tenantID,
	}

	validClient := &entities.Client{
		ClientID: "client-123",
		TenantID: tenantID.String(),
		IsActive: true,
	}

	mockUserRepo.On("FindByID", ctx, userID).Return(validUser, nil)
	mockClientRepo.On("GetByID", ctx, "client-123").Return(validClient, nil)
	// User already has the role
	mockRoleRepo.On("HasRole", ctx, userID.String(), "client-123", "user").Return(true, nil)

	uc := role.NewAssignRole(mockRoleRepo, mockUserRepo, mockClientRepo)

	// Act
	err := uc.Execute(ctx, role.AssignRoleRequest{
		UserID:    userID.String(),
		ClientID:  "client-123",
		TenantID:  tenantID.String(),
		Role:      "user",
		GrantedBy: "admin-123",
	})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already has role")
}

// TestAssignRole_InvalidRole tests error for invalid role
func TestAssignRole_InvalidRole(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRoleRepo := new(mocks.MockRoleRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockClientRepo := new(mocks.MockClientRepository)

	userID := uuid.New()

	uc := role.NewAssignRole(mockRoleRepo, mockUserRepo, mockClientRepo)

	// Act
	err := uc.Execute(ctx, role.AssignRoleRequest{
		UserID:    userID.String(),
		ClientID:  "client-123",
		TenantID:  "tenant-123",
		Role:      "superuser", // Invalid role
		GrantedBy: "admin-123",
	})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role")
}

// TestAssignRole_MissingFields tests validation for missing required fields
func TestAssignRole_MissingFields(t *testing.T) {
	ctx := context.Background()
	mockRoleRepo := new(mocks.MockRoleRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockClientRepo := new(mocks.MockClientRepository)

	userID := uuid.New()

	uc := role.NewAssignRole(mockRoleRepo, mockUserRepo, mockClientRepo)

	tests := []struct {
		name string
		req  role.AssignRoleRequest
	}{
		{
			name: "Missing user_id",
			req: role.AssignRoleRequest{
				ClientID:  "client-123",
				TenantID:  "tenant-123",
				Role:      "user",
				GrantedBy: "admin-123",
			},
		},
		{
			name: "Missing client_id",
			req: role.AssignRoleRequest{
				UserID:    userID.String(),
				TenantID:  "tenant-123",
				Role:      "user",
				GrantedBy: "admin-123",
			},
		},
		{
			name: "Missing tenant_id",
			req: role.AssignRoleRequest{
				UserID:    userID.String(),
				ClientID:  "client-123",
				Role:      "user",
				GrantedBy: "admin-123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uc.Execute(ctx, tt.req)
			assert.Error(t, err)
		})
	}
}

// TestAssignRole_TenantMismatch tests that user and client must belong to same tenant
func TestAssignRole_TenantMismatch(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRoleRepo := new(mocks.MockRoleRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockClientRepo := new(mocks.MockClientRepository)

	userID := uuid.New()
	userTenantID := uuid.New()
	clientTenantID := uuid.New()

	validUser := &entities.AdminUser{
		ID:       userID,
		TenantID: userTenantID, // Different tenant
	}

	validClient := &entities.Client{
		ClientID: "client-123",
		TenantID: clientTenantID.String(),
		IsActive: true,
	}

	mockUserRepo.On("FindByID", ctx, userID).Return(validUser, nil)
	mockClientRepo.On("GetByID", ctx, "client-123").Return(validClient, nil)

	uc := role.NewAssignRole(mockRoleRepo, mockUserRepo, mockClientRepo)

	// Act
	err := uc.Execute(ctx, role.AssignRoleRequest{
		UserID:    userID.String(),
		ClientID:  "client-123",
		TenantID:  clientTenantID.String(),
		Role:      "user",
		GrantedBy: "admin-123",
	})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant")
}

// TestAssignRole_DatabaseError tests database error handling
func TestAssignRole_DatabaseError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRoleRepo := new(mocks.MockRoleRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockClientRepo := new(mocks.MockClientRepository)

	userID := uuid.New()
	tenantID := uuid.New()

	validUser := &entities.AdminUser{
		ID:       userID,
		TenantID: tenantID,
	}

	validClient := &entities.Client{
		ClientID: "client-123",
		TenantID: tenantID.String(),
		IsActive: true,
	}

	mockUserRepo.On("FindByID", ctx, userID).Return(validUser, nil)
	mockClientRepo.On("GetByID", ctx, "client-123").Return(validClient, nil)
	mockRoleRepo.On("HasRole", ctx, userID.String(), "client-123", "user").Return(false, nil)
	mockRoleRepo.On("AssignRole", ctx, mock.AnythingOfType("*entities.UserRoleAssignment")).Return(errors.New("database error"))

	uc := role.NewAssignRole(mockRoleRepo, mockUserRepo, mockClientRepo)

	// Act
	err := uc.Execute(ctx, role.AssignRoleRequest{
		UserID:    userID.String(),
		ClientID:  "client-123",
		TenantID:  tenantID.String(),
		Role:      "user",
		GrantedBy: "admin-123",
	})

	// Assert
	assert.Error(t, err)
}

// TestAssignRole_InvalidUserIDFormat tests error for invalid user ID format
func TestAssignRole_InvalidUserIDFormat(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRoleRepo := new(mocks.MockRoleRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockClientRepo := new(mocks.MockClientRepository)

	uc := role.NewAssignRole(mockRoleRepo, mockUserRepo, mockClientRepo)

	// Act
	err := uc.Execute(ctx, role.AssignRoleRequest{
		UserID:    "not-a-uuid",
		ClientID:  "client-123",
		TenantID:  "tenant-123",
		Role:      "user",
		GrantedBy: "admin-123",
	})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user_id format")
}
