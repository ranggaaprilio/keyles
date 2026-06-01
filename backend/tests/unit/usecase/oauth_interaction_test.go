package usecase_test

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const interactionFrontendURL = "https://sso.example.com"

func newTestOAuthInteraction(
	clientRepo *mocks.MockClientRepository,
	roleRepo *mocks.MockRoleRepository,
	txnRepo *mocks.MockAuthorizationTransactionRepository,
	sessionRepo *mocks.MockSessionRepository,
	endUserRepo *mocks.MockEndUserRepository,
	auditRepo *mocks.MockAuditRepository,
) *auth.OAuthInteraction {
	auditHelper := auth.NewOAuthAuditHelper(auditRepo)
	return auth.NewOAuthInteraction(
		clientRepo,
		roleRepo,
		txnRepo,
		sessionRepo,
		endUserRepo,
		auditHelper,
		interactionFrontendURL,
	)
}

func validInitializeRequest() auth.InitializeAuthRequest {
	return auth.InitializeAuthRequest{
		ClientID:            "client_abc123",
		RedirectURI:         "https://app.example.com/callback",
		ResponseType:        "code",
		Scope:               "openid profile email",
		State:               "csrf_state_123",
		CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		CodeChallengeMethod: "S256",
		Prompt:              "",
		MaxAge:              nil,
		SessionCookie:       "",
		SourceIP:            "192.168.1.1",
	}
}

func validClient() *entities.Client {
	return &entities.Client{
		ClientID:            "client_abc123",
		TenantID:            "tenant_xyz",
		ClientName:          "Test App",
		ClientType:          entities.ClientTypeConfidential,
		AllowedRedirectURIs: []string{"https://app.example.com/callback"},
		IsActive:            true,
	}
}

func activeUser() *entities.User {
	return &entities.User{
		ID:          "user_123",
		TenantID:    "tenant_xyz",
		Email:       "test@example.com",
		Status:      entities.UserStatusActive,
		DisplayName: "Test User",
	}
}

func validSession() *repositories.Session {
	now := time.Now()
	return &repositories.Session{
		SessionID:       "sess_active123",
		UserID:          "user_123",
		TenantID:        "tenant_xyz",
		AuthenticatedAt: now,
		CreatedAt:       now,
		ExpiresAt:       now.Add(8 * time.Hour),
	}
}

// --- Validation Tests ---

func TestInitializeAuth_MissingClientID(t *testing.T) {
	ctx := context.Background()
	uc := newTestOAuthInteraction(
		new(mocks.MockClientRepository),
		new(mocks.MockRoleRepository),
		new(mocks.MockAuthorizationTransactionRepository),
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()
	req.ClientID = ""

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.IsLocalError)
	assert.Contains(t, resp.RedirectURL, "/oauth2/error")
	assert.Contains(t, resp.RedirectURL, "client_id")
}

func TestInitializeAuth_MissingRedirectURI(t *testing.T) {
	ctx := context.Background()
	uc := newTestOAuthInteraction(
		new(mocks.MockClientRepository),
		new(mocks.MockRoleRepository),
		new(mocks.MockAuthorizationTransactionRepository),
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()
	req.RedirectURI = ""

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.IsLocalError)
	assert.Contains(t, resp.RedirectURL, "redirect_uri")
}

func TestInitializeAuth_InvalidResponseType(t *testing.T) {
	ctx := context.Background()
	uc := newTestOAuthInteraction(
		new(mocks.MockClientRepository),
		new(mocks.MockRoleRepository),
		new(mocks.MockAuthorizationTransactionRepository),
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()
	req.ResponseType = "token"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.IsLocalError)
	assert.Contains(t, resp.RedirectURL, "unsupported_response_type")
}

func TestInitializeAuth_MissingState(t *testing.T) {
	ctx := context.Background()
	uc := newTestOAuthInteraction(
		new(mocks.MockClientRepository),
		new(mocks.MockRoleRepository),
		new(mocks.MockAuthorizationTransactionRepository),
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()
	req.State = ""

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.IsLocalError)
	assert.Contains(t, resp.RedirectURL, "state")
}

func TestInitializeAuth_MissingCodeChallenge(t *testing.T) {
	ctx := context.Background()
	uc := newTestOAuthInteraction(
		new(mocks.MockClientRepository),
		new(mocks.MockRoleRepository),
		new(mocks.MockAuthorizationTransactionRepository),
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()
	req.CodeChallenge = ""

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.IsLocalError)
	assert.Contains(t, resp.RedirectURL, "code_challenge")
}

func TestInitializeAuth_InvalidCodeChallengeMethod(t *testing.T) {
	ctx := context.Background()
	uc := newTestOAuthInteraction(
		new(mocks.MockClientRepository),
		new(mocks.MockRoleRepository),
		new(mocks.MockAuthorizationTransactionRepository),
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()
	req.CodeChallengeMethod = "plain"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.IsLocalError)
	assert.Contains(t, resp.RedirectURL, "S256")
}

func TestInitializeAuth_MissingOpenIDScope(t *testing.T) {
	ctx := context.Background()
	uc := newTestOAuthInteraction(
		new(mocks.MockClientRepository),
		new(mocks.MockRoleRepository),
		new(mocks.MockAuthorizationTransactionRepository),
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()
	req.Scope = "profile email"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.IsLocalError)
	assert.Contains(t, resp.RedirectURL, "openid")
}

// --- Client Lookup Tests ---

func TestInitializeAuth_ClientNotFound(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_nonexistent").Return(nil, errors.New("not found"))

	uc := newTestOAuthInteraction(
		clientRepo,
		new(mocks.MockRoleRepository),
		new(mocks.MockAuthorizationTransactionRepository),
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()
	req.ClientID = "client_nonexistent"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.IsLocalError)
	assert.Contains(t, resp.RedirectURL, "invalid_client")
}

func TestInitializeAuth_InactiveClient(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	inactiveClient := validClient()
	inactiveClient.IsActive = false
	clientRepo.On("GetByID", ctx, "client_abc123").Return(inactiveClient, nil)

	uc := newTestOAuthInteraction(
		clientRepo,
		new(mocks.MockRoleRepository),
		new(mocks.MockAuthorizationTransactionRepository),
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.IsLocalError)
	assert.Contains(t, resp.RedirectURL, "invalid_client")
}

// --- Redirect URI Validation ---

func TestInitializeAuth_InvalidRedirectURI(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)
	auditRepo := new(mocks.MockAuditRepository)
	auditRepo.On("Create", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	uc := newTestOAuthInteraction(
		clientRepo,
		new(mocks.MockRoleRepository),
		new(mocks.MockAuthorizationTransactionRepository),
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		auditRepo,
	)

	req := validInitializeRequest()
	req.RedirectURI = "https://evil.example.com/callback"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.IsLocalError)
	assert.Contains(t, resp.RedirectURL, "invalid_redirect_uri")
}

// --- Prompt Parameter Tests ---

func TestInitializeAuth_PromptNoneWithOtherValues(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	uc := newTestOAuthInteraction(
		clientRepo,
		new(mocks.MockRoleRepository),
		new(mocks.MockAuthorizationTransactionRepository),
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()
	req.Prompt = "none login"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.False(t, resp.IsLocalError)
	// Should redirect to the *validated* callback URI with error params
	assert.Contains(t, resp.RedirectURL, "https://app.example.com/callback")
	parsed, parseErr := url.Parse(resp.RedirectURL)
	require.NoError(t, parseErr)
	assert.Equal(t, "invalid_request", parsed.Query().Get("error"))
	assert.Equal(t, "csrf_state_123", parsed.Query().Get("state"))
}

func TestInitializeAuth_PromptSelectAccount(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	uc := newTestOAuthInteraction(
		clientRepo,
		new(mocks.MockRoleRepository),
		new(mocks.MockAuthorizationTransactionRepository),
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()
	req.Prompt = "select_account"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.False(t, resp.IsLocalError)
	parsed, parseErr := url.Parse(resp.RedirectURL)
	require.NoError(t, parseErr)
	assert.Equal(t, "unsupported_response_type", parsed.Query().Get("error"))
	assert.Equal(t, "csrf_state_123", parsed.Query().Get("state"))
}

// --- Transaction Creation & Login Redirect ---

func TestInitializeAuth_NoSession_RedirectsToLogin(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Create", ctx, mock.AnythingOfType("*repositories.AuthorizationTransaction"), auth.TransactionTTL).Return(nil)

	uc := newTestOAuthInteraction(
		clientRepo,
		new(mocks.MockRoleRepository),
		txnRepo,
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.False(t, resp.IsLocalError)
	assert.Contains(t, resp.RedirectURL, "/oauth2/login")
	assert.Contains(t, resp.RedirectURL, "transaction_id=")

	// Verify transaction was stored with correct fields
	createCall := txnRepo.Calls[0]
	txnArg := createCall.Arguments.Get(1).(*repositories.AuthorizationTransaction)
	assert.Equal(t, "client_abc123", txnArg.ClientID)
	assert.Equal(t, "tenant_xyz", txnArg.TenantID)
	assert.Equal(t, repositories.TransactionStagePendingLogin, txnArg.Stage)
	assert.NotEmpty(t, txnArg.TransactionID)
	assert.NotEmpty(t, txnArg.InteractionCSRFToken)
}

func TestInitializeAuth_TransactionCreationFail_LocalError(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Create", ctx, mock.AnythingOfType("*repositories.AuthorizationTransaction"), auth.TransactionTTL).Return(errors.New("redis down"))

	uc := newTestOAuthInteraction(
		clientRepo,
		new(mocks.MockRoleRepository),
		txnRepo,
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.IsLocalError)
	assert.Contains(t, resp.RedirectURL, "/oauth2/error")
	assert.Contains(t, resp.RedirectURL, "temporarily_unavailable")
}

// --- Session Reuse Tests ---

func TestInitializeAuth_EligibleSession_RedirectsToConsent(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Create", ctx, mock.AnythingOfType("*repositories.AuthorizationTransaction"), auth.TransactionTTL).Return(nil)
	txnRepo.On("UpdateStage", ctx, mock.AnythingOfType("string"), repositories.TransactionStagePendingConsent, "user_123", "sess_active123").Return(nil)

	sessionRepo := new(mocks.MockSessionRepository)
	sessionRepo.On("Get", ctx, "sess_active123").Return(validSession(), nil)

	endUserRepo := new(mocks.MockEndUserRepository)
	endUserRepo.On("GetByID", ctx, "user_123").Return(activeUser(), nil)

	roleRepo := new(mocks.MockRoleRepository)
	roleRepo.On("HasAnyRole", ctx, "user_123", "client_abc123").Return(true, nil)

	uc := newTestOAuthInteraction(clientRepo, roleRepo, txnRepo, sessionRepo, endUserRepo, new(mocks.MockAuditRepository))

	req := validInitializeRequest()
	req.SessionCookie = "sess_active123"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.False(t, resp.IsLocalError)
	assert.Contains(t, resp.RedirectURL, "/oauth2/consent")
	assert.Contains(t, resp.RedirectURL, "transaction_id=")
}

func TestInitializeAuth_SessionDifferentTenant_NoReuse(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Create", ctx, mock.AnythingOfType("*repositories.AuthorizationTransaction"), auth.TransactionTTL).Return(nil)

	wrongTenantSession := validSession()
	wrongTenantSession.TenantID = "tenant_other"

	sessionRepo := new(mocks.MockSessionRepository)
	sessionRepo.On("Get", ctx, "sess_other").Return(wrongTenantSession, nil)

	uc := newTestOAuthInteraction(
		clientRepo,
		new(mocks.MockRoleRepository),
		txnRepo,
		sessionRepo,
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()
	req.SessionCookie = "sess_other"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.False(t, resp.IsLocalError)
	assert.Contains(t, resp.RedirectURL, "/oauth2/login")
}

func TestInitializeAuth_DisabledUser_NoReuse(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Create", ctx, mock.AnythingOfType("*repositories.AuthorizationTransaction"), auth.TransactionTTL).Return(nil)

	disabledUser := activeUser()
	disabledUser.Status = entities.UserStatusDisabled

	sessionRepo := new(mocks.MockSessionRepository)
	sessionRepo.On("Get", ctx, "sess_disabled").Return(validSession(), nil)

	endUserRepo := new(mocks.MockEndUserRepository)
	endUserRepo.On("GetByID", ctx, "user_123").Return(disabledUser, nil)

	uc := newTestOAuthInteraction(
		clientRepo,
		new(mocks.MockRoleRepository),
		txnRepo,
		sessionRepo,
		endUserRepo,
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()
	req.SessionCookie = "sess_disabled"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.Contains(t, resp.RedirectURL, "/oauth2/login")
}

func TestInitializeAuth_UserNoRole_NoReuse(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Create", ctx, mock.AnythingOfType("*repositories.AuthorizationTransaction"), auth.TransactionTTL).Return(nil)

	sessionRepo := new(mocks.MockSessionRepository)
	sessionRepo.On("Get", ctx, "sess_norole").Return(validSession(), nil)

	endUserRepo := new(mocks.MockEndUserRepository)
	endUserRepo.On("GetByID", ctx, "user_123").Return(activeUser(), nil)

	roleRepo := new(mocks.MockRoleRepository)
	roleRepo.On("HasAnyRole", ctx, "user_123", "client_abc123").Return(false, nil)

	uc := newTestOAuthInteraction(clientRepo, roleRepo, txnRepo, sessionRepo, endUserRepo, new(mocks.MockAuditRepository))

	req := validInitializeRequest()
	req.SessionCookie = "sess_norole"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.Contains(t, resp.RedirectURL, "/oauth2/login")
}

// --- prompt=none Tests ---

func TestInitializeAuth_PromptNone_ReturnsConsentRequired(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Create", ctx, mock.AnythingOfType("*repositories.AuthorizationTransaction"), auth.TransactionTTL).Return(nil)

	sessionRepo := new(mocks.MockSessionRepository)
	sessionRepo.On("Get", ctx, "sess_active123").Return(validSession(), nil)

	endUserRepo := new(mocks.MockEndUserRepository)
	endUserRepo.On("GetByID", ctx, "user_123").Return(activeUser(), nil)

	roleRepo := new(mocks.MockRoleRepository)
	roleRepo.On("HasAnyRole", ctx, "user_123", "client_abc123").Return(true, nil)

	uc := newTestOAuthInteraction(clientRepo, roleRepo, txnRepo, sessionRepo, endUserRepo, new(mocks.MockAuditRepository))

	req := validInitializeRequest()
	req.Prompt = "none"
	req.SessionCookie = "sess_active123"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.False(t, resp.IsLocalError)

	parsed, parseErr := url.Parse(resp.RedirectURL)
	require.NoError(t, parseErr)
	assert.Equal(t, "consent_required", parsed.Query().Get("error"))
	assert.Contains(t, parsed.Query().Get("error_description"), "consent_required")
	assert.Equal(t, "csrf_state_123", parsed.Query().Get("state"))
}

func TestInitializeAuth_PromptNone_NoSession_ReturnsLoginRequired(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)
	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Create", ctx, mock.AnythingOfType("*repositories.AuthorizationTransaction"), auth.TransactionTTL).Return(nil)
	uc := newTestOAuthInteraction(
		clientRepo,
		new(mocks.MockRoleRepository),
		txnRepo,
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)
	req := validInitializeRequest()
	req.Prompt = "none"
	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.Contains(t, resp.RedirectURL, "error=login_required")
	assert.Contains(t, resp.RedirectURL, "https://app.example.com/callback")
}

// --- prompt=login Tests ---

func TestInitializeAuth_PromptLogin_EligibleSession_ForceLogin(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Create", ctx, mock.AnythingOfType("*repositories.AuthorizationTransaction"), auth.TransactionTTL).Return(nil)

	sessionRepo := new(mocks.MockSessionRepository)
	sessionRepo.On("Get", ctx, "sess_active123").Return(validSession(), nil)

	endUserRepo := new(mocks.MockEndUserRepository)
	endUserRepo.On("GetByID", ctx, "user_123").Return(activeUser(), nil)

	roleRepo := new(mocks.MockRoleRepository)
	roleRepo.On("HasAnyRole", ctx, "user_123", "client_abc123").Return(true, nil)

	uc := newTestOAuthInteraction(clientRepo, roleRepo, txnRepo, sessionRepo, endUserRepo, new(mocks.MockAuditRepository))

	req := validInitializeRequest()
	req.Prompt = "login"
	req.SessionCookie = "sess_active123"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.False(t, resp.IsLocalError)
	assert.Contains(t, resp.RedirectURL, "/oauth2/login")
}

// --- max_age Tests ---

func TestInitializeAuth_MaxAgeExceeded_RedirectsToLogin(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Create", ctx, mock.AnythingOfType("*repositories.AuthorizationTransaction"), auth.TransactionTTL).Return(nil)

	oldSession := validSession()
	oldSession.AuthenticatedAt = time.Now().Add(-5 * time.Minute)

	sessionRepo := new(mocks.MockSessionRepository)
	sessionRepo.On("Get", ctx, "sess_old").Return(oldSession, nil)

	endUserRepo := new(mocks.MockEndUserRepository)
	endUserRepo.On("GetByID", ctx, "user_123").Return(activeUser(), nil)

	roleRepo := new(mocks.MockRoleRepository)
	roleRepo.On("HasAnyRole", ctx, "user_123", "client_abc123").Return(true, nil)

	uc := newTestOAuthInteraction(clientRepo, roleRepo, txnRepo, sessionRepo, endUserRepo, new(mocks.MockAuditRepository))

	req := validInitializeRequest()
	maxAge := 60
	req.MaxAge = &maxAge
	req.SessionCookie = "sess_old"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.Contains(t, resp.RedirectURL, "/oauth2/login")
}

func TestInitializeAuth_InvalidMaxAge_ReturnsCallbackError(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)
	uc := newTestOAuthInteraction(
		clientRepo,
		new(mocks.MockRoleRepository),
		new(mocks.MockAuthorizationTransactionRepository),
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)
	req := validInitializeRequest()
	maxAge := -1
	req.MaxAge = &maxAge

	resp, err := uc.InitializeAuth(ctx, req)

	require.NoError(t, err)
	parsed, err := url.Parse(resp.RedirectURL)
	require.NoError(t, err)
	assert.Equal(t, "invalid_request", parsed.Query().Get("error"))
	assert.Equal(t, "csrf_state_123", parsed.Query().Get("state"))
}

func TestInitializeAuth_MaxAgeWithin_RedirectsToConsent(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Create", ctx, mock.AnythingOfType("*repositories.AuthorizationTransaction"), auth.TransactionTTL).Return(nil)
	txnRepo.On("UpdateStage", ctx, mock.AnythingOfType("string"), repositories.TransactionStagePendingConsent, "user_123", "sess_active123").Return(nil)

	recentSession := validSession()
	recentSession.AuthenticatedAt = time.Now().Add(-10 * time.Second)

	sessionRepo := new(mocks.MockSessionRepository)
	sessionRepo.On("Get", ctx, "sess_recent").Return(recentSession, nil)

	endUserRepo := new(mocks.MockEndUserRepository)
	endUserRepo.On("GetByID", ctx, "user_123").Return(activeUser(), nil)

	roleRepo := new(mocks.MockRoleRepository)
	roleRepo.On("HasAnyRole", ctx, "user_123", "client_abc123").Return(true, nil)

	uc := newTestOAuthInteraction(clientRepo, roleRepo, txnRepo, sessionRepo, endUserRepo, new(mocks.MockAuditRepository))

	req := validInitializeRequest()
	maxAge := 3600
	req.MaxAge = &maxAge
	req.SessionCookie = "sess_recent"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.Contains(t, resp.RedirectURL, "/oauth2/consent")
}

func TestInitializeAuth_PromptConsent_AlwaysRedirectsToConsent(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Create", ctx, mock.AnythingOfType("*repositories.AuthorizationTransaction"), auth.TransactionTTL).Return(nil)
	txnRepo.On("UpdateStage", ctx, mock.AnythingOfType("string"), repositories.TransactionStagePendingConsent, "user_123", "sess_active123").Return(nil)

	sessionRepo := new(mocks.MockSessionRepository)
	sessionRepo.On("Get", ctx, "sess_active123").Return(validSession(), nil)

	endUserRepo := new(mocks.MockEndUserRepository)
	endUserRepo.On("GetByID", ctx, "user_123").Return(activeUser(), nil)

	roleRepo := new(mocks.MockRoleRepository)
	roleRepo.On("HasAnyRole", ctx, "user_123", "client_abc123").Return(true, nil)

	uc := newTestOAuthInteraction(clientRepo, roleRepo, txnRepo, sessionRepo, endUserRepo, new(mocks.MockAuditRepository))

	req := validInitializeRequest()
	req.Prompt = "consent"
	req.SessionCookie = "sess_active123"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.Contains(t, resp.RedirectURL, "/oauth2/consent")
}

// --- Session Not Found ---

func TestInitializeAuth_SessionNotFound_RedirectsToLogin(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Create", ctx, mock.AnythingOfType("*repositories.AuthorizationTransaction"), auth.TransactionTTL).Return(nil)

	sessionRepo := new(mocks.MockSessionRepository)
	sessionRepo.On("Get", ctx, "sess_missing").Return(nil, nil)

	uc := newTestOAuthInteraction(
		clientRepo,
		new(mocks.MockRoleRepository),
		txnRepo,
		sessionRepo,
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()
	req.SessionCookie = "sess_missing"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.Contains(t, resp.RedirectURL, "/oauth2/login")
}

func TestInitializeAuth_EmptySessionCookie_NoSessionLookup(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Create", ctx, mock.AnythingOfType("*repositories.AuthorizationTransaction"), auth.TransactionTTL).Return(nil)

	sessionRepo := new(mocks.MockSessionRepository)
	// No session lookups should be made

	uc := newTestOAuthInteraction(
		clientRepo,
		new(mocks.MockRoleRepository),
		txnRepo,
		sessionRepo,
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()
	req.SessionCookie = ""

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.Contains(t, resp.RedirectURL, "/oauth2/login")

	// Assert no session repo calls were made
	sessionRepo.AssertNotCalled(t, "Get")
}

// --- Callback Error Redirect Preserves State ---

func TestInitializeAuth_CallbackErrorRedirect_PreservesState(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	uc := newTestOAuthInteraction(
		clientRepo,
		new(mocks.MockRoleRepository),
		new(mocks.MockAuthorizationTransactionRepository),
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()
	req.Prompt = "none login" // triggers invalid_request callback error

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.False(t, resp.IsLocalError)

	parsed, parseErr := url.Parse(resp.RedirectURL)
	require.NoError(t, parseErr)
	assert.Equal(t, "csrf_state_123", parsed.Query().Get("state"))
	assert.Equal(t, "invalid_request", parsed.Query().Get("error"))
}

// --- Transaction Stage Tests ---

func TestInitializeAuth_TransactionFieldsPopulated(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	var capturedTxn *repositories.AuthorizationTransaction
	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Create", ctx, mock.MatchedBy(func(txn *repositories.AuthorizationTransaction) bool {
		capturedTxn = txn
		return true
	}), auth.TransactionTTL).Return(nil)

	uc := newTestOAuthInteraction(
		clientRepo,
		new(mocks.MockRoleRepository),
		txnRepo,
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()

	_, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, capturedTxn)
	assert.Equal(t, repositories.TransactionStagePendingLogin, capturedTxn.Stage)
	assert.Equal(t, "client_abc123", capturedTxn.ClientID)
	assert.Equal(t, "tenant_xyz", capturedTxn.TenantID)
	assert.Equal(t, "https://app.example.com/callback", capturedTxn.RedirectURI)
	assert.Equal(t, "code", capturedTxn.ResponseType)
	assert.Equal(t, "openid profile email", capturedTxn.Scope)
	assert.Equal(t, "csrf_state_123", capturedTxn.State)
	assert.Equal(t, "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM", capturedTxn.CodeChallenge)
	assert.Equal(t, "S256", capturedTxn.CodeChallengeMethod)
	assert.NotEmpty(t, capturedTxn.TransactionID)
	assert.NotEmpty(t, capturedTxn.InteractionCSRFToken)
}

// --- Helper function tests ---

func TestParsePrompt(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
		want   []string
	}{
		{"empty", "", nil},
		{"single", "none", []string{"none"}},
		{"multiple", "none login", []string{"none", "login"}},
		{"extra spaces", "  consent   login  ", []string{"consent", "login"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := auth.ParsePrompt(tt.prompt)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainsString(t *testing.T) {
	assert.True(t, auth.ContainsString([]string{"none", "login"}, "none"))
	assert.False(t, auth.ContainsString([]string{"login"}, "none"))
	assert.False(t, auth.ContainsString(nil, "none"))
}

func TestGenerateSecureToken(t *testing.T) {
	token1, err := auth.GenerateSecureToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token1)

	token2, err := auth.GenerateSecureToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token2)

	// Tokens should be unique
	assert.NotEqual(t, token1, token2)

	// Should be base64url encoded (no +, /, or =)
	assert.NotContains(t, token1, "+")
	assert.NotContains(t, token1, "/")
	assert.NotContains(t, token1, "=")
}

// --- Audit logging for invalid callback ---

func TestInitializeAuth_InvalidCallback_AuditLogged(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	auditRepo := new(mocks.MockAuditRepository)
	auditRepo.On("Create", ctx, mock.MatchedBy(func(log *entities.AuditLog) bool {
		return log.EventType == entities.EventOAuthInvalidCallback
	})).Return(nil)

	uc := newTestOAuthInteraction(
		clientRepo,
		new(mocks.MockRoleRepository),
		new(mocks.MockAuthorizationTransactionRepository),
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		auditRepo,
	)

	req := validInitializeRequest()
	req.RedirectURI = "https://evil.example.com/callback"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.IsLocalError)

	auditRepo.AssertCalled(t, "Create", ctx, mock.MatchedBy(func(log *entities.AuditLog) bool {
		return log.EventType == entities.EventOAuthInvalidCallback
	}))
}

// --- UpdateStage failure fail-closed behavior ---

func TestInitializeAuth_UpdateStageFails_RedirectsToLocalError(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Create", ctx, mock.AnythingOfType("*repositories.AuthorizationTransaction"), auth.TransactionTTL).Return(nil)
	txnRepo.On("UpdateStage", ctx, mock.AnythingOfType("string"), repositories.TransactionStagePendingConsent, "user_123", "sess_active123").Return(errors.New("redis error"))

	sessionRepo := new(mocks.MockSessionRepository)
	sessionRepo.On("Get", ctx, "sess_active123").Return(validSession(), nil)

	endUserRepo := new(mocks.MockEndUserRepository)
	endUserRepo.On("GetByID", ctx, "user_123").Return(activeUser(), nil)

	roleRepo := new(mocks.MockRoleRepository)
	roleRepo.On("HasAnyRole", ctx, "user_123", "client_abc123").Return(true, nil)

	uc := newTestOAuthInteraction(clientRepo, roleRepo, txnRepo, sessionRepo, endUserRepo, new(mocks.MockAuditRepository))

	req := validInitializeRequest()
	req.SessionCookie = "sess_active123"

	resp, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.IsLocalError)
	assert.Contains(t, resp.RedirectURL, "/oauth2/error")
	assert.Contains(t, resp.RedirectURL, "temporarily_unavailable")
}

// --- Nonce is stored in the transaction ---

func TestInitializeAuth_NonceStoredInTransaction(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	var capturedTxn *repositories.AuthorizationTransaction
	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Create", ctx, mock.MatchedBy(func(txn *repositories.AuthorizationTransaction) bool {
		capturedTxn = txn
		return true
	}), auth.TransactionTTL).Return(nil)

	uc := newTestOAuthInteraction(
		clientRepo,
		new(mocks.MockRoleRepository),
		txnRepo,
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()
	req.Nonce = "nonce_abc123"

	_, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, capturedTxn)
	assert.Equal(t, "nonce_abc123", capturedTxn.Nonce)
}

// --- Prompt stored in transaction ---

func TestInitializeAuth_PromptStoredInTransaction(t *testing.T) {
	ctx := context.Background()
	clientRepo := new(mocks.MockClientRepository)
	clientRepo.On("GetByID", ctx, "client_abc123").Return(validClient(), nil)

	var capturedTxn *repositories.AuthorizationTransaction
	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	txnRepo.On("Create", ctx, mock.MatchedBy(func(txn *repositories.AuthorizationTransaction) bool {
		capturedTxn = txn
		return true
	}), auth.TransactionTTL).Return(nil)

	uc := newTestOAuthInteraction(
		clientRepo,
		new(mocks.MockRoleRepository),
		txnRepo,
		new(mocks.MockSessionRepository),
		new(mocks.MockEndUserRepository),
		new(mocks.MockAuditRepository),
	)

	req := validInitializeRequest()
	req.Prompt = "consent"

	_, err := uc.InitializeAuth(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, capturedTxn)
	assert.Equal(t, []string{"consent"}, capturedTxn.Prompt)
}
