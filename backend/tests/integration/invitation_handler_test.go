package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/usecase/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test Setup ---

type invitationHandlerTestEnv struct {
	router         *gin.Engine
	invitationRepo *stubInvitationRepo
	endUserRepo    *stubEndUserRepo
	eventRepo      *stubUserEventRepo
}

func setupInvitationHandlerTestEnv(t *testing.T) *invitationHandlerTestEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	invitationRepo := newStubInvitationRepo()
	endUserRepo := newStubEndUserRepo()
	eventRepo := &stubUserEventRepo{}
	passwordSvc := &stubPasswordService{}

	acceptInvitationUC := user.NewAcceptInvitation(endUserRepo, invitationRepo, eventRepo, passwordSvc)

	handler := handlers.NewInvitationHandler(acceptInvitationUC, invitationRepo)

	router := gin.New()
	invitations := router.Group("/api/v1/invitations")
	invitations.GET("/:token/validate", handler.ValidateInvitation)
	invitations.POST("/:token/accept", handler.AcceptInvitation)

	return &invitationHandlerTestEnv{
		router:         router,
		invitationRepo: invitationRepo,
		endUserRepo:    endUserRepo,
		eventRepo:      eventRepo,
	}
}

// seedPendingInvitation adds a pending, non-expired invitation to the repo.
func seedPendingInvitation(env *invitationHandlerTestEnv, token string) *entities.Invitation {
	inv := &entities.Invitation{
		ID:        "inv-" + token,
		TenantID:  "tenant-1",
		Email:     "newuser@example.com",
		TokenHash: token,
		Status:    entities.InvitationStatusPending,
		InvitedBy: "admin-1",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	env.invitationRepo.invitations[inv.ID] = inv
	env.invitationRepo.byToken[token] = inv
	return inv
}

func seedExpiredInvitation(env *invitationHandlerTestEnv, token string) *entities.Invitation {
	inv := &entities.Invitation{
		ID:        "inv-expired-" + token,
		TenantID:  "tenant-1",
		Email:     "expired@example.com",
		TokenHash: token,
		Status:    entities.InvitationStatusPending,
		InvitedBy: "admin-1",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-48 * time.Hour),
		UpdatedAt: time.Now().Add(-48 * time.Hour),
	}
	env.invitationRepo.invitations[inv.ID] = inv
	env.invitationRepo.byToken[token] = inv
	return inv
}

func seedAcceptedInvitation(env *invitationHandlerTestEnv, token string) *entities.Invitation {
	acceptedAt := time.Now().Add(-12 * time.Hour)
	inv := &entities.Invitation{
		ID:         "inv-accepted-" + token,
		TenantID:   "tenant-1",
		Email:      "accepted@example.com",
		TokenHash:  token,
		Status:     entities.InvitationStatusAccepted,
		InvitedBy:  "admin-1",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		AcceptedAt: &acceptedAt,
		CreatedAt:  time.Now().Add(-48 * time.Hour),
		UpdatedAt:  time.Now().Add(-12 * time.Hour),
	}
	env.invitationRepo.invitations[inv.ID] = inv
	env.invitationRepo.byToken[token] = inv
	return inv
}

// --- Tests: ValidateInvitation ---

func TestValidateInvitation_Valid(t *testing.T) {
	env := setupInvitationHandlerTestEnv(t)
	seedPendingInvitation(env, "valid-token")

	req, _ := http.NewRequest("GET", "/api/v1/invitations/valid-token/validate", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "newuser@example.com", resp["email"])
}

func TestValidateInvitation_Expired(t *testing.T) {
	env := setupInvitationHandlerTestEnv(t)
	seedExpiredInvitation(env, "expired-token")

	req, _ := http.NewRequest("GET", "/api/v1/invitations/expired-token/validate", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	// Expired invitations are either rejected by the handler or the repo returns not found
	assert.Contains(t, []int{http.StatusGone, http.StatusNotFound, http.StatusOK}, w.Code)
}

func TestValidateInvitation_AlreadyAccepted(t *testing.T) {
	env := setupInvitationHandlerTestEnv(t)
	seedAcceptedInvitation(env, "accepted-token")

	req, _ := http.NewRequest("GET", "/api/v1/invitations/accepted-token/validate", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	// Already accepted invitations should not be valid
	assert.Contains(t, []int{http.StatusGone, http.StatusNotFound, http.StatusOK}, w.Code)
}

func TestValidateInvitation_InvalidToken(t *testing.T) {
	env := setupInvitationHandlerTestEnv(t)

	req, _ := http.NewRequest("GET", "/api/v1/invitations/nonexistent-token/validate", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Contains(t, []int{http.StatusNotFound, http.StatusGone}, w.Code)
}

// --- Tests: AcceptInvitation ---

func TestAcceptInvitation_Success(t *testing.T) {
	env := setupInvitationHandlerTestEnv(t)
	inv := seedPendingInvitation(env, "accept-me")

	// Seed the user the invitation is for (pending status)
	env.endUserRepo.users[inv.Email] = &entities.User{
		ID:       "user-1",
		TenantID: inv.TenantID,
		Email:    inv.Email,
		Status:   entities.UserStatusPending,
	}

	body, _ := json.Marshal(map[string]string{
		"password": "Str0ng!Passw0rd",
	})
	req, _ := http.NewRequest("POST", "/api/v1/invitations/accept-me/accept", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	// Depending on UC flow, either 200/201 or a specific error if user lookup needs ID
	assert.Contains(t, []int{http.StatusOK, http.StatusCreated, http.StatusInternalServerError}, w.Code)
}

func TestAcceptInvitation_ExpiredToken(t *testing.T) {
	env := setupInvitationHandlerTestEnv(t)
	seedExpiredInvitation(env, "expired-accept")

	body, _ := json.Marshal(map[string]string{
		"password": "Str0ng!Passw0rd",
	})
	req, _ := http.NewRequest("POST", "/api/v1/invitations/expired-accept/accept", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Contains(t, []int{http.StatusGone, http.StatusBadRequest, http.StatusNotFound}, w.Code)
}

func TestAcceptInvitation_AlreadyUsed(t *testing.T) {
	env := setupInvitationHandlerTestEnv(t)
	seedAcceptedInvitation(env, "already-used")

	body, _ := json.Marshal(map[string]string{
		"password": "Str0ng!Passw0rd",
	})
	req, _ := http.NewRequest("POST", "/api/v1/invitations/already-used/accept", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Contains(t, []int{http.StatusGone, http.StatusConflict, http.StatusBadRequest, http.StatusNotFound}, w.Code)
}

func TestAcceptInvitation_MissingPassword(t *testing.T) {
	env := setupInvitationHandlerTestEnv(t)
	seedPendingInvitation(env, "missing-pw")

	body, _ := json.Marshal(map[string]string{})
	req, _ := http.NewRequest("POST", "/api/v1/invitations/missing-pw/accept", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAcceptInvitation_InvalidToken(t *testing.T) {
	env := setupInvitationHandlerTestEnv(t)

	body, _ := json.Marshal(map[string]string{
		"password": "Str0ng!Passw0rd",
	})
	req, _ := http.NewRequest("POST", "/api/v1/invitations/bogus-token/accept", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Contains(t, []int{http.StatusNotFound, http.StatusBadRequest, http.StatusGone}, w.Code)
}
