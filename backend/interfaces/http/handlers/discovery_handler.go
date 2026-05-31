package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/domain/services"
)

// OIDCConfiguration represents the OpenID Connect Discovery metadata
// per RFC 8414 and OpenID Connect Discovery 1.0 (FR-041)
type OIDCConfiguration struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserInfoEndpoint                  string   `json:"userinfo_endpoint"`
	JwksURI                           string   `json:"jwks_uri"`
	RegistrationEndpoint              string   `json:"registration_endpoint,omitempty"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	ResponseModesSupported            []string `json:"response_modes_supported,omitempty"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported,omitempty"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	RevocationEndpoint                string   `json:"revocation_endpoint,omitempty"`
	IntrospectionEndpoint             string   `json:"introspection_endpoint,omitempty"`
}

// DiscoveryHandler handles OIDC discovery endpoints
type DiscoveryHandler struct {
	tokenService services.TokenService
	issuer       string
}

// NewDiscoveryHandler creates a new DiscoveryHandler
func NewDiscoveryHandler(tokenService services.TokenService, issuer string) *DiscoveryHandler {
	return &DiscoveryHandler{
		tokenService: tokenService,
		issuer:       issuer,
	}
}

// OpenIDConfiguration returns the OIDC discovery metadata (FR-041)
// GET /.well-known/openid-configuration
func (h *DiscoveryHandler) OpenIDConfiguration(c *gin.Context) {
	config := OIDCConfiguration{
		Issuer:                h.issuer,
		AuthorizationEndpoint: h.issuer + "/oauth2/auth",
		TokenEndpoint:         h.issuer + "/oauth2/token",
		UserInfoEndpoint:      h.issuer + "/oauth2/userinfo",
		JwksURI:               h.issuer + "/.well-known/jwks.json",
		RevocationEndpoint:    h.issuer + "/oauth2/revoke",
		IntrospectionEndpoint: h.issuer + "/oauth2/introspect",

		// Supported scopes
		ScopesSupported: []string{
			"openid",
			"profile",
			"email",
			"offline_access",
		},

		// Response types supported (authorization code flow only per spec)
		ResponseTypesSupported: []string{
			"code",
		},

		// Response modes
		ResponseModesSupported: []string{
			"query",
		},

		// Grant types supported
		GrantTypesSupported: []string{
			"authorization_code",
			"refresh_token",
		},

		// Subject types
		SubjectTypesSupported: []string{
			"public",
		},

		// Token signing algorithms (RS256 only per FR-036)
		IDTokenSigningAlgValuesSupported: []string{
			"RS256",
		},

		// Token endpoint auth methods
		TokenEndpointAuthMethodsSupported: []string{
			"client_secret_basic",
			"client_secret_post",
			"none",
		},

		// PKCE - S256 only (FR-008)
		CodeChallengeMethodsSupported: []string{
			"S256",
		},

		// Supported claims
		ClaimsSupported: []string{
			"iss",
			"sub",
			"aud",
			"exp",
			"iat",
			"nbf",
			"jti",
			"email",
			"email_verified",
			"name",
			"given_name",
			"family_name",
			"tenant_id",
			"client_id",
			"scope",
			"roles",
		},
	}

	c.JSON(http.StatusOK, config)
}

// JWKS returns the JSON Web Key Set (FR-039, FR-040)
// GET /.well-known/jwks.json
func (h *DiscoveryHandler) JWKS(c *gin.Context) {
	jwks, err := h.tokenService.GetJWKS(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "Failed to retrieve public keys",
		})
		return
	}

	// Return JWKS in the standard format
	// Format: {"keys": [{"kty": "RSA", "kid": "...", "use": "sig", "alg": "RS256", "n": "...", "e": "..."}]}
	c.JSON(http.StatusOK, jwks)
}
