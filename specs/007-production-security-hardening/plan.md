# Implementation Plan: Production Security Hardening

**Branch**: `007-production-security-hardening` | **Date**: 2026-06-06 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/007-production-security-hardening/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Harden the Keyles SSO platform for production by adding TLS termination via Caddy reverse proxy, removing hardcoded secrets from docker-compose, implementing CSRF protection middleware, adding comprehensive security headers (CSP, HSTS, Permissions-Policy, COOP, COEP), extending existing rate limiting to all public endpoints, sanitizing error responses and logging, adding request size limits/timeouts, enforcing database TLS connections, and exposing a `/metrics` endpoint with security-relevant counters.

## Technical Context

**Language/Version**: Go 1.23, TypeScript 5.4  
**Primary Dependencies**: Gin 1.10, go-redis/v9, GORM, Caddy (Docker), prometheus/client_golang  
**Storage**: PostgreSQL 15 (TLS required), Redis 7  
**Testing**: Go testing + Testify, Vitest + React Testing Library  
**Target Platform**: Linux server (Docker Compose)  
**Project Type**: Web service (backend API + frontend SPA)  
**Performance Goals**: Rate-limited endpoints reject within 100ms; /metrics updates within 1s; no degradation at 100 concurrent users  
**Constraints**: Must not break existing OAuth flow; CSRF must exempt OAuth `state`-protected endpoints; fail-closed on Redis outage for rate limiting  
**Scale/Scope**: 10k users, 7 user stories, 20 functional requirements, 11 success criteria

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Clean Architecture Compliance

- [x] Domain layer has no imports from infrastructure or frameworks — Security middleware (CSRF, headers, rate limiting) lives in `interfaces/http/middleware/`, not domain
- [x] All repository/service interfaces defined in Domain — No new domain interfaces needed; security is cross-cutting concern at interface layer
- [x] Concrete implementations only in Infrastructure layer — Caddy TLS config, Prometheus metrics service in `infrastructure/`
- [x] Dependency arrows point inward — CSRF token generation uses domain-agnostic crypto; headers middleware is pure HTTP layer

### SOLID Compliance

- [x] Each module has single, well-defined responsibility — Each security middleware is a separate file
- [x] Domain depends only on abstractions (interfaces) — No domain changes required
- [x] No direct database/external API calls from business logic — Security headers and CSRF are handler-layer concerns

### Testing Requirements

- [x] Unit test plan for all business logic (target: 85%+ coverage) — CSRF middleware, security headers, metrics counters
- [x] Integration test plan for all handlers/controllers — Rate limiting on all public endpoints, error sanitization
- [x] Test isolation strategy documented — Mock Redis for rate limiting tests; mock HTTP client for header verification

### No Violations — Standard Clean Architecture compliance maintained.

## Project Structure

### Documentation (this feature)

```text
specs/007-production-security-hardening/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
backend/
├── cmd/server/
│   └── main.go                          # Add metrics server init, CSRF setup
├── interfaces/http/
│   ├── middleware/
│   │   ├── security_headers.go          # NEW: CSP, HSTS, Permissions-Policy, COOP, COEP
│   │   ├── csrf.go                      # NEW: CSRF token generation/validation
│   │   ├── request_limits.go            # NEW: Body size limits, timeouts
│   │   ├── rate_limit.go                # EXTEND: Add more endpoints, fail-closed
│   │   └── error_handler.go             # EXTEND: Stricter sanitization, PII masking
│   ├── handlers/
│   │   └── metrics_handler.go           # NEW: /metrics endpoint with security counters
│   └── router.go                        # Wire new middleware, exempt OAuth endpoints
├── infrastructure/
│   ├── config/
│   │   └── config.go                    # EXTEND: Add CSRF, TLS, metrics config vars
│   ├── monitoring/
│   │   └── metrics.go                   # NEW: Prometheus registry, security counters
│   └── certs/
│       └── dev-certs/                   # NEW: Self-signed cert generation script
├── domain/
│   └── entities/
│       └── security_config.go           # NEW: Security configuration validation
└── tests/
    ├── middleware/
    │   ├── csrf_test.go                 # NEW
    │   ├── security_headers_test.go     # NEW
    │   └── request_limits_test.go       # NEW
    └── integration/
        ├── rate_limit_extended_test.go  # NEW
        └── error_sanitization_test.go   # NEW

frontend/
├── src/
│   ├── services/
│   │   └── csrfService.ts               # NEW: CSRF token extraction, Axios interceptor
│   └── components/
│       └── SecurityMeta.tsx             # NEW: CSP nonce handling if needed
└── nginx/
    └── nginx.conf                       # EXTEND: Add CSP, HSTS, Permissions-Policy headers

docker-compose.yml                       # EXTEND: Add Caddy service, remove hardcoded secrets
docker-compose.prod.yml                  # NEW: Production overrides
.env.example                             # EXTEND: Add all new env vars
```

**Structure Decision**: Web application pattern (Option 2) — backend API + frontend SPA. Security hardening touches both layers: backend middleware for CSRF/headers/rate-limiting, frontend for CSRF token inclusion in Axios, nginx for additional headers, and Docker Compose for TLS/secrets.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Caddy reverse proxy (new container) | TLS termination with auto-renewal | Nginx TLS requires manual cert management; Traefik is heavier for single-service use |
| Prometheus metrics library | Standard format for external monitoring | Custom JSON endpoint would require operators to build their own parsers |
| CSRF double-submit cookie pattern | No server-side session storage needed | Server-side CSRF tokens require Redis storage per token, adding complexity |
