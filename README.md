# Keyles - Multi-Tenant SSO Platform

OAuth 2.0 + OpenID Connect identity provider with multi-tenant isolation,
role-based access control, a browser-facing consent flow, and production
security controls. Built with Go, React, PostgreSQL, and Redis using Clean
Architecture.

## Project Status

Keyles is in active development. The core SSO platform and production
security hardening are implemented through Feature 007.

| Feature | Status |
|---------|--------|
| Tenant registration and email verification | Implemented |
| SSO landing page and tenant dashboard | Implemented |
| OAuth 2.0 and OpenID Connect provider | Implemented |
| OAuth client registration and secret rotation | Implemented |
| User management and client-scoped RBAC | Implemented |
| Browser login, consent, and logout flow | Implemented |
| Production security hardening | Implemented |

## Features

- **OAuth 2.0 + OIDC**: Authorization Code Flow with S256 PKCE, token
  exchange, refresh, revocation, and introspection (RFC 6749, 7009, 7662)
- **Browser Consent Flow**: Redis-backed authorization transactions,
  end-user login, consent approval/denial, and provider-local logout
- **Multi-Tenant**: Complete tenant isolation with organization-scoped
  users, clients, and roles
- **User Management & RBAC**: Invitation-based onboarding, user lifecycle
  (enable/disable/delete), role assignment per client, session listing,
  and audit activity feeds
- **OAuth Client Management**: Register, update, rotate secrets, and
  manage redirect URIs for OAuth client applications
- **JWT Tokens**: RS256 asymmetric signing with JWKS endpoint and OIDC
  discovery
- **Integration Guide**: Public and dashboard documentation for client
  registration, authorization, token handling, refresh, revocation, and
  production readiness
- **Production Security**: Caddy TLS termination, externalized secrets,
  startup configuration validation, security headers, double-submit CSRF,
  request limits, structured sanitized logging, and fail-closed rate limiting
- **Operations**: PostgreSQL connection limits and statement timeouts,
  dependency health checks, and Prometheus security metrics
- **Clean Architecture**: Domain-driven design with strict dependency
  inversion; domain layer has zero infrastructure imports

## Architecture

```
├── backend/                # Go API server (Clean Architecture)
│   ├── cmd/server/         # Application entry point
│   ├── domain/             # Entities, repository & service interfaces
│   ├── usecase/            # Application business rules
│   │   ├── auth/           # OAuth flows, consent, logout
│   │   ├── client/         # OAuth client management
│   │   ├── user/           # User lifecycle & invitations
│   │   ├── tenant/         # Tenant registration & verification
│   │   └── role/           # Role assignment & revocation
│   ├── infrastructure/     # PostgreSQL, Redis, RSA token service, config
│   ├── interfaces/http/    # Gin handlers, middleware, routing
│   ├── migrations/         # SQL migration files
│   └── tests/              # Unit, integration, and mock fixtures
├── frontend/               # React + TypeScript SPA (Vite)
│   └── src/
│       ├── components/     # Shared UI primitives (Radix + Tailwind)
│       ├── pages/          # Admin, OAuth interaction, and documentation pages
│       ├── services/       # API clients (auth, OAuth interaction, users)
│       ├── stores/         # Zustand state management
│       └── hooks/          # Shared React hooks
├── specs/                  # Feature specs, plans, data models, contracts
│   ├── 001-tenant-registration/
│   ├── 002-sso-landing-page/
│   ├── 003-sso-auth-provider/
│   ├── 004-client-app-registration/
│   ├── 005-user-management-rbac/
│   ├── 006-oauth-consent-flow/
│   └── 007-production-security-hardening/
├── docs/                   # Product and educational documentation
├── Caddyfile               # TLS termination and reverse proxy
├── docker-compose.yml      # Local service topology
└── docker-compose.prod.yml # Production security overrides
```

## Technology Stack

| Layer    | Technologies |
|----------|-------------|
| Backend  | Go 1.23, Gin 1.10, GORM, go-redis/v9, bcrypt, golang-migrate |
| Frontend | TypeScript 5.4, React 18.3, Vite 5, React Router 6, Axios, React Hook Form, Zod, Zustand, Tailwind CSS, Radix UI |
| Storage  | PostgreSQL 15, Redis 7 |
| Auth     | OAuth 2.0, OIDC, RS256 JWT, S256 PKCE |
| Testing  | Go testing + Testify, Vitest + React Testing Library |
| Security | Caddy TLS, CSRF protection, bcrypt, security headers, rate limiting |
| Observability | Structured `slog` JSON logs, health checks, Prometheus metrics |
| Infra    | Docker Compose, Caddy 2, nginx, multi-stage Docker builds |

## Quick Start

### Prerequisites

- Docker 20+ and Docker Compose 2+
- Go 1.23+ (for local backend development)
- Node.js 20 LTS (for local frontend development)

### Running with Docker Compose

```bash
git clone https://github.com/ranggaaprilio/keyles.git
cd keyles

# Copy and configure the Compose environment
cp .env.example .env
# Set POSTGRES_PASSWORD, DB_PASSWORD, JWT_SECRET, and other required values

# Start PostgreSQL, Redis, backend, frontend, and Caddy
docker compose up -d

# Run migrations
cd backend && make migrate-up && make seed

# Access the application
# Frontend:  http://localhost:3000
# Backend:   http://localhost:8080
# Health:    http://localhost:8080/health
# HTTPS:     https://localhost
# Metrics:   http://localhost:8080/metrics
```

### Running Locally

See [backend/README.md](backend/README.md) and [frontend/README.md](frontend/README.md) for detailed local development setup.

### Production Deployment

Production deployment uses the base Compose file plus hardened overrides:

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

Before deployment, configure strong secrets, HTTPS URLs, `ACME_EMAIL`, and a
TLS-enabled PostgreSQL connection. The backend rejects insecure production
configuration at startup. See the
[production hardening quickstart](specs/007-production-security-hardening/quickstart.md)
for the complete checklist.

## API Endpoints

### OAuth 2.0 / OIDC

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/oauth2/auth` | Browser authorization (redirects to login or consent) |
| `POST` | `/oauth2/login` | End-user credential authentication |
| `GET` | `/oauth2/consent/:transactionId` | Read consent details |
| `POST` | `/oauth2/consent` | Approve or deny consent |
| `POST` | `/oauth2/logout` | Terminate provider-local SSO session |
| `POST` | `/oauth2/token` | Token endpoint (code exchange, refresh) |
| `POST` | `/oauth2/revoke` | Token revocation (RFC 7009) |
| `POST` | `/oauth2/introspect` | Token introspection (RFC 7662) |
| `GET` | `/oauth2/userinfo` | User profile (Bearer token) |
| `GET` | `/.well-known/openid-configuration` | OIDC discovery |
| `GET` | `/.well-known/jwks.json` | JSON Web Key Set |

### Tenant Registration

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/register` | Register new tenant |
| `GET` | `/api/v1/check-availability` | Check email or org name availability |
| `POST` | `/api/v1/verify-otp` | Verify email OTP |
| `POST` | `/api/v1/resend-otp` | Resend verification OTP |

### Authentication

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/login` | Admin login (returns JWT) |
| `GET` | `/api/v1/dashboard` | Tenant dashboard (requires JWT) |
| `GET` | `/api/v1/invitations/:token/validate` | Validate a user invitation |
| `POST` | `/api/v1/invitations/:token/accept` | Accept an invitation and set a password |

### Admin: Client Management

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/admin/clients` | Register OAuth client |
| `GET` | `/api/v1/admin/clients` | List clients |
| `GET` | `/api/v1/admin/clients/:clientId` | Get client details |
| `PUT` | `/api/v1/admin/clients/:clientId` | Update client |
| `DELETE` | `/api/v1/admin/clients/:clientId` | Delete client |
| `POST` | `/api/v1/admin/clients/:clientId/rotate-secret` | Rotate client secret |

### Admin: User Management

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/admin/users` | List users |
| `POST` | `/api/v1/admin/users/invite` | Invite new user |
| `GET` | `/api/v1/admin/users/:id` | Get user details |
| `PATCH` | `/api/v1/admin/users/:id` | Update user |
| `PATCH` | `/api/v1/admin/users/:id/status` | Enable or disable user |
| `DELETE` | `/api/v1/admin/users/:id` | Delete user |
| `POST` | `/api/v1/admin/users/:id/resend-invitation` | Resend invitation |
| `GET` | `/api/v1/admin/users/:id/roles` | List user's role assignments |
| `POST` | `/api/v1/admin/users/:id/roles` | Assign role to user |
| `DELETE` | `/api/v1/admin/users/:id/roles/:assignmentId` | Revoke role |
| `GET` | `/api/v1/admin/users/:id/sessions` | List user's active sessions |
| `DELETE` | `/api/v1/admin/users/:id/sessions/:sessionId` | Revoke session |
| `GET` | `/api/v1/admin/users/:id/activity` | List user's audit activity |

### Admin: Role Management

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/admin/roles/assign` | Assign user role |
| `POST` | `/api/v1/admin/roles/revoke` | Revoke user role |
| `GET` | `/api/v1/admin/roles/users/:userId` | List roles for a user |
| `GET` | `/api/v1/admin/roles/clients/:clientId` | List roles for a client |

## Security

- **PKCE**: S256 mandatory for all authorization code flows
- **RS256**: Asymmetric JWT signing with 2048-bit RSA keys
- **TLS**: Caddy terminates HTTPS and manages certificates; production
  configuration requires HTTPS issuer/frontend URLs
- **Secrets**: Runtime environment variables replace committed credentials;
  weak production configuration is rejected during startup
- **CSRF**: Double-submit cookie validation protects state-changing API
  operations, with explicit exemptions for protocol and public auth endpoints
- **Security Headers**: CSP, HSTS, X-Frame-Options, X-Content-Type-Options,
  Permissions-Policy, COOP, COEP, and Referrer-Policy
- **Rate Limiting**: Redis sliding windows protect login, registration, OTP,
  and token endpoints and fail closed when Redis is unavailable
- **Browser SSO**: Host-only, HttpOnly, SameSite=Lax cookie; no Domain
  attribute; Path=/
- **Error and Log Safety**: Sanitized responses, masked email addresses and
  file paths, structured JSON logs, and no production debug logging
- **Request Hardening**: Body-size limits plus read, write, and idle timeouts
- **Database Security**: TLS required in production, bounded connection pools,
  30-second statement timeout, and hashed client secrets
- **Password Hashing**: bcrypt with cost factor 10
- **Metrics**: `/metrics` exposes `keyles_security_events_total` and security
  event duration metrics for external monitoring

## Testing

### Backend

```bash
cd backend
make test              # All tests
make test-coverage     # With coverage report
make test-compose-e2e  # Full OAuth browser matrix against Compose
```

### Frontend

```bash
cd frontend
npm run test           # Vitest with jsdom
npm run lint           # ESLint (zero warnings)
npm run build          # TypeScript + Vite production build
```

## Documentation

- [Feature 001: Tenant Registration](specs/001-tenant-registration/spec.md)
- [Feature 002: SSO Landing Page](specs/002-sso-landing-page/spec.md)
- [Feature 003: SSO Auth Provider](specs/003-sso-auth-provider/spec.md)
- [Feature 004: Client App Registration](specs/004-client-app-registration/spec.md)
- [Feature 005: User Management & RBAC](specs/005-user-management-rbac/spec.md)
- [Feature 006: OAuth Consent Flow](specs/006-oauth-consent-flow/spec.md)
- [Feature 007: Production Security Hardening](specs/007-production-security-hardening/spec.md)
- [OAuth/OIDC Integration Guide](http://localhost:3000/docs/oauth)
- [Understanding Single Sign-On](docs/articles/understanding-single-sign-on.md)
- [Backend README](backend/README.md)
- [Frontend README](frontend/README.md)

## Contributing

1. Read [AGENTS.md](AGENTS.md) for repository guidelines
2. Follow Clean Architecture: domain layer has zero infrastructure imports
3. Write tests before implementation (TDD)
4. Run `make test` and `npm run test` before submitting PRs
5. Use conventional commits: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`

---

**Version**: 0.3.0 | **Status**: Features 001-007 Implemented, Active Development | **Last Updated**: 2026-06-06
