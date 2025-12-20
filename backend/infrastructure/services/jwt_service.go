/**
 * JWT Service for token generation and validation
 */

package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the claims in the JWT token
type JWTClaims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations
type JWTService struct {
	secretKey       []byte
	expirationHours int
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, expirationHours int) *JWTService {
	return &JWTService{
		secretKey:       []byte(secretKey),
		expirationHours: expirationHours,
	}
}

// GenerateToken generates a new JWT token for a user
func (s *JWTService) GenerateToken(userID, tenantID, email, role string) (string, error) {
	if userID == "" {
		return "", errors.New("user ID is required")
	}
	if tenantID == "" {
		return "", errors.New("tenant ID is required")
	}
	if email == "" {
		return "", errors.New("email is required")
	}
	if role == "" {
		return "", errors.New("role is required")
	}

	expirationTime := time.Now().Add(time.Duration(s.expirationHours) * time.Hour)

	claims := &JWTClaims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "keyles-sso",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	if tokenString == "" {
		return nil, errors.New("token is required")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// GetExpirationDuration returns the token expiration duration in seconds
func (s *JWTService) GetExpirationDuration() int {
	return s.expirationHours * 3600
}
