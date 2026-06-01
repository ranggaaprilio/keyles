# Keyles Backend - OAuth 2.0 + OIDC SSO Provider

Multi-tenant SSO authentication service implementing OAuth 2.0 Authorization Code Flow with PKCE and OpenID Connect.

## Features

- **OAuth 2.0 + OIDC**: Full Authorization Code Flow with PKCE
- **Multi-Tenant**: Complete tenant isolation with role-based access control
- **JWT Tokens**: RS256 asymmetric signing with JWKS endpoint
- **Security**: Rate limiting, PKCE mandatory, session management
- **Clean Architecture**: Domain-driven design with dependency inversion

## Quick Start

See [specs/003-sso-auth-provider/quickstart.md](../specs/003-sso-auth-provider/quickstart.md) for comprehensive setup guide.

### Prerequisites
- Go 1.23+
- PostgreSQL 14+
- Redis 7+
- golang-migrate CLI

### Installation

```bash
# Clone the repository
cd backend

# Install dependencies
go mod download

# Copy environment file
cp .env.example .env

# Edit .env with your configuration
nano .env

# Generate RSA keypair for JWT signing
make keygen

# Start Docker services (PostgreSQL + Redis)
make docker-up

# Run migrations
make migrate-up

# Seed test data
make seed

# Run the server
make run
```

## OAuth Configuration

### Environment Variables

Key OAuth-specific variables in `.env`:

```bash
# OAuth/OIDC Configuration
OAUTH_ISSUER=http://localhost:8080
OAUTH_ACCESS_TOKEN_TTL=900          # 15 minutes
OAUTH_REFRESH_TOKEN_TTL=604800      # 7 days
OAUTH_AUTH_CODE_TTL=300             # 5 minutes

# JWT Signing Keys
JWT_SIGNING_KEY_PATH=./keys/private.pem
JWT_PUBLIC_KEY_PATH=./keys/public.pem
JWT_KEY_ID=dev_key_001

# Rate Limiting
RATE_LIMIT_TOKEN_ENDPOINT=10        # requests per minute per client_id
```

### RSA Key Generation

Generate RSA keypair for JWT signing:

```bash
# Generate 2048-bit keys (default)
make keygen

# Or generate 4096-bit keys
go run cmd/keygen/main.go -size 4096

# Keys will be created in ./keys/
# - private.pem (keep secure, never commit!)
# - public.pem (distributed via JWKS endpoint)
```

**⚠️ Important**: Add `keys/` directory to `.gitignore` to prevent committing private keys!

### Test Data Seeding

The seed command creates:

- Dev tenant (`tenant_dev`)
- Admin user: `admin@dev-tenant.com` / `admin123`
- Regular user: `user@dev-tenant.com` / `user123`
- OAuth client: `dev_client_001` with redirect URIs
- Role assignments for both users

```bash
make seed
```

## Project Structure

```
backend/
├── cmd/
│   ├── server/          # Main application entry point
│   ├── keygen/          # RSA keypair generator
│   └── seed/            # Test data seeder
├── domain/              # Business logic & interfaces
│   ├── entities/        # Domain entities
│   ├── repositories/    # Repository interfaces
│   └── services/        # Service interfaces
├── usecase/             # Application business rules
│   ├── auth/            # OAuth flows & browser interaction
│   ├── client/          # Client management
│   ├── user/            # User lifecycle management
│   ├── tenant/          # Tenant registration
│   └── role/            # RBAC management
├── infrastructure/      # External implementations
│   ├── config/          # Configuration
│   ├── persistence/     # Repository implementations
│   │   ├── postgres/    # PostgreSQL
│   │   └── redis/       # Redis
│   └── services/        # External services
├── interfaces/          # HTTP handlers (outer layer)
│   └── http/
│       ├── handlers/    # Request handlers
│       ├── middleware/  # HTTP middleware
│       └── router.go    # Route definitions
├── migrations/          # Database migrations
└── tests/
    ├── unit/            # Unit tests
    ├── integration/     # Integration tests
    └── mocks/           # Mock implementations
```

## OAuth Endpoints

### Browser Authorization Flow (Feature 006)

- `GET /oauth2/auth` - Authorization endpoint (redirects to frontend login or consent)
- `POST /oauth2/login` - End-user credential authentication for OAuth
- `GET /oauth2/consent/:transactionId` - Read consent details for authenticated session
- `POST /oauth2/consent` - Approve or deny consent, redirect to callback
- `POST /oauth2/logout` - Terminate provider-local browser SSO session

### Token & Revocation

- `POST /oauth2/token` - Token endpoint (authorization_code, refresh_token grants)
- `POST /oauth2/revoke` - Token revocation (RFC 7009)
- `POST /oauth2/introspect` - Token introspection (RFC 7662)

### Discovery & Validation

- `GET /.well-known/openid-configuration` - OIDC discovery document
- `GET /.well-known/jwks.json` - JSON Web Key Set (public keys)
- `GET /oauth2/userinfo` - User profile endpoint (requires Bearer token)

### Admin Endpoints: Client Management

- `POST /api/v1/admin/clients` - Register OAuth client
- `GET /api/v1/admin/clients` - List clients
- `GET /api/v1/admin/clients/:clientId` - Get client details
- `PUT /api/v1/admin/clients/:clientId` - Update client
- `DELETE /api/v1/admin/clients/:clientId` - Delete client
- `POST /api/v1/admin/clients/:clientId/rotate-secret` - Rotate client secret

### Admin Endpoints: Role Management

- `POST /api/v1/admin/roles/assign` - Assign user role
- `POST /api/v1/admin/roles/revoke` - Revoke user role
- `GET /api/v1/admin/roles/users/:userId` - List roles for a user
- `GET /api/v1/admin/roles/clients/:clientId` - List roles for a client

### Admin Endpoints: User Management (Feature 005)

- `GET /api/v1/admin/users` - List users
- `POST /api/v1/admin/users/invite` - Invite a new user
- `GET /api/v1/admin/users/:id` - Get user details
- `PATCH /api/v1/admin/users/:id` - Update user
- `PATCH /api/v1/admin/users/:id/status` - Enable or disable user
- `DELETE /api/v1/admin/users/:id` - Delete user
- `POST /api/v1/admin/users/:id/resend-invitation` - Resend invitation email
- `GET /api/v1/admin/users/:id/roles` - List user's role assignments
- `POST /api/v1/admin/users/:id/roles` - Assign role to user
- `DELETE /api/v1/admin/users/:id/roles/:assignmentId` - Revoke role from user
- `GET /api/v1/admin/users/:id/sessions` - List user's active sessions
- `DELETE /api/v1/admin/users/:id/sessions/:sessionId` - Revoke a user session
- `GET /api/v1/admin/users/:id/activity` - List user's audit activity
## Development

### Available Make Commands

```bash
make help              # Show all available commands
make build             # Build the server binary
make run               # Run the server
make test              # Run all tests
make test-coverage     # Generate coverage report
make migrate-up        # Run database migrations
make migrate-down      # Rollback last migration
make keygen            # Generate RSA keypair
make seed              # Seed test data
make clean             # Remove build artifacts
make docker-up         # Start Docker services
make docker-down       # Stop Docker services
make dev-infra         # Start local Docker dependencies
make dev-db            # Run migrations and seed data
make dev               # Full dev setup (infra + db + run)
make test-docker       # Run integration tests with Docker dependencies
make test-compose-e2e  # Run the live OAuth browser matrix against Compose stack
make docker-build      # Build Docker image for native architecture
make docker-buildx     # Build multi-arch image (amd64+arm64) and push
```

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Unit tests only
go test ./domain/... ./usecase/...

# Integration tests (requires PostgreSQL + Redis)
go test ./tests/integration/...
```

## Testing OAuth Flow

### 1. Start Authorization Request

The browser-facing `/oauth2/auth` endpoint validates the request and redirects
the end-user to the frontend login page via an opaque transaction:

```bash
# Generate PKCE challenge first, then visit:
open "http://localhost:8080/oauth2/auth?client_id=dev_client_001&redirect_uri=http://localhost:3000/auth/callback&response_type=code&scope=openid%20profile%20email&state=random123&code_challenge=E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM&code_challenge_method=S256"
# → Redirects to http://localhost:3000/oauth2/login?transaction_id=...
```

### 2. Log In at the Frontend

The frontend login page at `/oauth2/login` sends credentials to `POST /oauth2/login`:

```bash
curl -X POST http://localhost:8080/oauth2/login \
  -H "Content-Type: application/json" \
  -d '{"transaction_id":"...","email":"user@dev-tenant.com","password":"user123"}'
# → Returns consent URL; sets host-only HttpOnly keyles_sso cookie
```

### 3. Approve or Deny Consent

```bash
# Read consent details
curl http://localhost:8080/oauth2/consent/:transactionId \
  -H "Cookie: keyles_sso=..."

# Approve consent
curl -X POST http://localhost:8080/oauth2/consent \
  -H "Content-Type: application/json" \
  -H "Cookie: keyles_sso=..." \
  -d '{"transaction_id":"...","csrf_token":"...","approved":true}'
# → Returns callback URL with authorization code and original state
```

### 4. Token Exchange

```bash
curl -X POST http://localhost:8080/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "code=AUTH_CODE_HERE" \
  -d "redirect_uri=http://localhost:3000/auth/callback" \
  -d "code_verifier=dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk" \
  -d "client_id=dev_client_001" \
  -d "client_secret=dev_client_secret_change_in_production"
```

### 5. End Provider Session

```bash
curl -X POST http://localhost:8080/oauth2/logout \
  -H "Cookie: keyles_sso=..." \
  -v
```

### 6. JWKS Endpoint

```bash
curl http://localhost:8080/.well-known/jwks.json
```

See [quickstart.md](../specs/006-oauth-consent-flow/quickstart.md) for the
complete browser-flow verification matrix.

## Architecture

This project follows **Clean Architecture** principles:

- **Domain Layer**: Pure business logic, no external dependencies
- **Use Case Layer**: Application-specific business rules
- **Infrastructure Layer**: Database, Redis, external services
- **Interfaces Layer**: HTTP handlers, middleware

**Dependency Rule**: All dependencies point inward toward the domain.

## Security

- **PKCE**: S256 mandatory for all authorization flows
- **RS256**: Asymmetric JWT signing with 2048-bit RSA keys
- **Rate Limiting**: 10 requests/minute per client_id on token endpoint; dual-key (source IP + tenant email) fixed-window throttle for OAuth login (5 failures / 15 min)
- **Token Expiration**: Access tokens 15 min, Refresh tokens 7 days, Auth codes 5 min
- **RBAC**: Role-based access control for client applications; active-user and role revalidated at each consent
- **Tenant Isolation**: Complete separation between tenants
- **Browser SSO**: Host-only HttpOnly SameSite=Lax cookie with Path=/; no Domain attribute
- **Source IP**: Direct TCP peer address for throttling and audit; forwarded headers not trusted
- **Fail-Closed**: Authorization, login, consent, and Redis-dependent operations return local errors during infrastructure outages; logout always expires the cookie

## Production Deployment

Before deploying to production:

1. Set `SERVER_ENV=production`
2. Use HTTPS (`SECURITY_COOKIE_SECURE=true`)
3. Generate production RSA keys (4096-bit recommended)
4. Configure strong database passwords
5. Enable Redis authentication
6. Set proper `OAUTH_ISSUER` (production domain)
7. Configure rate limiting on load balancer
8. Set up database backups
9. Configure logging aggregation
10. Review and test disaster recovery plan

See [quickstart.md Production Deployment Checklist](../specs/003-sso-auth-provider/quickstart.md#production-deployment-checklist).

## Documentation

- [Feature 003: SSO Auth Provider](../specs/003-sso-auth-provider/spec.md)
- [Feature 005: User Management & RBAC](../specs/005-user-management-rbac/spec.md)
- [Feature 006: OAuth Consent Flow](../specs/006-oauth-consent-flow/spec.md)
- [Feature 006: Implementation Plan](../specs/006-oauth-consent-flow/plan.md)
- [Feature 006: Quickstart Guide](../specs/006-oauth-consent-flow/quickstart.md)
- [Feature 006: Data Model](../specs/006-oauth-consent-flow/data-model.md)
- [Feature 006: API Contracts](../specs/006-oauth-consent-flow/contracts/openapi.yaml)
- [Feature 006: Tasks](../specs/006-oauth-consent-flow/tasks.md)

## User Management & RBAC (Feature 005)

### Additional Environment Variables

| Variable                       | Description                                                           | Default  |
| ------------------------------ | --------------------------------------------------------------------- | -------- |
| `INVITATION_BASE_URL`          | Base URL for invitation links (e.g. `https://app.example.com/invite`) | Required |
| `BREVO_INVITATION_TEMPLATE_ID` | Brevo transactional email template ID for invitation emails           | Required |

### Cron Jobs

Add the following cron entries for maintenance tasks:

```cron
# Expire stale invitations (hourly)
0 * * * * /path/to/cleanup --expire-invitations

# Purge old user events (daily at 2 AM)
0 2 * * * /path/to/cleanup --purge-user-events
```

### Redis Key Namespaces

| Pattern                     | Purpose                           | TTL                    |
| --------------------------- | --------------------------------- | ---------------------- |
| `user_blacklist:<user_id>`  | Disabled user session blacklist   | Until re-enabled       |
| `user_count:<tenant_id>`    | Cached user count per tenant      | 5 min                  |
| `invitation_exists:<token>` | Invitation token validation cache | Matches invitation TTL |

### Database Migrations

Feature 005 adds migrations 000009–000012:

| Migration                             | Description                                                    |
| ------------------------------------- | -------------------------------------------------------------- |
| `000009_extend_users`                 | Add `display_name`, `status`, `last_login_at`; trigram indexes |
| `000010_create_invitations`           | Invitation tracking with token, expiry, status                 |
| `000011_extend_user_role_assignments` | Partial index for active role lookups                          |
| `000012_create_user_events`           | Audit event log with composite indexes                         |

Feature 006 adds migrations 000013–000014:

| Migration                             | Description                                         |
| ------------------------------------- | --------------------------------------------------- |
| `000013_oauth_security_hardening`     | Additional OAuth audit indexes and constraints      |
| `000014_add_oauth_audit_event_types`  | Extend audit event enum with OAuth browser-flow types |

**Rollback**: To remove Feature 005 changes, migrate down 4 steps:

```bash
make migrate-down STEPS=4
# This removes migrations 000012, 000011, 000010, 000009
```

## OAuth Browser Consent Flow (Feature 006)

The browser-facing authorization flow stores validated OAuth request data in a
short-lived Redis transaction before redirecting to the frontend. The frontend
receives only an opaque `transaction_id`; client IDs, callback URIs, state, PKCE
values, and consent CSRF state remain server-controlled.

### Browser Flow Environment Variables

| Variable                                  | Description                                      | Default                 |
| ----------------------------------------- | ------------------------------------------------ | ----------------------- |
| `FRONTEND_URL`                            | Frontend base URL for login, consent, and errors | `http://localhost:3000` |
| `SECURITY_COOKIE_SECURE`                  | Require HTTPS for the SSO cookie                 | `false`                 |
| `SECURITY_SESSION_TTL`                    | Browser SSO session TTL in seconds               | `28800`                 |
| `OAUTH_AUTH_TRANSACTION_TTL`              | Authorization transaction TTL in seconds         | `600`                   |
| `RATE_LIMIT_OAUTH_LOGIN_FAILURES`         | Maximum failures per login bucket                | `5`                     |
| `RATE_LIMIT_OAUTH_LOGIN_WINDOW_SECONDS`   | Fixed login failure window in seconds            | `900`                   |

The `keyles_sso` cookie is host-only, `HttpOnly`, `SameSite=Lax`, and uses
`Path=/`. Do not add a cookie `Domain` setting.

### Browser Flow Security Notes

- OAuth login throttling uses both source-IP and tenant-scoped normalized-email
  Redis buckets.
- Source IP is read from the direct TCP peer. Forwarded IP headers are ignored
  until a trusted-proxy allowlist is implemented.
- Authorization initialization, login, consent reads, and consent decisions fail
  closed when Redis is unavailable.
- `POST /oauth2/logout` always expires the browser cookie, even when Redis session
  deletion fails. It does not revoke tokens issued to external clients.

### Browser Flow Redis Keys

| Pattern                                             | Purpose                         | Default TTL |
| --------------------------------------------------- | ------------------------------- | ----------- |
| `oauth:transaction:<transaction_id>`                | Validated authorization request | 10 min      |
| `oauth:session:<session_id>`                        | End-user browser SSO session    | 8 hours     |
| `oauth:login-failure:ip:<source_ip>`                | Source-IP login throttle        | 15 min      |
| `oauth:login-failure:email:<tenant_id>:<email_hash>`| Tenant-email login throttle     | 15 min      |

## License

[Your License Here]
