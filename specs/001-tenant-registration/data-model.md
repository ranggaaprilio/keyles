# Data Model: Multi-Tenant Registration with Email Verification

**Feature**: 001-tenant-registration  
**Date**: 2025-12-06  
**Database**: PostgreSQL 15+

## Entity Relationship Diagram

```
┌─────────────────────┐
│      Tenant         │
├─────────────────────┤
│ id (UUID, PK)       │
│ organization_name   │◄──┐
│ status              │   │
│ created_at          │   │
│ verified_at         │   │
│ updated_at          │   │
└─────────────────────┘   │
                          │ 1:N
                          │
┌─────────────────────┐   │
│     AdminUser       │   │
├─────────────────────┤   │
│ id (UUID, PK)       │   │
│ tenant_id (FK) ─────┼───┘
│ full_name           │
│ email               │
│ password_hash       │
│ role                │
│ created_at          │
│ updated_at          │
└─────────────────────┘
        │
        │ 1:1 (initially)
        │
┌─────────────────────┐
│  OTPVerification    │
├─────────────────────┤
│ id (UUID, PK)       │
│ tenant_id (FK) ─────┼───┐
│ otp_code            │   │
│ created_at          │   │
│ expires_at          │   │
│ verified_at         │   │
│ attempt_count       │   │
│ ip_address          │   │
└─────────────────────┘   │
                          │
┌─────────────────────┐   │
│     AuditLog        │   │
├─────────────────────┤   │
│ id (UUID, PK)       │   │
│ tenant_id (FK) ─────┼───┘
│ user_id (UUID)      │
│ event_type          │
│ event_data (JSONB)  │
│ ip_address          │
│ user_agent          │
│ created_at          │
└─────────────────────┘
```

## Entities

### 1. Tenant

**Purpose**: Represents an organization in the multi-tenant SSO platform

**Fields**:
- `id` (UUID, PRIMARY KEY): Unique tenant identifier
- `organization_name` (VARCHAR(100), UNIQUE, NOT NULL): Organization display name (case-insensitive uniqueness)
- `status` (VARCHAR(20), NOT NULL, DEFAULT 'pending_verification'): Tenant status
  - Values: `pending_verification`, `active`, `suspended`, `deleted`
- `created_at` (TIMESTAMP, NOT NULL, DEFAULT NOW()): Creation timestamp
- `verified_at` (TIMESTAMP, NULL): Email verification timestamp
- `updated_at` (TIMESTAMP, NOT NULL, DEFAULT NOW()): Last update timestamp

**Indexes**:
```sql
CREATE UNIQUE INDEX idx_tenants_org_name ON tenants(LOWER(organization_name));
CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_created_at ON tenants(created_at DESC);
```

**Validation Rules**:
- `organization_name`: 2-100 characters, alphanumeric + spaces + hyphens only
- `organization_name`: Must be unique (case-insensitive)
- `status`: Must be one of the allowed enum values
- `verified_at`: Can only be set when status changes to `active`

**State Transitions**:
```
pending_verification → active (on successful OTP verification)
active → suspended (admin action, future)
active → deleted (soft delete, future)
```

**Business Rules**:
- A tenant cannot be deleted if it has active users (future enforcement)
- Organization names are trimmed and normalized (spaces, case)
- Once verified, `verified_at` timestamp is immutable

---

### 2. AdminUser

**Purpose**: Represents the primary administrator user for a tenant

**Fields**:
- `id` (UUID, PRIMARY KEY): Unique user identifier
- `tenant_id` (UUID, FOREIGN KEY → tenants.id, NOT NULL): Associated tenant
- `full_name` (VARCHAR(100), NOT NULL): Administrator full name
- `email` (VARCHAR(255), UNIQUE, NOT NULL): Email address (case-insensitive uniqueness)
- `password_hash` (VARCHAR(255), NOT NULL): Bcrypt hashed password
- `role` (VARCHAR(20), NOT NULL, DEFAULT 'admin'): User role
  - Values: `admin`, `owner` (future: `member`, `viewer`)
- `created_at` (TIMESTAMP, NOT NULL, DEFAULT NOW()): Creation timestamp
- `updated_at` (TIMESTAMP, NOT NULL, DEFAULT NOW()): Last update timestamp

**Indexes**:
```sql
CREATE UNIQUE INDEX idx_users_email ON users(LOWER(email));
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_role ON users(role);
```

**Validation Rules**:
- `full_name`: 2-100 characters
- `email`: Valid RFC 5322 email format
- `email`: Must be unique across all tenants (case-insensitive)
- `password`: Minimum 8 characters, must contain:
  - At least one uppercase letter
  - At least one lowercase letter
  - At least one number
  - At least one special character (@$!%*?&)
- `password_hash`: Stored using bcrypt with cost factor 12

**Relationships**:
- Each admin user belongs to exactly one tenant (N:1)
- Each tenant has exactly one admin user initially (1:1, becomes 1:N in future)

**Business Rules**:
- Password must never be stored in plain text
- Email verification required before tenant activation
- Admin user cannot be deleted while tenant is active (future)
- Email changes require re-verification (future)

---

### 3. OTPVerification

**Purpose**: Stores email verification codes for tenant activation (short-lived data)

**Storage**: Redis (primary) with PostgreSQL fallback (audit trail)

**Redis Key Structure**:
```
Key: otp:{tenant_id}
Value: {otp_code}
TTL: 600 seconds (10 minutes)

Key: otp:attempts:{tenant_id}
Value: {attempt_count}
TTL: 600 seconds (10 minutes)
```

**PostgreSQL Schema** (audit/historical purposes):
- `id` (UUID, PRIMARY KEY): Unique OTP record identifier
- `tenant_id` (UUID, FOREIGN KEY → tenants.id, NOT NULL): Associated tenant
- `otp_code` (VARCHAR(6), NOT NULL): 6-digit numeric code
- `created_at` (TIMESTAMP, NOT NULL, DEFAULT NOW()): Generation timestamp
- `expires_at` (TIMESTAMP, NOT NULL): Expiration timestamp (created_at + 10 minutes)
- `verified_at` (TIMESTAMP, NULL): Verification timestamp
- `attempt_count` (INTEGER, NOT NULL, DEFAULT 0): Failed verification attempts
- `ip_address` (VARCHAR(45), NULL): Requester IP address
- `status` (VARCHAR(20), NOT NULL, DEFAULT 'pending'): OTP status
  - Values: `pending`, `verified`, `expired`, `invalidated`

**Indexes**:
```sql
CREATE INDEX idx_otp_tenant_id ON otp_verifications(tenant_id);
CREATE INDEX idx_otp_status ON otp_verifications(status);
CREATE INDEX idx_otp_created_at ON otp_verifications(created_at DESC);
```

**Validation Rules**:
- `otp_code`: Exactly 6 digits (000000-999999)
- `otp_code`: Cryptographically secure random generation
- `expires_at`: Exactly 10 minutes from `created_at`
- `attempt_count`: Maximum 5 attempts before invalidation

**Business Rules**:
- OTP can only be used once (single-use verification)
- OTP expires after 10 minutes
- Maximum 5 verification attempts per OTP
- Maximum 3 OTP generation requests per email per hour (rate limiting)
- Old OTPs are invalidated when a new OTP is requested
- Verified OTPs cannot be reused

**State Transitions**:
```
pending → verified (on successful verification)
pending → expired (after 10 minutes)
pending → invalidated (when new OTP is requested)
```

---

### 4. AuditLog

**Purpose**: Records all security-relevant events for compliance and monitoring

**Fields**:
- `id` (UUID, PRIMARY KEY): Unique log entry identifier
- `tenant_id` (UUID, FOREIGN KEY → tenants.id, NULL): Associated tenant (NULL for pre-verification events)
- `user_id` (UUID, NULL): Associated user (NULL for system events)
- `event_type` (VARCHAR(50), NOT NULL): Type of event
  - Values: `tenant_registration_initiated`, `tenant_registration_completed`, `otp_generated`, `otp_sent`, `otp_verification_attempted`, `otp_verification_succeeded`, `otp_verification_failed`, `otp_expired`, `tenant_verified`, `login_attempted`, `login_succeeded`, `login_failed`, `rate_limit_exceeded`
- `event_data` (JSONB, NULL): Event-specific metadata (structured)
- `ip_address` (VARCHAR(45), NULL): Client IP address
- `user_agent` (TEXT, NULL): Client user agent string
- `created_at` (TIMESTAMP, NOT NULL, DEFAULT NOW()): Event timestamp

**Indexes**:
```sql
CREATE INDEX idx_audit_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_event_type ON audit_logs(event_type);
CREATE INDEX idx_audit_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_event_data ON audit_logs USING GIN(event_data);
```

**Validation Rules**:
- `event_type`: Must be one of the predefined event types
- `event_data`: Valid JSON structure
- `created_at`: Immutable (no updates allowed)

**Business Rules**:
- Audit logs are append-only (no updates or deletes)
- All logs retained for minimum 90 days (compliance requirement)
- Sensitive data (passwords, OTP codes) never logged
- IP address and user agent logged for security analysis

**Event Data Examples**:

```json
// tenant_registration_initiated
{
  "organization_name": "Acme Corp",
  "admin_email": "admin@acme.com",
  "success": true
}

// otp_verification_attempted
{
  "otp_id": "uuid",
  "success": false,
  "reason": "invalid_code",
  "attempt_number": 3
}

// rate_limit_exceeded
{
  "endpoint": "/api/v1/resend-otp",
  "limit": 3,
  "window": "1 hour",
  "email": "admin@acme.com"
}
```

---

## Database Schema (PostgreSQL)

### Migration Scripts

**000001_create_tenants_table.up.sql**:
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending_verification',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMP NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_tenant_status CHECK (status IN ('pending_verification', 'active', 'suspended', 'deleted')),
    CONSTRAINT chk_org_name_length CHECK (LENGTH(TRIM(organization_name)) >= 2)
);

CREATE UNIQUE INDEX idx_tenants_org_name ON tenants(LOWER(organization_name));
CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_created_at ON tenants(created_at DESC);

COMMENT ON TABLE tenants IS 'Organizations using the SSO platform';
COMMENT ON COLUMN tenants.organization_name IS 'Unique organization name (case-insensitive)';
COMMENT ON COLUMN tenants.status IS 'Tenant lifecycle status';
```

**000001_create_tenants_table.down.sql**:
```sql
DROP TABLE IF EXISTS tenants CASCADE;
```

**000002_create_users_table.up.sql**:
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'admin',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_user_role CHECK (role IN ('admin', 'owner', 'member', 'viewer')),
    CONSTRAINT chk_full_name_length CHECK (LENGTH(TRIM(full_name)) >= 2),
    CONSTRAINT chk_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

CREATE UNIQUE INDEX idx_users_email ON users(LOWER(email));
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_role ON users(role);

COMMENT ON TABLE users IS 'User accounts associated with tenants';
COMMENT ON COLUMN users.tenant_id IS 'Associated tenant (foreign key)';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hashed password (cost factor 12)';
```

**000002_create_users_table.down.sql**:
```sql
DROP TABLE IF EXISTS users CASCADE;
```

**000003_create_audit_logs_table.up.sql**:
```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NULL REFERENCES tenants(id) ON DELETE SET NULL,
    user_id UUID NULL,
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB NULL,
    ip_address VARCHAR(45) NULL,
    user_agent TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_event_type_not_empty CHECK (LENGTH(TRIM(event_type)) > 0)
);

CREATE INDEX idx_audit_tenant_id ON audit_logs(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX idx_audit_user_id ON audit_logs(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_audit_event_type ON audit_logs(event_type);
CREATE INDEX idx_audit_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_event_data ON audit_logs USING GIN(event_data);

COMMENT ON TABLE audit_logs IS 'Append-only security and compliance audit trail';
COMMENT ON COLUMN audit_logs.event_type IS 'Type of event (e.g., tenant_registration_initiated)';
COMMENT ON COLUMN audit_logs.event_data IS 'Event-specific structured metadata (JSON)';
```

**000003_create_audit_logs_table.down.sql**:
```sql
DROP TABLE IF EXISTS audit_logs CASCADE;
```

---

## Redis Data Structures

### OTP Storage

**Key**: `otp:{tenant_id}`  
**Type**: String  
**Value**: `{otp_code}` (6-digit numeric string)  
**TTL**: 600 seconds (10 minutes)

**Example**:
```
Key: otp:123e4567-e89b-12d3-a456-426614174000
Value: 123456
TTL: 600
```

### OTP Attempt Tracking

**Key**: `otp:attempts:{tenant_id}`  
**Type**: Integer  
**Value**: Attempt count (0-5)  
**TTL**: 600 seconds (10 minutes, same as OTP)

**Example**:
```
Key: otp:attempts:123e4567-e89b-12d3-a456-426614174000
Value: 2
TTL: 600
```

### Rate Limiting - OTP Requests

**Key**: `rate:otp:request:{email}`  
**Type**: Integer  
**Value**: Request count  
**TTL**: 3600 seconds (1 hour)

**Example**:
```
Key: rate:otp:request:admin@acme.com
Value: 2
TTL: 3600
```

### Rate Limiting - Login Attempts (Future)

**Key**: `rate:login:{email}`  
**Type**: Integer  
**Value**: Failed login count  
**TTL**: 900 seconds (15 minutes)

---

## Data Access Patterns

### Tenant Registration Flow

1. **Check organization name availability**:
   ```sql
   SELECT COUNT(*) FROM tenants 
   WHERE LOWER(organization_name) = LOWER($1);
   ```

2. **Check email availability**:
   ```sql
   SELECT COUNT(*) FROM users 
   WHERE LOWER(email) = LOWER($1);
   ```

3. **Create tenant**:
   ```sql
   INSERT INTO tenants (organization_name, status)
   VALUES ($1, 'pending_verification')
   RETURNING id, created_at;
   ```

4. **Create admin user**:
   ```sql
   INSERT INTO users (tenant_id, full_name, email, password_hash, role)
   VALUES ($1, $2, $3, $4, 'admin')
   RETURNING id, created_at;
   ```

5. **Generate and store OTP in Redis**:
   ```redis
   SET otp:{tenant_id} {otp_code} EX 600
   SET otp:attempts:{tenant_id} 0 EX 600
   ```

6. **Log audit event**:
   ```sql
   INSERT INTO audit_logs (tenant_id, event_type, event_data, ip_address)
   VALUES ($1, 'tenant_registration_initiated', $2, $3);
   ```

### OTP Verification Flow

1. **Retrieve OTP from Redis**:
   ```redis
   GET otp:{tenant_id}
   GET otp:attempts:{tenant_id}
   ```

2. **Verify OTP code** (application logic):
   - Compare provided code with stored code
   - Check attempt count < 5

3. **On success**:
   ```sql
   UPDATE tenants 
   SET status = 'active', verified_at = NOW(), updated_at = NOW()
   WHERE id = $1;
   ```
   ```redis
   DEL otp:{tenant_id}
   DEL otp:attempts:{tenant_id}
   ```

4. **On failure**:
   ```redis
   INCR otp:attempts:{tenant_id}
   ```

5. **Log audit event**:
   ```sql
   INSERT INTO audit_logs (tenant_id, event_type, event_data, ip_address)
   VALUES ($1, 'otp_verification_attempted', $2, $3);
   ```

### Tenant Dashboard Query

```sql
SELECT 
    t.id,
    t.organization_name,
    t.status,
    t.created_at,
    t.verified_at,
    u.full_name,
    u.email
FROM tenants t
INNER JOIN users u ON u.tenant_id = t.id
WHERE t.id = $1 AND u.role = 'admin';
```

---

## Data Retention and Archival (Future)

**Active Data**:
- Tenants: Retained while status is `active` or `suspended`
- Users: Retained while associated tenant is active
- Audit Logs: Minimum 90 days, maximum 1 year (configurable)

**Deleted Data**:
- Tenants: Soft delete (status = `deleted`), hard delete after 30 days
- Users: Cascade delete with tenant or soft delete individually
- Audit Logs: Archived to cold storage after 1 year

**OTP Data**:
- Redis: Automatic expiration after 10 minutes (TTL)
- PostgreSQL: Retained for audit purposes (90 days), then purged
