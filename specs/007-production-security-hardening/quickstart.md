# Quickstart: Production Security Hardening

## Local Development Setup

### 1. Generate Development Secrets

```bash
cd backend
make generate-secrets
# Creates .env with random secrets for local development
```

### 2. Generate Self-Signed Certificates

```bash
# Option A: Using mkcert (recommended for local dev)
brew install mkcert
mkcert -install
mkcert -cert-file backend/infrastructure/certs/dev-certs/cert.pem \
       -key-file backend/infrastructure/certs/dev-certs/key.pem \
       localhost 127.0.0.1

# Option B: Using openssl
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout backend/infrastructure/certs/dev-certs/key.pem \
  -out backend/infrastructure/certs/dev-certs/cert.pem \
  -days 365 -subj "/CN=localhost"
```

### 3. Start Services with TLS

```bash
# Start all services (Caddy handles TLS termination)
docker compose up -d

# Access via HTTPS
curl -k https://localhost:443/health
# → {"status":"ok"}
```

### 4. Verify Security Headers

```bash
# Check backend security headers
curl -k -I https://localhost:443/health | grep -E "(Content-Security|Strict-Transport|Permissions-Policy|Cross-Origin)"

# Check frontend security headers
curl -k -I https://localhost:3000 | grep -E "(Content-Security|Strict-Transport|Permissions-Policy|Cross-Origin)"
```

### 5. Verify CSRF Protection

```bash
# Get CSRF token from cookie
curl -k -c /tmp/cookies https://localhost:443/api/v1/dashboard

# Attempt POST without CSRF token (should fail)
curl -k -b /tmp/cookies -X POST https://localhost:443/api/v1/admin/clients \
  -H "Content-Type: application/json" \
  -d '{"name":"test"}'
# → 403 CSRF_TOKEN_INVALID

# Attempt POST with CSRF token (should succeed)
curl -k -b /tmp/cookies -X POST https://localhost:443/api/v1/admin/clients \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: $(grep keyles_csrf /tmp/cookies | awk '{print $7}')" \
  -d '{"name":"test","redirect_uris":["http://localhost:3000/callback"]}'
```

### 6. Verify Rate Limiting

```bash
# Rapid-fire login requests (should trigger rate limit after 5)
for i in {1..7}; do
  curl -k -s -o /dev/null -w "%{http_code}\n" -X POST https://localhost:443/api/v1/login \
    -H "Content-Type: application/json" \
    -d '{"email":"test@test.com","password":"wrong"}'
done
# Expected: 200, 200, 200, 200, 200, 429, 429
```

### 7. Verify Metrics Endpoint

```bash
curl -k https://localhost:443/metrics | grep keyles_security
# → keyles_security_events_total{event_type="..."} N
```

## Production Deployment

### 1. Generate Production Secrets

```bash
# Generate strong secrets
openssl rand -base64 48  # JWT_SECRET (48 bytes)
openssl rand -base64 32  # DB_PASSWORD
openssl rand -base64 32  # REDIS_PASSWORD
```

### 2. Configure Production Environment

```bash
# .env.production
APP_ENV=production
DB_SSL_MODE=require
SECURITY_COOKIE_SECURE=true
JWT_SECRET=<generated-above>
DB_PASSWORD=<generated-above>
REDIS_PASSWORD=<generated-above>
OAUTH_ISSUER=https://sso.yourdomain.com
FRONTEND_URL=https://app.yourdomain.com
ACME_EMAIL=admin@yourdomain.com
LOG_LEVEL=info
```

### 3. Deploy with Production Override

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### 4. Verify Production Hardening

```bash
# All acceptance criteria from spec.md
# SC-001: HTTPS enforced
curl -I http://sso.yourdomain.com/health
# → 301 redirect to https://

# SC-002: No hardcoded secrets
grep -r "password\|secret\|token" docker-compose.yml
# → Only ${VAR} references, no literal values

# SC-004: Security headers present
curl -I https://sso.yourdomain.com/health
# → All security headers present

# SC-009: Database TLS
# Check PostgreSQL logs or run: SELECT ssl FROM pg_stat_ssl WHERE pid = pg_backend_pid();
# → ssl = true
```

## Troubleshooting

### Certificate Errors in Local Dev

```bash
# If using self-signed certs, add -k flag to curl or trust the cert
# For mkcert: mkcert -install adds to system trust store
```

### CSRF Token Mismatch

```bash
# Ensure cookies are being sent with requests
# Check that SameSite=Strict is not blocking cross-origin dev requests
# For local dev with separate frontend/backend, consider SameSite=Lax
```

### Rate Limiting Too Aggressive

```bash
# Adjust limits in .env
RATE_LIMIT_LOGIN_ATTEMPTS_PER_15MIN=10
RATE_LIMIT_OTP_REQUESTS_PER_HOUR=5
```
