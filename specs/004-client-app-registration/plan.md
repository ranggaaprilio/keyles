# Implementation Plan: OAuth Client Application Registration Portal

**Branch**: `004-client-app-registration` | **Date**: 2026-02-25 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/004-client-app-registration/spec.md`

## Summary

Implement a tenant-administrator-facing OAuth client application registration portal as the dedicated realization of 003-sso-auth-provider User Story 1. This feature extends the existing backend client CRUD (already scaffolded in `usecase/client/`, `interfaces/http/handlers/client_handler.go`, and `infrastructure/persistence/postgres/client_repository.go`) with: **public/confidential client type distinction**, **25-client-per-tenant quota enforcement**, **token revocation on client deletion**, **audit logging**, **pagination/search**, and a full React frontend dashboard with one-time secret display, inline documentation, and secret regeneration UX. Redis is used for caching client counts per tenant and rate limiting.

## Technical Context

**Language/Version**: Go 1.23 (backend), TypeScript 5.4 (frontend)
**Primary Dependencies**: Gin (HTTP), GORM (ORM), go-redis v9, React 18, TanStack Query v5, Zustand, Axios, Zod, Vite, Tailwind CSS, shadcn/ui (Radix)
**Storage**: PostgreSQL (persistent client data, audit logs) + Redis (tenant client count cache, rate limiting)
**Testing**: Go `testing` + `testify` + GoMock (backend); Vitest + React Testing Library (frontend)
**Target Platform**: Linux server (Docker), modern browsers
**Project Type**: Web application (backend + frontend)
**Performance Goals**: Client list API <200ms for 25 clients; registration portal usable by 100 concurrent admins per tenant
**Constraints**: Max 25 clients per tenant; client_secret displayed once then never stored in plain text; tenant isolation enforced at all layers
**Scale/Scope**: Multi-tenant SaaS; each tenant ≤25 clients; typical tenant 3-10 clients

## Constitution Check

_GATE: Must pass before Phase 0 research. Re-check after Phase 1 design._

**Clean Architecture Compliance**:

- [x] Domain layer has no imports from infrastructure or frameworks
- [x] All repository/service interfaces defined in Domain layer
- [x] Concrete implementations only in Infrastructure layer
- [x] Dependency arrows verified to point inward (toward Domain)

**Verified**: New `ClientType` field added to existing `entities.Client` (domain). `ClientRepository` interface (domain) extended with `CountByTenant()`. `AuditLogRepository` interface (domain) already exists. All concrete implementations (bcrypt hashing, Redis caching, PostgreSQL queries) remain in infrastructure layer. Use cases depend only on domain interfaces.

**SOLID Principles Compliance**:

- [x] Each module has single, well-defined responsibility (SRP)
- [x] Domain depends only on abstractions/interfaces (DIP)
- [x] No direct database/external API calls from business logic
- [x] Interface segregation verified for all contracts

**Verified**: Each use case has a single responsibility (CreateClient, UpdateClient, DeleteClient, etc.). New `CountByTenant` method added to existing `ClientRepository` interface rather than a new interface (ISP — single cohesive repository). Domain entities have no infrastructure imports. Quota enforcement logic lives in use case layer, not domain.

**Testing Requirements**:

- [x] Unit test plan documented for all business logic (target: ≥85% coverage)
- [x] Integration test plan for all handlers/controllers
- [x] Test isolation strategy defined (mocking approach)
- [x] Test-first workflow feasible for this feature

**Verified**: Unit tests for all 6 use cases using GoMock-generated mocks for `ClientRepository`, `PasswordService`, `AuditLogRepository`. Integration tests for all handler endpoints using httptest. Frontend uses Vitest + RTL with MSW for API mocking. Test-first feasible: interfaces defined first, then tests, then implementations.

**Code Conventions**:

- [x] Backend: Follows Effective Go, lowercase packages, exported function docs
- [x] Frontend: TypeScript strict mode, PascalCase components, functional components only
- [x] Clear separation between backend (domain/usecase/infrastructure) and frontend (components/services)

**Verified**: Existing codebase conventions followed: lowercase Go packages (`client`, `handlers`, `postgres`), exported docs on all public types/functions, PascalCase React components, functional components with hooks, services abstracted through API layer.

**Violations Requiring Justification** (fill Complexity Tracking section if any):

- [x] No constitution violations OR all violations documented with justification

## Project Structure

### Documentation (this feature)

```text
specs/004-client-app-registration/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   ├── openapi.yaml     # OpenAPI 3.0 contract for client management endpoints
│   └── README.md        # Contract documentation
└── tasks.md             # Phase 2 output (/speckit.tasks - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
backend/
├── domain/
│   ├── entities/
│   │   └── client.go              # MODIFY: Add ClientType field, update Validate()
│   ├── repositories/
│   │   └── client_repository.go   # MODIFY: Add CountByTenant(), ListByTenantPaginated()
│   └── services/
│       └── (no changes needed)
├── usecase/
│   └── client/
│       ├── create_client.go       # MODIFY: Add client type support, quota check, audit log
│       ├── update_client.go       # MODIFY: Add audit logging
│       ├── delete_client.go       # MODIFY: Add token revocation, audit logging
│       ├── rotate_secret.go       # MODIFY: Add audit logging, public client guard
│       ├── get_client.go          # MODIFY: Add client_type to response
│       └── list_clients.go        # MODIFY: Add pagination, search, client_type
├── infrastructure/
│   ├── persistence/
│   │   ├── postgres/
│   │   │   └── client_repository.go  # MODIFY: CountByTenant(), paginated list, client_type column
│   │   └── redis/
│   │       └── client_count_cache.go # NEW: Tenant client count cache
│   └── services/
│       └── (no changes needed)
├── interfaces/
│   └── http/
│       ├── handlers/
│       │   └── client_handler.go     # MODIFY: Add client_type to request/response, pagination params
│       └── router.go                 # MODIFY: (no route changes needed — routes already exist)
├── migrations/
│   ├── 000008_add_client_type_and_description.up.sql    # NEW: ALTER TABLE clients ADD client_type, description
│   └── 000008_add_client_type_and_description.down.sql  # NEW: Rollback
└── tests/
    ├── unit/
    │   └── client/                   # MODIFY: Add tests for new behaviors
    └── integration/
        └── client_handler_test.go    # MODIFY: Add tests for new endpoints/behaviors

frontend/
├── src/
│   ├── components/
│   │   ├── clients/
│   │   │   ├── ClientList.tsx            # NEW: Client list with search, pagination
│   │   │   ├── ClientCard.tsx            # NEW: Client summary card
│   │   │   ├── ClientDetail.tsx          # NEW: Client detail view
│   │   │   ├── CreateClientForm.tsx      # NEW: Registration form with type selection
│   │   │   ├── EditClientForm.tsx        # NEW: Edit client configuration
│   │   │   ├── SecretDisplay.tsx         # NEW: One-time secret display with copy
│   │   │   ├── RotateSecretDialog.tsx    # NEW: Confirmation dialog for secret rotation
│   │   │   ├── DeleteClientDialog.tsx    # NEW: Confirmation dialog for deletion
│   │   │   └── IntegrationDocs.tsx       # NEW: Inline OAuth flow documentation
│   │   └── ui/
│   │       └── (existing shadcn/ui components)
│   ├── pages/
│   │   └── ClientManagementPage.tsx      # NEW: Main client management page
│   ├── hooks/
│   │   └── useClients.ts                 # NEW: TanStack Query hooks for client CRUD
│   ├── services/
│   │   └── clientService.ts              # MODIFY: Add client_type, pagination params
│   ├── types/
│   │   └── client.ts                     # MODIFY: Add ClientType, pagination types
│   └── utils/
│       └── (no changes needed)
└── tests/
    └── unit/
        ├── components/
        │   └── clients/                  # NEW: Component unit tests
        └── hooks/
            └── useClients.test.ts        # NEW: Hook tests
```

**Structure Decision**: Web application (Option 2) — matches existing project layout exactly. No new layers or directories needed at the architecture level. Changes are additive to existing files and new files within established patterns.

## Complexity Tracking

> No constitution violations found. All changes follow existing Clean Architecture patterns.
