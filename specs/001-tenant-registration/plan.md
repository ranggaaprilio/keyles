# Implementation Plan: Multi-Tenant Registration with Email Verification

**Branch**: `001-tenant-registration` | **Date**: 2025-12-06 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-tenant-registration/spec.md`

## Summary

Build a multi-tenant SaaS SSO platform's first feature: tenant registration with email verification. Organizations register via a public form, creating a tenant with "pending verification" status. System sends a 6-digit OTP to the admin email, which must be verified within 10 minutes to activate the tenant. Includes rate limiting, audit logging, and multi-tenant data isolation.

**Technical Approach**: Clean Architecture with Golang backend (Gin framework, PostgreSQL, Redis), React/TypeScript frontend (shadcn/ui, React Email), Brevo for email delivery, database migrations via golang-migrate.

## Technical Context

**Language/Version**: 
- Backend: Go 1.22+
- Frontend: TypeScript 5.3+ (strict mode)
- Node.js: 20.x LTS

**Primary Dependencies**:
- Backend: Gin (HTTP), GORM (ORM), golang-migrate (migrations), go-redis (caching), Brevo Go SDK (email)
- Frontend: React 18+, shadcn/ui (components), React Email (email templates), TanStack Query (API state), Zod (validation)

**Storage**: 
- Primary: PostgreSQL 15+ (multi-tenant isolation via tenant_id)
- Cache: Redis 7+ (OTP storage, rate limiting)

**Testing**:
- Backend: Go testing package + testify + gomock (mocking)
- Frontend: Vitest + React Testing Library
- E2E: Playwright (optional for integration tests)

**Target Platform**: 
- Backend: Linux server (Docker containerized)
- Frontend: Modern browsers (Chrome 100+, Firefox 100+, Safari 15+)
- Deployment: Web application (SPA + REST API)

**Project Type**: Web application (backend + frontend)

**Performance Goals**:
- API response time: <200ms p95 for registration, <100ms p95 for OTP verification
- Email delivery: <60 seconds for 99% of OTP emails
- Concurrent registrations: Handle 100 simultaneous requests without degradation
- Database queries: <50ms for tenant uniqueness checks

**Constraints**:
- Multi-tenant isolation: Zero cross-tenant data leakage (enforced at database and application level)
- Security: HTTPS only, bcrypt password hashing (cost 12), cryptographically secure OTP generation
- Rate limiting: Redis-based, 3 OTP requests/hour per email, 5 verification attempts per OTP
- OTP expiration: 10-minute TTL enforced via Redis TTL
- Session management: Stateless registration flow - backend returns tenant_id in registration response, frontend stores temporarily for OTP verification page. No server-side sessions required for registration-to-verification flow.
- GDPR compliance: Audit logging for all user actions, data retention policies (future)

**Scale/Scope**:
- Initial target: 1,000 tenants, 10,000 users across all tenants
- Registration volume: 100 new tenants per day
- Database: 4 core entities (Tenant, AdminUser, OTPVerification, AuditLog)
- API endpoints: 6 endpoints (register, verify OTP, resend OTP, check availability, login, dashboard)
- Frontend pages: 3 pages (registration form, OTP verification, tenant dashboard)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Clean Architecture Compliance**:
- [x] Domain layer has no imports from infrastructure or frameworks (entities, interfaces only)
- [x] All repository/service interfaces defined in Domain layer (TenantRepository, OTPService, EmailService, AuditLogger)
- [x] Concrete implementations only in Infrastructure layer (PostgresRepository, RedisOTPService, BrevoEmailService)
- [x] Dependency arrows verified to point inward (Handlers → Use Cases → Domain ← Infrastructure)

**SOLID Principles Compliance**:
- [x] Each module has single, well-defined responsibility (TenantUseCase for registration, OTPUseCase for verification)
- [x] Domain depends only on abstractions/interfaces (TenantRepository interface, not PostgreSQL)
- [x] No direct database/external API calls from business logic (all via interfaces)
- [x] Interface segregation verified for all contracts (separate interfaces for tenant ops, OTP ops, email ops)

**Testing Requirements**:
- [x] Unit test plan documented for all business logic (target: ≥85% coverage for domain and use case layers)
- [x] Integration test plan for all handlers/controllers (6 endpoint handlers)
- [x] Test isolation strategy defined (gomock for repositories, testcontainers for integration tests)
- [x] Test-first workflow feasible for this feature (TDD approach for domain entities and use cases)

**Code Conventions**:
- [x] Backend: Follows Effective Go, lowercase packages, exported function docs
- [x] Frontend: TypeScript strict mode, PascalCase components, functional components only
- [x] Clear separation between backend (domain/usecase/infrastructure) and frontend (components/services)

**Violations Requiring Justification** (fill Complexity Tracking section if any):
- [x] No constitution violations - all principles satisfied

**Violations Requiring Justification** (fill Complexity Tracking section if any):
- [x] No constitution violations - all principles satisfied

## Project Structure

### Documentation (this feature)

```text
specs/001-tenant-registration/
├── spec.md              # Feature specification (completed)
├── plan.md              # This file (implementation plan)
├── research.md          # Technology decisions and best practices
├── data-model.md        # Entity definitions and relationships
├── quickstart.md        # Local development and testing guide
├── contracts/           # API contracts (OpenAPI/REST)
│   └── api.yaml         # REST API specification
└── checklists/
    └── requirements.md  # Specification quality validation (completed)
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   └── server/
│       └── main.go                    # Application entry point
├── domain/                            # Clean Architecture: Innermost layer
│   ├── entities/
│   │   ├── tenant.go                  # Tenant entity with business rules
│   │   ├── admin_user.go              # Admin user entity
│   │   ├── otp_verification.go        # OTP verification entity
│   │   └── audit_log.go               # Audit log entity
│   ├── repositories/                  # Repository interfaces (abstractions)
│   │   ├── tenant_repository.go       # Interface for tenant persistence
│   │   ├── user_repository.go         # Interface for user persistence
│   │   ├── otp_repository.go          # Interface for OTP persistence
│   │   └── audit_repository.go        # Interface for audit logging
│   └── services/                      # Service interfaces (abstractions)
│       ├── email_service.go           # Interface for email operations
│       ├── otp_service.go             # Interface for OTP generation/validation
│       └── password_service.go        # Interface for password hashing
├── usecase/                           # Application business rules
│   ├── tenant/
│   │   ├── register_tenant.go         # RegisterTenant use case
│   │   ├── verify_tenant.go           # VerifyTenant use case
│   │   └── check_availability.go      # CheckAvailability use case
│   └── auth/
│       └── authenticate_admin.go      # AuthenticateAdmin use case
├── infrastructure/                    # Concrete implementations (outer layer)
│   ├── persistence/
│   │   ├── postgres/
│   │   │   ├── tenant_repository.go   # PostgreSQL tenant repo implementation
│   │   │   ├── user_repository.go     # PostgreSQL user repo implementation
│   │   │   └── audit_repository.go    # PostgreSQL audit repo implementation
│   │   └── redis/
│   │       └── otp_repository.go      # Redis OTP repo implementation
│   ├── services/
│   │   ├── brevo_email.go             # Brevo email service implementation
│   │   ├── crypto_otp.go              # Cryptographic OTP service
│   │   └── bcrypt_password.go         # Bcrypt password service
│   └── config/
│       └── config.go                  # Configuration management
├── interfaces/                        # Handlers, controllers (outer layer)
│   └── http/
│       ├── handlers/
│       │   ├── registration_handler.go    # POST /api/v1/register
│       │   ├── verification_handler.go    # POST /api/v1/verify-otp
│       │   ├── resend_otp_handler.go      # POST /api/v1/resend-otp
│       │   ├── availability_handler.go    # GET /api/v1/check-availability
│       │   ├── auth_handler.go            # POST /api/v1/login
│       │   └── dashboard_handler.go       # GET /api/v1/dashboard
│       ├── middleware/
│       │   ├── cors.go                    # CORS middleware
│       │   ├── auth.go                    # JWT authentication middleware
│       │   ├── rate_limit.go              # Rate limiting middleware
│       │   └── error_handler.go           # Error handling middleware
│       └── router.go                      # Route definitions
├── migrations/                        # Database migrations
│   ├── 000001_create_tenants_table.up.sql
│   ├── 000001_create_tenants_table.down.sql
│   ├── 000002_create_users_table.up.sql
│   ├── 000002_create_users_table.down.sql
│   ├── 000003_create_audit_logs_table.up.sql
│   └── 000003_create_audit_logs_table.down.sql
├── tests/
│   ├── unit/
│   │   ├── domain/
│   │   │   ├── tenant_test.go
│   │   │   ├── admin_user_test.go
│   │   │   └── otp_verification_test.go
│   │   └── usecase/
│   │       ├── register_tenant_test.go
│   │       └── verify_tenant_test.go
│   └── integration/
│       ├── registration_test.go
│       ├── verification_test.go
│       └── auth_test.go
├── go.mod
├── go.sum
├── Dockerfile
└── .env.example

frontend/
├── src/
│   ├── components/
│   │   ├── ui/                        # shadcn/ui components
│   │   │   ├── button.tsx
│   │   │   ├── input.tsx
│   │   │   ├── form.tsx
│   │   │   ├── card.tsx
│   │   │   └── toast.tsx
│   │   ├── registration/
│   │   │   ├── RegistrationForm.tsx   # Main registration form
│   │   │   └── RegistrationSchema.ts  # Zod validation schema
│   │   ├── verification/
│   │   │   ├── OTPVerificationForm.tsx
│   │   │   └── ResendOTPButton.tsx
│   │   └── dashboard/
│   │       ├── TenantDashboard.tsx
│   │       └── TenantInfo.tsx
│   ├── pages/
│   │   ├── RegisterPage.tsx           # /register route
│   │   ├── VerifyOTPPage.tsx          # /verify-otp route
│   │   ├── LoginPage.tsx              # /login route
│   │   └── DashboardPage.tsx          # /dashboard route
│   ├── services/
│   │   ├── api/
│   │   │   ├── client.ts              # Axios/Fetch API client
│   │   │   ├── tenant.ts              # Tenant API calls
│   │   │   └── auth.ts                # Auth API calls
│   │   └── validation/
│   │       └── schemas.ts             # Shared Zod schemas
│   ├── hooks/
│   │   ├── useTenantRegistration.ts   # Registration mutation hook
│   │   ├── useOTPVerification.ts      # OTP verification hook
│   │   └── useAuth.ts                 # Authentication hook
│   ├── types/
│   │   ├── tenant.ts                  # Tenant TypeScript types
│   │   ├── user.ts                    # User TypeScript types
│   │   └── api.ts                     # API response types
│   ├── lib/
│   │   └── utils.ts                   # Utility functions
│   ├── App.tsx
│   ├── main.tsx
│   └── vite-env.d.ts
├── emails/                            # React Email templates
│   ├── components/
│   │   └── EmailLayout.tsx
│   └── templates/
│       └── OTPVerificationEmail.tsx   # OTP email template
├── tests/
│   ├── unit/
│   │   ├── components/
│   │   │   ├── RegistrationForm.test.tsx
│   │   │   └── OTPVerificationForm.test.tsx
│   │   └── services/
│   │       └── api.test.ts
│   └── integration/
│       └── registration-flow.test.ts
├── public/
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
├── tailwind.config.js
└── .env.example

.github/
└── workflows/
    ├── backend-ci.yml                 # Backend CI/CD
    └── frontend-ci.yml                # Frontend CI/CD

docker-compose.yml                     # Local development environment
README.md
.gitignore
```

**Structure Decision**: Web application structure selected. Backend uses Clean Architecture with domain/usecase/infrastructure/interfaces separation. Frontend uses component-based architecture with shadcn/ui. Database migrations managed via golang-migrate. Redis used for OTP storage and rate limiting. Brevo SDK integrated in infrastructure layer for email delivery.
└── tests/
    ├── unit/
    │   ├── domain/
    │   └── usecase/
    ├── integration/
    └── contract/

frontend/
├── src/
│   ├── components/      # PascalCase React components
│   ├── services/        # API client abstractions
│   ├── types/           # TypeScript type definitions
│   └── hooks/           # Custom React hooks
└── tests/
    ├── unit/
    └── integration/

# [REMOVE IF UNUSED] Option 3: Mobile + API (when "iOS/Android" detected)
api/
└── [same as backend above]

ios/ or android/
└── [platform-specific structure: feature modules, UI flows, platform tests]
```

**Structure Decision**: [Document the selected structure and reference the real
directories captured above]

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
