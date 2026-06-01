package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/infrastructure/config"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/interfaces/http/middleware"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/stretchr/testify/require"
)

const (
	browserFrontendURL = "https://sso.example.com"
	browserTenantID    = "ed32b07b-7587-48cc-a4a4-e2d1b85c779f"
	browserUserID      = "a203ac31-44b0-4a6f-bf45-2dfb38fce6f3"
	browserClientA     = "browser_client_a"
	browserClientB     = "browser_client_b"
	browserCallbackA   = "https://app.example.com/callback?existing=keep"
	browserCallbackB   = "https://other.example.com/callback"
	browserDirectIP    = "198.51.100.25"
)

type browserTransactionRepository struct {
	transactions map[string]*repositories.AuthorizationTransaction
	createErr    error
	getErr       error
	updateErr    error
	completeErr  error
}

func newBrowserTransactionRepository() *browserTransactionRepository {
	return &browserTransactionRepository{transactions: make(map[string]*repositories.AuthorizationTransaction)}
}

func (r *browserTransactionRepository) Create(_ context.Context, txn *repositories.AuthorizationTransaction, _ time.Duration) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.transactions[txn.TransactionID] = txn
	return nil
}

func (r *browserTransactionRepository) Get(_ context.Context, transactionID string) (*repositories.AuthorizationTransaction, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	txn := r.transactions[transactionID]
	if txn == nil || time.Now().After(txn.ExpiresAt) {
		return nil, nil
	}
	return txn, nil
}

func (r *browserTransactionRepository) UpdateStage(_ context.Context, transactionID string, stage repositories.AuthorizationTransactionStage, userID, sessionID string) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	txn := r.transactions[transactionID]
	if txn == nil || time.Now().After(txn.ExpiresAt) {
		return errors.New(repositories.ErrTransactionNotFound)
	}
	txn.Stage = stage
	txn.UserID = userID
	txn.SessionID = sessionID
	return nil
}

func (r *browserTransactionRepository) Complete(_ context.Context, transactionID string) (*repositories.AuthorizationTransaction, error) {
	if r.completeErr != nil {
		return nil, r.completeErr
	}
	txn := r.transactions[transactionID]
	if txn == nil || time.Now().After(txn.ExpiresAt) {
		return nil, errors.New(repositories.ErrTransactionNotFound)
	}
	if txn.Stage == repositories.TransactionStageCompleted {
		return nil, errors.New(repositories.ErrTransactionAlreadyCompleted)
	}
	if txn.Stage != repositories.TransactionStagePendingConsent {
		return nil, errors.New(repositories.ErrTransactionWrongStage)
	}
	now := time.Now()
	txn.Stage = repositories.TransactionStageCompleted
	txn.CompletedAt = &now
	return txn, nil
}

type browserSessionRepository struct {
	sessions  map[string]*repositories.Session
	createErr error
	getErr    error
	deleteErr error
}

func newBrowserSessionRepository() *browserSessionRepository {
	return &browserSessionRepository{sessions: make(map[string]*repositories.Session)}
}

func (r *browserSessionRepository) Create(_ context.Context, session *repositories.Session, _ time.Duration) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.sessions[session.SessionID] = session
	return nil
}

func (r *browserSessionRepository) Get(_ context.Context, sessionID string) (*repositories.Session, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	session := r.sessions[sessionID]
	if session == nil || time.Now().After(session.ExpiresAt) {
		return nil, nil
	}
	return session, nil
}

func (r *browserSessionRepository) Delete(_ context.Context, sessionID string) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	delete(r.sessions, sessionID)
	return nil
}

func (r *browserSessionRepository) Exists(ctx context.Context, sessionID string) (bool, error) {
	session, err := r.Get(ctx, sessionID)
	return session != nil, err
}

func (r *browserSessionRepository) Extend(_ context.Context, sessionID string, ttl time.Duration) error {
	session := r.sessions[sessionID]
	if session == nil {
		return errors.New("session not found")
	}
	session.ExpiresAt = time.Now().Add(ttl)
	return nil
}

type browserLoginThrottler struct {
	ipFailures       map[string]int
	emailFailures    map[string]int
	maxFailures      int
	lastSourceIP     string
	isThrottledErr   error
	recordFailureErr error
}

func newBrowserLoginThrottler() *browserLoginThrottler {
	return &browserLoginThrottler{
		ipFailures:    make(map[string]int),
		emailFailures: make(map[string]int),
		maxFailures:   2,
	}
}

func (t *browserLoginThrottler) IsThrottled(_ context.Context, sourceIP, tenantID, normalizedEmail string) (bool, error) {
	t.lastSourceIP = sourceIP
	if t.isThrottledErr != nil {
		return false, t.isThrottledErr
	}
	return t.ipFailures[sourceIP] >= t.maxFailures || t.emailFailures[tenantID+":"+normalizedEmail] >= t.maxFailures, nil
}

func (t *browserLoginThrottler) RecordFailure(_ context.Context, sourceIP, tenantID, normalizedEmail string) error {
	t.lastSourceIP = sourceIP
	if t.recordFailureErr != nil {
		return t.recordFailureErr
	}
	t.ipFailures[sourceIP]++
	t.emailFailures[tenantID+":"+normalizedEmail]++
	return nil
}

func (t *browserLoginThrottler) ClearEmailBucket(_ context.Context, tenantID, normalizedEmail string) error {
	delete(t.emailFailures, tenantID+":"+normalizedEmail)
	return nil
}

type oauthBrowserHarness struct {
	router        *gin.Engine
	clients       *MockIntegrationClientRepository
	roles         *MockIntegrationRoleRepository
	users         *MockIntegrationEndUserRepository
	codes         *MockIntegrationAuthCodeRepository
	refreshTokens *MockRefreshTokenRepository
	transactions  *browserTransactionRepository
	sessions      *browserSessionRepository
	throttler     *browserLoginThrottler
	audits        *MockIntegrationAuditRepository
}

func newOAuthBrowserHarness(t *testing.T) *oauthBrowserHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)

	clients := NewMockIntegrationClientRepository()
	roles := NewMockIntegrationRoleRepository()
	users := NewMockIntegrationEndUserRepository()
	codes := NewMockIntegrationAuthCodeRepository()
	refreshTokens := NewMockRefreshTokenRepository()
	transactions := newBrowserTransactionRepository()
	sessions := newBrowserSessionRepository()
	throttler := newBrowserLoginThrottler()
	audits := &MockIntegrationAuditRepository{}

	clients.clients[browserClientA] = browserClient(browserClientA, "Browser Client A", browserCallbackA)
	clients.clients[browserClientB] = browserClient(browserClientB, "Browser Client B", browserCallbackB)
	users.users[browserUserID] = &entities.User{
		ID:           browserUserID,
		TenantID:     browserTenantID,
		Email:        "person@example.com",
		DisplayName:  "Browser User",
		PasswordHash: "hashed_correct-password",
		Status:       entities.UserStatusActive,
	}
	roles.roles[browserUserID+":"+browserClientA] = []*entities.UserRoleAssignment{browserRole(browserClientA)}
	roles.roles[browserUserID+":"+browserClientB] = []*entities.UserRoleAssignment{browserRole(browserClientB)}

	auditHelper := auth.NewOAuthAuditHelper(audits)
	interaction := auth.NewOAuthInteraction(clients, roles, transactions, sessions, users, auditHelper, browserFrontendURL)
	issueToken := auth.NewIssueToken(codes, clients, refreshTokens, roles, NewMockTokenService(), browserFrontendURL, users)
	login := auth.NewAuthenticateEndUser(transactions, sessions, clients, users, roles, throttler, &MockPasswordServiceIntegration{}, auditHelper, time.Hour, browserFrontendURL)
	consentDetails := auth.NewGetConsentDetails(transactions, sessions, clients, users)
	consent := auth.NewConsentDecision(transactions, sessions, clients, users, roles, codes, auditHelper)
	logout := auth.NewLogoutEndUser(sessions, auditHelper)
	oauthHandler := handlers.NewOAuthHandlerFullBrowser(nil, issueToken, clients, nil, nil, nil, interaction, login, consentDetails, consent, logout, sessions, throttler, auditHelper, &config.Config{
		FrontendURL:          browserFrontendURL,
		SecuritySessionTTL:   3600,
		SecurityCookieSecure: true,
	})

	router := gin.New()
	router.Use(middleware.CORS(browserFrontendURL, "GET, POST, OPTIONS", "Content-Type"))
	router.GET("/oauth2/auth", oauthHandler.AuthorizeBrowser)
	router.POST("/oauth2/login", oauthHandler.Login)
	router.GET("/oauth2/consent/:transactionId", oauthHandler.ConsentDetail)
	router.POST("/oauth2/consent", oauthHandler.ConsentDecision)
	router.POST("/oauth2/logout", oauthHandler.Logout)
	router.POST("/oauth2/token", oauthHandler.Token)

	return &oauthBrowserHarness{
		router:        router,
		clients:       clients,
		roles:         roles,
		users:         users,
		codes:         codes,
		refreshTokens: refreshTokens,
		transactions:  transactions,
		sessions:      sessions,
		throttler:     throttler,
		audits:        audits,
	}
}

func browserClient(clientID, name, redirectURI string) *entities.Client {
	return &entities.Client{
		ClientID:            clientID,
		TenantID:            browserTenantID,
		ClientName:          name,
		ClientType:          entities.ClientTypePublic,
		AllowedRedirectURIs: []string{redirectURI},
		IsActive:            true,
	}
}

func browserRole(clientID string) *entities.UserRoleAssignment {
	return &entities.UserRoleAssignment{
		UserID:   browserUserID,
		ClientID: clientID,
		TenantID: browserTenantID,
		Role:     "member",
		IsActive: true,
	}
}

func (h *oauthBrowserHarness) begin(t *testing.T, clientID, redirectURI, prompt string, cookie *http.Cookie) (*httptest.ResponseRecorder, string) {
	return h.beginWithChallenge(t, clientID, redirectURI, prompt, "browser-pkce-challenge", cookie)
}

func (h *oauthBrowserHarness) beginWithChallenge(t *testing.T, clientID, redirectURI, prompt, challenge string, cookie *http.Cookie) (*httptest.ResponseRecorder, string) {
	t.Helper()
	values := url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"response_type":         {"code"},
		"scope":                 {"openid email"},
		"state":                 {"state-123"},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}
	if prompt != "" {
		values.Set("prompt", prompt)
	}
	recorder := h.request(t, http.MethodGet, "/oauth2/auth?"+values.Encode(), nil, cookie, map[string]string{
		"X-Forwarded-For": "203.0.113.99",
	})
	location := recorder.Header().Get("Location")
	parsed, err := url.Parse(location)
	require.NoError(t, err)
	return recorder, parsed.Query().Get("transaction_id")
}

func (h *oauthBrowserHarness) login(t *testing.T, transactionID, password string) *httptest.ResponseRecorder {
	t.Helper()
	return h.request(t, http.MethodPost, "/oauth2/login", map[string]any{
		"transaction_id": transactionID,
		"email":          "person@example.com",
		"password":       password,
	}, nil, map[string]string{
		"X-Forwarded-For": "203.0.113.99",
	})
}

func (h *oauthBrowserHarness) consent(t *testing.T, transactionID, csrfToken string, approved bool, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	return h.request(t, http.MethodPost, "/oauth2/consent", map[string]any{
		"transaction_id":         transactionID,
		"interaction_csrf_token": csrfToken,
		"approved":               approved,
	}, cookie, nil)
}

func (h *oauthBrowserHarness) request(t *testing.T, method, target string, body any, cookie *http.Cookie, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var payload *bytes.Reader
	if body == nil {
		payload = bytes.NewReader(nil)
	} else {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		payload = bytes.NewReader(data)
	}
	request := httptest.NewRequest(method, target, payload)
	request.RemoteAddr = browserDirectIP + ":4321"
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		request.AddCookie(cookie)
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	h.router.ServeHTTP(recorder, request)
	return recorder
}

func responseCookie(t *testing.T, recorder *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	cookies := recorder.Result().Cookies()
	require.NotEmpty(t, cookies)
	return cookies[0]
}

func redirectQuery(t *testing.T, recorder *httptest.ResponseRecorder) url.Values {
	t.Helper()
	var response struct {
		RedirectURL string `json:"redirect_url"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	parsed, err := url.Parse(response.RedirectURL)
	require.NoError(t, err)
	return parsed.Query()
}

func requireFrontendRedirect(t *testing.T, recorder *httptest.ResponseRecorder, path string) {
	t.Helper()
	require.Equal(t, http.StatusFound, recorder.Code)
	require.True(t, strings.HasPrefix(recorder.Header().Get("Location"), browserFrontendURL+path), recorder.Header().Get("Location"))
}

func TestDockerComposeOAuthBrowserFlow(t *testing.T) {
	if os.Getenv("KEYLES_COMPOSE_E2E") != "1" {
		t.Skip("set KEYLES_COMPOSE_E2E=1 to run the live Docker Compose OAuth matrix")
	}

	command := exec.Command("bash", "oauth_compose_e2e.sh")
	command.Env = os.Environ()
	output, err := command.CombinedOutput()
	t.Log(string(output))
	require.NoError(t, err)
}
