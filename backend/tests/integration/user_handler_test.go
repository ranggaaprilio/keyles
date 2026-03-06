/**
 * Integration tests for user management handler endpoints (T058)
 */

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"github.com/ranggaaprilio/keyles/domain/services"
	"github.com/ranggaaprilio/keyles/interfaces/http/handlers"
	"github.com/ranggaaprilio/keyles/usecase/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Stub implementations for user handler integration tests ---

type stubEndUserRepo struct {
	users       map[string]*entities.User
	tenantCount map[string]int
}

func newStubEndUserRepo() *stubEndUserRepo {
	return &stubEndUserRepo{
		users:       make(map[string]*entities.User),
		tenantCount: make(map[string]int),
	}
}

func (s *stubEndUserRepo) GetByID(_ context.Context, userID string) (*entities.User, error) {
	if u, ok := s.users[userID]; ok {
		return u, nil
	}
	return nil, errors.New("user not found")
}

func (s *stubEndUserRepo) GetByEmail(_ context.Context, tenantID, email string) (*entities.User, error) {
	for _, u := range s.users {
		if u.TenantID == tenantID && u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (s *stubEndUserRepo) Create(_ context.Context, u *entities.User) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	s.users[u.ID] = u
	s.tenantCount[u.TenantID]++
	return nil
}

func (s *stubEndUserRepo) Update(_ context.Context, u *entities.User) error {
	if _, ok := s.users[u.ID]; !ok {
		return errors.New("user not found")
	}
	s.users[u.ID] = u
	return nil
}

func (s *stubEndUserRepo) ListByTenant(_ context.Context, tenantID, search string, status entities.UserStatus, page, pageSize int) ([]*entities.User, int, error) {
	var result []*entities.User
	for _, u := range s.users {
		if u.TenantID != tenantID {
			continue
		}
		if status != "" && u.Status != status {
			continue
		}
		if search != "" && !containsIgnoreCase(u.Email, search) && !containsIgnoreCase(u.DisplayName, search) {
			continue
		}
		result = append(result, u)
	}
	total := len(result)
	start := (page - 1) * pageSize
	if start >= total {
		return []*entities.User{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return result[start:end], total, nil
}

func (s *stubEndUserRepo) CountByTenant(_ context.Context, tenantID string) (int, error) {
	count := 0
	for _, u := range s.users {
		if u.TenantID == tenantID {
			count++
		}
	}
	return count, nil
}

func (s *stubEndUserRepo) UpdateStatus(_ context.Context, userID string, status entities.UserStatus) error {
	u, ok := s.users[userID]
	if !ok {
		return errors.New("user not found")
	}
	u.Status = status
	return nil
}

func (s *stubEndUserRepo) UpdateLastLogin(_ context.Context, userID string, at time.Time) error {
	u, ok := s.users[userID]
	if !ok {
		return errors.New("user not found")
	}
	u.LastLoginAt = &at
	return nil
}

func (s *stubEndUserRepo) Delete(_ context.Context, userID string) error {
	u, ok := s.users[userID]
	if !ok {
		return errors.New("user not found")
	}
	delete(s.users, userID)
	s.tenantCount[u.TenantID]--
	return nil
}

// --- Stub InvitationRepository ---

type stubInvitationRepo struct {
	invitations map[string]*entities.Invitation
	byToken     map[string]*entities.Invitation // plainToken -> invitation (for test only)
}

func newStubInvitationRepo() *stubInvitationRepo {
	return &stubInvitationRepo{
		invitations: make(map[string]*entities.Invitation),
		byToken:     make(map[string]*entities.Invitation),
	}
}

func (s *stubInvitationRepo) Create(_ context.Context, inv *entities.Invitation) error {
	if inv.ID == "" {
		inv.ID = uuid.New().String()
	}
	s.invitations[inv.ID] = inv
	return nil
}

func (s *stubInvitationRepo) GetByToken(_ context.Context, plainToken string) (*entities.Invitation, error) {
	if inv, ok := s.byToken[plainToken]; ok {
		return inv, nil
	}
	return nil, errors.New("invitation not found")
}

func (s *stubInvitationRepo) GetPendingByEmail(_ context.Context, tenantID, email string) (*entities.Invitation, error) {
	for _, inv := range s.invitations {
		if inv.TenantID == tenantID && inv.Email == email && inv.Status == entities.InvitationStatusPending {
			return inv, nil
		}
	}
	return nil, errors.New("invitation not found")
}

func (s *stubInvitationRepo) UpdateStatus(_ context.Context, invitationID string, status entities.InvitationStatus, acceptedAt *time.Time) error {
	inv, ok := s.invitations[invitationID]
	if !ok {
		return errors.New("invitation not found")
	}
	inv.Status = status
	inv.AcceptedAt = acceptedAt
	return nil
}

func (s *stubInvitationRepo) ListByTenant(_ context.Context, tenantID string, page, pageSize int) ([]*entities.Invitation, int, error) {
	var result []*entities.Invitation
	for _, inv := range s.invitations {
		if inv.TenantID == tenantID {
			result = append(result, inv)
		}
	}
	return result, len(result), nil
}

func (s *stubInvitationRepo) ExpireStalePending(_ context.Context) (int64, error) {
	return 0, nil
}

// --- Stub UserEventRepository ---

type stubUserEventRepo struct {
	events []*entities.UserEvent
}

func (s *stubUserEventRepo) Record(_ context.Context, event *entities.UserEvent) error {
	s.events = append(s.events, event)
	return nil
}

func (s *stubUserEventRepo) ListByUser(_ context.Context, userID string, page, pageSize int) ([]*entities.UserEvent, int, error) {
	var result []*entities.UserEvent
	for _, e := range s.events {
		if e.UserID == userID {
			result = append(result, e)
		}
	}
	total := len(result)
	start := (page - 1) * pageSize
	if start >= total {
		return []*entities.UserEvent{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return result[start:end], total, nil
}

func (s *stubUserEventRepo) DeleteOlderThan(_ context.Context, before time.Time) (int64, error) {
	return 0, nil
}

// --- Stub EmailService ---

type stubEmailService struct {
	sentInvitations []string
}

func (s *stubEmailService) SendOTPEmail(_ context.Context, _, _, _, _ string) error  { return nil }
func (s *stubEmailService) SendWelcomeEmail(_ context.Context, _, _, _ string) error { return nil }
func (s *stubEmailService) SendInvitationEmail(_ context.Context, toEmail, _, _, _ string) error {
	s.sentInvitations = append(s.sentInvitations, toEmail)
	return nil
}

// --- Stub UserCountCache ---

type stubUserCountCache struct {
	counts map[string]int
}

func newStubUserCountCache() *stubUserCountCache {
	return &stubUserCountCache{counts: make(map[string]int)}
}

func (s *stubUserCountCache) Get(_ context.Context, tenantID string) (int, bool, error) {
	c, ok := s.counts[tenantID]
	return c, ok, nil
}
func (s *stubUserCountCache) Set(_ context.Context, tenantID string, count int) error {
	s.counts[tenantID] = count
	return nil
}
func (s *stubUserCountCache) Invalidate(_ context.Context, tenantID string) error {
	delete(s.counts, tenantID)
	return nil
}

// --- Stub UserBlacklist ---

type stubUserBlacklist struct {
	blacklisted map[string]bool
}

func newStubUserBlacklist() *stubUserBlacklist {
	return &stubUserBlacklist{blacklisted: make(map[string]bool)}
}

func (s *stubUserBlacklist) Add(_ context.Context, userID string, _ time.Duration) error {
	s.blacklisted[userID] = true
	return nil
}

func (s *stubUserBlacklist) IsBlacklisted(_ context.Context, userID string) (bool, error) {
	return s.blacklisted[userID], nil
}

// --- Stub RoleRepository ---

type stubRoleRepo struct {
	assignments []*entities.UserRoleAssignment
}

func (s *stubRoleRepo) AssignRole(_ context.Context, a *entities.UserRoleAssignment) error {
	s.assignments = append(s.assignments, a)
	return nil
}
func (s *stubRoleRepo) RevokeRole(_ context.Context, _, _, _ string) error { return nil }
func (s *stubRoleRepo) GetUserRoles(_ context.Context, _, _ string) ([]*entities.UserRoleAssignment, error) {
	return nil, nil
}
func (s *stubRoleRepo) HasRole(_ context.Context, _, _, _ string) (bool, error) { return false, nil }
func (s *stubRoleRepo) HasAnyRole(_ context.Context, _, _ string) (bool, error) { return false, nil }
func (s *stubRoleRepo) ListRolesByClient(_ context.Context, _ string) ([]*entities.UserRoleAssignment, error) {
	return nil, nil
}
func (s *stubRoleRepo) ListRolesByUser(_ context.Context, _ string) ([]*entities.UserRoleAssignment, error) {
	return nil, nil
}
func (s *stubRoleRepo) Assign(_ context.Context, a *entities.UserRoleAssignment) error {
	s.assignments = append(s.assignments, a)
	return nil
}
func (s *stubRoleRepo) Revoke(_ context.Context, _ int64, _ string) error { return nil }
func (s *stubRoleRepo) ListByUser(_ context.Context, userID string) ([]*entities.UserRoleAssignment, error) {
	var result []*entities.UserRoleAssignment
	for _, a := range s.assignments {
		if a.UserID == userID {
			result = append(result, a)
		}
	}
	return result, nil
}
func (s *stubRoleRepo) ListByClient(_ context.Context, _ string, _, _ int) ([]*entities.UserRoleAssignment, int, error) {
	return nil, 0, nil
}
func (s *stubRoleRepo) RevokeAllForUser(_ context.Context, userID, _ string) error {
	for _, a := range s.assignments {
		if a.UserID == userID {
			a.IsActive = false
		}
	}
	return nil
}
func (s *stubRoleRepo) GetActiveRoles(_ context.Context, userID, clientID string) ([]string, error) {
	var roles []string
	for _, a := range s.assignments {
		if a.UserID == userID && a.ClientID == clientID && a.IsActive {
			roles = append(roles, a.Role)
		}
	}
	return roles, nil
}

// --- Stub RefreshTokenRepository ---

type stubRefreshTokenRepo struct {
	tokens []*entities.RefreshToken
}

func (s *stubRefreshTokenRepo) Create(_ context.Context, t *entities.RefreshToken) error {
	s.tokens = append(s.tokens, t)
	return nil
}
func (s *stubRefreshTokenRepo) GetByToken(_ context.Context, _ string) (*entities.RefreshToken, error) {
	return nil, errors.New("not found")
}
func (s *stubRefreshTokenRepo) Revoke(_ context.Context, _, _ string) error { return nil }
func (s *stubRefreshTokenRepo) RevokeAllForUser(_ context.Context, _, _ string) error {
	return nil
}
func (s *stubRefreshTokenRepo) DeleteExpired(_ context.Context) (int64, error) { return 0, nil }
func (s *stubRefreshTokenRepo) IsRevoked(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (s *stubRefreshTokenRepo) UpdateLastUsed(_ context.Context, _ string) error { return nil }
func (s *stubRefreshTokenRepo) RevokeByClientID(_ context.Context, _ string) error {
	return nil
}
func (s *stubRefreshTokenRepo) RevokeByUserID(_ context.Context, userID string) error {
	for _, t := range s.tokens {
		if t.UserID == userID {
			t.RevokedFlag = true
		}
	}
	return nil
}
func (s *stubRefreshTokenRepo) ListByUserID(_ context.Context, userID string) ([]*entities.RefreshToken, error) {
	var result []*entities.RefreshToken
	for _, t := range s.tokens {
		if t.UserID == userID {
			result = append(result, t)
		}
	}
	return result, nil
}
func (s *stubRefreshTokenRepo) GetByID(_ context.Context, id int64) (*entities.RefreshToken, error) {
	for _, t := range s.tokens {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, errors.New("not found")
}

// --- Stub AuditRepository ---

type stubAuditRepo struct {
	logs []*entities.AuditLog
}

func (s *stubAuditRepo) Create(_ context.Context, log *entities.AuditLog) error {
	s.logs = append(s.logs, log)
	return nil
}
func (s *stubAuditRepo) FindByTenantID(_ context.Context, _ uuid.UUID, _, _ int) ([]*entities.AuditLog, error) {
	return nil, nil
}
func (s *stubAuditRepo) FindByEventType(_ context.Context, _ entities.EventType, _, _ int) ([]*entities.AuditLog, error) {
	return nil, nil
}
func (s *stubAuditRepo) FindRecent(_ context.Context, _ int) ([]*entities.AuditLog, error) {
	return nil, nil
}

// --- Stub UserRepository (AdminUser) ---

type stubAdminUserRepo struct {
	adminUsers map[uuid.UUID]*entities.AdminUser
}

func newStubAdminUserRepo() *stubAdminUserRepo {
	return &stubAdminUserRepo{adminUsers: make(map[uuid.UUID]*entities.AdminUser)}
}

func (s *stubAdminUserRepo) Create(_ context.Context, u *entities.AdminUser) error {
	s.adminUsers[u.ID] = u
	return nil
}
func (s *stubAdminUserRepo) FindByID(_ context.Context, id uuid.UUID) (*entities.AdminUser, error) {
	if u, ok := s.adminUsers[id]; ok {
		return u, nil
	}
	return nil, errors.New("user not found")
}
func (s *stubAdminUserRepo) FindByEmail(_ context.Context, email string) (*entities.AdminUser, error) {
	for _, u := range s.adminUsers {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}
func (s *stubAdminUserRepo) FindByTenantID(_ context.Context, tenantID uuid.UUID) ([]*entities.AdminUser, error) {
	var result []*entities.AdminUser
	for _, u := range s.adminUsers {
		if u.TenantID == tenantID {
			result = append(result, u)
		}
	}
	return result, nil
}
func (s *stubAdminUserRepo) Update(_ context.Context, u *entities.AdminUser) error {
	s.adminUsers[u.ID] = u
	return nil
}
func (s *stubAdminUserRepo) EmailExists(_ context.Context, email string) (bool, error) {
	for _, u := range s.adminUsers {
		if u.Email == email {
			return true, nil
		}
	}
	return false, nil
}
func (s *stubAdminUserRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(s.adminUsers, id)
	return nil
}

// --- Stub PasswordService ---

type stubPasswordService struct{}

func (s *stubPasswordService) Hash(password string) (string, error) {
	return "$2a$12$test-hash-" + password, nil
}

func (s *stubPasswordService) Verify(password, hash string) error {
	if hash == "$2a$12$test-hash-"+password {
		return nil
	}
	return errors.New("password mismatch")
}

// --- Helpers ---

func containsIgnoreCase(s, substr string) bool {
	sl := len(s)
	subl := len(substr)
	if subl > sl {
		return false
	}
	for i := 0; i <= sl-subl; i++ {
		if equalFoldAt(s, substr, i) {
			return true
		}
	}
	return false
}

func equalFoldAt(s, sub string, pos int) bool {
	for j := 0; j < len(sub); j++ {
		a, b := s[pos+j], sub[j]
		if a >= 'A' && a <= 'Z' {
			a += 32
		}
		if b >= 'A' && b <= 'Z' {
			b += 32
		}
		if a != b {
			return false
		}
	}
	return true
}

func setAdminContext(c *gin.Context, tenantID, adminID string) {
	c.Set("tenant_id", tenantID)
	c.Set("user_id", adminID)
}

func adminMiddleware(tenantID, adminID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		setAdminContext(c, tenantID, adminID)
		c.Next()
	}
}

// --- Test Setup ---

type userHandlerTestEnv struct {
	router           *gin.Engine
	endUserRepo      *stubEndUserRepo
	invitationRepo   *stubInvitationRepo
	eventRepo        *stubUserEventRepo
	emailService     *stubEmailService
	countCache       *stubUserCountCache
	blacklist        *stubUserBlacklist
	roleRepo         *stubRoleRepo
	refreshTokenRepo *stubRefreshTokenRepo
	auditRepo        *stubAuditRepo
	adminUserRepo    *stubAdminUserRepo
	tenantID         string
	adminID          string
}

func setupUserHandlerTestEnv(t *testing.T) *userHandlerTestEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	tenantID := uuid.New().String()
	adminID := uuid.New().String()

	endUserRepo := newStubEndUserRepo()
	invitationRepo := newStubInvitationRepo()
	eventRepo := &stubUserEventRepo{}
	emailSvc := &stubEmailService{}
	countCache := newStubUserCountCache()
	blacklist := newStubUserBlacklist()
	roleRepo := &stubRoleRepo{}
	refreshTokenRepo := &stubRefreshTokenRepo{}
	auditRepo := &stubAuditRepo{}
	adminUserRepo := newStubAdminUserRepo()

	inviteUserUC := user.NewInviteUser(endUserRepo, invitationRepo, eventRepo, emailSvc, countCache)
	getUserUC := user.NewGetUser(endUserRepo, roleRepo)
	listUsersUC := user.NewListUsers(endUserRepo)
	updateUserUC := user.NewUpdateUser(endUserRepo, auditRepo)
	disableUserUC := user.NewDisableUser(endUserRepo, adminUserRepo, refreshTokenRepo, blacklist, eventRepo, auditRepo)
	enableUserUC := user.NewEnableUser(endUserRepo, eventRepo, auditRepo)
	deleteUserUC := user.NewDeleteUser(endUserRepo, roleRepo, refreshTokenRepo, blacklist, auditRepo)
	resendInvitationUC := user.NewResendInvitation(endUserRepo, invitationRepo, eventRepo, emailSvc)

	handler := handlers.NewUserHandler(
		inviteUserUC, getUserUC, listUsersUC, updateUserUC,
		disableUserUC, enableUserUC, deleteUserUC, resendInvitationUC,
	)

	router := gin.New()
	admin := router.Group("/api/v1/admin", adminMiddleware(tenantID, adminID))
	{
		users := admin.Group("/users")
		{
			users.POST("/invite", handler.InviteUser)
			users.GET("", handler.ListUsers)
			users.GET("/:id", handler.GetUser)
			users.PATCH("/:id", handler.UpdateUser)
			users.PUT("/:id/status", handler.UpdateUserStatus)
			users.DELETE("/:id", handler.DeleteUser)
			users.POST("/:id/resend-invitation", handler.ResendInvitation)
		}
	}

	return &userHandlerTestEnv{
		router:           router,
		endUserRepo:      endUserRepo,
		invitationRepo:   invitationRepo,
		eventRepo:        eventRepo,
		emailService:     emailSvc,
		countCache:       countCache,
		blacklist:        blacklist,
		roleRepo:         roleRepo,
		refreshTokenRepo: refreshTokenRepo,
		auditRepo:        auditRepo,
		adminUserRepo:    adminUserRepo,
		tenantID:         tenantID,
		adminID:          adminID,
	}
}

func (env *userHandlerTestEnv) seedUser(id, email string, status entities.UserStatus) *entities.User {
	u := &entities.User{
		ID:          id,
		TenantID:    env.tenantID,
		Email:       email,
		DisplayName: "Test User",
		Status:      status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	env.endUserRepo.users[id] = u
	return u
}

// --- Tests ---

func TestInviteUser_Success(t *testing.T) {
	env := setupUserHandlerTestEnv(t)

	body, _ := json.Marshal(map[string]string{
		"email":        "alice@example.com",
		"display_name": "Alice Smith",
	})
	req, _ := http.NewRequest("POST", "/api/v1/admin/users/invite", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "alice@example.com", resp["email"])
	assert.Equal(t, "pending", resp["status"])
	assert.NotEmpty(t, resp["user_id"])
	assert.NotEmpty(t, resp["invitation_id"])
	assert.Len(t, env.emailService.sentInvitations, 1)
}

func TestInviteUser_DuplicateEmail(t *testing.T) {
	env := setupUserHandlerTestEnv(t)
	env.seedUser(uuid.New().String(), "alice@example.com", entities.UserStatusActive)

	body, _ := json.Marshal(map[string]string{
		"email":        "alice@example.com",
		"display_name": "Alice Duplicate",
	})
	req, _ := http.NewRequest("POST", "/api/v1/admin/users/invite", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "conflict", resp["error"])
}

func TestInviteUser_BadEmail(t *testing.T) {
	env := setupUserHandlerTestEnv(t)

	body, _ := json.Marshal(map[string]string{
		"email": "not-an-email",
	})
	req, _ := http.NewRequest("POST", "/api/v1/admin/users/invite", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInviteUser_QuotaExceeded(t *testing.T) {
	env := setupUserHandlerTestEnv(t)
	// Set cache to report quota max
	env.countCache.Set(context.Background(), env.tenantID, entities.MaxUsersPerTenant)

	body, _ := json.Marshal(map[string]string{
		"email":        "quota@example.com",
		"display_name": "Quota Test",
	})
	req, _ := http.NewRequest("POST", "/api/v1/admin/users/invite", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "quota_exceeded", resp["error"])
}

func TestListUsers_PaginatedList(t *testing.T) {
	env := setupUserHandlerTestEnv(t)

	// Seed 5 users
	for i := 0; i < 5; i++ {
		env.seedUser(uuid.New().String(), fmt.Sprintf("user%d@example.com", i), entities.UserStatusActive)
	}

	req, _ := http.NewRequest("GET", "/api/v1/admin/users?page=1&page_size=3", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	users := resp["users"].([]interface{})
	assert.Len(t, users, 3)

	pagination := resp["pagination"].(map[string]interface{})
	assert.Equal(t, float64(5), pagination["total_count"])
	assert.Equal(t, float64(1), pagination["page"])
}

func TestListUsers_SearchFilter(t *testing.T) {
	env := setupUserHandlerTestEnv(t)
	env.seedUser(uuid.New().String(), "alice@example.com", entities.UserStatusActive)
	env.seedUser(uuid.New().String(), "bob@example.com", entities.UserStatusActive)

	req, _ := http.NewRequest("GET", "/api/v1/admin/users?search=alice", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	users := resp["users"].([]interface{})
	assert.Len(t, users, 1)
}

func TestListUsers_StatusFilter(t *testing.T) {
	env := setupUserHandlerTestEnv(t)
	env.seedUser(uuid.New().String(), "active@example.com", entities.UserStatusActive)
	env.seedUser(uuid.New().String(), "pending@example.com", entities.UserStatusPending)

	req, _ := http.NewRequest("GET", "/api/v1/admin/users?status=pending", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	users := resp["users"].([]interface{})
	assert.Len(t, users, 1)
}

func TestGetUser_Success(t *testing.T) {
	env := setupUserHandlerTestEnv(t)
	userID := uuid.New().String()
	env.seedUser(userID, "alice@example.com", entities.UserStatusActive)

	req, _ := http.NewRequest("GET", "/api/v1/admin/users/"+userID, nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "alice@example.com", resp["email"])
	assert.Equal(t, "active", resp["status"])
}

func TestGetUser_NotFound(t *testing.T) {
	env := setupUserHandlerTestEnv(t)

	req, _ := http.NewRequest("GET", "/api/v1/admin/users/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateUser_DisplayName(t *testing.T) {
	env := setupUserHandlerTestEnv(t)
	userID := uuid.New().String()
	env.seedUser(userID, "alice@example.com", entities.UserStatusActive)

	newName := "Alice Updated"
	body, _ := json.Marshal(map[string]*string{
		"display_name": &newName,
	})
	req, _ := http.NewRequest("PATCH", "/api/v1/admin/users/"+userID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Alice Updated", resp["display_name"])
}

func TestUpdateUserStatus_Disable(t *testing.T) {
	env := setupUserHandlerTestEnv(t)
	userID := uuid.New().String()
	env.seedUser(userID, "alice@example.com", entities.UserStatusActive)

	body, _ := json.Marshal(map[string]string{"status": "disabled"})
	req, _ := http.NewRequest("PUT", "/api/v1/admin/users/"+userID+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "disabled", resp["status"])
	// Verify blacklist was set
	assert.True(t, env.blacklist.blacklisted[userID])
}

func TestUpdateUserStatus_DisableSelf(t *testing.T) {
	env := setupUserHandlerTestEnv(t)
	// admin trying to disable themselves
	env.seedUser(env.adminID, "admin@example.com", entities.UserStatusActive)

	body, _ := json.Marshal(map[string]string{"status": "disabled"})
	req, _ := http.NewRequest("PUT", "/api/v1/admin/users/"+env.adminID+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpdateUserStatus_Enable(t *testing.T) {
	env := setupUserHandlerTestEnv(t)
	userID := uuid.New().String()
	env.seedUser(userID, "alice@example.com", entities.UserStatusDisabled)

	body, _ := json.Marshal(map[string]string{"status": "active"})
	req, _ := http.NewRequest("PUT", "/api/v1/admin/users/"+userID+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "active", resp["status"])
}

func TestUpdateUserStatus_InvalidStatus(t *testing.T) {
	env := setupUserHandlerTestEnv(t)
	userID := uuid.New().String()
	env.seedUser(userID, "alice@example.com", entities.UserStatusActive)

	body, _ := json.Marshal(map[string]string{"status": "invalid"})
	req, _ := http.NewRequest("PUT", "/api/v1/admin/users/"+userID+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteUser_Success(t *testing.T) {
	env := setupUserHandlerTestEnv(t)
	userID := uuid.New().String()
	env.seedUser(userID, "alice@example.com", entities.UserStatusActive)

	req, _ := http.NewRequest("DELETE", "/api/v1/admin/users/"+userID, nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	// User removed from repo
	_, err := env.endUserRepo.GetByID(context.Background(), userID)
	assert.Error(t, err)
	// Blacklist set
	assert.True(t, env.blacklist.blacklisted[userID])
}

func TestDeleteUser_CannotDeleteSelf(t *testing.T) {
	env := setupUserHandlerTestEnv(t)
	env.seedUser(env.adminID, "admin@example.com", entities.UserStatusActive)

	req, _ := http.NewRequest("DELETE", "/api/v1/admin/users/"+env.adminID, nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestResendInvitation_Success(t *testing.T) {
	env := setupUserHandlerTestEnv(t)
	userID := uuid.New().String()
	env.seedUser(userID, "pending@example.com", entities.UserStatusPending)
	env.invitationRepo.invitations["inv1"] = &entities.Invitation{
		ID:       "inv1",
		TenantID: env.tenantID,
		Email:    "pending@example.com",
		Status:   entities.InvitationStatusPending,
	}

	req, _ := http.NewRequest("POST", "/api/v1/admin/users/"+userID+"/resend-invitation", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Invitation resent successfully.", resp["message"])
}

func TestResendInvitation_UserNotPending(t *testing.T) {
	env := setupUserHandlerTestEnv(t)
	userID := uuid.New().String()
	env.seedUser(userID, "active@example.com", entities.UserStatusActive)

	req, _ := http.NewRequest("POST", "/api/v1/admin/users/"+userID+"/resend-invitation", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Verify all required interfaces are satisfied at compile time
var _ repositories.EndUserRepository = (*stubEndUserRepo)(nil)
var _ repositories.InvitationRepository = (*stubInvitationRepo)(nil)
var _ repositories.UserEventRepository = (*stubUserEventRepo)(nil)
var _ repositories.RefreshTokenRepository = (*stubRefreshTokenRepo)(nil)
var _ repositories.RoleRepository = (*stubRoleRepo)(nil)
var _ repositories.AuditRepository = (*stubAuditRepo)(nil)
var _ repositories.UserRepository = (*stubAdminUserRepo)(nil)
var _ services.EmailService = (*stubEmailService)(nil)
var _ services.UserCountCache = (*stubUserCountCache)(nil)
var _ services.UserBlacklist = (*stubUserBlacklist)(nil)
var _ services.PasswordService = (*stubPasswordService)(nil)
