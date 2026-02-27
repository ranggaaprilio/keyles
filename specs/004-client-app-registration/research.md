# Research: OAuth Client Application Registration Portal

**Date**: 2026-02-25  
**Feature**: [spec.md](./spec.md)  
**Plan**: [plan.md](./plan.md)

## Overview

This research addresses technical decisions for extending the existing Keyles SSO client management with public/confidential client types, quota enforcement, token revocation on client deletion, audit logging, pagination/search, and a React frontend dashboard.

---

## 1. Public vs Confidential Client Type Handling

### Context

OAuth2 (RFC 6749 §2.1) defines two client types:

- **Confidential**: Clients that can maintain credential confidentiality (server-side apps)
- **Public**: Clients that cannot (SPAs, mobile/native apps) — MUST use PKCE only, no client_secret

The existing `clients` table has `client_secret VARCHAR(255) NOT NULL`. We need to support nullable secrets for public clients.

### Decision: Add `client_type` column with enum

- **Approach**: Add a `client_type VARCHAR(20) NOT NULL DEFAULT 'confidential'` column to the `clients` table
- **Validation**: Public clients skip secret generation and skip `client_secret` NOT NULL constraint
- **Migration**: `ALTER TABLE clients ADD COLUMN client_type VARCHAR(20) NOT NULL DEFAULT 'confidential'` + relax `client_secret` to nullable
- **Domain entity**: Add `ClientType string` field to `entities.Client` with constants `ClientTypeConfidential = "confidential"` and `ClientTypePublic = "public"`
- **Use case change**: `CreateClientUseCase` checks `ClientType` — if public, skip secret generation; if confidential, generate and hash secret as before
- **Handler change**: Accept `client_type` in CreateClientRequest; omit `client_secret` from response for public clients

### Rationale

- Matches OAuth2 RFC 6749 §2.1 exactly
- Minimal schema change (one column addition + relax NOT NULL on secret)
- Backward compatible — all existing clients default to `confidential`

### Alternatives Considered

| Alternative                                  | Why Rejected                                                    |
| -------------------------------------------- | --------------------------------------------------------------- |
| Separate tables for public/confidential      | Over-engineering; 95% of fields are identical                   |
| Boolean `is_public` flag                     | Less self-documenting; string enum is clearer for API consumers |
| Store empty-string secret for public clients | Confusing; violates semantic meaning. NULL is correct           |

---

## 2. Tenant Client Quota Enforcement (Max 25)

### Context

FR-020 requires a maximum of 25 clients per tenant. This must be enforced at creation time and be resistant to race conditions.

### Decision: Database-level count + Redis cache

1. **Primary enforcement**: `CreateClientUseCase` calls `clientRepo.CountByTenant(ctx, tenantID)` before creating. If count >= 25, return a quota error.
2. **Race condition protection**: Use PostgreSQL advisory lock or `SELECT COUNT(*) ... FOR UPDATE` within a transaction to prevent two concurrent requests both passing the check.
3. **Optional Redis cache**: Cache tenant client counts in Redis (`client_count:{tenant_id}`) with short TTL (60s) for fast reads. Invalidate on create/delete. This avoids hitting PostgreSQL on every creation attempt.
4. **Constant**: Define `MaxClientsPerTenant = 25` in the domain layer as a business rule constant.

### Rationale

- Database is the source of truth; Redis cache is an optimization only
- Advisory locks or row-level locking prevent the race condition where two concurrent requests both see count=24 and both succeed
- 25 is a reasonable limit that can be raised per-tenant later via a `max_clients` column on the `tenants` table

### Alternatives Considered

| Alternative                               | Why Rejected                                                                         |
| ----------------------------------------- | ------------------------------------------------------------------------------------ |
| Redis-only counting                       | Not reliable as source of truth; can drift from actual DB count                      |
| PostgreSQL CHECK constraint with function | Non-standard and fragile; better enforced at application level                       |
| No cache, always query DB                 | Acceptable for v1 since write frequency is low; Redis cache is optional optimization |

---

## 3. Token Revocation on Client Deletion

### Context

When a client is deleted (FR-018), all active access tokens and refresh tokens issued through that client must be immediately revoked. The current `DeleteClientUseCase` only soft-deletes the client record.

### Decision: Cascade revocation through use case composition

1. **Extend `DeleteClientUseCase`**: After soft-deleting the client, call `refreshTokenRepo.RevokeByClientID(ctx, clientID)` to revoke all refresh tokens
2. **Access token handling**: Since access tokens are stateless JWTs (RS256, 15-min TTL), true immediate revocation requires either:
   - **(a)** Add `client_id` to a Redis blacklist (`revoked_client:{client_id}`, TTL = max access token lifetime of 15 min), checked during token validation
   - **(b)** Accept that access tokens expire naturally within 15 minutes (pragmatic approach)
3. **Recommended**: Option (a) — Redis client blacklist with 15-min TTL. This satisfies SC-019 ("100% effectiveness within 1 second") without changing the JWT validation flow significantly.

### Implementation Details

- New Redis key pattern: `revoked_client:{client_id}` with TTL of 900s (15 min)
- Token validation middleware checks this key before accepting access tokens
- `RefreshTokenRepository` needs `RevokeByClientID(ctx, clientID string) error` method
- Audit log entry with type `client_deleted` recorded

### Rationale

- Redis blacklist is lightweight (one key per deleted client, auto-expires)
- Satisfies the spec's "within 1 second" requirement
- Refresh token revocation prevents long-term session continuation
- Audit log provides compliance trail

### Alternatives Considered

| Alternative                               | Why Rejected                                                 |
| ----------------------------------------- | ------------------------------------------------------------ |
| Accept 15-min window for access tokens    | Doesn't meet SC-019 "within 1 second"                        |
| Store access tokens in DB (opaque tokens) | Major architecture change; defeats purpose of stateless JWTs |
| Short-lived access tokens (1 min)         | Too aggressive; increases token refresh traffic              |

---

## 4. Audit Logging for Client Lifecycle Events

### Context

FR-021 requires logging all client registration, modification, secret regeneration, and deletion events. The existing `audit_logs` table and `AuditLogRepository` interface already exist in the codebase.

### Decision: Use existing audit logging infrastructure

1. **Extend use cases**: Each client use case (Create, Update, Delete, RotateSecret) calls `auditLogRepo.Create(ctx, &entities.AuditLog{...})` at the end of successful operations
2. **Event types**: `client_created`, `client_updated`, `client_deleted`, `client_secret_rotated`
3. **Payload**: Include `client_id`, `tenant_id`, `actor_user_id`, `changes` (for updates: which fields changed)
4. **No new infrastructure**: The existing `AuditLogRepository` and PostgreSQL `audit_logs` table are sufficient

### Rationale

- Reuses existing infrastructure — no new tables or repositories needed
- Consistent with existing audit logging patterns in the codebase
- Event types are specific enough for security monitoring

---

## 5. Pagination and Search for Client List

### Context

FR-028 (search/filtering) and FR-030 (pagination) require that the client list endpoint supports paginated results with search capabilities. The current `ListByTenant()` returns all clients.

### Decision: Offset-based pagination with name search

1. **Pagination style**: Offset-based pagination (simpler, sufficient for ≤25 items per tenant). Parameters: `page` (default 1), `page_size` (default 10, max 25).
2. **Search**: Filter by `client_name` using case-insensitive `ILIKE` pattern matching. Parameter: `search` (optional).
3. **New repository method**: `ListByTenantPaginated(ctx, tenantID, search string, page, pageSize int) ([]*entities.Client, total int, error)`
4. **Response**: Include `total`, `page`, `page_size`, `total_pages` in list response metadata.

### Rationale

- Offset-based is simpler and appropriate for small datasets (≤25 items)
- ILIKE search is PostgreSQL-native and performs well on small datasets
- No need for full-text search or Elasticsearch for this scale

### Alternatives Considered

| Alternative                 | Why Rejected                                         |
| --------------------------- | ---------------------------------------------------- |
| Cursor-based pagination     | Over-engineering for ≤25 items                       |
| Full-text search (tsvector) | Unnecessary complexity for searching client names    |
| Client-side filtering only  | Doesn't meet FR-028's server-side search requirement |

---

## 6. Frontend Client Management Dashboard Architecture

### Context

The frontend needs a new client management page with list view, detail view, create/edit forms, secret display, and documentation. Must integrate with existing React + TanStack Query + Zustand patterns.

### Decision: Component-based architecture with TanStack Query

1. **Page**: `ClientManagementPage.tsx` — main page with list + create button
2. **Components**:
   - `ClientList.tsx` — paginated list with search bar, renders `ClientCard` items
   - `ClientCard.tsx` — summary card showing name, type, client_id, status
   - `ClientDetail.tsx` — full detail view with redirect URIs, timestamps, action buttons
   - `CreateClientForm.tsx` — form using react-hook-form + zod, type selector (confidential/public), redirect URI management
   - `EditClientForm.tsx` — partial update form
   - `SecretDisplay.tsx` — one-time secret display with copy-to-clipboard, confirmation checkbox
   - `RotateSecretDialog.tsx` — confirmation dialog with warning text
   - `DeleteClientDialog.tsx` — confirmation dialog with warning text
   - `IntegrationDocs.tsx` — inline OAuth documentation with personalized code examples
3. **Hooks**: `useClients.ts` — TanStack Query hooks wrapping `clientService` (create, list, get, update, delete, rotateSecret)
4. **State**: No Zustand store needed — TanStack Query's cache handles server state; form state is local
5. **Routing**: Add `/admin/clients` route to `App.tsx` with `ProtectedRoute` wrapper

### UI/UX Decisions

- **Secret display**: Modal overlay after creation, copy button with visual feedback, checkbox "I have saved the secret" before dismissing
- **Type selector**: Radio buttons (Confidential / Public) with helper text explaining the difference
- **Redirect URIs**: Dynamic list editor with add/remove buttons, inline validation
- **Documentation**: Expandable section with tabbed code examples (cURL, JavaScript, Go) using the client's actual client_id

### Rationale

- Follows existing patterns (TanStack Query mutations, react-hook-form + zod validation)
- Component decomposition enables independent testing
- No global state needed — server state caching via TanStack Query is sufficient

---

## 7. Secret Generation and Hashing

### Context

Existing implementation generates 32-byte cryptographically random secrets (256 bits entropy) using `crypto/rand`, base64url encoded. Secrets are hashed with bcrypt (cost 12) via `PasswordService.Hash()`.

### Decision: Keep existing approach, no changes needed

- **client_id**: 24 bytes → 32 chars base64url (192 bits) — already sufficient
- **client_secret**: 32 bytes → 43 chars base64url (256 bits) — meets SC-004
- **Hashing**: bcrypt cost 12 via existing `PasswordService` — meets FR-008
- **Public clients**: Skip secret generation entirely; `ClientSecretHash` is NULL

### Rationale

- Existing implementation already meets all security requirements
- 256-bit entropy exceeds OWASP recommendations for OAuth client secrets
- bcrypt cost 12 provides adequate protection against brute force

---

## 8. Redirect URI Validation Enhancements

### Context

The existing `ValidateRedirectURI()` checks for scheme, host, and no fragments. FR-003 requires HTTPS enforcement in production. FR-004 allows localhost HTTP for development.

### Decision: Extend existing validation with environment awareness

1. **Production rule**: Reject non-HTTPS URIs unless host is `localhost` or `127.0.0.1`
2. **Development rule**: Allow HTTP for `localhost` and `127.0.0.1` hosts (any port)
3. **Implementation**: Add `ValidateRedirectURIStrict(uri string, allowInsecureLocalhost bool)` to domain entity
4. **Configuration**: Pass `allowInsecureLocalhost` flag from environment config (`APP_ENV != "production"`)

### Rationale

- Localhost exception is standard practice (Auth0, Google OAuth, GitHub all allow it)
- Environment-aware validation prevents accidental HTTP URIs in production
- Backward compatible with existing URIs

---

## Summary of Research Decisions

| #   | Topic                   | Decision                                                                    |
| --- | ----------------------- | --------------------------------------------------------------------------- |
| 1   | Client types            | `client_type` VARCHAR column, nullable `client_secret`                      |
| 2   | Quota enforcement       | DB count + Redis cache, PostgreSQL locking for race conditions              |
| 3   | Token revocation        | Redis blacklist (`revoked_client:{id}`, 15-min TTL) + refresh token cascade |
| 4   | Audit logging           | Use existing `AuditLogRepository`, add event types per use case             |
| 5   | Pagination/search       | Offset-based, `ILIKE` name search, metadata in response                     |
| 6   | Frontend architecture   | TanStack Query hooks, component decomposition, no Zustand                   |
| 7   | Secret generation       | Keep existing 256-bit crypto/rand + bcrypt cost 12                          |
| 8   | Redirect URI validation | Add HTTPS enforcement with localhost exception                              |
