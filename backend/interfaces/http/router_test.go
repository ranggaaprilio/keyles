package http

import (
	"testing"

	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
)

func TestRouterSetupDoesNotPanic(t *testing.T) {
	router := NewRouter(
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
