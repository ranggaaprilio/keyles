# Feature Specification: End-User Management with RBAC

**Feature Branch**: `005-user-management-rbac`  
**Created**: June 11, 2025  
**Status**: Draft  
**Input**: User description: "Build an end-user management system with RBAC so that tenant administrators can manage the users who authenticate through their SSO. This includes: inviting users to the tenant, assigning/revoking roles per client application, viewing user activity/sessions, and enabling/disabling users. The role system should integrate with the OAuth token flow so that user roles are included in JWT claims."

**Scope Note**: This feature provides the full lifecycle management of end-users within a tenant, extending the authentication infrastructure introduced in `003-sso-auth-provider`. The `003` feature established the foundational `users` and `user_role_assignments` tables; this feature builds the administrative surface area on top of that foundation—invitation flows, role management UI, session visibility, and JWT claim integration—so that tenant administrators have full operational control over their user population.

---

## Clarifications

### Session 2025-06-11

- Q: How are roles defined—are they a fixed platform-level set (e.g., `admin`, `viewer`) or can tenant admins create custom role names per client application? → A: Roles are free-form strings scoped to a `(tenant, client_app)` pair; there is no platform-enforced enumeration. Tenant admins define and assign whatever role names their applications consume.
- Q: What is the invitation flow—does an invited user set a password immediately, or do they authenticate via a time-limited magic link / OTP? → A: Invitation sends a time-limited email link (valid 72 hours). Clicking it opens a password-set form; the user sets their password and their account becomes active. Unaccepted invitations expire and can be resent.
- Q: Should session/activity visibility be read-only for administrators, or should admins be able to remotely terminate individual sessions? → A: Admins can both view active sessions and forcibly terminate (revoke) individual sessions for any user in their tenant.
- Q: When a user is disabled, should their existing active sessions be immediately terminated? → A: Yes. Disabling a user must immediately revoke all active sessions and refresh tokens for that user.
- Q: What is the maximum number of users per tenant? → A: 10,000 users per tenant.
- Q: Can roles be assigned at the tenant level (global) in addition to per-client? → A: Roles are exclusively per-client-application. There is no tenant-wide role; access to a client is determined by having at least one active role assignment for that client.

---

## User Scenarios & Testing _(mandatory)_

### User Story 1 — Administrator Invites a New User to the Tenant (Priority: P1)

A tenant administrator needs to onboard a new colleague or customer. They open the user management section of the admin portal, enter the new user's email address (and optionally a display name), and send an invitation. The invited person receives an email containing a secure, time-limited link. They click the link, set a password, and their account is immediately active and ready to authenticate via SSO.

**Why this priority**: Every tenant begins with zero end-users. Without a way to add users, no one can authenticate through the SSO, and all downstream features (role assignment, session management) have no subjects to operate on. This story is the entry point for the entire user lifecycle.

**Independent Test**: Can be fully tested by sending an invitation from the admin portal, verifying the email is dispatched, clicking the link within the 72-hour window, completing the password-set form, and then confirming the user can successfully authenticate through an OAuth authorization code flow. Delivers immediate value by populating the tenant with authenticatable users.

**Acceptance Scenarios**:

1. **Given** a tenant administrator is logged in, **When** they navigate to User Management, enter `alice@example.com` as the email and `Alice Smith` as the display name, and click "Send Invitation", **Then** the system creates a pending user record, generates a unique invitation token with a 72-hour expiry, dispatches an invitation email to `alice@example.com` containing an activation link, and displays a confirmation message to the administrator.

2. **Given** the invitation email has been sent, **When** Alice clicks the activation link within 72 hours, **Then** she is presented with a password-creation form pre-filled with her email address, and after submitting a valid password, her account status changes from `pending` to `active`, and she is shown a success screen.

3. **Given** an invitation link that has expired (older than 72 hours), **When** Alice clicks it, **Then** she sees an informative error page explaining the link has expired and providing instructions to request a new invitation from her administrator.

4. **Given** an administrator attempts to invite an email address that already has an active user account within the same tenant, **When** they submit the invitation, **Then** the system rejects it with a clear message: "A user with this email already exists in this tenant."

5. **Given** an administrator invites a user but the invitation is never accepted, **When** the administrator views the user list, **Then** the pending invitation is shown with its expiry timestamp and a "Resend Invitation" action available.

6. **Given** an administrator resends an invitation for a pending user, **When** they confirm the action, **Then** the original invitation token is invalidated, a new token with a fresh 72-hour window is generated, and a new invitation email is dispatched.

---

### User Story 2 — Administrator Assigns and Revokes Roles for a User per Client Application (Priority: P1)

A tenant administrator needs to control which users can access which client applications through SSO, and what roles they carry within those applications. After a user account is active, the administrator opens that user's profile, selects a client application from the dropdown, types in a role name (e.g., `editor`, `billing-viewer`, `super-admin`), and saves. The role assignment is immediately reflected in the JWT access tokens issued to that user for that client.

**Why this priority**: Role assignment is the mechanism by which the RBAC system delivers its core value. Without it, all users would have identical, undifferentiated access to all clients. This story enables fine-grained per-application authorization and directly satisfies the OAuth token claim integration requirement.

**Independent Test**: Can be fully tested by assigning the role `editor` to a user for a specific client app, triggering a fresh OAuth authorization code flow for that user and client, decoding the resulting access token JWT, and verifying the `roles` claim contains `["editor"]`. Delivers immediate value by enabling resource servers to enforce role-based authorization decisions.

**Acceptance Scenarios**:

1. **Given** a tenant administrator is viewing an active user's profile, **When** they select client application "Reporting Dashboard", type role name `analyst`, and click "Assign Role", **Then** the system creates a role assignment record linking the user, the client, and the role; displays the new assignment in the user's role list; and logs the action to the audit trail.

2. **Given** a user has the role `analyst` assigned for "Reporting Dashboard", **When** that user completes a new OAuth authorization code flow for "Reporting Dashboard", **Then** the access token JWT contains a `roles` claim with the value `["analyst"]`, and the ID token contains a `roles` claim scoped to that client.

3. **Given** a user has multiple roles (`analyst`, `report-exporter`) assigned for the same client, **When** that user authenticates, **Then** the `roles` claim in the JWT contains all active roles as an array: `["analyst", "report-exporter"]`.

4. **Given** an administrator revokes the `analyst` role from a user, **When** the revocation is saved, **Then** the role assignment is marked inactive, subsequent OAuth token requests for that client no longer include `analyst` in the `roles` claim, and the revocation is recorded in the audit trail with the acting administrator's identity.

5. **Given** a user with no role assignments for a particular client application, **When** that user attempts to authenticate through that client's OAuth flow, **Then** the authorization is denied at the consent/authorization step and the user sees an access-denied message.

6. **Given** an administrator attempts to assign a duplicate role (same user, same client, same role name), **When** they submit, **Then** the system rejects it with the message: "This role is already assigned to this user for the selected application."

---

### User Story 3 — Administrator Views User List and Searches/Filters Users (Priority: P1)

A tenant administrator needs situational awareness of their user population. They open the User Management section and see a paginated list of all users in their tenant, with key attributes visible at a glance (name, email, status, last login, number of role assignments). They can search by name or email and filter by status (active, pending, disabled).

**Why this priority**: Without visibility into the user population, administrators cannot perform any management actions—they cannot find the user they need to invite, assign roles to, or disable. A functional user list is the navigation hub for all other user management stories.

**Independent Test**: Can be fully tested independently by populating a tenant with 20+ users in various states and verifying the list view shows correct data, pagination works, and search/filter returns accurate results. Delivers value as a standalone administrative dashboard.

**Acceptance Scenarios**:

1. **Given** a tenant has 150 active users and 10 pending invitations, **When** an administrator opens User Management, **Then** they see a paginated list (default 25 per page) showing each user's display name, email, status badge (Active / Pending / Disabled), last login timestamp, and total number of role assignments across all clients.

2. **Given** an administrator is on the user list page, **When** they type `alice` into the search box, **Then** the list immediately filters to show only users whose name or email contains `alice` (case-insensitive), and the result count updates accordingly.

3. **Given** an administrator is on the user list page, **When** they apply the "Pending" status filter, **Then** only users with `pending` invitation status are shown, and the count reflects the filtered set.

4. **Given** the user list has more entries than fit on one page, **When** the administrator clicks "Next Page", **Then** the next set of users is shown in the same sorted order (default: most recently created first), and pagination controls correctly reflect the current page and total pages.

5. **Given** an administrator views a user entry in the list, **When** they click the user's name, **Then** they are taken to that user's detailed profile page showing all role assignments per client application and full activity history.

---

### User Story 4 — Administrator Views User Activity and Active Sessions (Priority: P2)

A tenant administrator needs to audit user behavior and investigate security incidents. They open a user's profile and see a list of that user's recent authentication events (logins, token refreshes, failed attempts) and a separate list of currently active sessions (refresh tokens), including the client application, approximate location (if available), and last activity time. The administrator can terminate individual sessions remotely.

**Why this priority**: Session visibility and remote termination are critical security controls. While they don't block basic authentication from working, they are essential for incident response (e.g., a compromised account) and compliance. Placed at P2 because the core user/role CRUD must exist before session management adds value.

**Independent Test**: Can be fully tested by authenticating a user multiple times across different clients, opening their admin profile, verifying the activity log entries appear with correct timestamps and client context, and confirming that clicking "Terminate Session" on an active session revokes the associated refresh token and prevents further token refresh.

**Acceptance Scenarios**:

1. **Given** a user has authenticated five times over the past week, **When** an administrator opens that user's Activity tab, **Then** they see a chronological list of events including: event type (login success, login failure, token refresh, logout), timestamp, client application name, and approximate IP address or geolocation.

2. **Given** a user has two active sessions (refresh tokens) across two different client applications, **When** an administrator views the Active Sessions tab, **Then** they see both sessions listed with: client application name, session creation time, last activity time, and a "Terminate" button for each.

3. **Given** an administrator clicks "Terminate" on one of the active sessions and confirms, **When** the action is processed, **Then** the corresponding refresh token is immediately revoked, subsequent attempts to use that refresh token return an error to the client application, and the terminated session is removed from the Active Sessions list.

4. **Given** a user has no current active sessions, **When** an administrator views the Active Sessions tab, **Then** an empty state message is shown: "No active sessions for this user."

5. **Given** an administrator views the Activity tab, **When** there are more than 50 events, **Then** the events are paginated with 25 per page, and the administrator can navigate between pages.

---

### User Story 5 — Administrator Enables and Disables a User Account (Priority: P2)

A tenant administrator needs to temporarily block a user's access—for example, when an employee is on leave, under investigation, or their account has been compromised. They open the user's profile and toggle the account status from Active to Disabled. Immediately, all of that user's active sessions are terminated and any in-progress OAuth flows fail. Re-enabling the account restores access without requiring a new invitation or password reset.

**Why this priority**: Account suspension is a fundamental security control. It is more targeted than deletion (the account and its history are preserved) and can be reversed. Placed at P2 because it is important for security but does not block the core authentication path for the majority of normal users.

**Independent Test**: Can be fully tested by authenticating a user to obtain a refresh token, disabling the account via the admin UI, attempting to use the refresh token (must fail), re-enabling the account, and confirming the user can authenticate again. Delivers a self-contained security control test.

**Acceptance Scenarios**:

1. **Given** a tenant administrator is viewing an active user's profile, **When** they click "Disable Account" and confirm the action, **Then** the user's status changes to `disabled`, all active refresh tokens for that user are immediately revoked, and the audit log records the disabling event with the administrator's identity and timestamp.

2. **Given** a user has been disabled, **When** they attempt to initiate a new OAuth authorization code flow, **Then** the SSO login page returns an error: "Your account has been disabled. Please contact your administrator." and no authorization code is issued.

3. **Given** a user has been disabled and has an active refresh token, **When** a client application attempts to use that refresh token, **Then** the token endpoint returns an `invalid_grant` error.

4. **Given** a disabled user's account, **When** an administrator clicks "Enable Account", **Then** the user's status reverts to `active`, and the user can immediately begin a new OAuth authentication flow (their previously revoked sessions are not restored, but new sessions can be created).

5. **Given** an administrator attempts to disable another tenant administrator's account, **When** the action is submitted, **Then** the system rejects it with: "You cannot disable an administrator account." (Tenant admins cannot lock each other out; only platform-level intervention can disable an admin.)

---

### User Story 6 — Administrator Removes a User from the Tenant (Priority: P3)

A tenant administrator needs to permanently remove a user who has left the organization or whose account should be fully decommissioned. They open the user's profile, select "Remove User", confirm the irreversible action, and the user's account—including all role assignments and session history—is deleted from the tenant.

**Why this priority**: Permanent user removal is a data hygiene and compliance feature (e.g., right-to-erasure requests). It is placed at P3 because disabling covers most operational needs; deletion is only necessary for data-minimization compliance or permanent offboarding and is less frequently needed.

**Independent Test**: Can be fully tested by creating a user with role assignments and an active session, deleting the user, and verifying: the user no longer appears in the list, their credentials are rejected at the login page, role assignments are gone, and refresh tokens are revoked.

**Acceptance Scenarios**:

1. **Given** an administrator opens a user's profile and clicks "Remove User", **When** the confirmation dialog is displayed, **Then** it clearly warns: "This will permanently delete [user email] and all their role assignments. This action cannot be undone."

2. **Given** an administrator confirms user removal, **When** the deletion is processed, **Then** the user record is permanently deleted, all their role assignments are removed, all active refresh tokens are revoked, and the deletion event is written to the audit log.

3. **Given** a removed user attempts to authenticate, **When** they enter their credentials on the SSO login page, **Then** the system returns: "No account found for this email address in this tenant."

4. **Given** a removed user had existing access tokens that have not yet expired, **When** a resource server validates such a token, **Then** the token is considered invalid (the user no longer exists). The system adds a short-lived blacklist entry to ensure immediate invalidation without waiting for token expiry.

5. **Given** an administrator attempts to remove their own account, **When** they submit the deletion, **Then** the system rejects it with: "You cannot remove your own account."

---

### User Story 7 — JWT Access Token Includes Role Claims from RBAC (Priority: P1)

A resource server (API) behind the SSO needs to make authorization decisions based on what roles a user holds within a specific client application. When the SSO issues an access token for a user authenticating via a particular client, the JWT must include a `roles` claim listing all active role assignments for that `(user, client)` pair. Resource servers can then enforce access control by reading the claim—without making additional API calls to the SSO.

**Why this priority**: The JWT claim injection is the direct integration point between the RBAC data model and the live OAuth token flow. It is what makes roles actionable for resource servers. Placed P1 because without it, role assignments in the admin portal have no runtime effect.

**Independent Test**: Can be fully tested without a UI by directly seeding role assignments in the database, completing an OAuth authorization code flow, and inspecting the decoded access token for the `roles` claim. No admin UI is needed for this test.

**Acceptance Scenarios**:

1. **Given** a user has roles `["viewer", "report-exporter"]` assigned for client app `reporting-dashboard`, **When** an access token is issued for that user authenticating via `reporting-dashboard`, **Then** the JWT payload contains `"roles": ["viewer", "report-exporter"]` and `"client_id": "reporting-dashboard"`.

2. **Given** a user has roles for multiple clients (`analyst` for `reporting-dashboard`, `deployer` for `ci-platform`), **When** an access token is issued for `reporting-dashboard`, **Then** the `roles` claim contains only `["analyst"]`, not the roles from `ci-platform`.

3. **Given** a user has no role assignments for a client, **When** the SSO authorization endpoint receives an authorization request for that client from that user, **Then** access is denied before a token is ever issued; the user sees an "Access Denied" message on the SSO consent page.

4. **Given** a role is revoked by an administrator while the user has a valid access token, **When** the access token is next used at a resource server, **Then** the resource server uses the roles embedded in the JWT until the token expires. New tokens issued after the revocation will not include the revoked role. (Access tokens are short-lived; the effective propagation delay equals the access token lifetime.)

5. **Given** a user authenticates and an ID token is issued, **When** a client application decodes the ID token, **Then** the ID token also contains a `roles` claim matching the user's active roles for that client, enabling front-end applications to adapt their UI based on the user's roles.

---

### Edge Cases

- What happens when an invitation email cannot be delivered (bounce or invalid address)? The invitation record remains in `pending` state; the administrator sees an email delivery warning in the portal and can retry or use a different address.
- What happens when an administrator clicks an invitation link that has already been used (replay attack)? The token is single-use; a second click returns a "This link has already been used" error and no action is taken.
- What happens if an administrator assigns a role to a `pending` (uninvited, unaccepted) user? The system allows it. The role is stored and will be included in JWTs once the user activates their account.
- What happens if a role assignment is created for a client that has since been deleted? The orphaned role assignment is cleaned up by a cascade delete on the `clients` table; the JWT issued for any remaining active client will not contain that assignment.
- What happens when a tenant reaches the 10,000 user limit and an administrator tries to send another invitation? The invitation is blocked with a clear quota-exceeded error before any email is sent.
- What happens when the same user email is invited by two administrators simultaneously (race condition)? A unique constraint on `(tenant_id, email)` in the users table ensures only one invitation is created; the second attempt receives an appropriate conflict error.
- What happens when an administrator views sessions for a user who is currently mid-authentication (authorization code not yet exchanged)? Authorization codes are ephemeral (Redis, 5-minute TTL); they are not shown as "sessions". Only issued refresh tokens (representing established sessions) are displayed.
- What happens when a role name contains Unicode characters or unusually long strings? Role names are validated to be 1–100 characters; Unicode is permitted. Excessively long or empty role names are rejected with a validation error.
- What happens when an administrator attempts to remove a user with a very large number of role assignments (hundreds)? All role assignments are deleted in the same transaction as the user deletion; the operation is atomic.
- What happens when a disabled user's client application tries to silently refresh their session using a refresh token that was issued before the account was disabled? The refresh token is revoked on account disable; the token endpoint returns `invalid_grant`.

---

## Requirements _(mandatory)_

### Functional Requirements

#### User Invitation & Account Lifecycle

- **FR-001**: System MUST allow tenant administrators to invite new users by providing an email address and optional display name via the admin portal.
- **FR-002**: System MUST validate invited email addresses for proper format and reject invitations for emails that are already associated with an active or pending user in the same tenant.
- **FR-003**: System MUST generate a cryptographically secure, unique, single-use invitation token per invitation and dispatch an activation email containing the time-limited link.
- **FR-004**: Invitation links MUST expire after 72 hours from the time of creation. Expired links MUST display a clear expiry message and direct the user to request a new invitation.
- **FR-005**: System MUST provide a password-creation form when an invited user follows a valid activation link, accepting the email as read-only and requiring a new password meeting minimum strength requirements.
- **FR-006**: Upon successful password creation, the user account status MUST transition from `pending` to `active`, and the invitation token MUST be immediately invalidated so it cannot be reused.
- **FR-007**: System MUST allow administrators to resend an invitation for any user in `pending` status; resending MUST invalidate the previous token and generate a new one with a fresh 72-hour window.
- **FR-008**: System MUST display pending invitations in the user list with their expiry timestamp and a "Resend Invitation" action.
- **FR-009**: System MUST enforce a maximum limit of 10,000 users per tenant (across all statuses). Invitation attempts that would exceed this limit MUST be rejected with a quota-exceeded error before sending any email.

#### User Listing & Search

- **FR-010**: System MUST provide a paginated user list (default 25 users per page, maximum 100 per page) accessible only to tenant administrators, showing display name, email, status, last login timestamp, and total active role assignment count.
- **FR-011**: System MUST support case-insensitive search by display name and email address within the user list, returning results in real time or near real time.
- **FR-012**: System MUST support filtering the user list by account status: All, Active, Pending, or Disabled.
- **FR-013**: System MUST provide a user detail view accessible from the user list, showing full profile, all role assignments grouped by client application, and a tab for activity/session history.

#### Role Management (RBAC)

- **FR-014**: System MUST allow tenant administrators to assign one or more named roles to any active or pending user, scoped to a specific client application registered within the same tenant.
- **FR-015**: Role names MUST be free-form strings of 1–100 characters (Unicode permitted). The platform MUST NOT restrict the set of valid role names beyond this length limit.
- **FR-016**: System MUST prevent duplicate role assignments: the combination of `(user_id, client_id, role_name)` within a tenant MUST be unique. Attempting to assign an already-active duplicate MUST return a clear error.
- **FR-017**: System MUST allow administrators to revoke an individual role assignment. Revocation MUST mark the assignment as inactive (soft delete) and be immediately effective for new token issuances.
- **FR-018**: System MUST display all current role assignments for a user grouped by client application on the user detail page, showing role name, assignment date, and assigning administrator.
- **FR-019**: System MUST prevent role assignments that reference a client application from a different tenant.
- **FR-020**: System MUST log every role assignment and revocation event to the audit trail, capturing: acting administrator, affected user, client application, role name, and timestamp.

#### JWT Claim Integration

- **FR-021**: When the SSO authorization endpoint issues an authorization code for a user-client pair, it MUST verify that the user has at least one active role assignment for that client. Users with no active role assignments for the requested client MUST be denied access with an "access_denied" error.
- **FR-022**: The access token JWT issued for a user-client pair MUST contain a `roles` claim (array of strings) listing all active role names assigned to the user for that specific client application at the time of token issuance.
- **FR-023**: The `roles` claim MUST be scoped to the authenticating client only. Roles assigned for other client applications MUST NOT appear in the token.
- **FR-024**: The ID token MUST also contain a `roles` claim with the same value as the access token `roles` claim, enabling front-end applications to adapt their UI without calling the userinfo endpoint.
- **FR-025**: The userinfo endpoint MUST return a `roles` field containing the user's active role names for the client identified by the access token used in the request.

#### Account Enable / Disable

- **FR-026**: System MUST allow tenant administrators to disable any non-administrator active user account. Disabling MUST be confirmed via an explicit confirmation prompt.
- **FR-027**: When a user account is disabled, the system MUST immediately revoke all active refresh tokens associated with that user, regardless of which client application issued them.
- **FR-028**: A disabled user MUST NOT be able to initiate a new OAuth authorization flow. Attempts to authenticate MUST return a user-friendly error on the SSO login page indicating the account is disabled.
- **FR-029**: System MUST allow administrators to re-enable a disabled user account. Re-enabling MUST restore the ability to authenticate without requiring a new invitation or password reset.
- **FR-030**: System MUST prevent administrators from disabling another tenant administrator's account, protecting against accidental lock-out scenarios.

#### Session Management

- **FR-031**: System MUST display all active sessions (issued, non-revoked, non-expired refresh tokens) for a user in a Sessions tab on the user detail page, including: client application name, session creation time, last activity time, and an approximate origin (e.g., masked IP or country, if available).
- **FR-032**: System MUST allow administrators to terminate (revoke) any individual active session for a user within their tenant, with a confirmation prompt.
- **FR-033**: System MUST provide a paginated activity log for each user showing authentication events: login success, login failure, token refresh, logout, and session termination. Events MUST include timestamp, client application, and event type.
- **FR-034**: Activity log entries MUST be retained for a minimum of 90 days and displayed in reverse chronological order.

#### User Deletion

- **FR-035**: System MUST allow tenant administrators to permanently delete a user from the tenant after explicit confirmation.
- **FR-036**: User deletion MUST cascade: all role assignments, active sessions (refresh tokens), and pending invitations for that user MUST be atomically removed in the same operation.
- **FR-037**: Deleted users MUST be added to a short-lived access-token blacklist (keyed by user ID) so that any access tokens issued before deletion are immediately invalidated without waiting for natural expiry.
- **FR-038**: System MUST prevent administrators from deleting their own account.
- **FR-039**: User deletion events MUST be written to the audit log with the acting administrator's identity, the deleted user's email and ID, and a timestamp.

#### Security & Audit

- **FR-040**: All user management operations (invitation, role assignment, role revocation, enable, disable, delete, session termination) MUST be recorded in the audit log with: event type, timestamp, acting administrator ID, and affected resource (user ID or session ID).
- **FR-041**: All user management API endpoints MUST enforce that the authenticated caller is a tenant administrator within the same tenant as the resource being acted upon. Cross-tenant access MUST be denied with a 403 error.
- **FR-042**: Invitation token validation MUST be resistant to timing attacks (constant-time comparison).

### Key Entities

- **User**: Represents an end-user who can authenticate through the tenant's SSO. Key attributes: unique identifier, tenant association, email address (unique per tenant), display name, hashed password, account status (`pending` / `active` / `disabled`), creation timestamp, last login timestamp. A user belongs to exactly one tenant and can have many role assignments across many client applications.

- **Invitation**: Represents a pending invitation sent to a prospective user. Key attributes: unique identifier, tenant association, target email address, invitation token (hashed), expiry timestamp (72 hours from creation), status (`pending` / `accepted` / `expired`), sending administrator ID, creation timestamp. An invitation transitions to `accepted` when the user completes password setup, or `expired` if the deadline passes.

- **Role Assignment**: Represents the grant of a named role to a user for a specific client application within a tenant. Key attributes: unique identifier, user ID, client application ID, tenant ID, role name (free-form string), active/inactive status, grant timestamp, and granting administrator ID. The combination `(user_id, client_id, role_name)` is unique. Role assignments are soft-deleted on revocation to preserve audit history.

- **Session (Refresh Token)**: Represents an established, long-lived authentication session for a user-client pair. An existing entity from `003-sso-auth-provider`, surfaced here for administrators. Key attributes: token hash, user ID, client ID, tenant ID, creation time, expiry time, last-used time, revocation status. One user can have multiple concurrent sessions across different or same client applications.

- **User Activity Event**: Represents a single auditable user authentication event. Key attributes: event type (login_success, login_failure, token_refresh, logout, session_terminated, account_disabled, account_enabled, user_deleted), timestamp, user ID, client application ID, IP address or origin (if available), and event details. Retained for 90 days minimum.

---

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: Tenant administrators can complete the full user invitation flow—from entering an email to the invited user having an active, authenticatable account—in under 5 minutes (including the user's password setup step).
- **SC-002**: 95% of invitation emails are delivered and result in a successful account activation within the 72-hour window, as measured in staging environments with real mail delivery.
- **SC-003**: Role assignment changes (both grant and revoke) are reflected in newly issued JWT tokens within 1 second of the administrator saving the change.
- **SC-004**: Disabling a user account terminates all active sessions within 1 second of the administrator confirming the action, preventing any further use of existing refresh tokens.
- **SC-005**: Deleted user access tokens are blocked at resource servers within 1 second of deletion (via blacklist propagation), regardless of the token's remaining TTL.
- **SC-006**: User list with 10,000 entries loads within 2 seconds and supports pagination, search, and filter operations that return results within 1 second.
- **SC-007**: All 19 user management event types (invitation sent, accepted, expired, resent; role assigned, revoked; account disabled, enabled, deleted; session terminated; etc.) are captured in the audit log with 100% accuracy.
- **SC-008**: JWT `roles` claims contain only the roles for the authenticating client; cross-client role leakage is zero across all test scenarios.
- **SC-009**: Users with zero active role assignments for a client are denied authorization at the SSO layer with 100% consistency; no access tokens are ever issued to unauthorized users.
- **SC-010**: Tenant isolation is maintained with 100% accuracy: administrators can only view, modify, and act upon users within their own tenant.
- **SC-011**: Zero invitation tokens can be reused after acceptance or expiry; each token is single-use and immediately invalidated on first use.
- **SC-012**: 90% of tenant administrators can successfully complete common tasks (invite user, assign role, disable account) without consulting documentation, as measured by task completion rate in usability testing.
- **SC-013**: The RBAC system supports tenants with up to 10,000 users and up to 25 client applications (the client quota from feature 004) with role assignments per user per client, with no degradation in authentication flow performance compared to a tenant with a single user.
- **SC-014**: Activity logs for a user with 1,000+ events load the first page (25 events) within 1 second.
- **SC-015**: All user management API endpoints return appropriate error responses (with clear, actionable messages) for 100% of invalid inputs, unauthorized access attempts, and quota violations.
- **SC-016**: Re-enabling a disabled user account restores their ability to authenticate within 1 second of the administrator confirming the enable action, without requiring a new invitation or password reset.
- **SC-017**: The user invitation system handles concurrent invitations for the same email (race condition) correctly 100% of the time, creating at most one user record per email per tenant.
