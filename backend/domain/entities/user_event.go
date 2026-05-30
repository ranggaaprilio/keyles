package entities

import "time"

// UserEventType represents the type of user event
type UserEventType string

const (
	EventTypeLoginSuccess       UserEventType = "login_success"
	EventTypeLoginFailure       UserEventType = "login_failure"
	EventTypeTokenRefresh       UserEventType = "token_refresh"
	EventTypeLogout             UserEventType = "logout"
	EventTypeSessionTerminated  UserEventType = "session_terminated"
	EventTypeAccountDisabled    UserEventType = "account_disabled"
	EventTypeAccountEnabled     UserEventType = "account_enabled"
	EventTypeRoleAssigned       UserEventType = "role_assigned"
	EventTypeRoleRevoked        UserEventType = "role_revoked"
	EventTypeUserInvited        UserEventType = "user_invited"
	EventTypeInvitationAccepted UserEventType = "invitation_accepted"
	EventTypeInvitationExpired  UserEventType = "invitation_expired"
	EventTypeInvitationResent   UserEventType = "invitation_resent"
	EventTypeUserDeleted        UserEventType = "user_deleted"
)

// UserEvent represents an event related to a user within a tenant
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

// TableName specifies the table name for GORM
func (UserEvent) TableName() string {
	return "user_events"
}
