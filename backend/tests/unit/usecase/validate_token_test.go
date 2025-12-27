package usecase

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ranggaaprilio/keyles/domain/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTokenServiceForValidation implements TokenService for validation testing
type MockTokenServiceForValidation struct {
	privateKey       *rsa.PrivateKey
	publicKey        *rsa.PublicKey
	validateFunc     func(ctx context.Context, token string) (*services.TokenClaims, error)
	shouldFailVerify bool
}

func NewMockTokenServiceForValidation() *MockTokenServiceForValidation {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	return &MockTokenServiceForValidation{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}
}

func (m *MockTokenServiceForValidation) SignIDToken(ctx context.Context, claims *services.TokenClaims) (string, error) {
	return m.signToken(claims)
}

func (m *MockTokenServiceForValidation) SignAccessToken(ctx context.Context, claims *services.TokenClaims) (string, error) {
	return m.signToken(claims)
}

func (m *MockTokenServiceForValidation) signToken(claims *services.TokenClaims) (string, error) {
	jwtClaims := jwt.MapClaims{
		"iss":       claims.Issuer,
		"sub":       claims.Subject,
		"aud":       claims.Audience,
		"exp":       claims.ExpiresAt.Unix(),
		"iat":       claims.IssuedAt.Unix(),
		"tenant_id": claims.TenantID,
	}
	if claims.Scope != "" {
		jwtClaims["scope"] = claims.Scope
	}
	if claims.Email != "" {
		jwtClaims["email"] = claims.Email
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtClaims)
	token.Header["kid"] = "test-key-1"
	return token.SignedString(m.privateKey)
}

func (m *MockTokenServiceForValidation) ValidateTokenSignature(ctx context.Context, token string) (*services.TokenClaims, error) {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, token)
	}

	if m.shouldFailVerify {
		return nil, errors.New("signature verification failed")
	}

	// Parse and verify the token
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return m.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	// Parse audience - can be string or []string
	var audience []string
	switch aud := claims["aud"].(type) {
	case string:
		audience = []string{aud}
	case []interface{}:
		for _, a := range aud {
			if s, ok := a.(string); ok {
				audience = append(audience, s)
			}
		}
	}

	return &services.TokenClaims{
		Issuer:    claims["iss"].(string),
		Subject:   claims["sub"].(string),
		Audience:  audience,
		ExpiresAt: time.Unix(int64(claims["exp"].(float64)), 0),
		IssuedAt:  time.Unix(int64(claims["iat"].(float64)), 0),
		TenantID:  claims["tenant_id"].(string),
	}, nil
}

func (m *MockTokenServiceForValidation) GetPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	return m.publicKey, nil
}

func (m *MockTokenServiceForValidation) GetJWKS(ctx context.Context) (*services.JWKS, error) {
	return nil, nil
}

func (m *MockTokenServiceForValidation) GetActiveKeyID(ctx context.Context) (string, error) {
	return "test-key-1", nil
}

var _ services.TokenService = (*MockTokenServiceForValidation)(nil)

// ValidateTokenInput represents input for token validation
type ValidateTokenInput struct {
	Token            string
	ExpectedTenantID string
	ExpectedAudience string
}

// ValidateTokenOutput represents the output of token validation
type ValidateTokenOutput struct {
	Valid  bool
	Claims *services.TokenClaims
	Error  error
}

// ValidateTokenUseCase validates JWT tokens
type ValidateTokenUseCase struct {
	tokenService services.TokenService
}

func NewValidateTokenUseCase(tokenService services.TokenService) *ValidateTokenUseCase {
	return &ValidateTokenUseCase{
		tokenService: tokenService,
	}
}

func (uc *ValidateTokenUseCase) Execute(ctx context.Context, input ValidateTokenInput) (*ValidateTokenOutput, error) {
	// Validate signature and parse claims
	claims, err := uc.tokenService.ValidateTokenSignature(ctx, input.Token)
	if err != nil {
		return &ValidateTokenOutput{
			Valid: false,
			Error: errors.New("invalid token signature"),
		}, nil
	}

	// Check expiration
	if time.Now().After(claims.ExpiresAt) {
		return &ValidateTokenOutput{
			Valid: false,
			Error: errors.New("token expired"),
		}, nil
	}

	// Validate tenant_id if provided
	if input.ExpectedTenantID != "" && claims.TenantID != input.ExpectedTenantID {
		return &ValidateTokenOutput{
			Valid: false,
			Error: errors.New("tenant_id mismatch"),
		}, nil
	}

	// Validate audience if provided
	if input.ExpectedAudience != "" {
		found := false
		for _, aud := range claims.Audience {
			if aud == input.ExpectedAudience {
				found = true
				break
			}
		}
		if !found {
			return &ValidateTokenOutput{
				Valid: false,
				Error: errors.New("audience mismatch"),
			}, nil
		}
	}

	return &ValidateTokenOutput{
		Valid:  true,
		Claims: claims,
		Error:  nil,
	}, nil
}

func TestValidateToken_ValidToken(t *testing.T) {
	tokenService := NewMockTokenServiceForValidation()
	ctx := context.Background()

	// Create a valid token
	claims := &services.TokenClaims{
		Issuer:    "https://sso.test.com",
		Subject:   "user_123",
		Audience:  []string{"client_abc"},
		ExpiresAt: time.Now().Add(15 * time.Minute),
		IssuedAt:  time.Now(),
		TenantID:  "tenant_xyz",
	}
	token, err := tokenService.SignAccessToken(ctx, claims)
	require.NoError(t, err)

	uc := NewValidateTokenUseCase(tokenService)
	output, err := uc.Execute(ctx, ValidateTokenInput{
		Token:            token,
		ExpectedTenantID: "tenant_xyz",
		ExpectedAudience: "client_abc",
	})

	require.NoError(t, err)
	assert.True(t, output.Valid)
	assert.Nil(t, output.Error)
	assert.Equal(t, "user_123", output.Claims.Subject)
	assert.Equal(t, "tenant_xyz", output.Claims.TenantID)
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	tokenService := NewMockTokenServiceForValidation()
	tokenService.shouldFailVerify = true
	ctx := context.Background()

	uc := NewValidateTokenUseCase(tokenService)
	output, err := uc.Execute(ctx, ValidateTokenInput{
		Token: "invalid.token.here",
	})

	require.NoError(t, err)
	assert.False(t, output.Valid)
	assert.Contains(t, output.Error.Error(), "invalid token signature")
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	tokenService := NewMockTokenServiceForValidation()
	ctx := context.Background()

	// Create an expired token
	expiredClaims := &services.TokenClaims{
		Issuer:    "https://sso.test.com",
		Subject:   "user_123",
		Audience:  []string{"client_abc"},
		ExpiresAt: time.Now().Add(-5 * time.Minute), // Expired 5 minutes ago
		IssuedAt:  time.Now().Add(-20 * time.Minute),
		TenantID:  "tenant_xyz",
	}
	token, err := tokenService.SignAccessToken(ctx, expiredClaims)
	require.NoError(t, err)

	// Override validate to return the expired claims without error (to test expiration logic)
	tokenService.validateFunc = func(ctx context.Context, token string) (*services.TokenClaims, error) {
		return expiredClaims, nil
	}

	uc := NewValidateTokenUseCase(tokenService)
	output, err := uc.Execute(ctx, ValidateTokenInput{
		Token: token,
	})

	require.NoError(t, err)
	assert.False(t, output.Valid)
	assert.Contains(t, output.Error.Error(), "token expired")
}

func TestValidateToken_TenantIDMismatch(t *testing.T) {
	tokenService := NewMockTokenServiceForValidation()
	ctx := context.Background()

	// Create a valid token for tenant_xyz
	claims := &services.TokenClaims{
		Issuer:    "https://sso.test.com",
		Subject:   "user_123",
		Audience:  []string{"client_abc"},
		ExpiresAt: time.Now().Add(15 * time.Minute),
		IssuedAt:  time.Now(),
		TenantID:  "tenant_xyz",
	}
	token, err := tokenService.SignAccessToken(ctx, claims)
	require.NoError(t, err)

	uc := NewValidateTokenUseCase(tokenService)
	output, err := uc.Execute(ctx, ValidateTokenInput{
		Token:            token,
		ExpectedTenantID: "different_tenant", // Mismatch
	})

	require.NoError(t, err)
	assert.False(t, output.Valid)
	assert.Contains(t, output.Error.Error(), "tenant_id mismatch")
}

func TestValidateToken_AudienceMismatch(t *testing.T) {
	tokenService := NewMockTokenServiceForValidation()
	ctx := context.Background()

	// Create a valid token
	claims := &services.TokenClaims{
		Issuer:    "https://sso.test.com",
		Subject:   "user_123",
		Audience:  []string{"client_abc"},
		ExpiresAt: time.Now().Add(15 * time.Minute),
		IssuedAt:  time.Now(),
		TenantID:  "tenant_xyz",
	}
	token, err := tokenService.SignAccessToken(ctx, claims)
	require.NoError(t, err)

	uc := NewValidateTokenUseCase(tokenService)
	output, err := uc.Execute(ctx, ValidateTokenInput{
		Token:            token,
		ExpectedAudience: "different_client", // Mismatch
	})

	require.NoError(t, err)
	assert.False(t, output.Valid)
	assert.Contains(t, output.Error.Error(), "audience mismatch")
}

func TestValidateToken_NoTenantValidation(t *testing.T) {
	tokenService := NewMockTokenServiceForValidation()
	ctx := context.Background()

	// Create a valid token
	claims := &services.TokenClaims{
		Issuer:    "https://sso.test.com",
		Subject:   "user_123",
		Audience:  []string{"client_abc"},
		ExpiresAt: time.Now().Add(15 * time.Minute),
		IssuedAt:  time.Now(),
		TenantID:  "tenant_xyz",
	}
	token, err := tokenService.SignAccessToken(ctx, claims)
	require.NoError(t, err)

	uc := NewValidateTokenUseCase(tokenService)
	output, err := uc.Execute(ctx, ValidateTokenInput{
		Token: token,
		// No ExpectedTenantID - should skip tenant validation
	})

	require.NoError(t, err)
	assert.True(t, output.Valid)
	assert.Nil(t, output.Error)
}

func TestValidateToken_NoAudienceValidation(t *testing.T) {
	tokenService := NewMockTokenServiceForValidation()
	ctx := context.Background()

	// Create a valid token
	claims := &services.TokenClaims{
		Issuer:    "https://sso.test.com",
		Subject:   "user_123",
		Audience:  []string{"client_abc"},
		ExpiresAt: time.Now().Add(15 * time.Minute),
		IssuedAt:  time.Now(),
		TenantID:  "tenant_xyz",
	}
	token, err := tokenService.SignAccessToken(ctx, claims)
	require.NoError(t, err)

	uc := NewValidateTokenUseCase(tokenService)
	output, err := uc.Execute(ctx, ValidateTokenInput{
		Token: token,
		// No ExpectedAudience - should skip audience validation
	})

	require.NoError(t, err)
	assert.True(t, output.Valid)
	assert.Nil(t, output.Error)
}

func TestValidateToken_MalformedToken(t *testing.T) {
	tokenService := NewMockTokenServiceForValidation()
	tokenService.validateFunc = func(ctx context.Context, token string) (*services.TokenClaims, error) {
		return nil, errors.New("malformed token")
	}
	ctx := context.Background()

	uc := NewValidateTokenUseCase(tokenService)
	output, err := uc.Execute(ctx, ValidateTokenInput{
		Token: "not-a-valid-jwt",
	})

	require.NoError(t, err)
	assert.False(t, output.Valid)
	assert.Contains(t, output.Error.Error(), "invalid token signature")
}

func TestValidateToken_EmptyToken(t *testing.T) {
	tokenService := NewMockTokenServiceForValidation()
	tokenService.validateFunc = func(ctx context.Context, token string) (*services.TokenClaims, error) {
		return nil, errors.New("empty token")
	}
	ctx := context.Background()

	uc := NewValidateTokenUseCase(tokenService)
	output, err := uc.Execute(ctx, ValidateTokenInput{
		Token: "",
	})

	require.NoError(t, err)
	assert.False(t, output.Valid)
}

func TestValidateToken_MultipleAudiences(t *testing.T) {
	tokenService := NewMockTokenServiceForValidation()
	ctx := context.Background()

	// Create a valid token with multiple audiences
	claims := &services.TokenClaims{
		Issuer:    "https://sso.test.com",
		Subject:   "user_123",
		Audience:  []string{"client_abc", "client_def", "client_ghi"},
		ExpiresAt: time.Now().Add(15 * time.Minute),
		IssuedAt:  time.Now(),
		TenantID:  "tenant_xyz",
	}
	token, err := tokenService.SignAccessToken(ctx, claims)
	require.NoError(t, err)

	uc := NewValidateTokenUseCase(tokenService)

	// Should pass when expected audience is in the list
	output, err := uc.Execute(ctx, ValidateTokenInput{
		Token:            token,
		ExpectedAudience: "client_def",
	})
	require.NoError(t, err)
	assert.True(t, output.Valid)

	// Should fail when expected audience is not in the list
	output2, err := uc.Execute(ctx, ValidateTokenInput{
		Token:            token,
		ExpectedAudience: "client_xyz",
	})
	require.NoError(t, err)
	assert.False(t, output2.Valid)
}
