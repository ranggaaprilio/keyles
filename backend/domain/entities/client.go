package entities

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Client type constants
const (
	ClientTypeConfidential = "confidential"
	ClientTypePublic       = "public"
)

// MaxClientsPerTenant defines the maximum number of OAuth clients per tenant
const MaxClientsPerTenant = 25

// Client represents an OAuth 2.0 / OIDC client application
type Client struct {
	ClientID            string
	TenantID            string
	ClientName          string
	Description         string // Application description (optional, max 500 chars)
	ClientType          string // "confidential" or "public"
	ClientSecretHash    string // Empty for public clients
	AllowedRedirectURIs []string
	IsActive            bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// Validate performs basic validation on the client entity
func (c *Client) Validate() error {
	if c.ClientID == "" {
		return errors.New("client_id cannot be empty")
	}
	if c.TenantID == "" {
		return errors.New("tenant_id cannot be empty")
	}
	if c.ClientName == "" {
		return errors.New("client_name cannot be empty")
	}
	if len(c.ClientName) < 3 || len(c.ClientName) > 100 {
		return errors.New("client_name must be between 3 and 100 characters")
	}
	if len(c.Description) > 500 {
		return errors.New("description must not exceed 500 characters")
	}

	// Validate client type
	if c.ClientType != ClientTypeConfidential && c.ClientType != ClientTypePublic {
		return fmt.Errorf("invalid client_type: must be '%s' or '%s'", ClientTypeConfidential, ClientTypePublic)
	}

	// Confidential clients must have a secret hash
	if c.ClientType == ClientTypeConfidential && c.ClientSecretHash == "" {
		return errors.New("client_secret_hash cannot be empty for confidential clients")
	}

	if len(c.AllowedRedirectURIs) == 0 {
		return errors.New("at least one redirect URI is required")
	}

	// Validate each redirect URI
	for _, uri := range c.AllowedRedirectURIs {
		if err := c.ValidateRedirectURI(uri); err != nil {
			return err
		}
	}

	return nil
}

// ValidateRedirectURI validates a single redirect URI format
func (c *Client) ValidateRedirectURI(redirectURI string) error {
	if redirectURI == "" {
		return errors.New("redirect_uri cannot be empty")
	}

	// Parse the URI
	parsedURL, err := url.Parse(redirectURI)
	if err != nil {
		return errors.New("invalid redirect_uri format")
	}

	// Must have a scheme (http or https)
	if parsedURL.Scheme == "" {
		return errors.New("redirect_uri must have a scheme (http or https)")
	}

	// Must have a host
	if parsedURL.Host == "" {
		return errors.New("redirect_uri must have a host")
	}

	// No fragments allowed per OAuth 2.0 spec
	if parsedURL.Fragment != "" {
		return errors.New("redirect_uri must not contain a fragment")
	}

	return nil
}

// ValidateRedirectURIStrict validates a redirect URI with HTTPS enforcement
// allowInsecureLocalhost: if true, allows http:// for localhost/127.0.0.1 (development)
func (c *Client) ValidateRedirectURIStrict(uri string, allowInsecureLocalhost bool) error {
	// First, run basic validation
	if err := c.ValidateRedirectURI(uri); err != nil {
		return err
	}

	parsedURL, _ := url.Parse(uri) // safe: already validated above

	// Check for HTTPS requirement
	if parsedURL.Scheme == "https" {
		return nil
	}

	// Allow HTTP only for localhost if permitted
	if allowInsecureLocalhost && parsedURL.Scheme == "http" {
		host := strings.Split(parsedURL.Host, ":")[0] // strip port
		if host == "localhost" || host == "127.0.0.1" {
			return nil
		}
	}

	return errors.New("redirect_uri must use HTTPS (except localhost)")
}

// IsURIAllowed checks if a given redirect URI is in the allowed list (exact match)
func (c *Client) IsURIAllowed(redirectURI string) bool {
	for _, allowedURI := range c.AllowedRedirectURIs {
		if strings.EqualFold(allowedURI, redirectURI) {
			return true
		}
	}
	return false
}

// IsEnabled checks if the client is active
func (c *Client) IsEnabled() bool {
	return c.IsActive
}
