# Epic 03: Operational Readiness

## Goal

Make Keyles operable in production: automated database migrations, structured logging, health check alerting, metrics export, backup/restore procedures, and a complete deployment guide so an operator can run and maintain the service without source code access.

## Why MVP

A service that cannot be deployed, monitored, backed up, or debugged in production is not an MVP — it's a prototype. Operators need to know the service is healthy, be alerted when it's not, have logs they can search, and be able to recover from failures. Without these, every incident becomes a crisis.

## Current State

- **Migrations**: SQL migration files exist (000001-000013) but are not applied automatically on startup. Operators must run them manually with a separate tool.
- **Logging**: Uses `log.Printf` throughout. No structured/JSON logging, no log levels at runtime, no correlation IDs, no request tracing.
- **Health checks**: Three endpoints exist (`/health`, `/health/db`, `/health/redis`) returning JSON. No alerting integration. No readiness/liveness probes beyond basic health.
- **Metrics**: No Prometheus metrics endpoint. No request latency, error rate, or business metrics (registrations, logins, token issues).
- **Backups**: No database backup scripts, no backup schedule, no restore procedure documented.
- **Deployment**: Docker Compose works for dev. No production deployment guide, no Kubernetes manifests, no Terraform, no cloud-specific instructions.
- **Admin bootstrapping**: `cmd/seed/main.go` exists for dev seeding. No documented path for creating the first admin user in production.
- **Cleanup**: `cmd/cleanup/main.go` has `expire-stale-invitations` and `purge-user-events` but no cron/scheduling mechanism.

## Tasks

### 3.1 — Automated database migrations on startup
- Integrate a migration runner (golang-migrate or custom) that applies pending SQL migrations on server start
- Track applied migrations in a `schema_migrations` table
- Add `--migrate-only` flag to run migrations and exit (for init containers)
- Add `--skip-migrations` flag to skip auto-migration (for rollback scenarios)
- Ensure migrations are idempotent and safe to run multiple times

### 3.2 — Structured logging
- Replace `log.Printf` with a structured logger (zerolog or zap)
- Emit JSON-formatted logs in production, text-formatted in development
- Add correlation ID (request ID) middleware — generate a unique ID per request, include in all log lines
- Add log levels (debug, info, warn, error) controllable via `LOG_LEVEL` env var
- Log key events: registration, login, OAuth authorization, token issue, token refresh, user invite, role change, status change

### 3.3 — Prometheus metrics endpoint
- Add `GET /metrics` endpoint exposing Prometheus-format metrics
- Track: HTTP request duration (histogram), HTTP request count (counter), active sessions, error rate by endpoint
- Track business metrics: registrations, logins, OAuth authorizations, token issues, token refreshes, user invitations
- Use `promhttp` handler for standard Go runtime metrics (goroutines, memory, GC)

### 3.4 — Health check improvements
- Add `/health/live` (liveness — is the process running?) and `/health/ready` (readiness — can it serve traffic?)
- Readiness check verifies DB and Redis connectivity
- Add health check response time tracking
- Document expected health check behavior for load balancer/orchestrator configuration

### 3.5 — Database backup and restore
- Create `scripts/backup-db.sh` — dumps PostgreSQL to timestamped file, optionally uploads to S3/GCS
- Create `scripts/restore-db.sh` — restores from a backup file
- Add backup schedule documentation (cron example)
- Test backup and restore end-to-end

### 3.6 — Admin user bootstrapping
- Create `cmd/bootstrap/main.go` — CLI that creates the first admin user for a tenant
- Alternatively: add a `/api/v1/setup` endpoint that works only when no admin users exist (first-run setup)
- Document the bootstrap process in deployment guide
- Ensure bootstrap cannot be exploited after initial setup

### 3.7 — Cron job scheduling for cleanup
- Add cron-like scheduling to the cleanup binary or integrate with the main server
- Schedule: expire stale invitations (every hour), purge old user events (daily)
- Document cron setup for production (systemd timer, Kubernetes CronJob, or external cron)

### 3.8 — Production deployment guide
- Write `docs/DEPLOYMENT.md` covering:
  - Infrastructure requirements (PostgreSQL 15+, Redis 7+, compute, TLS certs)
  - Environment variables and secrets setup
  - Docker Compose production deployment
  - Migration execution
  - Admin bootstrapping
  - Health check endpoints
  - Backup/restore procedures
  - Upgrade/rollback procedures
  - Monitoring setup (Prometheus scrape config, Grafana dashboard JSON)
  - Common troubleshooting scenarios

### 3.9 — Graceful degradation
- Ensure the service remains partially functional when Redis is down (degraded: no rate limiting, no OTP caching)
- Ensure the service remains partially functional when email service is down (registration works, OTP can be resent later)
- Return appropriate 503 with Retry-After when dependent services are unavailable
- Add circuit breaker for external service calls (Brevo API)

## Acceptance Criteria

1. Database migrations run automatically on server startup — no manual step required
2. All log output is structured JSON in production mode, with correlation IDs on every line
3. `GET /metrics` returns Prometheus-format metrics including HTTP and business metrics
4. `/health/live` returns 200 if process is running; `/health/ready` returns 200 only if DB and Redis are reachable
5. Backup script produces a restorable database dump; restore script successfully restores it
6. A documented, tested bootstrap path exists for creating the first admin in production
7. Cleanup jobs can be scheduled and run on a recurring basis
8. `docs/DEPLOYMENT.md` exists and covers all listed topics
9. Service degrades gracefully when Redis or email is unavailable (returns appropriate errors, does not crash)
