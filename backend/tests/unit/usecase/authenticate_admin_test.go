/**
 * Unit tests for AuthenticateAdmin use case
 */

package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.AdminUser, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.AdminUser), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.AdminUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.AdminUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// MockTenantRepository for testing
type MockTenantRepository struct {
	mock.Mock
}

func (m *MockTenantRepository) FindByID(ctx context.Context, id string) (*entities.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

func (m *MockTenantRepository) Create(ctx context.Context, tenant *entities.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) Update(ctx context.Context, tenant *entities.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) FindByOrganizationName(ctx context.Context, name string) (*entities.Tenant, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

// MockPasswordService for testing
type MockPasswordService struct {
	mock.Mock
}

func (m *MockPasswordService) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordService) Compare(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

// MockJWTService for testing
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateToken(userID, tenantID, email, role string) (string, error) {
	args := m.Called(userID, tenantID, email, role)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) ValidateToken(token string) (*auth.JWTClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.JWTClaims), args.Error(1)
}

func TestAuthenticateAdmin_Success(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockTenantRepo := new(MockTenantRepository)
	mockPasswordService := new(MockPasswordService)
	mockJWTService := new(MockJWTService)

	hashedPassword := "$2a$10$hashedpassword"
	tenant := &entities.Tenant{
		ID:               "tenant-123",
		OrganizationName: "Test Org",
		Status:           "active",
		CreatedAt:        time.Now(),
		VerifiedAt:       func() *time.Time { t := time.Now(); return &t }(),
	}

	user := &entities.AdminUser{
		ID:           "user-123",
		TenantID:     "tenant-123",
		FullName:     "John Doe",
		Email:        "john@example.com",
		PasswordHash: hashedPassword,
		Role:         "admin",
		CreatedAt:    time.Now(),
	}

	mockUserRepo.On("FindByEmail", mock.Anything, "john@example.com").Return(user, nil)
	mockPasswordService.On("Compare", hashedPassword, "password123").Return(nil)
	mockTenantRepo.On("FindByID", mock.Anything, "tenant-123").Return(tenant, nil)
	mockJWTService.On("GenerateToken", "user-123", "tenant-123", "john@example.com", "admin").Return("jwt-token", nil)

	useCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, mockPasswordService, mockJWTService)

	// Act
	result, err := useCase.Execute(context.Background(), "john@example.com", "password123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "jwt-token", result.Token)
	assert.Equal(t, "user-123", result.User.ID)
	assert.Equal(t, "john@example.com", result.User.Email)
	assert.Equal(t, "tenant-123", result.Tenant.ID)
	assert.Equal(t, "Test Org", result.Tenant.OrganizationName)
	mockUserRepo.AssertExpectations(t)
	mockTenantRepo.AssertExpectations(t)
	mockPasswordService.AssertExpectations(t)
	mockJWTService.AssertExpectations(t)
}

func TestAuthenticateAdmin_UserNotFound(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockTenantRepo := new(MockTenantRepository)
	mockPasswordService := new(MockPasswordService)
	mockJWTService := new(MockJWTService)

	mockUserRepo.On("FindByEmail", mock.Anything, "notfound@example.com").Return(nil, errors.New("user not found"))

	useCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, mockPasswordService, mockJWTService)

	// Act
	result, err := useCase.Execute(context.Background(), "notfound@example.com", "password123")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")
	mockUserRepo.AssertExpectations(t)
}

func TestAuthenticateAdmin_InvalidPassword(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockTenantRepo := new(MockTenantRepository)
	mockPasswordService := new(MockPasswordService)
	mockJWTService := new(MockJWTService)

	hashedPassword := "$2a$10$hashedpassword"
	user := &entities.AdminUser{
		ID:           "user-123",
		TenantID:     "tenant-123",
		Email:        "john@example.com",
		PasswordHash: hashedPassword,
	}

	mockUserRepo.On("FindByEmail", mock.Anything, "john@example.com").Return(user, nil)
	mockPasswordService.On("Compare", hashedPassword, "wrongpassword").Return(errors.New("password mismatch"))

	useCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, mockPasswordService, mockJWTService)

	// Act
	result, err := useCase.Execute(context.Background(), "john@example.com", "wrongpassword")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid credentials")
	mockUserRepo.AssertExpectations(t)
	mockPasswordService.AssertExpectations(t)
}

func TestAuthenticateAdmin_TenantNotFound(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockTenantRepo := new(MockTenantRepository)
	mockPasswordService := new(MockPasswordService)
	mockJWTService := new(MockJWTService)

	hashedPassword := "$2a$10$hashedpassword"
	user := &entities.AdminUser{
		ID:           "user-123",
		TenantID:     "tenant-123",
		Email:        "john@example.com",
		PasswordHash: hashedPassword,
	}

	mockUserRepo.On("FindByEmail", mock.Anything, "john@example.com").Return(user, nil)
	mockPasswordService.On("Compare", hashedPassword, "password123").Return(nil)
	mockTenantRepo.On("FindByID", mock.Anything, "tenant-123").Return(nil, errors.New("tenant not found"))

	useCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, mockPasswordService, mockJWTService)

	// Act
	result, err := useCase.Execute(context.Background(), "john@example.com", "password123")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "tenant not found")
	mockUserRepo.AssertExpectations(t)
	mockPasswordService.AssertExpectations(t)
	mockTenantRepo.AssertExpectations(t)
}

func TestAuthenticateAdmin_TenantNotActive(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockTenantRepo := new(MockTenantRepository)
	mockPasswordService := new(MockPasswordService)
	mockJWTService := new(MockJWTService)

	hashedPassword := "$2a$10$hashedpassword"
	tenant := &entities.Tenant{
		ID:               "tenant-123",
		OrganizationName: "Test Org",
		Status:           "pending",
		CreatedAt:        time.Now(),
	}

	user := &entities.AdminUser{
		ID:           "user-123",
		TenantID:     "tenant-123",
		Email:        "john@example.com",
		PasswordHash: hashedPassword,
	}

	mockUserRepo.On("FindByEmail", mock.Anything, "john@example.com").Return(user, nil)
	mockPasswordService.On("Compare", hashedPassword, "password123").Return(nil)
	mockTenantRepo.On("FindByID", mock.Anything, "tenant-123").Return(tenant, nil)

	useCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, mockPasswordService, mockJWTService)

	// Act
	result, err := useCase.Execute(context.Background(), "john@example.com", "password123")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "tenant not active")
	mockUserRepo.AssertExpectations(t)
	mockPasswordService.AssertExpectations(t)
	mockTenantRepo.AssertExpectations(t)
}

func TestAuthenticateAdmin_JWTGenerationFails(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockTenantRepo := new(MockTenantRepository)
	mockPasswordService := new(MockPasswordService)
	mockJWTService := new(MockJWTService)

	hashedPassword := "$2a$10$hashedpassword"
	tenant := &entities.Tenant{
		ID:               "tenant-123",
		OrganizationName: "Test Org",
		Status:           "active",
		CreatedAt:        time.Now(),
		VerifiedAt:       func() *time.Time { t := time.Now(); return &t }(),
	}

	user := &entities.AdminUser{
		ID:           "user-123",
		TenantID:     "tenant-123",
		Email:        "john@example.com",
		PasswordHash: hashedPassword,
		Role:         "admin",
	}

	mockUserRepo.On("FindByEmail", mock.Anything, "john@example.com").Return(user, nil)
	mockPasswordService.On("Compare", hashedPassword, "password123").Return(nil)
	mockTenantRepo.On("FindByID", mock.Anything, "tenant-123").Return(tenant, nil)
	mockJWTService.On("GenerateToken", "user-123", "tenant-123", "john@example.com", "admin").Return("", errors.New("jwt generation failed"))

	useCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, mockPasswordService, mockJWTService)

	// Act
	result, err := useCase.Execute(context.Background(), "john@example.com", "password123")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "jwt generation failed")
	mockUserRepo.AssertExpectations(t)
	mockPasswordService.AssertExpectations(t)
	mockTenantRepo.AssertExpectations(t)
	mockJWTService.AssertExpectations(t)
}

func TestAuthenticateAdmin_EmptyEmail(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockTenantRepo := new(MockTenantRepository)
	mockPasswordService := new(MockPasswordService)
	mockJWTService := new(MockJWTService)

	useCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, mockPasswordService, mockJWTService)

	// Act
	result, err := useCase.Execute(context.Background(), "", "password123")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "email is required")
}

func TestAuthenticateAdmin_EmptyPassword(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockTenantRepo := new(MockTenantRepository)
	mockPasswordService := new(MockPasswordService)
	mockJWTService := new(MockJWTService)

	useCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, mockPasswordService, mockJWTService)

	// Act
	result, err := useCase.Execute(context.Background(), "john@example.com", "")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "password is required")
}
