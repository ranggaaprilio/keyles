# Data Model: Core SSO Auth Provider

**Date**: December 26, 2025  
**Feature**: [spec.md](./spec.md)  
**Research**: [research.md](./research.md)

## Overview

This document defines the complete database schema for the Core SSO Auth Provider. The system uses:

- **PostgreSQL**: Persistent data (tenants, users, clients, roles, signing keys, audit logs)
- **Redis**: Ephemeral data (authorization codes, sessions, refresh tokens, rate limiting)

**Multi-Tenancy Strategy**: Shared schema with `tenant_id` column (Row-Level Security)

---

## PostgreSQL Schema

### Migration Order

Migrations must be applied in this order due to foreign key dependencies:

1. `000001_create_tenants.sql`
2. `000002_create_clients.sql`
3. `000003_create_users.sql`
4. `000004_create_user_role_assignments.sql`
5. `000005_create_refresh_tokens.sql`
6. `000006_create_signing_keys.sql`
7. `000007_create_audit_logs.sql`

---

### 1. Tenants Table

Represents platform tenants/organizations.

```sql
-- migrations/000001_create_tenants.up.sql
CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255) UNIQUE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT tenants_domain_lowercase CHECK (domain = LOWER(domain))
);

CREATE INDEX idx_tenants_domain ON tenants(domain);
CREATE INDEX idx_tenants_is_active ON tenants(is_active);

-- Trigger to update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- migrations/000001_create_tenants.down.sql
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;
DROP TABLE IF EXISTS tenants CASCADE;
DROP FUNCTION IF EXISTS update_updated_at_column();
```

**Fields**:

- `id`: Unique tenant identifier (e.g., `tenant_a`, UUID)
- `name`: Human-readable tenant name
- `domain`: Unique domain identifier for tenant
- `is_active`: Enable/disable tenant access
- `created_at`, `updated_at`: Audit timestamps

---

### 2. Clients Table

Represents registered OAuth2/OIDC client applications.

```sql
-- migrations/000002_create_clients.up.sql
CREATE TABLE IF NOT EXISTS clients (
    client_id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    client_secret VARCHAR(255) NOT NULL,
    client_name VARCHAR(255) NOT NULL,
    redirect_uris TEXT[] NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT clients_redirect_uris_not_empty CHECK (array_length(redirect_uris, 1) > 0)
);

CREATE INDEX idx_clients_tenant ON clients(tenant_id);
CREATE INDEX idx_clients_is_active ON clients(is_active);

CREATE TRIGGER update_clients_updated_at BEFORE UPDATE ON clients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- migrations/000002_create_clients.down.sql
DROP TRIGGER IF EXISTS update_clients_updated_at ON clients;
DROP TABLE IF EXISTS clients CASCADE;
```

**Fields**:

- `client_id`: Unique client identifier (generated)
- `tenant_id`: FK to tenants table (determines tenant context)
- `client_secret`: Hashed client secret for authentication
- `client_name`: Human-readable client name
- `redirect_uris`: Array of allowed redirect URIs (exact match validation)
- `is_active`: Enable/disable client
- `created_at`, `updated_at`: Audit timestamps

**Security Notes**:

- `client_secret` should be hashed using bcrypt before storage
- `redirect_uris` is an array to support multiple callbacks per client

---

### 3. Users Table

User accounts per tenant.

```sql
-- migrations/000003_create_users.up.sql
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    given_name VARCHAR(255),
    family_name VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT users_email_lowercase CHECK (email = LOWER(email)),
    CONSTRAINT users_unique_email_per_tenant UNIQUE (tenant_id, email)
);

CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_tenant_email ON users(tenant_id, email);
CREATE INDEX idx_users_is_active ON users(is_active);

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- migrations/000003_create_users.down.sql
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TABLE IF EXISTS users CASCADE;
```

**Fields**:

- `id`: Unique user identifier (UUID recommended)
- `tenant_id`: FK to tenants table
- `email`: User email address (unique per tenant)
- `email_verified`: Email verification status
- `password_hash`: Bcrypt hashed password
- `name`, `given_name`, `family_name`: User profile information (OIDC claims)
- `is_active`: Enable/disable user account
- `last_login_at`: Track last authentication time
- `created_at`, `updated_at`: Audit timestamps

**Security Notes**:

- Email is unique per tenant (not globally unique)
- Password hash using bcrypt with cost factor 12+
- Composite index on `(tenant_id, email)` for fast lookups

---

### 4. User Role Assignments Table

Role-based access control: which users can access which clients.

```sql
-- migrations/000004_create_user_role_assignments.up.sql
CREATE TABLE IF NOT EXISTS user_role_assignments (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id VARCHAR(255) NOT NULL REFERENCES clients(client_id) ON DELETE CASCADE,
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    granted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    granted_by VARCHAR(255),

    CONSTRAINT user_role_assignments_unique UNIQUE (user_id, client_id, role)
);

CREATE INDEX idx_user_role_assignments_user ON user_role_assignments(user_id);
CREATE INDEX idx_user_role_assignments_client ON user_role_assignments(client_id);
CREATE INDEX idx_user_role_assignments_tenant ON user_role_assignments(tenant_id);
CREATE INDEX idx_user_role_assignments_user_client ON user_role_assignments(user_id, client_id);
CREATE INDEX idx_user_role_assignments_is_active ON user_role_assignments(is_active);

-- migrations/000004_create_user_role_assignments.down.sql
DROP TABLE IF EXISTS user_role_assignments CASCADE;
```

**Fields**:

- `id`: Auto-increment primary key
- `user_id`: FK to users table
- `client_id`: FK to clients table
- `tenant_id`: FK to tenants table (denormalized for filtering)
- `role`: Role name (e.g., "admin", "user", "viewer")
- `is_active`: Enable/disable role assignment
- `granted_at`: When role was assigned
- `granted_by`: User ID of admin who assigned the role

**Usage**:

- Check if user has ANY active role for a client during authentication
- Roles can be used for fine-grained authorization in resource servers
- Composite unique constraint prevents duplicate role assignments

**Query Example**:

```sql
-- Check if user can access client
SELECT EXISTS (
    SELECT 1 FROM user_role_assignments
    WHERE user_id = $1
    AND client_id = $2
    AND is_active = true
) AS has_access;
```

---

### 5. Refresh Tokens Table

Long-lived tokens for token refresh flow (stored in PostgreSQL for revocation).

```sql
-- migrations/000005_create_refresh_tokens.up.sql
CREATE TABLE IF NOT EXISTS refresh_tokens (
    token VARCHAR(255) PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL REFERENCES clients(client_id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    scope TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_used_at TIMESTAMP WITH TIME ZONE,
    is_revoked BOOLEAN NOT NULL DEFAULT false,
    revoked_at TIMESTAMP WITH TIME ZONE,
    revoked_by VARCHAR(255)
);

CREATE INDEX idx_refresh_tokens_client ON refresh_tokens(client_id);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_tenant ON refresh_tokens(tenant_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_is_revoked ON refresh_tokens(is_revoked);
CREATE INDEX idx_refresh_tokens_user_client ON refresh_tokens(user_id, client_id);

-- migrations/000005_create_refresh_tokens.down.sql
DROP TABLE IF EXISTS refresh_tokens CASCADE;
```

**Fields**:

- `token`: Unique token value (cryptographically random, opaque)
- `client_id`: FK to clients table
- `user_id`: FK to users table
- `tenant_id`: FK to tenants table
- `scope`: OAuth scopes granted to this token
- `created_at`: Token creation timestamp
- `expires_at`: Token expiration (7 days from creation)
- `last_used_at`: Last time token was used for refresh
- `is_revoked`: Revocation flag
- `revoked_at`, `revoked_by`: Revocation audit trail

**Cleanup Strategy**:

```sql
-- Periodic cleanup of expired/revoked tokens (run via cron/scheduled job)
DELETE FROM refresh_tokens
WHERE expires_at < NOW() - INTERVAL '30 days'
   OR (is_revoked = true AND revoked_at < NOW() - INTERVAL '90 days');
```

**Revocation Queries**:

```sql
-- Revoke all refresh tokens for a user-client combination
UPDATE refresh_tokens
SET is_revoked = true, revoked_at = NOW(), revoked_by = $3
WHERE user_id = $1 AND client_id = $2 AND is_revoked = false;

-- Revoke all tokens for a user (account disabled)
UPDATE refresh_tokens
SET is_revoked = true, revoked_at = NOW(), revoked_by = 'system'
WHERE user_id = $1 AND is_revoked = false;
```

---

### 6. Signing Keys Table

RSA keypairs for JWT signing and verification.

```sql
-- migrations/000006_create_signing_keys.up.sql
CREATE TABLE IF NOT EXISTS signing_keys (
    kid VARCHAR(255) PRIMARY KEY,
    algorithm VARCHAR(50) NOT NULL DEFAULT 'RS256',
    private_key TEXT NOT NULL,
    public_key TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT signing_keys_algorithm_check CHECK (algorithm IN ('RS256', 'RS384', 'RS512'))
);

CREATE INDEX idx_signing_keys_is_active ON signing_keys(is_active);
CREATE INDEX idx_signing_keys_expires_at ON signing_keys(expires_at);

-- migrations/000006_create_signing_keys.down.sql
DROP TABLE IF EXISTS signing_keys CASCADE;
```

**Fields**:

- `kid`: Key ID (included in JWT header)
- `algorithm`: Signing algorithm (RS256 required, others optional)
- `private_key`: PEM-encoded RSA private key (encrypted at rest)
- `public_key`: PEM-encoded RSA public key
- `is_active`: Whether key can be used for signing new tokens
- `created_at`: Key creation timestamp
- `expires_at`: Optional key expiration for rotation

**Key Rotation Strategy**:

1. Generate new key with unique `kid`
2. Set new key as active (`is_active = true`)
3. Set old key as inactive (`is_active = false`)
4. Keep inactive keys for JWT verification (tokens signed with old keys still valid)
5. Remove keys after all tokens signed with them have expired

**Query Examples**:

```sql
-- Get active key for signing
SELECT kid, private_key FROM signing_keys
WHERE is_active = true AND algorithm = 'RS256'
ORDER BY created_at DESC LIMIT 1;

-- Get all public keys for JWKS endpoint
SELECT kid, algorithm, public_key FROM signing_keys
WHERE is_active = true OR expires_at > NOW();
```

**Security Notes**:

- Private keys should be encrypted at rest using application-level encryption
- Consider using HashiCorp Vault or AWS KMS for key management in production
- Rotate keys every 90-180 days

---

### 7. Audit Logs Table

Security audit trail for all authentication events.

```sql
-- migrations/000007_create_audit_logs.up.sql
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    user_id VARCHAR(255),
    tenant_id VARCHAR(255),
    client_id VARCHAR(255),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN NOT NULL,
    error_message TEXT,
    metadata JSONB,

    CONSTRAINT audit_logs_event_type_check CHECK (
        event_type IN (
            'authentication_attempt',
            'authentication_success',
            'authentication_failure',
            'token_issued',
            'token_refreshed',
            'token_revoked',
            'client_created',
            'client_updated',
            'client_deleted',
            'user_created',
            'user_updated',
            'user_disabled',
            'role_assigned',
            'role_revoked'
        )
    )
);

CREATE INDEX idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_client_id ON audit_logs(client_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX idx_audit_logs_success ON audit_logs(success);
CREATE INDEX idx_audit_logs_metadata ON audit_logs USING GIN (metadata);

-- Partitioning for performance (optional, for high-volume deployments)
-- CREATE TABLE audit_logs_2025_12 PARTITION OF audit_logs
--     FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');

-- migrations/000007_create_audit_logs.down.sql
DROP TABLE IF EXISTS audit_logs CASCADE;
```

**Fields**:

- `id`: Auto-increment primary key
- `event_type`: Type of event (constrained to specific values)
- `user_id`: User involved (nullable for system events)
- `tenant_id`: Tenant context (nullable for system events)
- `client_id`: Client involved (nullable for non-client events)
- `timestamp`: Event occurrence time
- `ip_address`: Source IP address
- `user_agent`: Client user agent string
- `success`: Whether operation succeeded
- `error_message`: Error details for failed operations
- `metadata`: Additional structured data (JSONB for flexible storage)

**Usage Examples**:

```sql
-- Log authentication attempt
INSERT INTO audit_logs (event_type, user_id, tenant_id, client_id, ip_address, user_agent, success)
VALUES ('authentication_attempt', $1, $2, $3, $4, $5, $6);

-- Query failed login attempts for user
SELECT * FROM audit_logs
WHERE user_id = $1
  AND event_type = 'authentication_failure'
  AND timestamp > NOW() - INTERVAL '1 hour'
ORDER BY timestamp DESC;

-- Security analysis: find brute force attempts
SELECT ip_address, COUNT(*) as attempt_count
FROM audit_logs
WHERE event_type = 'authentication_failure'
  AND timestamp > NOW() - INTERVAL '10 minutes'
GROUP BY ip_address
HAVING COUNT(*) > 10
ORDER BY attempt_count DESC;
```

**Retention Policy**:

```sql
-- Archive old audit logs (run monthly)
DELETE FROM audit_logs WHERE timestamp < NOW() - INTERVAL '2 years';
```

---

## Redis Data Structures

### 1. Authorization Codes

**TTL**: 5 minutes (300 seconds)

**Key Pattern**: `auth:code:{code_value}`

**Value Structure** (JSON):

```json
{
  "code": "auth_code_a1b2c3d4e5f6g7h8",
  "client_id": "abc123xyz",
  "user_id": "user_12345",
  "tenant_id": "tenant_a",
  "redirect_uri": "https://client-app.example.com/callback",
  "scope": "openid profile email",
  "code_challenge": "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
  "code_challenge_method": "S256",
  "created_at": "2025-12-26T10:30:15Z"
}
```

**Operations**:

```go
// Store authorization code
redisClient.Set(ctx,
    fmt.Sprintf("auth:code:%s", code),
    jsonData,
    5*time.Minute)

// Retrieve and delete (atomic)
data, _ := redisClient.GetDel(ctx, fmt.Sprintf("auth:code:%s", code)).Result()

// Check if already used
exists, _ := redisClient.Exists(ctx, fmt.Sprintf("auth:code:%s", code)).Result()
```

---

### 2. User Sessions

**TTL**: 8 hours (28800 seconds), sliding expiration

**Key Pattern**: `session:{session_id}`

**Value Structure** (JSON):

```json
{
  "session_id": "sess_9f8e7d6c5b4a3210",
  "user_id": "user_12345",
  "tenant_id": "tenant_a",
  "email": "user@tenant-a.com",
  "created_at": "2025-12-26T10:30:00Z",
  "expires_at": "2025-12-26T18:30:00Z",
  "ip_address": "203.0.113.42",
  "user_agent": "Mozilla/5.0...",
  "last_activity": "2025-12-26T10:30:00Z"
}
```

**Operations**:

```go
// Create session
redisClient.Set(ctx,
    fmt.Sprintf("session:%s", sessionID),
    jsonData,
    8*time.Hour)

// Get session and extend TTL (sliding expiration)
data, _ := redisClient.Get(ctx, fmt.Sprintf("session:%s", sessionID)).Result()
redisClient.Expire(ctx, fmt.Sprintf("session:%s", sessionID), 8*time.Hour)

// Delete session (logout)
redisClient.Del(ctx, fmt.Sprintf("session:%s", sessionID))
```

---

### 3. Refresh Token Cache

**TTL**: 7 days (604800 seconds)

**Key Pattern**: `refresh:{token_value}`

**Value Structure** (JSON - cached from PostgreSQL):

```json
{
  "token": "refresh_9z8y7x6w5v4u3t2s1r0q",
  "client_id": "abc123xyz",
  "user_id": "user_12345",
  "tenant_id": "tenant_a",
  "scope": "openid profile email",
  "expires_at": "2026-01-02T10:30:20Z",
  "is_revoked": false
}
```

**Operations**:

```go
// Cache refresh token after creation (write-through cache)
redisClient.Set(ctx,
    fmt.Sprintf("refresh:%s", token),
    jsonData,
    7*24*time.Hour)

// Check cache first, fall back to PostgreSQL
data, err := redisClient.Get(ctx, fmt.Sprintf("refresh:%s", token)).Result()
if err == redis.Nil {
    // Cache miss - query PostgreSQL and populate cache
    data = queryPostgreSQL(token)
    redisClient.Set(ctx, fmt.Sprintf("refresh:%s", token), data, 7*24*time.Hour)
}

// Invalidate cache on revocation
redisClient.Del(ctx, fmt.Sprintf("refresh:%s", token))
```

---

### 4. JWKS Cache

**TTL**: 1 hour (3600 seconds)

**Key Pattern**: `jwks:public_keys`

**Value Structure** (JSON - JWKS document):

```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "key_2025_01",
      "alg": "RS256",
      "n": "modulus_base64url_encoded_value",
      "e": "AQAB"
    }
  ]
}
```

**Operations**:

```go
// Cache JWKS document
redisClient.Set(ctx, "jwks:public_keys", jwksJSON, time.Hour)

// Retrieve cached JWKS
jwks, err := redisClient.Get(ctx, "jwks:public_keys").Result()
if err == redis.Nil {
    // Cache miss - generate from PostgreSQL and cache
    jwks = generateJWKS()
    redisClient.Set(ctx, "jwks:public_keys", jwks, time.Hour)
}

// Invalidate on key rotation
redisClient.Del(ctx, "jwks:public_keys")
```

---

### 5. Rate Limiting

**TTL**: 1 minute (60 seconds)

**Key Pattern**: `ratelimit:token:{client_id}`

**Value**: Integer counter

**Operations**:

```go
// Increment counter
key := fmt.Sprintf("ratelimit:token:%s", clientID)
count, _ := redisClient.Incr(ctx, key).Result()

// Set expiry on first request
if count == 1 {
    redisClient.Expire(ctx, key, time.Minute)
}

// Check limit
if count > 10 {
    return ErrRateLimitExceeded
}

// Alternative: Using INCR with PEXPIRE in pipeline
pipe := redisClient.Pipeline()
incrCmd := pipe.Incr(ctx, key)
pipe.PExpire(ctx, key, time.Minute)
pipe.Exec(ctx)
count := incrCmd.Val()
```

---

## Entity Relationships

```
┌──────────┐
│ tenants  │
└────┬─────┘
     │
     ├──────────┐
     │          │
     ▼          ▼
┌─────────┐  ┌────────┐
│ clients │  │ users  │
└────┬────┘  └───┬────┘
     │           │
     │           │
     │     ┌─────┴────────────┐
     │     │                  │
     ▼     ▼                  ▼
┌────────────────────┐  ┌─────────────────┐
│ user_role_         │  │ refresh_tokens  │
│   assignments      │  └─────────────────┘
└────────────────────┘

┌────────────────┐     ┌──────────────┐
│ signing_keys   │     │ audit_logs   │
└────────────────┘     └──────────────┘
```

**Relationships**:

- `tenants` (1) ──< (N) `clients`
- `tenants` (1) ──< (N) `users`
- `clients` (1) ──< (N) `user_role_assignments` ─> (N) `users`
- `users` (1) ──< (N) `refresh_tokens` ─> (N) `clients`
- `signing_keys`: Standalone (no foreign keys)
- `audit_logs`: References all entities but no FKs (for flexibility)

---

## Data Access Patterns

### 1. Authentication Flow

```sql
-- Step 1: Validate client (from authorization request)
SELECT c.*, t.is_active AS tenant_active
FROM clients c
JOIN tenants t ON c.tenant_id = t.id
WHERE c.client_id = $1 AND c.is_active = true AND t.is_active = true;

-- Step 2: Validate user credentials
SELECT id, tenant_id, password_hash, is_active
FROM users
WHERE tenant_id = $1 AND email = $2 AND is_active = true;

-- Step 3: Check user has permission for client
SELECT EXISTS (
    SELECT 1 FROM user_role_assignments
    WHERE user_id = $1 AND client_id = $2 AND is_active = true
) AS has_permission;

-- Step 4: Create audit log
INSERT INTO audit_logs (...) VALUES (...);
```

### 2. Token Exchange

```sql
-- Validate client credentials
SELECT client_id, tenant_id
FROM clients
WHERE client_id = $1 AND client_secret = $2 AND is_active = true;

-- Create refresh token record
INSERT INTO refresh_tokens (token, client_id, user_id, tenant_id, scope, expires_at)
VALUES ($1, $2, $3, $4, $5, NOW() + INTERVAL '7 days')
RETURNING *;
```

### 3. Token Refresh

```sql
-- Validate refresh token (check PostgreSQL if not in Redis)
SELECT *
FROM refresh_tokens
WHERE token = $1
  AND is_revoked = false
  AND expires_at > NOW();

-- Update last_used_at
UPDATE refresh_tokens
SET last_used_at = NOW()
WHERE token = $1;
```

### 4. Admin Operations

```sql
-- List clients for tenant
SELECT * FROM clients
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- Assign role to user
INSERT INTO user_role_assignments (user_id, client_id, tenant_id, role, granted_by)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, client_id, role) DO UPDATE
SET is_active = true, granted_at = NOW();

-- Revoke user access
UPDATE user_role_assignments
SET is_active = false
WHERE user_id = $1 AND client_id = $2;
```

---

## Performance Considerations

### Indexes Summary

| Table                   | Index                                     | Purpose                    |
| ----------------------- | ----------------------------------------- | -------------------------- |
| `tenants`               | `domain`, `is_active`                     | Fast tenant lookup         |
| `clients`               | `tenant_id`, `is_active`                  | Tenant filtering           |
| `users`                 | `(tenant_id, email)`, `is_active`         | Authentication lookup      |
| `user_role_assignments` | `(user_id, client_id)`, `is_active`       | Permission checks          |
| `refresh_tokens`        | `(user_id, client_id)`, `expires_at`      | Token validation & cleanup |
| `signing_keys`          | `is_active`, `expires_at`                 | Key selection              |
| `audit_logs`            | `timestamp DESC`, `event_type`, `user_id` | Audit queries              |

### Query Optimization

1. **Use prepared statements** for all queries (prevents SQL injection + performance)
2. **Connection pooling**: Configure pgx pool (25 max connections)
3. **Read replicas**: Route read queries to replicas if high load
4. **Partitioning**: Consider time-based partitioning for `audit_logs`
5. **EXPLAIN ANALYZE**: Profile slow queries during development

### Caching Strategy

1. **Write-through cache** for refresh tokens (Redis + PostgreSQL)
2. **Cache-aside** for JWKS (check Redis first, generate if miss)
3. **Session-only cache** for authorization codes and sessions (Redis only)
4. **Cache invalidation**: Delete Redis keys on revocation/updates

---

## Data Validation Rules

### Application-Level Validation

```go
// Client redirect URIs
- Must be HTTPS in production
- localhost allowed for development
- Exact match required (no wildcards)
- Max 10 URIs per client

// Passwords
- Min length: 12 characters
- Bcrypt hash with cost 12

// Emails
- RFC 5322 format validation
- Lowercase normalization
- Unique per tenant

// Client IDs/Secrets
- Generated using crypto/rand
- Min 32 characters
- Base64 URL-safe encoding

// Tokens
- Authorization codes: 32 bytes random
- Refresh tokens: 32 bytes random
- Session IDs: 16 bytes random (UUID also acceptable)
```

---

## Backup and Recovery

### Backup Strategy

**PostgreSQL**:

```bash
# Daily full backup
pg_dump -h localhost -U postgres keyles_sso > backup_$(date +%Y%m%d).sql

# Point-in-time recovery setup
# Enable WAL archiving in postgresql.conf
```

**Redis**:

```bash
# Enable AOF (Append-Only File) for durability
redis-cli CONFIG SET appendonly yes

# Manual snapshot
redis-cli BGSAVE
```

### Disaster Recovery

1. **RPO (Recovery Point Objective)**: 1 hour
2. **RTO (Recovery Time Objective)**: 4 hours
3. **Backup retention**: 30 days
4. **Test restores**: Monthly

---

## Migration Checklist

- [x] Migration files created in correct order
- [x] Up and down migrations defined
- [x] Foreign key constraints validated
- [x] Indexes optimized for query patterns
- [x] Triggers for updated_at timestamps
- [x] Check constraints for data validation
- [x] Unique constraints for business rules
- [x] Default values specified
- [x] NOT NULL constraints applied appropriately

**Next Steps**:

1. Create API contracts (OpenAPI specifications)
2. Define repository interfaces in Domain layer
3. Implement concrete repository classes in Infrastructure layer
