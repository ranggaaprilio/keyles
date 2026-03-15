/**
 * AuthenticateAdmin use case
 * Validates user credentials and returns JWT token
 */

package auth

import (
	"context"
	"errors"

	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// JWTClaims represents JWT token claims (re-exported from infrastructure for use case layer)
type JWTClaims struct {
	UserID   string
	TenantID string
	Email    string
	Role     string
}

// PasswordService interface for password operations
type PasswordService interface {
	Compare(hashedPassword, password string) error
}

// JWTService interface for JWT operations
type JWTService interface {
	GenerateToken(userID, tenantID, email, role string) (string, error)
	ValidateToken(token string) (*JWTClaims, error)
}

// AuthenticationResult represents the result of authentication
type AuthenticationResult struct {
	Token     string
	ExpiresIn int
	User      UserInfo
	Tenant    TenantInfo
}

// UserInfo contains user information for the response
type UserInfo struct {
	ID       string
	Email    string
	FullName string
	Role     string
}

// TenantInfo contains tenant information for the response
type TenantInfo struct {
	ID               string
	OrganizationName string
	Status           string
}

// AuthenticateAdminUseCase handles admin user authentication
type AuthenticateAdminUseCase struct {
	userRepo        repositories.UserRepository
	tenantRepo      repositories.TenantRepository
	passwordService PasswordService
	jwtService      JWTService
}

// NewAuthenticateAdminUseCase creates a new authenticate admin use case
func NewAuthenticateAdminUseCase(
	userRepo repositories.UserRepository,
	tenantRepo repositories.TenantRepository,
	passwordService PasswordService,
	jwtService JWTService,
) *AuthenticateAdminUseCase {
	return &AuthenticateAdminUseCase{
		userRepo:        userRepo,
		tenantRepo:      tenantRepo,
		passwordService: passwordService,
		jwtService:      jwtService,
	}
}

// Execute authenticates an admin user and returns a JWT token
func (uc *AuthenticateAdminUseCase) Execute(ctx context.Context, email, password string) (*AuthenticationResult, error) {
	// Step 1: Validate input
	if email == "" {
		return nil, errors.New("email is required")
	}
	if password == "" {
		return nil, errors.New("password is required")
	}

	// Step 2: Find user by email
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Step 3: Verify password
	err = uc.passwordService.Compare(user.PasswordHash, password)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Step 4: Load tenant
	tenant, err := uc.tenantRepo.FindByID(ctx, user.TenantID)
	if err != nil {
		return nil, errors.New("tenant not found")
	}
	if tenant == nil {
		return nil, errors.New("tenant not found")
	}

	// Step 5: Check tenant is active
	if tenant.Status != "active" {
		return nil, errors.New("tenant not active - please verify your email first")
	}

	// Step 6: Generate JWT token
	token, err := uc.jwtService.GenerateToken(
		user.ID.String(),
		tenant.ID.String(),
		user.Email,
		string(user.Role),
	)
	if err != nil {
		return nil, err
	}

	// Step 7: Return authentication result
	result := &AuthenticationResult{
		Token:     token,
		ExpiresIn: 86400, // 24 hours in seconds
		User: UserInfo{
			ID:       user.ID.String(),
			Email:    user.Email,
			FullName: user.FullName,
			Role:     string(user.Role),
		},
		Tenant: TenantInfo{
			ID:               tenant.ID.String(),
			OrganizationName: tenant.OrganizationName,
			Status:           string(tenant.Status),
		},
	}

	return result, nil
}
