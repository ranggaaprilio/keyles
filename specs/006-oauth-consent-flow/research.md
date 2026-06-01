# Research: OAuth Consent Flow and End-User Authentication

**Feature**: [spec.md](spec.md)  
**Date**: June 1, 2026

## Decision 1: Store OAuth Interaction State Server-Side

**Decision**: Create a short-lived Redis-backed authorization transaction after
validating the initial `/oauth2/auth` request. Frontend login and consent routes
receive only an opaque `transaction_id`.

**Rationale**: OAuth request fields such as `client_id`, `redirect_uri`, `scope`,
`state`, and PKCE challenge must not become trusted merely because a browser
resubmits them. Keeping the validated values in Redis gives the login and consent
steps a single source of truth, supports expiry and replay protection, and prevents
callback URI tampering.

**Alternatives considered**:

- Forward all OAuth parameters through frontend query strings and POST them back:
  rejected because it expands the tampering surface and duplicates validation.
- Put the full request in an encrypted browser token: rejected because it adds key
  management and rotation complexity without a benefit over Redis already used by
  the service.

## Decision 2: Use a Separate Redis Session and HttpOnly Cookie

**Decision**: Reuse and extend `SessionRepository` for client-agnostic browser SSO
sessions. Remove the legacy session `ClientID` field, store an opaque session
identifier in an `HttpOnly`, `SameSite=Lax`, host-only cookie with `Path=/`, no
`Domain` attribute, and `Secure` outside local development. Keep admin JWT
authentication unchanged.

**Rationale**: The SSO provider must own the end-user login session while external
clients receive OAuth tokens. An opaque cookie avoids exposing session contents to
frontend JavaScript and lets the server revoke sessions centrally.

**Alternatives considered**:

- Reuse the tenant-admin JWT: rejected because admin portal authentication and
external end-user SSO have different identities and trust boundaries.
- Store user identity directly in a readable cookie or local storage: rejected
  because browser scripts should not control the server-side login identity.
- Configure a parent-domain cookie: rejected because the backend host is the only
  component that needs the session cookie.

## Decision 3: Separate Initial Validation From Final Code Issuance

**Decision**: Refactor the authorization use cases so initial authorization
validates the client-facing request and creates a transaction, while consent
finalization revalidates the end-user and role then creates exactly one existing
five-minute authorization code.

**Rationale**: `AuthorizeClient.Execute` currently validates and immediately issues
a code, while `ConsentDecisionUseCase` independently implements another code
generator. A single finalization path removes duplicate behavior and ensures codes
are issued only after explicit consent.

**Alternatives considered**:

- Keep both generators and choose one in handlers: rejected because the two paths
  can diverge on tenant, PKCE, role, and expiry checks.
- Generate a code before consent and hold it in the browser: rejected because code
  issuance must follow authorization approval.

## Decision 4: Preserve OAuth `state`; Add Provider Interaction CSRF Protection

**Decision**: Preserve the client-supplied OAuth `state` unchanged in callback
responses. Separately create a random interaction CSRF token in the authorization
transaction, expose it only with authenticated consent details, and require it on
consent submission along with the SameSite session cookie.

**Rationale**: OAuth `state` binds the external client's request and callback; it is
not a substitute for protecting Keyles' own consent form. RFC 6749 recommends a
non-guessable client binding value in `state`, while the provider interaction needs
its own server-validated binding.

**Alternatives considered**:

- Use OAuth `state` as the consent POST CSRF token: rejected because it belongs to
  the external client and is intentionally returned through redirects.
- Depend only on transaction secrecy: rejected because an explicit interaction
  token makes the provider-side CSRF control testable and clearer.

## Decision 5: Implement OIDC Prompt and Authentication Age Semantics

**Decision**:

- `prompt=login`: require active reauthentication.
- `max_age`: require active reauthentication when the session authentication time
  is older than the requested non-negative number of seconds.
- `prompt=consent`: display consent.
- `prompt=none`: never display login or consent UI. Return `login_required` if no
  eligible session exists, otherwise `consent_required` because persisted consent
  grants are outside this feature.
- `select_account`: reject as unsupported in this single-session MVP.
- Reject `prompt=none` combined with any other prompt value.

**Rationale**: OpenID Connect Core defines `prompt` as a space-delimited set,
forbids combining `none` with other values, requires no UI for `none`, and requires
reauthentication when `max_age` is exceeded. The session therefore needs an
`AuthenticatedAt` timestamp.

**Alternatives considered**:

- Ignore unsupported prompt values: rejected because clients would receive
  behavior different from what they requested.
- Persist grants in PostgreSQL in this feature: deferred because the epic requires
  visible consent, while grant management adds schema, revocation, and UX scope.

## Decision 6: Keep Invalid Callback URI Failures Local

**Decision**: Render local frontend error pages when `client_id` or `redirect_uri`
cannot be trusted. Redirect callback-safe failures only after the callback URI has
passed exact registered-URI validation.

**Rationale**: RFC 6749 requires the authorization server not to redirect the
user-agent to a missing, invalid, or mismatched redirect URI. Existing callback
query parameters must be preserved when adding OAuth response fields.

**Alternatives considered**:

- Redirect every OAuth error to the supplied callback: rejected because an attacker
  could turn Keyles into an open redirect.

## Decision 7: Use Existing Stack and Testing Tools

**Decision**: Implement with Go 1.23, Gin, Redis, PostgreSQL/GORM, React 18,
TypeScript 5, React Router, Axios/fetch, Vitest, React Testing Library, and Go's
testing package with Testify.

**Rationale**: All needed infrastructure already exists in the repository. Redis is
already used for authorization codes and SSO sessions, and the frontend already
contains an unmounted consent component and OAuth types.

**Alternatives considered**:

- Add a new OAuth framework or frontend state library: rejected because the missing
  behavior fits the existing architecture and dependencies.

## Decision 8: Throttle OAuth Login Failures by IP and Tenant Email

**Decision**: Use Redis-backed fixed 15-minute counters for source IP address and
tenant-scoped normalized email. Start each TTL with the first failure and do not
extend it on later failures. Reject further verification when either key reaches
five failures. Increment both counters atomically on failed authentication and clear
the email counter after success.

**Rationale**: Dual-key throttling reduces password spraying and targeted attacks
without allowing an attacker to evade limits by changing only the attempted email.
Redis keeps limits consistent across backend replicas.

**Alternatives considered**:

- IP-only throttling: rejected because distributed attackers can continue targeting
  one account.
- Email-only throttling: rejected because an attacker can spray many accounts from
  one source.

## Decision 9: Add Provider-Local Logout

**Decision**: Add idempotent `POST /oauth2/logout`. Delete the current Redis session
when possible and always expire the browser cookie. Do not accept client-controlled
post-logout redirect URIs and do not revoke client tokens.

**Rationale**: End-users need a predictable way to end Keyles browser SSO without
expanding the MVP into OIDC RP-initiated logout.

**Alternatives considered**:

- Cookie expiry only: rejected because a live Redis session should be revoked when
  reachable.
- Full RP-initiated logout: deferred because client redirect validation and token
  lifecycle coordination are outside this feature.

## Decision 10: Audit Browser-Flow Security Outcomes

**Decision**: Extend the existing audit event constants and record structured audit
events for login success, login failure, throttling, consent approval, consent
denial, logout, and invalid callback attempts. Include known identifiers and
outcome codes, but exclude credentials, cookies, authorization codes, and PKCE
values. If PostgreSQL audit persistence fails, emit the same sanitized event
through structured application logging and continue the protocol outcome.

**Rationale**: These events support incident review and abuse investigation while
keeping secrets out of logs and audit storage. A structured-log fallback avoids
turning PostgreSQL audit availability into an OAuth service dependency.

**Alternatives considered**:

- Rely only on HTTP access logs: rejected because access logs do not contain the
  domain outcome context required for security review.

## Decision 11: Fail Closed When Redis Is Unavailable

**Decision**: Authorization initialization, login, consent reads, and consent
submissions return local `temporarily_unavailable` errors when Redis is unavailable.
They do not issue codes or redirect external callback errors. Logout still expires
the browser cookie even if Redis deletion fails.

**Rationale**: Sessions, transactions, and throttle counters are security controls.
Bypassing them or falling back to per-process memory risks inconsistent
authorization behavior.

**Alternatives considered**:

- In-memory fallback: rejected because replicas would disagree about session,
  replay, and throttle state.
- Continue without throttling or sessions: rejected because it weakens security
  controls during an outage.

## Decision 12: Use the Direct TCP Peer as the Source IP

**Decision**: Derive OAuth login-throttle and audit source-IP values from the direct
TCP peer address. Do not trust `X-Forwarded-For`, `Forwarded`, or similar headers in
this feature.

**Rationale**: The current service has no trusted-proxy allowlist. Accepting
forwarded-IP headers directly would let a client spoof throttle and audit identity.
A reverse-proxy deployment can add forwarded parsing later only alongside an
explicit trusted-proxy configuration.

**Alternatives considered**:

- Trust forwarded headers unconditionally: rejected because direct clients could
  rotate spoofed values and bypass source-IP throttling.

## Primary References

- [OAuth 2.0 Authorization Framework, RFC 6749](https://www.rfc-editor.org/rfc/rfc6749)
- [Proof Key for Code Exchange, RFC 7636](https://www.rfc-editor.org/rfc/rfc7636)
- [OpenID Connect Core 1.0, Authorization Endpoint](https://openid.net/specs/openid-connect-core-1_0-18.html#AuthorizationEndpoint)
