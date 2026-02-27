# Data Model: OAuth Client Application Registration Portal

**Date**: 2026-02-25  
**Feature**: [spec.md](./spec.md)  
**Research**: [research.md](./research.md)

## Overview

This document defines the data model changes for the OAuth Client Application Registration Portal. The feature extends the existing `clients` table and adds Redis caching. No new PostgreSQL tables are required.

**Storage**:

- **PostgreSQL**: Client records (extended with `client_type`, `description`), audit logs (existing table)
- **Redis**: Tenant client count cache, revoked client blacklist

---

## PostgreSQL Schema Changes

### Migration: 000008_add_client_type_and_description.up.sql

Extends the existing `clients` table to support public/confidential client types and application descriptions.

```sql
-- Add client_type column (confidential or public)
ALTER TABLE clients
    ADD COLUMN client_type VARCHAR(20) NOT NULL DEFAULT 'confidential';

-- Add description column for application metadata
ALTER TABLE clients
    ADD COLUMN description TEXT;

-- Relax client_secret NOT NULL constraint for public clients
ALTER TABLE clients
    ALTER COLUMN client_secret DROP NOT NULL;

-- Add check constraint for valid client types
ALTER TABLE clients
    ADD CONSTRAINT clients_valid_client_type
    CHECK (client_type IN ('confidential', 'public'));

-- Add check constraint: confidential clients MUST have a secret
ALTER TABLE clients
    ADD CONSTRAINT clients_secret_required_for_confidential
    CHECK (client_type = 'public' OR client_secret IS NOT NULL);

-- Add index for client_type queries
CREATE INDEX idx_clients_client_type ON clients(client_type);

-- Add composite index for tenant + active + type queries
CREATE INDEX idx_clients_tenant_active_type ON clients(tenant_id, is_active, client_type);

-- Add index for client_name search (ILIKE performance)
CREATE INDEX idx_clients_client_name_trgm ON clients USING gin (client_name gin_trgm_ops);
```

> **Note**: The trigram index (`gin_trgm_ops`) requires the `pg_trgm` extension. If not already enabled, add `CREATE EXTENSION IF NOT EXISTS pg_trgm;` to the migration.

### Migration: 000008_add_client_type_and_description.down.sql

```sql
-- Remove indexes
DROP INDEX IF EXISTS idx_clients_client_name_trgm;
DROP INDEX IF EXISTS idx_clients_tenant_active_type;
DROP INDEX IF EXISTS idx_clients_client_type;

-- Remove constraints
ALTER TABLE clients DROP CONSTRAINT IF EXISTS clients_secret_required_for_confidential;
ALTER TABLE clients DROP CONSTRAINT IF EXISTS clients_valid_client_type;

-- Restore NOT NULL on client_secret (set NULL values to empty string first)
UPDATE clients SET client_secret = '' WHERE client_secret IS NULL;
ALTER TABLE clients ALTER COLUMN client_secret SET NOT NULL;

-- Remove columns
ALTER TABLE clients DROP COLUMN IF EXISTS description;
ALTER TABLE clients DROP COLUMN IF EXISTS client_type;
```

---

## Updated Entity: Client

### Domain Entity (`domain/entities/client.go`)

```go
// Client type constants
const (
    ClientTypeConfidential = "confidential"
    ClientTypePublic       = "public"
)

// MaxClientsPerTenant defines the maximum number of OAuth clients per tenant
const MaxClientsPerTenant = 25

// Client represents an OAuth 2.0 / OIDC client application
type Client struct {
    ClientID            string
    TenantID            string
    ClientName          string
    Description         string     // Application description (optional)
    ClientType          string     // "confidential" or "public"
    ClientSecretHash    string     // NULL for public clients
    AllowedRedirectURIs []string
    IsActive            bool
    CreatedAt           time.Time
    UpdatedAt           time.Time
}
```

### Field Details

| Field                 | Type      | Required    | Description                                                       |
| --------------------- | --------- | ----------- | ----------------------------------------------------------------- |
| `ClientID`            | string    | Yes         | Unique identifier (24 bytes, base64url encoded)                   |
| `TenantID`            | string    | Yes         | FK to tenants table; determines tenant context                    |
| `ClientName`          | string    | Yes         | Human-readable name (3-100 chars)                                 |
| `Description`         | string    | No          | Application description (max 500 chars)                           |
| `ClientType`          | string    | Yes         | `"confidential"` or `"public"`                                    |
| `ClientSecretHash`    | string    | Conditional | bcrypt hash of secret; required for confidential, NULL for public |
| `AllowedRedirectURIs` | []string  | Yes         | At least one valid redirect URI                                   |
| `IsActive`            | bool      | Yes         | Active/revoked status (default: true)                             |
| `CreatedAt`           | time.Time | Yes         | Record creation timestamp                                         |
| `UpdatedAt`           | time.Time | Yes         | Last modification timestamp                                       |

### Validation Rules

| Rule                                                  | Enforcement                                      |
| ----------------------------------------------------- | ------------------------------------------------ |
| `ClientName` 3-100 characters                         | Domain entity `Validate()`                       |
| `Description` max 500 characters                      | Domain entity `Validate()`                       |
| `ClientType` must be "confidential" or "public"       | Domain entity `Validate()` + DB CHECK constraint |
| Confidential clients must have `ClientSecretHash`     | DB CHECK constraint + `Validate()`               |
| At least one redirect URI                             | DB CHECK constraint + `Validate()`               |
| Redirect URIs must be valid URLs with scheme and host | Domain entity `ValidateRedirectURI()`            |
| Redirect URIs must use HTTPS (except localhost)       | Domain entity `ValidateRedirectURIStrict()`      |
| No URI fragments allowed                              | Domain entity `ValidateRedirectURI()`            |
| Max 25 clients per tenant                             | Use case layer quota check                       |

---

## Updated Repository Interface

### ClientRepository (`domain/repositories/client_repository.go`)

New methods added to existing interface:

```go
type ClientRepository interface {
    // ... existing methods ...

    // CountByTenant returns the number of active clients for a tenant
    CountByTenant(ctx context.Context, tenantID string) (int, error)

    // ListByTenantPaginated retrieves clients with pagination and optional search
    // search: case-insensitive partial match on client_name (empty = no filter)
    // page: 1-based page number
    // pageSize: items per page (max 25)
    // Returns: clients, total count, error
    ListByTenantPaginated(ctx context.Context, tenantID string, search string, page int, pageSize int) ([]*entities.Client, int, error)
}
```

### RefreshTokenRepository (existing, new method)

```go
type RefreshTokenRepository interface {
    // ... existing methods ...

    // RevokeByClientID revokes all refresh tokens issued to a specific client
    RevokeByClientID(ctx context.Context, clientID string) error
}
```

---

## Redis Data Structures

### 1. Tenant Client Count Cache

**Purpose**: Cache the number of active clients per tenant to avoid frequent COUNT queries during registration.

```
Key:     client_count:{tenant_id}
Value:   integer (e.g., "7")
TTL:     60 seconds
```

**Operations**:

- **Read**: `GET client_count:{tenant_id}` — cache miss falls back to `CountByTenant()` DB query
- **Invalidate**: `DEL client_count:{tenant_id}` — on client create or delete
- **Note**: This is an optimization. The database count is the source of truth with row-level locking.

### 2. Revoked Client Blacklist

**Purpose**: Immediately invalidate access tokens for deleted clients without waiting for JWT expiry.

```
Key:     revoked_client:{client_id}
Value:   "1" (presence check only)
TTL:     900 seconds (15 minutes = max access token lifetime)
```

**Operations**:

- **Set**: `SET revoked_client:{client_id} "1" EX 900` — on client deletion
- **Check**: `EXISTS revoked_client:{client_id}` — during access token validation middleware
- **Auto-cleanup**: Keys expire after 15 minutes (no manual cleanup needed)

---

## Entity Relationships

```
┌──────────┐       ┌──────────────┐       ┌────────────────────┐
│ Tenants  │──1:N──│   Clients    │──1:N──│ Refresh Tokens     │
│          │       │              │       │                    │
│ id       │       │ client_id    │       │ client_id (FK)     │
│ name     │       │ tenant_id    │       │ token_hash         │
│ domain   │       │ client_name  │       │ expires_at         │
│ is_active│       │ description  │       └────────────────────┘
└──────────┘       │ client_type  │
                   │ client_secret│       ┌────────────────────┐
                   │ redirect_uris│──audit│ Audit Logs         │
                   │ is_active    │───────│                    │
                   │ created_at   │       │ event_type         │
                   │ updated_at   │       │ tenant_id          │
                   └──────────────┘       │ actor_id           │
                                          │ resource_id        │
                                          │ details (JSONB)    │
                                          └────────────────────┘
```

**Relationships**:

- **Tenant → Client**: One-to-many. Each client belongs to exactly one tenant. Max 25 per tenant.
- **Client → Refresh Tokens**: One-to-many. Tokens reference the client that issued them. Cascade revocation on client deletion.
- **Client → Audit Logs**: Logical relationship. Audit entries reference the client_id as `resource_id`.

---

## Data Access Patterns

| Operation                | Query Pattern                                 | Frequency             | Index Used                                                       |
| ------------------------ | --------------------------------------------- | --------------------- | ---------------------------------------------------------------- |
| Create client            | INSERT + CountByTenant check                  | Low (admin action)    | `idx_clients_tenant`                                             |
| List clients (paginated) | SELECT WHERE tenant_id + ILIKE + LIMIT/OFFSET | Medium                | `idx_clients_tenant_active_type`, `idx_clients_client_name_trgm` |
| Get client by ID         | SELECT WHERE client_id AND tenant_id          | Medium                | PK + `idx_clients_tenant`                                        |
| Update client            | UPDATE WHERE client_id                        | Low                   | PK                                                               |
| Delete client (soft)     | UPDATE is_active=false + revoke tokens        | Low                   | PK                                                               |
| Count by tenant          | SELECT COUNT WHERE tenant_id AND is_active    | Low (cached in Redis) | `idx_clients_tenant_active_type`                                 |
| Validate credentials     | SELECT WHERE client_id + bcrypt compare       | High (OAuth flow)     | PK                                                               |

---

## GORM Model (Infrastructure Layer)

```go
// PostgresClient is the GORM model for the clients table
type PostgresClient struct {
    ClientID            string         `gorm:"primaryKey;column:client_id"`
    TenantID            string         `gorm:"column:tenant_id;not null;index:idx_clients_tenant"`
    ClientName          string         `gorm:"column:client_name;not null"`
    Description         *string        `gorm:"column:description"`
    ClientType          string         `gorm:"column:client_type;not null;default:confidential"`
    ClientSecret        *string        `gorm:"column:client_secret"`
    RedirectURIs        pq.StringArray `gorm:"column:redirect_uris;type:text[];not null"`
    IsActive            bool           `gorm:"column:is_active;not null;default:true"`
    CreatedAt           time.Time      `gorm:"column:created_at;not null"`
    UpdatedAt           time.Time      `gorm:"column:updated_at;not null"`
}

func (PostgresClient) TableName() string { return "clients" }
```

---

## Migration Checklist

- [ ] Enable `pg_trgm` extension (if not already)
- [ ] Apply `000008_add_client_type_and_description.up.sql`
- [ ] Verify all existing clients default to `client_type = 'confidential'`
- [ ] Verify CHECK constraints are active
- [ ] Run `ANALYZE clients` after migration to update query planner statistics
