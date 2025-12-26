# Feature Specification: Core SSO Auth Provider

**Feature Branch**: `003-sso-auth-provider`  
**Created**: December 26, 2025  
**Status**: Draft  
**Input**: User description: "Design the logic architecture and workflow for a Core SSO Auth Provider based on the OpenID Connect (OIDC) protocol for a multi-tenant SaaS platform."

## Clarifications

### Session 2025-12-26

- Q: How should the system identify which tenant context to use during authentication? → A: Client-based (tenant determined by client_id lookup)
- Q: How should the system determine if a user is allowed to authenticate with a specific client application? → A: Role-based access control
- Q: What rate limits should be enforced on token endpoints? → A: 10 requests/minute per client_id

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Client App Registration (Priority: P1)

A tenant administrator needs to register a new client application (e.g., mobile app, web portal, third-party integration) that will use SSO authentication to access tenant resources. The administrator provides the client name, allowed redirect URIs, and receives client credentials.

**Why this priority**: Without client registration, no applications can authenticate users through the SSO provider. This is the foundational capability that enables all other SSO functionality.

**Independent Test**: Can be fully tested by creating a tenant, registering a client app with redirect URIs, and verifying that client_id and client_secret are generated and stored. Delivers immediate value by allowing tenant admins to prepare their applications for SSO integration.

**Acceptance Scenarios**:

1. **Given** a tenant administrator is logged into the admin portal, **When** they navigate to client app management and create a new client with name "Mobile App" and redirect URI "https://app.example.com/callback", **Then** the system generates a unique client_id and client_secret, stores the redirect URI, and displays the credentials for the administrator to copy.

2. **Given** a tenant administrator attempts to register a client, **When** they provide an invalid redirect URI format (not HTTPS in production), **Then** the system rejects the registration with a clear validation error message.

3. **Given** a tenant administrator has registered a client, **When** they view the client list, **Then** they see all registered clients with their client_ids, names, and allowed redirect URIs (but not secrets).

---

### User Story 2 - User Authentication Flow (Priority: P1)

An end user attempts to access a client application that requires authentication. They are redirected to the SSO provider, enter their credentials, and upon successful authentication, are redirected back to the client application with an authorization code.

**Why this priority**: This is the core SSO authentication experience. Without this, users cannot sign in to any applications. This delivers the primary value proposition of single sign-on.

**Independent Test**: Can be fully tested by initiating an OAuth2 authorization request from a registered client, completing the login flow, and verifying that an authorization code is returned to the redirect URI. Delivers immediate user value by enabling secure authentication.

**Acceptance Scenarios**:

1. **Given** an end user clicks "Sign in with SSO" in a client application, **When** they are redirected to the SSO provider with valid client_id, redirect_uri, code_challenge, and state parameters, **Then** they see the SSO login page.

2. **Given** a user is on the SSO login page, **When** they enter valid credentials for their tenant, **Then** the system validates the credentials, creates a user session, and redirects them back to the client's redirect_uri with an authorization_code and the original state parameter.

3. **Given** a user attempts to authenticate, **When** the client_id is invalid or the redirect_uri doesn't match the registered URIs, **Then** the system rejects the request and displays an error message without redirecting.

4. **Given** a user successfully authenticates, **When** the authorization code is generated, **Then** it expires after 5 minutes and can only be used once.

---

### User Story 3 - Token Exchange (Priority: P1)

A client application exchanges an authorization code for tokens. The client sends the authorization code along with the code_verifier (PKCE) to the SSO token endpoint and receives ID tokens, access tokens, and refresh tokens.

**Why this priority**: Without token exchange, the authorization code cannot be converted into usable authentication credentials. This completes the authentication flow and enables client applications to verify user identity and access resources.

**Independent Test**: Can be fully tested by obtaining an authorization code (from Story 2) and exchanging it with correct code_verifier and client credentials. Delivers value by providing JWTs that applications can validate and use for authorization decisions.

**Acceptance Scenarios**:

1. **Given** a client application has received an authorization code, **When** it sends a POST request to the token endpoint with the code, code_verifier, client_id, client_secret, and redirect_uri, **Then** the system validates all parameters and returns an ID token, access token, and refresh token.

2. **Given** a token exchange request, **When** the code_verifier doesn't match the original code_challenge (PKCE validation fails), **Then** the system rejects the request with an "invalid_grant" error.

3. **Given** a token exchange request, **When** the authorization code has expired (more than 5 minutes old) or has already been used, **Then** the system rejects the request with an "invalid_grant" error.

4. **Given** a successful token exchange, **When** the tokens are generated, **Then** the ID token contains user profile claims and tenant_id, the access token contains scope and short expiry, and the refresh token is valid for extended duration.

---

### User Story 4 - Token Validation & Signature Verification (Priority: P2)

A client application or resource server needs to validate tokens received from users. They fetch the public keys from the JWKS endpoint and verify the token signature to ensure it was issued by the trusted SSO provider.

**Why this priority**: This enables distributed token validation without requiring every token to be checked against the SSO server. Critical for scalability and security, but can be deferred slightly as initial implementations might use simpler validation.

**Independent Test**: Can be fully tested by obtaining a signed JWT token and verifying its signature using the public key from the JWKS endpoint. Delivers value by enabling secure, offline token validation.

**Acceptance Scenarios**:

1. **Given** a client application receives an ID token or access token, **When** it fetches the public keys from the SSO provider's JWKS endpoint and validates the token signature, **Then** the signature validation succeeds for valid tokens.

2. **Given** a token validation attempt, **When** the token signature doesn't match any public key in the JWKS, **Then** the validation fails and the token is rejected.

3. **Given** a client application validates a token, **When** the token has expired based on the "exp" claim, **Then** the validation fails even if the signature is valid.

4. **Given** a token validation attempt, **When** the tenant_id in the token doesn't match the expected tenant, **Then** the validation fails to prevent cross-tenant token usage.

---

### User Story 5 - Token Refresh (Priority: P2)

A client application's access token has expired, but the user session should continue. The client sends the refresh token to obtain a new access token without requiring the user to log in again.

**Why this priority**: Improves user experience by maintaining sessions without interruption. Important for production use but not critical for initial MVP authentication flows.

**Independent Test**: Can be fully tested by obtaining a refresh token, waiting for access token expiry, and exchanging the refresh token for a new access token. Delivers value by enabling long-lived sessions with short-lived access tokens.

**Acceptance Scenarios**:

1. **Given** a client application has a valid refresh token, **When** it sends the refresh token to the token endpoint, **Then** the system validates the refresh token and returns a new access token and optionally a new refresh token.

2. **Given** a refresh token request, **When** the refresh token has been revoked or doesn't exist, **Then** the system rejects the request with an "invalid_grant" error.

3. **Given** a refresh token request, **When** the client_id doesn't match the original client that obtained the refresh token, **Then** the system rejects the request to prevent token theft.

---

### User Story 6 - Token Revocation (Priority: P3)

A tenant administrator or user needs to revoke access tokens or refresh tokens (e.g., when a device is lost, an employee leaves, or suspicious activity is detected). The revoked tokens immediately become invalid.

**Why this priority**: Essential for security incident response but not needed for basic authentication flows. Can be added after core functionality is stable.

**Independent Test**: Can be fully tested by revoking a refresh token and attempting to use it for token refresh. Delivers security value by enabling administrators to terminate sessions remotely.

**Acceptance Scenarios**:

1. **Given** a tenant administrator views active sessions, **When** they select a session and revoke it, **Then** the associated refresh token is marked as revoked and all future attempts to use it fail.

2. **Given** a user logs out from a client application, **When** the client calls the revocation endpoint with the refresh token, **Then** the token is revoked and the user session is terminated.

3. **Given** a refresh token has been revoked, **When** any client attempts to use it for token refresh, **Then** the system returns an "invalid_grant" error.

---

### User Story 7 - Multi-Client Management (Priority: P3)

A tenant has multiple client applications (web app, mobile app, admin portal) that all need separate configurations but share the same user base. The tenant administrator manages each client's credentials and redirect URIs independently.

**Why this priority**: While important for production scenarios with multiple apps, a single client registration is sufficient for MVP. This can be layered on after core flows work.

**Independent Test**: Can be fully tested by registering multiple clients for one tenant and verifying that each has independent credentials and configurations. Delivers value by supporting complex multi-app ecosystems.

**Acceptance Scenarios**:

1. **Given** a tenant administrator has registered multiple clients, **When** they update the redirect URIs for one client, **Then** only that client's configuration changes and other clients are unaffected.

2. **Given** multiple clients exist for a tenant, **When** a user authenticates through Client A, **Then** the tokens issued are specific to Client A and cannot be used by Client B.

3. **Given** a tenant administrator needs to rotate credentials, **When** they regenerate the client_secret for one client, **Then** the old secret is invalidated for that client only.

---

### User Story 8 - User Role Management (Priority: P2)

A tenant administrator needs to control which users in their organization can authenticate with specific client applications. They assign roles or permissions to users for each client application.

**Why this priority**: Required to support role-based access control clarified in requirements. Essential for enterprise security but can be implemented after basic authentication flows are working.

**Independent Test**: Can be fully tested by creating users, assigning roles to specific clients, and verifying that users without roles are denied access during authentication. Delivers security value by enabling fine-grained access control.

**Acceptance Scenarios**:

1. **Given** a tenant administrator views a client application, **When** they assign a user to that client with a specific role, **Then** the user can successfully authenticate with that client.

2. **Given** a user without any role assignment for a client, **When** they attempt to authenticate with that client, **Then** the system denies access with an "insufficient_permissions" error.

3. **Given** a tenant administrator has assigned roles, **When** they revoke a user's role for a client, **Then** the user can no longer authenticate with that client and any active refresh tokens for that client-user combination are revoked.

4. **Given** a tenant administrator, **When** they view role assignments, **Then** they see which users have access to which clients and with what roles.

---

### Edge Cases

- What happens when a user's account is disabled while they have an active refresh token?
- How does the system handle clock skew when validating token expiration times?
- What happens when a client attempts to use an authorization code from a different tenant?
- How does the system prevent authorization code interception attacks?
- What happens if the redirect_uri contains query parameters that need to be preserved?
- What happens when rate limit (10 requests/minute per client) is exceeded?
- What happens when PKCE parameters (code_challenge, code_verifier) are missing from PKCE-required flows?
- How does the system handle concurrent token exchanges using the same authorization code?
- What happens when a tenant is deleted but has active tokens in circulation?
- What happens when a user's role for a client is revoked while they have an active session?
- How does the system handle a user who belongs to multiple tenants trying to authenticate?

## Requirements _(mandatory)_

### Functional Requirements

#### Client Management

- **FR-001**: System MUST allow tenant administrators to register client applications with a unique client_id, client_secret, client name, and one or more allowed redirect URIs.
- **FR-002**: System MUST validate that all redirect URIs use HTTPS protocol (except localhost for development environments).
- **FR-003**: System MUST generate cryptographically secure, unique client_id and client_secret values for each registered client.
- **FR-004**: System MUST allow tenant administrators to view, update, and delete registered client applications.
- **FR-005**: System MUST allow tenant administrators to regenerate client_secret without changing client_id.
- **FR-006**: System MUST enforce tenant isolation such that clients registered by Tenant A cannot be accessed or modified by Tenant B.

#### User Role Management

- **FR-006a**: System MUST allow tenant administrators to assign role-based permissions to users for specific client applications.
- **FR-006b**: System MUST allow tenant administrators to view, modify, and revoke user role assignments for clients.
- **FR-006c**: System MUST support multiple roles per user-client combination (e.g., "admin", "user", "viewer").
- **FR-006d**: System MUST deny authentication attempts when a user has no active role assignment for the requested client application.
- **FR-006e**: System MUST revoke all active refresh tokens for a user-client combination when their role assignment is revoked.

#### Authorization Flow (OIDC with PKCE)

- **FR-007**: System MUST implement the OAuth 2.0 Authorization Code Flow as defined in RFC 6749.
- **FR-008**: System MUST support PKCE (Proof Key for Code Exchange) as defined in RFC 7636 to prevent authorization code interception attacks.
- **FR-009**: System MUST validate the following parameters in authorization requests: client_id, redirect_uri, response_type=code, code_challenge, code_challenge_method=S256, state, scope.
- **FR-010**: System MUST verify that the redirect_uri in the authorization request exactly matches one of the registered redirect URIs for the client.
- **FR-010a**: System MUST determine the tenant context by looking up the tenant_id associated with the client_id from the authorization request.
- **FR-011**: System MUST present an authentication form to unauthenticated users when they arrive at the authorization endpoint.
- **FR-012**: System MUST validate user credentials against the tenant's user repository and verify the user has appropriate role-based permissions to access the requested client application.
- **FR-013**: System MUST create a user session upon successful authentication.
- **FR-014**: System MUST generate a single-use authorization code that expires after 5 minutes.
- **FR-015**: System MUST store the code_challenge and code_challenge_method associated with each authorization code for later PKCE verification.
- **FR-016**: System MUST redirect the user back to the client application's redirect_uri with the authorization_code and state parameter.
- **FR-017**: System MUST reject authorization requests where the client_id is invalid or inactive.
- **FR-018**: System MUST reject authorization requests where required parameters are missing or malformed.

#### Token Exchange

- **FR-019**: System MUST provide a token endpoint that exchanges authorization codes for tokens.
- **FR-020**: System MUST validate the following parameters in token exchange requests: grant_type=authorization_code, code, redirect_uri, code_verifier, client_id, client_secret.
- **FR-021**: System MUST verify that the code_verifier matches the code_challenge stored with the authorization code using the S256 method (SHA-256 hash).
- **FR-022**: System MUST verify that the redirect_uri in the token request matches the redirect_uri from the authorization request.
- **FR-023**: System MUST verify that the authorization code has not expired (under 5 minutes old).
- **FR-024**: System MUST verify that the authorization code has not been previously used (one-time use).
- **FR-025**: System MUST verify that the client_id and client_secret are valid and match the client that received the authorization code.
- **FR-026**: System MUST revoke an authorization code immediately after it is used for token exchange.
- **FR-027**: System MUST issue an ID token, access token, and refresh token upon successful token exchange.
- **FR-028**: System MUST reject token exchange requests with "invalid_grant" error when validation fails.

#### Token Specifications

- **FR-029**: System MUST generate tokens as signed JSON Web Tokens (JWT) conforming to RFC 7519.
- **FR-030**: ID token MUST contain standard OIDC claims: iss (issuer), sub (subject/user ID), aud (audience/client_id), exp (expiration time), iat (issued at time), and tenant_id.
- **FR-031**: ID token MUST contain user profile claims: email, name, email_verified, and any other tenant-specific claims.
- **FR-032**: Access token MUST contain: iss, sub, aud, exp, iat, scope, and tenant_id.
- **FR-033**: Access token MUST have a short validity period of 15 minutes.
- **FR-034**: Refresh token MUST be an opaque token (not a JWT) stored in the database with a reference to the user, client, and tenant.
- **FR-035**: Refresh token MUST have a validity period of 7 days.
- **FR-036**: System MUST sign all JWT tokens using RS256 (RSA Signature with SHA-256) asymmetric algorithm.
- **FR-037**: System MUST maintain a private key for signing tokens and expose public keys via a JWKS (JSON Web Key Set) endpoint.
- **FR-038**: System MUST include the "kid" (key ID) in the JWT header to identify which public key should be used for verification.

#### Token Validation

- **FR-039**: System MUST provide a JWKS endpoint (/.well-known/jwks.json) that returns public keys in JWKS format.
- **FR-040**: System MUST support key rotation by maintaining multiple active public keys in the JWKS endpoint.
- **FR-041**: Client applications and resource servers MUST be able to validate token signatures using the public keys from the JWKS endpoint.
- **FR-042**: System MUST validate that tokens are not expired based on the "exp" claim.
- **FR-043**: System MUST validate that tokens are not used before their "iat" (issued at) or "nbf" (not before) time.
- **FR-044**: System MUST enforce tenant isolation by validating that the tenant_id in the token matches the expected tenant context.

#### Token Refresh

- **FR-045**: System MUST provide a token endpoint that accepts refresh_token grant type.
- **FR-046**: System MUST validate that the refresh token exists, has not expired, and has not been revoked.
- **FR-047**: System MUST verify that the client_id in the refresh request matches the client that originally obtained the refresh token.
- **FR-048**: System MUST issue a new access token (and optionally a new refresh token) upon successful refresh.
- **FR-049**: System MUST allow refresh tokens to be used multiple times until they expire or are revoked.
- **FR-050**: System MUST update the last_used_at timestamp on refresh tokens each time they are used.

#### Token Revocation

- **FR-051**: System MUST provide a revocation endpoint that accepts tokens to be revoked.
- **FR-052**: System MUST allow tenant administrators to revoke refresh tokens for specific users or sessions.
- **FR-053**: System MUST mark revoked refresh tokens as invalid and reject any future attempts to use them.
- **FR-054**: System MUST revoke all refresh tokens for a user when their account is disabled or deleted.
- **FR-055**: System MUST provide a mechanism to check if a refresh token has been revoked during token refresh attempts.

#### Security & Compliance

- **FR-056**: System MUST enforce HTTPS for all OAuth/OIDC endpoints in production environments.
- **FR-057**: System MUST implement rate limiting on token endpoints with a limit of 10 requests per minute per client_id to prevent brute force attacks.
- **FR-058**: System MUST log all authentication attempts, token exchanges, and revocations for security audit purposes.
- **FR-059**: System MUST prevent timing attacks by using constant-time comparison for secret values.
- **FR-060**: System MUST validate that the state parameter returned in the authorization callback matches the original state to prevent CSRF attacks.
- **FR-061**: System MUST implement proper error handling that doesn't leak sensitive information (e.g., whether a client_id exists).

#### Discovery & Standards Compliance

- **FR-062**: System MUST provide an OpenID Connect Discovery document at /.well-known/openid-configuration.
- **FR-063**: Discovery document MUST include: issuer, authorization_endpoint, token_endpoint, jwks_uri, userinfo_endpoint, revocation_endpoint, scopes_supported, response_types_supported, grant_types_supported, token_endpoint_auth_methods_supported, and code_challenge_methods_supported.

### Key Entities

- **Client Application**: Represents a registered application that can authenticate users via the SSO provider. Attributes include: client_id (unique identifier), client_secret (confidential credential), client_name (human-readable name), redirect_uris (list of allowed callback URLs), tenant_id (owning tenant), created_at, updated_at, is_active (enable/disable flag).

- **Authorization Code**: A short-lived, single-use code issued after successful user authentication. Attributes include: code (unique identifier), client_id (associated client), user_id (authenticated user), tenant_id (tenant context), redirect_uri (callback URL), code_challenge (PKCE challenge), code_challenge_method (PKCE method), scope (requested permissions), expires_at (5-minute expiration), used (boolean flag for one-time use).

- **Refresh Token**: A long-lived, revocable token used to obtain new access tokens. Attributes include: token (unique identifier), client_id (associated client), user_id (token owner), tenant_id (tenant context), expires_at (long-term expiration), created_at, last_used_at, is_revoked (revocation flag), revoked_at.

- **User Session**: Represents an authenticated user's session with the SSO provider. Attributes include: session_id (unique identifier), user_id (authenticated user), tenant_id (tenant context), created_at, expires_at, last_activity_at, ip_address, user_agent.

- **Signing Key Pair**: Represents the asymmetric cryptographic keys used for signing and verifying tokens. Attributes include: key_id (kid), public_key (for distribution via JWKS), private_key (kept secure, used for signing), algorithm (RS256), created_at, is_active (for key rotation), expires_at (optional key expiration).

- **Audit Log**: Records all security-relevant events in the SSO system. Attributes include: event_type (authentication_attempt, token_issued, token_revoked, etc.), user_id, tenant_id, client_id, timestamp, ip_address, user_agent, success (boolean), error_message (if failed), metadata (additional context).

- **User Role Assignment**: Represents role-based permissions that control which users can authenticate with which client applications. Attributes include: user_id (user being granted access), client_id (client application), role (permission level or role name), tenant_id (tenant context), granted_at (timestamp), granted_by (administrator who assigned the role), is_active (enable/disable flag).

### Assumptions

1. **PKCE Enforcement**: Assuming PKCE is mandatory for all authorization flows to maximize security (even for confidential clients).
2. **Single Tenant Per User Context**: Assuming each user belongs to only one tenant, determined automatically via the client_id lookup during authentication.
3. **Key Rotation Strategy**: Assuming a manual key rotation process initially, with automated rotation as a future enhancement.
4. **Scope Support**: Assuming basic scope support (openid, profile, email) with extensibility for custom scopes.
5. **Multi-Factor Authentication**: Assuming MFA is handled separately and is not part of this SSO provider specification.
6. **User Consent**: Assuming implicit consent (no consent screen) for first-party client applications; third-party applications would require explicit consent in future iterations.
7. **Session Management**: Assuming sessions are stored server-side with a default timeout of 8 hours of inactivity.
8. **Rate Limit Response**: Assuming HTTP 429 (Too Many Requests) is returned when rate limits are exceeded, with retry-after header.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: Users can complete the full authentication flow (from client redirect to receiving tokens) in under 10 seconds under normal network conditions.
- **SC-002**: System can handle 500 concurrent authentication requests without performance degradation or timeouts.
- **SC-003**: 99% of valid authorization codes are successfully exchanged for tokens on the first attempt.
- **SC-004**: Token signature verification using public keys from JWKS endpoint succeeds with 100% accuracy for valid tokens.
- **SC-005**: Authorization code interception attacks are prevented through PKCE validation with 100% effectiveness (no successful attacks with invalid code_verifier).
- **SC-006**: Cross-tenant token usage attempts are blocked 100% of the time through tenant_id validation.
- **SC-007**: Expired authorization codes (over 5 minutes old) are rejected 100% of the time.
- **SC-008**: Revoked refresh tokens cannot be used to obtain new access tokens with 100% reliability.
- **SC-009**: System maintains 99.9% uptime for all OAuth/OIDC endpoints (authorization, token, JWKS).
- **SC-010**: Token generation and signing operations complete in under 100 milliseconds on average.
- **SC-011**: Public key retrieval from JWKS endpoint completes in under 50 milliseconds on average.
- **SC-012**: 100% of security-relevant events (authentication attempts, token issuance, revocations) are logged with complete audit trails.
- **SC-013**: Zero incidents of tokens being accepted after their expiration time.
- **SC-014**: Client applications can integrate with the SSO provider by following the OpenID Connect Discovery document without additional documentation.
- **SC-015**: Redirect URI validation prevents 100% of open redirect vulnerabilities.
- **SC-016**: Rate limiting successfully blocks requests exceeding 10 requests per minute per client_id with appropriate HTTP 429 responses.
- **SC-017**: Users without role-based permissions for a client are denied authentication 100% of the time.
- **SC-018**: Tenant isolation ensures that client_id lookups correctly identify tenant context with 100% accuracy.

## Workflow Documentation _(mandatory)_

This section provides step-by-step documentation of the validation and processing workflow from the time a request is received until tokens are issued, including JSON data structures.

### Workflow 1: Authorization Request to Authorization Code Issuance

**Step 1: Client Initiates Authorization Request**

The client application redirects the user to the SSO provider's authorization endpoint.

**Request Parameters**:

```json
{
  "client_id": "abc123xyz",
  "redirect_uri": "https://client-app.example.com/callback",
  "response_type": "code",
  "scope": "openid profile email",
  "state": "random-csrf-token-12345",
  "code_challenge": "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
  "code_challenge_method": "S256"
}
```

**Step 2: Authorization Endpoint Validation**

System validates the incoming request:

1. Verify `client_id` exists and is active
2. Lookup and set tenant context based on the client's tenant_id
3. Verify `redirect_uri` exactly matches one of the registered URIs for the client
4. Verify `response_type` is "code"
5. Verify `code_challenge_method` is "S256"
6. Verify `code_challenge` is present and properly formatted
7. Verify `state` is present (CSRF protection)
8. Verify `scope` contains at minimum "openid"

**Validation Failure Response**:
If validation fails (except for redirect_uri mismatch), display error page to user.
If redirect_uri is invalid, do not redirect to prevent open redirect vulnerabilities.

```json
{
  "error": "invalid_request",
  "error_description": "The redirect_uri does not match any registered URIs for this client"
}
```

**Step 3: User Authentication**

If validation succeeds and user is not authenticated, present login form.

**Login Request**:

```json
{
  "email": "user@tenant-a.com",
  "password": "secure-password-123"
}
```

**Step 4: Credential Validation**

System validates user credentials:

1. Verify user exists in the tenant (tenant context already determined from client_id lookup)
2. Query user repository for the tenant
3. Verify password hash matches
4. Check if user account is active and not disabled
5. Check if user has role-based permissions to access the requested client application (query user role assignments for this client_id)

**Authentication Failure Response**:

```json
{
  "error": "invalid_credentials",
  "error_description": "The email or password you entered is incorrect"
}
```

**Step 5: Session Creation**

Upon successful authentication, create user session:

**Session Object**:

```json
{
  "session_id": "sess_9f8e7d6c5b4a3210",
  "user_id": "user_12345",
  "tenant_id": "tenant_a",
  "email": "user@tenant-a.com",
  "created_at": "2025-12-26T10:30:00Z",
  "expires_at": "2025-12-26T18:30:00Z",
  "ip_address": "203.0.113.42",
  "user_agent": "Mozilla/5.0..."
}
```

**Step 6: Authorization Code Generation**

System generates single-use authorization code:

**Authorization Code Object**:

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
  "created_at": "2025-12-26T10:30:15Z",
  "expires_at": "2025-12-26T10:35:15Z",
  "used": false
}
```

**Step 7: Redirect to Client**

Redirect user back to client application with authorization code:

**Redirect URL**:

```
https://client-app.example.com/callback?code=auth_code_a1b2c3d4e5f6g7h8&state=random-csrf-token-12345
```

---

### Workflow 2: Token Exchange (Authorization Code to Tokens)

**Step 1: Client Sends Token Request**

Client application makes POST request to token endpoint.

**Token Request**:

```json
{
  "grant_type": "authorization_code",
  "code": "auth_code_a1b2c3d4e5f6g7h8",
  "redirect_uri": "https://client-app.example.com/callback",
  "code_verifier": "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
  "client_id": "abc123xyz",
  "client_secret": "secret_98765zyxwvut"
}
```

**Step 2: Token Endpoint Validation**

System validates token exchange request:

1. Verify `grant_type` is "authorization_code"
2. Verify `code` exists in database
3. Verify `code` has not expired (under 5 minutes old)
4. Verify `code` has not been used (`used` field is false)
5. Verify `client_id` and `client_secret` match and are valid
6. Verify `redirect_uri` matches the URI stored with the authorization code
7. Verify `code_verifier` matches the `code_challenge` using S256 method:
   - Compute: `BASE64URL(SHA256(code_verifier))`
   - Compare with stored `code_challenge`
8. Verify tenant_id from code matches tenant_id of client

**PKCE Verification Example**:

```
code_verifier: dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk
SHA256(code_verifier): 4f8...1c2 (binary)
BASE64URL(SHA256(code_verifier)): E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM
Must equal code_challenge: E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM ✓
```

**Validation Failure Response**:

```json
{
  "error": "invalid_grant",
  "error_description": "The authorization code is invalid, expired, or has already been used"
}
```

**Step 3: Mark Authorization Code as Used**

Update authorization code to prevent reuse:

```json
{
  "code": "auth_code_a1b2c3d4e5f6g7h8",
  "used": true,
  "used_at": "2025-12-26T10:30:20Z"
}
```

**Step 4: Generate ID Token (JWT)**

Create ID Token with user profile information.

**ID Token Payload**:

```json
{
  "iss": "https://sso.keyles.com",
  "sub": "user_12345",
  "aud": "abc123xyz",
  "exp": 1735213820,
  "iat": 1735210220,
  "tenant_id": "tenant_a",
  "email": "user@tenant-a.com",
  "email_verified": true,
  "name": "John Doe",
  "given_name": "John",
  "family_name": "Doe"
}
```

**ID Token Header**:

```json
{
  "alg": "RS256",
  "typ": "JWT",
  "kid": "key_2025_01"
}
```

**Signed ID Token** (example):

```
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImtleV8yMDI1XzAxIn0.eyJpc3MiOiJodHRwczovL3Nzby5rZXlsZXMuY29tIiwic3ViIjoidXNlcl8xMjM0NSIsImF1ZCI6ImFiYzEyM3h5eiIsImV4cCI6MTczNTIxMzgyMCwiaWF0IjoxNzM1MjEwMjIwLCJ0ZW5hbnRfaWQiOiJ0ZW5hbnRfYSIsImVtYWlsIjoidXNlckB0ZW5hbnQtYS5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwibmFtZSI6IkpvaG4gRG9lIiwiZ2l2ZW5fbmFtZSI6IkpvaG4iLCJmYW1pbHlfbmFtZSI6IkRvZSJ9.signature_bytes_here
```

**Step 5: Generate Access Token (JWT)**

Create Access Token for API authorization.

**Access Token Payload**:

```json
{
  "iss": "https://sso.keyles.com",
  "sub": "user_12345",
  "aud": "abc123xyz",
  "exp": 1735211120,
  "iat": 1735210220,
  "tenant_id": "tenant_a",
  "scope": "openid profile email",
  "client_id": "abc123xyz"
}
```

**Access Token Header**:

```json
{
  "alg": "RS256",
  "typ": "JWT",
  "kid": "key_2025_01"
}
```

**Signed Access Token** (example):

```
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImtleV8yMDI1XzAxIn0.eyJpc3MiOiJodHRwczovL3Nzby5rZXlsZXMuY29tIiwic3ViIjoidXNlcl8xMjM0NSIsImF1ZCI6ImFiYzEyM3h5eiIsImV4cCI6MTczNTIxMTEyMCwiaWF0IjoxNzM1MjEwMjIwLCJ0ZW5hbnRfaWQiOiJ0ZW5hbnRfYSIsInNjb3BlIjoib3BlbmlkIHByb2ZpbGUgZW1haWwiLCJjbGllbnRfaWQiOiJhYmMxMjN4eXoifQ.signature_bytes_here
```

**Step 6: Generate Refresh Token (Opaque)**

Create Refresh Token as database-backed opaque token.

**Refresh Token Object**:

```json
{
  "token": "refresh_9z8y7x6w5v4u3t2s1r0q",
  "client_id": "abc123xyz",
  "user_id": "user_12345",
  "tenant_id": "tenant_a",
  "created_at": "2025-12-26T10:30:20Z",
  "expires_at": "2026-01-25T10:30:20Z",
  "last_used_at": "2025-12-26T10:30:20Z",
  "is_revoked": false,
  "revoked_at": null
}
```

**Step 7: Return Token Response**

Send tokens to client application:

**Token Response**:

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImtleV8yMDI1XzAxIn0...",
  "token_type": "Bearer",
  "expires_in": 900,
  "id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImtleV8yMDI1XzAxIn0...",
  "refresh_token": "refresh_9z8y7x6w5v4u3t2s1r0q",
  "scope": "openid profile email"
}
```

---

### Workflow 3: Token Validation Using JWKS

**Step 1: Client Receives Token**

Client application receives JWT token (ID token or access token) from user request.

**Step 2: Fetch Public Keys**

Client makes GET request to JWKS endpoint to retrieve public keys.

**JWKS Endpoint Request**:

```
GET https://sso.keyles.com/.well-known/jwks.json
```

**JWKS Response**:

```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "key_2025_01",
      "alg": "RS256",
      "n": "modulus_base64url_encoded_value_here",
      "e": "AQAB"
    },
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "key_2024_12",
      "alg": "RS256",
      "n": "modulus_base64url_encoded_value_here",
      "e": "AQAB"
    }
  ]
}
```

**Step 3: Parse Token Header**

Extract header from JWT to get key ID.

**Token Header**:

```json
{
  "alg": "RS256",
  "typ": "JWT",
  "kid": "key_2025_01"
}
```

**Step 4: Select Public Key**

Match `kid` from token header with JWKS keys. Use the public key with matching `kid`.

**Step 5: Verify Signature**

Use the public key to verify the token signature:

1. Split JWT into header, payload, signature
2. Reconstruct signing input: `base64url(header).base64url(payload)`
3. Verify signature using RS256 algorithm with public key
4. If signature is valid, token is authentic

**Step 6: Validate Claims**

After signature verification, validate token claims:

1. Verify `iss` (issuer) matches expected SSO provider URL
2. Verify `aud` (audience) matches client_id
3. Verify `exp` (expiration) is in the future
4. Verify `iat` (issued at) is not in the future
5. Verify `tenant_id` matches the expected tenant context
6. Verify `sub` (subject) exists and is valid

**Validation Success**: Token is valid and claims can be trusted.

**Validation Failure Examples**:

```json
{
  "error": "invalid_token",
  "error_description": "Token signature verification failed"
}
```

```json
{
  "error": "expired_token",
  "error_description": "Token has expired"
}
```

```json
{
  "error": "invalid_tenant",
  "error_description": "Token tenant_id does not match expected tenant"
}
```

---

### Workflow 4: Token Refresh

**Step 1: Client Sends Refresh Request**

Client sends refresh token to obtain new access token.

**Refresh Request**:

```json
{
  "grant_type": "refresh_token",
  "refresh_token": "refresh_9z8y7x6w5v4u3t2s1r0q",
  "client_id": "abc123xyz",
  "client_secret": "secret_98765zyxwvut"
}
```

**Step 2: Validate Refresh Token**

System validates refresh token:

1. Verify `grant_type` is "refresh_token"
2. Query database for refresh token
3. Verify token exists
4. Verify `is_revoked` is false
5. Verify `expires_at` is in the future
6. Verify `client_id` matches the client that owns the token
7. Verify `client_secret` is correct
8. Verify tenant_id matches

**Validation Failure Response**:

```json
{
  "error": "invalid_grant",
  "error_description": "The refresh token is invalid, expired, or has been revoked"
}
```

**Step 3: Update Refresh Token Usage**

Update last_used timestamp:

```json
{
  "token": "refresh_9z8y7x6w5v4u3t2s1r0q",
  "last_used_at": "2025-12-26T11:45:00Z"
}
```

**Step 4: Generate New Access Token**

Create new access token (same structure as in Workflow 2, Step 5).

**Step 5: Optionally Generate New Refresh Token**

System may issue a new refresh token (refresh token rotation for enhanced security):

**New Refresh Token Object**:

```json
{
  "token": "refresh_8y7x6w5v4u3t2s1r0q9p",
  "client_id": "abc123xyz",
  "user_id": "user_12345",
  "tenant_id": "tenant_a",
  "created_at": "2025-12-26T11:45:00Z",
  "expires_at": "2026-01-25T11:45:00Z",
  "last_used_at": "2025-12-26T11:45:00Z",
  "is_revoked": false,
  "revoked_at": null
}
```

If new refresh token is issued, revoke old one:

```json
{
  "token": "refresh_9z8y7x6w5v4u3t2s1r0q",
  "is_revoked": true,
  "revoked_at": "2025-12-26T11:45:00Z"
}
```

**Step 6: Return Token Response**

Send new tokens to client:

**Refresh Response**:

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImtleV8yMDI1XzAxIn0...",
  "token_type": "Bearer",
  "expires_in": 900,
  "refresh_token": "refresh_8y7x6w5v4u3t2s1r0q9p",
  "scope": "openid profile email"
}
```

---

### Workflow 5: Token Revocation

**Step 1: Revocation Request**

Client or administrator sends revocation request.

**Revocation Request**:

```json
{
  "token": "refresh_9z8y7x6w5v4u3t2s1r0q",
  "token_type_hint": "refresh_token",
  "client_id": "abc123xyz",
  "client_secret": "secret_98765zyxwvut"
}
```

**Step 2: Validate Request**

System validates revocation request:

1. Verify client credentials (client_id and client_secret)
2. Query database for token
3. Verify token exists and belongs to the authenticated client

**Step 3: Revoke Token**

Update token to revoked status:

```json
{
  "token": "refresh_9z8y7x6w5v4u3t2s1r0q",
  "is_revoked": true,
  "revoked_at": "2025-12-26T12:00:00Z"
}
```

**Step 4: Return Success Response**

**Revocation Response**:

```json
{
  "status": "revoked"
}
```

Note: Per RFC 7009, revocation endpoint should return 200 OK even if token doesn't exist to prevent information leakage.

---

### Security Validation Summary

**Multi-Tenant Isolation Checkpoints**:

1. Authorization endpoint: Validate user belongs to tenant from client
2. Token exchange: Verify tenant_id matches across authorization code, user, and client
3. Token validation: Verify tenant_id claim matches expected tenant context
4. Refresh token: Verify tenant_id matches throughout refresh process

**PKCE Validation Flow**:

1. Authorization request: Store code_challenge and method with authorization code
2. Token exchange: Compute SHA256(code_verifier), base64url encode, compare with stored code_challenge
3. Reject if mismatch: Prevents authorization code interception attacks

**Redirect URI Validation**:

1. Authorization request: Exact match against registered URIs
2. Token exchange: Verify redirect_uri matches the one used in authorization request
3. No wildcards or pattern matching: Prevents open redirect vulnerabilities

**Token Expiration Enforcement**:

1. Authorization codes: 5-minute lifetime, one-time use
2. Access tokens: Short-lived (15 minutes default)
3. Refresh tokens: Long-lived (30 days default) but revocable
4. All expirations validated server-side and in JWT claims
