package entities

import (
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of audit event
type EventType string

const (
	EventRegistrationAttempt  EventType = "registration_attempt"
	EventRegistrationSuccess  EventType = "registration_success"
	EventRegistrationFailure  EventType = "registration_failure"
	EventOTPGenerated         EventType = "otp_generated"
	EventOTPSent              EventType = "otp_sent"
	EventOTPVerified          EventType = "otp_verified"
	EventOTPExpired           EventType = "otp_expired"
	EventOTPFailed            EventType = "otp_failed"
	EventLoginAttempt         EventType = "login_attempt"
	EventLoginSuccess         EventType = "login_success"
	EventLoginFailure         EventType = "login_failure"
	EventLogout               EventType = "logout"
	EventTenantActivated      EventType = "tenant_activated"
	EventTenantSuspended      EventType = "tenant_suspended"
)

// AuditLog represents security and activity events for compliance and monitoring
type AuditLog struct {
	ID          uuid.UUID
	TenantID    *uuid.UUID
	UserID      *uuid.UUID
	EventType   EventType
	EventData   map[string]interface{} // Additional context stored as JSON
	IPAddress   string
	UserAgent   string
	CreatedAt   time.Time
}

// NewAuditLog creates a new audit log entry
func NewAuditLog(eventType EventType, ipAddress, userAgent string) *AuditLog {
	return &AuditLog{
		ID:        uuid.New(),
		EventType: eventType,
		EventData: make(map[string]interface{}),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	}
}

// WithTenant associates the audit log with a tenant
func (a *AuditLog) WithTenant(tenantID uuid.UUID) *AuditLog {
	a.TenantID = &tenantID
	return a
}

// WithUser associates the audit log with a user
func (a *AuditLog) WithUser(userID uuid.UUID) *AuditLog {
	a.UserID = &userID
	return a
}

// WithData adds custom event data
func (a *AuditLog) WithData(key string, value interface{}) *AuditLog {
	if a.EventData == nil {
		a.EventData = make(map[string]interface{})
	}
	a.EventData[key] = value
	return a
}

// WithError adds error information to event data
func (a *AuditLog) WithError(err error) *AuditLog {
	if err != nil {
		a.WithData("error", err.Error())
	}
	return a
}
