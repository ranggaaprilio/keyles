package usecase_test

import (
"context"
"errors"
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/mock"

"github.com/ranggaaprilio/keyles/domain/entities"
"github.com/ranggaaprilio/keyles/tests/mocks"
"github.com/ranggaaprilio/keyles/usecase/role"
)

// TestRevokeRole_Success tests successful role revocation using new Revoke method
func TestRevokeRole_Success(t *testing.T) {
	ctx := context.Background()
	mockRoleRepo := new(mocks.MockRoleRepository)
	mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	mockEventRepo := new(mocks.MockUserEventRepository)
	mockAuditRepo := new(mocks.MockAuditRepository)

	mockRoleRepo.On("Revoke", ctx, int64(42), "admin-123").Return(nil)
	mockRefreshTokenRepo.On("RevokeAllForUser", ctx, "user-123", "client-123").Return(nil)
	mockEventRepo.On("Record", ctx, mock.AnythingOfType("*entities.UserEvent")).Return(nil)
	mockAuditRepo.On("Create", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

	uc := role.NewRevokeRole(mockRoleRepo, mockRefreshTokenRepo, mockEventRepo, mockAuditRepo)

	err := uc.Execute(ctx, role.RevokeRoleRequest{
AssignmentID: 42,
UserID:       "user-123",
ClientID:     "client-123",
TenantID:     "tenant-123",
RevokedBy:    "admin-123",
})

	assert.NoError(t, err)
	mockRoleRepo.AssertCalled(t, "Revoke", ctx, int64(42), "admin-123")
	mockRefreshTokenRepo.AssertCalled(t, "RevokeAllForUser", ctx, "user-123", "client-123")
}

// TestRevokeRole_RoleNotFound tests error when role assignment doesn't exist
func TestRevokeRole_RoleNotFound(t *testing.T) {
ctx := context.Background()
mockRoleRepo := new(mocks.MockRoleRepository)
mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)
mockEventRepo := new(mocks.MockUserEventRepository)
mockAuditRepo := new(mocks.MockAuditRepository)

mockRoleRepo.On("Revoke", ctx, int64(999), "admin-123").Return(errors.New("assignment not found"))

uc := role.NewRevokeRole(mockRoleRepo, mockRefreshTokenRepo, mockEventRepo, mockAuditRepo)

err := uc.Execute(ctx, role.RevokeRoleRequest{
AssignmentID: 999,
UserID:       "user-123",
RevokedBy:    "admin-123",
})

assert.Error(t, err)
assert.Contains(t, err.Error(), "failed to revoke role")
}

// TestRevokeRole_CascadeRefreshTokenRevocation tests that refresh tokens are revoked
func TestRevokeRole_CascadeRefreshTokenRevocation(t *testing.T) {
ctx := context.Background()
mockRoleRepo := new(mocks.MockRoleRepository)
mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)
mockEventRepo := new(mocks.MockUserEventRepository)
mockAuditRepo := new(mocks.MockAuditRepository)

mockRoleRepo.On("Revoke", ctx, int64(42), "admin-123").Return(nil)
mockRefreshTokenRepo.On("RevokeAllForUser", ctx, "user-123", "client-123").Return(nil)
mockEventRepo.On("Record", ctx, mock.AnythingOfType("*entities.UserEvent")).Return(nil)
mockAuditRepo.On("Create", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

uc := role.NewRevokeRole(mockRoleRepo, mockRefreshTokenRepo, mockEventRepo, mockAuditRepo)

err := uc.Execute(ctx, role.RevokeRoleRequest{
AssignmentID: 42,
UserID:       "user-123",
ClientID:     "client-123",
TenantID:     "tenant-123",
RevokedBy:    "admin-123",
})

assert.NoError(t, err)
mockRefreshTokenRepo.AssertCalled(t, "RevokeAllForUser", ctx, "user-123", "client-123")
}

// TestRevokeRole_NoClientIDSkipsCascade tests that cascade is skipped without clientID
func TestRevokeRole_NoClientIDSkipsCascade(t *testing.T) {
ctx := context.Background()
mockRoleRepo := new(mocks.MockRoleRepository)
mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)
mockEventRepo := new(mocks.MockUserEventRepository)
mockAuditRepo := new(mocks.MockAuditRepository)

mockRoleRepo.On("Revoke", ctx, int64(42), "admin-123").Return(nil)
mockEventRepo.On("Record", ctx, mock.AnythingOfType("*entities.UserEvent")).Return(nil)
mockAuditRepo.On("Create", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

uc := role.NewRevokeRole(mockRoleRepo, mockRefreshTokenRepo, mockEventRepo, mockAuditRepo)

err := uc.Execute(ctx, role.RevokeRoleRequest{
AssignmentID: 42,
UserID:       "user-123",
TenantID:     "tenant-123",
RevokedBy:    "admin-123",
})

assert.NoError(t, err)
mockRefreshTokenRepo.AssertNotCalled(t, "RevokeAllForUser")
}

// TestRevokeRole_MissingFields tests validation for missing required fields
func TestRevokeRole_MissingFields(t *testing.T) {
ctx := context.Background()
mockRoleRepo := new(mocks.MockRoleRepository)
mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)
mockEventRepo := new(mocks.MockUserEventRepository)
mockAuditRepo := new(mocks.MockAuditRepository)

uc := role.NewRevokeRole(mockRoleRepo, mockRefreshTokenRepo, mockEventRepo, mockAuditRepo)

tests := []struct {
name string
req  role.RevokeRoleRequest
msg  string
}{
{
name: "Missing assignment_id",
req:  role.RevokeRoleRequest{UserID: "u", RevokedBy: "a"},
msg:  "assignment_id is required",
},
{
name: "Missing user_id",
req:  role.RevokeRoleRequest{AssignmentID: 1, RevokedBy: "a"},
msg:  "user_id is required",
},
{
name: "Missing revoked_by",
req:  role.RevokeRoleRequest{AssignmentID: 1, UserID: "u"},
msg:  "revoked_by is required",
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
err := uc.Execute(ctx, tt.req)
assert.Error(t, err)
assert.Contains(t, err.Error(), tt.msg)
})
}
}

// TestRevokeRole_DatabaseError tests database error handling
func TestRevokeRole_DatabaseError(t *testing.T) {
ctx := context.Background()
mockRoleRepo := new(mocks.MockRoleRepository)
mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)
mockEventRepo := new(mocks.MockUserEventRepository)
mockAuditRepo := new(mocks.MockAuditRepository)

mockRoleRepo.On("Revoke", ctx, int64(42), "admin-123").Return(errors.New("database error"))

uc := role.NewRevokeRole(mockRoleRepo, mockRefreshTokenRepo, mockEventRepo, mockAuditRepo)

err := uc.Execute(ctx, role.RevokeRoleRequest{
AssignmentID: 42,
UserID:       "user-123",
RevokedBy:    "admin-123",
})

assert.Error(t, err)
}

// TestRevokeRole_EventRecorded verifies role_revoked event is recorded
func TestRevokeRole_EventRecorded(t *testing.T) {
ctx := context.Background()
mockRoleRepo := new(mocks.MockRoleRepository)
mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)
mockEventRepo := new(mocks.MockUserEventRepository)
mockAuditRepo := new(mocks.MockAuditRepository)

mockRoleRepo.On("Revoke", ctx, int64(42), "admin-123").Return(nil)
mockEventRepo.On("Record", ctx, mock.MatchedBy(func(e *entities.UserEvent) bool {
return e.EventType == entities.EventTypeRoleRevoked && e.UserID == "user-123"
})).Return(nil)
mockAuditRepo.On("Create", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

uc := role.NewRevokeRole(mockRoleRepo, mockRefreshTokenRepo, mockEventRepo, mockAuditRepo)

err := uc.Execute(ctx, role.RevokeRoleRequest{
AssignmentID: 42,
UserID:       "user-123",
TenantID:     "tenant-123",
RevokedBy:    "admin-123",
})

assert.NoError(t, err)
mockEventRepo.AssertExpectations(t)
}

// TestRevokeRole_RefreshTokenRevocationError tests that revoke still succeeds on cascade error
func TestRevokeRole_RefreshTokenRevocationError(t *testing.T) {
ctx := context.Background()
mockRoleRepo := new(mocks.MockRoleRepository)
mockRefreshTokenRepo := new(mocks.MockRefreshTokenRepository)
mockEventRepo := new(mocks.MockUserEventRepository)
mockAuditRepo := new(mocks.MockAuditRepository)

mockRoleRepo.On("Revoke", ctx, int64(42), "admin-123").Return(nil)
mockRefreshTokenRepo.On("RevokeAllForUser", ctx, "user-123", "client-123").Return(errors.New("cascade failed"))
mockEventRepo.On("Record", ctx, mock.AnythingOfType("*entities.UserEvent")).Return(nil)
mockAuditRepo.On("Create", ctx, mock.AnythingOfType("*entities.AuditLog")).Return(nil)

uc := role.NewRevokeRole(mockRoleRepo, mockRefreshTokenRepo, mockEventRepo, mockAuditRepo)

err := uc.Execute(ctx, role.RevokeRoleRequest{
AssignmentID: 42,
UserID:       "user-123",
ClientID:     "client-123",
TenantID:     "tenant-123",
RevokedBy:    "admin-123",
})

assert.NoError(t, err) // best effort - still succeeds
}
