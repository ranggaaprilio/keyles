# Implementation Tasks: Client Application Registration Portal

**Feature**: `004-client-registration-app`  
**Plan**: [plan.md](./plan.md) | **Spec**: [spec.md](./spec.md)  
**Generated**: February 7, 2026

## Overview

This document breaks down the implementation plan into concrete, executable tasks organized by phase and user story. Each task includes specific file paths, acceptance criteria, and parallelization opportunities.

**Total Tasks**: 180+ tasks across 4 phases (25 Phase 0 + 22 Phase 1 + 105 Phase 2 + 28 Phase 3)  
**Parallel Paths**: Backend (90+ tasks) and Frontend (50+ tasks) can be executed in parallel  
**Estimated Duration**: 6-12 weeks serial (4-6 weeks with parallelization)

---

## Phase 0: Research & Validation (2 weeks)

### Research Setup

- [ ] T001 Set up research documentation structure in `specs/004-client-registration-app/research.md`
- [ ] T002 Create research tracking spreadsheet for decisions, options, and recommendations

### Secure Random Generation Research

- [ ] T003 [P] Investigate `crypto/rand` package for client_id generation in Go 1.21+
- [ ] T004 [P] Document UUID v4 generation libraries and comparison (stdlib vs. google/uuid vs. others)
- [ ] T005 [P] Research secure client_secret generation (minimum 32 chars, entropy requirements)
- [ ] T006 [P] Evaluate bcrypt cost factor tradeoffs (cost 12 recommended in spec)
- [ ] T007 Compile findings into research.md with recommendations

### Rate Limiting Architecture Research

- [ ] T008 [P] Compare rate limiting patterns: token bucket vs sliding window vs fixed window
- [ ] T009 [P] Analyze Redis-backed rate limiting performance (single instance, 100 req/s baseline)
- [ ] T010 [P] Document existing `ulule/limiter` library capabilities and integration points
- [ ] T011 [P] Design rate limit key strategy: `client_register:{tenant_id}:{window}`
- [ ] T012 Compile rate limiting recommendation into research.md

### Frontend Validation Research

- [ ] T013 [P] Research RFC 3986 URI validation libraries for Go and TypeScript
- [ ] T014 [P] Analyze HTTPS enforcement patterns (exception: localhost HTTP)
- [ ] T015 [P] Document custom URI scheme validation (mobile app deep links)
- [ ] T016 [P] Research form validation patterns in React with TypeScript strict mode
- [ ] T017 Summarize validation strategy in research.md

### Session & Tenant Context Research

- [ ] T018 [P] Document existing session implementation in feature 003 codebase
- [ ] T019 [P] Research tenant context extraction from authenticated session
- [ ] T020 [P] Evaluate session middleware integration for client registration
- [ ] T021 Create architecture diagram showing session flow in research.md

### Database & Performance Research

- [ ] T022 [P] Analyze PostgreSQL query patterns for paginated client lists (1000+ records)
- [ ] T023 [P] Design database indexes for optimal performance (<100ms list queries)
- [ ] T024 [P] Research connection pooling configuration for shared PostgreSQL instance
- [ ] T025 Document schema decisions and index strategy in research.md

---

## Phase 1: Design & Contracts (2 weeks)

### Database Schema Design

- [ ] T026 Design `clients` table schema: id, client_id (UUID), client_secret_hash, name, tenant_id, status, created_at, updated_at, created_by, modified_by
- [ ] T027 Design `client_redirect_uris` table with foreign key to clients, validated URI, created_at
- [ ] T028 Design `client_secret_history` table: id, client_id, secret_hash, created_at, created_by, invalidated_at, reason
- [ ] T029 Design `client_audit_events` table: id, client_id, tenant_id, event_type, user_id, timestamp, ip_address, details (JSON)
- [ ] T030 Create migration files in `backend/migrations/` (T030 covers 000008-000011)
- [ ] T031 [P] Document schemas in `specs/004-client-registration-app/data-model.md` with relationships and constraints
- [ ] T032 [P] Create ER diagram showing table relationships (clients ← many redirect_uris, secret_history, audit_events)

### OpenAPI Contract Design

- [ ] T033 Design `POST /api/admin/clients` endpoint (register new client)
- [ ] T034 Design `GET /api/admin/clients` endpoint (list clients with pagination)
- [ ] T035 Design `GET /api/admin/clients/{client_id}` endpoint (client details)
- [ ] T036 Design `PATCH /api/admin/clients/{client_id}` endpoint (update client name and URIs)
- [ ] T037 Design `PATCH /api/admin/clients/{client_id}/status` endpoint (activate/deactivate)
- [ ] T038 Design `POST /api/admin/clients/{client_id}/rotate-secret` endpoint (regenerate secret)
- [ ] T039 Design `POST /api/admin/clients/{client_id}/redirect-uris` endpoint (add redirect URI)
- [ ] T040 Design `DELETE /api/admin/clients/{client_id}/redirect-uris/{uri_id}` endpoint (remove URI)
- [ ] T041 Design `GET /api/admin/clients/dashboard/stats` endpoint (dashboard statistics)
- [ ] T042 Create complete `specs/004-client-registration-app/contracts/openapi.yaml` with request/response schemas

### Developer Documentation

- [ ] T043 Write `specs/004-client-registration-app/quickstart.md` with backend setup (migrations, env vars)
- [ ] T044 [P] Document frontend setup (environment variables, API client configuration, component structure)
- [ ] T045 [P] Create local development workflow guide (database seeding, test data)
- [ ] T046 [P] Document API authentication: session-based tenant context requirement
- [ ] T047 Create API documentation in `specs/004-client-registration-app/contracts/README.md`

---

## Phase 2A: Implementation - User Story 1 (Client Registration Form) - 1 week

### Backend: Domain Layer

- [ ] T048 [P] Create `backend/domain/entities/client.go` with Client struct (id, client_id, secret_hash, name, tenant_id, status, timestamps, user tracking)
- [ ] T049 [P] Create `backend/domain/entities/client_redirect_uri.go` with ClientRedirectURI struct
- [ ] T050 Create `backend/domain/repositories/client_repository.go` interface with methods: Create, GetByID, GetByTenantID, ListByTenant, Update, Delete

### Backend: Use Cases

- [ ] T051 Create `backend/usecase/client/register_client.go` use case:
  - Input: client_name, redirect_uris[], tenant_id, user_id
  - Validation: name length 1-100, HTTPS URIs (except localhost), max 10 URIs
  - Output: generated client_id, client_secret (plaintext, shown once)
  - Side effects: save to database, log audit event
  - Accept criteria: ✓ unique client_id, ✓ bcrypt secret_hash, ✓ active status, ✓ timestamps set

- [ ] T052 Create `backend/usecase/client/validate_client_request.go` helper with validation functions for name and URIs

### Backend: Infrastructure - Database

- [ ] T053 [P] Implement `backend/infrastructure/persistence/postgres/client_repository.go` (Create, GetByID methods)
- [ ] T054 [P] Implement rate limiting repository in `backend/infrastructure/persistence/redis/rate_limiter.go` with CheckLimit and RecordAttempt methods
- [ ] T055 Modify database connection pool config in `backend/infrastructure/config/config.go` if needed for client repository

### Backend: Infrastructure - Services

- [ ] T056 [P] Implement `backend/infrastructure/services/secure_random.go` for UUID v4 client_id and 32-char secret generation
- [ ] T057 [P] Create `backend/infrastructure/services/redis_rate_limiter.go` wrapper for rate limiting logic

### Backend: HTTP Handlers

- [ ] T058 Create `backend/interfaces/http/handlers/client_handler.go` with POST handler for client registration
- [ ] T059 Create request/response types in handler for client creation
- [ ] T060 Implement middleware `backend/interfaces/http/middleware/client_tenant_context.go` to extract tenant from session
- [ ] T061 Implement middleware `backend/interfaces/http/middleware/admin_only.go` to verify admin role
- [ ] T062 Add client registration route to `backend/interfaces/http/router.go`: `POST /api/admin/clients`

### Backend: Testing

- [ ] T063 [P] Create `backend/tests/unit/domain/client_test.go` with entity validation tests
- [ ] T064 [P] Create `backend/tests/unit/usecase/register_client_test.go` with use case tests (mocked repositories)
- [ ] T065 [P] Create `backend/tests/integration/client_registration_test.go` with end-to-end registration flow
- [ ] T066 Create `backend/tests/mocks/client_repository.go` mock implementation
- [ ] T067 Create test data seeders in `backend/cmd/seed/` for client test fixtures

### Frontend: Components & Services

- [ ] T068 [P] Create `frontend/src/components/admin/ClientForm.tsx` component with:
  - Text input for application name (1-100 chars, validation)
  - Redirect URI form with add/remove buttons (max 10)
  - Submit and validation error display
  - Accept criteria: ✓ form validation, ✓ URI validation feedback, ✓ visual error states

- [ ] T069 [P] Create `frontend/src/components/admin/CredentialDisplay.tsx` component with:
  - Large display of client_id and client_secret
  - Warning message about one-time display
  - Copy button for each credential
  - Accept criteria: ✓ copy to clipboard, ✓ warning text, ✓ masked after navigation

- [ ] T070 [P] Create `frontend/src/services/clientService.ts` API client:
  - `registerClient()` method calling `POST /api/admin/clients`
  - Error handling and response parsing
  - Type-safe request/response

### Frontend: Types & Utils

- [ ] T071 [P] Create `frontend/src/types/client.ts` with TypeScript types: Client, CreateClientRequest, ClientResponse
- [ ] T072 [P] Create `frontend/src/utils/validation.ts` with:
  - `validateClientName()` - 1-100 chars
  - `validateRedirectURI()` - RFC 3986 format, HTTPS (except localhost)
  - `validateURICount()` - max 10

### Frontend: Testing

- [ ] T073 [P] Create `frontend/tests/unit/components/ClientForm.test.tsx` with form validation tests
- [ ] T074 [P] Create `frontend/tests/unit/components/CredentialDisplay.test.tsx` with credential display tests
- [ ] T075 [P] Create `frontend/tests/unit/utils/validation.test.ts` with validation function tests

---

## Phase 2B: Implementation - User Story 1 Part 2 (Credential Security) - 1 week

### Backend: Secret Storage & Validation

- [ ] T076 Integrate bcrypt hashing in `RegisterClient` use case (cost factor 12) in `backend/usecase/client/register_client.go`
- [ ] T077 Create `backend/domain/services/password_service.go` interface if not exists (reuse from feature 003)
- [ ] T078 Update `backend/infrastructure/services/bcrypt_password.go` to verify it works with RegisterClient

### Backend: Audit Logging

- [ ] T079 Create `backend/domain/repositories/client_audit_repository.go` interface with Log method
- [ ] T080 Implement `backend/infrastructure/persistence/postgres/client_audit_repository.go` with event logging
- [ ] T081 Integrate audit logging into `RegisterClient` use case with event: "client_created"
- [ ] T082 Update handler to log HTTP metadata (IP address, user agent) in audit event

### Frontend: Secure Credential Handling

- [ ] T083 Create `frontend/src/utils/secretHandling.ts` with:
  - `secureCopy()` - copy to clipboard without storing in state
  - `maskSecret()` - replace secret with asterisks after navigation
  - `preventScreenshot()` - disable screenshot capability on credential page

- [ ] T084 Implement copy-to-clipboard detection and success feedback in CredentialDisplay component
- [ ] T085 Add warning banner styling and accessibility (ARIA roles) to credential display

### Security Testing

- [ ] T086 [P] Create `backend/tests/integration/secret_storage_test.go` - verify secrets are bcrypt hashed, never plaintext
- [ ] T087 [P] Create security audit test for logging - verify secrets never appear in logs or error messages
- [ ] T088 [P] Create `frontend/tests/security/credentialDisplay.test.tsx` - verify secret masking on navigation

---

## Phase 2C: Implementation - User Story 2 (Client List & Management) - 1 week

### Backend: Use Cases

- [ ] T089 Create `backend/usecase/client/get_client.go` - retrieve single client by ID with tenant validation
- [ ] T090 Create `backend/usecase/client/list_clients.go` - paginated list of tenant's clients (default 20 per page)
- [ ] T091 Create `backend/usecase/client/update_client.go` - update client name and status (no direct secret update)
- [ ] T092 Create `backend/usecase/client/toggle_client_status.go` - activate/deactivate client (soft delete)

### Backend: Repository Implementation

- [ ] T093 Implement GetByID, GetByTenantID, ListByTenant, Update methods in `backend/infrastructure/persistence/postgres/client_repository.go`
- [ ] T094 Add database indexes for performance: (tenant_id, status), (tenant_id, created_at desc), (client_id)

### Backend: HTTP Handlers

- [ ] T095 Add GET handler in `backend/interfaces/http/handlers/client_handler.go` for client list (`GET /api/admin/clients`)
- [ ] T096 Add GET handler for client details (`GET /api/admin/clients/{client_id}`)
- [ ] T097 Add PATCH handler for updating client (`PATCH /api/admin/clients/{client_id}`)
- [ ] T098 Add PATCH handler for status toggle (`PATCH /api/admin/clients/{client_id}/status`)

### Backend: Testing

- [ ] T099 Create `backend/tests/unit/usecase/list_clients_test.go` with pagination tests
- [ ] T100 Create `backend/tests/unit/usecase/update_client_test.go`
- [ ] T101 Create `backend/tests/integration/client_management_test.go` with list, get, update workflows

### Frontend: Components

- [ ] T102 [P] Create `frontend/src/components/admin/ClientList.tsx` component with:
  - Sortable table (name, client_id, URIs count, status, created_at)
  - Pagination controls (20 items/page)
  - Row click to view details
  - Status badge (active/inactive)

- [ ] T103 [P] Create `frontend/src/components/admin/ClientDetails.tsx` component with:
  - Display all client properties (read-only except name and URIs)
  - Status toggle button with confirmation
  - Edit button for client name
  - Masked client_secret with "not visible" message
  - "Regenerate Secret" button

- [ ] T104 [P] Create `frontend/src/components/admin/ConfirmDialog.tsx` component (reusable confirmation modal)

### Frontend: Services

- [ ] T105 Extend `frontend/src/services/clientService.ts` with:
  - `listClients(page, perPage)` - GET /api/admin/clients
  - `getClient(client_id)` - GET /api/admin/clients/{client_id}
  - `updateClient(client_id, name)` - PATCH /api/admin/clients/{client_id}
  - `toggleClientStatus(client_id, active)` - PATCH /api/admin/clients/{client_id}/status

### Frontend: Custom Hooks

- [ ] T106 Create `frontend/src/hooks/useClientForm.ts` custom hook for form state management

### Frontend: Testing

- [ ] T107 [P] Create `frontend/tests/unit/components/ClientList.test.tsx` with table rendering and pagination
- [ ] T108 [P] Create `frontend/tests/unit/components/ClientDetails.test.tsx` with status toggle tests
- [ ] T109 [P] Create `frontend/tests/unit/services/clientService.test.ts` with API call mocking

---

## Phase 2D: Implementation - User Story 2 Part 2 (Redirect URI Management) - 1 week

### Backend: Use Cases

- [ ] T110 Create `backend/usecase/client/add_redirect_uri.go` - validate and add new URI to client (max 10 total)
- [ ] T111 Create `backend/usecase/client/remove_redirect_uri.go` - remove redirect URI by ID

### Backend: Validation & Constants

- [ ] T112 Create `backend/domain/services/uri_validator.go` interface for URI validation
- [ ] T113 Implement `backend/infrastructure/services/rfc3986_uri_validator.go` with:
  - HTTPS requirement enforcement (except localhost)
  - Custom scheme support (myapp://)
  - Length validation (max 2048 chars)
  - Format validation (must be valid URL)

- [ ] T114 Create `backend/usecase/client/validation_rules.go` with URI validation constants

### Backend: Repository Methods

- [ ] T115 Implement AddRedirectURI and RemoveRedirectURI methods in client repository

### Backend: HTTP Handlers

- [ ] T116 Add POST handler for adding redirect URI (`POST /api/admin/clients/{client_id}/redirect-uris`)
- [ ] T117 Add DELETE handler for removing redirect URI (`DELETE /api/admin/clients/{client_id}/redirect-uris/{uri_id}`)

### Backend: Testing

- [ ] T118 [P] Create `backend/tests/unit/services/uri_validator_test.go` with validation test cases
- [ ] T119 [P] Create `backend/tests/integration/redirect_uri_test.go` with add/remove workflows
- [ ] T120 Test edge cases: duplicate URIs, URL encoding, query parameters, localhost exceptions

### Frontend: Components

- [ ] T121 [P] Create `frontend/src/components/admin/RedirectURIManager.tsx` component with:
  - List of current redirect URIs
  - Add new URI form with validation
  - Delete button per URI with confirmation
  - Real-time validation feedback
  - Accept criteria: ✓ max 10 URIs enforced on UI, ✓ HTTPS warning for non-localhost, ✓ custom scheme support

### Frontend: Utilities

- [ ] T122 Update `frontend/src/utils/validation.ts` with comprehensive URI validation function
- [ ] T123 Create `frontend/src/utils/uriPatterns.ts` with URI scheme patterns for mobile apps

### Frontend: Testing

- [ ] T124 [P] Create `frontend/tests/unit/components/RedirectURIManager.test.tsx` with add/remove tests
- [ ] T125 [P] Create `frontend/tests/unit/utils/uriValidation.test.ts` with scheme and format tests

---

## Phase 2E: Implementation - User Story 3 (Client Secret Regeneration) - 1 week

### Backend: Use Cases & Service

- [ ] T126 Create `backend/usecase/client/regenerate_secret.go` use case:
  - Generate new 32-char secret
  - Invalidate old secret immediately (no grace period)
  - Record both secrets in history with reason "manual_rotation"
  - Log audit event with "secret_regenerated" type
  - Return new secret plaintext for display

- [ ] T127 Update ClientSecretHistory entity tracking: old secret → invalidated_at, new secret → created_at

### Backend: Repository Updates

- [ ] T128 Add SaveSecretHistory method to client audit/history repository
- [ ] T129 Update GetClient to return client_secret_hash as masked (never reveal in API)

### Backend: HTTP Handlers

- [ ] T130 Add POST handler for secret regeneration (`POST /api/admin/clients/{client_id}/rotate-secret`)
- [ ] T131 Add validation to prevent multiple regenerations in quick succession (client-side debounce + server-side check)

### Backend: Integration with OAuth Provider

- [ ] T132 Create synchronization mechanism so feature 003 (OAuth provider) invalidates old secrets immediately
- [ ] T133 Update feature 003 token endpoint to reject attempts with old client_secret with "invalid_client" error

### Backend: Testing

- [ ] T134 [P] Create `backend/tests/unit/usecase/regenerate_secret_test.go` with secret generation and invalidation
- [ ] T135 [P] Create `backend/tests/integration/secret_regeneration_test.go` with old/new secret validation
- [ ] T136 Test that old secret cannot be used for OAuth token exchange (integration with feature 003)

### Frontend: Components

- [ ] T137 [P] Create confirmation dialog for secret regeneration (already have ConfirmDialog)
- [ ] T138 Update ClientDetails to show "Regenerate Secret" button with confirmation flow
- [ ] T139 Integrate with clientService.regenerateSecret() API call
- [ ] T140 Display new secret on success (reuse CredentialDisplay component)

### Frontend: Testing

- [ ] T141 [P] Create `frontend/tests/unit/components/SecretRegeneration.test.tsx` with confirmation and display flow

---

## Phase 2F: Implementation - User Story 4 (Dashboard) - 1 week

### Backend: Use Cases & Queries

- [ ] T142 Create `backend/usecase/client/get_dashboard_stats.go` use case:
  - Count total active clients
  - Count total inactive clients
  - Count clients created in last 30 days
  - List recent authentication activity (last 7 days) aggregated by client

- [ ] T143 Create database query to fetch recent authentication requests from audit logs (optimized with indexes)
- [ ] T144 Create statistics aggregation logic in use case

### Backend: HTTP Handlers

- [ ] T145 Add GET handler for dashboard stats (`GET /api/admin/clients/dashboard/stats`)
- [ ] T146 Add caching layer (Redis, 5-min TTL) for dashboard stats to reduce database load

### Backend: Testing

- [ ] T147 [P] Create `backend/tests/unit/usecase/get_dashboard_stats_test.go`
- [ ] T148 [P] Create `backend/tests/integration/client_dashboard_test.go` with statistics validation

### Frontend: Components

- [ ] T149 [P] Create `frontend/src/components/admin/ClientDashboard.tsx` component with:
  - Summary cards: total active clients, total inactive clients, created last 30 days
  - Recent activity chart (last 7 days, authentications by client)
  - Configuration warnings (localhost only URIs in production)
  - Accept criteria: ✓ real-time stats, ✓ chart renders correctly, ✓ warnings displayed

- [ ] T150 Create `frontend/src/components/admin/ActivityChart.tsx` - bar or line chart visualization

### Frontend: Services & Hooks

- [ ] T151 Extend clientService with `getDashboardStats()` method
- [ ] T152 Create `frontend/src/hooks/useDashboard.ts` custom hook for data fetching and refresh

### Frontend: Testing

- [ ] T153 [P] Create `frontend/tests/unit/components/ClientDashboard.test.tsx` with stats rendering
- [ ] T154 [P] Create `frontend/tests/unit/hooks/useDashboard.test.ts` with data fetching logic

---

## Phase 2 Integration Testing

- [ ] T155 [P] Create `backend/tests/integration/client-management-flow.test.go` - end-to-end registration to management workflow
- [ ] T156 [P] Create `frontend/tests/integration/client-management-flow.test.tsx` - full UI workflow test

---

## Phase 3: Testing & Optimization (2 weeks)

### Security Audit & Feature Integration

- [ ] T157 Audit all endpoints for tenant isolation (verify tenant_id validation on all operations)
- [ ] T158 Verify client_secret never appears in logs, error responses, or database dumps
- [ ] T159 Test rate limiting accuracy (attempt 21 registrations, verify 21st is rejected)
- [ ] T160 Verify timing attack prevention with constant-time string comparison for secrets
- [ ] T181 [P] **Feature 003 OAuth Integration**: RegisterClient → Verify client_id appears in feature 003 ClientRepository.GetByID() with proper tenant_id isolation
- [ ] T182 [P] **Input Sanitization Integration**: Test XSS payloads in application_name and redirect_uris; verify HTML entity escaping and safe retrieval
- [ ] T183 [P] **HTTPS Enforcement Audit**: Verify all admin portal pages enforce HTTPS with HSTS headers in production environment
- [ ] T184 **OAuth Caching Verification**: Coordinate with feature 003 team; verify client registration invalidates any cached client lookups within 100ms
- [ ] T185 **Secret Invalidation in OAuth**: Secret regeneration → feature 003 token endpoint rejects old secret with 'invalid_client' error within 100ms
- [ ] T186 **OAuth Redirect URI Validation**: RegisterClient(multiple URIs) → feature 003 /authorize accepts registered URIs and rejects unregistered ones
- [ ] T187 **Inactive Client Blocking**: Toggle client inactive → feature 003 OAuth flow returns 'invalid_client' error, not authentication success

### Performance Testing

- [ ] T161 Load test client list endpoint with 1000+ clients (target: <100ms p95)
- [ ] T162 Load test client registration concurrent requests (target: 100 req/s, rate limiting enforced)
- [ ] T163 Profile database queries and verify indexes are used
- [ ] T164 Test connection pooling under load (25 max connections, no starvation)
- [ ] T191 **Latency Testing for Success Criteria SC-006**: Measure p95 latency for each endpoint using k6/wrk/JMeter:
  - (1) POST /api/admin/clients <500ms (registration with encryption)
  - (2) GET /api/admin/clients <100ms (paginated list)
  - (3) GET /api/admin/clients/{id} <100ms (detail view)
  - (4) PATCH /api/admin/clients/{id} <200ms (update name/URIs)
  - (5) PATCH /api/admin/clients/{id}/status <200ms (toggle active/inactive)
  - (6) GET /api/admin/clients/dashboard/stats <200ms (aggregate stats, 5-min Redis cache)
  - Success criteria: **ALL endpoints ≤ targets with 100+ concurrent users**

### End-to-End Testing

- [ ] T165 Test complete registration flow: form validation → credential display → list → management
- [ ] T166 Test secret regeneration with OAuth provider (feature 003 integration)
- [ ] T167 Test status toggle preventing authorization on inactive clients
- [ ] T168 Test dashboard with 50+ clients and various activity levels
- [ ] T192 **Edge Case Testing**:
  - (1) Concurrent admin attempts to register clients with identical names → Verify unique (tenant_id, application_name) constraint; one succeeds, second fails gracefully with "Name already exists for this tenant"
  - (2) Concurrent URI modifications to same client from different admins → Verify final state consistency (pessimistic lock or optimistic concurrency control)
  - (3) Duplicate redirect URI attempts within same registration → Verify prevented by unique constraint; clear error message shown
  - (4) Query parameters & fragments in redirect URIs (e.g., `https://example.com/callback?state=1`) → Verify normalized and preserved correctly
  - (5) UUID v4 collision (astronomically rare) → Verify regeneration loop instead of failure; log collision event for monitoring
  - (6) Cross-tenant access attempts (A's admin accessing B's clients) → Verify 403 Forbidden, no data leakage

### Documentation & Deployment

- [ ] T169 [P] Update `backend/README.md` with client registration setup instructions
- [ ] T170 [P] Update `frontend/README.md` with client management guide
- [ ] T171 [P] Create deployment guide for database migrations (000008-000011)
- [ ] T172 [P] Document configuration variables for rate limiting and feature flags

### Code Quality & Security Hardening

- [ ] T173 Run test coverage analysis (target: ≥85% for domain + usecase)
- [ ] T174 Run linting and formatting (golangci-lint, prettier)
- [ ] T175 Review code for security issues (SQL injection, XSS, CSRF, timing attacks)
- [ ] T176 Documentation review - all exported functions/types documented
- [ ] T188 **Implement Input Sanitization** (FR-038): HTML entity escaping for application_name before storage in RegisterClient handler; parameterized SQL queries enforced via pgx (prevents SQL injection)
- [ ] T189 **Implement HTTPS Enforcement** (FR-042): Add middleware in chi router to redirect HTTP→HTTPS in production; configure TLS settings; document env-based toggle
- [ ] T190 **Configure CSP Headers**: Set Content-Security-Policy headers on all admin portal responses (prevent inline scripts, restrict external assets)

### Preparation for Production

- [ ] T177 Create migration rollback plan for database changes
- [ ] T178 Document backup/restore procedures for client data
- [ ] T179 Create monitoring/alerting for rate limit abuse (>100 rejection/hour = alert)
- [ ] T180 Document incident response for compromised client_secret

---

## Task Dependencies & Parallel Execution

### Critical Path

```
T001-T025 (Phase 0 Research - can parallelize)
    ↓
T026-T047 (Phase 1 Design - can parallelize)
    ↓
T048-T087 (Phase 2A-2B User Story 1 - Backend & Frontend parallel)
    ↓
T089-T125 (Phase 2C-2D User Story 2 - Backend & Frontend parallel)
    ↓
T126-T154 (Phase 2E-2F User Story 3-4 - Backend & Frontend parallel)
    ↓
T155-T156 (Phase 2 Integration)
    ↓
T157-T192 (Phase 3 Testing, Security, OAuth Integration, & Optimization)
```

### Parallelization Opportunities

| Group        | Tasks     | Duration | Notes                                                           |
| ------------ | --------- | -------- | --------------------------------------------------------------- |
| Research     | T001-T025 | 2 weeks  | All 5 streams can run in parallel                               |
| Design       | T026-T047 | 2 weeks  | Schema, API, and docs can run in parallel                       |
| Backend US1  | T048-T087 | 1 week   | Domain + infrastructure in parallel                             |
| Frontend US1 | T068-T075 | 1 week   | Components + tests in parallel (parallel to backend)            |
| Backend US2  | T089-T125 | 1 week   | Use cases + handlers + tests in parallel                        |
| Frontend US2 | T102-T109 | 1 week   | Components + services + tests in parallel (parallel to backend) |
| Backend US3  | T126-T148 | 1 week   | Use case + handler + testing in parallel                        |
| Frontend US3 | T149-T154 | 1 week   | Components + hooks + testing in parallel (parallel to backend)  |
| Integration  | T155-T156 | 1 week   | E2E tests for all features                                      |
| Testing      | T157-T180 | 2 weeks  | Security, performance, and ops (parallel activities)            |

**Total Time**: 6-12 weeks serial → **4-6 weeks with full parallelization** (1 backend engineer + 1 frontend engineer)

---

## Success Criteria per Phase

### Phase 0 ✓

- [x] All research questions answered
- [x] Decision matrix completed for rate limiting, validation, session context
- [x] research.md document complete and reviewed

### Phase 1 ✓

- [ ] Database schema finalized and reviewed
- [ ] OpenAPI contract consistent with implementation
- [ ] Quickstart guide complete and tested

### Phase 2 ✓

- [ ] All 6 user stories implemented
- [ ] Code coverage ≥85% for domain + usecase
- [ ] All integration tests passing
- [ ] Endpoints functional and integrated with feature 003

### Phase 3 ✓ (Updated with 16 additional security/integration/edge case tasks)

- [ ] Security audit completed (tenant isolation, secret handling, rate limiting verified)
- [ ] Feature 003 integration tests passed (OAuth flows validated, secret invalidation verified)
- [ ] Input sanitization and HTTPS enforcement implemented
- [ ] Edge case testing completed (concurrency, constraints, cross-tenant isolation)
- [ ] Performance targets met (<500ms registration, <100ms list, <200ms dashboard)
- [ ] Zero outstanding bugs from manual testing
- [ ] Documentation complete and accurate
- [ ] Ready for production deployment

---

## Notes

- **Parallelization**: This task list is designed to allow backend and frontend work to happen in parallel after design phase completes
- **Test-Driven**: Each use case task should be preceded by failing tests (tests created in same phase)
- **Feature Integration**: Tasks T132-T133 require coordination with feature 003 (Core SSO Auth Provider) team
- **Database Migrations**: Must be reviewed and tested before production deployment
- **Rollback Plan**: Each migration must have corresponding down migration (automatic with used migration tool)
