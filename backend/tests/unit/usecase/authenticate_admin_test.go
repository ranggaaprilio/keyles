/**
 * Unit tests for AuthenticateAdmin use case
 */

package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
	mockUserRepo := new(mocks.MockUserRepository)
	mockTenantRepo := new(mocks.MockTenantRepository)
	mockPasswordService := new(mocks.MockPasswordService)
	mockJWTService := new(MockJWTService)

	tenantID := uuid.New()
	userID := uuid.New()
	
	hashedPassword := "$2a$10$hashedpassword"
	tenant := &entities.Tenant{
		ID:               tenantID,
		OrganizationName: "Test Org",
		Status:           "active",
		CreatedAt:        time.Now(),
		VerifiedAt:       func() *time.Time { t := time.Now(); return &t }(),
	}

	user := &entities.AdminUser{
		ID:           userID,
		TenantID:     tenantID,
		FullName:     "John Doe",
		Email:        "john@example.com",
		PasswordHash: hashedPassword,
		Role:         "admin",
		CreatedAt:    time.Now(),
	}

	mockUserRepo.On("FindByEmail", mock.Anything, "john@example.com").Return(user, nil)
	mockPasswordService.On("Compare", hashedPassword, "password123").Return(nil)
	mockTenantRepo.On("FindByID", mock.Anything, tenantID).Return(tenant, nil)
	mockJWTService.On("GenerateToken", userID.String(), tenantID.String(), "john@example.com", "admin").Return("jwt-token", nil)

	useCase := auth.NewAuthenticateAdminUseCase(mockUserRepo, mockTenantRepo, mockPasswordService, mockJWTService)

	// Act
	result, err := useCase.Execute(context.Background(), "john@example.com", "password123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "jwt-token", result.Token)
	assert.Equal(t, userID.String(), result.User.ID)
	assert.Equal(t, "john@example.com", result.User.Email)
	assert.Equal(t, "Test Org", result.Tenant.OrganizationName)
	mockUserRepo.AssertExpectations(t)
	mockPasswordService.AssertExpectations(t)
	mockTenantRepo.AssertExpectations(t)
	mockJWTService.AssertExpectations(t)
}

func TestAuthenticateAdmin_UserNotFound(t *testing.T) {
	// Arrange
	mockUserRepo := new(mocks.MockUserRepository)
	mockTenantRepo := new(mocks.MockTenantRepository)
	mockPasswordService := new(mocks.MockPasswordService)
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
	mockUserRepo := new(mocks.MockUserRepository)
	mockTenantRepo := new(mocks.MockTenantRepository)
	mockPasswordService := new(mocks.MockPasswordService)
	mockJWTService := new(MockJWTService)

	userID := uuid.New()
	tenantID := uuid.New()
	
	hashedPassword := "$2a$10$hashedpassword"
	user := &entities.AdminUser{
		ID:           userID,
		TenantID:     tenantID,
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
	mockUserRepo := new(mocks.MockUserRepository)
	mockTenantRepo := new(mocks.MockTenantRepository)
	mockPasswordService := new(mocks.MockPasswordService)
	mockJWTService := new(MockJWTService)

	userID := uuid.New()
	tenantID := uuid.New()
	
	hashedPassword := "$2a$10$hashedpassword"
	user := &entities.AdminUser{
		ID:           userID,
		TenantID:     tenantID,
		Email:        "john@example.com",
		PasswordHash: hashedPassword,
	}

	mockUserRepo.On("FindByEmail", mock.Anything, "john@example.com").Return(user, nil)
	mockPasswordService.On("Compare", hashedPassword, "password123").Return(nil)
	mockTenantRepo.On("FindByID", mock.Anything, tenantID).Return(nil, errors.New("tenant not found"))

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
