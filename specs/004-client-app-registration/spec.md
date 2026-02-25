# Feature Specification: OAuth Client Application Registration Portal

**Feature Branch**: `004-client-app-registration`  
**Created**: February 25, 2026  
**Status**: Draft  
**Input**: User description: "create a feature for the client can register their app to use the SSO, so they can get clientId clientSecret and anything that they need for use oauth API on feature sso-auth-provider"

**Scope Note**: This feature provides the dedicated implementation of client application registration as described in 003-sso-auth-provider User Story 1 ("Client App Registration"). It extends that story with full lifecycle management (dashboard, updates, secret rotation, deletion). There is no separate registration portal; this is the single source of truth for OAuth client management within the SSO platform.

## Clarifications

### Session 2026-02-25

- Q: Should the registration portal support both confidential and public client types, or only confidential clients? → A: Both confidential and public client types (public clients get no secret, use PKCE only)
- Q: Who can register and manage OAuth client applications? → A: Only tenant administrators
- Q: What is the maximum number of client applications per tenant? → A: 25 clients per tenant
- Q: What happens to active tokens when a client is deleted? → A: Immediately revoke all tokens (access, refresh) issued to the deleted client
- Q: What is the scope of this feature relative to 003-sso-auth-provider? → A: This feature IS the implementation of 003's client registration (User Story 1); no separate portal

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Developer Registers New OAuth Client Application (Priority: P1)

A developer wants to integrate their application with the SSO system. They access the client registration portal, fill in their application details (name, description, redirect URIs), and submit the registration. Upon successful registration, the system generates and displays unique OAuth credentials (client_id and client_secret) that the developer copies for use in their application.

**Why this priority**: This is the foundational capability that enables any third-party application to integrate with the SSO system. Without this, developers cannot obtain the credentials needed to implement OAuth authentication flows.

**Independent Test**: Can be fully tested by accessing the registration portal, submitting valid application details, and verifying that client_id and client_secret are generated and displayed. Delivers immediate value by providing developers with the credentials needed to start OAuth integration.

**Acceptance Scenarios**:

1. **Given** a developer accesses the client registration portal, **When** they fill in application name "My Mobile App", description "iOS mobile application", and redirect URI "https://myapp.example.com/auth/callback", **Then** the system validates the inputs, creates a new client record, generates unique client_id and client_secret, and displays the credentials with a one-time warning to copy and securely store the secret.

2. **Given** a developer submits a registration form, **When** they provide an invalid redirect URI (e.g., uses HTTP instead of HTTPS in production, or invalid URL format), **Then** the system rejects the submission with a clear validation error message specifying the redirect URI requirements.

3. **Given** a developer completes registration, **When** the credentials are displayed, **Then** the client_secret is shown only once with a prominent warning that it cannot be retrieved again, and the developer must confirm they have copied it before proceeding.

4. **Given** a developer registers a client, **When** the registration is successful, **Then** the system provides documentation links and code examples showing how to use the credentials in OAuth2/OIDC flows.

---

### User Story 2 - Developer Views and Manages Their Registered Applications (Priority: P1)

A developer who has registered one or more applications needs to view their registered clients, check the configuration details, and verify their redirect URIs. They access their client dashboard and see a list of all their registered applications with key information.

**Why this priority**: Developers need visibility into their registered applications to verify configurations, troubleshoot integration issues, and manage multiple client applications. This is essential for day-to-day OAuth integration work.

**Independent Test**: Can be fully tested by registering multiple clients, then accessing the dashboard to view the list of registered applications with their client_ids, names, descriptions, and redirect URIs. Delivers value by providing developers with a reference for their OAuth credentials and configurations.

**Acceptance Scenarios**:

1. **Given** a developer has registered three applications, **When** they access their client dashboard, **Then** they see a list displaying all three applications with client_id, application name, creation date, and status.

2. **Given** a developer is viewing their client list, **When** they select a specific client, **Then** they see detailed information including client_id, application name, description, all registered redirect URIs, and creation/modification timestamps (but not the client_secret).

3. **Given** a developer views their client details, **When** they need to check which redirect URIs are configured, **Then** all redirect URIs are displayed in a clear list format, with HTTPS URIs distinguishable from localhost development URIs.

---

### User Story 3 - Developer Updates Client Application Configuration (Priority: P2)

A developer needs to modify their registered client application because they've added new redirect URIs for additional environments (staging, production) or need to update the application description. They access their client details, edit the configuration, and save the changes.

**Why this priority**: As applications evolve, developers need to update redirect URIs for different environments and maintain accurate application metadata. This is important for flexibility but not critical for initial integration.

**Independent Test**: Can be fully tested by modifying an existing client's redirect URIs and description, saving the changes, and verifying the updates are persisted. Delivers value by allowing developers to adapt their OAuth configurations without creating new clients.

**Acceptance Scenarios**:

1. **Given** a developer has a registered client with one redirect URI, **When** they edit the client and add two additional redirect URIs "https://staging.myapp.com/callback" and "https://myapp.com/callback", **Then** the system validates the new URIs, saves all three redirect URIs, and displays a confirmation.

2. **Given** a developer edits their client configuration, **When** they attempt to add an invalid redirect URI, **Then** the system prevents saving and displays specific validation errors without losing the other valid changes they made.

3. **Given** a developer updates their client description, **When** they save the changes, **Then** the system updates only the description field without affecting other client properties like client_id or existing redirect URIs.

---

### User Story 4 - Developer Regenerates Client Secret (Priority: P2)

A developer's client_secret has been compromised or they need to rotate it for security compliance. They access their client details, request a new secret generation, and the system generates a new client_secret while invalidating the old one.

**Why this priority**: Security best practices require the ability to rotate credentials when they're compromised or as part of regular security maintenance. Important for security but not required for initial integration.

**Independent Test**: Can be fully tested by regenerating a client_secret, verifying the new secret works with OAuth flows, and confirming the old secret is invalidated. Delivers value by enabling security incident response and compliance with credential rotation policies.

**Acceptance Scenarios**:

1. **Given** a developer suspects their client_secret is compromised, **When** they click "Regenerate Secret" in their client details and confirm the action, **Then** the system generates a new client_secret, immediately invalidates the old secret, and displays the new secret with a warning to copy it.

2. **Given** a developer regenerates their client_secret, **When** they attempt to use the old secret for OAuth token requests, **Then** the system rejects the requests with an "invalid_client" error.

3. **Given** a developer initiates secret regeneration, **When** the confirmation prompt appears, **Then** it clearly warns that the old secret will stop working immediately and all applications using it must be updated with the new secret.

---

### User Story 5 - Developer Deletes Client Application (Priority: P3)

A developer has deprecated an application and no longer needs the OAuth client registration. They access their client details, request deletion, and the system removes the client registration and invalidates all associated credentials.

**Why this priority**: Proper credential lifecycle management requires the ability to decommission unused clients. This is a hygiene feature that supports security best practices but isn't needed for initial integration or active development.

**Independent Test**: Can be fully tested by deleting a client, verifying it no longer appears in the dashboard, and confirming that credentials are invalidated. Delivers value by allowing developers to clean up unused registrations and maintain security hygiene.

**Acceptance Scenarios**:

1. **Given** a developer has a registered client they no longer use, **When** they click "Delete Client" and confirm the irreversible action, **Then** the system permanently deletes the client registration, invalidates the client_id and client_secret, and removes it from their dashboard.

2. **Given** a developer deletes a client, **When** applications attempt to use its credentials for OAuth flows, **Then** the system rejects all requests with an "invalid_client" error.

3. **Given** a developer initiates client deletion, **When** the confirmation prompt appears, **Then** it clearly warns that this action is permanent, will immediately break any applications using these credentials, and cannot be undone.

---

### User Story 6 - Developer Accesses OAuth Integration Documentation (Priority: P2)

A developer has registered their client application and needs guidance on implementing the OAuth2/OIDC flows. They access documentation directly from the registration portal that includes code examples using their actual client_id.

**Why this priority**: Reducing integration friction is important for developer adoption. Clear documentation with personalized examples significantly speeds up integration, but developers can also work from generic documentation if needed.

**Independent Test**: Can be fully tested by registering a client and verifying that documentation links are provided with code examples that include the developer's actual client_id. Delivers value by reducing integration time and support burden.

**Acceptance Scenarios**:

1. **Given** a developer has registered a client, **When** they view their client details, **Then** they see links to documentation covering authorization code flow, PKCE, token exchange, and token refresh.

2. **Given** a developer accesses the OAuth flow documentation, **When** they view code examples, **Then** the examples include their actual client_id in the sample requests, making it easy to adapt the code for testing.

3. **Given** a developer views the documentation, **When** they need to understand redirect URI requirements, **Then** they see clear explanations of URI validation rules, HTTPS requirements, and localhost exceptions for development.

---

### Edge Cases

- What happens when a developer attempts to register a client with duplicate redirect URIs?
- How does the system handle when a developer tries to add a redirect URI that's already registered by another client in the same tenant?
- What happens if a developer closes the browser immediately after registration before copying the client_secret?
- How does the system handle very long application names or descriptions?
- What happens when a developer attempts to register more clients than allowed per tenant?
- How does the system handle redirect URIs with special characters or non-standard ports?
- What happens when a developer tries to delete a client that's actively being used by many users? All active access tokens and refresh tokens issued through the client are immediately revoked; affected users must re-authenticate through another client.

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: System MUST provide a web-based client registration portal accessible only to authenticated tenant administrators.
- **FR-002**: System MUST validate application names to ensure they are between 3 and 100 characters and contain no special characters that could cause security or display issues.
- **FR-003**: System MUST validate redirect URIs to ensure they use HTTPS protocol in production environments, have valid URL format, and match expected patterns.
- **FR-004**: System MUST allow localhost and HTTP redirect URIs for development purposes in non-production environments.
- **FR-005**: System MUST generate cryptographically secure, unique client_id values for each registered client application regardless of client type.
- **FR-005a**: System MUST allow developers to select a client type during registration: confidential (server-side apps) or public (SPAs, mobile/native apps).
- **FR-005b**: System MUST generate a client_secret only for confidential client types. Public clients MUST NOT receive a client_secret and MUST use PKCE-only flows.
- **FR-006**: System MUST generate cryptographically secure, unique client_secret values with sufficient entropy to resist brute-force attacks for confidential clients.
- **FR-007**: System MUST display the client_secret only once immediately after generation with a clear warning that it cannot be retrieved again (confidential clients only).
- **FR-008**: System MUST store client_secret in hashed form using a secure one-way hashing algorithm, never storing it in plain text (confidential clients only).
- **FR-009**: System MUST associate each registered client with the tenant context of the user who created it, ensuring tenant isolation.
- **FR-010**: System MUST allow users to register multiple redirect URIs for a single client application to support different environments.
- **FR-011**: System MUST provide a dashboard view showing all client applications registered within the tenant, accessible only to tenant administrators.
- **FR-012**: System MUST display client_id, application name, description, creation date, and status in the client list view.
- **FR-013**: System MUST provide a detailed view for each client showing all configuration including redirect URIs but excluding the client_secret.
- **FR-014**: System MUST allow users to edit client application names, descriptions, and redirect URIs after registration.
- **FR-015**: System MUST validate all updates to redirect URIs using the same validation rules as initial registration.
- **FR-016**: System MUST provide a secret regeneration capability that generates a new client_secret and immediately invalidates the previous one.
- **FR-017**: System MUST require explicit user confirmation before regenerating a client_secret, warning about service disruption.
- **FR-018**: System MUST provide a client deletion capability that permanently removes the client registration, invalidates all credentials, and immediately revokes all active access tokens and refresh tokens that were issued through the deleted client.
- **FR-019**: System MUST require explicit user confirmation before deleting a client, warning that the action is irreversible.
- **FR-020**: System MUST enforce a maximum limit of 25 client applications per tenant to prevent abuse. When the limit is reached, new registration attempts MUST be rejected with a clear message indicating the quota has been reached.
- **FR-021**: System MUST log all client registration, modification, secret regeneration, and deletion events to an audit trail for security monitoring.
- **FR-022**: System MUST prevent registration of clients with duplicate redirect URIs within the same tenant when such duplicates could cause ambiguity.
- **FR-023**: System MUST provide clear, actionable error messages for all validation failures during registration and updates.
- **FR-024**: System MUST include links to OAuth2/OIDC integration documentation from the client dashboard and details pages.
- **FR-025**: System MUST validate that client_id and redirect_uri pairs used in OAuth flows match registered client configurations.
- **FR-026**: Users MUST be authenticated as a tenant administrator before accessing the client registration portal. Non-admin users MUST be denied access with a clear authorization error.
- **FR-027**: System MUST handle UTF-8 characters properly in application names and descriptions to support internationalization.
- **FR-028**: System MUST provide search and filtering capabilities in the client list when users have many registered applications.
- **FR-029**: System MUST indicate the last modification date and time for each client application.
- **FR-030**: System MUST support pagination in the client list view for tenants with large numbers of registered clients.

### Key Entities

- **Client Application**: Represents a third-party application registered to use the SSO system for authentication. Key attributes include unique identifier (client_id), client type (confidential or public), hashed secret (client_secret_hash, confidential clients only), application name, description, list of allowed redirect URIs, tenant association, creation timestamp, last modified timestamp, and status (active, revoked). Confidential clients receive a client_secret for server-side authentication; public clients (SPAs, mobile/native apps) use PKCE-only flows without a secret. Each client belongs to exactly one tenant and can have multiple redirect URIs.

- **Redirect URI**: Represents an allowed OAuth callback URL for a registered client. Key attributes include the URI value and association with a client application. URIs must use HTTPS in production; HTTP is allowed for localhost/127.0.0.1 (development). Multiple redirect URIs can belong to one client application.

- **Client Registration Event**: Represents an audit log entry for client lifecycle events. Key attributes include event type (created, updated, secret_regenerated, deleted), timestamp, acting user, client application reference, tenant context, and relevant details (e.g., which fields were modified). Used for security monitoring and compliance.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: Developers can complete client application registration from portal access to receiving credentials in under 2 minutes for a straightforward registration.
- **SC-002**: 95% of client registration attempts with valid inputs succeed on the first try without errors.
- **SC-003**: All generated client_id values are unique across the entire platform with zero collisions.
- **SC-004**: All generated client_secret values meet entropy requirements of at least 256 bits to ensure cryptographic security.
- **SC-005**: Client credentials (client_id and client_secret pairs) are successfully usable in OAuth2 authorization flows immediately after registration without additional configuration or delays.
- **SC-006**: Developers with 10 or more registered clients can locate a specific client using search or filtering in under 10 seconds.
- **SC-007**: 90% of developers successfully copy and save their client_secret on first display without requesting support to regenerate it.
- **SC-008**: Redirect URI validation catches 100% of common security misconfigurations including HTTP in production and malformed URLs.
- **SC-009**: Client configuration updates (adding redirect URIs, changing descriptions) are reflected immediately in subsequent OAuth authorization requests.
- **SC-010**: Secret regeneration invalidates old credentials within 1 second, ensuring compromised secrets cannot be used for new OAuth flows.
- **SC-011**: The client registration portal supports at least 100 concurrent users per tenant without performance degradation.
- **SC-012**: Developer error rate when implementing OAuth2 flows decreases by 40% when using personalized documentation with their actual client_id compared to generic documentation.
- **SC-013**: All client registration, modification, and deletion events are captured in audit logs with 100% accuracy for security compliance.
- **SC-014**: The portal displays clear, actionable error messages for 100% of validation failures, specifying exactly what needs to be corrected.
- **SC-015**: Tenant isolation is maintained with 100% accuracy—users can only view and manage clients within their own tenant.
- **SC-016**: Zero client_secret values are ever logged, displayed in UI, or transmitted after the initial one-time display.
- **SC-017**: System handles UTF-8 characters in application names and descriptions without corruption or display issues for international users.
- **SC-018**: Client deletion prevents all subsequent OAuth flows using deleted credentials with 100% effectiveness within 1 second.
- **SC-019**: Time from developer registration to first successful OAuth authentication flow completion averages under 15 minutes including reading documentation and implementing code.
