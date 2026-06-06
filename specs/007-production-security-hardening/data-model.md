# Data Model: Production Security Hardening

## Overview

This feature does not introduce new database entities. Security hardening is implemented through middleware, configuration, and infrastructure changes. The following configuration entities and validation rules are defined.

## Configuration Entities

### SecurityConfig

Runtime security configuration loaded from environment variables.

| Field | Type | Validation | Default |
|-------|------|------------|---------|
| `AppEnv` | string | Must be `development`, `staging`, or `production` | `development` |
| `CSRFEnabled` | bool | — | `true` in production |
| `CSRFCookieName` | string | Non-empty, no whitespace | `keyles_csrf` |
| `CSRFHeaderName` | string | Non-empty | `X-CSRF-Token` |
| `CSRFTokenLength` | int | 16–64 bytes | `32` |
| `SecurityHeadersCSP` | string | Valid CSP directive string | `default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'` |
| `SecurityHeadersHSTS` | string | Must contain `max-age=` with value ≥ 31536000 | `max-age=31536000; includeSubDomains` |
| `RequestMaxBodySize` | int | > 0 | `1048576` (1MB) |
| `RequestReadTimeout` | duration | > 0 | `15s` |
| `RequestWriteTimeout` | duration | > 0 | `15s` |
| `RequestIdleTimeout` | duration | > 0 | `60s` |
| `DBSSLMode` | string | Production: must be `require`, `verify-ca`, or `verify-full` | `require` (production), `disable` (dev) |
| `DBMaxOpenConns` | int | > 0 | `25` |
| `DBMaxIdleConns` | int | > 0 | `10` |
| `DBConnMaxLifetime` | duration | > 0 | `5m` |
| `DBStatementTimeout` | duration | > 0 | `30s` |
| `MetricsEnabled` | bool | — | `true` |
| `MetricsPath` | string | Valid URL path | `/metrics` |
| `LogLevel` | string | Must be `debug`, `info`, `warn`, or `error` | `info` (production), `debug` (dev) |

### RateLimitConfig

Per-endpoint rate limiting configuration.

| Endpoint | Default Limit | Window | Redis Key Pattern |
|----------|--------------|--------|-------------------|
| `/api/v1/login` | 5 requests | 15 min | `ratelimit:login:{ip}:{email_hash}` |
| `/api/v1/register` | 3 requests | 1 hour | `ratelimit:register:{ip}` |
| `/api/v1/verify-otp` | 5 attempts | 10 min | `ratelimit:otp:verify:{ip}` |
| `/api/v1/resend-otp` | 3 requests | 1 hour | `ratelimit:otp:resend:{ip}` |
| `/oauth2/token` | 10 requests | 1 min | `ratelimit:token:{client_id}` |

### CSRF Exempt Paths

Paths exempt from CSRF validation (use alternative protection mechanisms).

| Path | Reason |
|------|--------|
| `/oauth2/auth` | Uses `state` parameter for CSRF protection |
| `/oauth2/token` | OAuth token endpoint — uses client_secret + PKCE |
| `/oauth2/revoke` | Uses bearer token authentication |
| `/oauth2/introspect` | Uses bearer token authentication |
| `/health` | Health check — no state change |
| `/.well-known/*` | OIDC discovery — public read-only |
| `/api/v1/register` | Tenant registration — no authenticated session |
| `/api/v1/check-availability` | Public read-only |
| `/api/v1/verify-otp` | OTP verification — no authenticated session |
| `/api/v1/resend-otp` | OTP resend — no authenticated session |
| `/api/v1/login` | Admin login — no authenticated session yet |

## Validation Rules

### Production Mode Validation (FR-004)

When `APP_ENV=production`, the backend MUST reject startup if:
- `JWT_SECRET` equals the default value `dev_jwt_secret_change_in_production`
- `JWT_SECRET` is shorter than 32 characters
- `DB_PASSWORD` is empty
- `DB_SSL_MODE` is `disable` or `allow`
- `BREVO_API_KEY` is empty
- `SECURITY_COOKIE_SECURE` is `false`
- `OAUTH_ISSUER` does not use HTTPS
- `FRONTEND_URL` does not use HTTPS

### Security Headers Validation (FR-005 to FR-007)

All HTTP responses MUST include:
- `Content-Security-Policy` with `script-src` restriction
- `Strict-Transport-Security` with `max-age >= 31536000`
- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `Permissions-Policy` with restricted features
- `Cross-Origin-Opener-Policy: same-origin`
- `Cross-Origin-Embedder-Policy: require-corp`

### Error Sanitization Validation (FR-011 to FR-012)

Error responses MUST NOT contain:
- Stack traces
- SQL queries or database connection strings
- Internal file paths (e.g., `/Users/`, `/home/`, `/app/`)
- Passwords, tokens, or secrets (even partial)
- Unmasked email addresses (must be `***@***` format)

## State Transitions

No state transitions — this feature adds cross-cutting security concerns, not entity lifecycle changes.
