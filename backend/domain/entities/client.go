package entities

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

// Client represents an OAuth 2.0 / OIDC client application
type Client struct {
	ClientID            string
	TenantID            string
	ClientName          string
	ClientSecretHash    string
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
	if c.ClientSecretHash == "" {
		return errors.New("client_secret_hash cannot be empty")
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
