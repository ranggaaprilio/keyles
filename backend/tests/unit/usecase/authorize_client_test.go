package usecase_test

import (
"context"
"errors"
"testing"
"time"

"github.com/ranggaaprilio/keyles/domain/entities"
"github.com/ranggaaprilio/keyles/tests/mocks"
"github.com/ranggaaprilio/keyles/usecase/auth"
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/mock"
)

func TestAuthorizeClient_Execute(t *testing.T) {
	ctx := context.Background()

	validRequest := auth.AuthorizeRequest{
		ClientID:            "client_abc123",
		RedirectURI:         "https://app.example.com/callback",
		ResponseType:        "code",
		Scope:               "openid profile email",
		State:               "csrf_token_123",
		CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		CodeChallengeMethod: "S256",
		UserID:              "user_123",
		TenantID:            "tenant_xyz",
	}

	testClient := &entities.Client{
		ClientID:            "client_abc123",
		TenantID:            "tenant_xyz",
		ClientName:          "Test App",
		ClientSecretHash:    "hashed_secret",
		AllowedRedirectURIs: []string{"https://app.example.com/callback", "https://app.example.com/oauth"},
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	activeUser := &entities.User{
		ID: "user_123", TenantID: "tenant_xyz", Status: entities.UserStatusActive,
	}

	tests := []struct {
		name          string
		request       auth.AuthorizeRequest
		setupMocks    func(*mocks.MockClientRepository, *mocks.MockRoleRepository, *mocks.MockAuthCodeRepository, *mocks.MockEndUserRepository)
		wantErr       bool
		errContains   string
		checkResponse func(*testing.T, *auth.AuthorizeResponse)
	}{
		{
			name:    "successful authorization",
			request: validRequest,
			setupMocks: func(cr *mocks.MockClientRepository, rr *mocks.MockRoleRepository, ar *mocks.MockAuthCodeRepository, eur *mocks.MockEndUserRepository) {
				cr.On("GetByClientID", ctx, "client_abc123", "tenant_xyz").Return(testClient, nil)
				eur.On("GetByID", ctx, "user_123").Return(activeUser, nil)
				rr.On("HasAnyRole", ctx, "user_123", "client_abc123").Return(true, nil)
				ar.On("Store", ctx, mock.AnythingOfType("*entities.AuthorizationCode"), 5*time.Minute).Return(nil)
			},
			wantErr: false,
			checkResponse: func(t *testing.T, resp *auth.AuthorizeResponse) {
				assert.NotEmpty(t, resp.Code)
				assert.Equal(t, "csrf_token_123", resp.State)
				assert.Equal(t, "https://app.example.com/callback", resp.RedirectURI)
			},
		},
		{
			name: "disabled user - access denied (FR-028)",
			request: auth.AuthorizeRequest{
				ClientID:            "client_abc123",
				RedirectURI:         "https://app.example.com/callback",
				ResponseType:        "code",
				Scope:               "openid",
				State:               "csrf_token",
				CodeChallenge:       "challenge",
				CodeChallengeMethod: "S256",
				UserID:              "user_disabled",
				TenantID:            "tenant_xyz",
			},
			setupMocks: func(cr *mocks.MockClientRepository, rr *mocks.MockRoleRepository, ar *mocks.MockAuthCodeRepository, eur *mocks.MockEndUserRepository) {
				cr.On("GetByClientID", ctx, "client_abc123", "tenant_xyz").Return(testClient, nil)
				eur.On("GetByID", ctx, "user_disabled").Return(&entities.User{
					ID: "user_disabled", TenantID: "tenant_xyz", Status: entities.UserStatusDisabled,
				}, nil)
			},
			wantErr:     true,
			errContains: "access_denied",
		},
		{
			name: "pending user with active roles - authorization succeeds",
			request: auth.AuthorizeRequest{
				ClientID:            "client_abc123",
				RedirectURI:         "https://app.example.com/callback",
				ResponseType:        "code",
				Scope:               "openid",
				State:               "csrf_token",
				CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
				CodeChallengeMethod: "S256",
				UserID:              "user_pending",
				TenantID:            "tenant_xyz",
			},
			setupMocks: func(cr *mocks.MockClientRepository, rr *mocks.MockRoleRepository, ar *mocks.MockAuthCodeRepository, eur *mocks.MockEndUserRepository) {
				cr.On("GetByClientID", ctx, "client_abc123", "tenant_xyz").Return(testClient, nil)
				eur.On("GetByID", ctx, "user_pending").Return(&entities.User{
					ID: "user_pending", TenantID: "tenant_xyz", Status: entities.UserStatusPending,
				}, nil)
				rr.On("HasAnyRole", ctx, "user_pending", "client_abc123").Return(true, nil)
				ar.On("Store", ctx, mock.AnythingOfType("*entities.AuthorizationCode"), 5*time.Minute).Return(nil)
			},
			wantErr: false,
			checkResponse: func(t *testing.T, resp *auth.AuthorizeResponse) {
				assert.NotEmpty(t, resp.Code)
			},
		},
		{
			name: "user not found in EndUserRepo - flow continues to role check",
			request: auth.AuthorizeRequest{
				ClientID:            "client_abc123",
				RedirectURI:         "https://app.example.com/callback",
				ResponseType:        "code",
				Scope:               "openid",
				State:               "csrf_token",
				CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
				CodeChallengeMethod: "S256",
				UserID:              "user_admin",
				TenantID:            "tenant_xyz",
			},
			setupMocks: func(cr *mocks.MockClientRepository, rr *mocks.MockRoleRepository, ar *mocks.MockAuthCodeRepository, eur *mocks.MockEndUserRepository) {
				cr.On("GetByClientID", ctx, "client_abc123", "tenant_xyz").Return(testClient, nil)
				eur.On("GetByID", ctx, "user_admin").Return(nil, errors.New("not found"))
				rr.On("HasAnyRole", ctx, "user_admin", "client_abc123").Return(true, nil)
				ar.On("Store", ctx, mock.AnythingOfType("*entities.AuthorizationCode"), 5*time.Minute).Return(nil)
			},
			wantErr: false,
			checkResponse: func(t *testing.T, resp *auth.AuthorizeResponse) {
				assert.NotEmpty(t, resp.Code)
			},
		},
		{
			name: "invalid client_id - client not found",
			request: auth.AuthorizeRequest{
				ClientID:            "invalid_client",
				RedirectURI:         "https://app.example.com/callback",
				ResponseType:        "code",
				Scope:               "openid",
				State:               "csrf_token",
				CodeChallenge:       "challenge",
				CodeChallengeMethod: "S256",
				UserID:              "user_123",
				TenantID:            "tenant_xyz",
			},
			setupMocks: func(cr *mocks.MockClientRepository, rr *mocks.MockRoleRepository, ar *mocks.MockAuthCodeRepository, eur *mocks.MockEndUserRepository) {
				cr.On("GetByClientID", ctx, "invalid_client", "tenant_xyz").Return(nil, errors.New("client not found"))
			},
			wantErr:     true,
			errContains: "invalid_client",
		},
		{
			name: "invalid redirect_uri - not in allowed list",
			request: auth.AuthorizeRequest{
				ClientID:            "client_abc123",
				RedirectURI:         "https://malicious.com/callback",
				ResponseType:        "code",
				Scope:               "openid",
				State:               "csrf_token",
				CodeChallenge:       "challenge",
				CodeChallengeMethod: "S256",
				UserID:              "user_123",
				TenantID:            "tenant_xyz",
			},
			setupMocks: func(cr *mocks.MockClientRepository, rr *mocks.MockRoleRepository, ar *mocks.MockAuthCodeRepository, eur *mocks.MockEndUserRepository) {
				cr.On("GetByClientID", ctx, "client_abc123", "tenant_xyz").Return(testClient, nil)
			},
			wantErr:     true,
			errContains: "redirect_uri",
		},
		{
			name: "user without role - access denied",
			request: auth.AuthorizeRequest{
				ClientID:            "client_abc123",
				RedirectURI:         "https://app.example.com/callback",
				ResponseType:        "code",
				Scope:               "openid",
				State:               "csrf_token",
				CodeChallenge:       "challenge",
				CodeChallengeMethod: "S256",
				UserID:              "user_no_role",
				TenantID:            "tenant_xyz",
			},
			setupMocks: func(cr *mocks.MockClientRepository, rr *mocks.MockRoleRepository, ar *mocks.MockAuthCodeRepository, eur *mocks.MockEndUserRepository) {
				cr.On("GetByClientID", ctx, "client_abc123", "tenant_xyz").Return(testClient, nil)
				eur.On("GetByID", ctx, "user_no_role").Return(&entities.User{
					ID: "user_no_role", TenantID: "tenant_xyz", Status: entities.UserStatusActive,
				}, nil)
				rr.On("HasAnyRole", ctx, "user_no_role", "client_abc123").Return(false, nil)
			},
			wantErr:     true,
			errContains: "access_denied",
		},
		{
			name: "invalid response_type",
			request: auth.AuthorizeRequest{
				ClientID:            "client_abc123",
				RedirectURI:         "https://app.example.com/callback",
				ResponseType:        "token",
				Scope:               "openid",
				State:               "csrf_token",
				CodeChallenge:       "challenge",
				CodeChallengeMethod: "S256",
				UserID:              "user_123",
				TenantID:            "tenant_xyz",
			},
			setupMocks: func(cr *mocks.MockClientRepository, rr *mocks.MockRoleRepository, ar *mocks.MockAuthCodeRepository, eur *mocks.MockEndUserRepository) {},
			wantErr:     true,
			errContains: "unsupported_response_type",
		},
		{
			name: "missing code_challenge - PKCE required",
			request: auth.AuthorizeRequest{
				ClientID:            "client_abc123",
				RedirectURI:         "https://app.example.com/callback",
				ResponseType:        "code",
				Scope:               "openid",
				State:               "csrf_token",
				CodeChallenge:       "",
				CodeChallengeMethod: "S256",
				UserID:              "user_123",
				TenantID:            "tenant_xyz",
			},
			setupMocks: func(cr *mocks.MockClientRepository, rr *mocks.MockRoleRepository, ar *mocks.MockAuthCodeRepository, eur *mocks.MockEndUserRepository) {},
			wantErr:     true,
			errContains: "invalid_request",
		},
		{
			name: "invalid code_challenge_method",
			request: auth.AuthorizeRequest{
				ClientID:            "client_abc123",
				RedirectURI:         "https://app.example.com/callback",
				ResponseType:        "code",
				Scope:               "openid",
				State:               "csrf_token",
				CodeChallenge:       "challenge",
				CodeChallengeMethod: "plain",
				UserID:              "user_123",
				TenantID:            "tenant_xyz",
			},
			setupMocks: func(cr *mocks.MockClientRepository, rr *mocks.MockRoleRepository, ar *mocks.MockAuthCodeRepository, eur *mocks.MockEndUserRepository) {},
			wantErr:     true,
			errContains: "invalid_request",
		},
		{
			name: "missing state parameter",
			request: auth.AuthorizeRequest{
				ClientID:            "client_abc123",
				RedirectURI:         "https://app.example.com/callback",
				ResponseType:        "code",
				Scope:               "openid",
				State:               "",
				CodeChallenge:       "challenge",
				CodeChallengeMethod: "S256",
				UserID:              "user_123",
				TenantID:            "tenant_xyz",
			},
			setupMocks: func(cr *mocks.MockClientRepository, rr *mocks.MockRoleRepository, ar *mocks.MockAuthCodeRepository, eur *mocks.MockEndUserRepository) {},
			wantErr:     true,
			errContains: "invalid_request",
		},
		{
			name: "inactive client",
			request: auth.AuthorizeRequest{
				ClientID:            "client_abc123",
				RedirectURI:         "https://app.example.com/callback",
				ResponseType:        "code",
				Scope:               "openid",
				State:               "csrf_token",
				CodeChallenge:       "challenge",
				CodeChallengeMethod: "S256",
				UserID:              "user_123",
				TenantID:            "tenant_xyz",
			},
			setupMocks: func(cr *mocks.MockClientRepository, rr *mocks.MockRoleRepository, ar *mocks.MockAuthCodeRepository, eur *mocks.MockEndUserRepository) {
				inactiveClient := *testClient
				inactiveClient.IsActive = false
				cr.On("GetByClientID", ctx, "client_abc123", "tenant_xyz").Return(&inactiveClient, nil)
			},
			wantErr:     true,
			errContains: "unauthorized_client",
		},
		{
			name: "scope must include openid",
			request: auth.AuthorizeRequest{
				ClientID:            "client_abc123",
				RedirectURI:         "https://app.example.com/callback",
				ResponseType:        "code",
				Scope:               "profile email",
				State:               "csrf_token",
				CodeChallenge:       "challenge",
				CodeChallengeMethod: "S256",
				UserID:              "user_123",
				TenantID:            "tenant_xyz",
			},
			setupMocks: func(cr *mocks.MockClientRepository, rr *mocks.MockRoleRepository, ar *mocks.MockAuthCodeRepository, eur *mocks.MockEndUserRepository) {},
			wantErr:     true,
			errContains: "invalid_scope",
		},
		{
			name:    "auth code store failure",
			request: validRequest,
			setupMocks: func(cr *mocks.MockClientRepository, rr *mocks.MockRoleRepository, ar *mocks.MockAuthCodeRepository, eur *mocks.MockEndUserRepository) {
				cr.On("GetByClientID", ctx, "client_abc123", "tenant_xyz").Return(testClient, nil)
				eur.On("GetByID", ctx, "user_123").Return(activeUser, nil)
				rr.On("HasAnyRole", ctx, "user_123", "client_abc123").Return(true, nil)
				ar.On("Store", ctx, mock.AnythingOfType("*entities.AuthorizationCode"), 5*time.Minute).Return(errors.New("redis connection failed"))
			},
			wantErr:     true,
			errContains: "server_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
mockClientRepo := new(mocks.MockClientRepository)
mockRoleRepo := new(mocks.MockRoleRepository)
mockAuthCodeRepo := new(mocks.MockAuthCodeRepository)
mockEndUserRepo := new(mocks.MockEndUserRepository)

if tt.setupMocks != nil {
tt.setupMocks(mockClientRepo, mockRoleRepo, mockAuthCodeRepo, mockEndUserRepo)
}

uc := auth.NewAuthorizeClient(mockClientRepo, mockRoleRepo, mockAuthCodeRepo, mockEndUserRepo)
resp, err := uc.Execute(ctx, tt.request)

if tt.wantErr {
assert.Error(t, err)
if tt.errContains != "" {
assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.checkResponse != nil {
					tt.checkResponse(t, resp)
				}
			}

			mockClientRepo.AssertExpectations(t)
			mockRoleRepo.AssertExpectations(t)
			mockAuthCodeRepo.AssertExpectations(t)
			mockEndUserRepo.AssertExpectations(t)
		})
	}
}

func TestAuthorizeClient_ValidateRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     auth.AuthorizeRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "valid request",
			request: auth.AuthorizeRequest{
				ClientID:            "client_123",
				RedirectURI:         "https://app.example.com/callback",
				ResponseType:        "code",
				Scope:               "openid profile",
				State:               "csrf_token",
				CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
				CodeChallengeMethod: "S256",
				UserID:              "user_123",
				TenantID:            "tenant_xyz",
			},
			wantErr: false,
		},
		{
			name: "missing client_id",
			request: auth.AuthorizeRequest{
				ClientID:            "",
				RedirectURI:         "https://app.example.com/callback",
				ResponseType:        "code",
				Scope:               "openid",
				State:               "csrf_token",
				CodeChallenge:       "challenge",
				CodeChallengeMethod: "S256",
			},
			wantErr:     true,
			errContains: "client_id",
		},
		{
			name: "missing redirect_uri",
			request: auth.AuthorizeRequest{
				ClientID:            "client_123",
				RedirectURI:         "",
				ResponseType:        "code",
				Scope:               "openid",
				State:               "csrf_token",
				CodeChallenge:       "challenge",
				CodeChallengeMethod: "S256",
			},
			wantErr:     true,
			errContains: "redirect_uri",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
err := auth.ValidateAuthorizeRequest(tt.request)
if tt.wantErr {
assert.Error(t, err)
if tt.errContains != "" {
assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthorizeClient_PKCEValidation(t *testing.T) {
	tests := []struct {
		name                string
		codeChallenge       string
		codeChallengeMethod string
		wantErr             bool
	}{
		{
			name:                "valid S256",
			codeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			codeChallengeMethod: "S256",
			wantErr:             false,
		},
		{
			name:                "missing code_challenge",
			codeChallenge:       "",
			codeChallengeMethod: "S256",
			wantErr:             true,
		},
		{
			name:                "plain method not allowed",
			codeChallenge:       "verifier",
			codeChallengeMethod: "plain",
			wantErr:             true,
		},
		{
			name:                "missing method",
			codeChallenge:       "challenge",
			codeChallengeMethod: "",
			wantErr:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
err := auth.ValidatePKCE(tt.codeChallenge, tt.codeChallengeMethod)
if tt.wantErr {
assert.Error(t, err)
} else {
assert.NoError(t, err)
}
})
	}
}
