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
// Per RFC 6749 Section 4.1.3 (Authorization Code Exchange)
func (h *OAuthHandler) Token(c *gin.Context) {
	// Extract form parameters (application/x-www-form-urlencoded per RFC 6749)
	grantType := c.PostForm("grant_type")
	code := c.PostForm("code")
	redirectURI := c.PostForm("redirect_uri")
	clientID := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")
	codeVerifier := c.PostForm("code_verifier")

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

	// Build token request
	req := auth.TokenRequest{
		GrantType:    grantType,
		Code:         code,
		RedirectURI:  redirectURI,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CodeVerifier: codeVerifier,
	}

	// Execute token exchange
	resp, err := h.issueTokenUC.Execute(c.Request.Context(), req)
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

	// Return token response per RFC 6749 Section 5.1
	c.JSON(http.StatusOK, resp)
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
