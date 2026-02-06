# Implementation Plan: Client Application Registration Portal

**Branch**: `004-client-registration-app` | **Date**: February 7, 2026 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification and 5 clarifications from `/specs/004-client-registration-app/spec.md`

**Note**: This plan follows the existing tech stack and Clean Architecture patterns from feature 003-sso-auth-provider.

## Summary

Build a self-service portal for tenant administrators to register and manage OAuth 2.0 client applications. Administrators can create new clients, configure redirect URIs, manage credentials (with one-time secret display), and toggle application status. The system integrates with the Core SSO Auth Provider (feature 003) to supply client configurations for authentication flows. Built with Go backend (CRUD use cases), PostgreSQL for persistent storage, Redis for rate limiting, and React/TypeScript frontend for an intuitive admin UI.

## Technical Context

**Language/Version**:

- Backend: Go 1.21+ (same as feature 003)
- Frontend: TypeScript 5.x (strict mode enabled, same as feature 003)

**Primary Dependencies**:

- **Backend**:
  - `github.com/jackc/pgx/v5` (PostgreSQL driver, existing)
  - `github.com/redis/go-redis/v9` (Redis client for rate limiting, existing)
  - `github.com/go-chi/chi/v5` (HTTP router, existing)
  - `github.com/ulule/limiter/v3` (Rate limiting library, existing)
  - `crypto/rand` (secure random generation for client_id/client_secret)
  - Standard library: `crypto/sha256`, `encoding/base64` (PKCE hashing for redirect URI validation)
- **Frontend**:
  - React 18+ (functional components only, same as feature 003)
  - React Router (SPA navigation, existing)
  - Axios or Fetch API (HTTP client, existing)
  - Shadcn/UI or Material-UI (UI components, existing)

**Storage**:

- **PostgreSQL 14+**: Persistent data (clients, redirect URIs, client secret history, audit events) - shares database with feature 003
- **Redis 7+**: Ephemeral data (rate limit counters per tenant, 1-hour TTL) - shares Redis instance with feature 003

**Testing**:

- **Backend**: Go's built-in `testing` package, manual mocks for interfaces
- **Frontend**: Jest + React Testing Library
- **Integration**: Go `httptest` package for API tests

**Target Platform**:

- Linux server (production, same as feature 003)
- Docker/docker-compose for local development
- Cloud-agnostic (AWS, GCP, Azure)

**Project Type**: Web application (backend API + frontend admin portal)

**Performance Goals**:

- Client registration: <500ms p95 latency (validation + encryption + DB insert)
- Client list retrieval: <100ms p95 (paginated queries with pagination)
- Dashboard load: <200ms p95 (aggregate statistics query)
- Throughput: 100 req/s per backend instance (admin traffic, not high-load)
- Rate limiting: Instant response with cached counter (Redis)

**Constraints**:

- Rate limiting: 20 client registrations per tenant per hour (configurable window)
- Client secret: Minimum 32 characters, never stored in plaintext (bcrypt hashing with cost factor 12)
- Client ID: UUID v4 format, globally unique
- Client status: Only active/inactive (soft delete, no permanent deletion)
- Secret regeneration: Immediate invalidation of old secret (no grace period)
- Redirect URI validation: RFC 3986 compliance, HTTPS required (except localhost)
- Security: Role-based access control (tenant administrators only)

**Scale/Scope**:

- Multi-tenant: 100+ tenants (shared with feature 003)
- Client applications: 10-50 OAuth clients per tenant initially, up to 100 long-term
- Administrators: 2-5 admins per tenant
- Redirect URIs: 1-10 per client application
- Concurrent requests: 50 admin requests/minute baseline
- Database: Shared PostgreSQL from feature 003

## Constitution Check

_GATE: Must pass before Phase 0 research. Re-check after Phase 1 design._

**Clean Architecture Compliance**:

- [x] Domain layer has no imports from infrastructure or frameworks
  - _Design_: Domain entities (Client, ClientRedirectURI, ClientSecretHistory, ClientAuditEvent) use only Go standard library types
  - _Design_: Repository interfaces (ClientRepository, ClientAuditRepository) defined in `domain/repositories/`
  - _Design_: Service interfaces defined in `domain/services/`
  - _Design_: Zero dependencies on pgx, go-redis, or chi in domain layer
- [x] All repository/service interfaces defined in Domain layer
  - _Design_: ClientRepository interface defines CRUD operations in `domain/repositories/client_repository.go`
  - _Design_: ClientAuditRepository interface in `domain/repositories/client_audit_repository.go`
  - _Design_: RateLimitService interface in `domain/services/rate_limit_service.go`
- [x] Concrete implementations only in Infrastructure layer
  - _Design_: PostgreSQL client repository in `infrastructure/persistence/postgres/client_repository.go`
  - _Design_: Redis rate limiter in `infrastructure/persistence/redis/rate_limiter.go`
  - _Design_: Audit repository in `infrastructure/persistence/postgres/client_audit_repository.go`
- [x] Dependency arrows verified to point inward (toward Domain)
  - _Design_: Handlers (outer) → Use Cases (middle) → Domain (inner)
  - _Design_: Infrastructure implements Domain interfaces
  - _Design_: Use cases accept repository interfaces, not concrete types

**SOLID Principles Compliance**:

- [x] Each module has single, well-defined responsibility (SRP)
  - _Design_: `RegisterClient` use case: single purpose (create new client)
  - _Design_: `UpdateClientRedirectURIs` use case: single purpose (manage URIs)
  - _Design_: `RegenerateClientSecret` use case: single purpose (credential rotation)
  - _Design_: Handlers only transform HTTP ↔ domain types
- [x] Domain depends only on abstractions/interfaces (DIP)
  - _Design_: Use cases accept `ClientRepository` interface, not concrete type
  - _Design_: Use cases accept `RateLimitService` interface, not Redis directly
  - _Design_: No concrete type dependencies in domain layer
- [x] No direct database/external API calls from business logic
  - _Design_: All database access through repository interfaces
  - _Design_: Rate limiting through RateLimitService interface
  - _Design_: Audit logging through ClientAuditRepository interface
- [x] Interface segregation verified for all contracts
  - _Design_: ClientRepository: 6 methods (Create, GetByID, Update, GetByTenantID, Delete, ListByTenant)
  - _Design_: ClientAuditRepository: 2 methods (Log, GetByClientID)
  - _Design_: RateLimitService: 2 methods (CheckLimit, RecordAttempt)

**Testing Requirements**:

- [x] Unit test plan documented for all business logic (target: ≥85% coverage)
  - _Plan_: Domain entities tested in `tests/unit/domain/`
  - _Plan_: Use cases tested in `tests/unit/usecase/` with mocked repositories
  - _Plan_: All 6 user stories have corresponding use case tests
  - _Plan_: Coverage target: 85%+ for domain + usecase layers
- [x] Integration test plan for all handlers/controllers
  - _Plan_: Integration tests in `tests/integration/`
  - _Plan_: Each endpoint tested with real PostgreSQL + Redis
  - _Plan_: Tests: `client_registration_test.go`, `client_management_test.go`, `client_dashboard_test.go`
- [x] Test isolation strategy defined (mocking approach)
  - _Plan_: Unit tests: Use manual mocks for repositories
  - _Plan_: Integration tests: Testcontainers for PostgreSQL and Redis
  - _Plan_: Each test creates isolated database schema
  - _Plan_: Redis FLUSHDB per test for isolation
- [x] Test-first workflow feasible for this feature
  - _Plan_: Write interface definitions first
  - _Plan_: Write failing tests for use cases
  - _Plan_: Implement use cases to pass tests
  - _Plan_: Write integration tests for handlers

**Code Conventions**:

- [x] Backend: Follows Effective Go, lowercase packages, exported function docs
  - _Design_: Package names: `domain`, `usecase`, `infrastructure`, `interfaces` (lowercase)
  - _Design_: Exported types/functions documented: `// Client represents an OAuth 2.0 client application`
  - _Design_: Standard Go project layout followed
- [x] Frontend: TypeScript strict mode, PascalCase components, functional components only
  - _Design_: tsconfig.json: `"strict": true`
  - _Design_: Components: `ClientRegistration`, `ClientManagement`, `Dashboard` (PascalCase)
  - _Design_: All components use hooks (useState, useEffect, custom hooks)
- [x] Clear separation between backend and frontend
  - _Design_: Backend: `backend/domain/`, `backend/usecase/`, `backend/infrastructure/`, `backend/interfaces/`
  - _Design_: Frontend: `frontend/src/components/`, `frontend/src/services/`, `frontend/src/hooks/`
  - _Design_: API contracts in `specs/004-client-registration-app/contracts/`

**Violations Requiring Justification** (fill Complexity Tracking section if any):

- [x] No constitution violations OR all violations documented with justification
  - _Status_: No violations. All Clean Architecture and SOLID principles satisfied.
  - _Validation_: Repository pattern used, dependency inversion enforced, interfaces in domain, implementations in infrastructure.

## Project Structure

### Documentation (this feature)

```text
specs/004-client-registration-app/
├── spec.md              # Feature specification (complete)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output: Technology decisions (to be created)
├── data-model.md        # Phase 1 output: Database schemas (to be created)
├── quickstart.md        # Phase 1 output: Developer setup guide (to be created)
├── contracts/           # Phase 1 output: API specifications (to be created)
│   ├── openapi.yaml     # OpenAPI 3.0 spec with client endpoints
│   └── README.md        # API documentation guide
├── checklists/          # Quality validation checklists
│   └── requirements.md  # Specification quality checklist (complete)
└── tasks.md             # Phase 2 output (/speckit.tasks command - to be created)
```

### Source Code (repository root)

```text
backend/
├── domain/              # Clean Architecture: Entities, business logic, interfaces
│   ├── entities/
│   │   ├── client.go                    # NEW: OAuth client entity
│   │   ├── client_redirect_uri.go       # NEW: Allowed redirect URI entity
│   │   ├── client_secret_history.go     # NEW: Secret rotation history
│   │   └── client_audit_event.go        # NEW: Audit log entity
│   ├── repositories/    # Interface definitions only
│   │   ├── client_repository.go         # NEW: Client CRUD interface
│   │   └── client_audit_repository.go   # NEW: Audit logging interface
│   └── services/        # Service interface definitions only
│       ├── rate_limit_service.go        # NEW: Rate limiting interface
│       └── random_service.go            # NEW: Secure random generation interface
├── usecase/             # Application business rules
│   └── client/          # NEW: Client application management
│       ├── register_client.go           # NEW: Create new client
│       ├── get_client.go                # NEW: Retrieve client details
│       ├── update_client.go             # NEW: Update application properties
│       ├── list_clients.go              # NEW: List tenant's clients with pagination
│       ├── toggle_client_status.go      # NEW: Activate/deactivate client
│       ├── regenerate_secret.go         # NEW: Rotate client_secret
│       ├── add_redirect_uri.go          # NEW: Add redirect URI to client
│       ├── remove_redirect_uri.go       # NEW: Remove redirect URI from client
│       └── get_dashboard_stats.go       # NEW: Get client statistics
├── infrastructure/      # Concrete implementations (DB, external APIs)
│   ├── persistence/     # Repository implementations
│   │   ├── postgres/
│   │   │   ├── client_repository.go         # NEW: PostgreSQL client CRUD
│   │   │   ├── client_audit_repository.go   # NEW: PostgreSQL audit logging
│   │   │   └── migrations/                  # NEW database migrations
│   │   │       ├── 000008_create_clients_table.up.sql
│   │   │       ├── 000008_create_clients_table.down.sql
│   │   │       ├── 000009_create_client_redirect_uris_table.up.sql
│   │   │       ├── 000009_create_client_redirect_uris_table.down.sql
│   │   │       ├── 000010_create_client_secret_history_table.up.sql
│   │   │       ├── 000010_create_client_secret_history_table.down.sql
│   │   │       ├── 000011_create_client_audit_events_table.up.sql
│   │   │       └── 000011_create_client_audit_events_table.down.sql
│   │   └── redis/
│   │       └── rate_limiter.go          # NEW: Redis rate limiting
│   └── services/        # External service implementations
│       ├── secure_random.go             # NEW: cryptographically secure random generation
│       └── redis_rate_limiter.go        # NEW: Redis-backed rate limiting
├── interfaces/          # Handlers, controllers (outer layer)
│   └── http/
│       ├── handlers/
│       │   └── client_handler.go        # NEW: Client registration & management endpoints
│       ├── middleware/
│       │   ├── client_tenant_context.go # NEW: Extract tenant from session
│       │   └── admin_only.go            # NEW: Verify admin role
│       └── router.go                    # Update: Add client routes
├── migrations/          # (Database migration files listed above)
├── tests/
│   ├── unit/
│   │   ├── domain/
│   │   │   ├── client_test.go           # NEW
│   │   │   ├── client_redirect_uri_test.go # NEW
│   │   │   └── validation_test.go       # NEW: Validation rules
│   │   └── usecase/
│   │       ├── register_client_test.go  # NEW
│   │       ├── update_client_test.go    # NEW
│   │       ├── regenerate_secret_test.go # NEW
│   │       ├── list_clients_test.go     # NEW
│   │       └── rate_limit_test.go       # NEW
│   ├── integration/
│   │   ├── client_registration_test.go  # NEW: POST /api/admin/clients
│   │   ├── client_management_test.go    # NEW: GET, PATCH, DELETE /api/admin/clients/*
│   │   ├── redirect_uri_test.go         # NEW: Redirect URI validation
│   │   ├── secret_regeneration_test.go  # NEW: Secret rotation flow
│   │   └── rate_limiting_test.go        # NEW: Rate limit enforcement
│   └── mocks/
│       ├── client_repository.go         # NEW
│       ├── client_audit_repository.go   # NEW
│       └── rate_limit_service.go        # NEW
├── go.mod
├── go.sum
└── README.md            # Update with client registration setup

frontend/
├── src/
│   ├── components/
│   │   └── admin/       # Admin portal components (new or updated)
│   │       ├── ClientRegistration.tsx      # NEW: Registration form UI
│   │       ├── ClientForm.tsx              # NEW: Create/edit client form
│   │       ├── ClientList.tsx              # NEW: Sortable client list
│   │       ├── ClientDetails.tsx           # NEW: Client detail view
│   │       ├── CredentialDisplay.tsx       # NEW: One-time secret display
│   │       ├── RedirectURIManager.tsx      # NEW: Add/remove redirect URIs
│   │       ├── ClientDashboard.tsx         # NEW: Overview dashboard
│   │       └── ConfirmDialog.tsx           # NEW: Confirmation modals
│   ├── services/        # API clients
│   │   └── clientService.ts             # NEW: Client management API
│   ├── hooks/           # Custom React hooks
│   │   ├── useClientForm.ts             # NEW: Client form logic
│   │   └── useDashboard.ts              # NEW: Dashboard data fetching
│   ├── types/           # TypeScript definitions
│   │   └── client.ts                    # NEW: Client type definitions
│   └── utils/           # Utility functions
│       ├── validation.ts                # NEW: Client/URI validation utilities
│       └── secretHandling.ts            # NEW: Secure secret handling
├── tests/
│   ├── unit/
│   │   ├── components/
│   │   │   ├── ClientForm.test.tsx      # NEW
│   │   │   ├── ClientList.test.tsx      # NEW
│   │   │   ├── CredentialDisplay.test.tsx # NEW
│   │   │   └── ClientDashboard.test.tsx # NEW
│   │   ├── services/
│   │   │   └── clientService.test.ts    # NEW
│   │   └── utils/
│   │       ├── validation.test.ts       # NEW
│   │       └── secretHandling.test.ts   # NEW
│   └── integration/
│       └── client-management-flow.test.tsx # NEW: Full flow test
└── README.md            # Update with client management guide
```

**Structure Decision**: Reuses existing web application structure from feature 003. Client registration is a focused feature with 4 database tables, 8-9 use cases, 1 admin handler, and 6+ frontend components. No disruption to existing tenant registration or OAuth auth provider; purely additive feature within the same architectural boundaries.

## Implementation Phases

### Phase 0: Research & Validation (Parallel)

**Dependencies**: Specification clarifications (COMPLETE) → Ready to proceed

**Outputs**:

1. `research.md` - Technology decisions and vendor analysis
2. `data-model.md` - Complete database schema with relationships
3. `contracts/openapi.yaml` - REST API specification

**Estimated Duration**: 1-2 weeks

**Parallel Research Streams**:

| Stream | Research Question                                                             | Responsibility |
| ------ | ----------------------------------------------------------------------------- | -------------- |
| A      | Best practices for secure client secret generation and storage in Go          | Backend Lead   |
| B      | Rate limiting patterns in Go (token bucket vs sliding window vs fixed window) | Backend Lead   |
| C      | Form validation and URI format testing in TypeScript/React                    | Frontend Lead  |
| D      | Session-based tenant context implementation compared to alternatives          | Architecture   |

### Phase 1: Design & Contracts (Sequential after Phase 0)

**Prerequisites**: Phase 0 research complete

**Deliverables**:

1. **data-model.md**:
   - 4 new database tables (clients, client_redirect_uris, client_secret_history, client_audit_events)
   - Relationships and constraints
   - Indexes for performance

2. **OpenAPI Contract** (`contracts/openapi.yaml`):
   - Client registration: `POST /api/admin/clients`
   - Client details: `GET /api/admin/clients/{client_id}`
   - Client list: `GET /api/admin/clients` (paginated)
   - Update client: `PATCH /api/admin/clients/{client_id}`
   - Toggle status: `PATCH /api/admin/clients/{client_id}/status`
   - Regenerate secret: `POST /api/admin/clients/{client_id}/rotate-secret`
   - Add URI: `POST /api/admin/clients/{client_id}/redirect-uris`
   - Remove URI: `DELETE /api/admin/clients/{client_id}/redirect-uris/{uri_id}`
   - Dashboard: `GET /api/admin/clients/dashboard/stats`

3. **quickstart.md**:
   - Backend setup (database migrations, environment variables)
   - Frontend setup (component structure, API integration)
   - Local development workflow

**Estimated Duration**: 1-2 weeks

### Phase 2: Implementation (Staged by user story priority)

**Prerequisites**: Phase 1 design complete

**Implementation Order** (following spec priority):

| Phase | Story                        | Backend Components                                      | Frontend Components              | Tests              | Duration |
| ----- | ---------------------------- | ------------------------------------------------------- | -------------------------------- | ------------------ | -------- |
| 2A    | P1: Registration Form        | RegisterClient use case, validation, Client entity      | ClientForm, CredentialDisplay    | Unit + integration | 1 week   |
| 2B    | P1: Credential Security      | Bcrypt integration, secret storage, one-time display    | Display copy button, mask logic  | Unit + integration | 1 week   |
| 2C    | P2: Client List & Management | ListClients, GetClient, UpdateClient use cases, handler | ClientList, ClientDetails, forms | Unit + integration | 1 week   |
| 2D    | P2: Redirect URI Management  | AddURI, RemoveURI use cases, validation                 | RedirectURIManager component     | Unit + integration | 1 week   |
| 2E    | P3: Secret Regeneration      | RegenerateSecret use case, history tracking             | Confirmation, new display        | Unit + integration | 1 week   |
| 2F    | P3: Dashboard                | GetDashboardStats use case, activity tracking           | ClientDashboard, charts          | Unit + integration | 1 week   |

**Estimated Duration**: 6 weeks (can be parallelized: backend + frontend in parallel, 3 weeks elapsed time)

### Phase 3: Testing & Deployment

**Prerequisites**: Phase 2 implementation complete

**Activities**:

- End-to-end testing (registration → management → status toggle)
- Security audit (credential handling, rate limiting, tenant isolation)
- Performance testing (pagination, bulk operations)
- Load testing (rate limit accuracy)

**Estimated Duration**: 1-2 weeks

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

**No violations identified.** The design fully complies with the constitution:

- Clean Architecture: Domain layer isolated, all dependencies point inward
- SOLID Principles: Single responsibility per use case, dependency inversion through interfaces
- Testing: Unit + integration strategy with 85%+ coverage target
- Code Conventions: Go (Effective Go) and TypeScript (strict mode) standards

## Resource Requirements

### Backend Development

- **Primary**: 1 backend engineer (full-time)
- **Support**: Architecture review (1-2 hours/week)
- **Testing**: 1 QA engineer (part of integration testing)

### Frontend Development

- **Primary**: 1 frontend engineer (full-time)
- **Design**: UI/UX review (2-4 hours/week)
- **Accessibility**: A11y audit (1 day)

### Infrastructure

- **Database**: Use existing PostgreSQL instance (feature 003)
- **Redis**: Use existing Redis instance (feature 003)
- **Deployment**: Existing CI/CD pipeline (GitHub Actions or similar)

### Timeline

- **Phase 0**: 1-2 weeks (parallel research)
- **Phase 1**: 1-2 weeks (design)
- **Phase 2**: 3-6 weeks (implementation, can parallelize backend+frontend)
- **Phase 3**: 1-2 weeks (testing, optimization)
- **Total**: 6-12 weeks (with parallelization: 4-6 weeks elapsed time)

## Success Metrics

| Metric                 | Target                              | Validation            |
| ---------------------- | ----------------------------------- | --------------------- |
| Code coverage          | ≥85% (domain + usecase)             | Unit tests passing    |
| API response time      | <500ms p95                          | Load testing results  |
| Rate limiting accuracy | 100% enforcement                    | Integration tests     |
| Tenant isolation       | 0 cross-tenant access               | Security audit        |
| User satisfaction      | Admin can register client in <2 min | Manual testing        |
| Deployment             | No rollback due to bugs             | Production monitoring |

## Integration Points

### With Feature 003 (SSO Auth Provider)

- **Data Integration**: Client registrations stored in shared PostgreSQL table, consumed by OAuth authorization endpoint
- **Secret Validation**: Client authentication uses same bcrypt hashing/validation as feature 003
- **Audit Logging**: All client events logged to shared audit_logs table for compliance
- **Rate Limiting**: Shared Redis rate limit counters (separate keys: `client_register:*` vs `oauth_token:*`)
- **Session Context**: Uses same session middleware to identify tenant and administrator

### With Feature 001 (Tenant Registration)

- **Tenant Data**: Relies on existing tenant records; only tenant admins can register clients
- **User Management**: Uses existing user_roles table to verify admin status
- **No Direct Integration**: Feature 004 doesn't modify tenant data; purely additive

## Risks & Mitigation

| Risk                                      | Impact | Probability | Mitigation                                                        |
| ----------------------------------------- | ------ | ----------- | ----------------------------------------------------------------- |
| Client secret exposure in logs            | High   | Medium      | Logging framework audit, structured logging, secret masking       |
| Rate limit bypass via parallel requests   | Medium | Low         | Atomic Redis operations, distributed rate limiting                |
| UI form validation incomplete             | Medium | Medium      | Comprehensive frontend + backend validation, shared contract      |
| Migration conflicts with feature 003      | Medium | Low         | Isolated migration numbers (000008-000011), reviewed before merge |
| Performance degradation with 100+ clients | Low    | Low         | Pagination, database indexing, connection pooling                 |

## Notes

- This feature is a supporting portal for feature 003 (Core SSO Auth Provider)
- Reuses all existing infrastructure (database, Redis, authentication, session management)
- Can be developed in parallel with other features (no blocking dependencies)
- Focuses on admin usability with clear UX for secure credential handling
- All clarifications from the specification review have been incorporated into the plan
