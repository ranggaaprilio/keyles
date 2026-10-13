/**
 * Shared in-memory repositories for OAuth integration tests
 */

package integration

import (
	"context"
	"errors"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// MockIntegrationAuthCodeRepository for integration tests
type MockIntegrationAuthCodeRepository struct {
	codes map[string]*entities.AuthorizationCode
}

func NewMockIntegrationAuthCodeRepository() *MockIntegrationAuthCodeRepository {
	return &MockIntegrationAuthCodeRepository{
		codes: make(map[string]*entities.AuthorizationCode),
	}
}

func (m *MockIntegrationAuthCodeRepository) Store(ctx context.Context, code *entities.AuthorizationCode, ttl time.Duration) error {
	m.codes[code.Code] = code
	return nil
}

func (m *MockIntegrationAuthCodeRepository) Get(ctx context.Context, code string) (*entities.AuthorizationCode, error) {
	if c, ok := m.codes[code]; ok {
		return c, nil
	}
	return nil, errors.New("authorization code not found")
}

func (m *MockIntegrationAuthCodeRepository) MarkAsUsed(ctx context.Context, code string) error {
	if c, ok := m.codes[code]; ok {
		c.MarkAsUsed()
		return nil
	}
	return errors.New("authorization code not found")
}

func (m *MockIntegrationAuthCodeRepository) Delete(ctx context.Context, code string) error {
	delete(m.codes, code)
	return nil
}

func (m *MockIntegrationAuthCodeRepository) Exists(ctx context.Context, code string) (bool, error) {
	_, ok := m.codes[code]
	return ok, nil
}

var _ repositories.AuthCodeRepository = (*MockIntegrationAuthCodeRepository)(nil)

// MockIntegrationRoleRepository for integration tests
type MockIntegrationRoleRepository struct {
	roles map[string][]*entities.UserRoleAssignment
}

func NewMockIntegrationRoleRepository() *MockIntegrationRoleRepository {
	return &MockIntegrationRoleRepository{
		roles: make(map[string][]*entities.UserRoleAssignment),
	}
}

func (m *MockIntegrationRoleRepository) AssignRole(ctx context.Context, assignment *entities.UserRoleAssignment) error {
	key := assignment.UserID + ":" + assignment.ClientID
	m.roles[key] = append(m.roles[key], assignment)
	return nil
}

func (m *MockIntegrationRoleRepository) RevokeRole(ctx context.Context, userID, clientID, role string) error {
	return nil
}

func (m *MockIntegrationRoleRepository) GetUserRoles(ctx context.Context, userID, clientID string) ([]*entities.UserRoleAssignment, error) {
	key := userID + ":" + clientID
	return m.roles[key], nil
}

func (m *MockIntegrationRoleRepository) HasRole(ctx context.Context, userID, clientID, role string) (bool, error) {
	key := userID + ":" + clientID
	for _, r := range m.roles[key] {
		if r.Role == role && r.IsActive {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockIntegrationRoleRepository) HasAnyRole(ctx context.Context, userID, clientID string) (bool, error) {
	key := userID + ":" + clientID
	for _, r := range m.roles[key] {
		if r.IsActive {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockIntegrationRoleRepository) ListRolesByClient(ctx context.Context, clientID string) ([]*entities.UserRoleAssignment, error) {
	return nil, nil
}

func (m *MockIntegrationRoleRepository) ListRolesByUser(ctx context.Context, userID string) ([]*entities.UserRoleAssignment, error) {
	return nil, nil
}

func (m *MockIntegrationRoleRepository) Assign(ctx context.Context, assignment *entities.UserRoleAssignment) error {
	key := assignment.UserID + ":" + assignment.ClientID
	m.roles[key] = append(m.roles[key], assignment)
	return nil
}

func (m *MockIntegrationRoleRepository) Revoke(ctx context.Context, assignmentID int64, revokedByUserID string) error {
	return nil
}

func (m *MockIntegrationRoleRepository) ListByUser(ctx context.Context, userID string) ([]*entities.UserRoleAssignment, error) {
	return nil, nil
}

func (m *MockIntegrationRoleRepository) ListByClient(ctx context.Context, clientID string, page, pageSize int) ([]*entities.UserRoleAssignment, int, error) {
	return nil, 0, nil
}

func (m *MockIntegrationRoleRepository) RevokeAllForUser(ctx context.Context, userID, revokedByUserID string) error {
	return nil
}

func (m *MockIntegrationRoleRepository) GetActiveRoles(ctx context.Context, userID, clientID string) ([]string, error) {
	key := userID + ":" + clientID
	var roles []string
	for _, r := range m.roles[key] {
		if r.IsActive {
			roles = append(roles, r.Role)
		}
	}
	return roles, nil
}

var _ repositories.RoleRepository = (*MockIntegrationRoleRepository)(nil)

// MockIntegrationEndUserRepository for integration tests
type MockIntegrationEndUserRepository struct {
	users map[string]*entities.User
}

func NewMockIntegrationEndUserRepository() *MockIntegrationEndUserRepository {
	return &MockIntegrationEndUserRepository{
		users: make(map[string]*entities.User),
	}
}

func (m *MockIntegrationEndUserRepository) GetByID(ctx context.Context, userID string) (*entities.User, error) {
	if u, ok := m.users[userID]; ok {
		return u, nil
	}
	return nil, errors.New("user not found")
}

func (m *MockIntegrationEndUserRepository) GetByEmail(ctx context.Context, tenantID, email string) (*entities.User, error) {
	for _, u := range m.users {
		if u.TenantID == tenantID && u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockIntegrationEndUserRepository) Create(ctx context.Context, user *entities.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockIntegrationEndUserRepository) Update(ctx context.Context, user *entities.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockIntegrationEndUserRepository) ListByTenant(ctx context.Context, tenantID, search string, status entities.UserStatus, page, pageSize int) ([]*entities.User, int, error) {
	return nil, 0, nil
}

func (m *MockIntegrationEndUserRepository) CountByTenant(ctx context.Context, tenantID string) (int, error) {
	return len(m.users), nil
}

func (m *MockIntegrationEndUserRepository) UpdateStatus(ctx context.Context, userID string, status entities.UserStatus) error {
	if u, ok := m.users[userID]; ok {
		u.Status = status
	}
	return nil
}

func (m *MockIntegrationEndUserRepository) UpdateLastLogin(ctx context.Context, userID string, at time.Time) error {
	return nil
}

func (m *MockIntegrationEndUserRepository) Delete(ctx context.Context, userID string) error {
	delete(m.users, userID)
	return nil
}

var _ repositories.EndUserRepository = (*MockIntegrationEndUserRepository)(nil)
