package auth_test

import (
	"context"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/services"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type validatorTokenService struct {
	claims *services.TokenClaims
	err    error
}

func (s *validatorTokenService) SignIDToken(context.Context, *services.TokenClaims) (string, error) {
	return "", nil
}
func (s *validatorTokenService) SignAccessToken(context.Context, *services.TokenClaims) (string, error) {
	return "", nil
}
func (s *validatorTokenService) ValidateTokenSignature(context.Context, string) (*services.TokenClaims, error) {
	return s.claims, s.err
}
func (s *validatorTokenService) GetPublicKey(context.Context, string) (*rsa.PublicKey, error) {
	return nil, nil
}
func (s *validatorTokenService) GetJWKS(context.Context) (*services.JWKS, error) {
	return nil, nil
}
func (s *validatorTokenService) GetActiveKeyID(context.Context) (string, error) {
	return "", nil
}

func setupValidator(t *testing.T, userStatus entities.UserStatus, clientRevoked bool) (*auth.AccessTokenValidator, *mocks.MockClientRepository) {
	t.Helper()
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	userRepo := new(mocks.MockEndUserRepository)
	revokedClients := new(mocks.MockRevokedClientCache)
	blacklist := new(mocks.MockUserBlacklist)
	client := &entities.Client{ClientID: "client", TenantID: "tenant", ClientType: entities.ClientTypeConfidential, IsActive: true}

	clientRepo.On("GetByID", ctx, "client").Return(client, nil)
	userRepo.On("GetByID", ctx, "user").Return(&entities.User{ID: "user", TenantID: "tenant", Status: userStatus}, nil)
	revokedClients.On("IsRevoked", ctx, "client").Return(clientRevoked, nil)
	if !clientRevoked {
		blacklist.On("IsBlacklisted", ctx, "user").Return(false, nil)
	}

	tokenService := &validatorTokenService{claims: &services.TokenClaims{
		Issuer: "https://issuer.example", Subject: "user", ClientID: "client", TenantID: "tenant",
		Audience: []string{"client"}, IssuedAt: time.Now().Add(-time.Minute), ExpiresAt: time.Now().Add(time.Minute),
		Roles: []string{"reader"},
	}}
	return auth.NewAccessTokenValidator(tokenService, clientRepo, userRepo, revokedClients, blacklist, "https://issuer.example"), clientRepo
}

func TestAccessTokenValidatorRejectsRevokedClient(t *testing.T) {
	validator, _ := setupValidator(t, entities.UserStatusActive, true)
	_, err := validator.Validate(context.Background(), "token", "")
	require.ErrorIs(t, err, auth.ErrInactiveAccessToken)
}

func TestAccessTokenValidatorRejectsDisabledUser(t *testing.T) {
	validator, _ := setupValidator(t, entities.UserStatusDisabled, false)
	_, err := validator.Validate(context.Background(), "token", "")
	require.ErrorIs(t, err, auth.ErrInactiveAccessToken)
}

func TestIntrospectionReturnsOwnedTokenMetadata(t *testing.T) {
	ctx := context.Background()
	validator, clientRepo := setupValidator(t, entities.UserStatusActive, false)
	clientRepo.On("ValidateCredentials", ctx, "client", "secret").Return(&entities.Client{
		ClientID: "client", TenantID: "tenant", ClientType: entities.ClientTypeConfidential, IsActive: true,
	}, nil)

	response, err := auth.NewIntrospectToken(clientRepo, validator).Execute(ctx, "token", "client", "secret")
	require.NoError(t, err)
	require.True(t, response.Active)
	require.Equal(t, []string{"reader"}, response.Roles)
}

func TestIntrospectionRejectsInvalidCallerCredentials(t *testing.T) {
	ctx := context.Background()
	validator, clientRepo := setupValidator(t, entities.UserStatusActive, false)
	clientRepo.On("ValidateCredentials", mock.Anything, "client", "wrong").Return(nil, errors.New("invalid"))

	response, err := auth.NewIntrospectToken(clientRepo, validator).Execute(ctx, "token", "client", "wrong")
	require.Nil(t, response)
	require.Error(t, err)
}

func TestAuthenticateOAuthClientAllowsPublicClientWithoutSecret(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "public-client").Return(&entities.Client{
		ClientID: "public-client", ClientType: entities.ClientTypePublic, IsActive: true,
	}, nil)

	client, err := auth.AuthenticateOAuthClient(ctx, clientRepo, "public-client", "", true)
	require.NoError(t, err)
	require.Equal(t, entities.ClientTypePublic, client.ClientType)
}
