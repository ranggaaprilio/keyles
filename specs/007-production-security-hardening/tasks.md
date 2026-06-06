# Tasks: Production Security Hardening

**Input**: Design documents from `/specs/007-production-security-hardening/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/security-contracts.md, quickstart.md

**Tests**: Tests are mandatory. Write tests first, confirm they fail, then implement.
Maintain at least 85% coverage for domain and business logic and add integration
coverage for every backend handler/controller and frontend API integration point.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Web app**: `backend/src/`, `frontend/src/`
- Backend follows Clean Architecture: `domain/`, `usecase/`, `infrastructure/`, `interfaces/http/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and dependency additions

- [X] T001 Add `github.com/prometheus/client_golang` dependency to `backend/go.mod`
- [X] T002 [P] Create `backend/infrastructure/certs/dev-certs/` directory with `.gitkeep`
- [X] T003 [P] Create `backend/infrastructure/monitoring/` directory with `.gitkeep`
- [X] T004 Add `frontend/nginx.conf` security header placeholders: `add_header Content-Security-Policy "default-src 'self'";`, `add_header Strict-Transport-Security "max-age=31536000; includeSubDomains";`, `add_header Permissions-Policy "camera=(), microphone=()";`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core security infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T005 [P] Add CSRF config fields to `backend/infrastructure/config/config.go`: `CSRFEnabled`, `CSRFCookieName`, `CSRFHeaderName`, `CSRFTokenLength`
- [X] T006 [P] Add security headers config fields to `backend/infrastructure/config/config.go`: `SecurityHeadersCSP`, `SecurityHeadersHSTS`
- [X] T007 [P] Add request/DB/metrics config fields to `backend/infrastructure/config/config.go`: `RequestMaxBodySize`, `RequestReadTimeout`, `RequestWriteTimeout`, `RequestIdleTimeout`, `MetricsEnabled`, `MetricsPath`, `DBMaxOpenConns`, `DBMaxIdleConns`, `DBConnMaxLifetime`, `DBStatementTimeout`
- [X] T008 Extend `backend/.env.example` with all new environment variables and documentation
- [X] T009 [P] Create `backend/domain/entities/security_config.go` with `SecurityConfig` struct and production validation function (FR-004: reject default/weak secrets, enforce HTTPS URLs, DB_SSL_MODE require)
- [X] T010 [P] Create `backend/infrastructure/monitoring/metrics.go` with Prometheus registry, `keyles_security_events_total` counter (labels: failed_login, rate_limit_triggered, csrf_rejected, tls_error), `keyles_security_event_duration_seconds` histogram, and `/metrics` HTTP handler
- [X] T011 [P] Create `backend/infrastructure/certs/generate-dev-certs.sh` script for self-signed certificate generation using openssl
- [X] T012 Wire config validation into `backend/cmd/server/main.go`: call `SecurityConfig.Validate()` on startup, initialize Prometheus metrics registry, load dev certs if present
- [X] T013 [P] Add `client_max_body_size 1m;` to `frontend/nginx.conf` for request body size limit (FR-009)
- [X] T014 Configure `ReadTimeout`, `WriteTimeout`, `IdleTimeout` on Gin HTTP server in `backend/cmd/server/main.go` from config values (FR-010)

**Checkpoint**: Foundation ready — config extended, metrics registry created, cert generation script available, production validation active, request limits and timeouts configured

---

## Phase 3: User Story 1 - Secure Transport for All Traffic (Priority: P1) 🎯 MVP

**Goal**: Add TLS termination via Caddy reverse proxy in docker-compose, remove hardcoded secrets from docker-compose.yml, enforce HTTPS for production

**Independent Test**: `curl -k https://localhost/health` returns 200 over HTTPS; `curl http://localhost/health` redirects to HTTPS; no hardcoded secrets in committed docker-compose.yml

- [X] T015 [P] [US1] Update `docker-compose.yml`: remove hardcoded `POSTGRES_PASSWORD`, `DB_PASSWORD`, `JWT_SECRET` defaults; replace with `${VAR}` syntax referencing `.env`
- [X] T016 [P] [US1] Create `docker-compose.prod.yml` with production overrides: `DB_SSL_MODE=require`, `SECURITY_COOKIE_SECURE=true`, `GIN_MODE=release`, Caddy environment variables
- [X] T017 [P] [US1] Add Caddy service to `docker-compose.yml`: image `caddy:2`, ports 80/443, volumes for `Caddyfile`, `caddy_data`, `caddy_config`; route to backend:8080
- [X] T018 [US1] Create `Caddyfile` with TLS config: `tls internal` for dev, `tls { env ACME_EMAIL }` for prod; reverse proxy to backend; HTTP→HTTPS redirect
- [X] T019 [US1] Update `backend/.env.example` with `ACME_EMAIL`, `CADDY_ADMIN_API` variables
- [X] T020 [US1] Update `backend/infrastructure/config/config.go` production validation: reject `DB_SSL_MODE=disable` or `allow` when `APP_ENV=production`
- [X] T021 [US1] Update `backend/Makefile`: add `generate-secrets` target to create random secrets for `.env`
- [X] T022 [US1] Add `backend/infrastructure/certs/dev-certs/` to `.gitignore`
- [X] T023 [P] [US1] Integration test: verify TLS connection in `backend/tests/integration/tls_test.go` — assert HTTPS redirect, cert validity
- [X] T024 [US1] Update `quickstart.md` (at `specs/007-production-security-hardening/quickstart.md`) with Caddy TLS setup instructions

**Checkpoint**: TLS termination active, secrets externalized, production config validated, HTTPS enforced

---

## Phase 4: User Story 2 - Secrets Management Without Hardcoded Values (Priority: P1)

**Goal**: Ensure zero hardcoded secrets in committed files; backend rejects startup with default/weak secrets in production mode; secret rotation documented

**Independent Test**: `grep -r "password\|secret\|token" docker-compose.yml` returns only `${VAR}` references; backend fails to start with `JWT_SECRET=dev_jwt_secret_change_in_production` and `APP_ENV=production`

> **Note**: T015, T016, T020, T021 from US1 already cover the core secrets work. This phase adds validation tests and rotation documentation.

- [X] T025 [P] [US2] Create `backend/tests/integration/secrets_validation_test.go` — assert no hardcoded secrets in `docker-compose.yml`, `docker-compose.prod.yml`, `.env.example`
- [X] T026 [P] [US2] Create `backend/tests/integration/production_config_test.go` — assert backend rejects startup with each default/weak secret value when `APP_ENV=production`
- [X] T027 [US2] Add secret rotation documentation to `backend/README.md` under "Production Deployment" section
- [X] T028 [US2] Create `backend/infrastructure/config/config_test.go` test cases for production validation: JWT_SECRET length, DB_SSL_MODE, SECURITY_COOKIE_SECURE, OAUTH_ISSUER HTTPS, FRONTEND_URL HTTPS

**Checkpoint**: Secrets fully externalized, production validation tested, rotation documented

---

## Phase 5: User Story 3 - Security Headers on Every Response (Priority: P2)

**Goal**: Add CSP, HSTS, Permissions-Policy, COOP, COEP headers to all backend API responses and frontend nginx responses

**Independent Test**: `curl -I https://localhost/health` shows all 7 required security headers; `curl -I https://localhost:3000` shows matching headers from nginx

- [X] T029 [P] [US3] Create `backend/interfaces/http/middleware/security_headers.go` with Gin middleware: sets CSP, HSTS, X-Frame-Options, X-Content-Type-Options, Permissions-Policy, COOP, COEP, Referrer-Policy on every response
- [X] T030 [P] [US3] Create `backend/interfaces/http/middleware/security_headers_test.go` — unit test asserting all 7 headers present with correct values
- [X] T031 [US3] Wire `SecurityHeaders()` middleware into `backend/interfaces/http/router.go` — apply globally before route handlers
- [X] T032 [P] [US3] Update `frontend/nginx.conf` — add `add_header` directives for CSP, HSTS, Permissions-Policy, COOP, COEP (matching backend values)
- [X] T033 [US3] Integration test: `backend/tests/integration/security_headers_test.go` — make HTTP requests to multiple endpoints, assert all security headers present
- [X] T034 [US3] Update `backend/.env.example` with `SECURITY_HEADERS_CSP`, `SECURITY_HEADERS_HSTS` variables

**Checkpoint**: All HTTP responses include required security headers from both backend and frontend

---

## Phase 6: User Story 4 - CSRF Protection for State-Changing Operations (Priority: P2)

**Goal**: Implement CSRF token generation/validation middleware for backend; frontend includes CSRF token in all state-changing Axios requests; OAuth endpoints exempted

**Independent Test**: POST to `/api/v1/admin/clients` without CSRF token returns 403; POST with valid CSRF token succeeds; POST to `/oauth2/token` without CSRF token succeeds (exempt)

- [X] T035 [P] [US4] Create `backend/interfaces/http/middleware/csrf.go` — double-submit cookie pattern: generate 32-byte random token, set as `keyles_csrf` cookie (HttpOnly=false, SameSite=Strict, Secure, Path=/), validate `X-CSRF-Token` header matches cookie on POST/PUT/DELETE
- [X] T036 [P] [US4] Create `backend/interfaces/http/middleware/csrf_test.go` — unit test: missing token → 403, valid token → 200, mismatched token → 403, GET/HEAD/OPTIONS exempt
- [X] T037 [US4] Wire `CSRF()` middleware into `backend/interfaces/http/router.go` — apply globally with exempt paths: `/oauth2/auth`, `/oauth2/token`, `/oauth2/revoke`, `/oauth2/introspect`, `/health`, `/.well-known/*`, `/api/v1/register`, `/api/v1/check-availability`, `/api/v1/verify-otp`, `/api/v1/resend-otp`, `/api/v1/login`
- [X] T038 [P] [US4] Create `frontend/src/services/csrfService.ts` — extract CSRF token from `keyles_csrf` cookie, provide `getCsrfToken()` function
- [X] T039 [US4] Add Axios interceptor in `frontend/src/services/apiClient.ts` (or equivalent) — attach `X-CSRF-Token` header to all POST/PUT/DELETE requests
- [X] T040 [US4] Update `backend/infrastructure/config/config.go` with `CSRF_COOKIE_NAME`, `CSRF_HEADER_NAME`, `CSRF_TOKEN_LENGTH` config fields
- [X] T041 [P] [US4] Integration test: `backend/tests/integration/csrf_test.go` — test CSRF protection on admin endpoints, exemption on OAuth endpoints
- [X] T042 [US4] Update `backend/.env.example` with `CSRF_ENABLED`, `CSRF_COOKIE_NAME`, `CSRF_HEADER_NAME`, `CSRF_TOKEN_LENGTH` variables

**Checkpoint**: CSRF protection active on all state-changing API endpoints; OAuth endpoints exempted; frontend sends CSRF token automatically

---

## Phase 7: User Story 5 - Safe Error Messages and Logging (Priority: P2)

**Goal**: Sanitize all error responses (no stack traces, SQL, paths); mask PII in logs (email → `***@***`); enforce LOG_LEVEL filtering (debug never in production)

**Independent Test**: Trigger server error → response contains no stack traces/SQL/paths; check logs → no passwords/tokens/unmasked emails; set `LOG_LEVEL=info` → no debug output

- [X] T043 [P] [US5] Extend `backend/interfaces/http/middleware/error_handler.go` — add email regex masking to `sanitizeError()` (pattern: `.*@.*` → `***@***`); add internal file path masking (`/Users/`, `/home/`, `/app/` → `[REDACTED]`)
- [X] T044 [P] [US5] Extend `backend/interfaces/http/middleware/error_handler_test.go` — test cases: error with email → masked, error with SQL query → generic message, error with stack trace → generic message, error with file path → redacted
- [X] T045 [US5] Audit all `log.Printf` calls in `backend/` — replace with structured logging using Go `log/slog` (Go 1.21+); ensure no sensitive data logged
- [X] T046 [US5] Create `backend/infrastructure/logging/logger.go` — structured slog wrapper with level filtering; `LOG_LEVEL=debug` in dev, `LOG_LEVEL=info` in production
- [X] T047 [US5] Update `backend/cmd/server/main.go` — initialize structured logger with `LOG_LEVEL` from config; replace all `log.Printf` with `logger.Info/Warn/Error`
- [X] T048 [P] [US5] Integration test: `backend/tests/integration/error_sanitization_test.go` — trigger various errors, assert sanitized responses; assert logs contain no sensitive data
- [X] T049 [US5] Update `backend/infrastructure/config/config.go` production validation: reject `LOG_LEVEL=debug` when `APP_ENV=production`

**Checkpoint**: All error responses sanitized; logs contain no sensitive data; debug logging disabled in production

---

## Phase 8: User Story 6 - Rate Limiting on Public Endpoints (Priority: P3)

**Goal**: Extend rate limiting to login, registration, OTP verify, OTP resend, token exchange; fail-closed on Redis outage; rate limit headers on all responses

**Independent Test**: Rapid-fire requests to `/api/v1/login` → 429 after 5 attempts with `X-RateLimit-*` headers; Redis down → requests rejected with 503

- [X] T050 [P] [US6] Extend `backend/interfaces/http/middleware/rate_limit.go` — add sliding window implementation using Redis sorted sets (`ZADD`, `ZREMRANGEBYSCORE`, `ZCARD`); change Redis failure behavior from `c.Next()` (fail-open) to reject with 503 (fail-closed)
- [X] T051 [P] [US6] Extend `backend/interfaces/http/middleware/rate_limit_test.go` — test sliding window accuracy, fail-closed behavior on Redis error, rate limit header presence
- [X] T052 [US6] Wire rate limiting to new endpoints in `backend/interfaces/http/router.go`: `/api/v1/login` (5/15min), `/api/v1/register` (3/1hr), `/api/v1/verify-otp` (5/10min), `/api/v1/resend-otp` (3/1hr)
- [X] T053 [US6] Update `backend/infrastructure/config/config.go` with per-endpoint rate limit config: `RateLimitRegisterPerHour`, `RateLimitVerifyOTPPer10Min`, `RateLimitResendOTPPerHour`
- [X] T054 [P] [US6] Integration test: `backend/tests/integration/rate_limit_extended_test.go` — test rate limiting on each new endpoint, assert 429 response with headers, assert fail-closed on Redis error
- [X] T055 [US6] Update `backend/.env.example` with new rate limit variables

**Checkpoint**: Rate limiting active on all public endpoints; fail-closed on Redis outage; proper rate limit headers

---

## Phase 9: User Story 7 - Database Connection Security (Priority: P3)

**Goal**: Enforce TLS for database connections in production; set connection pool limits; add statement timeout; hash/encrypt sensitive columns at rest

**Independent Test**: Production config → `DB_SSL_MODE=require` enforced; connection pool capped at 25 open/10 idle; long-running queries killed after 30s

- [X] T056 [P] [US7] Update `backend/infrastructure/persistence/postgres/helpers.go` (or database initialization) — set `DBMaxOpenConns=25`, `DBMaxIdleConns=10`, `DBConnMaxLifetime=5m` from config
- [X] T057 [P] [US7] Add statement timeout: execute `SET statement_timeout = '30s'` on each new database connection in `backend/infrastructure/persistence/postgres/helpers.go`
- [X] T058 [US7] Verify `backend/infrastructure/config/config.go` production validation: `DB_SSL_MODE` must be `require`, `verify-ca`, or `verify-full` when `APP_ENV=production`
- [X] T059 [US7] Audit `backend/migrations/` — verify client secrets are hashed (bcrypt or similar) in database
- [X] T060 [US7] If audit finds plaintext client secrets, create migration `000015_hash_client_secrets.up.sql` with matching `.down.sql` to bcrypt existing values
- [X] T061 [P] [US7] Integration test: `backend/tests/integration/database_security_test.go` — assert TLS connection in production config, assert connection pool limits, assert statement timeout kills long queries
- [X] T062 [US7] Update `backend/.env.example` with `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME`, `DB_STATEMENT_TIMEOUT` variables

**Checkpoint**: Database connections use TLS in production; pool limits enforced; statement timeout active; sensitive columns hashed

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T063 [P] Wire Prometheus metrics counter increments into: `error_handler.go` (failed_login, tls_error), `rate_limit.go` (rate_limit_triggered), `csrf.go` (csrf_rejected)
- [X] T064 [P] Update `backend/README.md` with production deployment checklist including all new security features
- [X] T065 [P] Run `make test` — ensure all new tests pass; run `make test-coverage` — verify 85%+ domain coverage
- [X] T066 [P] Run `npm run lint` and `npm run test` in `frontend/` — ensure CSRF interceptor doesn't break existing tests
- [X] T067 Run quickstart.md validation: execute all verification steps from `specs/007-production-security-hardening/quickstart.md`
- [X] T068 [P] Add `docker-compose.yml` and `docker-compose.prod.yml` to AGENTS.md security section

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **User Stories (Phase 3-9)**: All depend on Foundational phase completion
  - US1 (P1) and US2 (P1) are tightly coupled — US2 validation tests depend on US1 secrets externalization
  - US3 (P2), US4 (P2), US5 (P2) can proceed in parallel after Foundational
  - US6 (P3), US7 (P3) can proceed in parallel after Foundational
- **Polish (Phase 10)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational — No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational — depends on US1 docker-compose changes for validation tests
- **User Story 3 (P2)**: Can start after Foundational — No dependencies on other stories
- **User Story 4 (P2)**: Can start after Foundational — No dependencies on other stories
- **User Story 5 (P2)**: Can start after Foundational — No dependencies on other stories
- **User Story 6 (P3)**: Can start after Foundational — extends existing rate_limit.go (no story dependencies)
- **User Story 7 (P3)**: Can start after Foundational — No dependencies on other stories

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Config changes before middleware implementation
- Middleware before router wiring
- Core implementation before integration tests
- Story complete before moving to next priority

### Parallel Opportunities

- **Phase 1**: T002, T003 can run in parallel
- **Phase 2**: T005, T006, T007, T009, T010, T011, T013 can run in parallel
- **Phase 3 (US1)**: T015, T016, T017 can run in parallel; T023 can run in parallel with T018-T022
- **Phase 4 (US2)**: T025, T026 can run in parallel
- **Phase 5 (US3)**: T029, T030, T032 can run in parallel
- **Phase 6 (US4)**: T035, T036 can run in parallel; T038 can run in parallel with T037
- **Phase 7 (US5)**: T043, T044 can run in parallel; T048 can run in parallel with T045-T047
- **Phase 8 (US6)**: T050, T051 can run in parallel
- **Phase 9 (US7)**: T056, T057, T061 can run in parallel
- **Phase 10**: T063, T064, T065, T066, T068 can run in parallel

### Parallel Example: User Story 3 (Security Headers)

```bash
# Launch all parallel tasks together:
Task: "Create security_headers.go middleware in backend/interfaces/http/middleware/security_headers.go"
Task: "Create security_headers_test.go unit tests in backend/interfaces/http/middleware/security_headers_test.go"
Task: "Update frontend/nginx.conf with security headers in frontend/nginx.conf"

# After parallel tasks complete, wire and integrate:
Task: "Wire SecurityHeaders() middleware into backend/interfaces/http/router.go"
Task: "Integration test in backend/tests/integration/security_headers_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
3. Complete Phase 3: User Story 1 (TLS + secrets externalization)
4. **STOP and VALIDATE**: Test HTTPS, verify no hardcoded secrets, run production config validation
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add US1 (TLS + secrets) → Test independently → Deploy/Demo (MVP!)
3. Add US2 (secrets validation) → Test independently → Deploy/Demo
4. Add US3 (security headers) → Test independently → Deploy/Demo
5. Add US4 (CSRF) → Test independently → Deploy/Demo
6. Add US5 (error sanitization) → Test independently → Deploy/Demo
7. Add US6 (rate limiting) → Test independently → Deploy/Demo
8. Add US7 (database security) → Test independently → Deploy/Demo
9. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: US1 (TLS) → US2 (secrets validation)
   - Developer B: US3 (headers) → US4 (CSRF)
   - Developer C: US5 (logging) → US6 (rate limiting)
   - Developer D: US7 (database security)
3. Stories complete and integrate independently
