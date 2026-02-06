package handlers

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/usecase/auth"
)

// OAuthHandler handles OAuth 2.0 / OIDC endpoints
type OAuthHandler struct {
	authorizeClientUC *auth.AuthorizeClient
	issueTokenUC      *auth.IssueToken
	refreshTokenUC    *auth.RefreshToken
	revokeTokenUC     *auth.RevokeToken
	clientRepo        repositories.ClientRepository
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
) *OAuthHandler {
	return &OAuthHandler{
		authorizeClientUC: authorizeClientUC,
		issueTokenUC:      issueTokenUC,
		refreshTokenUC:    refreshTokenUC,
		revokeTokenUC:     revokeTokenUC,
		clientRepo:        clientRepo,
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
	tenantID := c.GetHeader("X-Tenant-ID")

	// If user is not authenticated, we would redirect to login page
	// For now, we require the headers to be set
	if userID == "" || tenantID == "" {
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
		TenantID:            tenantID,
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
