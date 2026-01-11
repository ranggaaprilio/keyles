package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserRepositoryForUserInfo for testing GetUserInfo
type MockUserRepositoryForUserInfo struct {
	mock.Mock
}

func (m *MockUserRepositoryForUserInfo) Create(ctx context.Context, user *entities.AdminUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepositoryForUserInfo) FindByID(ctx context.Context, id uuid.UUID) (*entities.AdminUser, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.AdminUser), args.Error(1)
}

func (m *MockUserRepositoryForUserInfo) FindByEmail(ctx context.Context, email string) (*entities.AdminUser, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.AdminUser), args.Error(1)
}

func (m *MockUserRepositoryForUserInfo) FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entities.AdminUser, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.AdminUser), args.Error(1)
}

func (m *MockUserRepositoryForUserInfo) Update(ctx context.Context, user *entities.AdminUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepositoryForUserInfo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepositoryForUserInfo) EmailExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func TestGetUserInfo_Success(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockUserRepo := new(MockUserRepositoryForUserInfo)

	useCase := auth.NewGetUserInfo(mockUserRepo)

	// Test user
	userID := uuid.New()
	tenantID := uuid.New()
	user := &entities.AdminUser{
		ID:       userID,
		TenantID: tenantID,
		Email:    "user@example.com",
		Role:     entities.UserRoleMember,
	}

	// Mock expectations
	mockUserRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Execute
	userInfo, err := useCase.Execute(ctx, userID.String())

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, userInfo)
	assert.Equal(t, userID.String(), userInfo.Sub)
	assert.Equal(t, user.Email, userInfo.Email)
	assert.Equal(t, tenantID.String(), userInfo.TenantID)
	assert.True(t, userInfo.EmailVerified)
	mockUserRepo.AssertExpectations(t)
}

func TestGetUserInfo_UserNotFound(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockUserRepo := new(MockUserRepositoryForUserInfo)

	useCase := auth.NewGetUserInfo(mockUserRepo)

	userID := uuid.New()

	// Mock expectations
	mockUserRepo.On("FindByID", ctx, userID).Return(nil, assert.AnError)

	// Execute
	userInfo, err := useCase.Execute(ctx, userID.String())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, userInfo)
	mockUserRepo.AssertExpectations(t)
}

func TestGetUserInfo_EmptyUserID(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockUserRepo := new(MockUserRepositoryForUserInfo)

	useCase := auth.NewGetUserInfo(mockUserRepo)

	// Execute with empty user ID
	userInfo, err := useCase.Execute(ctx, "")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, userInfo)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestGetUserInfo_InvalidUUID(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockUserRepo := new(MockUserRepositoryForUserInfo)

	useCase := auth.NewGetUserInfo(mockUserRepo)

	// Execute with invalid UUID
	userInfo, err := useCase.Execute(ctx, "invalid-uuid")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, userInfo)
	assert.Contains(t, err.Error(), "invalid user_id format")
}

func TestGetUserInfo_ClaimsMapping(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockUserRepo := new(MockUserRepositoryForUserInfo)

	useCase := auth.NewGetUserInfo(mockUserRepo)

	// Test user with all fields
	userID := uuid.New()
	tenantID := uuid.New()
	user := &entities.AdminUser{
		ID:       userID,
		TenantID: tenantID,
		Email:    "admin@example.com",
		Role:     entities.UserRoleAdmin,
	}

	// Mock expectations
	mockUserRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Execute
	userInfo, err := useCase.Execute(ctx, userID.String())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, userID.String(), userInfo.Sub)
	assert.Equal(t, user.Email, userInfo.Email)
	assert.True(t, userInfo.EmailVerified)
	assert.Equal(t, tenantID.String(), userInfo.TenantID)
	mockUserRepo.AssertExpectations(t)
}
