# Epic 06: Frontend Test Coverage & Quality Gates

## Goal

Make the React frontend safer to change as the admin surface grows.

## Why This Matters

The frontend already has Vitest and Testing Library. Expanding user management without strong component tests will make regressions likely.

## Scope

- Add tests for:
  - user list
  - invite dialog
  - user detail
  - role assignment
  - invitation acceptance
  - session and activity tabs
- Add or expand MSW handlers for user-management endpoints.
- Ensure CI runs lint, tests, and build.
- Keep TypeScript API types aligned with backend contracts.

## Acceptance Criteria

- `npm run lint` passes with zero warnings.
- `npm test` passes for existing and new frontend tests.
- `npm run build` passes.
- User-management UI has meaningful happy-path and error-state coverage.

## Suggested First Tasks

- Add MSW mocks for the new user APIs.
- Test `UserList` and `InviteUserDialog` first.
- Add tests alongside each new frontend component.
