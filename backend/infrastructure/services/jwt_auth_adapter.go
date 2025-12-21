/**
 * JWT Service Adapter for Auth Use Case
 * Adapts the infrastructure JWT service to the auth use case interface
 */

package services

import (
	"github.com/ranggaaprilio/keyles/usecase/auth"
)

// AuthJWTServiceAdapter adapts JWTService to auth.JWTService interface
type AuthJWTServiceAdapter struct {
	jwtService *JWTService
}

// NewAuthJWTServiceAdapter creates a new adapter
func NewAuthJWTServiceAdapter(jwtService *JWTService) auth.JWTService {
	return &AuthJWTServiceAdapter{jwtService: jwtService}
}

// GenerateToken delegates to the wrapped JWT service
func (a *AuthJWTServiceAdapter) GenerateToken(userID, tenantID, email, role string) (string, error) {
	return a.jwtService.GenerateToken(userID, tenantID, email, role)
}

// ValidateToken delegates to the wrapped JWT service and converts the claims
func (a *AuthJWTServiceAdapter) ValidateToken(token string) (*auth.JWTClaims, error) {
	claims, err := a.jwtService.ValidateToken(token)
	if err != nil {
		return nil, err
	}
	
	return &auth.JWTClaims{
		UserID:   claims.UserID,
		TenantID: claims.TenantID,
		Email:    claims.Email,
		Role:     claims.Role,
	}, nil
}
