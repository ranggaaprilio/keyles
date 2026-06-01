package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthBrowserConsent_ApprovePreservesCallbackAndRejectsReplay(t *testing.T) {
	harness := newOAuthBrowserHarness(t)
	_, transactionID := harness.begin(t, browserClientA, browserCallbackA, "", nil)
	loginResponse := harness.login(t, transactionID, "correct-password")
	require.Equal(t, http.StatusOK, loginResponse.Code)
	cookie := responseCookie(t, loginResponse)

	details := harness.request(t, http.MethodGet, "/oauth2/consent/"+transactionID, nil, cookie, nil)
	require.Equal(t, http.StatusOK, details.Code)
	var payload struct {
		TransactionID        string   `json:"transaction_id"`
		InteractionCSRFToken string   `json:"interaction_csrf_token"`
		ClientName           string   `json:"client_name"`
		Scopes               []string `json:"scopes"`
		UserDisplay          string   `json:"user_display"`
	}
	require.NoError(t, json.Unmarshal(details.Body.Bytes(), &payload))
	assert.Equal(t, transactionID, payload.TransactionID)
	assert.Equal(t, "Browser Client A", payload.ClientName)
	assert.Equal(t, []string{"openid", "email"}, payload.Scopes)
	assert.Equal(t, "Browser User", payload.UserDisplay)

	approved := harness.consent(t, transactionID, payload.InteractionCSRFToken, true, cookie)
	require.Equal(t, http.StatusOK, approved.Code)
	query := redirectQuery(t, approved)
	assert.Equal(t, "keep", query.Get("existing"))
	assert.Equal(t, "state-123", query.Get("state"))
	assert.NotEmpty(t, query.Get("code"))
	_, ok := harness.codes.codes[query.Get("code")]
	assert.True(t, ok)

	replay := harness.consent(t, transactionID, payload.InteractionCSRFToken, true, cookie)
	assert.Equal(t, http.StatusGone, replay.Code)
}

func TestOAuthBrowserConsent_DenyPreservesState(t *testing.T) {
	harness := newOAuthBrowserHarness(t)
	_, transactionID := harness.begin(t, browserClientA, browserCallbackA, "", nil)
	loginResponse := harness.login(t, transactionID, "correct-password")
	cookie := responseCookie(t, loginResponse)
	csrfToken := harness.transactions.transactions[transactionID].InteractionCSRFToken

	denied := harness.consent(t, transactionID, csrfToken, false, cookie)
	require.Equal(t, http.StatusOK, denied.Code)
	query := redirectQuery(t, denied)
	assert.Equal(t, "access_denied", query.Get("error"))
	assert.Equal(t, "state-123", query.Get("state"))
	assert.Equal(t, "keep", query.Get("existing"))
}
