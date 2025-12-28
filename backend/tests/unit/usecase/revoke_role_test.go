package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/role"
)

// TestRevokeRole_Success tests successful role revocation
func TestRevokeRole_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRoleRepo := new(mocks.MockRoleRepository)
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	// Role exists
	mockRoleRepo.On("HasRole", ctx, "user-123", "client-123", "user").Return(true, nil)
	mockRoleRepo.On("RevokeRole", ctx, "user-123", "client-123", "user").Return(nil)
	// Cascade revocation of refresh tokens per FR-006e
	mockRefreshTokenRepo.On("RevokeAllForUser", ctx, "user-123", "client-123").Return(nil)

	uc := role.NewRevokeRole(mockRoleRepo, mockRefreshTokenRepo)

	// Act
	err := uc.Execute(ctx, role.RevokeRoleRequest{
		UserID:   "user-123",
		ClientID: "client-123",
		Role:     "user",
	})

	// Assert
	assert.NoError(t, err)
	mockRoleRepo.AssertExpectations(t)
	mockRefreshTokenRepo.AssertExpectations(t)
}

// TestRevokeRole_RoleNotFound tests error when role doesn't exist
func TestRevokeRole_RoleNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRoleRepo := new(mocks.MockRoleRepository)
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	// Role doesn't exist
	mockRoleRepo.On("HasRole", ctx, "user-123", "client-123", "user").Return(false, nil)

	uc := role.NewRevokeRole(mockRoleRepo, mockRefreshTokenRepo)

	// Act
	err := uc.Execute(ctx, role.RevokeRoleRequest{
		UserID:   "user-123",
		ClientID: "client-123",
		Role:     "user",
	})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role not found")
}

// TestRevokeRole_CascadeRefreshTokenRevocation tests that refresh tokens are revoked when role is revoked
func TestRevokeRole_CascadeRefreshTokenRevocation(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRoleRepo := new(mocks.MockRoleRepository)
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	mockRoleRepo.On("HasRole", ctx, "user-123", "client-123", "admin").Return(true, nil)
	mockRoleRepo.On("RevokeRole", ctx, "user-123", "client-123", "admin").Return(nil)
	// This should be called to revoke all refresh tokens (FR-006e)
	mockRefreshTokenRepo.On("RevokeAllForUser", ctx, "user-123", "client-123").Return(nil)

	uc := role.NewRevokeRole(mockRoleRepo, mockRefreshTokenRepo)

	// Act
	err := uc.Execute(ctx, role.RevokeRoleRequest{
		UserID:   "user-123",
		ClientID: "client-123",
		Role:     "admin",
	})

	// Assert
	assert.NoError(t, err)
	mockRefreshTokenRepo.AssertCalled(t, "RevokeAllForUser", ctx, "user-123", "client-123")
}

// TestRevokeRole_MissingFields tests validation for missing required fields
func TestRevokeRole_MissingFields(t *testing.T) {
	ctx := context.Background()
	mockRoleRepo := new(mocks.MockRoleRepository)
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	uc := role.NewRevokeRole(mockRoleRepo, mockRefreshTokenRepo)

	tests := []struct {
		name string
		req  role.RevokeRoleRequest
	}{
		{
			name: "Missing user_id",
			req: role.RevokeRoleRequest{
				ClientID: "client-123",
				Role:     "user",
			},
		},
		{
			name: "Missing client_id",
			req: role.RevokeRoleRequest{
				UserID: "user-123",
				Role:   "user",
			},
		},
		{
			name: "Missing role",
			req: role.RevokeRoleRequest{
				UserID:   "user-123",
				ClientID: "client-123",
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

// TestRevokeRole_DatabaseError tests database error handling
func TestRevokeRole_DatabaseError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRoleRepo := new(mocks.MockRoleRepository)
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	mockRoleRepo.On("HasRole", ctx, "user-123", "client-123", "user").Return(true, nil)
	mockRoleRepo.On("RevokeRole", ctx, "user-123", "client-123", "user").Return(errors.New("database error"))

	uc := role.NewRevokeRole(mockRoleRepo, mockRefreshTokenRepo)

	// Act
	err := uc.Execute(ctx, role.RevokeRoleRequest{
		UserID:   "user-123",
		ClientID: "client-123",
		Role:     "user",
	})

	// Assert
	assert.Error(t, err)
}

// TestRevokeRole_RefreshTokenRevocationError tests handling of refresh token revocation errors
func TestRevokeRole_RefreshTokenRevocationError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRoleRepo := new(mocks.MockRoleRepository)
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	mockRoleRepo.On("HasRole", ctx, "user-123", "client-123", "user").Return(true, nil)
	mockRoleRepo.On("RevokeRole", ctx, "user-123", "client-123", "user").Return(nil)
	// Refresh token revocation fails
	mockRefreshTokenRepo.On("RevokeAllForUser", ctx, "user-123", "client-123").Return(errors.New("refresh token revocation failed"))

	uc := role.NewRevokeRole(mockRoleRepo, mockRefreshTokenRepo)

	// Act - should still succeed even if refresh token revocation fails (best effort)
	err := uc.Execute(ctx, role.RevokeRoleRequest{
		UserID:   "user-123",
		ClientID: "client-123",
		Role:     "user",
	})

	// Assert - role revocation should still succeed
	assert.NoError(t, err)
}
