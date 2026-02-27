# Tasks: OAuth Client Application Registration Portal

**Input**: Design documents from `/specs/004-client-app-registration/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Per constitution, tests are MANDATORY. All business logic requires unit tests (≥85% coverage), and all handlers require integration tests.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Backend (Clean Architecture)**: `backend/domain/`, `backend/usecase/`, `backend/infrastructure/`, `backend/interfaces/`
- **Frontend**: `frontend/src/components/`, `frontend/src/services/`, `frontend/src/hooks/`, `frontend/src/pages/`
- **Tests**: `backend/tests/unit/`, `backend/tests/integration/`, `backend/tests/mocks/`, `frontend/tests/unit/`

---

## Phase 1: Setup (Database Migration)

**Purpose**: Schema changes that all subsequent phases depend on

- [X] T001 Create migration files `backend/migrations/000008_add_client_type_and_description.up.sql` and `backend/migrations/000008_add_client_type_and_description.down.sql` per data-model.md (includes pg_trgm extension, client_type column, description column, relax client_secret NOT NULL, CHECK constraints, indexes)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Domain, repository, infrastructure, and frontend base changes that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Backend Domain & Repository Interfaces

- [X] T002 Update domain entity in `backend/domain/entities/client.go`: add `ClientType` and `Description` fields, add constants `ClientTypeConfidential`, `ClientTypePublic`, `MaxClientsPerTenant = 25`, update `Validate()` to enforce client type enum, description max 500 chars, and conditional secret requirement, add `ValidateRedirectURIStrict(uri string, allowInsecureLocalhost bool) error` for HTTPS enforcement with localhost exception per research.md §8
- [X] T003 [P] Update `ClientRepository` interface in `backend/domain/repositories/client_repository.go`: add `CountByTenant(ctx context.Context, tenantID string) (int, error)` and `ListByTenantPaginated(ctx context.Context, tenantID string, search string, page int, pageSize int) ([]*entities.Client, int, error)` methods
- [X] T004 [P] Update `RefreshTokenRepository` interface in `backend/domain/repositories/refresh_token_repository.go`: add `RevokeByClientID(ctx context.Context, clientID string) error` method

### Backend Infrastructure Implementations

- [X] T005 Update GORM model and implement new repository methods in `backend/infrastructure/persistence/postgres/client_repository.go`: update `PostgresClient` struct (add `Description *string`, `ClientType string`, make `ClientSecret *string` nullable), update `toDomain()`/`toModel()` conversions, implement `CountByTenant()` with row-level locking, implement `ListByTenantPaginated()` with ILIKE search and offset pagination
- [X] T006 [P] Implement `RevokeByClientID()` in `backend/infrastructure/persistence/postgres/refresh_token_repository.go`: UPDATE refresh_tokens SET is_revoked=true WHERE client_id=?
- [X] T007 [P] Create Redis client count cache in `backend/infrastructure/persistence/redis/client_count_cache.go`: implement `ClientCountCache` with `Get(tenantID)`, `Invalidate(tenantID)` methods using key pattern `client_count:{tenant_id}` with 60s TTL
- [X] T008 [P] Add revoked client blacklist check to token validation middleware: on client deletion, SET `revoked_client:{client_id}` with 900s TTL in Redis; check EXISTS during access token validation in auth middleware

### Backend Wiring

- [X] T009 Update dependency injection in `backend/cmd/server/main.go`: wire `AuditLogRepository` and Redis client count cache into client use case constructors, wire `RefreshTokenRepository` into `DeleteClientUseCase`
- [X] T010 [P] Generate/update GoMock mocks in `backend/tests/mocks/` for updated `ClientRepository`, `RefreshTokenRepository`, and `ClientCountCache` interfaces

### Frontend Base

- [X] T011 [P] Update frontend types in `frontend/src/types/client.ts`: add `ClientType` enum (`confidential` | `public`), `Description` field to client type, add `PaginatedResponse<T>` type with `total`, `page`, `page_size`, `total_pages`, add `CreateClientRequest`/`UpdateClientRequest` types with `client_type` field
- [X] T012 [P] Update frontend API service in `frontend/src/services/clientService.ts`: add `client_type` and `description` to create/update payloads, add `page`, `page_size`, `search` query params to list endpoint, update response types for pagination
- [X] T013 Create TanStack Query hooks in `frontend/src/hooks/useClients.ts`: implement `useClients(page, search)`, `useClient(id)`, `useCreateClient()`, `useUpdateClient()`, `useDeleteClient()`, `useRotateSecret()` wrapping clientService methods with proper cache invalidation

**Clean Architecture Checkpoint**:

- Domain layer established with interfaces only ✓
- Infrastructure layer provides implementations ✓
- No domain-to-infrastructure dependencies ✓

**Checkpoint**: Foundation ready — user story implementation can now begin

---

## Phase 3: User Story 1 — Developer Registers New OAuth Client Application (Priority: P1) 🎯 MVP

**Goal**: Tenant administrators can register confidential or public OAuth client applications and receive credentials (client_id + client_secret for confidential, client_id only for public)

**Independent Test**: Register a confidential client via POST `/api/v1/admin/clients` with name, description, type, and redirect URIs → receive client_id and client_secret. Register a public client → receive client_id only (no secret). Verify quota enforcement at 25 clients.

### Tests for User Story 1 (MANDATORY per constitution) ⚠️

> **Write these tests FIRST, ensure they FAIL before implementation**

- [X] T014 [P] [US1] Unit tests for `CreateClientUseCase` in `backend/tests/unit/client/create_client_test.go`: test confidential client creation (secret generated), public client creation (no secret), quota enforcement (reject at 25), redirect URI validation, audit log creation, invalid client type rejection
- [X] T015 [P] [US1] Integration test for POST `/api/v1/admin/clients` in `backend/tests/integration/client_handler_test.go`: test 201 response with client_id/client_secret for confidential, 201 without secret for public, 400 for invalid redirect URIs, 409 for quota exceeded, 401 for unauthenticated, 403 for non-admin

### Implementation for User Story 1

- [X] T016 [US1] Modify `CreateClientUseCase` in `backend/usecase/client/create_client.go`: add `AuditLogRepository` and `ClientCountCache` dependencies, accept `ClientType` and `Description` in input, check quota via `CountByTenant()` before creation, skip secret generation for public clients, create audit log entry with type `client_created`, invalidate count cache on success
- [X] T017 [US1] Modify `CreateClient` handler method in `backend/interfaces/http/handlers/client_handler.go`: add `client_type` (required, enum) and `description` (optional) to `CreateClientRequest` struct, return `client_secret: null` for public clients in response, return 409 with quota error message, add `client_type` and `description` to response struct
- [X] T018 [P] [US1] Create `CreateClientForm` component in `frontend/src/components/clients/CreateClientForm.tsx`: react-hook-form + zod validation, radio buttons for client type (confidential/public) with helper text, name input (3-100 chars), description textarea (max 500), dynamic redirect URI list editor with add/remove and inline HTTPS validation, submit button with loading state
- [X] T019 [P] [US1] Create `SecretDisplay` component in `frontend/src/components/clients/SecretDisplay.tsx`: modal overlay showing client_id and client_secret (confidential only), copy-to-clipboard buttons with visual feedback, prominent warning that secret cannot be retrieved again, "I have saved the secret" checkbox that enables dismiss button

**Checkpoint**: Confidential and public client registration works end-to-end. Credentials are generated, displayed once, and stored securely. Quota is enforced.

---

## Phase 4: User Story 2 — Developer Views and Manages Registered Applications (Priority: P1)

**Goal**: Tenant administrators can view a paginated, searchable list of their registered clients and access detailed configuration for each client

**Independent Test**: Register 3+ clients, access GET `/api/v1/admin/clients` → see paginated list with client_id, name, type, status. Access GET `/api/v1/admin/clients/{id}` → see full detail including redirect URIs (no secret). Search by name returns filtered results.

### Tests for User Story 2 (MANDATORY per constitution) ⚠️

- [X] T020 [P] [US2] Unit tests for `ListClientsUseCase` and `GetClientUseCase` in `backend/tests/unit/client/list_clients_test.go` and `backend/tests/unit/client/get_client_test.go`: test paginated listing with total count, search filter by name, page/page_size defaults, get client returns client_type and description, get client never returns secret hash
- [X] T021 [P] [US2] Integration tests for GET `/api/v1/admin/clients` and GET `/api/v1/admin/clients/{id}` in `backend/tests/integration/client_handler_test.go`: test 200 with pagination metadata, search query filtering, 404 for nonexistent client, tenant isolation (cannot see other tenant's clients)

### Implementation for User Story 2

- [X] T022 [US2] Modify `ListClientsUseCase` in `backend/usecase/client/list_clients.go`: accept `Search string`, `Page int`, `PageSize int` in input, validate page/pageSize defaults (page=1, pageSize=10, max=25), call `ListByTenantPaginated()`, return clients with pagination metadata (total, page, pageSize, totalPages)
- [X] T023 [US2] Modify `GetClientUseCase` in `backend/usecase/client/get_client.go`: include `ClientType` and `Description` in response, ensure secret hash is never returned
- [X] T024 [US2] Modify `ListClients` and `GetClient` handler methods in `backend/interfaces/http/handlers/client_handler.go`: parse `page`, `page_size`, `search` query params for list endpoint, return `ListClientsResponse` with pagination metadata, add `client_type` and `description` to `ClientResponse` struct
- [X] T025 [P] [US2] Create `ClientCard` component in `frontend/src/components/clients/ClientCard.tsx`: display client_name, client_type badge (confidential/public), client_id (truncated with copy), is_active status indicator, created_at date, click handler for navigation to detail view
- [X] T026 [US2] Create `ClientList` component in `frontend/src/components/clients/ClientList.tsx`: search input with debounced filtering, renders `ClientCard` items, pagination controls (prev/next/page numbers) using `useClients` hook, empty state when no clients, loading skeleton state
- [X] T027 [US2] Create `ClientDetail` component in `frontend/src/components/clients/ClientDetail.tsx`: display full client info (client_id, name, description, type, redirect URIs list, timestamps, status), action buttons for edit/rotate/delete (wired in later phases), back navigation to list
- [X] T028 [US2] Create `ClientManagementPage` in `frontend/src/pages/ClientManagementPage.tsx` and add `/admin/clients` and `/admin/clients/:id` routes to `frontend/src/App.tsx` with `ProtectedRoute` wrapper: page renders `ClientList` or `ClientDetail` based on route, "Register New Client" button opens `CreateClientForm`

**Checkpoint**: Full client dashboard with paginated list, search, and detail view working end-to-end. Combined with US1, this delivers a complete MVP.

---

## Phase 5: User Story 3 — Developer Updates Client Application Configuration (Priority: P2)

**Goal**: Tenant administrators can modify client name, description, and redirect URIs after registration

**Independent Test**: Update an existing client's description and redirect URIs via PUT `/api/v1/admin/clients/{id}` → changes persisted and reflected in detail view. Invalid redirect URIs rejected with validation error.

### Tests for User Story 3 (MANDATORY per constitution) ⚠️

- [X] T029 [P] [US3] Unit test for `UpdateClientUseCase` in `backend/tests/unit/client/update_client_test.go`: test partial update (only description), partial update (only redirect URIs), full update, redirect URI validation on update, audit log creation with changed fields, cannot change client_type after creation
- [X] T030 [P] [US3] Integration test for PUT `/api/v1/admin/clients/{id}` in `backend/tests/integration/client_handler_test.go`: test 200 with updated fields, 400 for invalid redirect URIs, 404 for nonexistent client, tenant isolation

### Implementation for User Story 3

- [X] T031 [US3] Modify `UpdateClientUseCase` in `backend/usecase/client/update_client.go`: add `AuditLogRepository` dependency, accept `Description` in input, create audit log entry with type `client_updated` and changed fields in details, reject attempts to change `client_type`
- [X] T032 [US3] Modify `UpdateClient` handler method in `backend/interfaces/http/handlers/client_handler.go`: add `description` field to `UpdateClientRequest` struct, ensure `client_type` is not accepted in update payload
- [X] T033 [US3] Create `EditClientForm` component in `frontend/src/components/clients/EditClientForm.tsx`: pre-populated form with current values, editable fields (name, description, redirect URIs), client_type displayed as read-only, same redirect URI validation as create form, save/cancel buttons with `useUpdateClient` mutation

**Checkpoint**: Client configuration updates work end-to-end with audit trail. Previous stories unaffected.

---

## Phase 6: User Story 4 — Developer Regenerates Client Secret (Priority: P2)

**Goal**: Tenant administrators can rotate a confidential client's secret, immediately invalidating the old one

**Independent Test**: POST `/api/v1/admin/clients/{id}/rotate-secret` for a confidential client → new secret returned, old secret invalidated. Same request for a public client → 400 error.

### Tests for User Story 4 (MANDATORY per constitution) ⚠️

- [X] T034 [P] [US4] Unit test for `RotateSecretUseCase` in `backend/tests/unit/client/rotate_secret_test.go`: test successful rotation for confidential client, reject rotation for public client with descriptive error, audit log creation with type `client_secret_rotated`, verify new secret hash differs from old
- [X] T035 [P] [US4] Integration test for POST `/api/v1/admin/clients/{id}/rotate-secret` in `backend/tests/integration/client_handler_test.go`: test 200 with new secret for confidential, 400 for public client, 404 for nonexistent client

### Implementation for User Story 4

- [X] T036 [US4] Modify `RotateSecretUseCase` in `backend/usecase/client/rotate_secret.go`: add `AuditLogRepository` dependency, guard against public client type (return error "secret rotation is not available for public clients"), create audit log entry with type `client_secret_rotated`
- [X] T037 [US4] Modify `RotateSecret` handler method in `backend/interfaces/http/handlers/client_handler.go`: return 400 with descriptive error for public clients, include `rotated_at` timestamp in response per OpenAPI contract
- [X] T038 [US4] Create `RotateSecretDialog` component in `frontend/src/components/clients/RotateSecretDialog.tsx`: confirmation dialog warning that old secret stops working immediately, confirm/cancel buttons, on confirm call `useRotateSecret` mutation, on success show `SecretDisplay` with new secret

**Checkpoint**: Secret rotation works for confidential clients, properly rejected for public clients. Audit trail captures rotation events.

---

## Phase 7: User Story 6 — Developer Accesses OAuth Integration Documentation (Priority: P2)

**Goal**: Developers see inline OAuth flow documentation with personalized code examples using their actual client_id

**Independent Test**: View a registered client's detail page → see documentation links and code examples with the client's actual client_id inserted into sample requests.

- [X] T039 [US6] Create `IntegrationDocs` component in `frontend/src/components/clients/IntegrationDocs.tsx`: expandable section in client detail, tabbed code examples (cURL, JavaScript, Go) for authorization code flow, PKCE flow, token exchange, and token refresh, dynamically insert the client's `client_id` into all examples, show different examples for confidential vs public client types, link to redirect URI requirements and HTTPS rules
- [X] T040 [P] [US6] Unit test for `IntegrationDocs` component in `frontend/tests/unit/components/clients/IntegrationDocs.test.tsx`: verify client_id is interpolated into code examples, verify confidential examples show client_secret placeholder, verify public examples show PKCE parameters, verify all OAuth flow tabs render

**Checkpoint**: Developers have personalized integration documentation accessible from the client detail view.

---

## Phase 8: User Story 5 — Developer Deletes Client Application (Priority: P3)

**Goal**: Tenant administrators can permanently deactivate a client, immediately revoking all associated tokens

**Independent Test**: DELETE `/api/v1/admin/clients/{id}` → client deactivated, removed from dashboard, all refresh tokens revoked, access tokens blacklisted in Redis for 15 minutes. Subsequent OAuth flows with deleted credentials fail immediately.

### Tests for User Story 5 (MANDATORY per constitution) ⚠️

- [X] T041 [P] [US5] Unit test for `DeleteClientUseCase` in `backend/tests/unit/client/delete_client_test.go`: test soft-delete sets is_active=false, refresh tokens revoked via `RevokeByClientID()`, Redis blacklist entry created with 900s TTL, audit log entry with type `client_deleted`, count cache invalidated
- [X] T042 [P] [US5] Integration test for DELETE `/api/v1/admin/clients/{id}` in `backend/tests/integration/client_handler_test.go`: test 204 no content response, client no longer appears in list, 404 for already-deleted client, tenant isolation

### Implementation for User Story 5

- [X] T043 [US5] Modify `DeleteClientUseCase` in `backend/usecase/client/delete_client.go`: add `AuditLogRepository`, `RefreshTokenRepository`, and Redis client dependencies, call `refreshTokenRepo.RevokeByClientID()` after soft-delete, set `revoked_client:{client_id}` in Redis with 900s TTL, create audit log entry with type `client_deleted`, invalidate count cache
- [X] T044 [US5] Modify `DeleteClient` handler method in `backend/interfaces/http/handlers/client_handler.go`: return 204 No Content on success per OpenAPI contract (currently may return 200)
- [X] T045 [US5] Create `DeleteClientDialog` component in `frontend/src/components/clients/DeleteClientDialog.tsx`: confirmation dialog with prominent warning that action is irreversible, explains all tokens will be revoked and applications will stop working, confirm/cancel buttons, on confirm call `useDeleteClient` mutation, redirect to client list on success

**Checkpoint**: Client deletion with full token revocation cascade works end-to-end. All 6 user stories complete.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Quality assurance, architecture compliance, and documentation

- [X] T046 [P] Verify ≥85% unit test coverage for domain layer (`backend/domain/entities/client.go`) and use case layer (`backend/usecase/client/*.go`) — run `go test -coverprofile=coverage.out ./...` and check coverage report
- [X] T047 [P] Architecture compliance review: verify no `domain/` imports from `infrastructure/` or `interfaces/`, all repository usage via interfaces, all use cases depend only on domain abstractions
- [X] T048 Run `specs/004-client-app-registration/quickstart.md` validation: follow the quickstart guide end-to-end to register a confidential client, a public client, list/search clients, update configuration, rotate secret, delete with token verification
- [X] T049 [P] Code cleanup: add GoDoc comments on all exported types and functions, verify error messages match OpenAPI contract examples, ensure consistent JSON field naming (snake_case) across all handler responses
- [X] T050 [P] [US1] Unit tests for `CreateClientForm` and `SecretDisplay` in `frontend/tests/unit/components/clients/CreateClientForm.test.tsx` and `frontend/tests/unit/components/clients/SecretDisplay.test.tsx`: test form validation (name length, URI format), client type radio selection, secret copy-to-clipboard, "I have saved" checkbox flow
- [X] T051 [P] [US2] Unit tests for `ClientList` and `ClientCard` in `frontend/tests/unit/components/clients/ClientList.test.tsx` and `frontend/tests/unit/components/clients/ClientCard.test.tsx`: test pagination controls, search input debounce, empty state, card rendering with type badge

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (migration must exist before infrastructure code references new columns)
- **User Stories (Phases 3–8)**: All depend on Phase 2 completion
  - US1 (Phase 3) and US2 (Phase 4) are both P1 — implement sequentially (US1 first for MVP)
  - US3 (Phase 5), US4 (Phase 6), US6 (Phase 7) are all P2 — can run in parallel after US1+US2
  - US5 (Phase 8) is P3 — can start after Phase 2 but benefits from US2 frontend being done
- **Polish (Phase 9)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (Register)**: Depends on Foundational only — no inter-story dependencies
- **US2 (View/Manage)**: Depends on Foundational only — benefits from US1 for testing (needs clients to list)
- **US3 (Update)**: Depends on Foundational only — frontend `EditClientForm` is accessed from `ClientDetail` (US2)
- **US4 (Rotate Secret)**: Depends on Foundational only — frontend `RotateSecretDialog` is accessed from `ClientDetail` (US2)
- **US5 (Delete)**: Depends on Foundational only — frontend `DeleteClientDialog` is accessed from `ClientDetail` (US2)
- **US6 (Docs)**: Depends on Foundational only — frontend `IntegrationDocs` is rendered within `ClientDetail` (US2)

### Within Each User Story

1. Tests MUST be written and FAIL before implementation
2. Use case logic before handler (domain → outward)
3. Backend before frontend (API must exist for frontend to consume)
4. Core implementation before integration
5. Story complete before moving to next priority

### Parallel Opportunities

**Phase 2 (Foundational)**:

```
Parallel batch 1: T003, T004 (different interface files)
Parallel batch 2: T005, T006, T007, T008 (different implementation files)
Parallel batch 3: T010, T011, T012 (mocks + frontend, different layers)
Sequential: T002 (entity, used by all), T009 (DI, depends on implementations), T013 (hooks, depends on service)
```

**Per User Story (e.g., US1)**:

```
Parallel: T014, T015 (test files are independent)
Sequential: T016 → T017 (use case before handler)
Parallel: T018, T019 (different frontend components)
```

**Cross-Story (after Phase 2)**:

```
US3, US4, US6 can all proceed in parallel (P2 stories, different files)
```

---

## Parallel Example: User Story 1

```bash
# Launch all tests in parallel (write first, expect failures):
Task T014: "Unit tests for CreateClientUseCase in backend/tests/unit/client/create_client_test.go"
Task T015: "Integration test for POST /api/v1/admin/clients in backend/tests/integration/client_handler_test.go"

# Sequential backend implementation:
Task T016: "Modify CreateClientUseCase in backend/usecase/client/create_client.go"
Task T017: "Modify CreateClient handler in backend/interfaces/http/handlers/client_handler.go"

# Launch frontend components in parallel:
Task T018: "Create CreateClientForm in frontend/src/components/clients/CreateClientForm.tsx"
Task T019: "Create SecretDisplay in frontend/src/components/clients/SecretDisplay.tsx"
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2)

1. Complete Phase 1: Setup (migration)
2. Complete Phase 2: Foundational (domain + infrastructure + frontend base)
3. Complete Phase 3: US1 — Register clients (core capability)
4. Complete Phase 4: US2 — View & manage clients (dashboard)
5. **STOP and VALIDATE**: Test registration + dashboard flow end-to-end
6. Deploy/demo if ready — developers can register clients and manage them

### Incremental Delivery

1. Setup + Foundational → Infrastructure ready
2. US1 (Register) → Test independently → **MVP: Developers can get credentials**
3. US2 (View/Manage) → Test independently → **Dashboard available**
4. US3 + US4 + US6 (Update + Rotate + Docs) → Test independently → **Full lifecycle**
5. US5 (Delete) → Test independently → **Complete feature**
6. Polish → Quality assurance → **Production ready**

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: US1 (Register) → US3 (Update)
   - Developer B: US2 (View/Manage) → US4 (Rotate Secret)
   - Developer C: US6 (Docs) → US5 (Delete)
3. Stories complete and integrate independently via shared interfaces

---

## Notes

- [P] tasks = different files, no dependencies on in-progress tasks
- [Story] label maps task to specific user story for traceability
- Each user story is independently completable and testable
- Verify tests fail before implementing (test-first per constitution)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Existing client CRUD is already scaffolded — tasks MODIFY existing files (not create from scratch)
- Handler file `client_handler.go` is modified across multiple stories but in different methods (no conflicts)
- Redis infrastructure (count cache + blacklist) is set up in Foundational but consumed by US1 and US5 respectively
