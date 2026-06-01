package handlers

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
	"github.com/ranggaaprilio/keyles/infrastructure/config"
	"github.com/ranggaaprilio/keyles/usecase/auth"
)

// OAuthHandler handles OAuth 2.0 / OIDC endpoints
type OAuthHandler struct {
	authorizeClientUC   *auth.AuthorizeClient
	issueTokenUC        *auth.IssueToken
	refreshTokenUC      *auth.RefreshToken
	revokeTokenUC       *auth.RevokeToken
	introspectTokenUC   *auth.IntrospectToken
	clientRepo          repositories.ClientRepository
	oauthInteraction    *auth.OAuthInteraction
	authenticateEndUser *auth.AuthenticateEndUser
	getConsentDetails   *auth.GetConsentDetails
	consentDecision     *auth.ConsentDecision
	logoutEndUser       *auth.LogoutEndUser
	sessionRepo         repositories.SessionRepository
	loginThrottler      services.LoginThrottler
	auditHelper         *auth.OAuthAuditHelper
	config              *config.Config
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(
	authorizeClientUC *auth.AuthorizeClient,
	issueTokenUC *auth.IssueToken,
	clientRepo repositories.ClientRepository,
) *OAuthHandler {
	return &OAuthHandler{
		authorizeClientUC: authorizeClientUC,
		issueTokenUC:      issueTokenUC,
		clientRepo:        clientRepo,
	}
}

// NewOAuthHandlerWithRefresh creates a new OAuth handler with refresh token support
func NewOAuthHandlerWithRefresh(
	authorizeClientUC *auth.AuthorizeClient,
	issueTokenUC *auth.IssueToken,
	clientRepo repositories.ClientRepository,
	refreshTokenUC *auth.RefreshToken,
) *OAuthHandler {
	return &OAuthHandler{
		authorizeClientUC: authorizeClientUC,
		issueTokenUC:      issueTokenUC,
		refreshTokenUC:    refreshTokenUC,
		clientRepo:        clientRepo,
	}
}

// NewOAuthHandlerWithRevoke creates a new OAuth handler with revoke token support
func NewOAuthHandlerWithRevoke(
	authorizeClientUC *auth.AuthorizeClient,
	issueTokenUC *auth.IssueToken,
	clientRepo repositories.ClientRepository,
	revokeTokenUC *auth.RevokeToken,
) *OAuthHandler {
	return &OAuthHandler{
		authorizeClientUC: authorizeClientUC,
		issueTokenUC:      issueTokenUC,
		revokeTokenUC:     revokeTokenUC,
		clientRepo:        clientRepo,
	}
}

// NewOAuthHandlerFull creates a new OAuth handler with all features
func NewOAuthHandlerFull(
	authorizeClientUC *auth.AuthorizeClient,
	issueTokenUC *auth.IssueToken,
	clientRepo repositories.ClientRepository,
	refreshTokenUC *auth.RefreshToken,
	revokeTokenUC *auth.RevokeToken,
	introspectTokenUC *auth.IntrospectToken,
) *OAuthHandler {
	return &OAuthHandler{
		authorizeClientUC: authorizeClientUC,
		issueTokenUC:      issueTokenUC,
		refreshTokenUC:    refreshTokenUC,
		revokeTokenUC:     revokeTokenUC,
		introspectTokenUC: introspectTokenUC,
		clientRepo:        clientRepo,
	}
}

// NewOAuthHandlerFullBrowser creates a new OAuth handler with all features
// including the browser-facing login and consent flow.
func NewOAuthHandlerFullBrowser(
	authorizeClientUC *auth.AuthorizeClient,
	issueTokenUC *auth.IssueToken,
	clientRepo repositories.ClientRepository,
	refreshTokenUC *auth.RefreshToken,
	revokeTokenUC *auth.RevokeToken,
	introspectTokenUC *auth.IntrospectToken,
	oauthInteraction *auth.OAuthInteraction,
	authenticateEndUser *auth.AuthenticateEndUser,
	getConsentDetails *auth.GetConsentDetails,
	consentDecision *auth.ConsentDecision,
	logoutEndUser *auth.LogoutEndUser,
	sessionRepo repositories.SessionRepository,
	loginThrottler services.LoginThrottler,
	auditHelper *auth.OAuthAuditHelper,
	config *config.Config,
) *OAuthHandler {
	return &OAuthHandler{
		authorizeClientUC:   authorizeClientUC,
		issueTokenUC:        issueTokenUC,
		refreshTokenUC:      refreshTokenUC,
		revokeTokenUC:       revokeTokenUC,
		introspectTokenUC:   introspectTokenUC,
		clientRepo:          clientRepo,
		oauthInteraction:    oauthInteraction,
		authenticateEndUser: authenticateEndUser,
		getConsentDetails:   getConsentDetails,
		consentDecision:     consentDecision,
		logoutEndUser:       logoutEndUser,
		sessionRepo:         sessionRepo,
		loginThrottler:      loginThrottler,
		auditHelper:         auditHelper,
		config:              config,
	}
}

// Authorize handles the OAuth 2.0 authorization endpoint (GET /oauth2/auth)
// Per RFC 6749 Section 4.1.1 and OpenID Connect Core 1.0 Section 3.1.2.1
func (h *OAuthHandler) Authorize(c *gin.Context) {
	// Extract query parameters per FR-009
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	responseType := c.Query("response_type")
	scope := c.Query("scope")
	state := c.Query("state")
	codeChallenge := c.Query("code_challenge")
	codeChallengeMethod := c.Query("code_challenge_method")

	// Get user context from headers (in real implementation, this would come from session/cookie)
	userID := c.GetHeader("X-User-ID")
	// If user is not authenticated, we would redirect to login page
	// For now, we require the headers to be set
	if userID == "" {
		// In a real implementation, we would redirect to login page
		// preserving the authorization request parameters
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "login_required",
			"error_description": "User must be authenticated to authorize",
		})
		return
	}

	// Build authorization request
	req := auth.AuthorizeRequest{
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		ResponseType:        responseType,
		Scope:               scope,
		State:               state,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		UserID:              userID,
	}

	// Execute authorization
	resp, err := h.authorizeClientUC.Execute(c.Request.Context(), req)
	if err != nil {
		// Handle OAuth errors
		oauthErr, ok := err.(*auth.OAuthError)
		if ok {
			statusCode := mapOAuthErrorToStatus(oauthErr.Code)
			c.JSON(statusCode, gin.H{
				"error":             oauthErr.Code,
				"error_description": oauthErr.Description,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": err.Error(),
		})
		return
	}

	// Build redirect URL with authorization code and state (FR-016)
	redirectURL, err := url.Parse(resp.RedirectURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "Failed to build redirect URL",
		})
		return
	}

	query := redirectURL.Query()
	query.Set("code", resp.Code)
	query.Set("state", resp.State)
	redirectURL.RawQuery = query.Encode()

	// Redirect to client with authorization code (HTTP 302)
	c.Redirect(http.StatusFound, redirectURL.String())
}

// Token handles the OAuth 2.0 token endpoint (POST /oauth2/token)
// Per RFC 6749 Section 4.1.3 (Authorization Code Exchange) and Section 6 (Refresh Token)
func (h *OAuthHandler) Token(c *gin.Context) {
	// Extract form parameters (application/x-www-form-urlencoded per RFC 6749)
	grantType := c.PostForm("grant_type")
	clientID := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")

	// Support Basic Auth for client credentials
	if clientID == "" || clientSecret == "" {
		basicClientID, basicClientSecret, hasBasic := c.Request.BasicAuth()
		if hasBasic {
			if clientID == "" {
				clientID = basicClientID
			}
			if clientSecret == "" {
				clientSecret = basicClientSecret
			}
		}
	}

	// Route based on grant_type
	switch grantType {
	case "authorization_code":
		h.handleAuthorizationCodeGrant(c, clientID, clientSecret)
	case "refresh_token":
		h.handleRefreshTokenGrant(c, clientID, clientSecret)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             auth.ErrUnsupportedGrantType,
			"error_description": "only grant_type=authorization_code and grant_type=refresh_token are supported",
		})
	}
}

// handleAuthorizationCodeGrant handles the authorization_code grant type
func (h *OAuthHandler) handleAuthorizationCodeGrant(c *gin.Context, clientID, clientSecret string) {
	code := c.PostForm("code")
	redirectURI := c.PostForm("redirect_uri")
	codeVerifier := c.PostForm("code_verifier")

	// Build token request
	req := auth.TokenRequest{
		GrantType:    "authorization_code",
		Code:         code,
		RedirectURI:  redirectURI,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CodeVerifier: codeVerifier,
	}

	// Execute token exchange
	resp, err := h.issueTokenUC.Execute(c.Request.Context(), req)
	if err != nil {
		h.handleOAuthError(c, err)
		return
	}

	// Return token response per RFC 6749 Section 5.1
	c.JSON(http.StatusOK, resp)
}

// handleRefreshTokenGrant handles the refresh_token grant type (FR-043 through FR-047)
func (h *OAuthHandler) handleRefreshTokenGrant(c *gin.Context, clientID, clientSecret string) {
	refreshToken := c.PostForm("refresh_token")
	scope := c.PostForm("scope") // Optional: can request reduced scope

	// Check if refresh token use case is available
	if h.refreshTokenUC == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             auth.ErrUnsupportedGrantType,
			"error_description": "refresh_token grant type is not enabled",
		})
		return
	}

	// Build refresh token request
	req := auth.RefreshTokenRequest{
		RefreshToken: refreshToken,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scope:        scope,
	}

	// Execute refresh token exchange
	resp, err := h.refreshTokenUC.Execute(c.Request.Context(), req)
	if err != nil {
		h.handleOAuthError(c, err)
		return
	}

	// Return token response
	c.JSON(http.StatusOK, resp)
}

// handleOAuthError handles OAuth errors and returns appropriate HTTP response
func (h *OAuthHandler) handleOAuthError(c *gin.Context, err error) {
	oauthErr, ok := err.(*auth.OAuthError)
	if ok {
		statusCode := mapOAuthErrorToStatus(oauthErr.Code)
		c.JSON(statusCode, gin.H{
			"error":             oauthErr.Code,
			"error_description": oauthErr.Description,
		})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"error":             "server_error",
		"error_description": err.Error(),
	})
}

// mapOAuthErrorToStatus maps OAuth error codes to HTTP status codes
func mapOAuthErrorToStatus(errorCode string) int {
	switch errorCode {
	case auth.ErrInvalidRequest:
		return http.StatusBadRequest
	case auth.ErrUnauthorizedClient:
		return http.StatusUnauthorized
	case auth.ErrAccessDenied:
		return http.StatusForbidden
	case auth.ErrUnsupportedResponseType:
		return http.StatusBadRequest
	case auth.ErrInvalidScope:
		return http.StatusBadRequest
	case auth.ErrServerError:
		return http.StatusInternalServerError
	case auth.ErrInvalidClient:
		return http.StatusUnauthorized
	case auth.ErrInvalidGrant:
		return http.StatusBadRequest
	case auth.ErrUnsupportedGrantType:
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}

// Revoke handles the OAuth 2.0 token revocation endpoint (POST /oauth2/revoke)
// Per RFC 7009 - OAuth 2.0 Token Revocation
func (h *OAuthHandler) Revoke(c *gin.Context) {
	// Extract form parameters
	token := c.PostForm("token")
	tokenTypeHint := c.PostForm("token_type_hint")
	clientID, clientSecret := clientCredentials(c)

	if h.clientRepo != nil {
		if _, err := auth.AuthenticateOAuthClient(c.Request.Context(), h.clientRepo, clientID, clientSecret, true); err != nil {
			h.handleOAuthError(c, err)
			return
		}
	}

	// Check if revoke token use case is available
	if h.revokeTokenUC == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             auth.ErrServerError,
			"error_description": "token revocation is not enabled",
		})
		return
	}

	// Build revoke token request
	req := auth.RevokeTokenRequest{
		Token:         token,
		TokenTypeHint: tokenTypeHint,
		ClientID:      clientID,
	}

	// Execute token revocation
	err := h.revokeTokenUC.Execute(c.Request.Context(), req)
	if err != nil {
		h.handleOAuthError(c, err)
		return
	}

	// Per RFC 7009, return HTTP 200 OK with empty body on success
	c.Status(http.StatusOK)
}

// Introspect returns active OAuth token metadata to the confidential client
// that owns the token audience.
func (h *OAuthHandler) Introspect(c *gin.Context) {
	if h.introspectTokenUC == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": auth.ErrUnsupportedGrantType})
		return
	}

	clientID, clientSecret := clientCredentials(c)
	response, err := h.introspectTokenUC.Execute(c.Request.Context(), c.PostForm("token"), clientID, clientSecret)
	if err != nil {
		h.handleOAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func clientCredentials(c *gin.Context) (string, string) {
	clientID := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")
	if basicClientID, basicClientSecret, ok := c.Request.BasicAuth(); ok {
		if clientID == "" {
			clientID = basicClientID
		}
		if clientSecret == "" {
			clientSecret = basicClientSecret
		}
	}
	return clientID, clientSecret
}

// AuthorizeBrowser handles GET /oauth2/auth for the browser-facing flow.
// It reads the SSO session cookie and delegates to OAuthInteraction.InitializeAuth
// to validate the request, create an authorization transaction, and determine
// whether to redirect to the frontend login page, consent page, or an error page.
func (h *OAuthHandler) AuthorizeBrowser(c *gin.Context) {
	if h.oauthInteraction == nil {
		c.Redirect(http.StatusFound, h.buildLocalErrorURL("temporarily_unavailable", "OAuth service is not available"))
		return
	}

	// Extract query parameters
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	responseType := c.Query("response_type")
	scope := c.Query("scope")
	state := c.Query("state")
	codeChallenge := c.Query("code_challenge")
	codeChallengeMethod := c.Query("code_challenge_method")
	nonce := c.Query("nonce")
	prompt := c.Query("prompt")
	maxAgeStr := c.Query("max_age")

	var maxAge *int
	maxAgeInvalid := false
	if maxAgeStr != "" {
		if v, err := strconv.Atoi(maxAgeStr); err == nil {
			maxAge = &v
		} else {
			maxAgeInvalid = true
		}
	}

	// Read SSO session cookie
	sessionCookie, _ := c.Cookie("keyles_sso")

	// Source IP: use direct TCP peer, not forwarded headers
	sourceIP := directPeerIP(c)

	req := auth.InitializeAuthRequest{
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		ResponseType:        responseType,
		Scope:               scope,
		State:               state,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		Nonce:               nonce,
		Prompt:              prompt,
		MaxAge:              maxAge,
		MaxAgeInvalid:       maxAgeInvalid,
		SessionCookie:       sessionCookie,
		SourceIP:            sourceIP,
	}

	resp, err := h.oauthInteraction.InitializeAuth(c.Request.Context(), req)
	if err != nil {
		// Infrastructure/Redis errors — redirect to frontend error page
		c.Redirect(http.StatusFound, h.buildLocalErrorURL("temporarily_unavailable", "An internal error occurred"))
		return
	}

	if resp.IsLocalError {
		// Local error: redirect to the URL returned (frontend error page)
		c.Redirect(http.StatusFound, resp.RedirectURL)
		return
	}

	// Normal redirect (302) to the frontend login/consent page or callback
	c.Redirect(http.StatusFound, resp.RedirectURL)
}

// loginRequest is the JSON body for POST /oauth2/login.
type loginRequest struct {
	TransactionID string `json:"transaction_id"`
	Email         string `json:"email"`
	Password      string `json:"password"`
}

// Login handles POST /oauth2/login for the browser-facing authentication flow.
func (h *OAuthHandler) Login(c *gin.Context) {
	if h.authenticateEndUser == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":             "temporarily_unavailable",
			"error_description": "login is not available",
		})
		return
	}

	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "failed to parse request body",
		})
		return
	}

	sourceIP := directPeerIP(c)

	result, err := h.authenticateEndUser.Execute(c.Request.Context(), auth.AuthenticateEndUserRequest{
		TransactionID: req.TransactionID,
		Email:         req.Email,
		Password:      req.Password,
		SourceIP:      sourceIP,
	})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrThrottled):
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":             "throttled",
				"error_description": "too many login attempts; please try again later",
			})
		case errors.Is(err, auth.ErrInvalidCredentials), errors.Is(err, auth.ErrTransactionExpired):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":             "invalid_credentials",
				"error_description": "invalid credentials or expired transaction",
			})
		default:
			// Redis or other infrastructure errors
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "temporarily_unavailable",
			})
		}
		return
	}

	// Set SSO session cookie
	sessionTTL := h.sessionTTL()
	httpOnly := true
	secure := h.config != nil && h.config.SecurityCookieSecure
	c.SetCookie(
		"keyles_sso",
		result.SessionID,
		sessionTTL,
		"/",
		"", // no Domain — host-only cookie
		secure,
		httpOnly,
	)
	// SameSite=Lax must be set explicitly; gin SetCookie uses SameSiteDefaultMode
	// which maps to Lax in modern browsers, but let's be explicit.
	if len(c.Writer.Header()["Set-Cookie"]) > 0 {
		cookies := c.Writer.Header()["Set-Cookie"]
		last := cookies[len(cookies)-1]
		last += "; SameSite=Lax"
		cookies[len(cookies)-1] = last
		c.Writer.Header()["Set-Cookie"] = cookies
	}

	c.JSON(http.StatusOK, gin.H{
		"redirect_url": result.RedirectURL,
	})
}

// ConsentDetail handles GET /oauth2/consent/:transactionId for the browser-facing flow.
// Returns the consent screen data (client name, scopes, user display name).
func (h *OAuthHandler) ConsentDetail(c *gin.Context) {
	if h.getConsentDetails == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":             "temporarily_unavailable",
			"error_description": "consent service is not available",
		})
		return
	}
	transactionID := c.Param("transactionId")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "missing transaction ID",
		})
		return
	}
	sessionID, _ := c.Cookie("keyles_sso")
	if sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "login_required",
			"error_description": "session not found; please sign in",
		})
		return
	}
	result, err := h.getConsentDetails.Execute(c.Request.Context(), transactionID, sessionID)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrSessionMissing):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":             "login_required",
				"error_description": "session not found or expired",
			})
		case errors.Is(err, auth.ErrTransactionExpired):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":             "invalid_request",
				"error_description": "authorization request expired or not found",
			})
		case errors.Is(err, auth.ErrTransactionWrongStage):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":             "invalid_request",
				"error_description": "authorization request is not awaiting consent",
			})
		case errors.Is(err, auth.ErrSessionUserMismatch):
			c.JSON(http.StatusForbidden, gin.H{
				"error":             "access_denied",
				"error_description": "session user does not match the authorization request",
			})
		case errors.Is(err, auth.ErrConsentDenied):
			c.JSON(http.StatusForbidden, gin.H{
				"error":             "access_denied",
				"error_description": "user is disabled or has no role for this client",
			})
		case errors.Is(err, auth.ErrTemporarilyUnavailable):
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":             "temporarily_unavailable",
				"error_description": "authorization interaction is temporarily unavailable",
			})
		default:
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "temporarily_unavailable",
			})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"transaction_id":         result.TransactionID,
		"interaction_csrf_token": result.InteractionCSRFToken,
		"client_id":              result.ClientID,
		"client_name":            result.ClientName,
		"client_logo_uri":        result.ClientLogoURI,
		"scopes":                 result.Scopes,
		"user_display":           result.UserDisplay,
	})
}

// consentDecisionRequest is the JSON body for POST /oauth2/consent.
type consentDecisionRequest struct {
	TransactionID        string `json:"transaction_id"`
	InteractionCSRFToken string `json:"interaction_csrf_token"`
	Approved             bool   `json:"approved"`
}

// ConsentDecision handles POST /oauth2/consent for the browser-facing flow.
// The user approves or denies the authorization request.
func (h *OAuthHandler) ConsentDecision(c *gin.Context) {
	if h.consentDecision == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":             "temporarily_unavailable",
			"error_description": "consent service is not available",
		})
		return
	}

	var req consentDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "failed to parse request body",
		})
		return
	}

	// Read SSO session cookie to identify the user
	sessionID, _ := c.Cookie("keyles_sso")
	if sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "login_required",
			"error_description": "you need to sign in to continue",
		})
		return
	}

	result, err := h.consentDecision.Execute(c.Request.Context(), auth.ConsentDecisionRequest{
		TransactionID:        req.TransactionID,
		InteractionCSRFToken: req.InteractionCSRFToken,
		Approved:             req.Approved,
		SessionID:            sessionID,
	})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrTransactionExpired):
			c.JSON(http.StatusGone, gin.H{
				"error":             "invalid_request",
				"error_description": "authorization request expired or already consumed",
			})
		case errors.Is(err, auth.ErrInvalidCSRFToken):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":             "invalid_request",
				"error_description": "invalid or expired consent session",
			})
		case errors.Is(err, auth.ErrSessionUserMismatch):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":             "login_required",
				"error_description": "session mismatch; please sign in again",
			})
		default:
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":             "temporarily_unavailable",
				"error_description": "authorization interaction is temporarily unavailable",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"redirect_url": result.RedirectURL})
}

// Logout handles POST /oauth2/logout.
// Terminates the SSO session by deleting it from Redis, emits an audit event,
// and sets an expired cookie. Always succeeds — if no session exists, the
// response is still 204.
func (h *OAuthHandler) Logout(c *gin.Context) {
	sessionID, _ := c.Cookie("keyles_sso")
	sourceIP := directPeerIP(c)

	// Delete session from Redis (best-effort; ignore errors)
	if h.logoutEndUser != nil && sessionID != "" {
		_ = h.logoutEndUser.Execute(c.Request.Context(), sessionID, sourceIP)
	}

	// Always set expired cookie to clear the client
	secure := h.config != nil && h.config.SecurityCookieSecure
	c.SetCookie(
		"keyles_sso",
		"",
		-1,
		"/",
		"",
		secure,
		true, // HttpOnly
	)
	// Enforce SameSite=Lax explicitly
	if len(c.Writer.Header()["Set-Cookie"]) > 0 {
		cookies := c.Writer.Header()["Set-Cookie"]
		last := cookies[len(cookies)-1]
		last += "; SameSite=Lax"
		cookies[len(cookies)-1] = last
		c.Writer.Header()["Set-Cookie"] = cookies
	}

	c.Status(http.StatusNoContent)
}

// directPeerIP returns the TCP peer IP without trusting forwarded headers.
func directPeerIP(c *gin.Context) string {
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err == nil {
		return host
	}
	return c.Request.RemoteAddr
}

// buildLocalErrorURL constructs a redirect URL to the frontend error page
// with the given error code and description. Used for local errors where
// no valid redirect_uri is available.
func (h *OAuthHandler) buildLocalErrorURL(errorCode, description string) string {
	return fmt.Sprintf("%s/oauth2/error?error=%s&error_description=%s",
		h.frontendURL(),
		url.QueryEscape(errorCode),
		url.QueryEscape(description),
	)
}

// buildCallbackErrorURL constructs a redirect URL back to the client's
// redirect_uri with error parameters and the original state parameter.
// Used when the redirect_uri has been validated and the error must be
// communicated to the relying party per RFC 6749 §4.1.2.1.
func (h *OAuthHandler) buildCallbackErrorURL(redirectURI, state, errorCode, description string) string {
	u, err := url.Parse(redirectURI)
	if err != nil {
		// Malformed redirect URI — fall back to local error page
		return h.buildLocalErrorURL(errorCode, description)
	}
	q := u.Query()
	q.Set("error", errorCode)
	q.Set("error_description", description)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// frontendURL returns the configured frontend URL or a sensible default.
func (h *OAuthHandler) frontendURL() string {
	if h.config != nil && h.config.FrontendURL != "" {
		return h.config.FrontendURL
	}
	return "http://localhost:3000"
}

// sessionTTL returns the configured session TTL in seconds.
func (h *OAuthHandler) sessionTTL() int {
	if h.config != nil && h.config.SecuritySessionTTL > 0 {
		return h.config.SecuritySessionTTL
	}
	return 28800 // 8 hours default
}
