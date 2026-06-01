package integration

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthBrowserLogout_ClearsHostOnlyCookieAndRequiresLogin(t *testing.T) {
	harness := newOAuthBrowserHarness(t)
	_, transactionID := harness.begin(t, browserClientA, browserCallbackA, "", nil)
	loginResponse := harness.login(t, transactionID, "correct-password")
	cookie := responseCookie(t, loginResponse)

	logout := harness.request(t, http.MethodPost, "/oauth2/logout", nil, cookie, nil)
	require.Equal(t, http.StatusNoContent, logout.Code)
	expiredCookie := responseCookie(t, logout)
	assert.Equal(t, "keyles_sso", expiredCookie.Name)
	assert.Empty(t, expiredCookie.Domain)
	assert.True(t, expiredCookie.MaxAge < 0)
	assert.True(t, expiredCookie.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, expiredCookie.SameSite)
	_, ok := harness.sessions.sessions[cookie.Value]
	assert.False(t, ok)

	afterLogout, _ := harness.begin(t, browserClientA, browserCallbackA, "", cookie)
	requireFrontendRedirect(t, afterLogout, "/oauth2/login")
}

func TestOAuthBrowserLogout_MissingSessionAndRedisOutageStillClearCookie(t *testing.T) {
	harness := newOAuthBrowserHarness(t)

	missing := harness.request(t, http.MethodPost, "/oauth2/logout", nil, nil, nil)
	assert.Equal(t, http.StatusNoContent, missing.Code)
	assert.True(t, responseCookie(t, missing).MaxAge < 0)

	harness.sessions.getErr = errors.New("redis unavailable")
	harness.sessions.deleteErr = errors.New("redis unavailable")
	outage := harness.request(t, http.MethodPost, "/oauth2/logout", nil, &http.Cookie{
		Name:  "keyles_sso",
		Value: "unavailable-session",
	}, nil)
	assert.Equal(t, http.StatusNoContent, outage.Code)
	assert.True(t, responseCookie(t, outage).MaxAge < 0)
}
