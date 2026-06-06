# Research: Production Security Hardening

## Decision 1: TLS Termination — Caddy Reverse Proxy

**What was chosen**: Caddy as a reverse proxy in docker-compose.yml for automatic TLS termination.

**Why chosen**: 
- Automatic HTTPS with zero configuration — Caddy provisions and renews Let's Encrypt certs automatically
- Self-signed cert generation for local dev via `caddy untrusted-root` or `mkcert`
- Simpler than Traefik (fewer config lines, less resource usage) for a single-service setup
- Native support for security headers via `header` directive

**Alternatives considered**:
- **Traefik**: More feature-rich but heavier; overkill for single-service TLS termination
- **Nginx with certbot**: Requires manual cert management and cron jobs for renewal
- **Gin native TLS**: `gin.RunTLS()` works but doesn't handle cert renewal; Caddy is purpose-built for this

**Implementation approach**:
- Add `caddy` service to `docker-compose.yml` routing to backend on port 8080
- Local dev: use `tls internal` for self-signed certs
- Production: use `tls { env ACME_EMAIL }` for Let's Encrypt
- Caddy handles HTTP→HTTPS redirect automatically

---

## Decision 2: CSRF Protection — Double-Submit Cookie Pattern

**What was chosen**: Double-submit cookie pattern (CSRF token in cookie + request header) with SameSite=Strict cookies.

**Why chosen**:
- No server-side storage needed — token is cryptographically generated and verified statelessly
- Compatible with existing Redis-backed session architecture
- Frontend reads token from cookie, sends via `X-CSRF-Token` header
- OAuth endpoints exempted (they use `state` parameter for equivalent protection)

**Alternatives considered**:
- **Synchronizer Token Pattern**: Requires server-side token storage per session — adds Redis complexity
- **SameSite cookies only**: Not sufficient alone — older browsers don't support SameSite=Strict
- **Custom header only**: Doesn't protect against CSRF if attacker can set custom headers via Flash/Silverlight

**Implementation approach**:
- New middleware `interfaces/http/middleware/csrf.go`
- Generate 32-byte random token, set as `csrf_token` cookie (HttpOnly=false, SameSite=Strict, Path=/)
- Validate `X-CSRF-Token` header matches cookie value on POST/PUT/DELETE
- Exempt paths: `/oauth2/auth`, `/oauth2/token`, `/oauth2/revoke`, `/oauth2/introspect`, `/health`, `/.well-known/*`
- Config: `CSRF_COOKIE_NAME`, `CSRF_HEADER_NAME`, `CSRF_TOKEN_LENGTH`

---

## Decision 3: Security Headers — Gin Middleware + Nginx

**What was chosen**: Dual-layer approach — Gin middleware for API responses, nginx config for frontend SPA.

**Why chosen**:
- API responses (JSON) need headers from backend directly
- Frontend SPA served by nginx needs headers at the HTTP server level
- Consistent header values across both layers

**Header values**:
```
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'
Strict-Transport-Security: max-age=31536000; includeSubDomains
Permissions-Policy: camera=(), microphone=(), geolocation=(), payment=()
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Embedder-Policy: require-corp
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Referrer-Policy: strict-origin-when-cross-origin
```

**Alternatives considered**:
- **unrolled/secure**: Go library for security headers — adds external dependency for simple header setting
- **Nginx-only**: Would miss headers on direct API calls that bypass nginx

---

## Decision 4: Rate Limiting Extension — Redis Sliding Window

**What was chosen**: Extend existing `middleware/rate_limit.go` with sliding window algorithm and fail-closed behavior.

**Why chosen**:
- Existing rate limiter uses fixed window (Redis INCR + EXPIRE) — adequate but can allow 2x burst at window boundaries
- Sliding window (Redis sorted sets) provides smoother rate limiting
- Fail-closed: on Redis error, reject requests rather than allowing unlimited access (security-first)

**Current state**: `rate_limit.go` exists with fixed-window counter. Already has `X-RateLimit-*` headers.

**Changes needed**:
- Add sliding window implementation using Redis sorted sets (`ZADD`, `ZREMRANGEBYSCORE`, `ZCARD`)
- Change Redis failure behavior from `c.Next()` (fail-open) to reject with 503 (fail-closed)
- Wire to endpoints: `/api/v1/login`, `/api/v1/register`, `/api/v1/verify-otp`, `/api/v1/resend-otp`, `/oauth2/token`
- Config per-endpoint limits via environment variables

**Alternatives considered**:
- **Token bucket**: More complex to implement in Redis; sliding window is simpler and sufficient
- **In-memory rate limiter**: Doesn't work with multi-instance deployments

---

## Decision 5: Metrics — Prometheus client_golang

**What was chosen**: `github.com/prometheus/client_golang` for `/metrics` endpoint.

**Why chosen**:
- Industry standard for Go metrics — compatible with Prometheus, Grafana, Datadog, etc.
- Simple HTTP handler (`promhttp.Handler()`) — minimal code
- Counter type for security events is straightforward

**Metrics to expose**:
```go
securityEventsTotal = prometheus.NewCounterVec(
    Name: "keyles_security_events_total",
    Labels: ["event_type"] // failed_login, rate_limit_triggered, csrf_rejected, tls_error
)
securityEventsDuration = prometheus.NewHistogramVec(
    Name: "keyles_security_event_duration_seconds",
    Labels: ["event_type"]
)
```

**Alternatives considered**:
- **Custom JSON endpoint**: Would require operators to write custom parsers
- **OpenTelemetry**: Heavier dependency; overkill for simple counter metrics
- **StatsD**: Requires external StatsD server; Prometheus pull model is simpler

---

## Decision 6: Secrets Management — Docker Compose .env + Production Override

**What was chosen**: Move all secrets to `.env` file (gitignored), use `docker-compose.prod.yml` for production overrides.

**Why chosen**:
- `.env` is the standard Docker Compose pattern for local secrets
- `docker-compose.prod.yml` allows production-specific values without duplicating the base config
- Backend already validates secrets on startup (config.go lines 225-255) — just need to extend validation

**Current hardcoded secrets in docker-compose.yml**:
- `DB_PASSWORD: postgres` → move to `${DB_PASSWORD}`
- `REDIS_PASSWORD: redis` → move to `${REDIS_PASSWORD}`
- JWT secret default in config.go already validated — no change needed

**Implementation**:
- Create `.env` from `.env.example` with generated secrets
- Update `docker-compose.yml` to use `${VAR}` syntax
- Create `docker-compose.prod.yml` with production defaults (DB_SSL_MODE=require, etc.)
- Add `make generate-secrets` script to create random secrets for `.env`

---

## Decision 7: Database TLS — sslmode=require with CA Verification

**What was chosen**: `DB_SSL_MODE=require` for production, with option for `verify-full` when CA certificate is available.

**Why chosen**:
- `require` encrypts traffic without requiring CA verification — good baseline for production
- `verify-full` requires CA cert and hostname verification — stronger but needs cert infrastructure
- Config already has `DB_SSL_MODE` field — just need to change default and enforce in production validation

**Current state**: `DB_SSL_MODE` defaults to `disable` in config.go line 152.

**Changes needed**:
- Change default to `require` in production mode validation
- Add production check: reject if `DB_SSL_MODE` is `disable` or `allow` when `APP_ENV=production`
- Document how to configure `verify-full` with CA certificate

---

## Decision 8: Request Size Limits — Gin + Nginx Dual Layer

**What was chosen**: Body size limits at both nginx (`client_max_body_size`) and Gin (`MaxMultipartMemory`) levels.

**Why chosen**:
- Nginx layer rejects oversized requests before they reach the application — saves resources
- Gin layer provides defense-in-depth for direct API calls that bypass nginx
- Timeouts prevent slow-loris attacks

**Implementation**:
- Nginx: `client_max_body_size 1m;` (1MB max for JSON payloads)
- Gin: `server.MaxMultipartMemory = 8 << 20` (8MB), `ReadTimeout: 15s`, `WriteTimeout: 15s`, `IdleTimeout: 60s`
- New middleware `request_limits.go` for application-level enforcement

---

## Decision 9: Error Sanitization — Structured Logging with PII Masking

**What was chosen**: Extend existing `error_handler.go` sanitization with structured logging and PII masking.

**Why chosen**:
- Existing `sanitizeError()` function (lines 109-145) uses string matching — effective but incomplete
- Need to add email masking (regex: `.*@.*` → `***@***`)
- Need structured logging with log level filtering — `LOG_LEVEL=info` in production hides debug logs

**Current state**: `error_handler.go` has `sanitizeError()` with pattern matching for sensitive keywords. Recovery handler prevents stack trace exposure.

**Changes needed**:
- Add email regex masking to `sanitizeError()`
- Add structured logger (slog or zap) with level filtering
- Ensure `LOG_LEVEL` env var controls output — debug never in production
- Audit all `log.Printf` calls for sensitive data leakage
