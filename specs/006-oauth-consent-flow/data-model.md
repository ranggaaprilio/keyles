# Data Model: OAuth Consent Flow and End-User Authentication

**Feature**: [spec.md](spec.md)  
**Date**: June 1, 2026

## Overview

This feature adds no PostgreSQL tables. It introduces one ephemeral Redis entity
and extends the existing Redis-backed end-user session shape. Existing clients,
end-users, roles, and authorization codes remain the durable or established model.

## Authorization Transaction

**Storage**: Redis  
**Key pattern**: `oauth:transaction:{transaction_id}`  
**Default TTL**: 10 minutes, configurable

| Field | Type | Rules |
| --- | --- | --- |
| `TransactionID` | string | Cryptographically random opaque identifier; required |
| `ClientID` | string | Required; resolved and validated before storage |
| `TenantID` | string | Required; derived from registered client |
| `RedirectURI` | string | Required; exact registered URI |
| `ResponseType` | string | Required; `code` only |
| `Scope` | string | Required; includes `openid` |
| `State` | string | Required; preserved unchanged for client callback |
| `CodeChallenge` | string | Required; valid S256 base64url challenge |
| `CodeChallengeMethod` | string | Required; `S256` only |
| `Nonce` | string | Optional OIDC nonce |
| `Prompt` | []string | Optional validated prompt values |
| `MaxAgeSeconds` | *int | Optional; non-negative |
| `UserID` | string | Set after login or eligible session reuse |
| `SessionID` | string | Set when transaction is bound to a session |
| `InteractionCSRFToken` | string | Cryptographically random opaque provider-form CSRF token |
| `Stage` | string | `pending_login`, `pending_consent`, or `completed` |
| `CreatedAt` | time | Required |
| `ExpiresAt` | time | Required |
| `CompletedAt` | *time | Set once approved or denied |

### Validation Rules

- Initial storage occurs only after registered client, callback URI, response type,
  scope, PKCE, `state`, `prompt`, and `max_age` validation.
- A consent read requires `pending_consent`, a valid session cookie, and matching
  session and transaction user identifiers.
- A consent write also requires a matching `InteractionCSRFToken`.
- Consent approval revalidates active end-user status and the current client-role
  assignment.
- Approval and denial atomically consume the transaction so replay cannot issue a
  second code or callback.

### State Transitions

```text
validated request
  -> pending_login
  -> pending_consent
  -> completed

validated request with eligible session
  -> pending_consent
  -> completed

prompt=none
  -> callback error without UI
```

## End-User Session

**Storage**: Redis  
**Key pattern**: `oauth:session:{session_id}`  
**Default TTL**: 8 hours, configurable

| Field | Type | Rules |
| --- | --- | --- |
| `SessionID` | string | Cryptographically random opaque identifier; required |
| `UserID` | string | Required |
| `TenantID` | string | Required |
| `AuthenticatedAt` | time | Required; reset only by active login |
| `CreatedAt` | time | Required |
| `ExpiresAt` | time | Required |
| `Metadata` | map | Optional audit context such as user agent |

### Validation Rules

- The session is distinct from tenant-admin JWT authentication.
- The SSO session is client-agnostic. Remove the legacy `ClientID` field from this
  Redis shape rather than using it as an authorization decision input.
- Session reuse across clients is allowed, but authorization still validates that
  the end-user remains active, belongs to the client's tenant, and has an active role
  for that client before binding a new authorization transaction.
- `prompt=login` and exceeded `max_age` bypass session reuse and require login.
- Missing, expired, or deleted sessions are treated as unauthenticated.

## OAuth Login Failure Counter

**Storage**: Redis  
**Key patterns**:

```text
oauth:login-failure:ip:{source_ip}
oauth:login-failure:email:{tenant_id}:{normalized_email_hash}
```

**TTL**: Fixed 15-minute window beginning with the first failure

| Field | Type | Rules |
| --- | --- | --- |
| `Key` | string | Source-IP or tenant-email key |
| `FailureCount` | integer | Incremented atomically on failed authentication |
| `ExpiresAt` | time | Set by Redis TTL |

### Validation Rules

- Check both keys before password verification.
- Reject authentication when either counter has reached five failures.
- Increment both keys atomically on generic authentication failure. Create the
  15-minute TTL only when a key is first created; later increments do not extend it.
- Clear the tenant-email counter after successful login.
- Do not expose which counter caused a throttle response.
- Derive `{source_ip}` from the direct TCP peer address. Do not trust forwarded-IP
  headers without a separately designed trusted-proxy allowlist.

## Security Audit Event

**Storage**: Existing audit repository  

Browser-flow events reuse the existing audit abstraction. Each event records an
event type and outcome code plus tenant, client, user, transaction, and source-IP
identifiers when known. Payloads must exclude passwords, cookies, authorization
codes, PKCE values, and other secrets. Extend the existing domain event-type
constants for these events. If audit persistence fails, emit the same sanitized
event through structured application logging.

| Event type | Trigger |
| --- | --- |
| `oauth_login_succeeded` | End-user session created |
| `oauth_login_failed` | Generic authentication failure |
| `oauth_login_throttled` | Either Redis throttle key reached its limit |
| `oauth_consent_approved` | Authorization code issued |
| `oauth_consent_denied` | Client callback receives `access_denied` |
| `oauth_logout` | Provider-local logout requested |
| `oauth_invalid_callback` | Callback URI rejected |

## Authorization Code

**Storage**: Existing Redis repository  
**Key pattern**: `oauth:authcode:{code}`  
**TTL**: 5 minutes

The existing `entities.AuthorizationCode` remains the output of approved consent.
Its fields continue to include client, tenant, user, callback URI, scope, PKCE
challenge, expiry, and one-time use state. Add optional `Nonce` and
`AuthenticatedAt` fields so approved browser interactions preserve OIDC replay and
authentication-age context through token issuance. Existing atomic token-endpoint
consumption remains unchanged.

## Existing Related Entities

### Client

The existing PostgreSQL client supplies tenant context, active state, allowed
callback URIs, client name, and optional display metadata when available.

### End User

The existing PostgreSQL end-user supplies tenant identity, email, display name,
password hash, and account status. OAuth login accepts active users only.

### User Role Assignment

The existing PostgreSQL role assignment remains the authorization gate for a
user-client pair. Role access is checked during login continuation and again before
authorization-code issuance.
