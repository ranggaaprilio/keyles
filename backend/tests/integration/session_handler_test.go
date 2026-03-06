package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/usecase/auth"
	"github.com/ranggaaprilio/keyles/usecase/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test Setup ---

type sessionHandlerTestEnv struct {
	router           *gin.Engine
	endUserRepo      *stubEndUserRepo
	refreshTokenRepo *stubRefreshTokenRepo
	eventRepo        *stubUserEventRepo
	tenantID         string
	adminID          string
}

func setupSessionHandlerTestEnv(t *testing.T) *sessionHandlerTestEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	tenantID := uuid.New().String()
	adminID := uuid.New().String()

	endUserRepo := newStubEndUserRepo()
	refreshTokenRepo := &stubRefreshTokenRepo{}
	eventRepo := &stubUserEventRepo{}

	revokeTokenUC := auth.NewRevokeToken(refreshTokenRepo)
	listSessionsUC := user.NewListSessions(endUserRepo, refreshTokenRepo)
	revokeSessionUC := user.NewRevokeSession(endUserRepo, refreshTokenRepo, eventRepo)

	handler := handlers.NewSessionHandler(revokeTokenUC, listSessionsUC, revokeSessionUC)

	router := gin.New()
	admin := router.Group("/api/v1/admin", adminMiddleware(tenantID, adminID))
	users := admin.Group("/users")
	users.GET("/:id/sessions", handler.ListUserSessions)
	users.DELETE("/:id/sessions/:sessionId", handler.RevokeUserSession)

	return &sessionHandlerTestEnv{
		router:           router,
		endUserRepo:      endUserRepo,
		refreshTokenRepo: refreshTokenRepo,
		eventRepo:        eventRepo,
		tenantID:         tenantID,
		adminID:          adminID,
	}
}

// seedUserSession registers a user and 0+ sessions in the test stubs.
func seedUserSession(env *sessionHandlerTestEnv, userID string) {
	env.endUserRepo.users[userID] = &entities.User{
		ID:       userID,
		TenantID: env.tenantID,
		Email:    userID + "@example.com",
		Status:   entities.UserStatusActive,
	}
}

func addSession(env *sessionHandlerTestEnv, userID, clientID string, id int64, valid bool) {
	now := time.Now()
	token := &entities.RefreshToken{
		ID:        id,
		Token:     "token-" + userID,
		UserID:    userID,
		ClientID:  clientID,
		TenantID:  env.tenantID,
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
	}
	if !valid {
		token.RevokedFlag = true
		token.RevokedAt = &now
		token.RevokedReason = "test"
	}
	env.refreshTokenRepo.tokens = append(env.refreshTokenRepo.tokens, token)
}

// --- Tests: ListUserSessions ---

func TestListUserSessions_Success(t *testing.T) {
	env := setupSessionHandlerTestEnv(t)
	userID := uuid.New().String()
	clientID := uuid.New().String()

	seedUserSession(env, userID)
	addSession(env, userID, clientID, 1, true)
	addSession(env, userID, clientID, 2, true)

	req, _ := http.NewRequest("GET", "/api/v1/admin/users/"+userID+"/sessions", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	sessions := resp["sessions"].([]interface{})
	assert.Len(t, sessions, 2)
}

func TestListUserSessions_Empty(t *testing.T) {
	env := setupSessionHandlerTestEnv(t)
	userID := uuid.New().String()
	seedUserSession(env, userID)

	req, _ := http.NewRequest("GET", "/api/v1/admin/users/"+userID+"/sessions", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	sessions := resp["sessions"].([]interface{})
	assert.Len(t, sessions, 0)
}

func TestListUserSessions_UserNotFound(t *testing.T) {
	env := setupSessionHandlerTestEnv(t)
	missingID := uuid.New().String()

	req, _ := http.NewRequest("GET", "/api/v1/admin/users/"+missingID+"/sessions", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- Tests: RevokeUserSession ---

func TestRevokeUserSession_Success(t *testing.T) {
	env := setupSessionHandlerTestEnv(t)
	userID := uuid.New().String()
	clientID := uuid.New().String()

	seedUserSession(env, userID)
	addSession(env, userID, clientID, 10, true)

	req, _ := http.NewRequest("DELETE", "/api/v1/admin/users/"+userID+"/sessions/10", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRevokeUserSession_NotFound(t *testing.T) {
	env := setupSessionHandlerTestEnv(t)
	userID := uuid.New().String()
	seedUserSession(env, userID)

	req, _ := http.NewRequest("DELETE", "/api/v1/admin/users/"+userID+"/sessions/999", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Contains(t, []int{http.StatusNotFound, http.StatusInternalServerError}, w.Code)
}

func TestRevokeUserSession_InvalidSessionID(t *testing.T) {
	env := setupSessionHandlerTestEnv(t)
	userID := uuid.New().String()

	req, _ := http.NewRequest("DELETE", "/api/v1/admin/users/"+userID+"/sessions/abc", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- Tests: ListUserActivity (via ActivityHandler) ---

type activityHandlerTestEnv struct {
	router      *gin.Engine
	endUserRepo *stubEndUserRepo
	eventRepo   *stubUserEventRepo
	tenantID    string
	adminID     string
}

func setupActivityHandlerTestEnv(t *testing.T) *activityHandlerTestEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	tenantID := uuid.New().String()
	adminID := uuid.New().String()

	endUserRepo := newStubEndUserRepo()
	eventRepo := &stubUserEventRepo{}

	listUserActivityUC := user.NewListUserActivity(endUserRepo, eventRepo)

	handler := handlers.NewActivityHandler(listUserActivityUC)

	router := gin.New()
	admin := router.Group("/api/v1/admin", adminMiddleware(tenantID, adminID))
	users := admin.Group("/users")
	users.GET("/:id/activity", handler.ListUserActivity)

	return &activityHandlerTestEnv{
		router:      router,
		endUserRepo: endUserRepo,
		eventRepo:   eventRepo,
		tenantID:    tenantID,
		adminID:     adminID,
	}
}

func TestListUserActivity_Success(t *testing.T) {
	env := setupActivityHandlerTestEnv(t)
	userID := uuid.New().String()

	env.endUserRepo.users[userID] = &entities.User{
		ID:       userID,
		TenantID: env.tenantID,
		Email:    "active@example.com",
		Status:   entities.UserStatusActive,
	}

	now := time.Now()
	env.eventRepo.events = append(env.eventRepo.events,
		&entities.UserEvent{ID: 1, TenantID: env.tenantID, UserID: userID, EventType: entities.EventTypeLoginSuccess, OccurredAt: now},
		&entities.UserEvent{ID: 2, TenantID: env.tenantID, UserID: userID, EventType: entities.EventTypeTokenRefresh, OccurredAt: now},
	)

	req, _ := http.NewRequest("GET", "/api/v1/admin/users/"+userID+"/activity", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	events := resp["events"].([]interface{})
	assert.Len(t, events, 2)
}

func TestListUserActivity_Empty(t *testing.T) {
	env := setupActivityHandlerTestEnv(t)
	userID := uuid.New().String()

	env.endUserRepo.users[userID] = &entities.User{
		ID:       userID,
		TenantID: env.tenantID,
		Email:    "empty@example.com",
		Status:   entities.UserStatusActive,
	}

	req, _ := http.NewRequest("GET", "/api/v1/admin/users/"+userID+"/activity", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	events := resp["events"].([]interface{})
	assert.Len(t, events, 0)
}

func TestListUserActivity_Pagination(t *testing.T) {
	env := setupActivityHandlerTestEnv(t)
	userID := uuid.New().String()

	env.endUserRepo.users[userID] = &entities.User{
		ID:       userID,
		TenantID: env.tenantID,
		Email:    "paginated@example.com",
		Status:   entities.UserStatusActive,
	}

	now := time.Now()
	for i := 0; i < 5; i++ {
		env.eventRepo.events = append(env.eventRepo.events, &entities.UserEvent{
			ID:         int64(i + 1),
			TenantID:   env.tenantID,
			UserID:     userID,
			EventType:  entities.EventTypeLoginSuccess,
			OccurredAt: now,
		})
	}

	req, _ := http.NewRequest("GET", "/api/v1/admin/users/"+userID+"/activity?page=1&page_size=2", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	pagination := resp["pagination"].(map[string]interface{})
	assert.Equal(t, float64(5), pagination["total_count"])
}

func TestListUserActivity_UserNotFound(t *testing.T) {
	env := setupActivityHandlerTestEnv(t)
	missingID := uuid.New().String()

	req, _ := http.NewRequest("GET", "/api/v1/admin/users/"+missingID+"/activity", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	// The use case returns "user not found" error
	assert.Contains(t, []int{http.StatusNotFound, http.StatusInternalServerError}, w.Code)
}
