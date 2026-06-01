# Keyles — Multi-Tenant SSO Platform

OAuth 2.0 + OpenID Connect identity provider with multi-tenant isolation,
role-based access control, and a browser-facing consent flow. Built in Go
and React with Clean Architecture.

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
- **Security**: Dual-key login throttling (source IP + tenant email),
  host-only HttpOnly SSO cookies, direct-peer source IP, PKCE mandatory,
  fail-closed Redis behavior, and bcrypt credential hashing
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
│       ├── pages/          # OAuth login, consent, logout, error pages
│       ├── services/       # API clients (auth, OAuth interaction, users)
│       ├── stores/         # Zustand state management
│       └── hooks/          # Shared React hooks
├── specs/                  # Feature specs, plans, data models, contracts
│   ├── 003-sso-auth-provider/
│   ├── 004-client-app-registration/
│   ├── 005-user-management-rbac/
│   └── 006-oauth-consent-flow/
└── docker-compose.yml      # PostgreSQL, Redis, backend, frontend
```

## Technology Stack

| Layer    | Technologies |
|----------|-------------|
| Backend  | Go 1.23, Gin 1.10, GORM, go-redis/v9, bcrypt, golang-migrate |
| Frontend | TypeScript 5.4, React 18.3, Vite 5, React Router 6, Axios, React Hook Form, Zod, Zustand, Tailwind CSS, Radix UI |
| Storage  | PostgreSQL 15, Redis 7 |
| Auth     | OAuth 2.0, OIDC, RS256 JWT, S256 PKCE |
| Testing  | Go testing + Testify, Vitest + React Testing Library |
| Infra    | Docker Compose, multi-stage Docker builds |

## Quick Start

### Prerequisites

- Docker 20+ and Docker Compose 2+
- Go 1.23+ (for local backend development)
- Node.js 20 LTS (for local frontend development)

### Running with Docker Compose

```bash
git clone https://github.com/ranggaaprilio/keyles.git
cd keyles

# Copy and configure environment
cp backend/.env.example backend/.env

# Start all services (PostgreSQL, Redis, backend, frontend)
docker compose up -d

# Run migrations
cd backend && make migrate-up && make seed

# Access the application
# Frontend:  http://localhost:3000
# Backend:   http://localhost:8080
# Health:    http://localhost:8080/health
```

### Running Locally

See [backend/README.md](backend/README.md) and [frontend/README.md](frontend/README.md) for detailed local development setup.

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
- **Rate Limiting**: Token endpoint throttled per client; dual-key
  (source IP + tenant email) fixed-window throttle for OAuth login
- **Browser SSO**: Host-only, HttpOnly, SameSite=Lax cookie; no Domain
  attribute; Path=/
- **Source IP**: Direct TCP peer address for throttling and audit;
  forwarded headers are not trusted
- **Fail-Closed**: Authorization, login, and consent operations return
  local errors during Redis outages; logout always expires the cookie
- **Password Hashing**: bcrypt with cost factor 10
- **HTTPS**: Required in production (`SECURITY_COOKIE_SECURE=true`)

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

- [Feature 003: SSO Auth Provider](specs/003-sso-auth-provider/spec.md)
- [Feature 004: Client App Registration](specs/004-client-app-registration/spec.md)
- [Feature 005: User Management & RBAC](specs/005-user-management-rbac/spec.md)
- [Feature 006: OAuth Consent Flow](specs/006-oauth-consent-flow/spec.md)
- [Backend README](backend/README.md)
- [Frontend README](frontend/README.md)

## Contributing

1. Read [AGENTS.md](AGENTS.md) for repository guidelines
2. Follow Clean Architecture: domain layer has zero infrastructure imports
3. Write tests before implementation (TDD)
4. Run `make test` and `npm run test` before submitting PRs
5. Use conventional commits: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`

## License

MIT License — see [LICENSE](LICENSE)

---

**Version**: 0.2.0 | **Status**: Active Development | **Last Updated**: 2026-06-01
