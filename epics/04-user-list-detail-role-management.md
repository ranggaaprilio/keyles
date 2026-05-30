# Epic 04: User List, Detail, and Role Management

## Goal

Build the admin APIs and UI for listing users, viewing user details, and assigning app-specific roles.

## Why This Matters

Once users can be invited, administrators need operational control over who has access to each OAuth client.

## Scope

- Add user endpoints:
  - `GET /api/v1/admin/users`
  - `GET /api/v1/admin/users/:id`
  - `PATCH /api/v1/admin/users/:id`
- Add role endpoints:
  - `GET /api/v1/admin/users/:id/roles`
  - `POST /api/v1/admin/users/:id/roles`
  - `DELETE /api/v1/admin/users/:id/roles/:assignmentId`
- Support free-form role names with 1-100 character validation.
- Add frontend user list with search, status filter, and pagination.
- Add user detail page with role assignment and revocation.

## Acceptance Criteria

- Admin can search users by email or display name.
- Admin can filter users by active, pending, or disabled status.
- Roles are scoped per OAuth client.
- Duplicate active role assignment returns conflict.
- New JWTs reflect role changes.

## Suggested First Tasks

- Implement backend list/get user use cases first.
- Update role use cases to support free-form roles.
- Build `UserManagementPage`, then `UserDetail`.
