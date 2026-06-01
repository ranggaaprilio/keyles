package integration

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthBrowserAuth_LoginCookieThrottleAndCORS(t *testing.T) {
	harness := newOAuthBrowserHarness(t)

	authResponse, transactionID := harness.begin(t, browserClientA, browserCallbackA, "", nil)
	requireFrontendRedirect(t, authResponse, "/oauth2/login")
	require.NotEmpty(t, transactionID)

	preflight := harness.request(t, http.MethodOptions, "/oauth2/login", nil, nil, map[string]string{
		"Origin": "https://sso.example.com",
	})
	assert.Equal(t, http.StatusNoContent, preflight.Code)
	assert.Equal(t, browserFrontendURL, preflight.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", preflight.Header().Get("Access-Control-Allow-Credentials"))
	assert.NotEqual(t, "*", preflight.Header().Get("Access-Control-Allow-Origin"))

	loginResponse := harness.login(t, transactionID, "correct-password")
	require.Equal(t, http.StatusOK, loginResponse.Code)
	assert.Equal(t, browserDirectIP, harness.throttler.lastSourceIP)
	cookie := responseCookie(t, loginResponse)
	assert.Equal(t, "keyles_sso", cookie.Name)
	assert.Empty(t, cookie.Domain)
	assert.True(t, cookie.HttpOnly)
	assert.True(t, cookie.Secure)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)

	_, failedTransactionID := harness.begin(t, browserClientA, browserCallbackA, "", nil)
	assert.Equal(t, http.StatusUnauthorized, harness.login(t, failedTransactionID, "wrong-password").Code)
	assert.Equal(t, http.StatusUnauthorized, harness.login(t, failedTransactionID, "wrong-password").Code)
	assert.Equal(t, http.StatusTooManyRequests, harness.login(t, failedTransactionID, "wrong-password").Code)
	assert.Equal(t, browserDirectIP, harness.throttler.lastSourceIP)
}

func TestOAuthBrowserAuth_InvalidCallbackAndRedisOutageStayLocal(t *testing.T) {
	harness := newOAuthBrowserHarness(t)

	invalidCallback, _ := harness.begin(t, browserClientA, "https://attacker.example.com/callback", "", nil)
	requireFrontendRedirect(t, invalidCallback, "/oauth2/error")
	assert.Contains(t, invalidCallback.Header().Get("Location"), "invalid_redirect_uri")
	assert.False(t, strings.Contains(invalidCallback.Header().Get("Location"), "attacker.example.com"))

	harness.transactions.createErr = errors.New("redis unavailable")
	outage, _ := harness.begin(t, browserClientA, browserCallbackA, "", nil)
	requireFrontendRedirect(t, outage, "/oauth2/error")
	assert.Contains(t, outage.Header().Get("Location"), "temporarily_unavailable")
}
