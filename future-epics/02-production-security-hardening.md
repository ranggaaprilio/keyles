# Epic 02: Production Security Hardening

## Goal

Harden the entire stack for production deployment: enforce HTTPS/TLS, remove hardcoded secrets, add security headers, implement CSRF protection, apply input sanitization, and ensure no sensitive data is exposed in logs or error messages.

## Why MVP

An SSO platform handles authentication credentials, authorization codes, and identity tokens. Deploying without HTTPS, with hardcoded secrets, or without proper security headers is unacceptable for any production service. Security hardening is not a nice-to-have — it is a prerequisite for going live.

## Current State

- **Docker Compose**: Runs plain HTTP on ports 8080 (backend) and 3000 (frontend). No TLS termination.
- **Secrets**: `docker-compose.yml` contains hardcoded dev passwords for PostgreSQL and default JWT secret. `.env.example` exists but no guidance on production secret rotation.
- **Nginx**: Frontend nginx config has basic security headers (X-Frame-Options, X-Content-Type-Options, XSS-Protection, Referrer-Policy) but is missing CSP, HSTS, and Permissions-Policy.
- **Backend**: No application-level security headers. CORS is configurable but defaults are broad. Error messages may leak internal details. No request body size limits.
- **CSRF**: Only the OAuth `state` parameter provides CSRF protection. API endpoints (POST/PUT/DELETE) have no CSRF tokens.
- **Input validation**: Domain entities validate business rules but there is no centralized input sanitization against XSS or SQL injection at the handler layer (though parameterized queries via GORM mitigate SQL injection).

## Tasks

### 2.1 — Add TLS termination
- Add Traefik or Caddy as a reverse proxy in `docker-compose.yml` for TLS termination
- Alternatively, configure nginx in the frontend container for TLS with Let's Encrypt certs
- Generate self-signed certs for local development
- Document production TLS setup (certificate provisioning, auto-renewal)

### 2.2 — Remove hardcoded secrets from docker-compose.yml
- Move all secrets to a `.env` file (gitignored) referenced by docker-compose
- Add `docker-compose.prod.yml` override for production settings
- Add secrets validation on backend startup (reject default/weak secrets in production mode)
- Document secret generation and rotation procedures

### 2.3 — Add Content Security Policy (CSP)
- Add CSP header in nginx config for the frontend
- Add CSP header in backend Gin middleware for API responses
- Policy should restrict script-src, style-src, connect-src, frame-ancestors appropriately
- Test with report-only mode first, then enforce

### 2.4 — Add HSTS and remaining security headers
- Add `Strict-Transport-Security` header (max-age=31536000; includeSubDomains)
- Add `Permissions-Policy` header
- Add `Cross-Origin-Opener-Policy` and `Cross-Origin-Embedder-Policy` headers
- Verify all security headers are present on every response

### 2.5 — Add CSRF protection for API endpoints
- Implement CSRF token generation and validation middleware for the backend
- Frontend: include CSRF token in all state-changing requests (POST/PUT/DELETE)
- Exempt OAuth endpoints that use `state` parameter
- Use SameSite=Strict or SameSite=Lax cookies where applicable

### 2.6 — Add request size limits and timeout
- Configure `gin` with `MaxMultipartMemory` and read timeout
- Add request body size limits in nginx (`client_max_body_size`)
- Add connection timeout, read timeout, and send timeout
- Prevent slow-loris attacks with appropriate timeouts

### 2.7 — Sanitize error messages and logging
- Audit all error responses: ensure no stack traces, SQL queries, or internal paths are exposed
- Implement centralized error response formatting
- Ensure log messages don't contain passwords, tokens, or PII (email addresses in logs should be masked/hashed)
- Add `LOG_LEVEL` filtering (debug logs should never appear in production)

### 2.8 — Add rate limiting for all public endpoints
- Extend rate limiting beyond OAuth token endpoint to cover login, registration, OTP verify, OTP resend
- Use Redis-backed sliding window rate limiter (already initialized, just need to wire more endpoints)
- Add rate limit response headers (`X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`)

### 2.9 — Database connection hardening
- Enforce `DB_SSL_MODE=require` (or `verify-full`) for production
- Add connection pool limits with sensible defaults for production
- Add statement timeout to prevent long-running queries
- Encrypt sensitive columns at rest (client secrets, if not already hashed)

## Acceptance Criteria

1. All traffic between client and server is encrypted (HTTPS) in production config
2. No hardcoded secrets exist in any committed file
3. Backend rejects startup in production mode with default/weak secrets
4. CSP header is present and restricts script sources appropriately
5. HSTS header is present with appropriate max-age
6. All state-changing API requests require a valid CSRF token
7. Error responses never contain stack traces, SQL, or internal paths
8. Sensitive data (passwords, tokens, PII) never appears in logs at INFO level or above
9. Rate limiting is active on login, registration, OTP, and token endpoints
10. Database connects over TLS in production configuration
