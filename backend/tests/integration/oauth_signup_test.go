package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/infrastructure/config"
	infraServices "github.com/ranggaaprilio/keyles/infrastructure/services"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const signupNewEmail = "fresh@example.com"

// memorySignupStore is an in-memory PendingSignupStore for HTTP-level tests.
type memorySignupStore struct {
	records  map[string]*repositories.PendingSignup
	attempts map[string]int
}

func (s *memorySignupStore) key(tenantID, email string) string { return tenantID + "|" + email }

func (s *memorySignupStore) AllowSend(context.Context, string, string, time.Duration, int) (bool, error) {
	return true, nil
}

func (s *memorySignupStore) Save(_ context.Context, p *repositories.PendingSignup, _ time.Duration) error {
	cp := *p
	s.records[s.key(p.TenantID, p.Email)] = &cp
	s.attempts[s.key(p.TenantID, p.Email)] = 0
	return nil
}

func (s *memorySignupStore) Get(_ context.Context, tenantID, email string) (*repositories.PendingSignup, error) {
	return s.records[s.key(tenantID, email)], nil
}

func (s *memorySignupStore) Verify(_ context.Context, tenantID, email, codeHash string, maxAttempts int) (*repositories.PendingSignup, error) {
	k := s.key(tenantID, email)
	rec := s.records[k]
	if rec == nil {
		return nil, nil
	}
	if rec.CodeHash == codeHash {
		delete(s.records, k)
		return rec, nil
	}
	if s.attempts[k]++; s.attempts[k] >= maxAttempts {
		delete(s.records, k)
	}
	return nil, nil
}

func (s *memorySignupStore) Delete(_ context.Context, tenantID, email string) error {
	delete(s.records, s.key(tenantID, email))
	return nil
}

// signupEmailCapture records the codes "emailed" to each address.
type signupEmailCapture struct {
	stubEmailService
	codes map[string]string
}

func (c *signupEmailCapture) SendSignupOTPEmail(_ context.Context, toEmail, _, otpCode, _ string) error {
	c.codes[toEmail] = otpCode
	return nil
}

// globalEmailIndex answers EmailExists from the harness end-user store.
type globalEmailIndex struct {
	repositories.UserRepository
	users *MockIntegrationEndUserRepository
}

func (g *globalEmailIndex) EmailExists(_ context.Context, email string) (bool, error) {
	for _, u := range g.users.users {
		if u.Email == email {
			return true, nil
		}
	}
	return false, nil
}

type signupHarness struct {
	*oauthBrowserHarness
	emails *signupEmailCapture
}

func newSignupHarness(t *testing.T) *signupHarness {
	t.Helper()
	h := newOAuthBrowserHarness(t)
	client := h.clients.clients[browserClientA]
	client.AllowSelfSignup = true
	client.SelfSignupRole = "member"

	emails := &signupEmailCapture{codes: map[string]string{}}
	auditHelper := auth.NewOAuthAuditHelper(h.audits)
	completer := auth.NewLoginCompleter(h.transactions, h.sessions, h.users, h.throttler, auditHelper, time.Hour, browserFrontendURL)
	selfSignup := auth.NewSelfSignup(
		h.transactions, h.clients, h.users, &globalEmailIndex{users: h.users}, h.roles,
		&memorySignupStore{records: map[string]*repositories.PendingSignup{}, attempts: map[string]int{}},
		&MockPasswordServiceIntegration{}, infraServices.NewCryptoOTPService(), emails,
		newStubUserCountCache(), auditHelper, completer,
	)
	handler := handlers.NewOAuthSignupHandler(selfSignup, &config.Config{SecuritySessionTTL: 3600, SecurityCookieSecure: true})
	h.router.GET("/oauth2/signup/context", handler.Context)
	h.router.POST("/oauth2/signup", handler.Signup)
	h.router.POST("/oauth2/signup/resend", handler.Resend)
	h.router.POST("/oauth2/signup/verify", handler.Verify)

	return &signupHarness{oauthBrowserHarness: h, emails: emails}
}

func (h *signupHarness) signup(t *testing.T, transactionID, email string) *httptest.ResponseRecorder {
	t.Helper()
	return h.request(t, http.MethodPost, "/oauth2/signup", map[string]any{
		"transaction_id": transactionID,
		"email":          email,
		"password":       "Str0ng!Passw0rd",
		"display_name":   "Fresh Person",
	}, nil, nil)
}

func (h *signupHarness) verify(t *testing.T, transactionID, email, code string) *httptest.ResponseRecorder {
	t.Helper()
	return h.request(t, http.MethodPost, "/oauth2/signup/verify", map[string]any{
		"transaction_id": transactionID,
		"email":          email,
		"code":           code,
	}, nil, nil)
}

func TestOAuthSignup_FullFlowIssuesAuthorizationCode(t *testing.T) {
	h := newSignupHarness(t)
	_, transactionID := h.begin(t, browserClientA, browserCallbackA, "", nil)

	ctxResp := h.request(t, http.MethodGet, "/oauth2/signup/context?transaction_id="+transactionID, nil, nil, nil)
	require.Equal(t, http.StatusOK, ctxResp.Code)
	assert.JSONEq(t, `{"client_name":"Browser Client A","signup_enabled":true}`, ctxResp.Body.String())

	signupResp := h.signup(t, transactionID, "Fresh@Example.com")
	require.Equal(t, http.StatusAccepted, signupResp.Code)
	code := h.emails.codes[signupNewEmail]
	require.Len(t, code, 6)

	verifyResp := h.verify(t, transactionID, signupNewEmail, code)
	require.Equal(t, http.StatusOK, verifyResp.Code, verifyResp.Body.String())
	cookie := responseCookie(t, verifyResp)
	assert.Equal(t, "keyles_sso", cookie.Name)
	assert.Contains(t, verifyResp.Header().Get("Set-Cookie"), "SameSite=Lax")
	assert.Equal(t, transactionID, redirectQuery(t, verifyResp).Get("transaction_id"))

	// The new session carries the user straight through consent.
	csrfToken := h.transactions.transactions[transactionID].InteractionCSRFToken
	approved := h.consent(t, transactionID, csrfToken, true, cookie)
	require.Equal(t, http.StatusOK, approved.Code, approved.Body.String())
	assert.NotEmpty(t, redirectQuery(t, approved).Get("code"))
}

func TestOAuthSignup_ExistingEmailResponseMatchesNewEmail(t *testing.T) {
	h := newSignupHarness(t)
	_, transactionID := h.begin(t, browserClientA, browserCallbackA, "", nil)

	fresh := h.signup(t, transactionID, signupNewEmail)
	existing := h.signup(t, transactionID, "person@example.com")

	assert.Equal(t, fresh.Code, existing.Code)
	assert.Equal(t, fresh.Body.String(), existing.Body.String())
	_, sent := h.emails.codes["person@example.com"]
	assert.False(t, sent, "no code may be sent to an already registered address")
}

func TestOAuthSignup_WrongCodeIsRejectedGenerically(t *testing.T) {
	h := newSignupHarness(t)
	_, transactionID := h.begin(t, browserClientA, browserCallbackA, "", nil)
	require.Equal(t, http.StatusAccepted, h.signup(t, transactionID, signupNewEmail).Code)

	wrong := "000000"
	if h.emails.codes[signupNewEmail] == wrong {
		wrong = "111111"
	}
	resp := h.verify(t, transactionID, signupNewEmail, wrong)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &body))
	assert.Equal(t, "invalid_code", body["error"])
	assert.Empty(t, resp.Header().Get("Set-Cookie"))
}

func TestOAuthSignup_DisabledClientAndBadInput(t *testing.T) {
	h := newSignupHarness(t)
	_, transactionID := h.begin(t, browserClientB, browserCallbackB, "", nil)

	ctxResp := h.request(t, http.MethodGet, "/oauth2/signup/context?transaction_id="+transactionID, nil, nil, nil)
	require.Equal(t, http.StatusOK, ctxResp.Code)
	assert.JSONEq(t, `{"client_name":"Browser Client B","signup_enabled":false}`, ctxResp.Body.String())
	assert.Equal(t, http.StatusForbidden, h.signup(t, transactionID, signupNewEmail).Code)

	unknown := h.request(t, http.MethodGet, "/oauth2/signup/context?transaction_id=missing", nil, nil, nil)
	assert.Equal(t, http.StatusNotFound, unknown.Code)

	_, txnA := h.begin(t, browserClientA, browserCallbackA, "", nil)
	weak := h.request(t, http.MethodPost, "/oauth2/signup", map[string]any{
		"transaction_id": txnA, "email": signupNewEmail, "password": "weak", "display_name": "Fresh Person",
	}, nil, nil)
	assert.Equal(t, http.StatusBadRequest, weak.Code)
}

func TestOAuthSignup_CodeIsBoundToTheRequestingTransaction(t *testing.T) {
	h := newSignupHarness(t)
	_, victimTxn := h.begin(t, browserClientA, browserCallbackA, "", nil)
	_, otherTxn := h.begin(t, browserClientA, browserCallbackA, "", nil)
	require.Equal(t, http.StatusAccepted, h.signup(t, victimTxn, signupNewEmail).Code)

	resp := h.verify(t, otherTxn, signupNewEmail, h.emails.codes[signupNewEmail])

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	_, err := h.users.GetByEmail(context.Background(), browserTenantID, signupNewEmail)
	assert.Error(t, err, "no user may be created from another transaction")
}
