# Feature Specification: Client Application Registration Portal

**Feature Branch**: `004-client-registration-app`  
**Created**: February 7, 2026  
**Status**: Draft  
**Input**: User description: "create registration app for client, you need to adjust this feature following feature sso-auth-provider that was created before"

## Clarifications

### Session 2026-02-07

- Q: When a client secret is regenerated, should the old secret remain valid for a grace period to allow zero-downtime migration, or be immediately invalidated? → A: Immediate invalidation (old secret stops working instantly, most secure, requires coordination)
- Q: When a tenant administrator exceeds the rate limit (20 registrations per hour), what should happen? → A: Block with error and reset time displayed (provides clear feedback on when they can retry)
- Q: How should the system determine which tenant an administrator is managing when they access the client registration portal? → A: Session-based tenant context from login (tenant_id automatically determined from authenticated user's session)
- Q: Should the system support permanent deletion of client applications, or only deactivation? → A: Deactivation only (soft delete) - clients remain in database for audit trails but cannot authenticate
- Q: What should "recent activity" on the client dashboard include, and what time window should be used? → A: Authentication requests (authorization/token endpoint) in last 7 days (provides operational visibility, balanced performance)

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Client Application Registration Form (Priority: P1)

A tenant administrator needs to register a new client application (web app, mobile app, or third-party integration) that will authenticate users through the SSO provider. They access the client registration portal, fill out application details including name and redirect URIs, and receive client credentials to integrate with their application.

**Why this priority**: This is the foundational capability that enables tenant administrators to onboard their applications to use the SSO service. Without this, applications cannot authenticate users. This directly supports FR-001 from the sso-auth-provider feature.

**Independent Test**: Can be fully tested by logging in as a tenant administrator, accessing the client registration form, creating a new client with valid details, and verifying that client_id and client_secret are generated and displayed. Delivers immediate value by allowing administrators to prepare their applications for SSO integration.

**Acceptance Scenarios**:

1. **Given** a tenant administrator is logged into the admin portal, **When** they navigate to "Register New Client Application", **Then** they are presented with a registration form requesting application name and at least one redirect URI.

2. **Given** the registration form is displayed, **When** the administrator enters a valid application name "Mobile App" and redirect URI "https://app.example.com/callback", **Then** the form accepts the input and enables the submit button.

3. **Given** valid application details are entered, **When** the administrator submits the form, **Then** the system generates a unique client_id and cryptographically secure client_secret, stores the application with status "active", and displays the credentials with a warning to copy them securely.

4. **Given** the administrator enters an invalid redirect URI (e.g., using HTTP instead of HTTPS), **When** they attempt to submit, **Then** the system rejects the submission with a clear validation error indicating HTTPS is required for production endpoints.

5. **Given** the administrator leaves the application name empty, **When** they attempt to submit, **Then** the system displays a validation error requiring the application name.

6. **Given** multiple redirect URIs are needed, **When** the administrator adds additional URIs (up to 10), **Then** each URI is validated independently and stored with the client application.

---

### User Story 2 - Client Credentials Display and Security (Priority: P1)

After successfully registering a client application, the administrator must securely receive and store the client_secret because it will only be displayed once. The system provides clear instructions and a copy button to help administrators securely capture the credentials.

**Why this priority**: Critical for security - the client_secret must be shown immediately but never stored in plaintext or displayed again. This prevents credential compromise while ensuring administrators can successfully complete integration.

**Independent Test**: Can be fully tested by completing client registration and verifying that the client_secret is displayed once with clear copy functionality, and that subsequent views of the client do not reveal the secret. Delivers security value by enforcing secure credential handling.

**Acceptance Scenarios**:

1. **Given** a client application has been successfully created, **When** the success page loads, **Then** the administrator sees both client_id and client_secret displayed prominently with a "Copy" button for each.

2. **Given** credentials are displayed, **When** the administrator sees the page, **Then** a clear warning message states "This is the only time the client secret will be displayed. Store it securely."

3. **Given** the administrator clicks the "Copy" button for client_secret, **When** the button is clicked, **Then** the secret is copied to clipboard and a confirmation message appears.

4. **Given** the administrator navigates away from the credentials page, **When** they return to view the client application details, **Then** the client_secret is masked (e.g., "••••••••") and cannot be revealed again.

5. **Given** the administrator has lost the client_secret, **When** they view the client details, **Then** they see an option to "Regenerate Client Secret" which creates a new secret and invalidates the old one.

---

### User Story 3 - Client Application List and Management (Priority: P2)

A tenant administrator needs to view all registered client applications for their organization, see their configuration details, and manage them (update redirect URIs, regenerate secrets, or deactivate applications).

**Why this priority**: Essential for ongoing management of multiple applications, but basic registration (P1) must work first. This provides operational capabilities for maintaining application configurations.

**Independent Test**: Can be fully tested by registering multiple clients, viewing the list, updating a redirect URI, and verifying changes are saved correctly. Delivers management value by enabling administrators to maintain their application portfolio.

**Acceptance Scenarios**:

1. **Given** a tenant administrator has registered multiple client applications, **When** they navigate to "My Applications", **Then** they see a list of all registered clients showing client_id, application name, number of redirect URIs, creation date, and status.

2. **Given** the administrator views the application list, **When** they click on a specific application, **Then** they see detailed configuration including all redirect URIs, creation date, last modified date, and status.

3. **Given** the administrator is viewing a client's details, **When** they update the redirect URIs (add, modify, or remove), **Then** the changes are validated and saved, and a confirmation message is displayed.

4. **Given** the administrator needs to disable an application, **When** they toggle the status to "inactive", **Then** the system confirms the action and marks the client as inactive, preventing new authorization requests.

5. **Given** an inactive client exists, **When** the administrator views it, **Then** they see the inactive status clearly indicated and have the option to reactivate it.

---

### User Story 4 - Redirect URI Validation and Management (Priority: P2)

A tenant administrator needs to add, update, or remove redirect URIs for their client applications as their application deployment evolves (e.g., adding staging environments, updating domain names, adding mobile deep links).

**Why this priority**: Important for maintaining security as applications evolve, but initial registration (P1) must support redirect URIs first. This enables flexible application lifecycle management.

**Independent Test**: Can be fully tested by creating a client, adding multiple redirect URIs with different schemes (https://, app://, localhost), and verifying validation rules are applied correctly. Delivers flexibility value for complex application deployments.

**Acceptance Scenarios**:

1. **Given** a client application is being configured, **When** the administrator adds a redirect URI "https://staging.example.com/callback" to an existing production URI, **Then** the system validates the format and stores both URIs.

2. **Given** the administrator is configuring redirect URIs, **When** they add a localhost URI "http://localhost:3000/callback" for development, **Then** the system allows HTTP for localhost addresses only.

3. **Given** the administrator adds a mobile deep link URI "myapp://callback", **When** they submit it, **Then** the system validates the custom scheme format and stores it alongside HTTPS URIs.

4. **Given** the administrator attempts to add more than 10 redirect URIs, **When** they submit the 11th URI, **Then** the system rejects it with an error indicating the maximum limit has been reached.

5. **Given** the administrator removes a redirect URI, **When** they confirm the removal, **Then** the system deletes the URI and displays a warning if any existing authorization flows might be affected.

---

### User Story 5 - Client Secret Regeneration (Priority: P3)

A tenant administrator needs to regenerate the client_secret for security reasons (credential rotation policy, suspected compromise, or employee turnover). The system generates a new secret and immediately invalidates the old one. The administrator must coordinate the credential update with their application deployment to avoid service interruption.

**Why this priority**: Important for security operations but not critical for initial application registration. Can be added after core registration and management functionality is stable.

**Independent Test**: Can be fully tested by regenerating a client secret, verifying the old secret no longer works for authentication, and confirming the new secret functions correctly. Delivers security value through credential rotation capabilities.

**Acceptance Scenarios**:

1. **Given** a tenant administrator views a client's details, **When** they click "Regenerate Client Secret", **Then** the system displays a confirmation dialog warning that the old secret will be invalidated.

2. **Given** the administrator confirms secret regeneration, **When** the new secret is generated, **Then** the system displays the new client_secret exactly once (same as initial registration) and immediately invalidates the old secret.

3. **Given** a new secret has been generated, **When** the client application attempts to authenticate using the old secret, **Then** the authentication fails with an "invalid_client" error.

4. **Given** a new secret has been generated, **When** the client application uses the new secret, **Then** authentication succeeds and tokens are issued normally.

---

### User Story 6 - Multi-Application Overview Dashboard (Priority: P3)

A tenant administrator managing multiple client applications needs an overview dashboard showing the health and usage statistics of all their registered applications (total active applications, recent authentication activity from the last 7 days, configuration status).

**Why this priority**: Provides valuable operational insights but basic registration and management (P1, P2) must work first. This adds observability to the client management experience.

**Independent Test**: Can be fully tested by registering multiple clients, performing authentication flows, and verifying that the dashboard displays accurate statistics. Delivers operational value through usage visibility.

**Acceptance Scenarios**:

1. **Given** a tenant administrator has multiple registered clients, **When** they access the applications dashboard, **Then** they see summary statistics including total active applications, total inactive applications, and applications created in the last 30 days.

2. **Given** the administrator views the dashboard, **When** authentication events (authorization requests or token exchanges) occur through their clients within the last 7 days, **Then** they see recent activity including which applications were used and when.

3. **Given** the administrator needs to identify issues, **When** they view the dashboard, **Then** they see any clients with configuration warnings (e.g., no active redirect URIs, secrets not rotated in 90+ days).

---

### Edge Cases

- What happens when a tenant administrator attempts to register a client with a redirect URI that conflicts with another tenant's registered URI?
- How does the system handle extremely long application names or redirect URIs?
- What happens when an administrator attempts to delete a client that has active user sessions or refresh tokens?
- How does the system prevent accidental client_secret exposure in browser developer tools or logs?
- What happens when redirect URIs contain query parameters or fragments?
- How does the system handle concurrent updates to the same client by multiple administrators?
- What happens when a client_id collision occurs during generation (extremely rare but theoretically possible)?
- How does the system validate custom URI schemes for mobile and desktop applications?
- What happens when an administrator tries to add duplicate redirect URIs to the same client?
- How does the system enforce tenant isolation to prevent one tenant from viewing another tenant's clients?

## Requirements _(mandatory)_

### Functional Requirements

#### Client Registration

- **FR-001**: System MUST provide an authenticated portal accessible only to tenant administrators for registering client applications.
- **FR-002**: System MUST collect client application information including application name and at least one redirect URI during registration.
- **FR-003**: System MUST generate a cryptographically secure, unique client_id for each registered application using UUID v4 or equivalent.
- **FR-004**: System MUST generate a cryptographically secure client_secret of at least 32 characters using a secure random number generator.
- **FR-005**: System MUST store client_secret using bcrypt hashing with cost factor 12 before persisting to database.
- **FR-006**: System MUST display the generated client_id and client_secret exactly once immediately after successful registration.
- **FR-007**: System MUST provide clear visual warnings that the client_secret will only be displayed once and must be stored securely.
- **FR-008**: System MUST provide copy-to-clipboard functionality for both client_id and client_secret.
- **FR-009**: System MUST enforce tenant isolation such that clients registered by Tenant A cannot be viewed or managed by Tenant B.
- **FR-010**: System MUST set initial client status to "active" upon successful registration.

#### Redirect URI Validation

- **FR-011**: System MUST validate that all redirect URIs use HTTPS protocol except for localhost addresses which may use HTTP.
- **FR-012**: System MUST validate redirect URI format according to RFC 3986 URI specification.
- **FR-013**: System MUST allow localhost URIs with HTTP scheme in the format "http://localhost:_" or "http://127.0.0.1:_" for development purposes.
- **FR-014**: System MUST support custom URI schemes for mobile and desktop applications (e.g., "myapp://callback").
- **FR-015**: System MUST allow a maximum of 10 redirect URIs per client application.
- **FR-016**: System MUST validate that redirect URIs do not exceed 2048 characters in length.
- **FR-017**: System MUST allow administrators to add, update, and remove redirect URIs for existing client applications.
- **FR-018**: System MUST prevent duplicate redirect URIs within the same client application.
- **FR-019**: System MUST preserve query parameters and fragments in redirect URIs if provided by the administrator.

#### Client Management

- **FR-020**: System MUST provide a list view of all client applications registered by the tenant administrator's organization.
- **FR-021**: System MUST display client_id, application name, creation date, status, and redirect URI count in the client list.
- **FR-022**: System MUST provide detailed view of individual client applications showing all configuration details.
- **FR-023**: System MUST mask client_secret with placeholder characters (e.g., "••••••••") in all views except the initial registration success page.
- **FR-024**: System MUST allow tenant administrators to update application name and redirect URIs for existing clients.
- **FR-025**: System MUST allow tenant administrators to toggle client status between active and inactive.
- **FR-026**: System MUST prevent authorization requests from inactive clients by validating client status during OAuth flows.
- **FR-027**: System MUST display last modified timestamp and last modified by user for each client application.
- **FR-028**: System MUST validate application names are between 1 and 100 characters in length.

#### Client Secret Management

- **FR-029**: System MUST provide functionality to regenerate client_secret for existing applications.
- **FR-030**: System MUST display a confirmation dialog before regenerating client_secret warning that the old secret will be invalidated.
- **FR-031**: System MUST invalidate the previous client_secret immediately upon regeneration of a new secret.
- **FR-032**: System MUST display the newly generated client_secret exactly once after regeneration with the same security warnings as initial registration.
- **FR-033**: System MUST log all client_secret regeneration events in the audit log with timestamp and administrator identity.
- **FR-034**: System MUST reject authentication attempts using invalidated client_secrets with an "invalid_client" error response.

#### Security & Validation

- **FR-035**: System MUST implement role-based access control to ensure only tenant administrators can access client registration functionality.
  - **FR-035a** (Clarification): The `admin_only` middleware MUST verify two conditions on every request: (1) authenticated user's role='admin' (or 'tenant_admin'), AND (2) user's assigned tenant_id matches the tenant context from their session. A user who is an admin for Tenant A MUST NOT access Tenant B's clients even if they have admin role.
- **FR-036**: System MUST validate that the authenticated user belongs to the tenant they are managing clients for by extracting tenant_id from the user's authenticated session and enforcing this context for all client management operations.
- **FR-037**: System MUST implement rate limiting on client registration to prevent abuse (maximum 20 client registrations per tenant per hour) and display a clear error message with the reset time when the limit is exceeded.
  - **FR-037a** (Clarification): Rate limit check MUST execute BEFORE expensive operations (URI validation, secret generation, database insert) to minimize resource consumption on abuse attempts. When limit exceeded, HTTP 429 response MUST include `Retry-After` header with epoch timestamp of when limit resets and human-readable message: "Rate limit exceeded. Next registration attempt available at [TIME]. Please try again in [DURATION] minutes."
- **FR-038**: System MUST sanitize all user input to prevent XSS attacks in application names and redirect URIs.
  - **FR-038a** (Clarification): Application names MUST have HTML entities escaped (e.g., `<script>` → `&lt;script&gt;`) before storage and on retrieval. All SQL queries MUST use parameterized queries (automatic in pgx driver) to prevent SQL injection. Redirect URIs are already validated by RFC 3986 parser which rejects malformed input.
- **FR-039**: System MUST log all client registration, modification, and deletion events in the audit log.
- **FR-040**: System MUST never display client_secret in application logs, error messages, or browser developer tools.
- **FR-041**: System MUST use constant-time comparison when validating client_secret to prevent timing attacks.
- **FR-042**: System MUST enforce HTTPS for all client management portal pages in production environments.
  - **FR-042a** (Clarification): In production environments (identified by config flag `ENVIRONMENT=production`), backend MUST redirect all HTTP requests to HTTPS with 301 status code. Frontend deployment MUST configure TLS certificates and HSTS headers (Strict-Transport-Security: max-age=31536000). Development environments MAY use HTTP for localhost testing. All sensitive operations MUST enforce HTTPS regardless of environment setting.

#### Dashboard & Monitoring

- **FR-043**: System MUST provide a dashboard showing total active clients, total inactive clients, recent registrations, and recent authentication activity (requests to authorization and token endpoints) from the last 7 days for the tenant.
- **FR-044**: System MUST display configuration warnings for clients that have potential security issues (e.g., only localhost redirect URIs in production tenants).
- **FR-045**: System MUST show the creation date and last modification date for each client application.

#### Integration with SSO Auth Provider

- **FR-046**: System MUST store client configurations in a format compatible with the Core SSO Auth Provider (feature 003-sso-auth-provider).
- **FR-047**: System MUST ensure registered clients are immediately available for use in OAuth/OIDC authorization flows.
- **FR-048**: System MUST validate that client_id used in authorization requests matches an active registered client.
- **FR-049**: System MUST validate that redirect_uri in authorization requests exactly matches one of the registered redirect URIs for the client.

### Key Entities

- **Client Application**: Represents a registered application that can authenticate users via the SSO provider. Attributes include: id (primary key), client_id (unique UUID), client_secret_hash (bcrypt hashed secret), application_name (human-readable name), tenant_id (owning tenant foreign key), status (active/inactive), created_at, updated_at, created_by_user_id, last_modified_by_user_id.

- **Client Redirect URI**: Represents an allowed callback URL for a client application. Attributes include: id (primary key), client_id (foreign key to Client Application), redirect_uri (validated URL), created_at. Relationship: One client application has many redirect URIs (1:N relationship, max 10 per client).

- **Client Secret History**: Tracks secret regeneration events for audit purposes. Attributes include: id (primary key), client_id (foreign key), secret_hash (bcrypt hashed), created_at, created_by_user_id, invalidated_at, reason (rotation_policy, suspected_compromise, manual_regeneration). Relationship: One client has many secret history records (1:N).

- **Client Audit Event**: Records all client management actions for security auditing. Attributes include: id (primary key), client_id (foreign key), tenant_id (foreign key), event_type (created, updated, secret_regenerated, status_changed, deleted), performed_by_user_id (administrator who performed action), timestamp, ip_address, details (JSON metadata).

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: Tenant administrators can successfully register a new client application and receive credentials in under 2 minutes from form access to credential display.
- **SC-002**: System generates cryptographically secure client_id and client_secret that meet industry security standards (minimum entropy equivalent to UUID v4 and 32-character random strings).
- **SC-003**: 100% of redirect URI validation rules are enforced preventing insecure configurations (no HTTP URIs except localhost).
- **SC-004**: Client credentials (client_secret) are never stored in plaintext and are only displayed once to administrators, with zero instances of secrets appearing in logs or subsequent views.
- **SC-005**: Tenant isolation is enforced with zero occurrences of cross-tenant data access in client management operations.
- **SC-006**: Administrators can successfully manage (view, update, activate/deactivate) all registered client applications with operation completion under 5 seconds per action.
- **SC-007**: All client registration and management actions are logged in audit trail with 100% traceability including timestamp and administrator identity.
- **SC-008**: Registered clients are immediately functional in OAuth/OIDC flows without requiring system restarts or cache invalidation.
- **SC-009**: Client secret regeneration completes successfully with old secrets invalidated and new secrets functional within 1 second of confirmation.
- **SC-010**: System maintains performance supporting up to 100 client applications per tenant without degradation in listing or search functionality.

## Assumptions

- Tenant administrators have already been authenticated and have appropriate permissions through the tenant registration and authentication system (features 001 and 003).
- The underlying SSO Auth Provider (feature 003) is available to consume client registration data via shared database or API.
- Email notifications for security events (secret regeneration, client creation) are handled by existing notification infrastructure.
- Tenant administrators understand OAuth/OIDC client credentials and their security implications (documentation/tooltips can provide guidance).
- Client applications being registered will implement proper OAuth/OIDC client protocols including secure storage of client_secret.
- The system supports industry-standard bcrypt hashing available in the application runtime environment.
- Database schema supports tenant_id partitioning for multi-tenant data isolation.
- Administrators access the portal through modern web browsers supporting standard HTML5, CSS3, and JavaScript features.

## Dependencies

- **Feature 001 (Tenant Registration)**: Required for tenant existence and administrator account management.
- **Feature 003 (SSO Auth Provider)**: Client registration portal provides the configuration data needed for OAuth/OIDC flows.
- Shared database or data access layer for storing client configurations accessible by both the registration portal and auth provider.
- Secure random number generation library for client_id and client_secret generation.
- Bcrypt hashing library for client_secret storage.
- Audit logging infrastructure for tracking client management events.
- Authentication middleware for verifying tenant administrator sessions.
- Frontend framework for rendering registration forms and client management interfaces.

## Out of Scope

- User authentication flows (handled by feature 003-sso-auth-provider).
- Tenant registration and onboarding (handled by feature 001-tenant-registration).
- Token issuance, validation, and refresh logic (handled by feature 003-sso-auth-provider).
- User role assignments to client applications (handled by feature 003-sso-auth-provider).
- Advanced client analytics and usage metrics beyond basic counts.
- Client application testing tools or OAuth/OIDC debugging utilities.
- Automated client secret rotation policies (manual regeneration only).
- Client deletion functionality (can be added in future iterations; deactivation is supported).
- Multi-factor authentication for client management portal access (assumed to be handled by tenant admin authentication).
- Client logo upload or branding customization.
- Webhook or API-based client registration (web portal UI only).
