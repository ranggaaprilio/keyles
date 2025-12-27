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
	clientRepo        repositories.ClientRepository
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(
	authorizeClientUC *auth.AuthorizeClient,
	clientRepo repositories.ClientRepository,
) *OAuthHandler {
	return &OAuthHandler{
		authorizeClientUC: authorizeClientUC,
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
		return http.StatusBadRequest
	case auth.ErrInvalidGrant:
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}
