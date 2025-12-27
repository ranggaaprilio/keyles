package integration

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/domain/services"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTokenServiceForDiscovery implements TokenService for discovery tests
type MockTokenServiceForDiscovery struct {
	privateKey *rsa.PrivateKey
	keyID      string
}

func NewMockTokenServiceForDiscovery() *MockTokenServiceForDiscovery {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	return &MockTokenServiceForDiscovery{
		privateKey: privateKey,
		keyID:      "test-key-discovery-1",
	}
}

func (m *MockTokenServiceForDiscovery) SignIDToken(ctx context.Context, claims *services.TokenClaims) (string, error) {
	return "mock-id-token", nil
}

func (m *MockTokenServiceForDiscovery) SignAccessToken(ctx context.Context, claims *services.TokenClaims) (string, error) {
	return "mock-access-token", nil
}

func (m *MockTokenServiceForDiscovery) ValidateTokenSignature(ctx context.Context, token string) (*services.TokenClaims, error) {
	return nil, nil
}

func (m *MockTokenServiceForDiscovery) GetPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	return &m.privateKey.PublicKey, nil
}

func (m *MockTokenServiceForDiscovery) GetJWKS(ctx context.Context) (*services.JWKS, error) {
	publicKey := &m.privateKey.PublicKey
	n := base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())
	eBytes := make([]byte, 3)
	eBytes[0] = byte((publicKey.E >> 16) & 0xff)
	eBytes[1] = byte((publicKey.E >> 8) & 0xff)
	eBytes[2] = byte(publicKey.E & 0xff)
	e := base64.RawURLEncoding.EncodeToString(eBytes)

	return &services.JWKS{
		Keys: []services.JWK{
			{
				KeyType:   "RSA",
				Use:       "sig",
				KeyID:     m.keyID,
				Algorithm: "RS256",
				N:         n,
				E:         e,
			},
		},
	}, nil
}

func (m *MockTokenServiceForDiscovery) GetActiveKeyID(ctx context.Context) (string, error) {
	return m.keyID, nil
}

var _ services.TokenService = (*MockTokenServiceForDiscovery)(nil)

func setupDiscoveryRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	tokenService := NewMockTokenServiceForDiscovery()
	discoveryHandler := handlers.NewDiscoveryHandler(tokenService, "https://sso.test.com")

	router.GET("/.well-known/openid-configuration", discoveryHandler.OpenIDConfiguration)
	router.GET("/.well-known/jwks.json", discoveryHandler.JWKS)

	return router
}

func TestDiscovery_OpenIDConfiguration(t *testing.T) {
	router := setupDiscoveryRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/openid-configuration", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var config handlers.OIDCConfiguration
	err := json.Unmarshal(w.Body.Bytes(), &config)
	require.NoError(t, err)

	assert.Equal(t, "https://sso.test.com", config.Issuer)
	assert.Equal(t, "https://sso.test.com/oauth2/auth", config.AuthorizationEndpoint)
	assert.Equal(t, "https://sso.test.com/oauth2/token", config.TokenEndpoint)
	assert.Equal(t, "https://sso.test.com/.well-known/jwks.json", config.JwksURI)
	assert.Contains(t, config.ScopesSupported, "openid")
	assert.Contains(t, config.ResponseTypesSupported, "code")
	assert.Contains(t, config.GrantTypesSupported, "authorization_code")
	assert.Contains(t, config.CodeChallengeMethodsSupported, "S256")
	assert.Contains(t, config.IDTokenSigningAlgValuesSupported, "RS256")
}

func TestDiscovery_JWKS(t *testing.T) {
	router := setupDiscoveryRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var jwks services.JWKS
	err := json.Unmarshal(w.Body.Bytes(), &jwks)
	require.NoError(t, err)

	require.NotEmpty(t, jwks.Keys)
	key := jwks.Keys[0]

	assert.Equal(t, "RSA", key.KeyType)
	assert.Equal(t, "sig", key.Use)
	assert.Equal(t, "RS256", key.Algorithm)
	assert.NotEmpty(t, key.KeyID)
	assert.NotEmpty(t, key.N)
	assert.NotEmpty(t, key.E)
}
