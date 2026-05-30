# Epic 02: End-User Domain Foundation

## Goal

Add the backend foundation for tenant end users, invitations, user events, and immediate account invalidation.

## Why This Matters

The repo currently has admin users and role assignments, but it does not yet have the full end-user lifecycle described in `specs/005-user-management-rbac/`.

## Scope

- Add migrations `000009` through `000012`.
- Create domain entities: `User`, `Invitation`, and `UserEvent`.
- Add repository interfaces: `EndUserRepository`, `InvitationRepository`, and `UserEventRepository`.
- Extend `RoleRepository` and `RefreshTokenRepository` for user lifecycle operations.
- Add `UserBlacklist` service interface and Redis implementation.
- Add user count cache for tenant quota checks.
- Extend email service with invitation email support.

## Acceptance Criteria

- Migrations run cleanly up and down.
- Domain tests cover validation, status constants, and invitation expiry.
- Domain layer has no infrastructure or HTTP imports.
- New repository interfaces are mockable for use case tests.

## Suggested First Tasks

- Start with the migration files from `specs/005-user-management-rbac/tasks.md`.
- Add domain entities and tests before infrastructure.
- Generate or update mocks after interfaces stabilize.
