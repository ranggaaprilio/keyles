# Implementation Plan: End-User Management with RBAC

**Branch**: `005-user-management-rbac` | **Date**: 2025-06-11 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/005-user-management-rbac/spec.md`

## Summary

Implement a tenant-administrator-facing end-user management portal with role-based access control (RBAC). This feature extends the existing SSO platform with: **user invitation and account lifecycle** (pending → active → disabled), **free-form role assignment per client application**, **JWT `roles` claim injection into access and ID tokens**, **Redis-backed immediate token invalidation** (blacklist on delete/disable), **session management**, **per-user activity log** (90-day retention), and a full React frontend admin dashboard covering user list, detail, invite/accept flows, and role management. Four new migrations (000009–000012) extend the existing `users` and `user_role_assignments` tables and introduce `invitations` and `user_events` tables.

**Critical architectural note**: The existing `AdminUser` entity in `domain/entities/admin_user.go` already maps to the `users` PostgreSQL table. Feature 005 extends that same table and introduces a second domain entity — `User` (`domain/entities/user.go`) — to represent end-users invited by administrators. During implementation, the existing `UserRepository` interface (which serves `AdminUser` for admin authentication) must not be broken; new end-user operations are introduced via a new `EndUserRepository` interface, preventing regression in the admin authentication flow while cleanly separating concerns.

---

## Technical Context

**Language/Version**: Go 1.23 (backend), TypeScript 5.4 (frontend)
**Primary Dependencies**: Gin (HTTP), GORM (ORM), go-redis v9, React 18, TanStack Query v5, Zustand, Axios, Zod, Vite, Tailwind CSS, shadcn/ui (Radix), Brevo (email)
**Storage**: PostgreSQL (users, invitations, role_assignments, user_events, audit_logs) + Redis (user blacklist, tenant user count cache, invitation token prefix index)
**Testing**: Go `testing` + `testify` + GoMock (backend); Vitest + React Testing Library + MSW (frontend)
**Target Platform**: Linux server (Docker), modern browsers
**Project Type**: Web application (backend + frontend)
**Performance Goals**: User list API (10,000 entries) <2s; paginated page <1s; role change reflected in JWT within 1s; blacklist propagation on delete/disable within 1s
**Constraints**: Max 10,000 users per tenant; invitation tokens expire after 72 hours; tokens single-use; free-form role names (1–100 chars, no whitelist); tenant isolation enforced at all layers; admin cannot disable/delete their own account
**Scale/Scope**: Multi-tenant SaaS; each tenant ≤10,000 users ≤25 clients; typical tenant 10–200 users

---

## Constitution Check

_GATE: Must pass before implementation. Re-check after each phase._

**Clean Architecture Compliance**:

- [x] Domain layer has no imports from infrastructure or frameworks
- [x] All repository/service interfaces defined in Domain layer
- [x] Concrete implementations only in Infrastructure layer
- [x] Dependency arrows verified to point inward (toward Domain)

**Verified**: New `User` and `Invitation` entities reside in `domain/entities/` with no external imports. `EndUserRepository`, `InvitationRepository`, `UserEventRepository`, and `UserBlacklist` cache interface reside in domain layer. All concrete implementations (bcrypt, go-redis, GORM queries, Brevo email) remain in `infrastructure/`. Use cases depend only on domain interfaces. Existing `UserRepository` (for `AdminUser`) is untouched. JWT claim injection lives in the `auth/issue_token.go` use case (not in domain entities), satisfying DIP.

**SOLID Principles Compliance**:

- [x] Each module has single, well-defined responsibility (SRP)
- [x] Domain depends only on abstractions/interfaces (DIP)
- [x] No direct database/external API calls from business logic
- [x] Interface segregation verified for all contracts

**Verified**: `InviteUser`, `AcceptInvitation`, `ResendInvitation`, `ListUsers`, `GetUser`, `UpdateUser`, `EnableUser`, `DisableUser`, `DeleteUser`, `ListSessions`, `RevokeSession`, `ListUserActivity` are each single-responsibility use cases. Free-form role validation (1–100 chars) lives in use-case layer, not domain entity, to avoid over-constraining the entity. `EmailService` interface extended with `SendInvitationEmail`; the existing `SendOTPEmail` and `SendWelcomeEmail` methods are unaffected (ISP). `RoleRepository` extended with pagination and revocation-metadata methods without modifying existing method signatures.

**Testing Requirements**:

- [x] Unit test plan documented for all business logic (target: ≥85% coverage)
- [x] Integration test plan for all handlers/controllers
- [x] Test isolation strategy defined (mocking approach)
- [x] Test-first workflow feasible for this feature

**Verified**: GoMock-generated mocks for all four new repository interfaces and `UserBlacklist` cache. Unit tests for all 12 new use cases. Integration tests for all 17 new HTTP endpoints using `httptest` and a test PostgreSQL database. Frontend: Vitest + RTL with MSW service worker for API mocking. Test-first feasible: interfaces defined first, mocks generated, tests written, implementations filled.

**Code Conventions**:

- [x] Backend: Follows Effective Go, lowercase packages, exported function docs
- [x] Frontend: TypeScript strict mode, PascalCase components, functional components only
- [x] Clear separation between backend (domain/usecase/infrastructure) and frontend (components/services)

**Verified**: New Go packages follow existing naming: `user` (use cases), `postgres` (infra), `redis` (infra). All exported types/functions carry godoc comments. React components are PascalCase functional components with hooks. API calls are abstracted in `services/api/user.ts`, not called directly from components.

**Violations Requiring Justification**:

- [x] `UserRoleAssignment.ValidRoles` whitelist **REMOVED** — FR-015 mandates free-form role names. The existing `ValidRoles = []string{"admin", "user", "viewer"}` check in `domain/entities/user_role.go` is removed; validation is replaced by a length check (1–100 chars) in the `AssignRole` use case. **Impact**: existing tests in `assign_role_test.go` and `revoke_role_test.go` that assert on valid/invalid role names must be updated to reflect the new free-form validation logic.
- [x] `GetUserInfo` use case **MODIFIED** — adds `roles` claim to `UserInfoClaims`. Strictly speaking, a new use case would be cleaner, but the single-responsibility principle is still satisfied since `GetUserInfo` still has one job (return user claims for the UserInfo endpoint); it is simply returning more claims. The change is additive and non-breaking.

---

## Architecture Overview

### Clean Architecture Layers for This Feature

```
┌─────────────────────────────────────────────────────────────────────┐
│  Interfaces Layer (interfaces/http/)                                 │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐  │
│  │  UserHandler      │  │  RoleHandler      │  │  SessionHandler  │  │
│  │  InviteHandler    │  │  (extended)       │  │  (extended)      │  │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  Middleware: BlacklistCheckMiddleware (new)                   │   │
│  └──────────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────────┤
│  Use Case Layer (usecase/)                                           │
│  ┌──────────────────────────────┐  ┌──────────────────────────────┐ │
│  │  user/                       │  │  auth/ (modified)            │ │
│  │  invite_user.go              │  │  authorize_client.go → +role │ │
│  │  accept_invitation.go        │  │    check (FR-021)            │ │
│  │  resend_invitation.go        │  │  issue_token.go → +roles     │ │
│  │  list_users.go               │  │    claim (FR-022/024)        │ │
│  │  get_user.go                 │  │  get_userinfo.go → +roles    │ │
│  │  update_user.go              │  │    field (FR-025)            │ │
│  │  enable_user.go              │  └──────────────────────────────┘ │
│  │  disable_user.go             │  ┌──────────────────────────────┐ │
│  │  delete_user.go              │  │  role/ (extended)            │ │
│  │  list_sessions.go            │  │  assign_role.go → free-form  │ │
│  │  revoke_session.go           │  │  revoke_role.go → soft delete│ │
│  │  list_user_activity.go       │  │  list_user_roles.go → paginated│
│  └──────────────────────────────┘  └──────────────────────────────┘ │
├─────────────────────────────────────────────────────────────────────┤
│  Domain Layer (domain/)                                              │
│  Entities: User, Invitation, UserRoleAssignment (extended)           │
│  New entities: UserEvent                                             │
│  Repositories: EndUserRepository, InvitationRepository,             │
│                UserEventRepository, RoleRepository (extended)        │
│  Services: EmailService (extended), UserBlacklist (new)             │
├─────────────────────────────────────────────────────────────────────┤
│  Infrastructure Layer (infrastructure/)                              │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  postgres/                                                   │   │
│  │  end_user_repository.go (NEW)                                │   │
│  │  invitation_repository.go (NEW)                              │   │
│  │  user_event_repository.go (NEW)                              │   │
│  │  role_repository.go (MODIFY: + revoked_at, revoked_by)       │   │
│  └──────────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  redis/                                                      │   │
│  │  user_blacklist.go (NEW)                                     │   │
│  │  user_count_cache.go (NEW)                                   │   │
│  └──────────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  services/                                                   │   │
│  │  brevo_email.go (MODIFY: + SendInvitationEmail)              │   │
│  └──────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

### Key Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| End-user entity naming | `User` in `domain/entities/user.go`, distinct from `AdminUser` | `AdminUser` is the admin portal user; `User` is the OAuth end-user. Same `users` table but different Go types until a future consolidation PR. |
| Repository naming | `EndUserRepository` (not `UserRepository`) | `UserRepository` is already defined and used for `AdminUser` auth. Avoids breaking existing flows; clear intent. |
| Role name validation | Use-case layer length check (1–100 chars) | FR-015: free-form roles. Removes `ValidRoles` whitelist from entity. Validation belongs in use case, not domain. |
| Token invalidation strategy | Redis blacklist `user_blacklist:{user_id}` TTL=900s | Immediate effect without waiting for JWT natural expiry. TTL matches max access token lifetime (FR-037). |
| Invitation token storage | bcrypt hash in DB; 8-char prefix in Redis | Raw token never persisted. Redis prefix cache enables fast existence check before expensive bcrypt comparison (FR-042). |
| Background job approach | `cmd/cleanup/main.go` extended; scheduled via container cron | Reuses existing cleanup command pattern already present in the codebase. |

---

## Database Migration Strategy

**Next migration number**: 000009 (current highest: 000008)

### Migration Dependencies (apply in order)

```
000009_extend_users          ← prerequisite for all others
000010_create_invitations    ← depends on users table
000011_extend_user_role_assignments ← depends on users table
000012_create_user_events    ← independent (denormalized tenant_id)
```

### Schema Changes Summary

| Migration | Table | Type | Key Changes |
|---|---|---|---|
| 000009 | `users` | ALTER | Add `display_name VARCHAR(255)`, `status VARCHAR(20) DEFAULT 'active'` + CHECK constraint, `last_login_at TIMESTAMPTZ`; add trgm + composite indexes |
| 000010 | `invitations` | CREATE | New table: id, tenant_id, email, display_name, token_hash (UNIQUE), status, invited_by (FK→users), expires_at, accepted_at |
| 000011 | `user_role_assignments` | ALTER | Add `revoked_at TIMESTAMPTZ`, `revoked_by VARCHAR(255)` (FK→users); add partial index `WHERE is_active = true` |
| 000012 | `user_events` | CREATE | New table: BIGSERIAL id, tenant_id, user_id, client_id (nullable), event_type + CHECK, ip_address INET, details JSONB, occurred_at |

### Pre-migration Requirements

- Enable `pg_trgm` extension (required by 000009 trigram indexes): `CREATE EXTENSION IF NOT EXISTS pg_trgm;`
- Existing users will default to `status = 'active'` via the `DEFAULT 'active'` clause — no backfill needed.
- Run `ANALYZE users, invitations, user_role_assignments, user_events;` after all migrations to refresh query planner statistics.

### Post-deployment Jobs (scheduled)

| Job | Command | Schedule | Purpose |
|---|---|---|---|
| Expire stale invitations | `InvitationRepository.ExpireStalePending` | Every hour | Mark past-deadline `pending` invitations as `expired` |
| Purge old user events | `UserEventRepository.DeleteOlderThan(90 days)` | Daily at 02:00 UTC | Enforce 90-day retention (FR-034) |

---

## Backend Implementation Phases

### Phase B1 — Domain Layer (foundation; no infrastructure dependencies)

**Deliverables**:

1. **`domain/entities/user.go`** (NEW)
   - `UserStatus` type with `pending`, `active`, `disabled` constants
   - `User` struct with all fields from data-model.md
   - `MaxUsersPerTenant = 10_000`
   - `NewUser()` constructor, `Validate()` method

2. **`domain/entities/invitation.go`** (NEW)
   - `InvitationStatus` type with `pending`, `accepted`, `expired` constants
   - `Invitation` struct; `InvitationTTL = 72 * time.Hour`

3. **`domain/entities/user_event.go`** (NEW)
   - `UserEvent` struct with all 14 `event_type` constants

4. **`domain/entities/user_role.go`** (MODIFY)
   - Remove `ValidRoles` whitelist and `IsValidRole()` method
   - Add `RevokedAt *time.Time`, `RevokedBy *string` fields to `UserRoleAssignment`
   - Add `MaxRoleNameLength = 100` constant
   - Update `Validate()`: replace whitelist check with `len(role) >= 1 && len(role) <= 100`

5. **`domain/repositories/end_user_repository.go`** (NEW)
   - `EndUserRepository` interface with all methods from data-model.md

6. **`domain/repositories/invitation_repository.go`** (NEW)
   - `InvitationRepository` interface

7. **`domain/repositories/user_event_repository.go`** (NEW)
   - `UserEventRepository` interface

8. **`domain/repositories/role_repository.go`** (MODIFY)
   - Add `Assign`, `Revoke`, `ListByUser`, `ListByClient`, `RevokeAllForUser` methods
   - Existing `AssignRole`, `RevokeRole`, `GetUserRoles`, `HasRole`, `HasAnyRole`, `ListRolesByClient`, `ListRolesByUser` signatures unchanged

9. **`domain/services/email_service.go`** (MODIFY)
   - Add `SendInvitationEmail(ctx, toEmail, toName, inviteURL, orgName string) error`

10. **`domain/services/user_blacklist.go`** (NEW)
    - `UserBlacklist` interface: `Add(ctx, userID string, ttl time.Duration) error`, `IsBlacklisted(ctx, userID string) (bool, error)`

### Phase B2 — Infrastructure Layer

**Deliverables**:

1. **`migrations/000009_extend_users.{up,down}.sql`** (NEW) — full SQL from data-model.md
2. **`migrations/000010_create_invitations.{up,down}.sql`** (NEW)
3. **`migrations/000011_extend_user_role_assignments.{up,down}.sql`** (NEW)
4. **`migrations/000012_create_user_events.{up,down}.sql`** (NEW)

5. **`infrastructure/persistence/postgres/end_user_repository.go`** (NEW)
   - GORM model `PostgresUser` extended with `DisplayName`, `Status`, `LastLoginAt`
   - Implements `EndUserRepository`; uses `ILIKE` + trgm for search, `LIMIT/OFFSET` for pagination

6. **`infrastructure/persistence/postgres/invitation_repository.go`** (NEW)
   - GORM model `PostgresInvitation`
   - `GetByToken`: iterates candidates via Redis prefix hint, then bcrypt comparison
   - `ExpireStalePending`: bulk UPDATE where `expires_at < NOW()`

7. **`infrastructure/persistence/postgres/user_event_repository.go`** (NEW)
   - GORM model `PostgresUserEvent`
   - `Record`: fire-and-forget INSERT; errors logged, not propagated to caller

8. **`infrastructure/persistence/postgres/role_repository.go`** (MODIFY)
   - Add `Assign` (INSERT), `Revoke` (UPDATE `is_active=false`, set `revoked_at`, `revoked_by`), `ListByUser`, `ListByClient`, `RevokeAllForUser`
   - Existing methods unchanged

9. **`infrastructure/persistence/redis/user_blacklist.go`** (NEW)
   - `UserBlacklist` implementation using `SET EX 900` / `EXISTS`
   - Key pattern: `user_blacklist:{user_id}`

10. **`infrastructure/persistence/redis/user_count_cache.go`** (NEW)
    - `GET` / `SET EX 60` / `DEL` for key `user_count:{tenant_id}`

11. **`infrastructure/services/brevo_email.go`** (MODIFY)
    - Add `SendInvitationEmail` using existing Brevo client — separate transactional template

### Phase B3 — Use Cases

**Deliverables** (all in `usecase/user/`):

| File | Use Case | Key Logic |
|---|---|---|
| `invite_user.go` | `InviteUser` | Quota check → email uniqueness check → generate token (crypto/rand 32 bytes) → bcrypt hash → save Invitation → enqueue email → log `user_invited` event |
| `accept_invitation.go` | `AcceptInvitation` | Load invitation by token → verify not expired/used → validate password → create `User` (status=active) → mark invitation accepted → log `invitation_accepted` |
| `resend_invitation.go` | `ResendInvitation` | Verify user is `pending` → expire old invitation → create new invitation → send email → log `invitation_resent` |
| `list_users.go` | `ListUsers` | Paginated query with optional search + status filter → return users + total count |
| `get_user.go` | `GetUser` | Fetch user by ID (tenant-scoped) → attach role assignments and active sessions |
| `update_user.go` | `UpdateUser` | Update `display_name` only → log audit event |
| `enable_user.go` | `EnableUser` | Set status=active → log `account_enabled` event + audit log |
| `disable_user.go` | `DisableUser` | Guard: cannot disable self or another admin → set status=disabled → revoke all refresh tokens → add to Redis blacklist → log `account_disabled` event + audit log |
| `delete_user.go` | `DeleteUser` | Guard: cannot delete self → cascade: revoke roles + sessions → add to Redis blacklist → hard delete user row → log `user_deleted` audit entry |
| `list_sessions.go` | `ListSessions` | Query active (non-revoked, non-expired) refresh tokens for user |
| `revoke_session.go` | `RevokeSession` | Revoke single refresh token by ID (tenant-scoped) → log `session_terminated` event |
| `list_user_activity.go` | `ListUserActivity` | Paginated query of `user_events` for user, descending by `occurred_at` |

**Modified use cases**:

| File | Change |
|---|---|
| `usecase/auth/authorize_client.go` | **Before issuing code**: call `RoleRepository.HasAnyRole(userID, clientID)` → reject with `access_denied` if false (FR-021) |
| `usecase/auth/issue_token.go` | **Before signing**: call `RoleRepository.GetActiveRoles(userID, clientID)` → add `roles []string` claim to access token and ID token (FR-022/024) |
| `usecase/auth/get_userinfo.go` | **On response**: fetch roles for the client extracted from the access token → add `roles` field to `UserInfoClaims` (FR-025) |
| `usecase/role/assign_role.go` | Replace whitelist validation with 1–100 char length check; call new `RoleRepository.Assign` method; log `role_assigned` event |
| `usecase/role/revoke_role.go` | Call new `RoleRepository.Revoke` method (soft delete with revokedBy); log `role_revoked` event |

### Phase B4 — HTTP Interfaces

**Deliverables**:

1. **`interfaces/http/handlers/user_handler.go`** (NEW)
   - `UserHandler` struct wiring all 12 user use cases
   - Endpoints: `ListUsers`, `InviteUser`, `GetUser`, `UpdateUser`, `DeleteUser`, `UpdateUserStatus`, `ResendInvitation`

2. **`interfaces/http/handlers/invitation_handler.go`** (NEW)
   - `InvitationHandler` for the public `POST /api/v1/invitations/{token}/accept` endpoint
   - No authentication middleware on this route

3. **`interfaces/http/handlers/role_handler.go`** (MODIFY)
   - Add `ListUserRoles`, extend `AssignRole` (new role path), extend `RevokeRole` (assignment ID path)

4. **`interfaces/http/handlers/session_handler.go`** (MODIFY)
   - Add `ListUserSessions`, `RevokeUserSession` (single session by ID)

5. **`interfaces/http/handlers/activity_handler.go`** (NEW)
   - `ActivityHandler` with `ListUserActivity`

6. **`interfaces/http/middleware/blacklist_check.go`** (NEW)
   - Middleware to call `UserBlacklist.IsBlacklisted(userID)` on every authenticated request
   - Integrates before the admin auth middleware chain

7. **`interfaces/http/router.go`** (MODIFY)
   - Register all 17 new routes under `/api/v1/admin/users/…` and `/api/v1/invitations/…`
   - Wire blacklist middleware

### Phase B5 — Background Jobs

**Deliverables**:

1. **`cmd/cleanup/main.go`** (MODIFY)
   - Add `--expire-invitations` flag: calls `InvitationRepository.ExpireStalePending`
   - Add `--purge-user-events` flag: calls `UserEventRepository.DeleteOlderThan(time.Now().Add(-90 * 24 * time.Hour))`
   - Both operations log count of affected rows

---

## Frontend Implementation Phases

### Phase F1 — Foundation (types, API client, routing)

**Deliverables**:

1. **`src/types/user.ts`** (NEW)
   - `User`, `UserStatus`, `Invitation`, `RoleAssignment`, `UserEvent`, `UserSession` TypeScript types
   - Pagination types: `PaginatedResponse<T>`, `UserListFilters`

2. **`src/services/api/user.ts`** (NEW)
   - `listUsers(filters)`, `inviteUser(req)`, `getUser(id)`, `updateUser(id, req)`, `deleteUser(id)`, `updateUserStatus(id, status)`, `resendInvitation(id)`

3. **`src/services/api/role.ts`** (MODIFY)
   - Add `listUserRoles(userId)`, `assignRole(userId, req)`, `revokeRole(userId, assignmentId)`

4. **`src/services/api/session.ts`** (NEW)
   - `listSessions(userId)`, `revokeSession(userId, sessionId)`, `listUserActivity(userId, page)`

5. **`src/hooks/useUsers.ts`** (NEW)
   - TanStack Query hooks: `useUsers(filters)`, `useUser(id)`, `useInviteUser()`, `useUpdateUser()`, `useDeleteUser()`, `useUpdateUserStatus()`, `useResendInvitation()`

6. **`src/hooks/useRoles.ts`** (NEW)
   - `useUserRoles(userId)`, `useAssignRole()`, `useRevokeRole()`

7. **`src/hooks/useSessions.ts`** (NEW)
   - `useUserSessions(userId)`, `useRevokeSession()`, `useUserActivity(userId, page)`

8. **`src/pages/UserManagementPage.tsx`** (NEW)
   - Route: `/admin/users` — renders `UserList` with header toolbar

9. **`src/pages/AcceptInvitationPage.tsx`** (NEW)
   - Route: `/invite/:token` (public, no auth) — renders password-creation form

### Phase F2 — User List & Search

**Deliverables** (in `src/components/users/`):

1. **`UserList.tsx`** — paginated table with columns: display name, email, status badge, last login, role count, actions menu
2. **`UserListFilters.tsx`** — search input + status tabs (All / Active / Pending / Disabled)
3. **`UserStatusBadge.tsx`** — reusable colour-coded badge component
4. **`InviteUserDialog.tsx`** — modal with email + optional display name fields; shows quota remaining
5. **`DeleteUserDialog.tsx`** — confirmation dialog with destructive action warning

### Phase F3 — User Detail & Role Management

**Deliverables** (in `src/components/users/`):

1. **`UserDetail.tsx`** — profile header + tabbed view (Overview / Roles / Sessions / Activity)
2. **`UserRoles.tsx`** — roles table grouped by client; assign role form with client selector + role name input; revoke role action
3. **`AssignRoleDialog.tsx`** — modal: client dropdown (from existing clients API) + free-form role name input + Zod validation (1–100 chars)
4. **`EnableDisableDialog.tsx`** — confirmation dialog for enable/disable with status-appropriate messaging

### Phase F4 — Invitation Acceptance Flow

**Deliverables**:

1. **`src/components/users/AcceptInvitationForm.tsx`** — read-only email display, password + confirm password fields, password strength indicator, submit button
2. **`src/components/users/InvitationExpired.tsx`** — error state shown when token is expired/used, directs user to contact admin
3. **`src/pages/AcceptInvitationPage.tsx`** — validates token presence in URL, renders form or error state

### Phase F5 — Sessions & Activity

**Deliverables** (in `src/components/users/`):

1. **`UserSessions.tsx`** — active sessions table: client name, created at, last activity, masked origin; "Terminate" action per row
2. **`UserActivity.tsx`** — paginated activity log: event type icon, description, timestamp, client name; infinite scroll or page buttons

---

## Testing Strategy

### Backend — Unit Tests (Go)

Target coverage: **≥85%** across all new files.

| Package | Test File | Mocks Required |
|---|---|---|
| `usecase/user` | `invite_user_test.go` | `EndUserRepository`, `InvitationRepository`, `EmailService`, `UserEventRepository` |
| `usecase/user` | `accept_invitation_test.go` | `EndUserRepository`, `InvitationRepository`, `PasswordService` |
| `usecase/user` | `resend_invitation_test.go` | `EndUserRepository`, `InvitationRepository`, `EmailService` |
| `usecase/user` | `list_users_test.go` | `EndUserRepository` |
| `usecase/user` | `get_user_test.go` | `EndUserRepository`, `RoleRepository` |
| `usecase/user` | `disable_user_test.go` | `EndUserRepository`, `UserBlacklist`, `RefreshTokenRepository`, `UserEventRepository`, `AuditRepository` |
| `usecase/user` | `delete_user_test.go` | `EndUserRepository`, `RoleRepository`, `RefreshTokenRepository`, `UserBlacklist`, `AuditRepository` |
| `usecase/user` | `list_sessions_test.go` | `RefreshTokenRepository` |
| `usecase/user` | `revoke_session_test.go` | `RefreshTokenRepository`, `UserEventRepository` |
| `usecase/auth` | `authorize_client_test.go` (MODIFY) | Add test: user with no roles → `access_denied` |
| `usecase/auth` | `issue_token_test.go` (MODIFY) | Add test: `roles` claim present in token payload |
| `usecase/role` | `assign_role_test.go` (MODIFY) | Update: free-form role validation; add: `role_assigned` event |
| `usecase/role` | `revoke_role_test.go` (MODIFY) | Update: uses `RoleRepository.Revoke`; add: `role_revoked` event |
| `domain` | `user_test.go` (NEW) | None (pure domain logic) |
| `domain` | `invitation_test.go` (NEW) | None |

**Mock generation**: Run `go generate ./...` after each new interface is defined. Mock files live in `backend/tests/mocks/`.

### Backend — Integration Tests

| Test File | Endpoints Covered |
|---|---|
| `tests/integration/user_handler_test.go` (NEW) | All 7 admin user endpoints |
| `tests/integration/invitation_handler_test.go` (NEW) | `POST /api/v1/invitations/{token}/accept` |
| `tests/integration/role_handler_test.go` (MODIFY) | 3 new role endpoints |
| `tests/integration/session_handler_test.go` (MODIFY) | 2 new session endpoints, 1 activity endpoint |

Integration tests use a real PostgreSQL test database (`docker-compose.test.yml`), applied migrations, and `httptest.NewRecorder`. Each test suite runs migrations up in `TestMain` and truncates tables between tests.

### Frontend — Unit Tests (Vitest + RTL)

| Test File | Coverage |
|---|---|
| `tests/unit/components/users/UserList.test.tsx` | Renders, pagination, filter tabs, search debounce |
| `tests/unit/components/users/InviteUserDialog.test.tsx` | Form validation, quota error display |
| `tests/unit/components/users/UserRoles.test.tsx` | Role list render, assign dialog, revoke confirm |
| `tests/unit/components/users/AcceptInvitationForm.test.tsx` | Password validation, submit success, expired token state |
| `tests/unit/hooks/useUsers.test.ts` | Query/mutation hooks with MSW handlers |
| `tests/unit/hooks/useRoles.test.ts` | Assign/revoke hooks |

MSW handlers in `tests/mocks/handlers/user.ts` mirror the OpenAPI contract for all 17 endpoints.

---

## Project Structure

### Documentation (this feature)

```text
specs/005-user-management-rbac/
├── plan.md              # This file
├── data-model.md        # Schema, entities, repository interfaces
├── contracts/
│   ├── openapi.yaml     # OpenAPI 3.0 contract (17 endpoints)
│   └── README.md        # Endpoint summary + JWT claim changes
└── tasks.md             # Phase 2 output (/speckit.tasks — NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
backend/
├── domain/
│   ├── entities/
│   │   ├── user.go                     # NEW: User, UserStatus, MaxUsersPerTenant
│   │   ├── invitation.go               # NEW: Invitation, InvitationStatus, InvitationTTL
│   │   ├── user_event.go               # NEW: UserEvent, event_type constants
│   │   └── user_role.go                # MODIFY: remove ValidRoles whitelist, add RevokedAt/RevokedBy, MaxRoleNameLength
│   ├── repositories/
│   │   ├── end_user_repository.go      # NEW: EndUserRepository interface
│   │   ├── invitation_repository.go    # NEW: InvitationRepository interface
│   │   ├── user_event_repository.go    # NEW: UserEventRepository interface
│   │   └── role_repository.go          # MODIFY: add Assign, Revoke, ListByUser, ListByClient, RevokeAllForUser
│   └── services/
│       ├── email_service.go            # MODIFY: add SendInvitationEmail
│       └── user_blacklist.go           # NEW: UserBlacklist interface
├── usecase/
│   ├── user/
│   │   ├── invite_user.go              # NEW
│   │   ├── accept_invitation.go        # NEW
│   │   ├── resend_invitation.go        # NEW
│   │   ├── list_users.go               # NEW
│   │   ├── get_user.go                 # NEW
│   │   ├── update_user.go              # NEW
│   │   ├── enable_user.go              # NEW
│   │   ├── disable_user.go             # NEW
│   │   ├── delete_user.go              # NEW
│   │   ├── list_sessions.go            # NEW
│   │   ├── revoke_session.go           # NEW
│   │   └── list_user_activity.go       # NEW
│   ├── auth/
│   │   ├── authorize_client.go         # MODIFY: add HasAnyRole check before issuing auth code
│   │   ├── issue_token.go              # MODIFY: add roles claim to access + ID token
│   │   └── get_userinfo.go             # MODIFY: add roles field to UserInfoClaims
│   └── role/
│       ├── assign_role.go              # MODIFY: free-form validation, new Assign method, event log
│       └── revoke_role.go              # MODIFY: soft-delete via new Revoke method, event log
├── infrastructure/
│   ├── persistence/
│   │   ├── postgres/
│   │   │   ├── end_user_repository.go       # NEW: GORM impl of EndUserRepository
│   │   │   ├── invitation_repository.go     # NEW: GORM impl of InvitationRepository
│   │   │   ├── user_event_repository.go     # NEW: GORM impl of UserEventRepository
│   │   │   └── role_repository.go           # MODIFY: implement new methods
│   │   └── redis/
│   │       ├── user_blacklist.go            # NEW: UserBlacklist impl (SET/EXISTS, TTL=900s)
│   │       └── user_count_cache.go          # NEW: user_count:{tenant_id} cache (TTL=60s)
│   └── services/
│       └── brevo_email.go                   # MODIFY: add SendInvitationEmail
├── interfaces/
│   └── http/
│       ├── handlers/
│       │   ├── user_handler.go              # NEW: 7 user CRUD/status endpoints
│       │   ├── invitation_handler.go        # NEW: public accept-invitation endpoint
│       │   ├── activity_handler.go          # NEW: user activity log endpoint
│       │   ├── role_handler.go              # MODIFY: 3 new role endpoints
│       │   └── session_handler.go           # MODIFY: list sessions + individual revoke
│       ├── middleware/
│       │   └── blacklist_check.go           # NEW: Redis blacklist check on authenticated routes
│       └── router.go                        # MODIFY: register 17 new routes + blacklist middleware
├── migrations/
│   ├── 000009_extend_users.up.sql           # NEW
│   ├── 000009_extend_users.down.sql         # NEW
│   ├── 000010_create_invitations.up.sql     # NEW
│   ├── 000010_create_invitations.down.sql   # NEW
│   ├── 000011_extend_user_role_assignments.up.sql  # NEW
│   ├── 000011_extend_user_role_assignments.down.sql # NEW
│   ├── 000012_create_user_events.up.sql     # NEW
│   └── 000012_create_user_events.down.sql   # NEW
├── cmd/
│   └── cleanup/
│       └── main.go                          # MODIFY: add --expire-invitations + --purge-user-events flags
└── tests/
    ├── mocks/
    │   ├── end_user_repository.go            # NEW: GoMock generated
    │   ├── invitation_repository.go          # NEW: GoMock generated
    │   ├── user_event_repository.go          # NEW: GoMock generated
    │   └── user_blacklist.go                 # NEW: GoMock generated
    ├── unit/
    │   ├── domain/
    │   │   ├── user_test.go                  # NEW
    │   │   └── invitation_test.go            # NEW
    │   └── usecase/
    │       ├── invite_user_test.go            # NEW
    │       ├── accept_invitation_test.go      # NEW
    │       ├── resend_invitation_test.go      # NEW
    │       ├── list_users_test.go             # NEW
    │       ├── get_user_test.go               # NEW
    │       ├── disable_user_test.go           # NEW
    │       ├── delete_user_test.go            # NEW
    │       ├── revoke_session_test.go         # NEW
    │       ├── authorize_client_test.go       # MODIFY
    │       ├── issue_token_test.go            # MODIFY
    │       ├── assign_role_test.go            # MODIFY
    │       └── revoke_role_test.go            # MODIFY
    └── integration/
        ├── user_handler_test.go               # NEW
        ├── invitation_handler_test.go         # NEW
        ├── role_handler_test.go               # MODIFY
        └── session_handler_test.go            # MODIFY

frontend/
├── src/
│   ├── types/
│   │   └── user.ts                           # NEW: User, Invitation, RoleAssignment, UserEvent, UserSession types
│   ├── services/
│   │   └── api/
│   │       ├── user.ts                       # NEW: user CRUD + status + invitation API calls
│   │       ├── role.ts                       # MODIFY: add listUserRoles, assignRole, revokeRole
│   │       └── session.ts                    # NEW: listSessions, revokeSession, listUserActivity
│   ├── hooks/
│   │   ├── useUsers.ts                       # NEW: TanStack Query hooks for all user operations
│   │   ├── useRoles.ts                       # NEW: TanStack Query hooks for role CRUD
│   │   └── useSessions.ts                    # NEW: hooks for sessions + activity
│   ├── components/
│   │   └── users/
│   │       ├── UserList.tsx                  # NEW: paginated table with search + filter
│   │       ├── UserListFilters.tsx           # NEW: search input + status tab bar
│   │       ├── UserStatusBadge.tsx           # NEW: colour-coded status badge
│   │       ├── UserDetail.tsx                # NEW: profile header + tabbed detail view
│   │       ├── UserRoles.tsx                 # NEW: role assignments grouped by client
│   │       ├── AssignRoleDialog.tsx          # NEW: modal to assign free-form role
│   │       ├── EnableDisableDialog.tsx       # NEW: confirmation dialog for status change
│   │       ├── InviteUserDialog.tsx          # NEW: invite form with quota display
│   │       ├── DeleteUserDialog.tsx          # NEW: destructive delete confirmation
│   │       ├── AcceptInvitationForm.tsx      # NEW: password-creation form (public)
│   │       ├── InvitationExpired.tsx         # NEW: error state for expired/used tokens
│   │       ├── UserSessions.tsx              # NEW: active sessions table with terminate action
│   │       └── UserActivity.tsx             # NEW: paginated event log
│   └── pages/
│       ├── UserManagementPage.tsx            # NEW: /admin/users route
│       └── AcceptInvitationPage.tsx          # NEW: /invite/:token route (public)
└── tests/
    └── unit/
        ├── components/
        │   └── users/
        │       ├── UserList.test.tsx          # NEW
        │       ├── InviteUserDialog.test.tsx  # NEW
        │       ├── UserRoles.test.tsx         # NEW
        │       └── AcceptInvitationForm.test.tsx # NEW
        └── hooks/
            ├── useUsers.test.ts               # NEW
            └── useRoles.test.ts               # NEW
```

**Structure Decision**: All changes follow the established Clean Architecture layout. No new top-level directories. New packages `usecase/user/` and `components/users/` mirror the existing `usecase/client/` and `components/clients/` patterns introduced in feature 004.

---

## Deployment Considerations

### Environment Variables (additions)

| Variable | Description | Example |
|---|---|---|
| `INVITATION_BASE_URL` | Base URL prepended to invitation tokens for email links | `https://app.keyles.io` |
| `BREVO_INVITATION_TEMPLATE_ID` | Brevo transactional email template ID for invitations | `42` |

### Rollout Sequence

1. **Apply migrations** (000009–000012) via `migrate up` before deploying new application code — the `ADD COLUMN IF NOT EXISTS` and `CREATE TABLE IF NOT EXISTS` clauses are safe to run against the live schema.
2. **Deploy backend** — new routes are additive; existing `/api/v1/admin/roles/*` and `/api/v1/admin/sessions/*` endpoints remain backward-compatible.
3. **Deploy frontend** — new pages are behind authenticated admin routes; no breaking changes to existing pages.
4. **Schedule background jobs** — add cron entries for invitation expiry (hourly) and user event purge (daily).
5. **Verify Redis** — confirm `user_blacklist:*` and `user_count:*` key patterns are not colliding with any existing Redis key namespaces.

### Rollback

- **Backend**: Deploy previous image; migrations are rolled back with `migrate down 4` (removes 000009–000012).
- **Frontend**: Deploy previous build; no localStorage or IndexedDB state is introduced.
- **Data risk**: Rolling back 000009 drops `display_name`, `status`, `last_login_at` columns — no data loss risk for existing `AdminUser` records since those columns are new. Rolling back 000010–000012 drops new tables entirely.

### Performance Validation Checklist (pre-production)

- [ ] `EXPLAIN ANALYZE` on `SELECT * FROM users WHERE tenant_id = $1 AND status = $2` confirms `idx_users_tenant_status` usage
- [ ] `EXPLAIN ANALYZE` on full-text search query confirms trgm index usage (`idx_users_display_name_trgm`, `idx_users_email_trgm`)
- [ ] Redis blacklist check adds <1ms latency per request (benchmark with `redis-benchmark`)
- [ ] User list endpoint with 10,000 rows returns first page in <2s under load (k6 smoke test)
- [ ] `ANALYZE users, invitations, user_role_assignments, user_events;` run after migration

---

## Complexity Tracking

**Justified violations**:

1. **`ValidRoles` whitelist removal** — `domain/entities/user_role.go` — required by FR-015 (free-form roles). Existing tests updated. Risk: LOW (additive change; no existing tenant data uses constrained roles since this is the first user management feature).

2. **`UserRepository` / `EndUserRepository` split** — The existing `UserRepository` interface (serving `AdminUser` for admin authentication) is preserved intact. A separate `EndUserRepository` is introduced for end-user CRUD. This temporary duplication is intentional; a future consolidation sprint can merge the two once admin-user auth is also migrated to the new entity model. Risk: LOW (no cross-contamination; parallel interfaces with no shared methods).

3. **Shared `users` table, two Go entities** — `AdminUser` and `User` both map to the `users` PostgreSQL table. GORM's `TableName()` on both types prevents conflicts as long as they are never used in the same GORM auto-migration call. Migrations are hand-written SQL; GORM auto-migrate is not used in this project. Risk: LOW.
