# Implementation Plan: OAuth Consent Flow and End-User Authentication

**Branch**: `006-oauth-consent-flow` | **Date**: June 1, 2026 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `/specs/006-oauth-consent-flow/spec.md`

## Summary

Complete the browser-facing OAuth 2.0 / OIDC Authorization Code flow by replacing
the placeholder `X-User-ID` behavior with a Redis-backed end-user SSO session, an
opaque Redis-backed authorization transaction, frontend login and consent routes,
provider-local logout, and safe OAuth error handling. Preserve the existing token,
refresh, revocation, introspection, discovery, and JWKS behavior.

The backend validates and stores trusted OAuth request values before the browser
reaches the frontend. The frontend receives only a short-lived `transaction_id`,
renders the interaction, and sends login or consent decisions to backend endpoints
that revalidate session, role, CSRF, expiry, and one-time use. Redis-backed
throttling limits login failures by both source IP and tenant-scoped normalized
email. Browser-flow security outcomes are audited without recording secrets.

## Technical Context

**Language/Version**: Go 1.23.0; TypeScript 5.4; React 18.3  
**Primary Dependencies**: Gin 1.10, Redis go-redis/v9, GORM, bcrypt, React Router 6, Axios, React Hook Form, Zod, Tailwind CSS, Vitest, React Testing Library  
**Storage**: Existing PostgreSQL for clients/users/roles/audit logs; Redis for authorization transactions, end-user sessions, login-throttle counters, and authorization codes  
**Testing**: Go `testing` + Testify; Vitest + React Testing Library; Docker Compose manual flow verification  
**Target Platform**: Linux backend service and browser frontend served through Docker Compose  
**Project Type**: Multi-project web application with Go API and React frontend  
**Performance Goals**: OAuth interaction reads and writes remain O(1) Redis operations; login adds bounded Redis throttle checks, one tenant-scoped user lookup, and one role check  
**Constraints**: S256 PKCE required; invalid callback URIs must never receive redirects; no admin JWT reuse; no trusted OAuth parameters from frontend; host-only HttpOnly `SameSite=Lax` cookie with `Path=/`; direct TCP peer source IP only with no forwarded-header trust; fail closed when Redis is unavailable except local cookie expiry during logout  
**Scale/Scope**: Existing multi-tenant platform limits; one new Redis transaction entity, one client-agnostic extended Redis session shape, two fixed-window Redis throttle counters per failed login, backend OAuth handler/use-case refactor, four public frontend routes, no PostgreSQL migration

## Constitution Check

*GATE: Passed before Phase 0 research. Re-checked after Phase 1 design.*

### Clean Architecture Compliance

- [x] Domain layer has no imports from infrastructure or HTTP packages.
- [x] Authorization transactions and login throttling are defined through domain
  repository or service interfaces.
- [x] Security audit recording reuses the domain audit abstraction.
- [x] Redis implementations remain in `backend/infrastructure/persistence/redis/`.
- [x] Cookie parsing, setting, expiry, and HTTP redirects remain in
  `backend/interfaces/http/handlers/`.
- [x] Application rules for transaction progression, login, consent, logout,
  prompt handling, session age, throttling, and fail-closed errors remain in
  `backend/usecase/auth/`.

### SOLID Compliance

- [x] Authorization initialization, end-user login, consent-detail read, consent
  submission, and provider-local logout have separate responsibilities.
- [x] Use cases depend on repository, password-service, throttle, and audit
  abstractions.
- [x] The existing duplicate authorization-code creation paths are consolidated
  into one final approval path.
- [x] Frontend pages depend on a dedicated OAuth interaction service instead of
  embedding network calls in presentational components.

### Testing Requirements

- [x] Unit-test plan covers transaction validation and transitions, prompt parsing,
  max-age behavior, credential rejection, dual-key throttling, role checks,
  approval/denial, logout, audit payload sanitization, and Redis outages.
- [x] Handler integration-test plan covers redirects, host-only cookies,
  interaction APIs, invalid callback safety, one-time consent, logout, throttle
  responses, fail-closed errors, and existing token exchange.
- [x] Frontend-test plan covers login, consent, logout, error pages, and route
  wiring.
- [x] Test isolation uses mocked domain interfaces for use cases and disposable
  Redis behavior for handler integration tests.
- [x] Domain/business-logic target remains at least 85% coverage.

### Post-Design Re-Check

- [x] No constitution violations introduced by the Phase 1 design.
- [x] All outbound persistence, audit, throttling, and password-verification
  dependencies are represented as interfaces.
- [x] No PostgreSQL migration or cross-layer dependency is required.

## Project Structure

### Documentation (this feature)

```text
specs/006-oauth-consent-flow/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── README.md
│   └── openapi.yaml
└── tasks.md              # Generated by /speckit.tasks, not this command
```

### Source Code (repository root)

```text
backend/
├── cmd/server/main.go
├── domain/
│   ├── entities/
│   │   ├── authorization_code.go                     # Carry auth time and nonce
│   │   └── audit_log.go                              # Extend OAuth event types
│   ├── repositories/
│   │   ├── authorization_transaction_repository.go  # New
│   │   └── session_repository.go                    # Remove client binding; add auth time
│   └── services/
│       └── login_throttler.go                        # New abstraction
├── infrastructure/
│   ├── config/
│   │   ├── config.go                                # Frontend/cookie/TTL config
│   │   └── config_test.go
│   └── persistence/redis/
│       ├── authorization_transaction_repository.go  # New
│       ├── authorization_transaction_repository_test.go
│       ├── login_throttler.go                        # New dual-key counters
│       ├── login_throttler_test.go
│       └── session_repository.go                    # Reuse and extend
├── interfaces/http/
│   ├── handlers/
│   │   └── oauth_handler.go                         # Redirect, cookie, APIs
│   └── router.go                                    # Add interaction routes
├── usecase/auth/
│   ├── authorize_client.go                          # Split initial validation
│   ├── authenticate_end_user.go                     # New
│   ├── consent_decision.go                          # Consolidate final issuance
│   ├── get_consent_details.go                       # New
│   ├── logout_end_user.go                           # New
│   └── oauth_interaction.go                         # New transaction/prompt rules
└── tests/
    ├── integration/oauth_auth_test.go
    ├── integration/oauth_consent_test.go             # New
    ├── integration/oauth_logout_test.go              # New
    ├── unit/usecase/authenticate_end_user_test.go     # New
    ├── unit/usecase/consent_decision_test.go          # New
    ├── unit/usecase/logout_end_user_test.go           # New
    └── unit/usecase/oauth_interaction_test.go         # New

frontend/
├── src/
│   ├── App.tsx                                      # Add public OAuth routes
│   ├── components/auth/
│   │   ├── ConsentScreen.tsx                        # Reuse presentational UI
│   │   └── OAuthErrorPanel.tsx                      # New shared errors
│   ├── pages/
│   │   ├── OAuthLoginPage.tsx                       # New
│   │   ├── OAuthConsentPage.tsx                     # New
│   │   ├── OAuthLogoutPage.tsx                      # New
│   │   └── OAuthErrorPage.tsx                       # New
│   ├── services/
│   │   └── oauthInteractionService.ts               # New credentialed API client
│   └── types/oauth.ts                               # Add interaction contracts
└── src/tests/
    ├── OAuthLoginPage.test.tsx                       # New
    ├── OAuthConsentPage.test.tsx                     # New
    ├── OAuthLogoutPage.test.tsx                      # New
    └── OAuthErrorPage.test.tsx                       # New

docker-compose.yml                                   # Add browser-flow env
backend/.env.example                                 # Add browser-flow env
```

**Structure Decision**: Keep the existing backend Clean Architecture boundaries and
React component/page/service organization. Use Redis for ephemeral browser-flow
state and distributed login counters because Redis already owns OAuth codes and
sessions. Reuse the existing audit repository abstraction. No database migration or
new package family is needed.

## Implementation Design

### 1. Initialize Authorization Interaction

`GET /oauth2/auth` validates the external request before creating an authorization
transaction. It parses optional `nonce`, `prompt`, and `max_age`, validates the
registered callback exactly, resolves the tenant from the client, and stores a
random interaction CSRF token.

The handler reads the end-user session cookie:

- Missing, expired, forced-login, too-old, disabled-user, tenant-mismatched, or
  role-ineligible session: redirect to frontend login.
- Eligible client-agnostic session: revalidate the active user and current role,
  bind the transaction to that session, and redirect to consent.
- `prompt=none` requiring either login or visible consent: redirect the applicable
  OIDC error to the already-validated callback without rendering UI.
- Redis failure: render a local `temporarily_unavailable` error without an external
  callback redirect.

### 2. Authenticate End User

`POST /oauth2/login` accepts `transaction_id`, email, and password. Before password
verification, the use case checks Redis-backed fixed 15-minute failure buckets for
source IP and tenant-scoped normalized email. The first failure creates a bucket TTL;
later failures increment atomically without extending it. Either bucket reaching
five failures rejects the attempt generically. Failure increments both buckets;
success clears the email bucket.

The use case loads the transaction and client tenant, fetches the tenant-scoped
end-user, checks active status and client role, verifies bcrypt password through
`PasswordService`, creates an end-user session, binds the transaction, records last
login, and returns the frontend consent URL. The handler sets a host-only HttpOnly
`SameSite=Lax` cookie with `Path=/` and no `Domain` attribute. Login failures return
one generic user-facing credential error.

### 3. Render and Submit Consent

`GET /oauth2/consent/:transactionId` reads consent details for the currently bound
session. The response includes display-safe client details, signed-in user display
text, scopes, and the interaction CSRF token.

`POST /oauth2/consent` accepts `transaction_id`, interaction CSRF token, and
`approved`. It atomically completes the transaction. Approval revalidates the active
end-user status and current role then stores one existing authorization code. Denial
builds an `access_denied` callback. The frontend navigates to the returned redirect URL.
Redis failures return local `temporarily_unavailable` responses and never issue a
code.

### 4. Preserve OIDC Authentication Context

Carry the authorization transaction's `nonce` and the end-user session's
`AuthenticatedAt` value into the approved authorization code. Extend
`services.TokenClaims` and authorization-code token issuance so ID tokens include
`nonce` when requested and `auth_time` when authentication context is available.

### 5. End the Provider Session

`POST /oauth2/logout` is idempotent. It deletes the Redis SSO session when possible
and always returns an expired host-only cookie. It does not revoke external-client
tokens and does not accept a post-logout redirect URI. If Redis is unavailable, the
handler still expires the browser cookie.

### 6. Audit Security Outcomes

Extend the existing audit event constants and use the existing audit abstraction to
record login success, login failure, throttling, consent approval, consent denial,
logout, and invalid callback attempts. Include known tenant, client, user,
transaction, source-IP, and outcome identifiers. Never record passwords, cookies,
authorization codes, or PKCE values. If PostgreSQL audit persistence fails, emit
the same sanitized event through structured application logging and continue the
protocol outcome; audit storage availability must not become a new OAuth outage
dependency.

### 7. Render Safe Errors

Invalid client and callback URI failures redirect to
`FRONTEND_URL/oauth2/error?...` because no external callback is trusted.
Authorization-initialization Redis failures use that same local frontend redirect;
JSON interaction endpoints return local `503` error responses. Later callback-safe
OAuth failures use the validated callback URI and preserve `state`.

### 8. Derive Source IP Safely

OAuth login throttling and audit events use the request's direct TCP peer address.
Do not consume `X-Forwarded-For`, `Forwarded`, or similar headers in this feature.
Deployments that later add a reverse proxy must introduce and validate an explicit
trusted-proxy allowlist before enabling forwarded client-IP parsing.

### 9. Configuration

Add and validate:

- `FRONTEND_URL`
- `SECURITY_COOKIE_SECURE`
- `SECURITY_SESSION_TTL`
- `OAUTH_AUTH_TRANSACTION_TTL`
- `RATE_LIMIT_OAUTH_LOGIN_FAILURES`
- `RATE_LIMIT_OAUTH_LOGIN_WINDOW_SECONDS`

Document local Docker Compose values and production HTTPS requirements. Do not add
cookie-domain configuration: the SSO cookie is deliberately host-only.

## Test Strategy

### Backend Unit Tests

- Initial request validation: cryptographically random transaction and CSRF
  identifiers, S256 PKCE, registered callback, scope, prompt list, `prompt=none`
  conflict, unsupported `select_account`, and non-negative `max_age`.
- Login: tenant-scoped lookup, generic failures, active status, role, bcrypt,
  cryptographically random client-agnostic session creation, last-login update,
  transaction binding, fixed-window dual-key throttling, direct-peer source-IP
  extraction, email bucket clearing on success, and Redis failure.
- Consent: detail ownership, CSRF token, active-user and role re-check, approval,
  denial, expiry, replay protection, and Redis failure.
- OIDC claims: approved-code propagation and ID-token output for `nonce` and
  `auth_time`.
- Logout: idempotent Redis deletion and unconditional cookie expiry.
- Audit: required event coverage, secret-field exclusion, and sanitized structured
  log fallback when audit persistence fails.
- Config: defaults, invalid TTL and limiter values, frontend URL validation, and
  secure production settings.

### Backend Integration Tests

- `/oauth2/auth` redirects unauthenticated users to frontend login.
- Eligible client-agnostic session skips login across clients in the same tenant;
  disabled users, role loss, `prompt=login`, and exceeded `max_age` do not.
- `/oauth2/login` sets a host-only cookie with `Path=/` and no `Domain`.
- Either throttle bucket blocks verification after five failures in a fixed
  15-minute window without trusting spoofed forwarded-IP headers.
- Credentialed frontend-origin CORS requests are accepted without wildcard origin
  responses.
- Consent read exposes only display-safe details.
- Approval redirects with code and original `state`; denial redirects with
  `access_denied`.
- `/oauth2/logout` deletes the session, expires the cookie, and remains successful
  when the session is missing or Redis is unavailable.
- Redis outages fail closed for auth, login, and consent with local errors.
- Invalid callback stays local; callback query parameters are preserved.
- Header-only `X-User-ID` no longer authenticates.
- Approved code still exchanges once through `/oauth2/token`, while refresh,
  revocation, introspection, discovery, JWKS, and userinfo regression suites pass.

### Frontend Tests

- Route wiring for login, consent, logout, and error pages.
- OAuth login form sends credentials with the transaction identifier and navigates
  to the returned URL.
- Consent container loads details, renders existing `ConsentScreen`, submits allow
  and deny, and navigates to the returned callback.
- Logout page calls provider-local logout and renders a signed-out state.
- Error page maps common OAuth codes including `temporarily_unavailable` to
  user-friendly text.

### Manual Docker Compose Test

Follow [quickstart.md](quickstart.md) with a seeded client and end-user, then verify
approve, deny, session reuse, forced login, silent errors, dual-key throttling,
logout, Redis-outage behavior, and PKCE rejection.

## Complexity Tracking

No constitution violations require justification.
