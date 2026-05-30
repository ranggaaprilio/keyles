# Epic 05: Sessions, Activity Log, Disable, and Delete

## Goal

Give administrators control over user sessions, account status, deletion, and audit visibility.

## Why This Matters

Role management is incomplete without the ability to terminate access immediately and inspect account activity.

## Scope

- Add session endpoints:
  - `GET /api/v1/admin/users/:id/sessions`
  - `DELETE /api/v1/admin/users/:id/sessions/:sessionId`
- Add activity endpoint:
  - `GET /api/v1/admin/users/:id/activity`
- Add account status endpoint:
  - `PATCH /api/v1/admin/users/:id/status`
- Add delete endpoint:
  - `DELETE /api/v1/admin/users/:id`
- Revoke all refresh tokens when disabling or deleting users.
- Add Redis blacklist checks for immediate access-token invalidation.
- Add frontend tabs for sessions and activity.

## Acceptance Criteria

- Disabled users cannot refresh tokens or access protected APIs.
- Deleted users have roles and sessions revoked.
- Admin cannot disable or delete themselves.
- User activity is paginated and scoped to the tenant.
- Session revoke removes the selected active session.

## Suggested First Tasks

- Implement blacklist middleware after auth middleware.
- Add disable user use case with tests.
- Add sessions and activity UI after endpoints are stable.
