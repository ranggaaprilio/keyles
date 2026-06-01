# Feature Specification: OAuth Consent Flow and End-User Authentication

**Feature Branch**: `006-oauth-consent-flow`  
**Created**: June 1, 2026  
**Status**: Draft  
**Input**: `future-epics/01-oauth-consent-flow.md`

## Summary

Complete the browser-facing OAuth 2.0 / OpenID Connect Authorization Code flow so an
external client can redirect an end-user to Keyles, authenticate the user with a
separate SSO session cookie, collect consent, receive an authorization code, and
exchange that code through the existing token endpoint.

The implementation must remove the placeholder `X-User-ID` authorization behavior.
The backend remains the authority for authorization-request validation, session
validation, consent decisions, and redirect construction. The frontend renders the
login, consent, and error experiences without becoming a source of trusted OAuth
state.

## Clarifications

### Session 2026-06-01

- Q: How should the new OAuth end-user login endpoint limit failed authentication attempts? → A: Allow 5 failed attempts per 15 minutes for both source IP address and tenant-scoped email.
- Q: How should an end-user terminate their Keyles browser SSO session? → A: Add `POST /oauth2/logout` to delete the current Redis session and expire its cookie.
- Q: Which browser-flow security events should Keyles audit? → A: Audit login success, login failure, throttling, consent approval, consent denial, logout, and invalid callback attempts.
- Q: How should the browser OAuth flow behave when Redis is unavailable? → A: Fail closed with a local temporary-unavailable error for authorization, login, and consent; logout still expires the browser cookie.
- Q: What domain scope should the end-user SSO cookie use? → A: Use a host-only cookie with `Path=/` and no `Domain` attribute.

## User Scenarios and Testing

### User Story 1 - Authenticate an End-User During OAuth (Priority: P1)

An unauthenticated end-user starts sign-in from an external client. Keyles validates
the authorization request, sends the browser to the SSO login page, validates the
user's tenant-scoped credentials, creates an end-user session, and continues to the
consent screen.

**Independent Test**: Start a valid `/oauth2/auth` request without an SSO session,
log in with an active tenant user who has a role for the client, and verify that the
browser reaches the consent page with an HttpOnly session cookie.

**Acceptance Scenarios**:

1. **Given** a valid authorization request and no end-user session, **When** the
   browser requests `/oauth2/auth`, **Then** Keyles redirects to
   `/oauth2/login?transaction_id=...` on the configured frontend.
2. **Given** a valid authorization transaction, **When** an active end-user submits
   valid email and password credentials, **Then** Keyles creates a server-side
   session and redirects to `/oauth2/consent?transaction_id=...`.
3. **Given** an invalid password, disabled user, tenant mismatch, or user without an
   active role for the requesting client, **When** login is submitted, **Then**
   Keyles rejects the login without creating a session.
4. **Given** an expired or unknown transaction identifier, **When** the login page
   loads or submits, **Then** Keyles renders a user-friendly OAuth error state.
5. **Given** five failed OAuth login attempts within 15 minutes for either a source
   IP address or tenant-scoped email, **When** another login is submitted for that
   limited key, **Then** Keyles rejects the attempt without verifying credentials.

### User Story 2 - Approve or Deny Consent (Priority: P1)

An authenticated end-user sees the requesting client and scopes, then either allows
or denies access.

**Independent Test**: Continue from a valid authenticated transaction, load consent
details, approve once, and verify that the client receives a single-use code and the
original `state`. Repeat with deny and verify `error=access_denied`.

**Acceptance Scenarios**:

1. **Given** a valid authenticated authorization transaction, **When** the consent
   page loads, **Then** it displays the client name, optional logo, signed-in user,
   and requested scope descriptions.
2. **Given** the user approves, **When** consent is submitted, **Then** Keyles
   generates a five-minute single-use authorization code and redirects to the
   registered callback URI with `code` and the original `state`.
3. **Given** the user denies, **When** consent is submitted, **Then** Keyles
   redirects to the registered callback URI with `error=access_denied`,
   `error_description`, and the original `state`.
4. **Given** an invalid or replayed transaction identifier, **When** consent is
   submitted, **Then** Keyles rejects the request and does not issue a code.

### User Story 3 - Reuse and Refresh End-User Sessions (Priority: P1)

An end-user with a valid SSO session can continue a later authorization request
without signing in again unless the client requests reauthentication.

**Independent Test**: Complete one login, begin a second valid authorization request
for another registered client in the same tenant, and verify direct consent routing.
Repeat after removing eligibility, with `prompt=login`, and with an exceeded
`max_age` value to verify forced reauthentication.

**Acceptance Scenarios**:

1. **Given** a valid SSO session, **When** a normal authorization request arrives,
   **Then** Keyles skips login and redirects to consent.
2. **Given** `prompt=login`, **When** a valid SSO session exists, **Then** Keyles
   requires active reauthentication before consent.
3. **Given** `max_age` is exceeded, **When** a valid SSO session exists, **Then**
   Keyles requires active reauthentication before consent.
4. **Given** `prompt=consent`, **When** an eligible SSO session exists, **Then**
   Keyles displays consent before issuing an authorization code.
5. **Given** `prompt=none` and user interaction would be required, **When** the
   authorization request is processed, **Then** Keyles returns the applicable OIDC
   error to the validated client callback without displaying login or consent UI.
6. **Given** a valid client-agnostic SSO session, **When** another registered client
   in the same tenant starts authorization and the user remains active with a role
   for that client, **Then** Keyles reuses the session and routes directly to consent.
7. **Given** an SSO session whose user is disabled or no longer has a role for the
   requesting client, **When** authorization starts, **Then** Keyles does not reuse
   the session for silent continuation.

### User Story 4 - Handle OAuth Errors Safely (Priority: P2)

An end-user sees a clear error when an OAuth request cannot proceed safely.

**Independent Test**: Request authorization with an invalid client, mismatched
redirect URI, expired transaction, and denied consent, then verify safe local or
client error handling as appropriate.

**Acceptance Scenarios**:

1. **Given** an invalid `client_id` or unregistered `redirect_uri`, **When**
   authorization validation fails, **Then** Keyles shows a local error page and
   never redirects to the untrusted URI.
2. **Given** a validated callback URI and a later authorization error, **When** the
   flow fails, **Then** Keyles returns the error and original `state` to that
   callback URI.
3. **Given** a common OAuth error, **When** the frontend renders it, **Then** the
   user sees a concise explanation without sensitive implementation details.

### User Story 5 - Complete the Existing PKCE Token Flow (Priority: P1)

An external client exchanges the approved authorization code using its PKCE verifier
and receives tokens through the existing token endpoint.

**Independent Test**: Complete browser login and consent, exchange the returned code
with its matching verifier, and verify access, ID, and refresh tokens are returned.

**Acceptance Scenarios**:

1. **Given** an approved authorization code, **When** the client exchanges it with
   the matching S256 verifier, **Then** the existing `/oauth2/token` endpoint returns
   tokens.
2. **Given** a missing or mismatched verifier, **When** token exchange is attempted,
   **Then** Keyles returns `invalid_grant`.
3. **Given** an already-used or expired code, **When** token exchange is attempted,
   **Then** Keyles returns `invalid_grant`.

### User Story 6 - End the Keyles Browser SSO Session (Priority: P2)

An end-user can explicitly sign out of the Keyles browser session without changing
the external client's token lifecycle.

**Independent Test**: Complete OAuth login, submit `POST /oauth2/logout`, then start
a new authorization request and verify that Keyles requires end-user login again.

**Acceptance Scenarios**:

1. **Given** an active end-user SSO session, **When** the browser submits
   `POST /oauth2/logout`, **Then** Keyles deletes the Redis session and returns an
   expired session cookie.
2. **Given** no active end-user SSO session, **When** the browser submits
   `POST /oauth2/logout`, **Then** Keyles still returns success and an expired
   session cookie.
3. **Given** a successful provider-local logout, **When** an external client starts
   authorization again, **Then** Keyles requires login.

## Requirements

### Functional Requirements

- **FR-001**: The backend MUST validate the initial `/oauth2/auth` request before
  redirecting to any frontend route.
- **FR-002**: The backend MUST require `client_id`, registered `redirect_uri`,
  `response_type=code`, `scope` containing `openid`, non-empty `state`,
  `code_challenge`, and `code_challenge_method=S256`.
- **FR-003**: The backend MUST create a short-lived, opaque authorization
  transaction after initial validation and MUST store trusted OAuth parameters
  server-side.
- **FR-004**: Frontend OAuth routes MUST receive only an opaque transaction
  identifier and MUST NOT be trusted to resubmit mutable OAuth request parameters.
- **FR-005**: End-user login MUST use email and password against the tenant resolved
  from the registered OAuth client.
- **FR-006**: End-user login MUST reject disabled or pending users and users without
  an active role for the requesting client.
- **FR-007**: Successful end-user login MUST create a Redis-backed session separate
  from admin JWT authentication. The session MUST be client-agnostic so an eligible
  user can reuse one Keyles SSO session across registered clients in the same tenant.
- **FR-008**: The end-user session MUST be represented in the browser by an
  `HttpOnly`, `SameSite=Lax`, host-only cookie with `Path=/` and no `Domain`
  attribute. The cookie MUST be `Secure` outside local development.
- **FR-009**: Session and authorization transaction TTL values MUST be configurable;
  defaults are eight hours and ten minutes respectively.
- **FR-010**: Authenticated users MUST be routed to a consent page displaying client
  metadata, the signed-in user, and requested scopes.
- **FR-011**: Consent approval MUST generate the existing five-minute,
  single-use authorization code with the original PKCE fields.
- **FR-012**: Consent denial MUST redirect to the validated callback URI with
  `error=access_denied`, an error description, and the original `state`.
- **FR-013**: Consent submissions MUST validate the end-user session, transaction
  ownership, transaction expiry, one-time use, active end-user status, and current
  client-role assignment before returning a response.
- **FR-014**: The backend MUST preserve the original `state` unchanged in success
  and callback-safe error responses.
- **FR-015**: Invalid clients and invalid callback URIs MUST result in a local error
  page without redirecting to an untrusted URI.
- **FR-016**: `prompt=login` and exceeded `max_age` MUST force active
  reauthentication.
- **FR-017**: `prompt=consent` MUST force the consent UI.
- **FR-018**: `prompt=none` MUST NOT display login or consent UI and MUST return an
  OIDC interaction error when silent completion is not possible.
- **FR-019**: `max_age` MUST be parsed as a non-negative integer and active
  reauthentication caused by `max_age` MUST be represented by the session
  authentication timestamp for downstream token claims where supported.
- **FR-020**: The backend MUST add `FRONTEND_URL`, cookie security, session TTL, and
  authorization-transaction TTL configuration and document them in environment
  examples and Docker Compose.
- **FR-021**: The existing `/oauth2/token`, refresh, revocation, introspection,
  discovery, and JWKS behavior MUST remain compatible.
- **FR-022**: The OAuth end-user login endpoint MUST reject authentication attempts
  after five failures within 15 minutes for either the source IP address or the
  tenant-scoped normalized email. A successful login MUST clear the email failure
  bucket. Each failure bucket MUST use a fixed 15-minute window beginning with its
  first failure; subsequent failures MUST NOT extend the window. Rate-limit responses
  MUST not reveal which key was limited.
- **FR-023**: The backend MUST expose `POST /oauth2/logout` to delete the current
  Redis-backed end-user session and expire the browser session cookie. The endpoint
  MUST be idempotent and MUST NOT accept a client-controlled post-logout redirect
  URI in this feature.

### Non-Functional Requirements

- **NFR-001**: Domain and use-case code MUST depend on repository and service
  interfaces; cookie parsing and HTTP redirects remain in the HTTP interface layer.
- **NFR-002**: Authorization transactions and end-user sessions MUST use
  cryptographically random opaque identifiers.
- **NFR-003**: Login failures MUST not reveal whether the email, password, role, or
  account status caused rejection.
- **NFR-004**: Browser credential requests MUST use CORS credentials where frontend
  and backend origins differ in development.
- **NFR-005**: Backend business logic and handlers MUST have unit and integration
  coverage; frontend login, consent, routing, and error behavior MUST have Vitest
  coverage.
- **NFR-006**: OAuth end-user login throttling MUST use Redis-backed counters so
  limits apply consistently across backend instances.
- **NFR-007**: The backend MUST emit structured security audit events for OAuth
  end-user login success, login failure, throttling, consent approval, consent
  denial, provider-local logout, and invalid callback attempts. Events MUST record
  an outcome code and relevant tenant, client, user, transaction, and source IP
  identifiers when known. Events MUST NOT contain passwords, session-cookie values,
  authorization codes, PKCE values, or other secrets.
- **NFR-008**: When Redis is unavailable, authorization initialization, OAuth
  end-user login, and consent read or submission MUST fail closed with a local
  `temporarily_unavailable` error page or response and MUST NOT issue authorization
  codes or redirect OAuth errors to an external callback. Provider-local logout
  MUST still expire the browser cookie even when Redis session deletion cannot
  complete.
- **NFR-009**: OAuth login throttling and audit events MUST derive the source IP from
  the direct TCP peer address and MUST NOT trust forwarded IP headers in this
  feature. A future reverse-proxy deployment MUST define an explicit trusted-proxy
  allowlist before forwarded client-IP parsing is enabled.

## Key Entities

- **Authorization Transaction**: Short-lived server-side record containing validated
  OAuth request fields, tenant and client context, session binding, current stage,
  and one-time completion state.
- **End-User Session**: Client-agnostic Redis-backed browser session with user,
  tenant, authentication timestamp, creation time, and expiration.
- **Authorization Code**: Existing single-use five-minute code containing client,
  user, tenant, callback URI, scopes, and PKCE challenge.
- **Client**: Existing registered OAuth client with tenant ownership, allowed
  callback URIs, status, and display metadata.
- **End User**: Existing tenant user with email, password hash, status, and
  client-role assignments.

## Edge Cases

- The authorization transaction expires while the login or consent screen is open.
- A user logs in successfully in one tab while another tab reuses an old
  transaction.
- A session cookie points to a deleted Redis session.
- A user is disabled or loses their client role between login and consent approval.
- A valid session is reused for another client in the same tenant after the user's
  account is disabled or the required client role is removed.
- `prompt=none` is combined with another prompt value.
- `max_age=0` forces active reauthentication.
- A callback URI already contains query parameters.
- A user denies consent after authenticating.
- A replayed consent submission attempts to generate a second authorization code.
- Docker Compose runs frontend and backend on different localhost ports.
- Either the source IP address or tenant-scoped email reaches the OAuth login
  failure threshold while the other key remains below its threshold.
- A client sends spoofed forwarded-IP headers directly to the backend.
- The browser submits logout after the Redis SSO session has already expired.
- Redis becomes unavailable during authorization initialization, login, consent,
  or provider-local logout.

## Assumptions

- The first implementation always displays consent before issuing a code unless
  `prompt=none` can complete through a future persisted consent grant. Persisted
  grants are outside this feature; therefore `prompt=none` returns
  `consent_required` after successful silent session validation.
- `select_account` is rejected as unsupported because Keyles stores one end-user
  identity per browser session in this MVP.
- Existing client metadata may not yet include a logo URI; the consent UI falls back
  to its current placeholder.
- No PostgreSQL migration is required. New authorization transactions and sessions
  are ephemeral Redis records.

## Success Criteria

- **SC-001**: A Docker Compose test client can complete login, consent approval, and
  token exchange end-to-end using S256 PKCE.
- **SC-002**: Denial returns `access_denied` with the original `state`.
- **SC-003**: Header-only `X-User-ID` authentication no longer authorizes browser
  requests.
- **SC-004**: Reauthentication occurs for `prompt=login`, `max_age=0`, and expired
  session age.
- **SC-005**: Invalid callback URIs never receive redirects.
- **SC-006**: Backend tests, frontend tests, lint, and builds pass.
- **SC-007**: Cross-client session reuse occurs only for active users with a current
  role for the requesting client.
- **SC-008**: Spoofed forwarded-IP headers do not change the source-IP throttle key
  or audit identifier.
