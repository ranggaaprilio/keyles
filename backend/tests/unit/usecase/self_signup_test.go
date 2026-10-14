package usecase_test

import (
	"context"
	"errors"
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

const (
	signupTxnID    = "txn_signup"
	signupTenantID = "tenant_001"
	signupClientID = "client_001"
	signupEmail    = "new.user@example.com"
	signupPassword = "Str0ng!Passw0rd"
	signupCode     = "123456"
	signupRole     = "member"
)

// fakeSignupStore is an in-memory PendingSignupStore with the same
// semantics as the Redis implementation (attempt cap, delete on success).
type fakeSignupStore struct {
	records   map[string]*repositories.PendingSignup
	attempts  map[string]int
	sendCount map[string]int
	blockSend bool
}

func newFakeSignupStore() *fakeSignupStore {
	return &fakeSignupStore{
		records:   map[string]*repositories.PendingSignup{},
		attempts:  map[string]int{},
		sendCount: map[string]int{},
	}
}

func (s *fakeSignupStore) key(tenantID, email string) string { return tenantID + "|" + email }

func (s *fakeSignupStore) AllowSend(_ context.Context, tenantID, email string, _ time.Duration, maxPerHour int) (bool, error) {
	if s.blockSend {
		return false, nil
	}
	k := s.key(tenantID, email)
	s.sendCount[k]++
	return s.sendCount[k] <= maxPerHour, nil
}

func (s *fakeSignupStore) Save(_ context.Context, p *repositories.PendingSignup, _ time.Duration) error {
	cp := *p
	s.records[s.key(p.TenantID, p.Email)] = &cp
	s.attempts[s.key(p.TenantID, p.Email)] = 0
	return nil
}

func (s *fakeSignupStore) Get(_ context.Context, tenantID, email string) (*repositories.PendingSignup, error) {
	return s.records[s.key(tenantID, email)], nil
}

func (s *fakeSignupStore) Verify(_ context.Context, tenantID, email, codeHash string, maxAttempts int) (*repositories.PendingSignup, error) {
	k := s.key(tenantID, email)
	rec := s.records[k]
	if rec == nil {
		return nil, nil
	}
	if rec.CodeHash == codeHash {
		delete(s.records, k)
		return rec, nil
	}
	s.attempts[k]++
	if s.attempts[k] >= maxAttempts {
		delete(s.records, k)
	}
	return nil, nil
}

func (s *fakeSignupStore) Delete(_ context.Context, tenantID, email string) error {
	delete(s.records, s.key(tenantID, email))
	return nil
}

// fixedOTP always generates the same code and validates 6-digit format.
type fixedOTP struct{}

func (fixedOTP) Generate() (string, error) { return signupCode, nil }
func (fixedOTP) Validate(otp string) bool {
	if len(otp) != 6 {
		return false
	}
	for _, r := range otp {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

type signupFixture struct {
	uc          *auth.SelfSignup
	txn         *repositories.AuthorizationTransaction
	client      *entities.Client
	store       *fakeSignupStore
	txnRepo     *mocks.MockAuthorizationTransactionRepository
	clientRepo  *mocks.MockClientRepository
	endUserRepo *mocks.MockEndUserRepository
	userRepo    *mocks.MockUserRepository
	roleRepo    *mocks.MockRoleRepository
	sessionRepo *mocks.MockSessionRepository
	throttler   *mocks.MockLoginThrottler
	passwordSvc *mocks.MockPasswordService
	emailSvc    *mocks.MockEmailService
	countCache  *mocks.MockUserCountCache
}

func newSignupFixture(t *testing.T) *signupFixture {
	t.Helper()
	now := time.Now()
	f := &signupFixture{
		txn: &repositories.AuthorizationTransaction{
			TransactionID: signupTxnID,
			ClientID:      signupClientID,
			TenantID:      signupTenantID,
			Stage:         repositories.TransactionStagePendingLogin,
			CreatedAt:     now,
			ExpiresAt:     now.Add(10 * time.Minute),
		},
		client: &entities.Client{
			ClientID:        signupClientID,
			TenantID:        signupTenantID,
			ClientName:      "Acme Portal",
			IsActive:        true,
			AllowSelfSignup: true,
			SelfSignupRole:  signupRole,
		},
		store:       newFakeSignupStore(),
		txnRepo:     new(mocks.MockAuthorizationTransactionRepository),
		clientRepo:  new(mocks.MockClientRepository),
		endUserRepo: new(mocks.MockEndUserRepository),
		userRepo:    new(mocks.MockUserRepository),
		roleRepo:    new(mocks.MockRoleRepository),
		sessionRepo: new(mocks.MockSessionRepository),
		throttler:   new(mocks.MockLoginThrottler),
		passwordSvc: new(mocks.MockPasswordService),
		emailSvc:    new(mocks.MockEmailService),
		countCache:  new(mocks.MockUserCountCache),
	}
	auditRepo := new(mocks.MockAuditRepository)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	f.txnRepo.On("Get", mock.Anything, signupTxnID).Return(f.txn, nil).Maybe()
	f.txnRepo.On("ExtendTTL", mock.Anything, signupTxnID, mock.Anything, mock.Anything).Return(nil).Maybe()
	f.clientRepo.On("GetByID", mock.Anything, signupClientID).Return(f.client, nil).Maybe()
	f.passwordSvc.On("Hash", signupPassword).Return("bcrypt-hash", nil).Maybe()
	f.countCache.On("Get", mock.Anything, signupTenantID).Return(1, true, nil).Maybe()
	f.countCache.On("Invalidate", mock.Anything, signupTenantID).Return(nil).Maybe()
	f.throttler.On("IsThrottled", mock.Anything, mock.Anything, signupTenantID, signupEmail).Return(false, nil).Maybe()
	f.throttler.On("RecordFailure", mock.Anything, mock.Anything, signupTenantID, signupEmail).Return(nil).Maybe()
	f.throttler.On("ClearEmailBucket", mock.Anything, signupTenantID, signupEmail).Return(nil).Maybe()

	auditHelper := auth.NewOAuthAuditHelper(auditRepo)
	completer := auth.NewLoginCompleter(f.txnRepo, f.sessionRepo, f.endUserRepo, f.throttler, auditHelper, time.Hour, "https://sso.example.com/")
	f.uc = auth.NewSelfSignup(
		f.txnRepo, f.clientRepo, f.endUserRepo, f.userRepo, f.roleRepo, f.store,
		f.passwordSvc, fixedOTP{}, f.emailSvc, f.countCache, auditHelper, completer,
	)
	return f
}

func validSignupRequest() auth.SignupRequest {
	return auth.SignupRequest{
		TransactionID: signupTxnID,
		Email:         "  New.User@Example.com ",
		Password:      signupPassword,
		DisplayName:   "New User",
		SourceIP:      "198.51.100.1",
	}
}

func verifyRequest(code string) auth.VerifySignupRequest {
	return auth.VerifySignupRequest{TransactionID: signupTxnID, Email: signupEmail, Code: code, SourceIP: "198.51.100.1"}
}

// requestCode runs a successful sign-up request for a new email.
func (f *signupFixture) requestCode(t *testing.T) {
	t.Helper()
	f.userRepo.On("EmailExists", mock.Anything, signupEmail).Return(false, nil).Once()
	f.emailSvc.On("SendSignupOTPEmail", mock.Anything, signupEmail, "New User", signupCode, "Acme Portal").Return(nil).Once()
	require.NoError(t, f.uc.Request(context.Background(), validSignupRequest()))
}

func TestSelfSignup_GetContext(t *testing.T) {
	f := newSignupFixture(t)

	got, err := f.uc.GetContext(context.Background(), signupTxnID)
	require.NoError(t, err)
	assert.Equal(t, "Acme Portal", got.ClientName)
	assert.True(t, got.SignupEnabled)

	f.client.AllowSelfSignup = false
	got, err = f.uc.GetContext(context.Background(), signupTxnID)
	require.NoError(t, err)
	assert.False(t, got.SignupEnabled)
}

func TestSelfSignup_Request_NewEmailStoresPendingAndSendsCode(t *testing.T) {
	f := newSignupFixture(t)

	f.requestCode(t)

	rec := f.store.records[f.store.key(signupTenantID, signupEmail)]
	require.NotNil(t, rec)
	assert.Equal(t, "bcrypt-hash", rec.PasswordHash)
	assert.Equal(t, "New User", rec.DisplayName)
	assert.NotEqual(t, signupCode, rec.CodeHash, "plaintext code must not be stored")
	f.emailSvc.AssertExpectations(t)
	f.txnRepo.AssertCalled(t, "ExtendTTL", mock.Anything, signupTxnID, auth.SignupTransactionTTL, auth.SignupTransactionMaxLifetime)
	f.endUserRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestSelfSignup_Request_ExistingEmailLooksIdenticalButSendsNothing(t *testing.T) {
	f := newSignupFixture(t)
	f.userRepo.On("EmailExists", mock.Anything, signupEmail).Return(true, nil)

	err := f.uc.Request(context.Background(), validSignupRequest())

	require.NoError(t, err)
	f.emailSvc.AssertNotCalled(t, "SendSignupOTPEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	assert.Empty(t, f.store.records)
	f.passwordSvc.AssertCalled(t, "Hash", signupPassword)
}

func TestSelfSignup_Request_CooldownSuppressesEmailSilently(t *testing.T) {
	f := newSignupFixture(t)
	f.store.blockSend = true
	f.userRepo.On("EmailExists", mock.Anything, signupEmail).Return(false, nil)

	require.NoError(t, f.uc.Request(context.Background(), validSignupRequest()))
	f.emailSvc.AssertNotCalled(t, "SendSignupOTPEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestSelfSignup_Request_EmailFailureIsHiddenAndDropsPendingRecord(t *testing.T) {
	f := newSignupFixture(t)
	f.userRepo.On("EmailExists", mock.Anything, signupEmail).Return(false, nil)
	f.emailSvc.On("SendSignupOTPEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("smtp down"))

	err := f.uc.Request(context.Background(), validSignupRequest())

	require.NoError(t, err, "a send failure must look like success so it cannot reveal unregistered emails")
	assert.Empty(t, f.store.records)
}

func TestSelfSignup_Request_Rejections(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(f *signupFixture)
		wantErr error
	}{
		{"signup disabled", func(f *signupFixture) { f.client.AllowSelfSignup = false }, auth.ErrSignupDisabled},
		{"client disabled", func(f *signupFixture) { f.client.IsActive = false }, auth.ErrSignupInvalidTransaction},
		{"transaction past login", func(f *signupFixture) {
			f.txn.Stage = repositories.TransactionStagePendingConsent
		}, auth.ErrSignupInvalidTransaction},
		{"quota reached", func(f *signupFixture) {
			f.countCache.ExpectedCalls = nil
			f.countCache.On("Get", mock.Anything, signupTenantID).Return(entities.MaxUsersPerTenant, true, nil)
		}, auth.ErrSignupUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newSignupFixture(t)
			tt.mutate(f)

			err := f.uc.Request(context.Background(), validSignupRequest())

			assert.ErrorIs(t, err, tt.wantErr)
			f.emailSvc.AssertNotCalled(t, "SendSignupOTPEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestSelfSignup_Request_ValidationErrors(t *testing.T) {
	for name, mutate := range map[string]func(*auth.SignupRequest){
		"weak password": func(r *auth.SignupRequest) { r.Password = "password" },
		"bad email":     func(r *auth.SignupRequest) { r.Email = "not-an-email" },
		"short name":    func(r *auth.SignupRequest) { r.DisplayName = "A" },
	} {
		t.Run(name, func(t *testing.T) {
			f := newSignupFixture(t)
			req := validSignupRequest()
			mutate(&req)

			err := f.uc.Request(context.Background(), req)

			var validationErr *auth.SignupValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	}
}

func TestSelfSignup_Resend_OnlySendsForPendingSignup(t *testing.T) {
	f := newSignupFixture(t)

	require.NoError(t, f.uc.Resend(context.Background(), signupTxnID, signupEmail))
	f.emailSvc.AssertNotCalled(t, "SendSignupOTPEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	f.requestCode(t)
	f.emailSvc.On("SendSignupOTPEmail", mock.Anything, signupEmail, "New User", signupCode, "Acme Portal").Return(nil).Once()
	require.NoError(t, f.uc.Resend(context.Background(), signupTxnID, signupEmail))
	f.emailSvc.AssertNumberOfCalls(t, "SendSignupOTPEmail", 2)
}

func TestSelfSignup_Verify_CreatesActiveUserWithDefaultRoleAndLogsIn(t *testing.T) {
	f := newSignupFixture(t)
	f.requestCode(t)

	var created *entities.User
	f.endUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.User")).
		Run(func(args mock.Arguments) { created = args.Get(1).(*entities.User) }).Return(nil)
	f.roleRepo.On("Assign", mock.Anything, mock.AnythingOfType("*entities.UserRoleAssignment")).Return(nil)
	f.sessionRepo.On("Create", mock.Anything, mock.Anything, time.Hour).Return(nil)
	f.txnRepo.On("UpdateStage", mock.Anything, signupTxnID, repositories.TransactionStagePendingConsent, mock.Anything, mock.Anything).Return(nil)
	f.endUserRepo.On("UpdateLastLogin", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	resp, err := f.uc.Verify(context.Background(), verifyRequest(signupCode))

	require.NoError(t, err)
	assert.Equal(t, "https://sso.example.com/oauth2/consent?transaction_id="+signupTxnID, resp.RedirectURL)
	assert.NotEmpty(t, resp.SessionID)
	require.NotNil(t, created)
	assert.Equal(t, entities.UserStatusActive, created.Status)
	assert.Equal(t, "bcrypt-hash", created.PasswordHash)
	assert.Equal(t, signupTenantID, created.TenantID)
	assert.Equal(t, signupEmail, created.Email)

	assignment := f.roleRepo.Calls[0].Arguments.Get(1).(*entities.UserRoleAssignment)
	assert.Equal(t, signupRole, assignment.Role)
	assert.Equal(t, signupClientID, assignment.ClientID)
	assert.Equal(t, created.ID, assignment.UserID)
	assert.Equal(t, created.ID, assignment.GrantedBy)
	assert.True(t, assignment.IsActive)

	// Code is single-use.
	_, err = f.uc.Verify(context.Background(), verifyRequest(signupCode))
	assert.ErrorIs(t, err, auth.ErrInvalidSignupCode)
}

func TestSelfSignup_Verify_WrongCodeAttemptsAreCapped(t *testing.T) {
	f := newSignupFixture(t)
	f.requestCode(t)

	for i := 0; i < auth.SignupOTPMaxAttempts; i++ {
		_, err := f.uc.Verify(context.Background(), verifyRequest("000000"))
		assert.ErrorIs(t, err, auth.ErrInvalidSignupCode)
	}
	// Failed codes must not lock the address owner out of password login.
	f.throttler.AssertNotCalled(t, "RecordFailure", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	_, err := f.uc.Verify(context.Background(), verifyRequest(signupCode))
	assert.ErrorIs(t, err, auth.ErrInvalidSignupCode, "correct code must fail after attempts are exhausted")
	f.endUserRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestSelfSignup_CodeOnlyValidInTheTransactionThatRequestedIt(t *testing.T) {
	f := newSignupFixture(t)
	f.requestCode(t) // victim, transaction txn_signup

	// Another party replaces the pending sign-up from their own transaction.
	attackerTxn := *f.txn
	attackerTxn.TransactionID = "txn_attacker"
	f.txnRepo.On("Get", mock.Anything, "txn_attacker").Return(&attackerTxn, nil)
	f.txnRepo.On("ExtendTTL", mock.Anything, "txn_attacker", mock.Anything, mock.Anything).Return(nil)
	f.passwordSvc.On("Hash", "Att4cker!Pass").Return("attacker-hash", nil)
	f.store.sendCount = map[string]int{} // past the resend cooldown
	f.userRepo.On("EmailExists", mock.Anything, signupEmail).Return(false, nil).Once()
	f.emailSvc.On("SendSignupOTPEmail", mock.Anything, signupEmail, "Attacker", signupCode, "Acme Portal").Return(nil).Once()
	attackerReq := validSignupRequest()
	attackerReq.TransactionID = "txn_attacker"
	attackerReq.Password = "Att4cker!Pass"
	attackerReq.DisplayName = "Attacker"
	require.NoError(t, f.uc.Request(context.Background(), attackerReq))

	// The victim enters the newest emailed code in their own browser: rejected,
	// so no account is created with the attacker's password.
	_, err := f.uc.Verify(context.Background(), verifyRequest(signupCode))
	assert.ErrorIs(t, err, auth.ErrInvalidSignupCode)
	f.endUserRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)

	// A resend from a foreign transaction does not trigger an email either.
	require.NoError(t, f.uc.Resend(context.Background(), signupTxnID, signupEmail))
	f.emailSvc.AssertNumberOfCalls(t, "SendSignupOTPEmail", 2)
}

func TestSelfSignup_Verify_MalformedCodeRejectedWithoutLookup(t *testing.T) {
	f := newSignupFixture(t)

	_, err := f.uc.Verify(context.Background(), verifyRequest("12ab"))

	assert.ErrorIs(t, err, auth.ErrInvalidSignupCode)
	f.txnRepo.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
}

func TestSelfSignup_Verify_RoleAssignFailureRollsBackUser(t *testing.T) {
	f := newSignupFixture(t)
	f.requestCode(t)
	f.endUserRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	f.roleRepo.On("Assign", mock.Anything, mock.Anything).Return(errors.New("db down"))
	f.endUserRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)

	_, err := f.uc.Verify(context.Background(), verifyRequest(signupCode))

	assert.ErrorIs(t, err, auth.ErrTemporarilyUnavailable)
	f.endUserRepo.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
	f.sessionRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
}

func TestSelfSignup_Verify_DuplicateEmailRaceIsGenericFailure(t *testing.T) {
	f := newSignupFixture(t)
	f.requestCode(t)
	f.endUserRepo.On("Create", mock.Anything, mock.Anything).Return(repositories.ErrDuplicateEmail)

	_, err := f.uc.Verify(context.Background(), verifyRequest(signupCode))

	assert.ErrorIs(t, err, auth.ErrInvalidSignupCode)
	f.roleRepo.AssertNotCalled(t, "Assign", mock.Anything, mock.Anything)
}
