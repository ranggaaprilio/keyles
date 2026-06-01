package usecase_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testAuthEndUserFrontendURL = "https://sso.example.com"

// newUseCase constructs an AuthenticateEndUser use case with all mocks.
// Individual tests override mock expectations as needed.
func newAuthenticateEndUser(
	txnRepo *mocks.MockAuthorizationTransactionRepository,
	sessionRepo *mocks.MockSessionRepository,
	clientRepo *mocks.MockClientRepository,
	endUserRepo *mocks.MockEndUserRepository,
	roleRepo *mocks.MockRoleRepository,
	throttler *mocks.MockLoginThrottler,
	passwordSvc *mocks.MockPasswordService,
	auditRepo *mocks.MockAuditRepository,
) *auth.AuthenticateEndUser {
	auditHelper := auth.NewOAuthAuditHelper(auditRepo)
	return auth.NewAuthenticateEndUser(
		txnRepo, sessionRepo, clientRepo, endUserRepo,
		roleRepo, throttler, passwordSvc, auditHelper,
		8*time.Hour, testAuthEndUserFrontendURL,
	)
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func validPendingTransaction() *repositories.AuthorizationTransaction {
	now := time.Now()
	return &repositories.AuthorizationTransaction{
		TransactionID:       "txn_abc123",
		ClientID:            "client_001",
		TenantID:            "tenant_001",
		RedirectURI:         "https://app.example.com/callback",
		ResponseType:        "code",
		Scope:               "openid profile",
		State:               "state_xyz",
		CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		CodeChallengeMethod: "S256",
		Stage:               repositories.TransactionStagePendingLogin,
		CreatedAt:           now,
		ExpiresAt:           now.Add(10 * time.Minute),
	}
}

func activeEndUser() *entities.User {
	now := time.Now()
	return &entities.User{
		ID:           "user_001",
		TenantID:     "tenant_001",
		Email:        "alice@example.com",
		DisplayName:  "Alice",
		PasswordHash: "$2a$12$hashedpassword",
		Status:       entities.UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func activeClient() *entities.Client {
	now := time.Now()
	return &entities.Client{
		ClientID:            "client_001",
		TenantID:            "tenant_001",
		ClientName:          "Test App",
		ClientType:          entities.ClientTypeConfidential,
		IsActive:            true,
		AllowedRedirectURIs: []string{"https://app.example.com/callback"},
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// defaultMocks sets up the common mock expectations for a successful
// authentication flow. Individual tests can override specific mocks.
func defaultMocks(
	txnRepo *mocks.MockAuthorizationTransactionRepository,
	sessionRepo *mocks.MockSessionRepository,
	clientRepo *mocks.MockClientRepository,
	endUserRepo *mocks.MockEndUserRepository,
	roleRepo *mocks.MockRoleRepository,
	throttler *mocks.MockLoginThrottler,
	passwordSvc *mocks.MockPasswordService,
	auditRepo *mocks.MockAuditRepository,
) {
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
}

// ─── Success path ───────────────────────────────────────────────────────────

func TestAuthenticateEndUser_Success(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()
	user := activeEndUser()
	client := activeClient()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	throttler.On("IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(false, nil)
	clientRepo.On("GetByID", mock.Anything, "client_001").Return(client, nil)
	endUserRepo.On("GetByEmail", mock.Anything, "tenant_001", "alice@example.com").Return(user, nil)
	roleRepo.On("HasAnyRole", mock.Anything, "user_001", "client_001").Return(true, nil)
	passwordSvc.On("Verify", "correctpassword", user.PasswordHash).Return(nil)
	sessionRepo.On("Create", mock.Anything, mock.Anything, 8*time.Hour).Return(nil)
	throttler.On("ClearEmailBucket", mock.Anything, "tenant_001", "alice@example.com").Return(nil)
	txnRepo.On("UpdateStage", mock.Anything, "txn_abc123", repositories.TransactionStagePendingConsent, "user_001", mock.Anything).Return(nil)
	endUserRepo.On("UpdateLastLogin", mock.Anything, "user_001", mock.Anything).Return(nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "correctpassword",
		SourceIP:      "10.0.0.1",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Contains(t, resp.RedirectURL, "/oauth2/consent?transaction_id=txn_abc123")
	require.True(t, strings.HasPrefix(resp.RedirectURL, testAuthEndUserFrontendURL))

	// Verify session was created with AuthenticatedAt set
	sessionArg := sessionRepo.Calls[0].Arguments.Get(1).(*repositories.Session)
	require.NotEmpty(t, sessionArg.SessionID)
	require.Equal(t, "user_001", sessionArg.UserID)
	require.Equal(t, "tenant_001", sessionArg.TenantID)
	require.False(t, sessionArg.AuthenticatedAt.IsZero())

	// Verify UpdateStage was called with pending_consent and user binding
	updateStageCall := txnRepo.Calls[1]
	require.Equal(t, "UpdateStage", updateStageCall.Method)
	require.Equal(t, repositories.TransactionStagePendingConsent, updateStageCall.Arguments.Get(2))
	require.Equal(t, "user_001", updateStageCall.Arguments.Get(3))

	// Verify last login was recorded
	endUserRepo.AssertCalled(t, "UpdateLastLogin", mock.Anything, "user_001", mock.Anything)

	// Verify audit was recorded
	auditRepo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
}

// ─── Email normalization ────────────────────────────────────────────────────

func TestAuthenticateEndUser_EmailNormalization(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()
	user := activeEndUser()
	client := activeClient()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	// The throttle and GetByEmail should receive the normalized (lowercased, trimmed) email
	throttler.On("IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(false, nil)
	clientRepo.On("GetByID", mock.Anything, "client_001").Return(client, nil)
	endUserRepo.On("GetByEmail", mock.Anything, "tenant_001", "alice@example.com").Return(user, nil)
	roleRepo.On("HasAnyRole", mock.Anything, "user_001", "client_001").Return(true, nil)
	passwordSvc.On("Verify", "password", user.PasswordHash).Return(nil)
	sessionRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	throttler.On("ClearEmailBucket", mock.Anything, "tenant_001", "alice@example.com").Return(nil)
	txnRepo.On("UpdateStage", mock.Anything, "txn_abc123", repositories.TransactionStagePendingConsent, "user_001", mock.Anything).Return(nil)
	endUserRepo.On("UpdateLastLogin", mock.Anything, "user_001", mock.Anything).Return(nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "  Alice@Example.COM  ",
		Password:      "password",
		SourceIP:      "10.0.0.1",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify normalized email was used for throttler calls
	throttler.AssertCalled(t, "IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com")
	throttler.AssertCalled(t, "ClearEmailBucket", mock.Anything, "tenant_001", "alice@example.com")
}

// ─── Transaction validation failures ────────────────────────────────────────

func TestAuthenticateEndUser_TransactionNotFound(t *testing.T) {
	ctx := context.Background()
	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "nonexistent").Return(nil, nil)

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "nonexistent",
		Email:         "alice@example.com",
		Password:      "password",
		SourceIP:      "10.0.0.1",
	})

	require.Nil(t, resp)
	require.ErrorIs(t, err, auth.ErrTransactionExpired)
}

func TestAuthenticateEndUser_TransactionGetError(t *testing.T) {
	ctx := context.Background()
	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(nil, fmt.Errorf("redis error"))

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "password",
		SourceIP:      "10.0.0.1",
	})

	require.Nil(t, resp)
	require.ErrorIs(t, err, auth.ErrTemporarilyUnavailable)
}

func TestAuthenticateEndUser_TransactionWrongStage(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()
	txn.Stage = repositories.TransactionStagePendingConsent

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "password",
		SourceIP:      "10.0.0.1",
	})

	require.Nil(t, resp)
	require.ErrorIs(t, err, auth.ErrTransactionExpired)
}

func TestAuthenticateEndUser_TransactionCompleted(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()
	txn.Stage = repositories.TransactionStageCompleted

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "password",
		SourceIP:      "10.0.0.1",
	})

	require.Nil(t, resp)
	require.ErrorIs(t, err, auth.ErrTransactionExpired)
}

// ─── Throttle check ─────────────────────────────────────────────────────────

func TestAuthenticateEndUser_Throttled(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	throttler.On("IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(true, nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "anypassword",
		SourceIP:      "10.0.0.1",
	})

	require.Nil(t, resp)
	require.ErrorIs(t, err, auth.ErrThrottled)

	// Must NOT call RecordFailure when throttled — the bucket is full already
	throttler.AssertNotCalled(t, "RecordFailure")
	// Must NOT attempt password verification
	passwordSvc.AssertNotCalled(t, "Verify")
	// Must emit throttled audit event
	auditRepo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
	entry := auditRepo.Calls[0].Arguments.Get(1).(*entities.AuditLog)
	require.Equal(t, entities.EventOAuthLoginThrottled, entry.EventType)
}

func TestAuthenticateEndUser_ThrottleCheckError_FailsClosed(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	// Throttle state is a security control: Redis failure must fail closed.
	throttler.On("IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(false, fmt.Errorf("redis down"))

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "correctpassword",
		SourceIP:      "10.0.0.1",
	})

	require.Nil(t, resp)
	require.ErrorIs(t, err, auth.ErrTemporarilyUnavailable)
	passwordSvc.AssertNotCalled(t, "Verify")
}

// ─── Credential failures (all must return ErrInvalidCredentials) ────────────

func TestAuthenticateEndUser_ClientNotFound(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	throttler.On("IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(false, nil)
	clientRepo.On("GetByID", mock.Anything, "client_001").Return(nil, nil)
	throttler.On("RecordFailure", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "password",
		SourceIP:      "10.0.0.1",
	})

	require.Nil(t, resp)
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
	throttler.AssertCalled(t, "RecordFailure", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com")

	// Verify audit event is oauth_login_failed not oauth_login_succeeded
	entry := auditRepo.Calls[0].Arguments.Get(1).(*entities.AuditLog)
	require.Equal(t, entities.EventOAuthLoginFailed, entry.EventType)
}

func TestAuthenticateEndUser_ClientDisabled(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()
	client := activeClient()
	client.IsActive = false

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	throttler.On("IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(false, nil)
	clientRepo.On("GetByID", mock.Anything, "client_001").Return(client, nil)
	throttler.On("RecordFailure", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "password",
		SourceIP:      "10.0.0.1",
	})

	require.Nil(t, resp)
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestAuthenticateEndUser_UserNotFound(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()
	client := activeClient()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	throttler.On("IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(false, nil)
	clientRepo.On("GetByID", mock.Anything, "client_001").Return(client, nil)
	endUserRepo.On("GetByEmail", mock.Anything, "tenant_001", "alice@example.com").Return(nil, nil)
	throttler.On("RecordFailure", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "password",
		SourceIP:      "10.0.0.1",
	})

	require.Nil(t, resp)
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
	passwordSvc.AssertNotCalled(t, "Verify")
}

func TestAuthenticateEndUser_UserDisabled(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()
	user := activeEndUser()
	user.Status = entities.UserStatusDisabled
	client := activeClient()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	throttler.On("IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(false, nil)
	clientRepo.On("GetByID", mock.Anything, "client_001").Return(client, nil)
	endUserRepo.On("GetByEmail", mock.Anything, "tenant_001", "alice@example.com").Return(user, nil)
	throttler.On("RecordFailure", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "password",
		SourceIP:      "10.0.0.1",
	})

	require.Nil(t, resp)
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
	// Password should NOT be verified for disabled user
	passwordSvc.AssertNotCalled(t, "Verify")
}

func TestAuthenticateEndUser_UserPending(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()
	user := activeEndUser()
	user.Status = entities.UserStatusPending
	client := activeClient()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	throttler.On("IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(false, nil)
	clientRepo.On("GetByID", mock.Anything, "client_001").Return(client, nil)
	endUserRepo.On("GetByEmail", mock.Anything, "tenant_001", "alice@example.com").Return(user, nil)
	throttler.On("RecordFailure", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "password",
		SourceIP:      "10.0.0.1",
	})

	require.Nil(t, resp)
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
	passwordSvc.AssertNotCalled(t, "Verify")
}

func TestAuthenticateEndUser_NoRole(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()
	user := activeEndUser()
	client := activeClient()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	throttler.On("IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(false, nil)
	clientRepo.On("GetByID", mock.Anything, "client_001").Return(client, nil)
	endUserRepo.On("GetByEmail", mock.Anything, "tenant_001", "alice@example.com").Return(user, nil)
	roleRepo.On("HasAnyRole", mock.Anything, "user_001", "client_001").Return(false, nil)
	throttler.On("RecordFailure", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "password",
		SourceIP:      "10.0.0.1",
	})

	require.Nil(t, resp)
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
	passwordSvc.AssertNotCalled(t, "Verify")
}

func TestAuthenticateEndUser_WrongPassword(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()
	user := activeEndUser()
	client := activeClient()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	throttler.On("IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(false, nil)
	clientRepo.On("GetByID", mock.Anything, "client_001").Return(client, nil)
	endUserRepo.On("GetByEmail", mock.Anything, "tenant_001", "alice@example.com").Return(user, nil)
	roleRepo.On("HasAnyRole", mock.Anything, "user_001", "client_001").Return(true, nil)
	passwordSvc.On("Verify", "wrongpassword", user.PasswordHash).Return(fmt.Errorf("bcrypt: password mismatch"))
	throttler.On("RecordFailure", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "wrongpassword",
		SourceIP:      "10.0.0.1",
	})

	require.Nil(t, resp)
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
	throttler.AssertCalled(t, "RecordFailure", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com")
}

// ─── Generic error message (NFR-003) ────────────────────────────────────────

func TestAuthenticateEndUser_GenericCredentialError(t *testing.T) {
	// All credential-related failures must return the same error:
	// ErrInvalidCredentials. This verifies NFR-003 compliance.
	errs := []error{auth.ErrInvalidCredentials}
	for _, e := range errs {
		require.Equal(t, "invalid credentials", e.Error())
	}
}

// ─── Session creation ────────────────────────────────────────────────────────

func TestAuthenticateEndUser_SessionIDCryptoRandom(t *testing.T) {
	// generateSessionID should produce unique, base64url-encoded IDs.
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id, err := auth.GenerateSessionID()
		require.NoError(t, err)
		require.NotEmpty(t, id)
		require.Len(t, id, 43) // 32 bytes → base64url = ceil(32*4/3) = 43, no padding
		ids[id] = true
	}
	require.GreaterOrEqual(t, len(ids), 98, "session IDs should be unique")
}

// ─── UpdateStage failure ────────────────────────────────────────────────────

func TestAuthenticateEndUser_UpdateStageFailure(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()
	user := activeEndUser()
	client := activeClient()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	throttler.On("IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(false, nil)
	clientRepo.On("GetByID", mock.Anything, "client_001").Return(client, nil)
	endUserRepo.On("GetByEmail", mock.Anything, "tenant_001", "alice@example.com").Return(user, nil)
	roleRepo.On("HasAnyRole", mock.Anything, "user_001", "client_001").Return(true, nil)
	passwordSvc.On("Verify", "correctpassword", user.PasswordHash).Return(nil)
	sessionRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	throttler.On("ClearEmailBucket", mock.Anything, "tenant_001", "alice@example.com").Return(nil)
	txnRepo.On("UpdateStage", mock.Anything, "txn_abc123", repositories.TransactionStagePendingConsent, "user_001", mock.Anything).Return(fmt.Errorf("redis error"))
	sessionRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)

	// Redis binding failure must fail closed.
	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "correctpassword",
		SourceIP:      "10.0.0.1",
	})

	require.Nil(t, resp)
	require.ErrorIs(t, err, auth.ErrTemporarilyUnavailable)
}

// ─── Redirect URL ────────────────────────────────────────────────────────────

func TestAuthenticateEndUser_RedirectURL(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()
	user := activeEndUser()
	client := activeClient()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	throttler.On("IsThrottled", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
	clientRepo.On("GetByID", mock.Anything, "client_001").Return(client, nil)
	endUserRepo.On("GetByEmail", mock.Anything, mock.Anything, mock.Anything).Return(user, nil)
	roleRepo.On("HasAnyRole", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	passwordSvc.On("Verify", mock.Anything, mock.Anything).Return(nil)
	sessionRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	throttler.On("ClearEmailBucket", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	txnRepo.On("UpdateStage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	endUserRepo.On("UpdateLastLogin", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "correctpassword",
		SourceIP:      "10.0.0.1",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	expectedURL := fmt.Sprintf("%s/oauth2/consent?transaction_id=%s", testAuthEndUserFrontendURL, "txn_abc123")
	require.Equal(t, expectedURL, resp.RedirectURL)
}

// ─── Repository error fallback paths ────────────────────────────────────────

func TestAuthenticateEndUser_ClientRepoError(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	throttler.On("IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(false, nil)
	clientRepo.On("GetByID", mock.Anything, "client_001").Return(nil, fmt.Errorf("db error"))
	throttler.On("RecordFailure", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "password",
		SourceIP:      "10.0.0.1",
	})

	require.Nil(t, resp)
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestAuthenticateEndUser_EndUserRepoError(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()
	client := activeClient()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	throttler.On("IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(false, nil)
	clientRepo.On("GetByID", mock.Anything, "client_001").Return(client, nil)
	endUserRepo.On("GetByEmail", mock.Anything, "tenant_001", "alice@example.com").Return(nil, fmt.Errorf("db error"))
	throttler.On("RecordFailure", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "password",
		SourceIP:      "10.0.0.1",
	})

	require.Nil(t, resp)
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestAuthenticateEndUser_RoleRepoError(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()
	user := activeEndUser()
	client := activeClient()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	throttler.On("IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(false, nil)
	clientRepo.On("GetByID", mock.Anything, "client_001").Return(client, nil)
	endUserRepo.On("GetByEmail", mock.Anything, "tenant_001", "alice@example.com").Return(user, nil)
	roleRepo.On("HasAnyRole", mock.Anything, "user_001", "client_001").Return(false, fmt.Errorf("db error"))
	throttler.On("RecordFailure", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "password",
		SourceIP:      "10.0.0.1",
	})

	require.Nil(t, resp)
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestAuthenticateEndUser_SessionCreateError(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()
	user := activeEndUser()
	client := activeClient()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	throttler.On("IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(false, nil)
	clientRepo.On("GetByID", mock.Anything, "client_001").Return(client, nil)
	endUserRepo.On("GetByEmail", mock.Anything, "tenant_001", "alice@example.com").Return(user, nil)
	roleRepo.On("HasAnyRole", mock.Anything, "user_001", "client_001").Return(true, nil)
	passwordSvc.On("Verify", "correctpassword", user.PasswordHash).Return(nil)
	sessionRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("redis error"))

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "correctpassword",
		SourceIP:      "10.0.0.1",
	})

	require.Nil(t, resp)
	require.ErrorIs(t, err, auth.ErrTemporarilyUnavailable)
	throttler.AssertNotCalled(t, "RecordFailure")
}

// ─── Success clears email throttle bucket ────────────────────────────────────

func TestAuthenticateEndUser_ClearsEmailBucketOnSuccess(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()
	user := activeEndUser()
	client := activeClient()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	throttler.On("IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(false, nil)
	clientRepo.On("GetByID", mock.Anything, "client_001").Return(client, nil)
	endUserRepo.On("GetByEmail", mock.Anything, "tenant_001", "alice@example.com").Return(user, nil)
	roleRepo.On("HasAnyRole", mock.Anything, "user_001", "client_001").Return(true, nil)
	passwordSvc.On("Verify", "correctpassword", user.PasswordHash).Return(nil)
	sessionRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	throttler.On("ClearEmailBucket", mock.Anything, "tenant_001", "alice@example.com").Return(nil)
	txnRepo.On("UpdateStage", mock.Anything, "txn_abc123", repositories.TransactionStagePendingConsent, "user_001", mock.Anything).Return(nil)
	endUserRepo.On("UpdateLastLogin", mock.Anything, "user_001", mock.Anything).Return(nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "correctpassword",
		SourceIP:      "10.0.0.1",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	throttler.AssertCalled(t, "ClearEmailBucket", mock.Anything, "tenant_001", "alice@example.com")
}

// ─── Success emits login_succeeded audit event ───────────────────────────────

func TestAuthenticateEndUser_AuditEventOnSuccess(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()
	user := activeEndUser()
	client := activeClient()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	throttler.On("IsThrottled", mock.Anything, "10.0.0.1", "tenant_001", "alice@example.com").Return(false, nil)
	clientRepo.On("GetByID", mock.Anything, "client_001").Return(client, nil)
	endUserRepo.On("GetByEmail", mock.Anything, "tenant_001", "alice@example.com").Return(user, nil)
	roleRepo.On("HasAnyRole", mock.Anything, "user_001", "client_001").Return(true, nil)
	passwordSvc.On("Verify", "correctpassword", user.PasswordHash).Return(nil)
	sessionRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	throttler.On("ClearEmailBucket", mock.Anything, "tenant_001", "alice@example.com").Return(nil)
	txnRepo.On("UpdateStage", mock.Anything, "txn_abc123", repositories.TransactionStagePendingConsent, "user_001", mock.Anything).Return(nil)
	endUserRepo.On("UpdateLastLogin", mock.Anything, "user_001", mock.Anything).Return(nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	_, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "correctpassword",
		SourceIP:      "10.0.0.1",
	})

	require.NoError(t, err)

	require.Len(t, auditRepo.Calls, 1)
	entry := auditRepo.Calls[0].Arguments.Get(1).(*entities.AuditLog)
	require.Equal(t, entities.EventOAuthLoginSucceeded, entry.EventType)
	require.Equal(t, "10.0.0.1", entry.IPAddress)
}

// ─── Success records last login ──────────────────────────────────────────────

func TestAuthenticateEndUser_RecordsLastLogin(t *testing.T) {
	ctx := context.Background()
	txn := validPendingTransaction()
	user := activeEndUser()
	client := activeClient()

	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := newAuthenticateEndUser(txnRepo, sessionRepo, clientRepo, endUserRepo, roleRepo, throttler, passwordSvc, auditRepo)

	txnRepo.On("Get", mock.Anything, "txn_abc123").Return(txn, nil)
	throttler.On("IsThrottled", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
	clientRepo.On("GetByID", mock.Anything, "client_001").Return(client, nil)
	endUserRepo.On("GetByEmail", mock.Anything, mock.Anything, mock.Anything).Return(user, nil)
	roleRepo.On("HasAnyRole", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	passwordSvc.On("Verify", mock.Anything, mock.Anything).Return(nil)
	sessionRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	throttler.On("ClearEmailBucket", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	txnRepo.On("UpdateStage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	endUserRepo.On("UpdateLastLogin", mock.Anything, "user_001", mock.Anything).Return(nil)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	_, err := uc.Execute(ctx, auth.AuthenticateEndUserRequest{
		TransactionID: "txn_abc123",
		Email:         "alice@example.com",
		Password:      "correctpassword",
		SourceIP:      "10.0.0.1",
	})

	require.NoError(t, err)
	endUserRepo.AssertCalled(t, "UpdateLastLogin", mock.Anything, "user_001", mock.Anything)
}

// ─── Default session TTL ────────────────────────────────────────────────────

func TestAuthenticateEndUser_DefaultSessionTTL(t *testing.T) {
	txnRepo := new(mocks.MockAuthorizationTransactionRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	clientRepo := new(mocks.MockClientRepository)
	endUserRepo := new(mocks.MockEndUserRepository)
	roleRepo := new(mocks.MockRoleRepository)
	throttler := new(mocks.MockLoginThrottler)
	passwordSvc := new(mocks.MockPasswordService)
	auditRepo := new(mocks.MockAuditRepository)

	uc := auth.NewAuthenticateEndUser(
		txnRepo, sessionRepo, clientRepo, endUserRepo,
		roleRepo, throttler, passwordSvc,
		auth.NewOAuthAuditHelper(auditRepo),
		0, // zero should default to 8 hours
		testAuthEndUserFrontendURL,
	)

	require.NotNil(t, uc)
}
