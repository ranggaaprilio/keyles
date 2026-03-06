package entities

import (
	"time"
)

// UserEventType represents the type of user activity event
type UserEventType string

const (
	// EventTypeLoginSuccess indicates a successful user login
	EventTypeLoginSuccess UserEventType = "login_success"
	// EventTypeLoginFailure indicates a failed login attempt
	EventTypeLoginFailure UserEventType = "login_failure"
	// EventTypeTokenRefresh indicates a token refresh operation
	EventTypeTokenRefresh UserEventType = "token_refresh"
	// EventTypeLogout indicates a user-initiated logout
	EventTypeLogout UserEventType = "logout"
	// EventTypeSessionTerminated indicates an admin-terminated session
	EventTypeSessionTerminated UserEventType = "session_terminated"
	// EventTypeAccountDisabled indicates an admin disabled a user account
	EventTypeAccountDisabled UserEventType = "account_disabled"
	// EventTypeAccountEnabled indicates an admin re-enabled a user account
	EventTypeAccountEnabled UserEventType = "account_enabled"
	// EventTypeRoleAssigned indicates a role was assigned to a user
	EventTypeRoleAssigned UserEventType = "role_assigned"
	// EventTypeRoleRevoked indicates a role was revoked from a user
	EventTypeRoleRevoked UserEventType = "role_revoked"
	// EventTypeUserInvited indicates a user invitation was sent
	EventTypeUserInvited UserEventType = "user_invited"
	// EventTypeInvitationAccepted indicates a user accepted their invitation
	EventTypeInvitationAccepted UserEventType = "invitation_accepted"
	// EventTypeInvitationExpired indicates an invitation expired without being accepted
	EventTypeInvitationExpired UserEventType = "invitation_expired"
	// EventTypeInvitationResent indicates an invitation was resent to a user
	EventTypeInvitationResent UserEventType = "invitation_resent"
	// EventTypeUserDeleted indicates a user was permanently deleted
	EventTypeUserDeleted UserEventType = "user_deleted"
)

// UserEvent represents a single activity event for a user in the audit/activity log
type UserEvent struct {
	ID          int64
	TenantID    string
	UserID      string
	ClientID    *string
	EventType   UserEventType
	IPAddress   *string
	CountryCode *string
	Details     map[string]any
	OccurredAt  time.Time
}
