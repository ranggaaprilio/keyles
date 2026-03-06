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

- Go 1.21+
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
│   ├── auth/            # OAuth flows
│   ├── client/          # Client management
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

### Authorization Flow

- `GET /oauth2/auth` - Authorization endpoint (user authentication)
- `POST /oauth2/token` - Token endpoint (code exchange, refresh)
- `POST /oauth2/revoke` - Token revocation endpoint

### Discovery & Validation

- `GET /.well-known/openid-configuration` - OIDC discovery document
- `GET /.well-known/jwks.json` - JSON Web Key Set (public keys)
- `GET /oauth2/userinfo` - User profile endpoint

### Admin Endpoints

- `POST /api/admin/clients` - Register OAuth client
- `GET /api/admin/clients` - List clients
- `GET /api/admin/clients/:id` - Get client details
- `PUT /api/admin/clients/:id` - Update client
- `DELETE /api/admin/clients/:id` - Delete client
- `POST /api/admin/clients/:id/rotate-secret` - Rotate client secret

- `POST /api/admin/roles/assign` - Assign user role
- `POST /api/admin/roles/revoke` - Revoke user role
- `GET /api/admin/roles/users/:userId` - List user roles

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
make dev               # Full dev setup (docker + migrate + seed + run)
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

### 1. Manual Authorization Request

```bash
# Visit in browser (generates PKCE challenge first)
open "http://localhost:8080/oauth2/auth?client_id=dev_client_001&redirect_uri=http://localhost:3000/auth/callback&response_type=code&scope=openid%20profile%20email&state=random123&code_challenge=E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM&code_challenge_method=S256"
```

### 2. Token Exchange

```bash
# After receiving authorization code from callback
curl -X POST http://localhost:8080/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "code=AUTH_CODE_HERE" \
  -d "redirect_uri=http://localhost:3000/auth/callback" \
  -d "code_verifier=dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk" \
  -d "client_id=dev_client_001" \
  -d "client_secret=dev_client_secret_change_in_production"
```

### 3. JWKS Endpoint

```bash
# Fetch public keys for token validation
curl http://localhost:8080/.well-known/jwks.json
```

## Architecture

This project follows **Clean Architecture** principles:

- **Domain Layer**: Pure business logic, no external dependencies
- **Use Case Layer**: Application-specific business rules
- **Infrastructure Layer**: Database, Redis, external services
- **Interfaces Layer**: HTTP handlers, middleware

**Dependency Rule**: All dependencies point inward toward the domain.

## Security

- **PKCE**: Mandatory for all authorization flows
- **RS256**: Asymmetric JWT signing with 2048-bit RSA keys
- **Rate Limiting**: 10 requests/minute per client_id on token endpoint
- **Token Expiration**: Access tokens 15min, Refresh tokens 7 days
- **RBAC**: Role-based access control for client applications
- **Tenant Isolation**: Complete separation between tenants

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

- [Feature Specification](../specs/003-sso-auth-provider/spec.md)
- [Implementation Plan](../specs/003-sso-auth-provider/plan.md)
- [Quickstart Guide](../specs/003-sso-auth-provider/quickstart.md)
- [Data Model](../specs/003-sso-auth-provider/data-model.md)
- [API Contracts](../specs/003-sso-auth-provider/contracts/openapi.yaml)
- [Tasks](../specs/003-sso-auth-provider/tasks.md)

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

**Rollback**: To remove Feature 005 changes, migrate down 4 steps:

```bash
make migrate-down STEPS=4
# This removes migrations 000012, 000011, 000010, 000009
```

## License

[Your License Here]
