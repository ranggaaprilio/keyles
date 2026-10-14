package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ranggaaprilio/keyles/infrastructure/config"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
)

func TestRouterSetupDoesNotPanic(t *testing.T) {
	router := NewRouter(
		&config.Config{},
		nil,
		nil,
		&handlers.RegistrationHandler{},
		&handlers.AvailabilityHandler{},
		&handlers.VerificationHandler{},
		&handlers.ResendOTPHandler{},
		&handlers.AuthHandler{},
		&handlers.DashboardHandler{},
		&handlers.HealthHandler{},
		&handlers.ClientHandler{},
		&handlers.OAuthHandler{},
		&handlers.OAuthSignupHandler{},
		&handlers.DiscoveryHandler{},
		&handlers.RoleHandler{},
		&handlers.SessionHandler{},
		&handlers.UserHandler{},
		&handlers.InvitationHandler{},
		&handlers.ActivityHandler{},
		&handlers.UserinfoHandler{},
		nil,
		nil,
		nil,
		"*",
		"GET,POST,PUT,DELETE,OPTIONS",
		"Origin,Content-Type,Accept,Authorization",
	)

	router.Setup()
}

// Regression: the placeholder header-only authorization endpoint must stay
// removed. It issued authorization codes for any user named in X-User-ID.
func TestRouterDoesNotExposeHeaderOnlyAuthorize(t *testing.T) {
	router := NewRouter(
		&config.Config{},
		nil,
		nil,
		&handlers.RegistrationHandler{},
		&handlers.AvailabilityHandler{},
		&handlers.VerificationHandler{},
		&handlers.ResendOTPHandler{},
		&handlers.AuthHandler{},
		&handlers.DashboardHandler{},
		&handlers.HealthHandler{},
		&handlers.ClientHandler{},
		&handlers.OAuthHandler{},
		&handlers.OAuthSignupHandler{},
		&handlers.DiscoveryHandler{},
		&handlers.RoleHandler{},
		&handlers.SessionHandler{},
		&handlers.UserHandler{},
		&handlers.InvitationHandler{},
		&handlers.ActivityHandler{},
		&handlers.UserinfoHandler{},
		nil,
		nil,
		nil,
		"*",
		"GET,POST,PUT,DELETE,OPTIONS",
		"Origin,Content-Type,Accept,Authorization",
	)
	router.Setup()

	req := httptest.NewRequest(http.MethodPost, "/oauth2/auth?client_id=c&redirect_uri=https://rp.example/cb&response_type=code&scope=openid&state=s&code_challenge=x&code_challenge_method=S256", nil)
	req.Header.Set("X-User-ID", "victim-user-id")
	rec := httptest.NewRecorder()
	router.engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for POST /oauth2/auth, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "" {
		t.Fatalf("expected no redirect, got Location=%q", loc)
	}
}
