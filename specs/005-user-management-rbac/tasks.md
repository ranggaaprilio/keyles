# Tasks: End-User Management with RBAC

**Input**: Design documents from `/specs/005-user-management-rbac/`
**Prerequisites**: plan.md ✓, spec.md ✓, data-model.md ✓
**Feature Branch**: `005-user-management-rbac`

**Tests**: Per constitution, tests are MANDATORY. All business logic requires unit tests (≥85% coverage), and all handlers require integration tests.

**Organization**: Tasks follow the plan's B1–B5 (backend) and F1–F5 (frontend) phases. Within each use-case and frontend phase, tasks are grouped by user story with tests written first. B1 and B2 are fully foundational — no user story work can begin until they are complete.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no inter-task dependencies)
- **[Story]**: Which user story this task belongs to — US1 through US7 (setup/foundational/polish phases have no story label)
- Include exact file paths in descriptions

## Path Conventions

- **Backend (Clean Architecture)**: `backend/domain/`, `backend/usecase/`, `backend/infrastructure/`, `backend/interfaces/`
- **Frontend**: `frontend/src/components/`, `frontend/src/services/`, `frontend/src/hooks/`, `frontend/src/pages/`
- **Tests**: `backend/tests/unit/`, `backend/tests/integration/`, `backend/tests/mocks/`, `frontend/tests/unit/`

## User Story Index

| ID  | Story                                          | Priority |
| --- | ---------------------------------------------- | -------- |
| US1 | Administrator Invites a New User               | P1       |
| US2 | Administrator Assigns and Revokes Roles        | P1       |
| US3 | Administrator Views User List & Search/Filter  | P1       |
| US4 | Administrator Views Activity & Active Sessions | P2       |
| US5 | Administrator Enables and Disables a User      | P2       |
| US6 | Administrator Removes a User from the Tenant   | P3       |
| US7 | JWT Access Token Includes Role Claims (RBAC)   | P1       |

---

## Phase 1: Setup — Database Migrations

**Purpose**: Four sequential SQL migrations that all subsequent phases depend on. Apply in order: 000009 → 000010 → 000011 → 000012.

**⚠️ PREREQUISITE**: Enable `pg_trgm` extension before running 000009: `CREATE EXTENSION IF NOT EXISTS pg_trgm;`

- [ ] T001 Create migration files `backend/migrations/000009_extend_users.up.sql` and `backend/migrations/000009_extend_users.down.sql` per data-model.md §000009: add `display_name VARCHAR(255)`, `status VARCHAR(20) NOT NULL DEFAULT 'active'` + `users_valid_status` CHECK constraint, `last_login_at TIMESTAMPTZ`; create indexes `idx_users_status`, `idx_users_tenant_status`, `idx_users_tenant_email_lower`, `idx_users_display_name_trgm`, `idx_users_email_trgm`

- [ ] T002 Create migration files `backend/migrations/000010_create_invitations.up.sql` and `backend/migrations/000010_create_invitations.down.sql` per data-model.md §000010: new `invitations` table with columns id, tenant_id (FK→tenants), email, display_name, token_hash (UNIQUE), status + CHECK constraint, invited_by (FK→users), expires_at, accepted_at, timestamps; create indexes `idx_invitations_tenant`, `idx_invitations_email`, `idx_invitations_status`, `idx_invitations_expires_at`; add `update_updated_at_column()` trigger

- [ ] T003 Create migration files `backend/migrations/000011_extend_user_role_assignments.up.sql` and `backend/migrations/000011_extend_user_role_assignments.down.sql` per data-model.md §000011: add `revoked_at TIMESTAMPTZ` and `revoked_by VARCHAR(255) REFERENCES users(id)` columns to `user_role_assignments`; create partial index `idx_ura_user_client_active ON user_role_assignments(user_id, client_id) WHERE is_active = true`

- [ ] T004 Create migration files `backend/migrations/000012_create_user_events.up.sql` and `backend/migrations/000012_create_user_events.down.sql` per data-model.md §000012: new `user_events` table with BIGSERIAL id, tenant_id, user_id, client_id (nullable), event_type + 14-value CHECK constraint, ip_address INET, country_code CHAR(2), details JSONB, occurred_at; create indexes `idx_user_events_user_recent`, `idx_user_events_tenant_type`, `idx_user_events_occurred_at`

---

## Phase 2: B1 — Domain Layer

**Purpose**: All new and modified domain entities, repository interfaces, and service interfaces. No infrastructure imports. Must be complete before Phase 3.

**⚠️ CRITICAL**: All repository interfaces must be defined here before any Infrastructure or Use Case code is written.

### New Domain Entities

- [ ] T005 Create domain entity `backend/domain/entities/user.go`: define `UserStatus` type with constants `UserStatusPending`, `UserStatusActive`, `UserStatusDisabled`; define `User` struct with fields ID, TenantID, Email, DisplayName, PasswordHash, Status, LastLoginAt *time.Time, CreatedAt, UpdatedAt; add `MaxUsersPerTenant = 10_000` constant; add `NewUser()` constructor and `Validate()` method (email format, status enum, display name max 255 chars)

- [ ] T006 [P] Create domain entity `backend/domain/entities/invitation.go`: define `InvitationStatus` type with constants `InvitationStatusPending`, `InvitationStatusAccepted`, `InvitationStatusExpired`; define `Invitation` struct (ID, TenantID, Email, DisplayName, TokenHash, Status, InvitedBy, ExpiresAt, AcceptedAt *time.Time, CreatedAt, UpdatedAt); add `InvitationTTL = 72 * time.Hour` constant; add `IsExpired() bool` and `IsAccepted() bool` helper methods

- [ ] T007 [P] Create domain entity `backend/domain/entities/user_event.go`: define `UserEventType` type; add all 14 event-type constants: `EventTypeLoginSuccess`, `EventTypeLoginFailure`, `EventTypeTokenRefresh`, `EventTypeLogout`, `EventTypeSessionTerminated`, `EventTypeAccountDisabled`, `EventTypeAccountEnabled`, `EventTypeRoleAssigned`, `EventTypeRoleRevoked`, `EventTypeUserInvited`, `EventTypeInvitationAccepted`, `EventTypeInvitationExpired`, `EventTypeInvitationResent`, `EventTypeUserDeleted`; define `UserEvent` struct with all fields from data-model.md (ID int64, TenantID, UserID, ClientID *string, EventType, IPAddress *string, CountryCode *string, Details map[string]any, OccurredAt)

- [ ] T008 Modify domain entity `backend/domain/entities/user_role.go`: remove `ValidRoles` var and `IsValidRole()` method; add `RevokedAt *time.Time` and `RevokedBy *string` fields to `UserRoleAssignment` struct; add `MaxRoleNameLength = 100` constant; update `Validate()` to replace whitelist check with length validation: `len(role) >= 1 && len(role) <= MaxRoleNameLength`; update all existing tests in `backend/tests/unit/usecase/assign_role_test.go` and `backend/tests/unit/usecase/revoke_role_test.go` that assert on old valid/invalid role names

### New Repository Interfaces

- [ ] T009 Create repository interface `backend/domain/repositories/end_user_repository.go`: define `EndUserRepository` interface with methods `GetByID(ctx, userID string) (*entities.User, error)`, `GetByEmail(ctx, tenantID, email string) (*entities.User, error)`, `Create(ctx, user *entities.User) error`, `Update(ctx, user *entities.User) error`, `ListByTenant(ctx, tenantID, search string, status entities.UserStatus, page, pageSize int) ([]*entities.User, int, error)`, `CountByTenant(ctx, tenantID string) (int, error)`, `UpdateStatus(ctx, userID string, status entities.UserStatus) error`, `UpdateLastLogin(ctx, userID string, at time.Time) error`, `Delete(ctx, userID string) error`

- [ ] T010 [P] Create repository interface `backend/domain/repositories/invitation_repository.go`: define `InvitationRepository` interface with methods `Create(ctx, inv *entities.Invitation) error`, `GetByToken(ctx, plainToken string) (*entities.Invitation, error)`, `GetPendingByEmail(ctx, tenantID, email string) (*entities.Invitation, error)`, `UpdateStatus(ctx, invitationID string, status entities.InvitationStatus, acceptedAt *time.Time) error`, `ListByTenant(ctx, tenantID string, page, pageSize int) ([]*entities.Invitation, int, error)`, `ExpireStalePending(ctx context.Context) (int64, error)`

- [ ] T011 [P] Create repository interface `backend/domain/repositories/user_event_repository.go`: define `UserEventRepository` interface with methods `Record(ctx, event *entities.UserEvent) error`, `ListByUser(ctx, userID string, page, pageSize int) ([]*entities.UserEvent, int, error)`, `DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)`

### Modified Repository Interfaces

- [ ] T012 Modify repository interface `backend/domain/repositories/role_repository.go`: add new methods `Assign(ctx, assignment *entities.UserRoleAssignment) error`, `Revoke(ctx, assignmentID int64, revokedByUserID string) error`, `ListByUser(ctx, userID string) ([]*entities.UserRoleAssignment, error)`, `ListByClient(ctx, clientID string, page, pageSize int) ([]*entities.UserRoleAssignment, int, error)`, `RevokeAllForUser(ctx, userID, revokedByUserID string) error`; preserve all existing method signatures unchanged (`AssignRole`, `RevokeRole`, `GetUserRoles`, `HasRole`, `HasAnyRole`, `ListRolesByClient`, `ListRolesByUser`)

- [ ] T013 [P] Modify repository interface `backend/domain/repositories/refresh_token_repository.go`: add `RevokeByUserID(ctx context.Context, userID string) error` method to support revoking all sessions for a user across all client applications (needed by disable_user and delete_user use cases); preserve all existing method signatures unchanged

### New Service Interfaces

- [ ] T014 [P] Create service interface `backend/domain/services/user_blacklist.go`: define `UserBlacklist` interface with methods `Add(ctx context.Context, userID string, ttl time.Duration) error` and `IsBlacklisted(ctx context.Context, userID string) (bool, error)`; add godoc comment explaining the Redis key pattern `user_blacklist:{user_id}` and 900s TTL purpose

- [ ] T015 [P] Modify service interface `backend/domain/services/email_service.go`: add `SendInvitationEmail(ctx context.Context, toEmail, toName, inviteURL, orgName string) error` method to the `EmailService` interface; preserve all existing method signatures (`SendOTPEmail`, `SendWelcomeEmail`) unchanged

### Domain Layer Tests

- [ ] T016 [P] Create unit tests `backend/tests/unit/domain/user_test.go`: test `User.Validate()` (valid user passes, empty email fails, empty tenantID fails, invalid status fails, display name >255 chars fails), test `NewUser()` sets correct defaults, test `UserStatus` constants
- [ ] T017 [P] Create unit tests `backend/tests/unit/domain/invitation_test.go`: test `Invitation.IsExpired()` (past expiry returns true, future expiry returns false), test `Invitation.IsAccepted()`, test `InvitationStatus` constants, test `InvitationTTL = 72h`

**Clean Architecture Checkpoint**:
- Domain layer has no imports from `infrastructure/`, `interfaces/`, or external frameworks ✓
- All repository and service interfaces defined in domain layer ✓
- `UserRepository` (serving `AdminUser`) is untouched ✓

**Checkpoint**: Domain layer established. Proceed to Infrastructure.

---

## Phase 3: B2 — Infrastructure Layer

**Purpose**: SQL migrations applied, all new Postgres GORM implementations, Redis implementations, Brevo email extension, and GoMock mocks. Must be complete before any use cases are written.

**⚠️ CRITICAL**: Apply migrations (T001–T004) against your development database before writing infrastructure code.

### Postgres Implementations

- [ ] T018 Create `backend/infrastructure/persistence/postgres/end_user_repository.go`: define `PostgresUser` GORM model (extends existing `users` table) with `DisplayName *string`, `Status string`, `LastLoginAt *time.Time`, `IsActive bool` fields and `TableName() = "users"`; implement all `EndUserRepository` methods: `Create` (INSERT), `GetByID` and `GetByEmail` (SELECT by PK/tenant+email), `ListByTenant` (SELECT with `ILIKE` search on display_name/email using trgm indexes + status filter + LIMIT/OFFSET pagination), `CountByTenant` (SELECT COUNT), `UpdateStatus` (UPDATE status WHERE id), `UpdateLastLogin` (UPDATE last_login_at WHERE id), `Update` (UPDATE display_name WHERE id), `Delete` (DELETE WHERE id)

- [ ] T019 [P] Create `backend/infrastructure/persistence/postgres/invitation_repository.go`: define `PostgresInvitation` GORM model with all fields from data-model.md and `TableName() = "invitations"`; implement `InvitationRepository`: `Create` (INSERT + store 8-char prefix in Redis key `invitation_exists:{prefix}` with 72h TTL), `GetByToken` (check Redis prefix index first → iterate candidates → bcrypt comparison → return ErrNotFound if no match), `GetPendingByEmail` (SELECT WHERE tenant_id + email + status='pending'), `UpdateStatus` (UPDATE status + accepted_at), `ListByTenant` (paginated SELECT), `ExpireStalePending` (bulk UPDATE WHERE expires_at < NOW() AND status='pending')

- [ ] T020 [P] Create `backend/infrastructure/persistence/postgres/user_event_repository.go`: define `PostgresUserEvent` GORM model with all fields from data-model.md using `datatypes.JSON` for details and `TableName() = "user_events"`; implement `UserEventRepository`: `Record` (fire-and-forget INSERT — errors are logged but not propagated to caller), `ListByUser` (SELECT WHERE user_id ORDER BY occurred_at DESC with LIMIT/OFFSET, returns total count), `DeleteOlderThan` (DELETE WHERE occurred_at < before)

- [ ] T021 Modify `backend/infrastructure/persistence/postgres/role_repository.go`: add implementations for all 5 new `RoleRepository` methods: `Assign` (INSERT with duplicate-key error handling → return ErrDuplicateRole if `(user_id, client_id, role, is_active=true)` already exists), `Revoke` (UPDATE is_active=false, revoked_at=NOW(), revoked_by=revokedByUserID WHERE id), `ListByUser` (SELECT all assignments for userID including inactive), `ListByClient` (SELECT active assignments for clientID with pagination), `RevokeAllForUser` (UPDATE is_active=false, revoked_at=NOW(), revoked_by=revokedByUserID WHERE user_id + is_active=true); preserve all existing method implementations unchanged

- [ ] T022 [P] Modify `backend/infrastructure/persistence/postgres/refresh_token_repository.go` (or `refresh_token_repository_gorm.go`): implement new `RevokeByUserID(ctx, userID string) error` method using `UPDATE refresh_tokens SET is_revoked=true WHERE user_id = ?`; add appropriate index hint comment referencing `idx_refresh_tokens_user`

### Redis Implementations

- [ ] T023 [P] Create `backend/infrastructure/persistence/redis/user_blacklist.go`: implement `UserBlacklist` interface using go-redis v9; `Add(ctx, userID, ttl)` → `SET user_blacklist:{userID} "1" EX 900`; `IsBlacklisted(ctx, userID)` → `EXISTS user_blacklist:{userID}` returns bool; include godoc and key-pattern constants

- [ ] T024 [P] Create `backend/infrastructure/persistence/redis/user_count_cache.go`: implement a `UserCountCache` struct with `Get(ctx, tenantID string) (int, bool, error)` → `GET user_count:{tenantID}` (returns value + cache-hit bool), `Set(ctx, tenantID string, count int) error` → `SET user_count:{tenantID} {count} EX 60`, `Invalidate(ctx, tenantID string) error` → `DEL user_count:{tenantID}`; expose the struct (not an interface — it is a concrete cache utility used directly by use cases)

### Email Service

- [ ] T025 [P] Modify `backend/infrastructure/services/brevo_email.go`: add `SendInvitationEmail(ctx, toEmail, toName, inviteURL, orgName string) error` using the existing Brevo client and a new transactional template ID from env var `BREVO_INVITATION_TEMPLATE_ID`; read `INVITATION_BASE_URL` env var and prepend to token when building inviteURL; preserve all existing `SendOTPEmail` and `SendWelcomeEmail` implementations

### Mock Generation

- [ ] T026 Generate GoMock mocks for all new interfaces — run `go generate ./...` after adding `//go:generate mockgen` directives on each new interface file:
  - `backend/tests/mocks/end_user_repository.go` (mock for `EndUserRepository`)
  - `backend/tests/mocks/invitation_repository.go` (mock for `InvitationRepository`)
  - `backend/tests/mocks/user_event_repository.go` (mock for `UserEventRepository`)
  - `backend/tests/mocks/user_blacklist.go` (mock for `UserBlacklist`)
  - Update `backend/tests/mocks/role_repository.go` (regenerate with new `Assign`, `Revoke`, `ListByUser`, `ListByClient`, `RevokeAllForUser` methods)
  - Update `backend/tests/mocks/refresh_token_repository.go` (regenerate with new `RevokeByUserID` method)

**Checkpoint**: Infrastructure complete. Mocks available. Use case phase can begin.

---

## Phase 4: B3 — Use Cases

**Purpose**: All new and modified use cases, organized by user story. Tests are written **FIRST** and must FAIL before implementation begins.

**⚠️ TEST-FIRST RULE**: For each user story sub-section below — write and commit the test file(s) first, confirm they fail (compilation errors or assertion failures), then implement the use case(s).

---

### User Story 7 — JWT Role Claims (Priority: P1) 🎯 Implement First

**Goal**: Access tokens and ID tokens include a `roles` claim scoped to the authenticating client. Authorization is denied before any code is issued if the user has no active role for the requested client.

**Why first**: US7 is a cross-cutting authentication concern. It modifies the core OAuth token flow (`authorize_client`, `issue_token`, `get_userinfo`) that all other stories depend on. Implementing this first validates the RBAC data model integration path before any UI is built.

**Independent Test (no UI needed)**: Seed `user_role_assignments` directly in the DB, run an OAuth authorization code flow, decode the JWT, assert `roles` claim contains the seeded role names.

#### Tests for US7 (MANDATORY) ⚠️

> **Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T027 [P] [US7] Modify unit test `backend/tests/unit/usecase/authorize_client_test.go`: add test case "user with no active role for client → returns access_denied error"; add test case "user with active role(s) → authorization succeeds and code is returned"; use mock `RoleRepository.HasAnyRole` returning false/true respectively

- [ ] T028 [P] [US7] Modify unit test `backend/tests/unit/usecase/issue_token_test.go` (or create alongside): add test case "issued access token JWT contains `roles` claim with active role names for the user+client pair"; add test case "roles claim contains only roles for the authenticating client, not other clients"; use mock `RoleRepository.GetActiveRoles` returning test role slices; verify token payload parsing

#### Implementation for US7

- [ ] T029 [US7] Modify `backend/usecase/auth/authorize_client.go`: inject `RoleRepository repositories.RoleRepository` as a new dependency in `AuthorizeClient` struct; before generating the authorization code, call `roleRepo.HasAnyRole(ctx, req.UserID, req.ClientID)` — if false, return `&OAuthError{Code: ErrAccessDenied, Description: "user has no active role for this application"}` (FR-021); update `NewAuthorizeClient` constructor accordingly

- [ ] T030 [US7] Modify `backend/usecase/auth/issue_token.go`: inject `RoleRepository repositories.RoleRepository` dependency in `IssueToken` struct; before signing the JWT, call `roleRepo.GetActiveRoles(ctx, userID, clientID)` to fetch `[]string` of active role names; add `roles` field (array of strings) to both the access token JWT claims and the ID token JWT claims (FR-022, FR-024); update `NewIssueToken` constructor; if `GetActiveRoles` returns an empty slice, include an empty array `[]` in the claim (not null)

- [ ] T031 [US7] Modify `backend/usecase/auth/get_userinfo.go`: after fetching user info, extract `clientID` from the access token claims; call `roleRepo.GetActiveRoles(ctx, userID, clientID)` and add `Roles []string` field to the `UserInfoClaims` response struct (FR-025); update `NewGetUserInfo` constructor with `RoleRepository` dependency

---

### User Story 1 — User Invitation (Priority: P1)

**Goal**: Administrators can invite new users via email; invited users click a time-limited link, set a password, and their account becomes active.

**Independent Test**: POST `/api/v1/admin/users/invite` with email + display name → invitation record created, email dispatched, token 72h TTL. POST `/api/v1/invitations/{token}/accept` with valid token + password → user status transitions `pending→active`, token invalidated. Expired token → 410 response.

#### Tests for US1 (MANDATORY) ⚠️

> **Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T032 [P] [US1] Create unit test `backend/tests/unit/usecase/invite_user_test.go`: test quota enforcement (CountByTenant at 10,000 → reject); test email-already-exists rejection; test happy path (token generated, bcrypt hashed, invitation created, email sent, `user_invited` event recorded); test pending invitation already exists for same email → reject with conflict; mocks: `EndUserRepository`, `InvitationRepository`, `EmailService`, `UserEventRepository`, `UserCountCache`

- [ ] T033 [P] [US1] Create unit test `backend/tests/unit/usecase/accept_invitation_test.go`: test valid token + valid password → user created (status=active), invitation marked accepted, token invalidated; test expired invitation → error; test already-accepted invitation → error; test weak password rejection; mocks: `EndUserRepository`, `InvitationRepository`

- [ ] T034 [P] [US1] Create unit test `backend/tests/unit/usecase/resend_invitation_test.go`: test user in `pending` status → old invitation expired, new invitation created, email sent; test user NOT in `pending` status → error; test new token has fresh 72h expiry; mocks: `EndUserRepository`, `InvitationRepository`, `EmailService`

#### Implementation for US1

- [ ] T035 [US1] Create `backend/usecase/user/invite_user.go`: implement `InviteUser` use case; quota check via `EndUserRepository.CountByTenant` (with `UserCountCache` — read cache first, fall back to DB, write-through); email uniqueness check via `EndUserRepository.GetByEmail`; generate 32-byte cryptographically secure token via `crypto/rand`; bcrypt-hash the token; create `Invitation` record via `InvitationRepository.Create`; call `EmailService.SendInvitationEmail`; record `user_invited` event via `UserEventRepository.Record`; create a `pending` `User` record; invalidate `UserCountCache` on success

- [ ] T036 [US1] Create `backend/usecase/user/accept_invitation.go`: implement `AcceptInvitation` use case; call `InvitationRepository.GetByToken(plainToken)` (bcrypt comparison); verify invitation is `pending` and not expired; validate password strength (min 8 chars, mix of cases and digits as per existing password validation); hash password with bcrypt cost≥12; call `EndUserRepository.Update` to set `PasswordHash` and `Status=active`; call `InvitationRepository.UpdateStatus(accepted, now)`; record `invitation_accepted` event

- [ ] T037 [US1] Create `backend/usecase/user/resend_invitation.go`: implement `ResendInvitation` use case; verify target user exists and has `pending` status (reject otherwise); expire current pending invitation via `InvitationRepository.UpdateStatus(expired, nil)`; generate new token (same crypto/rand approach); create new `Invitation` with fresh `ExpiresAt = now + 72h`; call `EmailService.SendInvitationEmail`; record `invitation_resent` event

---

### User Story 3 — User List & Search (Priority: P1)

**Goal**: Administrators see a paginated, searchable, filterable list of all users in their tenant.

**Independent Test**: Seed 30+ users in various statuses. GET `/api/v1/admin/users?page=1&page_size=25` → 25 users, total=30+, pagination metadata. GET with `?search=alice` → filtered results. GET with `?status=pending` → only pending users.

#### Tests for US3 (MANDATORY) ⚠️

- [ ] T038 [P] [US3] Create unit test `backend/tests/unit/usecase/list_users_test.go`: test default pagination (page=1, pageSize=25); test custom page/pageSize; test search filter passed to repository; test status filter; test page/pageSize clamping (max 100); test tenant isolation (tenantID from caller context); mocks: `EndUserRepository`

- [ ] T039 [P] [US3] Create unit test `backend/tests/unit/usecase/get_user_test.go`: test happy path returns user with role assignments and active sessions; test cross-tenant access → returns not-found error (tenant isolation); mocks: `EndUserRepository`, `RoleRepository`

#### Implementation for US3

- [ ] T040 [US3] Create `backend/usecase/user/list_users.go`: implement `ListUsers` use case; accept `ListUsersInput` (TenantID, Search string, StatusFilter UserStatus, Page int, PageSize int); validate/clamp pagination (default page=1, pageSize=25, max pageSize=100); call `EndUserRepository.ListByTenant`; return `ListUsersOutput` with users slice + total count + pagination metadata (total pages)

- [ ] T041 [US3] Create `backend/usecase/user/get_user.go`: implement `GetUser` use case; fetch user by ID scoped to caller's tenantID (return ErrNotFound if tenantID mismatch); fetch all role assignments via `RoleRepository.ListByUser`; fetch active sessions count via `RefreshTokenRepository` (or include raw count in response); return a `GetUserOutput` with user + role assignments grouped by clientID

- [ ] T042 [US3] Create `backend/usecase/user/update_user.go`: implement `UpdateUser` use case; accept `UpdateUserInput` (UserID, TenantID, DisplayName string); validate tenantID scoping; call `EndUserRepository.Update` with new display name; record `audit_log` entry with event type `user_updated`; this use case only allows updating `display_name` — status changes go through `EnableUser`/`DisableUser`

---

### User Story 2 — Role Management (Priority: P1)

**Goal**: Administrators assign free-form role names to users per client application; roles are immediately reflected in new JWT tokens.

**Independent Test**: POST `/api/v1/admin/users/{id}/roles` with clientID + roleName → assignment created, `role_assigned` event logged. Duplicate → 409. GET new JWT → `roles` claim contains new role. DELETE assignment → `role_revoked` event, next JWT omits role.

#### Tests for US2 (MANDATORY) ⚠️

- [ ] T043 [P] [US2] Modify unit test `backend/tests/unit/usecase/assign_role_test.go`: update existing tests to use new free-form validation (1–100 chars) instead of whitelist; add test: empty role name → error; add test: role name > 100 chars → error; add test: duplicate active assignment → ErrDuplicateRole; add test: `role_assigned` user event recorded; add test: cross-tenant assignment rejected; verify `RoleRepository.Assign` (not legacy `AssignRole`) is called

- [ ] T044 [P] [US2] Modify unit test `backend/tests/unit/usecase/revoke_role_test.go`: update existing tests to verify `RoleRepository.Revoke(assignmentID, revokedByUserID)` (soft delete) is called instead of `RevokeRole`; add test: `role_revoked` user event recorded with acting admin ID; add test: revoking non-existent assignment → not-found error

#### Implementation for US2

- [ ] T045 [US2] Modify `backend/usecase/role/assign_role.go`: replace `tempAssignment.IsValidRole()` whitelist check with length validation `len(req.Role) >= 1 && len(req.Role) <= entities.MaxRoleNameLength`; replace `roleRepo.AssignRole` call with `roleRepo.Assign` (new method with full assignment object); inject `UserEventRepository` dependency; add `UserEventRepository.Record` call for `EventTypeRoleAssigned` event after successful assignment; add `AuditLogRepository` entry; update constructor

- [ ] T046 [US2] Modify `backend/usecase/role/revoke_role.go`: update to call `roleRepo.Revoke(assignmentID int64, revokedByUserID string)` (soft delete with metadata) instead of `roleRepo.RevokeRole`; inject `UserEventRepository` dependency; add `UserEventRepository.Record` call for `EventTypeRoleRevoked` event; add `AuditLogRepository` entry; update constructor; function signature accepts `assignmentID int64` and `revokedBy string`

---

### User Story 4 — Sessions & Activity Log (Priority: P2)

**Goal**: Administrators view active sessions and paginated activity events per user; can terminate individual sessions.

**Independent Test**: Authenticate user twice across two clients. GET `/api/v1/admin/users/{id}/sessions` → 2 sessions with client name, created/last-activity times. DELETE `/api/v1/admin/users/{id}/sessions/{sessionId}` → 204, refresh token revoked, session gone from list. GET `/api/v1/admin/users/{id}/activity?page=1` → login events paginated.

#### Tests for US4 (MANDATORY) ⚠️

- [ ] T047 [P] [US4] Create unit test `backend/tests/unit/usecase/list_sessions_test.go`: test returns only non-revoked, non-expired refresh tokens for a given userID; test tenant isolation (sessions from another tenant not returned); test empty result returns empty slice (not nil); mocks: `RefreshTokenRepository`

- [ ] T048 [P] [US4] Create unit test `backend/tests/unit/usecase/revoke_session_test.go`: test happy path: `RefreshTokenRepository.Revoke` called with correct tokenHash + revokedBy; test `session_terminated` event recorded; test cross-tenant session revocation → error; test already-revoked session → error; mocks: `RefreshTokenRepository`, `UserEventRepository`

#### Implementation for US4

- [ ] T049 [US4] Create `backend/usecase/user/list_sessions.go`: implement `ListSessions` use case; fetch all non-revoked, non-expired refresh tokens for `userID` from `RefreshTokenRepository` (scoped to caller's tenantID); map tokens to `SessionOutput` structs (token ID, clientID, createdAt, lastUsedAt, expiresAt); return slice

- [ ] T050 [US4] Create `backend/usecase/user/revoke_session.go`: implement `RevokeSession` use case; look up refresh token by ID and verify it belongs to the target userID within the caller's tenant; call `RefreshTokenRepository.Revoke(tokenHash, adminUserID)`; record `EventTypeSessionTerminated` event via `UserEventRepository.Record`

- [ ] T051 [US4] Create `backend/usecase/user/list_user_activity.go`: implement `ListUserActivity` use case; call `UserEventRepository.ListByUser(userID, page, pageSize)` with default pageSize=25, max=100; verify tenantID scoping; return events slice + total count + pagination metadata

---

### User Story 5 — Enable & Disable User (Priority: P2)

**Goal**: Administrators can disable a user (immediately revoking all sessions and blocking new auth) and re-enable them without a new invitation.

**Independent Test**: Authenticate user → obtain refresh token. POST `/api/v1/admin/users/{id}/status` with `{"status":"disabled"}` → user disabled, refresh token revoked, Redis blacklist entry set. Attempt token refresh → `invalid_grant`. POST `{"status":"active"}` → user re-enabled, new auth flow succeeds.

#### Tests for US5 (MANDATORY) ⚠️

- [ ] T052 [P] [US5] Create unit test `backend/tests/unit/usecase/disable_user_test.go`: test admin disabling self → error "cannot disable your own account"; test admin disabling another admin → error "cannot disable an administrator account"; test happy path: status set to disabled, `RevokeByUserID` called, `UserBlacklist.Add` called with 900s TTL, `account_disabled` event recorded, audit log entry created; mocks: `EndUserRepository`, `RefreshTokenRepository`, `UserBlacklist`, `UserEventRepository`, `AuditRepository`

#### Implementation for US5

- [ ] T053 [US5] Create `backend/usecase/user/disable_user.go`: implement `DisableUser` use case; guard: reject if `req.TargetUserID == req.AdminUserID` (cannot disable self); guard: fetch target user, reject if target is also a tenant admin (check user role via `RoleRepository` or admin flag); call `EndUserRepository.UpdateStatus(userID, disabled)`; call `RefreshTokenRepository.RevokeByUserID(userID)` to revoke all active sessions; call `UserBlacklist.Add(userID, 900*time.Second)` for immediate token invalidation; record `EventTypeAccountDisabled` event; create audit log entry

- [ ] T054 [US5] Create `backend/usecase/user/enable_user.go`: implement `EnableUser` use case (simpler counterpart to DisableUser); call `EndUserRepository.UpdateStatus(userID, active)`; record `EventTypeAccountEnabled` event; create audit log entry; note: previously revoked sessions are NOT restored — user must re-authenticate to create a new session

---

### User Story 6 — Delete User (Priority: P3)

**Goal**: Administrators permanently remove a user, cascading all role assignments, sessions, and pending invitations; deleted user's access tokens are immediately invalidated.

**Independent Test**: Create user with role assignments and active session. DELETE `/api/v1/admin/users/{id}` → 204, user gone from list, credentials rejected at login, refresh token revoked, Redis blacklist set, audit log entry written.

#### Tests for US6 (MANDATORY) ⚠️

- [ ] T055 [P] [US6] Create unit test `backend/tests/unit/usecase/delete_user_test.go`: test admin deleting self → error "cannot delete your own account"; test cascade order: `RoleRepository.RevokeAllForUser` called, then `RefreshTokenRepository.RevokeByUserID` called, then `UserBlacklist.Add` called, then `EndUserRepository.Delete` called — all in same logical transaction; test `user_deleted` audit log entry written with deleted user's email + ID; mocks: `EndUserRepository`, `RoleRepository`, `RefreshTokenRepository`, `UserBlacklist`, `AuditRepository`

#### Implementation for US6

- [ ] T056 [US6] Create `backend/usecase/user/delete_user.go`: implement `DeleteUser` use case; guard: reject if `req.TargetUserID == req.AdminUserID` (cannot delete self); soft-cascade: call `RoleRepository.RevokeAllForUser(userID, adminUserID)` to soft-delete all role assignments; call `RefreshTokenRepository.RevokeByUserID(userID)` to revoke all sessions; call `UserBlacklist.Add(userID, 900*time.Second)` for access-token blacklist (FR-037); call `EndUserRepository.Delete(userID)` (hard delete — cascade in DB handles invitations); write `user_deleted` audit log entry with deleted user's email, ID, and acting admin identity

**Checkpoint**: All 12 use cases implemented and unit-tested. Auth token flow (US7) enhanced. Proceed to HTTP interfaces.

---

## Phase 5: B4 — HTTP Interfaces

**Purpose**: All new and modified HTTP handlers, blacklist middleware, router registration, and DI wiring. Integration tests written first.

### Middleware

- [ ] T057 Create `backend/interfaces/http/middleware/blacklist_check.go`: implement `BlacklistCheckMiddleware` as a Gin middleware; on every authenticated request, extract `userID` from JWT claims context; call `UserBlacklist.IsBlacklisted(ctx, userID)` → if true, abort with 401 JSON `{"error": "token_invalid", "error_description": "account has been revoked"}` (covers deleted and disabled users); inject `UserBlacklist` dependency; middleware must execute before (or immediately after) the existing auth middleware parses the token

### Integration Tests (write first)

- [ ] T058 [P] Create integration test `backend/tests/integration/user_handler_test.go`: test all 7 admin user endpoints: `POST /api/v1/admin/users/invite` (201 pending user + invitation, 409 duplicate email, 409 quota exceeded, 400 bad email), `GET /api/v1/admin/users` (200 paginated list, search, status filter), `GET /api/v1/admin/users/{id}` (200 user detail with roles, 404 not found, 403 cross-tenant), `PATCH /api/v1/admin/users/{id}` (200 updated display name), `PATCH /api/v1/admin/users/{id}/status` (200 enabled/disabled, 400 cannot disable self, 400 cannot disable admin), `DELETE /api/v1/admin/users/{id}` (204 deleted, 400 cannot delete self), `POST /api/v1/admin/users/{id}/resend-invitation` (200 new invitation sent, 400 user not pending)

- [ ] T059 [P] Create integration test `backend/tests/integration/invitation_handler_test.go`: test public `POST /api/v1/invitations/{token}/accept` endpoint: 200 with valid token + strong password, 410 with expired token, 410 with already-used token, 400 with weak password, 400 with missing fields — **no auth middleware on this route**

- [ ] T060 [P] Modify integration test `backend/tests/integration/role_handler_test.go`: add tests for 3 new role endpoints: `GET /api/v1/admin/users/{userId}/roles` (200 paginated list grouped by client, 404 user not found), `POST /api/v1/admin/users/{userId}/roles` (201 assignment created, 409 duplicate, 400 role name too long, 400 empty role name), `DELETE /api/v1/admin/users/{userId}/roles/{assignmentId}` (204 revoked, 404 not found, 403 cross-tenant)

- [ ] T061 [P] Modify integration test `backend/tests/integration/session_handler_test.go`: add tests for 3 new session/activity endpoints: `GET /api/v1/admin/users/{userId}/sessions` (200 active sessions list, empty state), `DELETE /api/v1/admin/users/{userId}/sessions/{sessionId}` (204 revoked, 404 not found), `GET /api/v1/admin/users/{userId}/activity` (200 paginated activity events, 25 per page default, 404 user not found)

### New Handlers

- [ ] T062 Create `backend/interfaces/http/handlers/user_handler.go`: implement `UserHandler` struct wiring all 7 admin user operations (InviteUser, GetUser, ListUsers, UpdateUser, UpdateUserStatus, DeleteUser, ResendInvitation); implement Gin handler methods: `InviteUser` (parse email + display_name, call InviteUser UC, return 201 with user stub), `GetUser` (return user + roles + active session count), `ListUsers` (parse page/pageSize/search/status query params), `UpdateUser` (PATCH display_name only), `UpdateUserStatus` (PATCH status field — routes to EnableUser or DisableUser UC based on value), `DeleteUser` (DELETE → 204), `ResendInvitation` (POST → 200 confirmation); all endpoints extract tenantID and adminUserID from JWT claims context

- [ ] T063 Create `backend/interfaces/http/handlers/invitation_handler.go`: implement `InvitationHandler` with single public `AcceptInvitation` handler method; parse `{token}` path param and `{"password": "..."}` JSON body; call `AcceptInvitation` UC; on success return 200 `{"message": "Account activated successfully"}`; on expired/used token return 410 Gone with `{"error": "invitation_expired", "error_description": "..."}`; **this route must have NO auth middleware**

- [ ] T064 Create `backend/interfaces/http/handlers/activity_handler.go`: implement `ActivityHandler` with `ListUserActivity` handler method; parse `{userId}` path param + `page`, `page_size` query params; call `ListUserActivity` UC; return 200 with paginated events; enforce tenantID scoping from JWT claims

- [ ] T065 Modify `backend/interfaces/http/handlers/role_handler.go`: add `ListUserRoles` handler (GET roles for a user, grouped by clientID); extend `AssignRole` to use new path pattern `/users/{userId}/roles` and call updated `AssignRole` UC; extend `RevokeRole` to use path `/users/{userId}/roles/{assignmentId}` (int64 ID, not legacy string triple) and call updated `RevokeRole` UC; preserve all existing role handler methods and paths

- [ ] T066 Modify `backend/interfaces/http/handlers/session_handler.go`: add `ListUserSessions` handler (GET active sessions for a user); add `RevokeUserSession` handler (DELETE single session by ID with confirmation); preserve all existing session handler methods

### Router & Wiring

- [ ] T067 Modify `backend/interfaces/http/router.go`: register all 17 new routes; wire `BlacklistCheckMiddleware` into the authenticated middleware chain (before admin route group); admin user routes under `/api/v1/admin/users/`: GET / (list), POST /invite, GET /{id}, PATCH /{id}, PATCH /{id}/status, DELETE /{id}, POST /{id}/resend-invitation, GET /{id}/roles, POST /{id}/roles, DELETE /{id}/roles/{assignmentId}, GET /{id}/sessions, DELETE /{id}/sessions/{sessionId}, GET /{id}/activity; public route: POST `/api/v1/invitations/{token}/accept` (no auth middleware)

- [ ] T068 Update DI wiring in `backend/cmd/server/main.go`: instantiate `EndUserRepository`, `InvitationRepository`, `UserEventRepository`, `UserBlacklist` (Redis), `UserCountCache` (Redis); instantiate all 12 new user use cases with their dependencies; inject updated `RoleRepository` implementation and `RefreshTokenRepository` (with `RevokeByUserID`) into `DisableUser`, `DeleteUser`; inject `RoleRepository` into `AuthorizeClient` and `IssueToken`; wire `UserHandler`, `InvitationHandler`, `ActivityHandler` into router

**Checkpoint**: All 17 endpoints wired and integration-tested. End-to-end backend flow from admin request to JWT claims working.

---

## Phase 6: B5 — Background Jobs

**Purpose**: Scheduled cleanup jobs for invitation expiry and user-event retention. Extends the existing `cmd/cleanup` pattern.

- [ ] T069 Modify `backend/cmd/cleanup/main.go`: add `--expire-invitations` CLI flag that calls `InvitationRepository.ExpireStalePending(ctx)` and logs the count of affected rows; add `--purge-user-events` CLI flag that calls `UserEventRepository.DeleteOlderThan(ctx, time.Now().Add(-90*24*time.Hour))` and logs the count of deleted rows; update the help text and README/docs comment to document both new flags and suggested cron schedules (expire-invitations: every hour; purge-user-events: daily at 02:00 UTC)

---

## Phase 7: F1 — Frontend Foundation

**Purpose**: All TypeScript types, API service functions, TanStack Query hooks, pages, and route wiring that F2–F5 depend on.

**⚠️ CRITICAL**: F2–F5 components depend on these hooks and types. Complete this phase before component work begins.

- [ ] T070 Create `frontend/src/types/user.ts`: define TypeScript types: `UserStatus` (union `'pending' | 'active' | 'disabled'`), `User` (id, tenantId, email, displayName, passwordHash excluded, status, lastLoginAt, createdAt, updatedAt, roleCount), `Invitation` (id, tenantId, email, displayName, status, expiresAt, createdAt), `RoleAssignment` (id, userId, clientId, clientName, roleName, isActive, grantedAt, grantedBy, revokedAt, revokedBy), `UserEvent` (id, eventType, clientId, clientName, ipAddress, countryCode, details, occurredAt), `UserSession` (id, clientId, clientName, createdAt, lastUsedAt, expiresAt); add `PaginatedResponse<T>` with total/page/pageSize/totalPages; add `UserListFilters` (search, status, page, pageSize); add `InviteUserRequest`, `AcceptInvitationRequest`, `AssignRoleRequest` request types

- [ ] T071 [P] Create `frontend/src/services/api/user.ts`: implement typed API functions: `listUsers(filters: UserListFilters): Promise<PaginatedResponse<User>>`, `inviteUser(req: InviteUserRequest): Promise<User>`, `getUser(id: string): Promise<User>`, `updateUser(id: string, req: {displayName: string}): Promise<User>`, `deleteUser(id: string): Promise<void>`, `updateUserStatus(id: string, status: UserStatus): Promise<User>`, `resendInvitation(userId: string): Promise<void>`; all use the existing Axios base client

- [ ] T072 [P] Modify `frontend/src/services/api/role.ts` (or create if not exists): add `listUserRoles(userId: string): Promise<RoleAssignment[]>`, `assignRole(userId: string, req: AssignRoleRequest): Promise<RoleAssignment>`, `revokeRole(userId: string, assignmentId: number): Promise<void>`; preserve any existing role API functions

- [ ] T073 [P] Create `frontend/src/services/api/session.ts`: implement `listSessions(userId: string): Promise<UserSession[]>`, `revokeSession(userId: string, sessionId: string): Promise<void>`, `listUserActivity(userId: string, page: number, pageSize?: number): Promise<PaginatedResponse<UserEvent>>`

- [ ] T074 Create `frontend/src/hooks/useUsers.ts`: implement TanStack Query v5 hooks: `useUsers(filters: UserListFilters)` (query with staleTime=30s), `useUser(id: string)` (query), `useInviteUser()` (mutation that invalidates `['users']` on success), `useUpdateUser()` (mutation), `useDeleteUser()` (mutation that invalidates `['users']`), `useUpdateUserStatus()` (mutation that invalidates specific `['user', id]`), `useResendInvitation()` (mutation)

- [ ] T075 [P] Create `frontend/src/hooks/useRoles.ts`: implement `useUserRoles(userId: string)` (query), `useAssignRole()` (mutation that invalidates `['user', userId, 'roles']`), `useRevokeRole()` (mutation that invalidates `['user', userId, 'roles']`)

- [ ] T076 [P] Create `frontend/src/hooks/useSessions.ts`: implement `useUserSessions(userId: string)` (query with staleTime=15s), `useRevokeSession()` (mutation that invalidates `['user', userId, 'sessions']`), `useUserActivity(userId: string, page: number)` (query)

- [ ] T077 Create `frontend/src/pages/UserManagementPage.tsx` and wire admin routes in `frontend/src/App.tsx`: add `/admin/users` route (renders `UserManagementPage`) and `/admin/users/:id` route (renders user detail) both wrapped in `ProtectedRoute` with admin role check; add public route `/invite/:token` (renders `AcceptInvitationPage`); `UserManagementPage` renders `UserList` with a header toolbar "Invite User" button

- [ ] T078 [P] Create `frontend/src/pages/AcceptInvitationPage.tsx`: public page (no auth); extracts `token` from URL params; renders `AcceptInvitationForm` or `InvitationExpired` based on whether token is present and valid; handles `useAcceptInvitation` mutation; on success redirect to login page with success message

**Clean Architecture Checkpoint** (Frontend):
- API calls abstracted in `services/api/*.ts`, not called from components ✓
- Components receive data via hooks, not direct API calls ✓
- TypeScript strict mode enforced ✓

---

## Phase 8: F2 — User List & Search (US3)

**Goal**: Administrators see the full paginated, searchable, filterable user dashboard.

**Independent Test**: Render `UserManagementPage` with MSW mocking `GET /api/v1/admin/users` → list renders with status badges, pagination controls, search input. Apply filter → filtered results show. Click invite → dialog opens.

### Tests for F2 (MANDATORY) ⚠️

- [ ] T079 [P] [US3] Create frontend unit test `frontend/tests/unit/components/users/UserList.test.tsx`: test renders user rows with correct columns (name, email, status badge, last login, role count); test search input calls `useUsers` with updated search param after debounce; test status tab switch updates filter; test pagination next/prev buttons; test empty state renders when no users; test loading skeleton shown while fetching; use MSW to mock list endpoint

- [ ] T080 [P] [US3] Create frontend unit test `frontend/tests/unit/components/users/InviteUserDialog.test.tsx`: test form renders with email (required) and display name (optional) fields; test submit with valid email → calls `useInviteUser` mutation; test submit with invalid email format → shows validation error; test quota-exceeded API error → shows quota message; test dialog closes on success

### Implementation for F2

- [ ] T081 [P] [US3] Create `frontend/src/components/users/UserStatusBadge.tsx`: reusable colour-coded badge; `active` → green, `pending` → yellow, `disabled` → grey/red; accepts `status: UserStatus` prop; use shadcn/ui `Badge` variant; export as named component

- [ ] T082 [P] [US3] Create `frontend/src/components/users/UserListFilters.tsx`: search input with 300ms debounce (update URL search params); status tab bar with four tabs: All / Active / Pending / Disabled; emits `onFiltersChange(filters: UserListFilters)` callback; controlled component accepting current filter values as props

- [ ] T083 [P] [US3] Create `frontend/src/components/users/InviteUserDialog.tsx`: modal dialog (shadcn/ui `Dialog`); react-hook-form + Zod validation (email: RFC 5322 format, displayName: optional max 255 chars); shows remaining quota `{10000 - currentCount} slots remaining`; uses `useInviteUser()` mutation; shows loading state on submit; closes and shows success toast on completion; shows API error inline on failure

- [ ] T084 [P] [US3] Create `frontend/src/components/users/DeleteUserDialog.tsx`: confirmation dialog with destructive styling; displays `"This will permanently delete {email} and all their role assignments. This action cannot be undone."`; require the admin to type the user's email to confirm; uses `useDeleteUser()` mutation; redirects to user list on success

- [ ] T085 [US3] Create `frontend/src/components/users/UserList.tsx`: paginated table using `useUsers(filters)` hook; columns: display name (link to detail), email, `UserStatusBadge`, last login (formatted relative time), role count, actions menu (Edit, Disable/Enable, Delete); integrate `UserListFilters` for search + tab filter; pagination controls (prev/next/page numbers); loading skeleton (10 rows); empty state component; "Invite User" button opens `InviteUserDialog`; clicking row name navigates to `/admin/users/:id`

**Checkpoint**: User list dashboard fully functional. Admins can see, search, filter, invite, and delete users.

---

## Phase 9: F3 — User Detail & Role Management (US2)

**Goal**: Administrators view a user's full profile with tabbed role management and controls.

**Independent Test**: Render `UserDetail` for a user with 3 role assignments across 2 clients → roles grouped by client visible. Click "Assign Role" → dialog opens, submit → role appears. Click revoke → role removed with confirmation. Click "Disable Account" → confirmation, then status badge updates.

### Tests for F3 (MANDATORY) ⚠️

- [ ] T086 [P] [US2] Create frontend unit test `frontend/tests/unit/components/users/UserRoles.test.tsx`: test roles render grouped by client application name; test "Assign Role" button opens `AssignRoleDialog`; test assigning a role calls `useAssignRole` mutation and updates list; test clicking revoke on a role shows confirmation and calls `useRevokeRole`; test empty state when no roles assigned

### Implementation for F3

- [ ] T087 [P] [US2] Create `frontend/src/components/users/AssignRoleDialog.tsx`: modal dialog; client dropdown populated from existing clients API (`useClients` hook); free-form role name text input; Zod validation (1–100 chars, no leading/trailing whitespace); uses `useAssignRole()` mutation; shows duplicate-role API error inline; closes on success and shows toast

- [ ] T088 [P] [US2] Create `frontend/src/components/users/EnableDisableDialog.tsx`: confirmation dialog; messaging adapts to current status: disable → "This will immediately terminate all active sessions for {email}."; enable → "This will restore {email}'s ability to authenticate."; uses `useUpdateUserStatus()` mutation; shows loading state; closes and refreshes user detail on success

- [ ] T089 [P] [US2] Create `frontend/src/components/users/UserRoles.tsx`: role assignments table grouped by client application name (accordion or section headers); columns: role name, granted at, granted by, revoke button; "Assign Role" button at top-right opens `AssignRoleDialog`; revoke confirmation inline (popover or inline confirm); uses `useUserRoles(userId)` and `useRevokeRole()` hooks; empty state: "No roles assigned. Assign a role to grant access to a client application."

- [ ] T090 [US2] Create `frontend/src/components/users/UserDetail.tsx`: profile header (avatar initials, display name, email, `UserStatusBadge`, last login); action buttons toolbar: "Edit Display Name" (inline edit), "Disable / Enable Account" (opens `EnableDisableDialog`), "Delete User" (opens `DeleteUserDialog`); tabbed content (shadcn/ui `Tabs`): **Roles** tab renders `UserRoles`, **Sessions** tab renders `UserSessions`, **Activity** tab renders `UserActivity`; uses `useUser(id)` hook; back navigation to user list

**Checkpoint**: User detail page with role management working end-to-end.

---

## Phase 10: F4 — Invitation Acceptance Flow (US1)

**Goal**: Invited users can click their email link, set a password, and activate their account.

**Independent Test**: Navigate to `/invite/{validToken}` → form renders with read-only email, password fields. Submit strong password → account activated, redirect to login. Navigate with expired token → `InvitationExpired` error state shown.

### Tests for F4 (MANDATORY) ⚠️

- [ ] T091 [P] [US1] Create frontend unit test `frontend/tests/unit/components/users/AcceptInvitationForm.test.tsx`: test form renders with read-only email field pre-populated; test password strength indicator updates as user types; test mismatched confirm password shows validation error; test password < 8 chars shows validation error; test successful submit calls `acceptInvitation` API and redirects to login; test API 410 expired error renders `InvitationExpired` component

### Implementation for F4

- [ ] T092 [P] [US1] Create `frontend/src/components/users/InvitationExpired.tsx`: error state component shown when invitation token is expired or already used; displays clear message: "This invitation link has expired or has already been used."; provides guidance: "Please contact your administrator to request a new invitation."; styled with appropriate error icon; no interactive controls needed (read-only state)

- [ ] T093 [P] [US1] Create `frontend/src/components/users/AcceptInvitationForm.tsx`: form with read-only email address display (extracted from invitation token pre-fetch); password input with strength indicator (visual bar: weak/fair/strong); confirm password input with match validation; submit button disabled until passwords match and meet minimum strength; uses `useAcceptInvitation()` mutation from a new `useInvitation` hook wrapping `POST /api/v1/invitations/{token}/accept`; on 200: redirect to login with `?activated=true` query param for success toast; on 410: render `InvitationExpired`; on 400: display inline field errors

- [ ] T094 [US1] Add `useAcceptInvitation` mutation hook to `frontend/src/hooks/useSessions.ts` or create `frontend/src/hooks/useInvitation.ts`: wrap `POST /api/v1/invitations/{token}/accept`; no auth required; handle 200 (success), 410 (expired/used), 400 (validation error) response shapes; update `AcceptInvitationPage.tsx` to use this hook and pass token from URL params

**Checkpoint**: Invitation acceptance flow complete. Full US1 story testable end-to-end from invite email to activated account.

---

## Phase 11: F5 — Sessions & Activity (US4, US5)

**Goal**: Administrators see active sessions and can terminate them; activity log provides audit trail per user.

**Independent Test**: Authenticate user via two different clients. Open Sessions tab → 2 rows visible with client names and last-activity times. Click "Terminate" → session removed. Open Activity tab → login events listed in reverse-chronological order with pagination.

- [ ] T095 [P] [US4] Create `frontend/src/components/users/UserSessions.tsx`: active sessions table using `useUserSessions(userId)` hook; columns: client application name, session created at, last activity (formatted relative time), masked origin (IP prefix or country code if available); "Terminate" button per row (with inline confirmation popover — no full modal needed); uses `useRevokeSession()` mutation (invalidates sessions query on success); empty state: "No active sessions for this user."; loading skeleton (3 rows)

- [ ] T096 [P] [US4] Create `frontend/src/components/users/UserActivity.tsx`: paginated activity log using `useUserActivity(userId, page)` hook; each row: event type icon (colour-coded by severity: login_failure = red, session_terminated = orange, others = grey), event type label (human-readable), client application name (if applicable), formatted timestamp (relative + absolute on hover), IP address / country code; pagination controls (prev/next); empty state: "No activity recorded for this user."; loading skeleton (5 rows); default 25 events per page

**Checkpoint**: Full session management and activity log working. US4 and US5 fully supported.

---

## Phase 12: Polish & Cross-Cutting Concerns

**Purpose**: Coverage verification, architecture compliance, frontend hook tests, performance validation.

- [ ] T097 [P] Create frontend unit tests `frontend/tests/unit/hooks/useUsers.test.ts` and `frontend/tests/unit/hooks/useRoles.test.ts`: test `useUsers` query hook with MSW handlers (correct URL params, paginated response shape); test `useInviteUser` mutation (optimistic updates, cache invalidation of `['users']`); test `useUserRoles` query; test `useAssignRole` mutation (cache invalidation); test `useRevokeRole` mutation; use `renderHook` from React Testing Library

- [ ] T098 [P] Verify ≥85% unit test coverage on all new backend packages — run `go test -coverprofile=coverage.out ./backend/...` and `go tool cover -func=coverage.out`; check coverage for: `domain/entities/` (user.go, invitation.go, user_event.go, user_role.go), `usecase/user/` (all 12 use cases), `usecase/auth/` (authorize_client.go, issue_token.go, get_userinfo.go), `usecase/role/` (assign_role.go, revoke_role.go); file a gap ticket for any package below 85%

- [ ] T099 [P] Architecture compliance review — verify: (1) no `infrastructure/` or `interfaces/` imports from `domain/`; (2) no `infrastructure/` imports from `usecase/`; (3) `UserRepository` (AdminUser auth) is entirely untouched (run `git diff backend/domain/repositories/user_repository.go` and `backend/infrastructure/persistence/postgres/user_repository.go`); (4) all new `usecase/user/` files import only `domain/` packages; use `go list -f '{{.Imports}}' ./...` to audit

- [ ] T100 [P] Performance validation — run `EXPLAIN ANALYZE` on key query patterns: `ListByTenant` with trgm search (confirm `idx_users_display_name_trgm` used), `GetActiveRoles` (confirm `idx_ura_user_client_active` partial index used), `ListByUser` on user_events (confirm `idx_user_events_user_recent` used); benchmark Redis blacklist check latency with `redis-benchmark`; run `ANALYZE users, invitations, user_role_assignments, user_events;` after all migrations

- [ ] T101 [P] Add GoDoc comments on all exported types and functions in new files: `backend/domain/entities/user.go`, `backend/domain/entities/invitation.go`, `backend/domain/entities/user_event.go`, `backend/domain/repositories/end_user_repository.go`, `backend/domain/repositories/invitation_repository.go`, `backend/domain/repositories/user_event_repository.go`, `backend/domain/services/user_blacklist.go`; verify all JSON field names in handler responses use snake_case consistently; verify error messages match OpenAPI contract examples in `specs/005-user-management-rbac/contracts/openapi.yaml`

- [ ] T102 [P] Document deployment additions in project README or ops runbook: add `INVITATION_BASE_URL` and `BREVO_INVITATION_TEMPLATE_ID` to environment variable table; document cron entries: `0 * * * * ./cleanup --expire-invitations` and `0 2 * * * ./cleanup --purge-user-events`; document Redis key namespaces (`user_blacklist:*`, `user_count:*`, `invitation_exists:*`); document rollback procedure (migrate down 4 removes 000009–000012)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — apply migrations to development DB immediately
- **B1 Domain (Phase 2)**: No code dependencies — write interfaces first (enablers for all other phases)
- **B2 Infrastructure (Phase 3)**: Depends on Phase 1 (migrations must exist) + Phase 2 (interfaces must be defined)
- **B3 Use Cases (Phase 4)**: Depends on Phase 2 (interfaces) + Phase 3 (mocks generated)
  - US7 FIRST (modifies core auth flow; validates role data access path)
  - US1, US2, US3 are all P1 — implement sequentially: US7 → US3 → US1 → US2
  - US4, US5 are P2 — implement after P1 stories complete
  - US6 is P3 — implement last
- **B4 HTTP Interfaces (Phase 5)**: Depends on Phase 4 (use cases must compile)
- **B5 Background Jobs (Phase 6)**: Depends on Phase 3 (repositories) — can be done at any point after Phase 3
- **F1 Frontend Foundation (Phase 7)**: Can begin in parallel with Phase 3 (types/mocks need no backend); API service calls can be implemented against OpenAPI contract
- **F2–F5 Frontend (Phases 8–11)**: All depend on Phase 7; can begin once Phase 5 backend endpoints are running
- **Polish (Phase 12)**: Depends on all previous phases

### User Story Dependencies

| Story | Depends On | Notes |
| ----- | ---------- | ----- |
| US7 (JWT Claims) | B1, B2 only | Implement first; enables RBAC enforcement in auth flow |
| US3 (User List) | B1, B2 only | No inter-story deps; read-only query path |
| US1 (Invite) | US3 (list users helps test invitation), B1, B2 | Invitation acceptance is fully independent (public endpoint) |
| US2 (Roles) | US1 (need active users to assign roles to), US7 (roles → JWT) | Frontend role management lives in UserDetail from US3 |
| US4 (Sessions) | US1 (need sessions to exist), B1, B2 | UI lives in UserDetail tab from US3 |
| US5 (Enable/Disable) | US1 (need active users), B1, B2 | Simple status change; reuses infrastructure from US4 (RevokeByUserID) |
| US6 (Delete) | US1, US2, US4, US5 | Cascade operation; all sub-systems must exist first |

### Within Each User Story (use case phase)

1. Tests MUST be written and FAIL before implementation
2. Domain entities before repository methods
3. Repository interfaces before infrastructure implementations
4. Use case logic before handler (domain → outward)
5. Integration tests before handler implementation
6. Backend before frontend (API must exist for frontend to consume)

### Parallel Opportunities

**Phase 2 (B1 Domain)**:
```
Parallel batch 1: T006, T007, T008 (different entity files, after T005 entity is defined)
Parallel batch 2: T010, T011, T013, T014, T015 (different interface files, after T009 EndUserRepository defined)
Sequential: T005 (User entity, depended on by all repos), T009 (EndUserRepository), T012 (RoleRepository extends existing)
Parallel: T016, T017 (domain unit tests, independent of each other)
```

**Phase 3 (B2 Infrastructure)**:
```
Sequential: T018 (EndUserRepository, core GORM model)
Parallel batch 1: T019, T020, T022, T023, T024, T025 (different files, all depend on T018 model pattern)
Sequential: T021 (extends existing role_repository.go), T026 (mocks, must be last — depends on all interfaces)
```

**Phase 4 (B3 Use Cases)**:
```
US7: T027, T028 parallel (test files) → T029 → T030 → T031 (sequential, each depends on previous)
US1: T032, T033, T034 parallel (test files) → T035 → T036 → T037 (sequential use cases)
US3: T038, T039 parallel (test files) → T040 → T041 → T042 (sequential)
US2: T043, T044 parallel (test files) → T045 → T046 (sequential)
US4: T047, T048 parallel (test files) → T049 → T050 → T051 (sequential)
US5: T052 (test) → T053 → T054
US6: T055 (test) → T056
```

**Phase 5 (B4 HTTP)**:
```
T057 (middleware, no deps)
Parallel: T058, T059, T060, T061 (integration tests for different handlers)
Sequential: T062 → T063, T064, T065, T066 (handlers after integration tests)
Sequential: T067 (router, depends on all handlers) → T068 (DI, depends on router)
```

**Frontend (Phases 7–11)**:
```
Phase 7: T070 first (types) → T071, T072, T073 parallel → T074 (useUsers, after types+services) → T075, T076 parallel → T077, T078 parallel
Phase 8: T079, T080 parallel (tests) → T081, T082, T083, T084 parallel → T085
Phase 9: T086 (test) → T087, T088, T089 parallel → T090
Phase 10: T091 (test) → T092, T093 parallel → T094
Phase 11: T095, T096 parallel
```

---

## Parallel Execution Example: User Story 1 (Invitation)

```bash
# Step 1 — Write tests first (run in parallel across two terminals):
Task T032: "Unit test for InviteUser use case in backend/tests/unit/usecase/invite_user_test.go"
Task T033: "Unit test for AcceptInvitation use case in backend/tests/unit/usecase/accept_invitation_test.go"
Task T034: "Unit test for ResendInvitation use case in backend/tests/unit/usecase/resend_invitation_test.go"
# → Confirm all three fail (compilation error or mock assertion failure)

# Step 2 — Implement use cases sequentially (each builds on prior):
Task T035: "Implement InviteUser use case in backend/usecase/user/invite_user.go"
Task T036: "Implement AcceptInvitation use case in backend/usecase/user/accept_invitation.go"
Task T037: "Implement ResendInvitation use case in backend/usecase/user/resend_invitation.go"
# → Run go test ./backend/tests/unit/usecase/... and confirm green

# Step 3 — Integration test + handler (test first):
Task T058: "Integration test for user_handler_test.go (invite, resend endpoints)"
# → Confirm test fails
Task T062: "Create UserHandler in backend/interfaces/http/handlers/user_handler.go"
# → Confirm integration test green

# Step 4 — Frontend (tests first, then parallel implementation):
Task T091: "Unit test for AcceptInvitationForm.test.tsx"
# → Confirm test fails
Tasks T092, T093: "Create InvitationExpired.tsx and AcceptInvitationForm.tsx" (parallel)
Task T094: "Update AcceptInvitationPage.tsx with useInvitation hook"
```

---

## Implementation Strategy

### MVP Scope (deliver in order)

1. **Phase 1 + Phase 2 + Phase 3**: Migrations + Domain + Infrastructure (blocking foundation)
2. **US7 (JWT Claims)**: Implement immediately — validates the RBAC data path with no UI required
3. **US3 (User List)**: Simplest P1 story — read-only queries, no state transitions
4. **US1 (User Invitation)**: Core user lifecycle entry point — enables all other stories
5. **US2 (Role Management)**: Directly delivers RBAC value; US7 + US1 must exist first

**After MVP**: US4 (Sessions/Activity), US5 (Enable/Disable), US6 (Delete User) — all P2/P3 security controls.

### Summary

| Phase | Tasks | Story Coverage |
| ----- | ----- | -------------- |
| Setup (Migrations) | T001–T004 | — |
| B1 Domain | T005–T017 | Foundational |
| B2 Infrastructure | T018–T026 | Foundational |
| B3 Use Cases | T027–T056 | US1–US7 |
| B4 HTTP Interfaces | T057–T068 | US1–US7 |
| B5 Background Jobs | T069 | — |
| F1 Frontend Foundation | T070–T078 | Foundational |
| F2 User List & Search | T079–T085 | US3 |
| F3 User Detail & Roles | T086–T090 | US2 |
| F4 Invitation Flow | T091–T094 | US1 |
| F5 Sessions & Activity | T095–T096 | US4, US5 |
| Polish | T097–T102 | Cross-cutting |
| **Total** | **102 tasks** | **US1–US7** |

**Parallel opportunities identified**: 40+ tasks marked [P] across all phases.
