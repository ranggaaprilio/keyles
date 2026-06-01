package integration

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthBrowserSession_CrossClientReuseAndEligibility(t *testing.T) {
	harness := newOAuthBrowserHarness(t)
	_, transactionID := harness.begin(t, browserClientA, browserCallbackA, "", nil)
	loginResponse := harness.login(t, transactionID, "correct-password")
	cookie := responseCookie(t, loginResponse)

	reused, _ := harness.begin(t, browserClientB, browserCallbackB, "", cookie)
	requireFrontendRedirect(t, reused, "/oauth2/consent")

	harness.users.users[browserUserID].Status = entities.UserStatusDisabled
	disabled, _ := harness.begin(t, browserClientB, browserCallbackB, "", cookie)
	requireFrontendRedirect(t, disabled, "/oauth2/login")

	harness.users.users[browserUserID].Status = entities.UserStatusActive
	harness.roles.roles[browserUserID+":"+browserClientB] = nil
	roleRemoved, _ := harness.begin(t, browserClientB, browserCallbackB, "", cookie)
	requireFrontendRedirect(t, roleRemoved, "/oauth2/login")

	harness.roles.roles[browserUserID+":"+browserClientB] = []*entities.UserRoleAssignment{browserRole(browserClientB)}
	forcedLogin, _ := harness.begin(t, browserClientB, browserCallbackB, "login", cookie)
	requireFrontendRedirect(t, forcedLogin, "/oauth2/login")
}

func TestOAuthBrowserSession_PromptNoneReturnsOIDCErrors(t *testing.T) {
	harness := newOAuthBrowserHarness(t)

	withoutSession, _ := harness.begin(t, browserClientA, browserCallbackA, "none", nil)
	require.Equal(t, http.StatusFound, withoutSession.Code)
	callback, err := url.Parse(withoutSession.Header().Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, "login_required", callback.Query().Get("error"))
	assert.Equal(t, "state-123", callback.Query().Get("state"))

	_, transactionID := harness.begin(t, browserClientA, browserCallbackA, "", nil)
	cookie := responseCookie(t, harness.login(t, transactionID, "correct-password"))
	withSession, _ := harness.begin(t, browserClientA, browserCallbackA, "none", cookie)
	callback, err = url.Parse(withSession.Header().Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, "consent_required", callback.Query().Get("error"))
	assert.Equal(t, "state-123", callback.Query().Get("state"))
}
