# Data Model: End-User Management with RBAC

**Date**: June 11, 2025  
**Feature**: [spec.md](./spec.md)

## Overview

This document defines the database schema changes for the End-User Management with RBAC feature. The existing `users` and `user_role_assignments` tables (introduced in `003-sso-auth-provider`) provide the foundational schema. This feature extends those tables, introduces an `invitations` table, adds a `user_events` table for activity logging, and adds a Redis-backed user blacklist for immediate access-token invalidation.

**Storage**:

- **PostgreSQL**: Users (extended), invitations, role assignments (extended), user activity events, audit logs (existing)
- **Redis**: User blacklist (for immediate access-token invalidation on delete/disable), active session count cache

**Multi-tenancy strategy**: Shared schema with `tenant_id` column on every table (Row-Level Security consistent with existing tables).

---

## PostgreSQL Schema Changes

### Migration Order

Apply migrations in order due to foreign key dependencies:

1. `000009_extend_users.up.sql` — Extend existing `users` table
2. `000010_create_invitations.up.sql` — New `invitations` table
3. `000011_extend_user_role_assignments.up.sql` — Extend existing `user_role_assignments` table
4. `000012_create_user_events.up.sql` — New `user_events` table for activity log

---

### Migration: 000009_extend_users

Extends the existing `users` table with display name, account status enum, and last-login tracking.

```sql
-- migrations/000009_extend_users.up.sql

-- Add display_name for UI presentation
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS display_name VARCHAR(255);

-- Add status with pending/active/disabled lifecycle
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active';

ALTER TABLE users
    ADD CONSTRAINT users_valid_status
    CHECK (status IN ('pending', 'active', 'disabled'));

-- Track last successful authentication
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;

-- Index on status for admin filter queries
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- Composite index for tenant + status admin queries
CREATE INDEX IF NOT EXISTS idx_users_tenant_status ON users(tenant_id, status);

-- Composite index for tenant + email (case-insensitive search)
CREATE INDEX IF NOT EXISTS idx_users_tenant_email_lower ON users(tenant_id, LOWER(email));

-- Trigram index for display_name and email search (requires pg_trgm)
CREATE INDEX IF NOT EXISTS idx_users_display_name_trgm ON users USING gin (display_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_email_trgm ON users USING gin (email gin_trgm_ops);


-- migrations/000009_extend_users.down.sql
DROP INDEX IF EXISTS idx_users_email_trgm;
DROP INDEX IF EXISTS idx_users_display_name_trgm;
DROP INDEX IF EXISTS idx_users_tenant_email_lower;
DROP INDEX IF EXISTS idx_users_tenant_status;
DROP INDEX IF EXISTS idx_users_status;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_valid_status;
ALTER TABLE users DROP COLUMN IF EXISTS last_login_at;
ALTER TABLE users DROP COLUMN IF EXISTS status;
ALTER TABLE users DROP COLUMN IF EXISTS display_name;
```

---

### Migration: 000010_create_invitations

New table to track pending user invitations.

```sql
-- migrations/000010_create_invitations.up.sql
CREATE TABLE IF NOT EXISTS invitations (
    id               VARCHAR(255) PRIMARY KEY,
    tenant_id        VARCHAR(255) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email            VARCHAR(255) NOT NULL,
    display_name     VARCHAR(255),
    token_hash       VARCHAR(255) NOT NULL UNIQUE,  -- bcrypt hash of the invitation token
    status           VARCHAR(20)  NOT NULL DEFAULT 'pending',
    invited_by       VARCHAR(255) NOT NULL REFERENCES users(id),
    expires_at       TIMESTAMP WITH TIME ZONE NOT NULL,
    accepted_at      TIMESTAMP WITH TIME ZONE,
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT invitations_valid_status CHECK (status IN ('pending', 'accepted', 'expired')),
    -- At most one pending invitation per email per tenant
    CONSTRAINT invitations_unique_pending_email UNIQUE (tenant_id, email, status)
        DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX idx_invitations_tenant ON invitations(tenant_id);
CREATE INDEX idx_invitations_email ON invitations(tenant_id, email);
CREATE INDEX idx_invitations_status ON invitations(status);
CREATE INDEX idx_invitations_expires_at ON invitations(expires_at);

CREATE TRIGGER update_invitations_updated_at BEFORE UPDATE ON invitations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- migrations/000010_create_invitations.down.sql
DROP TRIGGER IF EXISTS update_invitations_updated_at ON invitations;
DROP TABLE IF EXISTS invitations CASCADE;
```

---

### Migration: 000011_extend_user_role_assignments

Extends the existing `user_role_assignments` table with revocation metadata to preserve full audit history via soft deletes.

```sql
-- migrations/000011_extend_user_role_assignments.up.sql

-- Track who revoked an assignment and when
ALTER TABLE user_role_assignments
    ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE user_role_assignments
    ADD COLUMN IF NOT EXISTS revoked_by VARCHAR(255) REFERENCES users(id);

-- Index for active-only role lookups (most common query path)
CREATE INDEX IF NOT EXISTS idx_ura_user_client_active
    ON user_role_assignments(user_id, client_id)
    WHERE is_active = true;


-- migrations/000011_extend_user_role_assignments.down.sql
DROP INDEX IF EXISTS idx_ura_user_client_active;
ALTER TABLE user_role_assignments DROP COLUMN IF EXISTS revoked_by;
ALTER TABLE user_role_assignments DROP COLUMN IF EXISTS revoked_at;
```

---

### Migration: 000012_create_user_events

New table for per-user activity log (login events, session terminations, administrative actions).

```sql
-- migrations/000012_create_user_events.up.sql
CREATE TABLE IF NOT EXISTS user_events (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     VARCHAR(255) NOT NULL,                              -- denormalized for partition pruning
    user_id       VARCHAR(255) NOT NULL,
    client_id     VARCHAR(255),                                       -- NULL for non-OAuth events
    event_type    VARCHAR(50)  NOT NULL,
    ip_address    INET,
    country_code  CHAR(2),
    details       JSONB        NOT NULL DEFAULT '{}',
    occurred_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT user_events_valid_type CHECK (event_type IN (
        'login_success',
        'login_failure',
        'token_refresh',
        'logout',
        'session_terminated',
        'account_disabled',
        'account_enabled',
        'role_assigned',
        'role_revoked',
        'user_invited',
        'invitation_accepted',
        'invitation_expired',
        'invitation_resent',
        'user_deleted'
    ))
);

-- Primary access pattern: list events for a user, most recent first
CREATE INDEX idx_user_events_user_recent
    ON user_events(user_id, occurred_at DESC);

-- Admin filtering by tenant + event type
CREATE INDEX idx_user_events_tenant_type
    ON user_events(tenant_id, event_type, occurred_at DESC);

-- Cleanup: find events older than retention window
CREATE INDEX idx_user_events_occurred_at
    ON user_events(occurred_at);


-- migrations/000012_create_user_events.down.sql
DROP TABLE IF EXISTS user_events CASCADE;
```

> **Retention cleanup**: A scheduled job (cron/pg_cron) MUST delete rows where `occurred_at < NOW() - INTERVAL '90 days'` to enforce the 90-day retention requirement. Example:
> ```sql
> DELETE FROM user_events WHERE occurred_at < NOW() - INTERVAL '90 days';
> ```

---

## Updated Entity: User

### Domain Entity (`domain/entities/user.go`)

```go
// UserStatus represents the lifecycle state of a user account
type UserStatus string

const (
    UserStatusPending  UserStatus = "pending"
    UserStatusActive   UserStatus = "active"
    UserStatusDisabled UserStatus = "disabled"
)

// MaxUsersPerTenant defines the tenant user quota
const MaxUsersPerTenant = 10_000

// User represents an end-user who can authenticate via the tenant's SSO
type User struct {
    ID           string
    TenantID     string
    Email        string      // unique per tenant, lowercased
    DisplayName  string      // optional, for UI presentation
    PasswordHash string      // bcrypt hash; empty for pending (no password set yet)
    Status       UserStatus  // pending | active | disabled
    LastLoginAt  *time.Time
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### Field Details

| Field          | Type        | Required | Description                                              |
| -------------- | ----------- | -------- | -------------------------------------------------------- |
| `ID`           | string      | Yes      | Unique user identifier (UUID v4)                         |
| `TenantID`     | string      | Yes      | FK to tenants; determines tenant scope                   |
| `Email`        | string      | Yes      | RFC 5322 address, lowercased, unique per tenant          |
| `DisplayName`  | string      | No       | Human-readable name (max 255 chars); optional            |
| `PasswordHash` | string      | No       | bcrypt hash (cost ≥ 12); empty until invitation accepted |
| `Status`       | UserStatus  | Yes      | `pending` → `active` → `disabled` lifecycle             |
| `LastLoginAt`  | *time.Time  | No       | Updated on every successful authentication               |
| `CreatedAt`    | time.Time   | Yes      | Record creation timestamp                                |
| `UpdatedAt`    | time.Time   | Yes      | Last modification timestamp                              |

---

## New Entity: Invitation

### Domain Entity (`domain/entities/invitation.go`)

```go
// InvitationStatus represents the lifecycle state of an invitation
type InvitationStatus string

const (
    InvitationStatusPending  InvitationStatus = "pending"
    InvitationStatusAccepted InvitationStatus = "accepted"
    InvitationStatusExpired  InvitationStatus = "expired"
)

// InvitationTTL is the duration a token remains valid after issuance
const InvitationTTL = 72 * time.Hour

// Invitation represents a pending user activation request
type Invitation struct {
    ID          string
    TenantID    string
    Email       string
    DisplayName string
    TokenHash   string           // bcrypt hash of the one-time token
    Status      InvitationStatus // pending | accepted | expired
    InvitedBy   string           // user ID of the sending administrator
    ExpiresAt   time.Time
    AcceptedAt  *time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

---

## Updated Entity: Role Assignment

### Domain Entity (`domain/entities/role_assignment.go`)

```go
// RoleAssignment represents the grant of a named role to a user for a specific client
type RoleAssignment struct {
    ID         int
    UserID     string
    ClientID   string
    TenantID   string
    RoleName   string     // free-form, 1-100 chars
    IsActive   bool
    GrantedAt  time.Time
    GrantedBy  string     // admin user ID
    RevokedAt  *time.Time
    RevokedBy  *string    // admin user ID who revoked
}

// MaxRoleNameLength is the maximum allowed length for a role name
const MaxRoleNameLength = 100
```

---

## Redis Data Structures

### 1. User Blacklist (Immediate Token Invalidation)

**Purpose**: Immediately invalidate access tokens for deleted or disabled users without waiting for JWT natural expiry.

```
Key:     user_blacklist:{user_id}
Value:   "1" (presence-check only)
TTL:     900 seconds (15 minutes = max access token lifetime)
```

**Operations**:
- **Set** (on user delete or disable): `SET user_blacklist:{user_id} "1" EX 900`
- **Check** (during access token validation middleware): `EXISTS user_blacklist:{user_id}`
- **Auto-cleanup**: Keys expire automatically; no manual cleanup required

### 2. Tenant User Count Cache

**Purpose**: Avoid frequent COUNT queries when enforcing the 10,000-user quota.

```
Key:     user_count:{tenant_id}
Value:   integer string (e.g., "342")
TTL:     60 seconds
```

**Operations**:
- **Read**: `GET user_count:{tenant_id}` → cache miss falls back to COUNT query
- **Invalidate**: `DEL user_count:{tenant_id}` → on user create (invitation sent) or user delete

### 3. Invitation Token Lookup (Fast Validation)

**Purpose**: Fast O(1) lookup of invitation token validity during the activation flow, avoiding a full bcrypt comparison for non-existent tokens.

```
Key:     invitation_exists:{token_prefix_8_chars}
Value:   invitation_id
TTL:     259200 seconds (72 hours, matching InvitationTTL)
```

> **Note**: The raw token is never stored in Redis. Only a short prefix (first 8 characters) is indexed for existence checking. The actual token is always validated via bcrypt against the database `token_hash` column.

---

## Entity Relationships

```
┌──────────┐       ┌──────────────────┐       ┌────────────────────────┐
│ Tenants  │──1:N──│     Users        │──1:N──│  Role Assignments      │
│          │       │                  │       │                        │
│ id       │       │ id               │       │ id                     │
│ name     │       │ tenant_id        │       │ user_id (FK)           │
│ domain   │       │ email            │       │ client_id (FK)         │
│ is_active│       │ display_name     │       │ tenant_id              │
└──────────┘       │ password_hash    │       │ role_name              │
     │             │ status           │       │ is_active              │
     │             │ last_login_at    │       │ granted_by             │
     │             └──────────────────┘       │ revoked_by             │
     │                      │                 └────────────────────────┘
     │                      │
     │             ┌──────────────────┐       ┌────────────────────────┐
     │─────1:N────►│  Invitations     │       │  User Events           │
                   │                  │       │                        │
                   │ id               │       │ id (BIGSERIAL)         │
                   │ tenant_id        │       │ tenant_id              │
                   │ email            │       │ user_id                │
                   │ token_hash       │       │ client_id (nullable)   │
                   │ status           │       │ event_type             │
                   │ invited_by (FK)  │       │ ip_address             │
                   │ expires_at       │       │ details (JSONB)        │
                   └──────────────────┘       │ occurred_at            │
                                              └────────────────────────┘

┌──────────┐       ┌──────────────────┐
│ Clients  │──1:N──│ Role Assignments │
│          │       │ (see above)      │
│ client_id│       └──────────────────┘
│ tenant_id│
└──────────┘──1:N──► Refresh Tokens (existing, from 003)
```

**Key relationships**:

- **Tenant → User**: One-to-many. Each user belongs to exactly one tenant. Max 10,000 users per tenant.
- **Tenant → Invitation**: One-to-many. Invitations are scoped to the tenant; at most one `pending` invitation per email per tenant.
- **User → Role Assignment**: One-to-many. A user may have multiple role assignments across multiple clients.
- **Client → Role Assignment**: One-to-many. A client may have many users assigned roles.
- **User → User Events**: One-to-many. Chronological activity log. Retained 90 days.
- **User → Refresh Tokens**: One-to-many. Existing relationship from 003; surfaced as "sessions" in this feature.

---

## Repository Interface Changes

### UserRepository (`domain/repositories/user_repository.go`)

New and modified methods added to the existing interface:

```go
type UserRepository interface {
    // --- existing methods (from 003) ---
    GetByID(ctx context.Context, userID string) (*entities.User, error)
    GetByEmail(ctx context.Context, tenantID, email string) (*entities.User, error)
    Create(ctx context.Context, user *entities.User) error
    Update(ctx context.Context, user *entities.User) error

    // --- new methods for 005 ---

    // ListByTenant returns paginated users filtered by optional status and search term.
    // search: case-insensitive partial match on display_name or email (empty = no filter)
    // status: filter by account status (empty = all statuses)
    // Returns: users, total count across all pages, error
    ListByTenant(ctx context.Context, tenantID string, search string, status entities.UserStatus, page int, pageSize int) ([]*entities.User, int, error)

    // CountByTenant returns the total number of users (all statuses) in a tenant.
    CountByTenant(ctx context.Context, tenantID string) (int, error)

    // UpdateStatus updates the account status and triggers session revocation side-effects (handled at use-case layer).
    UpdateStatus(ctx context.Context, userID string, status entities.UserStatus) error

    // UpdateLastLogin records the most recent successful authentication timestamp.
    UpdateLastLogin(ctx context.Context, userID string, at time.Time) error

    // Delete permanently removes the user record. Cascade deletes handle role assignments and invitations.
    Delete(ctx context.Context, userID string) error
}
```

### InvitationRepository (`domain/repositories/invitation_repository.go`)

New repository:

```go
type InvitationRepository interface {
    // Create inserts a new invitation record.
    Create(ctx context.Context, inv *entities.Invitation) error

    // GetByToken fetches a pending invitation by validating the provided plain token
    // against stored hashes. Returns ErrNotFound if no match or token is expired/used.
    GetByToken(ctx context.Context, plainToken string) (*entities.Invitation, error)

    // GetPendingByEmail returns the active pending invitation for an email in a tenant, if any.
    GetPendingByEmail(ctx context.Context, tenantID, email string) (*entities.Invitation, error)

    // UpdateStatus transitions an invitation's status (e.g., pending → accepted or expired).
    UpdateStatus(ctx context.Context, invitationID string, status entities.InvitationStatus, acceptedAt *time.Time) error

    // ListByTenant returns all invitations for a tenant (admin view).
    ListByTenant(ctx context.Context, tenantID string, page int, pageSize int) ([]*entities.Invitation, int, error)

    // ExpireStalePending marks all pending invitations past their expires_at as expired.
    // Intended to be called by a scheduled background job.
    ExpireStalePending(ctx context.Context) (int64, error)
}
```

### RoleAssignmentRepository (`domain/repositories/role_assignment_repository.go`)

Extended from existing (003) interface:

```go
type RoleAssignmentRepository interface {
    // --- existing methods ---
    // GetActiveRoles returns role names for a (user, client) pair.
    GetActiveRoles(ctx context.Context, userID, clientID string) ([]string, error)

    // HasAnyRole returns true if the user has at least one active role for the client.
    HasAnyRole(ctx context.Context, userID, clientID string) (bool, error)

    // --- new methods for 005 ---

    // Assign creates a new role assignment. Returns ErrDuplicateRole if already active.
    Assign(ctx context.Context, assignment *entities.RoleAssignment) error

    // Revoke soft-deletes a role assignment by ID, recording revokedBy and revokedAt.
    Revoke(ctx context.Context, assignmentID int, revokedByUserID string) error

    // ListByUser returns all role assignments for a user, grouped by client, including inactive ones.
    ListByUser(ctx context.Context, userID string) ([]*entities.RoleAssignment, error)

    // ListByClient returns all active role assignments for a client application (admin view).
    ListByClient(ctx context.Context, clientID string, page int, pageSize int) ([]*entities.RoleAssignment, int, error)

    // RevokeAllForUser revokes all active role assignments for a user (used on account disable/delete).
    RevokeAllForUser(ctx context.Context, userID, revokedByUserID string) error
}
```

### UserEventRepository (`domain/repositories/user_event_repository.go`)

New repository:

```go
type UserEventRepository interface {
    // Record inserts a new event. Fire-and-forget; errors are logged but do not block the caller.
    Record(ctx context.Context, event *entities.UserEvent) error

    // ListByUser returns paginated activity events for a user in reverse chronological order.
    ListByUser(ctx context.Context, userID string, page int, pageSize int) ([]*entities.UserEvent, int, error)

    // DeleteOlderThan deletes events older than the given timestamp (used by retention job).
    DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}
```

---

## Data Access Patterns

| Operation                         | Query Pattern                                                         | Frequency              | Index Used                                    |
| --------------------------------- | --------------------------------------------------------------------- | ---------------------- | --------------------------------------------- |
| List users (paginated + search)   | SELECT WHERE tenant_id + ILIKE + status + LIMIT/OFFSET                | Medium (admin)         | `idx_users_tenant_status`, trgm indexes       |
| Get user by email                 | SELECT WHERE tenant_id + LOWER(email)                                 | High (auth flow)       | `idx_users_tenant_email_lower`                |
| Count users per tenant            | SELECT COUNT WHERE tenant_id                                          | Low (cached in Redis)  | `idx_users_tenant_status`                     |
| Update user status                | UPDATE status WHERE id                                                | Low (admin action)     | PK                                            |
| Get active roles for token        | SELECT role_name WHERE user_id + client_id + is_active = true         | High (token issuance)  | `idx_ura_user_client_active`                  |
| Check any role exists             | SELECT EXISTS WHERE user_id + client_id + is_active = true            | High (auth flow)       | `idx_ura_user_client_active`                  |
| Assign role                       | INSERT with UNIQUE constraint guard                                   | Low (admin action)     | PK + unique constraint                        |
| Revoke role                       | UPDATE is_active=false WHERE id                                       | Low (admin action)     | PK                                            |
| Validate invitation token         | SELECT WHERE token_hash bcrypt compare + status='pending' + expires   | Low (one-time)         | `token_hash` UNIQUE index                     |
| Expire stale invitations          | UPDATE status='expired' WHERE expires_at < NOW()                      | Scheduled (background) | `idx_invitations_expires_at`                  |
| List user activity events         | SELECT WHERE user_id ORDER BY occurred_at DESC + LIMIT/OFFSET         | Medium (admin)         | `idx_user_events_user_recent`                 |
| Revoke all sessions for user      | UPDATE refresh_tokens SET is_revoked=true WHERE user_id               | Low (admin action)     | `idx_refresh_tokens_user`                     |
| Check user blacklist (Redis)      | EXISTS user_blacklist:{user_id}                                       | High (every API call)  | Redis O(1)                                    |

---

## GORM Models (Infrastructure Layer)

```go
// PostgresUser is the GORM model for the users table (extended for 005)
type PostgresUser struct {
    ID           string         `gorm:"primaryKey;column:id"`
    TenantID     string         `gorm:"column:tenant_id;not null;index:idx_users_tenant"`
    Email        string         `gorm:"column:email;not null"`
    DisplayName  *string        `gorm:"column:display_name"`
    PasswordHash string         `gorm:"column:password_hash;not null"`
    Status       string         `gorm:"column:status;not null;default:active"`
    LastLoginAt  *time.Time     `gorm:"column:last_login_at"`
    IsActive     bool           `gorm:"column:is_active;not null;default:true"`
    CreatedAt    time.Time      `gorm:"column:created_at;not null"`
    UpdatedAt    time.Time      `gorm:"column:updated_at;not null"`
}
func (PostgresUser) TableName() string { return "users" }

// PostgresInvitation is the GORM model for the invitations table
type PostgresInvitation struct {
    ID          string     `gorm:"primaryKey;column:id"`
    TenantID    string     `gorm:"column:tenant_id;not null;index:idx_invitations_tenant"`
    Email       string     `gorm:"column:email;not null"`
    DisplayName *string    `gorm:"column:display_name"`
    TokenHash   string     `gorm:"column:token_hash;not null;uniqueIndex"`
    Status      string     `gorm:"column:status;not null;default:pending"`
    InvitedBy   string     `gorm:"column:invited_by;not null"`
    ExpiresAt   time.Time  `gorm:"column:expires_at;not null"`
    AcceptedAt  *time.Time `gorm:"column:accepted_at"`
    CreatedAt   time.Time  `gorm:"column:created_at;not null"`
    UpdatedAt   time.Time  `gorm:"column:updated_at;not null"`
}
func (PostgresInvitation) TableName() string { return "invitations" }

// PostgresRoleAssignment is the GORM model for user_role_assignments (extended for 005)
type PostgresRoleAssignment struct {
    ID        int        `gorm:"primaryKey;column:id;autoIncrement"`
    UserID    string     `gorm:"column:user_id;not null"`
    ClientID  string     `gorm:"column:client_id;not null"`
    TenantID  string     `gorm:"column:tenant_id;not null"`
    Role      string     `gorm:"column:role;not null"`
    IsActive  bool       `gorm:"column:is_active;not null;default:true"`
    GrantedAt time.Time  `gorm:"column:granted_at;not null"`
    GrantedBy string     `gorm:"column:granted_by"`
    RevokedAt *time.Time `gorm:"column:revoked_at"`
    RevokedBy *string    `gorm:"column:revoked_by"`
}
func (PostgresRoleAssignment) TableName() string { return "user_role_assignments" }

// PostgresUserEvent is the GORM model for the user_events table
type PostgresUserEvent struct {
    ID         int64          `gorm:"primaryKey;column:id;autoIncrement"`
    TenantID   string         `gorm:"column:tenant_id;not null"`
    UserID     string         `gorm:"column:user_id;not null"`
    ClientID   *string        `gorm:"column:client_id"`
    EventType  string         `gorm:"column:event_type;not null"`
    IPAddress  *string        `gorm:"column:ip_address"`
    CountryCode *string       `gorm:"column:country_code"`
    Details    datatypes.JSON `gorm:"column:details;not null;default:'{}'"`
    OccurredAt time.Time      `gorm:"column:occurred_at;not null"`
}
func (PostgresUserEvent) TableName() string { return "user_events" }
```

---

## Migration Checklist

- [ ] Enable `pg_trgm` extension (if not already active from feature 004): `CREATE EXTENSION IF NOT EXISTS pg_trgm;`
- [ ] Apply `000009_extend_users.up.sql`
- [ ] Verify all existing users default to `status = 'active'`
- [ ] Apply `000010_create_invitations.up.sql`
- [ ] Apply `000011_extend_user_role_assignments.up.sql`
- [ ] Apply `000012_create_user_events.up.sql`
- [ ] Schedule retention cron job for `user_events` (delete rows older than 90 days)
- [ ] Schedule expiry job for `invitations` (`ExpireStalePending`, run every hour)
- [ ] Run `ANALYZE users, invitations, user_role_assignments, user_events;` after migrations
