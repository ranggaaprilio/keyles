package integration

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOAuthBrowserErrors_InvalidClientAndJSONRedisOutageStayLocal(t *testing.T) {
	harness := newOAuthBrowserHarness(t)

	invalidClient, _ := harness.begin(t, "unknown-client", browserCallbackA, "", nil)
	requireFrontendRedirect(t, invalidClient, "/oauth2/error")
	assert.Contains(t, invalidClient.Header().Get("Location"), "invalid_client")

	_, transactionID := harness.begin(t, browserClientA, browserCallbackA, "", nil)
	loginResponse := harness.login(t, transactionID, "correct-password")
	cookie := responseCookie(t, loginResponse)
	harness.sessions.getErr = errors.New("redis unavailable")

	details := harness.request(t, http.MethodGet, "/oauth2/consent/"+transactionID, nil, cookie, nil)
	assert.Equal(t, http.StatusServiceUnavailable, details.Code)
	assert.Contains(t, details.Body.String(), "temporarily_unavailable")
}
