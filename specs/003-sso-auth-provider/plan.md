# Implementation Plan: Core SSO Auth Provider

**Branch**: `003-sso-auth-provider` | **Date**: 2025-01-16 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-sso-auth-provider/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement a production-grade OAuth 2.0 + OpenID Connect (OIDC) Single Sign-On provider for multi-tenant SaaS architecture. The system enables centralized authentication with role-based access control, supporting Authorization Code Flow with PKCE for security. Built using Go backend with Clean Architecture principles, PostgreSQL for persistent data, Redis for ephemeral data (authorization codes, sessions, refresh tokens), and React/TypeScript frontend for admin portal and consent UI.

## Technical Context

**Language/Version**:

- Backend: Go 1.21+ (latest stable)
- Frontend: TypeScript 5.x (strict mode enabled)

**Primary Dependencies**:

- **Backend**:
  - `github.com/ory/fosite` (OAuth 2.0/OIDC framework)
  - `github.com/jackc/pgx/v5` (PostgreSQL driver with connection pooling)
  - `github.com/redis/go-redis/v9` (Redis client with context support)
  - `github.com/go-chi/chi/v5` (HTTP router, lightweight and idiomatic)
  - `github.com/ulule/limiter/v3` (Rate limiting library)
  - Standard library: `crypto/rsa`, `crypto/sha256` (JWT signing, PKCE)
- **Frontend**:
  - React 18+ (functional components only)
  - React Router (SPA navigation)
  - Axios or Fetch API (HTTP client)
  - Shadcn/UI or Material-UI (UI components)

**Storage**:

- **PostgreSQL 14+**: Persistent data (tenants, clients, users, roles, refresh_tokens, signing_keys, audit_logs)
- **Redis 7+**: Ephemeral data (authorization codes 5min TTL, sessions 8hr TTL, refresh token cache 7day TTL, JWKS cache 1hr TTL, rate limit counters 1min TTL)

**Testing**:

- **Backend**: Go's built-in `testing` package, `go-mockgen` or manual mocks for interfaces
- **Frontend**: Jest + React Testing Library
- **Integration**: Go `httptest` package for API tests

**Target Platform**:

- Linux server (production)
- Docker/docker-compose for local development
- Cloud-agnostic (deployable to AWS, GCP, Azure)

**Project Type**: Web application (backend API + frontend admin portal)

**Performance Goals**:

- Authorization endpoint: <100ms p95 latency
- Token endpoint: <200ms p95 (includes DB + Redis operations)
- Discovery endpoints: <50ms p95 (cached responses)
- Throughput: 1000 req/s per backend instance
- JWKS caching: 1-hour TTL to minimize crypto operations

**Constraints**:

- Authorization codes: 5-minute expiration (RFC 6749)
- Access tokens: 15 minutes expiration (user-specified, high security)
- Refresh tokens: 7 days expiration (user-specified, high security)
- Session cookies: 8 hours expiration
- Rate limiting: 10 requests/minute per client_id on token endpoint
- JWT signing: RS256 with 2048-bit RSA keys minimum
- PKCE mandatory for all clients (no plain flow)

**Scale/Scope**:

- Multi-tenant: 100+ tenants initially, 1000+ long-term
- Users: 10k-100k users per tenant
- Concurrent sessions: 50k active sessions
- Client applications: 10-50 OAuth clients per tenant
- Database: PostgreSQL with connection pooling (25 max connections)
- Redis: Single instance initially, Redis Cluster for scale

## Constitution Check

_GATE: Must pass before Phase 0 research. Re-check after Phase 1 design._

**Clean Architecture Compliance**:

- [x] Domain layer has no imports from infrastructure or frameworks
  - _Design_: Domain entities (Client, User, Tenant, RefreshToken) use only Go standard library types
  - _Design_: Repository interfaces (ClientRepository, UserRepository, OTPRepository) defined in `domain/repositories/`
  - _Design_: Service interfaces (EmailService, OTPService, PasswordService) defined in `domain/services/`
  - _Design_: Zero dependencies on pgx, go-redis, fosite, or chi in domain layer
- [x] All repository/service interfaces defined in Domain layer
  - _Design_: All 6 repository interfaces in `domain/repositories/`
  - _Design_: All 3 service interfaces in `domain/services/`
  - _Design_: AuthService interface in `domain/services/auth_service.go`
- [x] Concrete implementations only in Infrastructure layer
  - _Design_: PostgreSQL repositories in `infrastructure/persistence/postgres/`
  - _Design_: Redis repositories in `infrastructure/persistence/redis/`
  - _Design_: Brevo email implementation in `infrastructure/services/brevo_email.go`
  - _Design_: Bcrypt password implementation in `infrastructure/services/bcrypt_password.go`
  - _Design_: Fosite OAuth adapter in `infrastructure/services/fosite_oauth_adapter.go`
- [x] Dependency arrows verified to point inward (toward Domain)
  - _Design_: Handlers (outer) → Use Cases (middle) → Domain (inner)
  - _Design_: Infrastructure implements Domain interfaces
  - _Design_: Use cases accept repository interfaces, not concrete types
  - _Design_: No domain imports from `infrastructure/`, `interfaces/`, or external frameworks

**SOLID Principles Compliance**:

- [x] Each module has single, well-defined responsibility (SRP)
  - _Design_: `AuthenticateAdmin` use case: single purpose (admin login)
  - _Design_: `RegisterTenant` use case: single purpose (tenant registration)
  - _Design_: `IssueToken` use case: single purpose (OAuth token issuance)
  - _Design_: Handlers only transform HTTP ↔ domain types, no business logic
- [x] Domain depends only on abstractions/interfaces (DIP)
  - _Design_: Use cases accept `UserRepository` interface, not `PostgresUserRepository`
  - _Design_: Use cases accept `OTPService` interface, not `CryptoOTPService`
  - _Design_: No concrete type dependencies in domain layer
- [x] No direct database/external API calls from business logic
  - _Design_: All database access through repository interfaces
  - _Design_: All email sending through EmailService interface
  - _Design_: OAuth logic through OAuthProvider interface (fosite abstraction)
- [x] Interface segregation verified for all contracts
  - _Design_: ClientRepository: 4 methods (Create, GetByID, Update, Delete)
  - _Design_: UserRepository: 5 methods (Create, GetByEmail, Update, UpdatePassword, ListByTenant)
  - _Design_: OTPService: 3 methods (Generate, Verify, Invalidate)
  - _Design_: No bloated interfaces with unused methods

**Testing Requirements**:

- [x] Unit test plan documented for all business logic (target: ≥85% coverage)
  - _Plan_: Domain entities tested in `tests/unit/domain/`
  - _Plan_: Use cases tested in `tests/unit/usecase/` with mocked repositories
  - _Plan_: All 8 user stories have corresponding use case tests
  - _Plan_: Coverage target: 85%+ for domain + usecase layers
- [x] Integration test plan for all handlers/controllers
  - _Plan_: Integration tests in `tests/integration/`
  - _Plan_: Each endpoint tested with real PostgreSQL + Redis (test containers)
  - _Plan_: Existing tests: `auth_test.go`, `registration_test.go`, `verification_test.go`
  - _Plan_: New tests: `oauth_auth_test.go`, `oauth_token_test.go`, `client_management_test.go`, `role_management_test.go`
- [x] Test isolation strategy defined (mocking approach)
  - _Plan_: Unit tests: Use `tests/mocks/` (manual or generated mocks)
  - _Plan_: Integration tests: Testcontainers for PostgreSQL and Redis
  - _Plan_: Each test creates isolated database schema (transaction rollback)
  - _Plan_: Redis FLUSHDB per test for isolation
- [x] Test-first workflow feasible for this feature
  - _Plan_: Write interface definitions first
  - _Plan_: Write failing tests for use cases
  - _Plan_: Implement use cases to pass tests
  - _Plan_: Write integration tests for handlers
  - _Plan_: Implement handlers
  - _Plan_: Follows Red-Green-Refactor cycle

**Code Conventions**:

- [x] Backend: Follows Effective Go, lowercase packages, exported function docs
  - _Design_: Package names: `domain`, `usecase`, `infrastructure`, `interfaces` (lowercase, single-word)
  - _Design_: Exported types/functions documented: `// Client represents an OAuth 2.0 client...`
  - _Design_: Standard Go project layout followed
- [x] Frontend: TypeScript strict mode, PascalCase components, functional components only
  - _Design_: tsconfig.json: `"strict": true`
  - _Design_: Components: `LoginForm`, `ConsentScreen`, `ClientManagement` (PascalCase)
  - _Design_: All components use hooks (useState, useEffect, custom hooks)
  - _Design_: No class components
- [x] Clear separation between backend (domain/usecase/infrastructure) and frontend (components/services)
  - _Design_: Backend: `backend/domain/`, `backend/usecase/`, `backend/infrastructure/`, `backend/interfaces/`
  - _Design_: Frontend: `frontend/src/components/`, `frontend/src/services/`, `frontend/src/hooks/`
  - _Design_: API contracts in `specs/003-sso-auth-provider/contracts/openapi.yaml`
  - _Design_: Frontend services use API client abstraction (axios/fetch wrapper)

**Violations Requiring Justification** (fill Complexity Tracking section if any):

- [x] No constitution violations OR all violations documented with justification
  - _Status_: No violations. All Clean Architecture and SOLID principles satisfied in design.
  - _Validation_: Repository pattern used, dependency inversion enforced, interfaces defined in domain, concrete implementations in infrastructure.
  - _Frontend_: TypeScript strict mode, functional components, proper separation of concerns.

## Project Structure

### Documentation (this feature)

```text
specs/003-sso-auth-provider/
├── spec.md              # Feature specification (complete)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output: Technology decisions (complete)
├── data-model.md        # Phase 1 output: Database schemas (complete)
├── quickstart.md        # Phase 1 output: Developer setup guide (complete)
├── contracts/           # Phase 1 output: API specifications (complete)
│   ├── openapi.yaml     # OpenAPI 3.0 spec with 14 endpoints
│   └── README.md        # API documentation guide
├── checklists/          # Quality validation checklists
│   └── requirements.md  # Specification quality checklist (complete)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT YET CREATED)
```

### Source Code (repository root)

```text
backend/
├── domain/              # Clean Architecture: Entities, business logic, interfaces
│   ├── entities/
│   │   ├── admin_user.go         # Existing
│   │   ├── audit_log.go          # Existing
│   │   ├── otp_verification.go   # Existing
│   │   ├── tenant.go             # Existing
│   │   ├── client.go             # NEW: OAuth client entity
│   │   ├── authorization_code.go # NEW: OAuth auth code entity
│   │   ├── refresh_token.go      # NEW: Refresh token entity (persistent)
│   │   ├── signing_key.go        # NEW: JWT signing key entity
│   │   └── user_role.go          # NEW: RBAC role assignment entity
│   ├── repositories/    # Interface definitions only
│   │   ├── audit_repository.go       # Existing
│   │   ├── otp_repository.go         # Existing
│   │   ├── tenant_repository.go      # Existing
│   │   ├── user_repository.go        # Existing
│   │   ├── client_repository.go      # NEW: CRUD for OAuth clients
│   │   ├── auth_code_repository.go   # NEW: Ephemeral auth codes (Redis)
│   │   ├── session_repository.go     # NEW: User sessions (Redis)
│   │   ├── refresh_token_repository.go # NEW: PostgreSQL + Redis cache
│   │   ├── signing_key_repository.go # NEW: RSA key management
│   │   └── role_repository.go        # NEW: RBAC role assignments
│   └── services/        # Service interface definitions only
│       ├── email_service.go        # Existing
│       ├── otp_service.go          # Existing
│       ├── password_service.go     # Existing
│       ├── oauth_provider.go       # NEW: OAuth 2.0 + OIDC operations
│       └── token_service.go        # NEW: JWT generation/validation
├── usecase/             # Application business rules
│   ├── auth/
│   │   ├── authenticate_admin.go    # Existing
│   │   ├── authorize_client.go      # NEW: OAuth authorization endpoint logic
│   │   ├── issue_token.go           # NEW: Token exchange logic
│   │   ├── refresh_token.go         # NEW: Refresh token flow
│   │   ├── revoke_token.go          # NEW: Token revocation
│   │   ├── validate_token.go        # NEW: JWT validation + userinfo
│   │   └── consent_decision.go      # NEW: User consent handling
│   ├── client/          # NEW: OAuth client management
│   │   ├── create_client.go         # NEW
│   │   ├── get_client.go            # NEW
│   │   ├── update_client.go         # NEW
│   │   ├── delete_client.go         # NEW
│   │   ├── list_clients.go          # NEW
│   │   └── rotate_secret.go         # NEW
│   ├── role/            # NEW: RBAC role management
│   │   ├── assign_role.go           # NEW
│   │   ├── revoke_role.go           # NEW
│   │   └── list_user_roles.go       # NEW
│   └── tenant/
│       ├── check_availability.go    # Existing
│       ├── register_tenant.go       # Existing
│       └── resend_otp.go            # Existing
├── infrastructure/      # Concrete implementations (DB, external APIs)
│   ├── config/
│   │   └── config.go                # Existing: Environment configuration
│   ├── persistence/     # Repository implementations
│   │   ├── postgres/
│   │   │   ├── tenant_repository.go     # Existing
│   │   │   ├── user_repository.go       # Existing
│   │   │   ├── audit_repository.go      # Existing
│   │   │   ├── client_repository.go     # NEW: PostgreSQL client CRUD
│   │   │   ├── refresh_token_repository.go # NEW: PostgreSQL refresh tokens
│   │   │   ├── signing_key_repository.go # NEW: PostgreSQL key storage
│   │   │   └── role_repository.go       # NEW: PostgreSQL role assignments
│   │   └── redis/
│   │       ├── otp_repository.go        # Existing
│   │       ├── auth_code_repository.go  # NEW: Redis auth codes (5min TTL)
│   │       ├── session_repository.go    # NEW: Redis sessions (8hr TTL)
│   │       ├── token_cache_repository.go # NEW: Redis refresh token cache
│   │       └── rate_limiter.go          # NEW: Redis rate limiting
│   └── services/        # External service implementations
│       ├── bcrypt_password.go       # Existing
│       ├── brevo_email.go           # Existing
│       ├── crypto_otp.go            # Existing
│       ├── jwt_auth_adapter.go      # Existing
│       ├── jwt_service.go           # Existing
│       ├── fosite_oauth_provider.go # NEW: Fosite OAuth adapter
│       └── rsa_token_service.go     # NEW: RS256 JWT signing
├── interfaces/          # Handlers, controllers (outer layer)
│   └── http/
│       ├── handlers/
│       │   ├── auth_handler.go          # Existing: Admin auth
│       │   ├── tenant_handler.go        # Existing: Tenant registration
│       │   ├── oauth_handler.go         # NEW: /oauth2/auth, /oauth2/token, /oauth2/revoke
│       │   ├── discovery_handler.go     # NEW: /.well-known/openid-configuration, /jwks
│       │   ├── userinfo_handler.go      # NEW: /oauth2/userinfo
│       │   ├── client_handler.go        # NEW: Admin client CRUD
│       │   └── role_handler.go          # NEW: Admin role management
│       ├── middleware/
│       │   ├── auth.go                  # Existing: JWT validation middleware
│       │   ├── rate_limiter.go          # NEW: Rate limiting middleware
│       │   ├── tenant_context.go        # NEW: Tenant identification middleware
│       │   └── cors.go                  # Existing: CORS handling
│       └── router.go    # Route definitions
├── migrations/
│   ├── 000001_create_tenants_table.up.sql    # Existing
│   ├── 000001_create_tenants_table.down.sql  # Existing
│   ├── 000002_create_users_table.up.sql      # Existing
│   ├── 000002_create_users_table.down.sql    # Existing
│   ├── 000003_create_audit_logs_table.up.sql # Existing
│   ├── 000003_create_audit_logs_table.down.sql # Existing
│   ├── 000004_create_clients_table.up.sql    # NEW
│   ├── 000004_create_clients_table.down.sql  # NEW
│   ├── 000005_create_user_role_assignments_table.up.sql # NEW
│   ├── 000005_create_user_role_assignments_table.down.sql # NEW
│   ├── 000006_create_refresh_tokens_table.up.sql # NEW
│   ├── 000006_create_refresh_tokens_table.down.sql # NEW
│   └── 000007_create_signing_keys_table.up.sql # NEW
│       000007_create_signing_keys_table.down.sql # NEW
├── cmd/
│   ├── server/
│   │   └── main.go      # Application entry point
│   ├── keygen/          # NEW
│   │   └── main.go      # NEW: RSA keypair generation utility
│   └── seed/            # NEW
│       └── main.go      # NEW: Test data seeder
├── tests/
│   ├── unit/
│   │   ├── domain/
│   │   │   ├── client_test.go           # NEW
│   │   │   ├── refresh_token_test.go    # NEW
│   │   │   └── user_role_test.go        # NEW
│   │   └── usecase/
│   │       ├── authorize_client_test.go # NEW
│   │       ├── issue_token_test.go      # NEW
│   │       ├── refresh_token_test.go    # NEW
│   │       ├── revoke_token_test.go     # NEW
│   │       ├── create_client_test.go    # NEW
│   │       └── assign_role_test.go      # NEW
│   ├── integration/
│   │   ├── auth_test.go                 # Existing
│   │   ├── registration_test.go         # Existing
│   │   ├── verification_test.go         # Existing
│   │   ├── oauth_auth_test.go           # NEW: Authorization flow
│   │   ├── oauth_token_test.go          # NEW: Token exchange
│   │   ├── oauth_refresh_test.go        # NEW: Refresh token flow
│   │   ├── oauth_revoke_test.go         # NEW: Token revocation
│   │   ├── discovery_test.go            # NEW: OIDC discovery
│   │   ├── client_management_test.go    # NEW: Client CRUD
│   │   └── role_management_test.go      # NEW: Role assignments
│   └── mocks/
│       ├── audit_repository.go          # Existing
│       ├── email_service.go             # Existing
│       ├── otp_repository.go            # Existing
│       ├── tenant_repository.go         # Existing
│       ├── user_repository.go           # Existing
│       ├── client_repository.go         # NEW
│       ├── auth_code_repository.go      # NEW
│       ├── session_repository.go        # NEW
│       ├── refresh_token_repository.go  # NEW
│       ├── signing_key_repository.go    # NEW
│       ├── role_repository.go           # NEW
│       ├── oauth_provider.go            # NEW
│       └── token_service.go             # NEW
├── go.mod
├── go.sum
├── .env.example         # NEW: Environment template
├── Makefile             # NEW: Common development commands
└── README.md            # Update with OAuth setup instructions

frontend/
├── src/
│   ├── components/
│   │   ├── auth/        # Existing: Login, registration
│   │   │   ├── LoginForm.tsx           # Existing
│   │   │   ├── ConsentScreen.tsx       # NEW: OAuth consent UI
│   │   │   └── OAuthCallback.tsx       # NEW: Callback handler
│   │   ├── admin/       # Admin portal components
│   │   │   ├── Dashboard.tsx           # Existing
│   │   │   ├── ClientManagement.tsx    # NEW: Client CRUD UI
│   │   │   ├── ClientForm.tsx          # NEW: Create/edit client
│   │   │   ├── RoleManagement.tsx      # NEW: Role assignment UI
│   │   │   └── UserRoles.tsx           # NEW: User role list
│   │   └── common/      # Shared components
│   │       ├── Button.tsx              # Existing
│   │       ├── Input.tsx               # Existing
│   │       └── Modal.tsx               # Existing
│   ├── services/        # API clients
│   │   ├── api.ts                      # Existing: Base API client
│   │   ├── authService.ts              # Existing: Auth API
│   │   ├── oauthService.ts             # NEW: OAuth flow API
│   │   ├── clientService.ts            # NEW: Client management API
│   │   └── roleService.ts              # NEW: Role management API
│   ├── hooks/           # Custom React hooks
│   │   ├── useAuth.ts                  # Existing
│   │   ├── useOAuth.ts                 # NEW: OAuth flow hook
│   │   └── usePKCE.ts                  # NEW: PKCE challenge generation
│   ├── contexts/        # React contexts
│   │   └── AuthContext.tsx             # Existing
│   ├── types/           # TypeScript definitions
│   │   ├── auth.ts                     # Existing
│   │   ├── oauth.ts                    # NEW: OAuth types
│   │   ├── client.ts                   # NEW: Client types
│   │   └── role.ts                     # NEW: Role types
│   ├── utils/           # Utility functions
│   │   ├── validators.ts               # Existing
│   │   ├── pkce.ts                     # NEW: PKCE helper
│   │   └── tokenStorage.ts             # NEW: Token storage helper
│   ├── App.tsx
│   ├── main.tsx
│   └── index.css
├── tests/
│   ├── unit/
│   │   ├── components/
│   │   │   ├── ConsentScreen.test.tsx  # NEW
│   │   │   ├── ClientManagement.test.tsx # NEW
│   │   │   └── RoleManagement.test.tsx # NEW
│   │   └── hooks/
│   │       ├── useOAuth.test.ts        # NEW
│   │       └── usePKCE.test.ts         # NEW
│   └── integration/
│       └── oauth-flow.test.tsx         # NEW
├── .env.example         # NEW: Frontend environment template
├── package.json
├── tsconfig.json
├── vite.config.ts
└── README.md            # Update with OAuth integration guide
```

**Structure Decision**: Web application structure selected (Option 2). Backend follows Clean Architecture with domain/usecase/infrastructure/interfaces separation. Frontend uses component-based React architecture with services layer for API abstraction. Existing backend structure preserved; new OAuth/OIDC entities, repositories, use cases, and handlers added alongside existing tenant registration logic. 7 new database tables (4 migrations). Frontend adds admin UI for client/role management and OAuth consent screen.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

**No violations identified.** The design fully complies with the constitution:

- Clean Architecture: Domain layer isolated from infrastructure, all dependencies point inward
- SOLID Principles: Single responsibility, dependency inversion through interfaces, proper abstraction
- Testing: Unit + integration test strategy defined with 85%+ coverage target
- Code Conventions: Go (Effective Go, lowercase packages) and TypeScript (strict mode, functional components) standards followed

No complexity tracking entries required.
