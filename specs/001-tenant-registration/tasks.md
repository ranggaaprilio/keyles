# Tasks: Multi-Tenant Registration with Email Verification

**Input**: Design documents from `/specs/001-tenant-registration/`
**Prerequisites**: plan.md (complete), spec.md (complete), research.md (complete), data-model.md (complete), contracts/api.yaml (complete)

**Tests**: Per constitution, tests are MANDATORY. All business logic requires unit tests (≥85% coverage), and all handlers require integration tests.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Web app (Clean Architecture)**: 
  - Backend: `backend/domain/`, `backend/usecase/`, `backend/infrastructure/`, `backend/interfaces/`
  - Frontend: `frontend/src/components/`, `frontend/src/services/`
  - Tests: `backend/tests/unit/`, `backend/tests/integration/`, `frontend/tests/`
- Migrations: `backend/migrations/`
- Email templates: `frontend/emails/templates/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create backend directory structure (domain/usecase/infrastructure/interfaces)
- [ ] T002 Initialize Go module in backend/ with go.mod (github.com/yourusername/keyles)
- [ ] T003 [P] Create frontend directory with Vite + React + TypeScript (npm create vite@latest frontend -- --template react-ts)
- [ ] T004 [P] Install backend dependencies (Gin, GORM, go-redis, Brevo SDK, bcrypt, testify, gomock)
- [ ] T005 [P] Install frontend dependencies (React 18, shadcn/ui, TanStack Query, Zod, React Email)
- [ ] T006 [P] Configure TypeScript strict mode in frontend/tsconfig.json
- [ ] T007 [P] Setup Tailwind CSS and shadcn/ui configuration in frontend/
- [ ] T008 Create docker-compose.yml with PostgreSQL 15, Redis 7, backend, frontend services
- [ ] T009 [P] Create backend/.env.example with all required variables
- [ ] T010 [P] Create frontend/.env.example with VITE_API_URL and VITE_APP_NAME
- [ ] T011 Create .gitignore files for backend/ and frontend/
- [ ] T012 Create README.md at repository root with project overview

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T013 Create database migration 000001_create_tenants_table.up.sql in backend/migrations/
- [ ] T014 Create database migration 000001_create_tenants_table.down.sql in backend/migrations/
- [ ] T015 [P] Create database migration 000002_create_users_table.up.sql in backend/migrations/
- [ ] T016 [P] Create database migration 000002_create_users_table.down.sql in backend/migrations/
- [ ] T017 [P] Create database migration 000003_create_audit_logs_table.up.sql in backend/migrations/
- [ ] T018 [P] Create database migration 000003_create_audit_logs_table.down.sql in backend/migrations/
- [ ] T019 Define Tenant entity in backend/domain/entities/tenant.go with validation methods
- [ ] T020 [P] Define AdminUser entity in backend/domain/entities/admin_user.go with validation methods
- [ ] T021 [P] Define OTPVerification entity in backend/domain/entities/otp_verification.go with validation
- [ ] T022 [P] Define AuditLog entity in backend/domain/entities/audit_log.go
- [ ] T023 Define TenantRepository interface in backend/domain/repositories/tenant_repository.go
- [ ] T024 [P] Define UserRepository interface in backend/domain/repositories/user_repository.go
- [ ] T025 [P] Define OTPRepository interface in backend/domain/repositories/otp_repository.go
- [ ] T026 [P] Define AuditRepository interface in backend/domain/repositories/audit_repository.go
- [ ] T027 Define EmailService interface in backend/domain/services/email_service.go
- [ ] T028 [P] Define OTPService interface in backend/domain/services/otp_service.go
- [ ] T029 [P] Define PasswordService interface in backend/domain/services/password_service.go
- [ ] T030 Implement PostgresTenantRepository in backend/infrastructure/persistence/postgres/tenant_repository.go
- [ ] T031 [P] Implement PostgresUserRepository in backend/infrastructure/persistence/postgres/user_repository.go
- [ ] T032 [P] Implement PostgresAuditRepository in backend/infrastructure/persistence/postgres/audit_repository.go
- [ ] T033 Implement RedisOTPRepository in backend/infrastructure/persistence/redis/otp_repository.go
- [ ] T034 [P] Implement BrevoEmailService in backend/infrastructure/services/brevo_email.go
- [ ] T035 [P] Implement CryptoOTPService in backend/infrastructure/services/crypto_otp.go
- [ ] T036 [P] Implement BcryptPasswordService in backend/infrastructure/services/bcrypt_password.go
- [ ] T037 Create configuration loader in backend/infrastructure/config/config.go
- [ ] T038 Setup CORS middleware in backend/interfaces/http/middleware/cors.go
- [ ] T039 [P] Setup rate limiting middleware in backend/interfaces/http/middleware/rate_limit.go
- [ ] T040 [P] Setup error handler middleware in backend/interfaces/http/middleware/error_handler.go (sanitize error messages to prevent information disclosure: hide stack traces, database constraint details, internal server details per FR-021)
- [ ] T041 Create Gin router setup in backend/interfaces/http/router.go
- [ ] T042 Create main.go entry point in backend/cmd/server/main.go with dependency injection
- [ ] T042b [P] Implement database transaction wrapper with SERIALIZABLE isolation for tenant creation in backend/infrastructure/persistence/postgres/transaction.go (prevents race conditions per FR-024)

**Clean Architecture Checkpoint**: 
- Domain layer established with interfaces only
- Infrastructure layer provides implementations
- No domain-to-infrastructure dependencies

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Tenant Registration Form (Priority: P1) 🎯 MVP

**Goal**: Organizations can register via a public form, creating a tenant with "pending verification" status

**Independent Test**: Can be fully tested by filling the form with valid data and verifying tenant record is created in database

### Tests for User Story 1 (MANDATORY per constitution) ⚠️

> **CONSTITUTION REQUIREMENT: Write these tests FIRST, ensure they FAIL before implementation**
> **Target: ≥85% coverage for domain/business logic, integration tests for all handlers**

- [X] T043 [P] [US1] Unit test for Tenant entity validation in backend/tests/unit/domain/tenant_test.go
- [X] T044 [P] [US1] Unit test for AdminUser entity validation in backend/tests/unit/domain/admin_user_test.go
- [X] T045 [P] [US1] Unit test for RegisterTenant use case in backend/tests/unit/usecase/register_tenant_test.go
- [X] T046 [P] [US1] Unit test for CheckAvailability use case in backend/tests/unit/usecase/check_availability_test.go
- [X] T047 [P] [US1] Integration test for registration handler in backend/tests/integration/registration_test.go
- [X] T048 [P] [US1] Integration test for availability handler in backend/tests/integration/availability_test.go

### Implementation for User Story 1

> **Clean Architecture Task Order**: Domain → Use Cases → Infrastructure → Interfaces (Handlers)

- [X] T049 [US1] Implement RegisterTenant use case in backend/usecase/tenant/register_tenant.go (orchestrates validation, uniqueness check, tenant creation, OTP generation)
- [X] T050 [US1] Implement CheckAvailability use case in backend/usecase/tenant/check_availability.go
- [X] T051 [US1] Implement registration handler POST /api/v1/register in backend/interfaces/http/handlers/registration_handler.go
- [X] T052 [US1] Implement availability handler GET /api/v1/check-availability in backend/interfaces/http/handlers/availability_handler.go
- [X] T053 [US1] Add request/response validation schemas in backend/interfaces/http/handlers/schemas.go
- [X] T054 [US1] Wire up registration and availability routes in backend/interfaces/http/router.go
- [X] T055 [P] [US1] Create RegistrationForm component in frontend/src/components/registration/RegistrationForm.tsx
- [X] T056 [P] [US1] Create Zod validation schema in frontend/src/components/registration/RegistrationSchema.ts
- [X] T057 [P] [US1] Create shadcn/ui form components (Input, Button, Label, Card) in frontend/src/components/ui/
- [X] T058 [US1] Create RegisterPage in frontend/src/pages/RegisterPage.tsx
- [X] T059 [P] [US1] Create tenant API client in frontend/src/services/api/tenant.ts
- [X] T060 [P] [US1] Create useTenantRegistration hook in frontend/src/hooks/useTenantRegistration.ts
- [X] T061 [P] [US1] Create TypeScript types in frontend/src/types/tenant.ts and frontend/src/types/api.ts
- [X] T062 [P] [US1] Create frontend unit tests for RegistrationForm in frontend/tests/unit/components/RegistrationForm.test.tsx

**Architecture Verification**:
- Domain has no infrastructure imports ✓
- Use cases depend only on domain interfaces ✓
- Handlers depend on use cases, not domain directly ✓

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently (tenant registration creates pending tenant)

---

## Phase 4: User Story 2 - Email OTP Verification (Priority: P2)

**Goal**: After registration, system sends OTP to email, user verifies within 10 minutes to activate tenant

**Independent Test**: Can be tested by registering a tenant (using US1), receiving OTP email, entering OTP, and confirming tenant status changes to "active"

### Tests for User Story 2 (MANDATORY per constitution) ⚠️

- [X] T063 [P] [US2] Unit test for OTPVerification entity in backend/tests/unit/domain/otp_verification_test.go
- [X] T064 [P] [US2] Unit test for VerifyTenant use case in backend/tests/unit/usecase/verify_tenant_test.go
- [X] T065 [P] [US2] Unit test for ResendOTP use case in backend/tests/unit/usecase/resend_otp_test.go
- [X] T066 [P] [US2] Integration test for OTP verification handler in backend/tests/integration/verification_test.go
- [X] T067 [P] [US2] Integration test for resend OTP handler in backend/tests/integration/resend_otp_test.go
- [X] T068 [P] [US2] Integration test for email sending (mock Brevo) in backend/tests/integration/email_test.go

### Implementation for User Story 2

- [X] T069 [P] [US2] Create OTP email template using React Email in frontend/emails/templates/OTPVerificationEmail.tsx
- [X] T070 [P] [US2] Create email layout component in frontend/emails/components/EmailLayout.tsx
- [X] T071 [P] [US2] Build and export email templates to HTML in frontend/emails/build/
- [X] T072 [US2] Implement VerifyTenant use case in backend/usecase/tenant/verify_tenant.go (validates OTP, updates tenant status, logs audit)
- [X] T073 [US2] Implement ResendOTP use case in backend/usecase/tenant/resend_otp.go (invalidates old OTP, generates new, sends email)
- [X] T074 [US2] Implement verification handler POST /api/v1/verify-otp in backend/interfaces/http/handlers/verification_handler.go
- [X] T075 [US2] Implement resend OTP handler POST /api/v1/resend-otp in backend/interfaces/http/handlers/resend_otp_handler.go
- [X] T076 [US2] Wire up verification and resend routes in backend/interfaces/http/router.go
- [ ] T077 [P] [US2] Create OTPVerificationForm component in frontend/src/components/verification/OTPVerificationForm.tsx
- [ ] T078 [P] [US2] Create ResendOTPButton component in frontend/src/components/verification/ResendOTPButton.tsx
- [ ] T079 [US2] Create VerifyOTPPage in frontend/src/pages/VerifyOTPPage.tsx
- [ ] T080 [P] [US2] Create useOTPVerification hook in frontend/src/hooks/useOTPVerification.ts
- [ ] T081 [P] [US2] Add OTP verification to API client in frontend/src/services/api/tenant.ts
- [ ] T082 [P] [US2] Create Toast notification component in frontend/src/components/ui/toast.tsx
- [ ] T083 [P] [US2] Create frontend unit tests for OTPVerificationForm in frontend/tests/unit/components/OTPVerificationForm.test.tsx

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently (registration + verification flow complete)

---

## Phase 5: User Story 3 - Post-Verification Tenant Access (Priority: P3)

**Goal**: Verified admin users can login and access tenant dashboard

**Independent Test**: Can be tested by verifying a tenant (using US1 and US2), logging in with credentials, and accessing dashboard

### Tests for User Story 3 (MANDATORY per constitution) ⚠️

- [ ] T084 [P] [US3] Unit test for AuthenticateAdmin use case in backend/tests/unit/usecase/authenticate_admin_test.go
- [ ] T085 [P] [US3] Integration test for login handler in backend/tests/integration/auth_test.go
- [ ] T086 [P] [US3] Integration test for dashboard handler (JWT protected) in backend/tests/integration/dashboard_test.go

### Implementation for User Story 3

- [ ] T087 [P] [US3] Create JWT token generation utility in backend/infrastructure/services/jwt_service.go
- [ ] T088 [US3] Implement AuthenticateAdmin use case in backend/usecase/auth/authenticate_admin.go (validates credentials, checks tenant status, generates JWT)
- [ ] T089 [US3] Implement auth middleware in backend/interfaces/http/middleware/auth.go (JWT validation)
- [ ] T090 [US3] Implement login handler POST /api/v1/login in backend/interfaces/http/handlers/auth_handler.go
- [ ] T091 [US3] Implement dashboard handler GET /api/v1/dashboard in backend/interfaces/http/handlers/dashboard_handler.go
- [ ] T092 [US3] Wire up auth and dashboard routes in backend/interfaces/http/router.go (dashboard requires auth middleware)
- [ ] T093 [P] [US3] Create LoginPage in frontend/src/pages/LoginPage.tsx
- [ ] T094 [P] [US3] Create DashboardPage in frontend/src/pages/DashboardPage.tsx
- [ ] T095 [P] [US3] Create TenantDashboard component in frontend/src/components/dashboard/TenantDashboard.tsx
- [ ] T096 [P] [US3] Create TenantInfo component in frontend/src/components/dashboard/TenantInfo.tsx
- [ ] T097 [P] [US3] Create auth API client in frontend/src/services/api/auth.ts
- [ ] T098 [P] [US3] Create useAuth hook in frontend/src/hooks/useAuth.ts (login, logout, token management)
- [ ] T099 [P] [US3] Setup React Router in frontend/src/App.tsx with routes (/, /register, /verify-otp, /login, /dashboard)
- [ ] T100 [P] [US3] Create protected route wrapper for dashboard in frontend/src/components/ProtectedRoute.tsx
- [ ] T101 [P] [US3] Create frontend unit tests for login and dashboard in frontend/tests/unit/

**Checkpoint**: All user stories should now be independently functional (complete registration → verification → login → dashboard flow)

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T102 [P] Add comprehensive logging to all use cases using structured logging (logrus or zap)
- [ ] T103 [P] Add health check endpoints in backend/interfaces/http/handlers/health_handler.go (/health, /health/db, /health/redis)
- [ ] T104 [P] Create API documentation using Swagger/OpenAPI annotations in handlers
- [ ] T105 [P] Add frontend error boundary component in frontend/src/components/ErrorBoundary.tsx
- [ ] T106 [P] Implement loading states and skeletons in frontend components
- [ ] T107 [P] Add frontend form accessibility (ARIA labels, keyboard navigation)
- [ ] T108 Verify ≥85% test coverage for backend domain and usecase layers (run go test -cover)
- [ ] T109 Verify all 6 handlers have integration tests passing
- [ ] T110 [P] Create Dockerfile for backend with multi-stage build
- [ ] T111 [P] Create Dockerfile for frontend (nginx serving static build)
- [ ] T112 Update docker-compose.yml with built images and health checks
- [ ] T113 [P] Add CI/CD workflow in .github/workflows/backend-ci.yml (test, lint, build)
- [ ] T114 [P] Add CI/CD workflow in .github/workflows/frontend-ci.yml (test, lint, build)
- [ ] T115 Architecture compliance review - verify no domain-to-infrastructure dependencies
- [ ] T116 Security review - verify bcrypt hashing, JWT secrets, rate limiting, CORS
- [ ] T117 [P] Create comprehensive README.md with setup instructions
- [ ] T118 Run all manual test scenarios from quickstart.md
- [ ] T119 Performance testing - verify <200ms p95 for registration, <100ms for verification
- [ ] T120 Database indexing review - ensure all indexes from data-model.md are created

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Phase 6)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Builds on US1 (registration must exist to send OTP)
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Requires US1 and US2 (login requires verified tenant)

### Within Each User Story

- Tests (mandatory) MUST be written and FAIL before implementation
- Domain entities and interfaces before use cases
- Use cases before handlers
- Handlers before frontend components
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

**Phase 1 (Setup)**:
- T003, T004, T005, T006, T007, T009, T010 can all run in parallel

**Phase 2 (Foundational)**:
- After migrations (T013-T018), all entity definitions (T019-T022) can run in parallel
- After entities, all interface definitions (T023-T029) can run in parallel
- After interfaces, all implementations (T030-T036) can run in parallel
- All middleware (T038-T040) can run in parallel

**User Story 1**:
- All unit tests (T043-T046) can run in parallel
- Frontend components (T055-T057, T059-T061) can run in parallel after backend API is ready

**User Story 2**:
- All unit tests (T063-T068) can run in parallel
- Email templates (T069-T071) can run in parallel with backend use cases
- Frontend components (T077-T078, T080-T082) can run in parallel

**User Story 3**:
- All unit tests (T084-T086) can run in parallel
- Frontend pages and components (T093-T096) can run in parallel

**Phase 6 (Polish)**:
- Most tasks (T102-T107, T110-T111, T113-T114, T117) can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Unit test for Tenant entity validation in backend/tests/unit/domain/tenant_test.go"
Task: "Unit test for AdminUser entity validation in backend/tests/unit/domain/admin_user_test.go"
Task: "Unit test for RegisterTenant use case in backend/tests/unit/usecase/register_tenant_test.go"
Task: "Unit test for CheckAvailability use case in backend/tests/unit/usecase/check_availability_test.go"

# After backend API is implemented, launch frontend tasks together:
Task: "Create RegistrationForm component in frontend/src/components/registration/RegistrationForm.tsx"
Task: "Create Zod validation schema in frontend/src/components/registration/RegistrationSchema.ts"
Task: "Create shadcn/ui form components in frontend/src/components/ui/"
Task: "Create tenant API client in frontend/src/services/api/tenant.ts"
Task: "Create TypeScript types in frontend/src/types/tenant.ts"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (12 tasks)
2. Complete Phase 2: Foundational (30 tasks)
3. Complete Phase 3: User Story 1 (20 tasks)
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

**MVP Result**: Organizations can register, tenant records are created, basic validation works

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready (42 tasks)
2. Add User Story 1 → Test independently → Deploy/Demo (20 tasks total: 62)
3. Add User Story 2 → Test independently → Deploy/Demo (21 tasks total: 83)
4. Add User Story 3 → Test independently → Deploy/Demo (18 tasks total: 101)
5. Polish phase → Production-ready (19 tasks total: 120)

Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (42 tasks)
2. Once Foundational is done:
   - **Developer A**: User Story 1 (registration form)
   - **Developer B**: User Story 2 (email OTP) - starts after US1 backend API exists
   - **Developer C**: User Story 3 (login/dashboard) - starts after US1 and US2 backend APIs exist
3. Stories complete and integrate independently

---

## Task Counts by Phase

- **Phase 1 (Setup)**: 12 tasks
- **Phase 2 (Foundational)**: 30 tasks
- **Phase 3 (User Story 1)**: 20 tasks
- **Phase 4 (User Story 2)**: 21 tasks
- **Phase 5 (User Story 3)**: 18 tasks
- **Phase 6 (Polish)**: 19 tasks

**Total**: 120 tasks

**Estimated Effort**:
- Setup: 1-2 days
- Foundational: 3-4 days
- User Story 1: 2-3 days
- User Story 2: 2-3 days
- User Story 3: 2 days
- Polish: 1-2 days

**Total Estimated Time**: 11-16 days (single developer) or 5-8 days (3 developers working in parallel)

---

## Notes

- **[P]** tasks = different files, no dependencies - can run in parallel
- **[Story]** label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing (TDD approach)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- All tasks follow Clean Architecture principles (domain → usecase → infrastructure → interfaces)
- Frontend tasks can start once corresponding backend APIs are implemented
- Database migrations must be tested with both up and down scripts
