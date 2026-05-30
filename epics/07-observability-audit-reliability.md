# Epic 07: Observability, Audit, and Admin Reliability

## Goal

Improve production readiness with logs, audit visibility, cleanup jobs, and API documentation.

## Why This Matters

SSO systems need traceability. Admins should know what happened, and operators should be able to troubleshoot without leaking sensitive data.

## Scope

- Add structured backend logging for key use cases and handlers.
- Send frontend `ErrorBoundary` failures to an error monitoring service.
- Add tenant admin audit log viewer.
- Extend `cmd/cleanup` to expire invitations and purge old user events.
- Document cleanup cron schedules.
- Update OpenAPI docs for OAuth, client management, and user-management endpoints.
- Review error responses for sensitive data leakage.

## Acceptance Criteria

- Security-sensitive actions create audit entries.
- Cleanup command can expire stale invitations.
- Old user events can be purged by retention policy.
- Error logs are useful without exposing secrets or stack traces.
- API docs match implemented routes.

## Suggested First Tasks

- Add cleanup flags for invitation expiry and user-event retention.
- Add audit viewer after user activity endpoints exist.
- Document all new environment variables.
