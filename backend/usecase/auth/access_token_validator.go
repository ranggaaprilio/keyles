package auth

import (
	"context"
	"errors"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
)

var ErrInactiveAccessToken = errors.New("access token is inactive")

// AccessTokenValidator validates signed OAuth access tokens against live
// account and client revocation state.
type AccessTokenValidator struct {
	tokenService       services.TokenService
	clientRepo         repositories.ClientRepository
	endUserRepo        repositories.EndUserRepository
	revokedClientCache services.RevokedClientCache
	userBlacklist      services.UserBlacklist
	issuer             string
}

func NewAccessTokenValidator(
	tokenService services.TokenService,
	clientRepo repositories.ClientRepository,
	endUserRepo repositories.EndUserRepository,
	revokedClientCache services.RevokedClientCache,
	userBlacklist services.UserBlacklist,
	issuer string,
) *AccessTokenValidator {
	return &AccessTokenValidator{
		tokenService:       tokenService,
		clientRepo:         clientRepo,
		endUserRepo:        endUserRepo,
		revokedClientCache: revokedClientCache,
		userBlacklist:      userBlacklist,
		issuer:             issuer,
	}
}

func (v *AccessTokenValidator) Validate(ctx context.Context, token, expectedClientID string) (*services.TokenClaims, error) {
	claims, err := v.tokenService.ValidateTokenSignature(ctx, token)
	if err != nil {
		return nil, ErrInactiveAccessToken
	}

	now := time.Now()
	if claims.Issuer != v.issuer || claims.Subject == "" || claims.ClientID == "" || claims.TenantID == "" ||
		claims.ExpiresAt.IsZero() || claims.IssuedAt.IsZero() || !now.Before(claims.ExpiresAt) ||
		(!claims.NotBefore.IsZero() && now.Before(claims.NotBefore)) ||
		(!claims.IssuedAt.IsZero() && now.Before(claims.IssuedAt)) ||
		!containsAudience(claims.Audience, claims.ClientID) {
		return nil, ErrInactiveAccessToken
	}
	if expectedClientID != "" && (claims.ClientID != expectedClientID || !containsAudience(claims.Audience, expectedClientID)) {
		return nil, ErrInactiveAccessToken
	}

	client, err := v.clientRepo.GetByID(ctx, claims.ClientID)
	if err != nil || client == nil || !client.IsEnabled() || client.TenantID != claims.TenantID {
		return nil, ErrInactiveAccessToken
	}
	if v.revokedClientCache != nil {
		revoked, err := v.revokedClientCache.IsRevoked(ctx, claims.ClientID)
		if err != nil || revoked {
			return nil, ErrInactiveAccessToken
		}
	}

	user, err := v.endUserRepo.GetByID(ctx, claims.Subject)
	if err != nil || user == nil || user.TenantID != claims.TenantID || user.Status == entities.UserStatusDisabled {
		return nil, ErrInactiveAccessToken
	}
	if v.userBlacklist != nil {
		revoked, err := v.userBlacklist.IsBlacklisted(ctx, claims.Subject)
		if err != nil || revoked {
			return nil, ErrInactiveAccessToken
		}
	}
	return claims, nil
}

func containsAudience(audiences []string, expected string) bool {
	for _, audience := range audiences {
		if audience == expected {
			return true
		}
	}
	return false
}

type IntrospectionResponse struct {
	Active   bool     `json:"active"`
	ClientID string   `json:"client_id,omitempty"`
	Sub      string   `json:"sub,omitempty"`
	Scope    string   `json:"scope,omitempty"`
	Aud      []string `json:"aud,omitempty"`
	Iss      string   `json:"iss,omitempty"`
	Exp      int64    `json:"exp,omitempty"`
	Iat      int64    `json:"iat,omitempty"`
	TenantID string   `json:"tenant_id,omitempty"`
	Roles    []string `json:"roles,omitempty"`
}

type IntrospectToken struct {
	clientRepo repositories.ClientRepository
	validator  *AccessTokenValidator
}

func NewIntrospectToken(clientRepo repositories.ClientRepository, validator *AccessTokenValidator) *IntrospectToken {
	return &IntrospectToken{clientRepo: clientRepo, validator: validator}
}

func (uc *IntrospectToken) Execute(ctx context.Context, token, clientID, clientSecret string) (*IntrospectionResponse, error) {
	if token == "" {
		return nil, &OAuthError{Code: ErrInvalidRequest, Description: "token is required"}
	}
	if _, err := authenticateOAuthClient(ctx, uc.clientRepo, clientID, clientSecret, false); err != nil {
		return nil, err
	}
	claims, err := uc.validator.Validate(ctx, token, clientID)
	if err != nil {
		return &IntrospectionResponse{Active: false}, nil
	}
	return &IntrospectionResponse{
		Active:   true,
		ClientID: claims.ClientID,
		Sub:      claims.Subject,
		Scope:    claims.Scope,
		Aud:      claims.Audience,
		Iss:      claims.Issuer,
		Exp:      claims.ExpiresAt.Unix(),
		Iat:      claims.IssuedAt.Unix(),
		TenantID: claims.TenantID,
		Roles:    claims.Roles,
	}, nil
}
