package integration

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/tests/mocks"
	"github.com/ranggaaprilio/keyles/usecase/auth"
)

// setupRevokeTestRouter creates a test router with all necessary use cases
func setupRevokeTestRouter(
	mockRefreshTokenRepo *mocks.MockRefreshTokenRepository,
) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create RevokeToken use case
	revokeTokenUC := auth.NewRevokeToken(mockRefreshTokenRepo)

	// Create handler with revoke use case
	handler := handlers.NewOAuthHandlerWithRevoke(nil, nil, nil, revokeTokenUC)

	// Register routes
	router.POST("/oauth2/revoke", handler.Revoke)

	return router
}

// TestOAuthRevoke_Success tests successful token revocation
func TestOAuthRevoke_Success(t *testing.T) {
	// Arrange
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	validToken := &entities.RefreshToken{
		ID:          1,
		Token:       "hashed-token",
		UserID:      "user-123",
		ClientID:    "test-client",
		TenantID:    "tenant-123",
		RevokedFlag: false,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	mockRefreshTokenRepo.On("GetByToken", mock.Anything, mock.AnythingOfType("string")).Return(validToken, nil)
	mockRefreshTokenRepo.On("Revoke", mock.Anything, mock.AnythingOfType("string"), "client_request").Return(nil)

	router := setupRevokeTestRouter(mockRefreshTokenRepo)

	// Act
	form := url.Values{}
	form.Set("token", "test-refresh-token")
	form.Set("token_type_hint", "refresh_token")

	req := httptest.NewRequest("POST", "/oauth2/revoke", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("test-client:test-secret")))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - RFC 7009 requires HTTP 200 OK for successful revocation
	assert.Equal(t, http.StatusOK, w.Code)
	mockRefreshTokenRepo.AssertExpectations(t)
}

// TestOAuthRevoke_InvalidToken tests RFC 7009 compliant behavior for invalid tokens
func TestOAuthRevoke_InvalidToken(t *testing.T) {
	// Arrange
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	mockRefreshTokenRepo.On("GetByToken", mock.Anything, mock.AnythingOfType("string")).Return(nil, errors.New("not found"))

	router := setupRevokeTestRouter(mockRefreshTokenRepo)

	// Act
	form := url.Values{}
	form.Set("token", "invalid-token")

	req := httptest.NewRequest("POST", "/oauth2/revoke", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("test-client:test-secret")))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - RFC 7009: server responds with 200 OK even for invalid tokens
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestOAuthRevoke_MissingToken tests error handling for missing token parameter
func TestOAuthRevoke_MissingToken(t *testing.T) {
	// Arrange
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	router := setupRevokeTestRouter(mockRefreshTokenRepo)

	// Act
	form := url.Values{}
	// No token provided

	req := httptest.NewRequest("POST", "/oauth2/revoke", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("test-client:test-secret")))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestOAuthRevoke_AlreadyRevokedToken tests idempotent revocation
func TestOAuthRevoke_AlreadyRevokedToken(t *testing.T) {
	// Arrange
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	revokedToken := &entities.RefreshToken{
		ID:          1,
		Token:       "hashed-token",
		UserID:      "user-123",
		ClientID:    "test-client",
		TenantID:    "tenant-123",
		RevokedFlag: true,
		RevokedAt:   func() *time.Time { t := time.Now(); return &t }(),
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	mockRefreshTokenRepo.On("GetByToken", mock.Anything, mock.AnythingOfType("string")).Return(revokedToken, nil)

	router := setupRevokeTestRouter(mockRefreshTokenRepo)

	// Act
	form := url.Values{}
	form.Set("token", "already-revoked-token")

	req := httptest.NewRequest("POST", "/oauth2/revoke", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("test-client:test-secret")))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - should succeed (idempotent)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestOAuthRevoke_AccessTokenHint tests handling of access_token hint
func TestOAuthRevoke_AccessTokenHint(t *testing.T) {
	// Arrange
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	router := setupRevokeTestRouter(mockRefreshTokenRepo)

	// Act
	form := url.Values{}
	form.Set("token", "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...")
	form.Set("token_type_hint", "access_token")

	req := httptest.NewRequest("POST", "/oauth2/revoke", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("test-client:test-secret")))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - access tokens are JWTs, can't be revoked but return 200 OK
	assert.Equal(t, http.StatusOK, w.Code)
	mockRefreshTokenRepo.AssertNotCalled(t, "GetByToken")
}

// TestOAuthRevoke_NoHint tests default behavior without hint
func TestOAuthRevoke_NoHint(t *testing.T) {
	// Arrange
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)

	validToken := &entities.RefreshToken{
		ID:          1,
		Token:       "hashed-token",
		UserID:      "user-123",
		ClientID:    "test-client",
		TenantID:    "tenant-123",
		RevokedFlag: false,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	mockRefreshTokenRepo.On("GetByToken", mock.Anything, mock.AnythingOfType("string")).Return(validToken, nil)
	mockRefreshTokenRepo.On("Revoke", mock.Anything, mock.AnythingOfType("string"), "client_request").Return(nil)

	router := setupRevokeTestRouter(mockRefreshTokenRepo)

	// Act
	form := url.Values{}
	form.Set("token", "test-refresh-token")
	// No hint provided - should default to refresh_token

	req := httptest.NewRequest("POST", "/oauth2/revoke", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("test-client:test-secret")))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockRefreshTokenRepo.AssertExpectations(t)
}
