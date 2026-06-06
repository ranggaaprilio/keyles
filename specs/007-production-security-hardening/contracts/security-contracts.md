# API Contracts: Production Security Hardening

## Security Headers Contract

All HTTP responses from the backend MUST include the following headers:

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

## CSRF Token Contract

### Token Generation (GET requests that render forms)

**Response Header**:
```
Set-Cookie: keyles_csrf=<32-byte-hex-token>; Path=/; SameSite=Strict; Secure; HttpOnly=false
```

### Token Validation (POST/PUT/DELETE requests)

**Request Header**:
```
X-CSRF-Token: <token-value>
```

**Validation**: Token in `X-CSRF-Token` header MUST match value in `keyles_csrf` cookie.

**Failure Response** (403 Forbidden):
```json
{
  "success": false,
  "error": {
    "code": "CSRF_TOKEN_INVALID",
    "message": "Invalid or missing CSRF token."
  }
}
```

## Rate Limiting Contract

### Rate Limit Headers (all rate-limited endpoints)

```
X-RateLimit-Limit: <max-requests>
X-RateLimit-Remaining: <remaining-requests>
X-RateLimit-Reset: <unix-timestamp>
```

### Rate Limit Exceeded Response (429 Too Many Requests)

```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Please try again later.",
    "retryAfter": <seconds>
  }
}
```

## Metrics Endpoint Contract

### GET /metrics

**Response Content-Type**: `text/plain; version=0.0.4; charset=utf-8`

**Response Body** (Prometheus exposition format):
```
# HELP keyles_security_events_total Total count of security events by type
# TYPE keyles_security_events_total counter
keyles_security_events_total{event_type="failed_login"} 42
keyles_security_events_total{event_type="rate_limit_triggered"} 15
keyles_security_events_total{event_type="csrf_rejected"} 8
keyles_security_events_total{event_type="tls_error"} 0

# HELP keyles_security_event_duration_seconds Duration of security event processing
# TYPE keyles_security_event_duration_seconds histogram
keyles_security_event_duration_seconds_bucket{event_type="failed_login",le="0.001"} 40
keyles_security_event_duration_seconds_bucket{event_type="failed_login",le="0.01"} 42
keyles_security_event_duration_seconds_bucket{event_type="failed_login",le="0.1"} 42
keyles_security_event_duration_seconds_sum{event_type="failed_login"} 0.042
keyles_security_event_duration_seconds_count{event_type="failed_login"} 42
```

**Authentication**: No authentication required — metrics endpoint is read-only and exposes only aggregate counters (no sensitive data).

**Access Control**: In production, this endpoint SHOULD be restricted to internal monitoring networks only (via firewall or reverse proxy rules).

## Error Response Contract

All error responses MUST follow this sanitized format:

```json
{
  "success": false,
  "error": {
    "code": "<ERROR_CODE>",
    "message": "<user-friendly-message>"
  }
}
```

**Rules**:
- `message` MUST NOT contain stack traces, SQL queries, file paths, or internal details
- `message` MUST NOT contain passwords, tokens, secrets, or unmasked PII
- `code` MUST be a stable, documented error code (e.g., `INTERNAL_SERVER_ERROR`, `BAD_REQUEST`)
- OAuth errors follow RFC 6749 format: `{ "error": "...", "error_description": "..." }`
