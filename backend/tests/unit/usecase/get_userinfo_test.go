package usecase

import (
"context"
"testing"

"github.com/google/uuid"
"github.com/ranggaaprilio/keyles/domain/entities"
"github.com/ranggaaprilio/keyles/domain/repositories"
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

// MockRoleRepositoryForUserInfo for testing GetUserInfo
type MockRoleRepositoryForUserInfo struct {
	mock.Mock
}

func (m *MockRoleRepositoryForUserInfo) AssignRole(ctx context.Context, assignment *entities.UserRoleAssignment) error {
	args := m.Called(ctx, assignment)
	return args.Error(0)
}

func (m *MockRoleRepositoryForUserInfo) RevokeRole(ctx context.Context, userID, clientID, role string) error {
	args := m.Called(ctx, userID, clientID, role)
	return args.Error(0)
}

func (m *MockRoleRepositoryForUserInfo) GetUserRoles(ctx context.Context, userID, clientID string) ([]*entities.UserRoleAssignment, error) {
	args := m.Called(ctx, userID, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserRoleAssignment), args.Error(1)
}

func (m *MockRoleRepositoryForUserInfo) HasRole(ctx context.Context, userID, clientID, role string) (bool, error) {
	args := m.Called(ctx, userID, clientID, role)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoleRepositoryForUserInfo) HasAnyRole(ctx context.Context, userID, clientID string) (bool, error) {
	args := m.Called(ctx, userID, clientID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoleRepositoryForUserInfo) ListRolesByClient(ctx context.Context, clientID string) ([]*entities.UserRoleAssignment, error) {
	args := m.Called(ctx, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserRoleAssignment), args.Error(1)
}

func (m *MockRoleRepositoryForUserInfo) ListRolesByUser(ctx context.Context, userID string) ([]*entities.UserRoleAssignment, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserRoleAssignment), args.Error(1)
}

func (m *MockRoleRepositoryForUserInfo) Assign(ctx context.Context, assignment *entities.UserRoleAssignment) error {
	args := m.Called(ctx, assignment)
	return args.Error(0)
}

func (m *MockRoleRepositoryForUserInfo) Revoke(ctx context.Context, assignmentID int64, revokedByUserID string) error {
	args := m.Called(ctx, assignmentID, revokedByUserID)
	return args.Error(0)
}

func (m *MockRoleRepositoryForUserInfo) ListByUser(ctx context.Context, userID string) ([]*entities.UserRoleAssignment, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserRoleAssignment), args.Error(1)
}

func (m *MockRoleRepositoryForUserInfo) ListByClient(ctx context.Context, clientID string, page, pageSize int) ([]*entities.UserRoleAssignment, int, error) {
	args := m.Called(ctx, clientID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*entities.UserRoleAssignment), args.Int(1), args.Error(2)
}

func (m *MockRoleRepositoryForUserInfo) RevokeAllForUser(ctx context.Context, userID, revokedByUserID string) error {
	args := m.Called(ctx, userID, revokedByUserID)
	return args.Error(0)
}

func (m *MockRoleRepositoryForUserInfo) GetActiveRoles(ctx context.Context, userID, clientID string) ([]string, error) {
	args := m.Called(ctx, userID, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

var _ repositories.RoleRepository = (*MockRoleRepositoryForUserInfo)(nil)

func TestGetUserInfo_Success(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := new(MockUserRepositoryForUserInfo)
	mockRoleRepo := new(MockRoleRepositoryForUserInfo)

	useCase := auth.NewGetUserInfo(mockUserRepo, mockRoleRepo)

	userID := uuid.New()
	tenantID := uuid.New()
	user := &entities.AdminUser{
		ID:       userID,
		TenantID: tenantID,
		Email:    "user@example.com",
		Role:     entities.UserRoleMember,
	}

	mockUserRepo.On("FindByID", ctx, userID).Return(user, nil)
	mockRoleRepo.On("GetActiveRoles", ctx, userID.String(), "client-abc").Return([]string{"admin", "viewer"}, nil)

	userInfo, err := useCase.Execute(ctx, userID.String(), "client-abc")

	require.NoError(t, err)
	assert.NotNil(t, userInfo)
	assert.Equal(t, userID.String(), userInfo.Sub)
	assert.Equal(t, user.Email, userInfo.Email)
	assert.Equal(t, tenantID.String(), userInfo.TenantID)
	assert.True(t, userInfo.EmailVerified)
	assert.Equal(t, []string{"admin", "viewer"}, userInfo.Roles)
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestGetUserInfo_UserNotFound(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := new(MockUserRepositoryForUserInfo)
	mockRoleRepo := new(MockRoleRepositoryForUserInfo)

	useCase := auth.NewGetUserInfo(mockUserRepo, mockRoleRepo)

	userID := uuid.New()

	mockUserRepo.On("FindByID", ctx, userID).Return(nil, assert.AnError)

	userInfo, err := useCase.Execute(ctx, userID.String(), "client-abc")

	assert.Error(t, err)
	assert.Nil(t, userInfo)
	mockUserRepo.AssertExpectations(t)
}

func TestGetUserInfo_EmptyUserID(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := new(MockUserRepositoryForUserInfo)
	mockRoleRepo := new(MockRoleRepositoryForUserInfo)

	useCase := auth.NewGetUserInfo(mockUserRepo, mockRoleRepo)

	userInfo, err := useCase.Execute(ctx, "", "client-abc")

	assert.Error(t, err)
	assert.Nil(t, userInfo)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestGetUserInfo_InvalidUUID(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := new(MockUserRepositoryForUserInfo)
	mockRoleRepo := new(MockRoleRepositoryForUserInfo)

	useCase := auth.NewGetUserInfo(mockUserRepo, mockRoleRepo)

	userInfo, err := useCase.Execute(ctx, "invalid-uuid", "client-abc")

	assert.Error(t, err)
	assert.Nil(t, userInfo)
	assert.Contains(t, err.Error(), "invalid user_id format")
}

func TestGetUserInfo_ClaimsMapping(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := new(MockUserRepositoryForUserInfo)
	mockRoleRepo := new(MockRoleRepositoryForUserInfo)

	useCase := auth.NewGetUserInfo(mockUserRepo, mockRoleRepo)

	userID := uuid.New()
	tenantID := uuid.New()
	user := &entities.AdminUser{
		ID:       userID,
		TenantID: tenantID,
		Email:    "admin@example.com",
		Role:     entities.UserRoleAdmin,
	}

	mockUserRepo.On("FindByID", ctx, userID).Return(user, nil)
	mockRoleRepo.On("GetActiveRoles", ctx, userID.String(), "my-client").Return([]string{"admin"}, nil)

	userInfo, err := useCase.Execute(ctx, userID.String(), "my-client")

	require.NoError(t, err)
	assert.Equal(t, userID.String(), userInfo.Sub)
	assert.Equal(t, user.Email, userInfo.Email)
	assert.True(t, userInfo.EmailVerified)
	assert.Equal(t, tenantID.String(), userInfo.TenantID)
	assert.Equal(t, []string{"admin"}, userInfo.Roles)
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestGetUserInfo_RolesErrorDefaultsToEmpty(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := new(MockUserRepositoryForUserInfo)
	mockRoleRepo := new(MockRoleRepositoryForUserInfo)

	useCase := auth.NewGetUserInfo(mockUserRepo, mockRoleRepo)

	userID := uuid.New()
	tenantID := uuid.New()
	user := &entities.AdminUser{
		ID:       userID,
		TenantID: tenantID,
		Email:    "user@example.com",
		Role:     entities.UserRoleMember,
	}

	mockUserRepo.On("FindByID", ctx, userID).Return(user, nil)
	mockRoleRepo.On("GetActiveRoles", ctx, userID.String(), "client-xyz").Return([]string(nil), assert.AnError)

	userInfo, err := useCase.Execute(ctx, userID.String(), "client-xyz")

	require.NoError(t, err)
	assert.NotNil(t, userInfo)
	assert.Equal(t, []string{}, userInfo.Roles)
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestGetUserInfo_EmptyClientIDStillWorks(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := new(MockUserRepositoryForUserInfo)
	mockRoleRepo := new(MockRoleRepositoryForUserInfo)

	useCase := auth.NewGetUserInfo(mockUserRepo, mockRoleRepo)

	userID := uuid.New()
	tenantID := uuid.New()
	user := &entities.AdminUser{
		ID:       userID,
		TenantID: tenantID,
		Email:    "user@example.com",
		Role:     entities.UserRoleMember,
	}

	mockUserRepo.On("FindByID", ctx, userID).Return(user, nil)
	mockRoleRepo.On("GetActiveRoles", ctx, userID.String(), "").Return([]string{}, nil)

	userInfo, err := useCase.Execute(ctx, userID.String(), "")

	require.NoError(t, err)
	assert.NotNil(t, userInfo)
	assert.Equal(t, []string{}, userInfo.Roles)
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}
