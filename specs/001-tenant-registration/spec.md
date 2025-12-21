# Feature Specification: Multi-Tenant Registration with Email Verification

**Feature Branch**: `001-tenant-registration`  
**Created**: 2025-12-06  
**Status**: Draft  
**Input**: User description: "Build an SSO server as SAAS platform using OIDC flow. First feature: create a register new tenant page. Platform will have multi-tenant. Email needs to be validated once user has successfully registered. Use OTP method that will be sent to email client"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Tenant Registration Form (Priority: P1)

A new organization administrator visits the SSO SaaS platform to register their organization as a new tenant. They fill out a registration form with organization details and their admin account information, then submit the form to initiate the tenant creation process.

**Why this priority**: This is the entry point to the entire platform. Without tenant registration, no organizations can use the SSO service. This is the absolute minimum viable feature.

**Independent Test**: Can be fully tested by accessing the registration page, filling the form with valid data, and verifying that a tenant record is created in the system (even without email verification).

**Acceptance Scenarios**:

1. **Given** a user visits the platform landing page, **When** they click "Register New Organization", **Then** they are presented with a registration form
2. **Given** the registration form is displayed, **When** the user fills in all required fields (organization name, admin email, admin name, password), **Then** all fields accept valid input
3. **Given** all required fields are filled with valid data, **When** the user submits the form, **Then** a new tenant is created with status "pending verification"
4. **Given** invalid data is entered (e.g., invalid email format), **When** the user tries to submit, **Then** clear validation errors are displayed
5. **Given** an organization name already exists, **When** the user tries to register, **Then** an error message indicates the organization name is taken

---

### User Story 2 - Email OTP Verification (Priority: P2)

After successfully submitting the registration form, the system sends a one-time password (OTP) to the admin's email address. The user receives the email, enters the OTP on a verification page, and their tenant account becomes active.

**Why this priority**: Email verification is critical for security and preventing abuse, but the registration form (P1) must exist first. This adds the security layer to the registration process.

**Independent Test**: Can be tested by registering a tenant (using P1), receiving an OTP email, entering the OTP on the verification page, and confirming the tenant status changes to "active".

**Acceptance Scenarios**:

1. **Given** a tenant has been registered, **When** the registration is submitted, **Then** an OTP is generated and sent to the admin email within 1 minute
2. **Given** an OTP email has been sent, **When** the user checks their email, **Then** they receive a well-formatted email with a 6-digit OTP code and clear instructions
3. **Given** the user has received the OTP, **When** they enter the correct OTP on the verification page, **Then** the tenant status changes to "active" and they are redirected to a success page
4. **Given** the user enters an incorrect OTP, **When** they submit it, **Then** an error message is displayed and they can retry
5. **Given** an OTP has been sent, **When** 10 minutes have passed, **Then** the OTP expires and cannot be used for verification
6. **Given** an OTP has expired, **When** the user requests a new OTP, **Then** a fresh OTP is generated and sent to the email

---

### User Story 3 - Post-Verification Tenant Access (Priority: P3)

Once the tenant is verified, the admin user can log in to their tenant dashboard to configure their SSO settings and manage their organization's identity provider setup.

**Why this priority**: This provides immediate value after verification and guides users to the next steps, but the core registration and verification (P1, P2) must work first.

**Independent Test**: Can be tested by verifying a tenant (using P1 and P2), then logging in with the admin credentials and accessing the tenant dashboard.

**Acceptance Scenarios**:

1. **Given** a tenant has been verified, **When** the admin user logs in with their credentials, **Then** they are authenticated and redirected to the tenant dashboard
2. **Given** the admin is logged into the dashboard, **When** they view their tenant information, **Then** they see their organization name, admin email, and tenant status as "active"
3. **Given** the admin is on the dashboard, **When** they navigate to SSO configuration, **Then** they can view OIDC configuration options (to be implemented in future features)

---

### Edge Cases

- What happens when a user tries to register with an email that already exists as an admin for another tenant?
- What happens when the email service is temporarily unavailable during OTP sending?
- What happens when a user requests multiple OTPs in quick succession (rate limiting)?
- What happens when a user closes the browser before entering the OTP and returns later?
- What happens when a user tries to use the same OTP twice?
- How does the system handle special characters in organization names?
- What happens when the network fails during registration form submission?
- How does the system prevent brute-force OTP guessing attempts?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a public registration page accessible without authentication
- **FR-002**: System MUST collect tenant information including organization name, admin full name, admin email, and password
- **FR-003**: System MUST validate email addresses using standard RFC 5322 format validation
- **FR-004**: System MUST enforce password complexity requirements (minimum 8 characters, at least one uppercase, one lowercase, one number, one special character)
- **FR-005**: System MUST check for unique organization names across all tenants before allowing registration
- **FR-006**: System MUST check for unique admin email addresses across all tenants before allowing registration
- **FR-007**: System MUST create tenant records with initial status "pending verification" upon successful form submission
- **FR-008**: System MUST generate a random 6-digit OTP code upon tenant registration
- **FR-009**: System MUST send OTP codes to the admin email address within 1 minute of registration
- **FR-010**: System MUST set OTP expiration time to 10 minutes from generation
- **FR-011**: System MUST provide an OTP verification page accessible via a link sent in the email or by redirecting after registration
- **FR-012**: System MUST validate OTP codes against the tenant's stored OTP
- **FR-013**: System MUST update tenant status from "pending verification" to "active" upon successful OTP verification
- **FR-014**: System MUST invalidate OTP codes after successful verification (single use)
- **FR-015**: System MUST allow users to request a new OTP if the previous one has expired
- **FR-016**: System MUST implement rate limiting for OTP requests (maximum 3 OTP requests per email per hour)
- **FR-017**: System MUST implement rate limiting for OTP verification attempts (maximum 5 attempts per OTP)
- **FR-018**: System MUST store passwords using bcrypt hashing with cost factor 12
- **FR-019**: System MUST support multi-tenant data isolation at the database level
- **FR-020**: System MUST log all registration attempts, OTP generation, and verification events for security auditing
- **FR-021**: System MUST provide clear error messages for validation failures without exposing security details
- **FR-022**: System MUST redirect verified users to a tenant dashboard or welcome page after successful verification
- **FR-023**: Email template MUST include tenant organization name, OTP code, expiration time, and support contact information
- **FR-024**: System MUST handle concurrent registration attempts gracefully (prevent race conditions)
- **FR-025**: System MUST provide a resend OTP functionality on the verification page

### Key Entities

- **Tenant**: Represents an organization in the multi-tenant SSO platform. Key attributes include unique tenant ID, organization name (unique), status (pending verification, active, suspended), creation timestamp, verification timestamp. Each tenant is isolated from other tenants.

- **Admin User**: Represents the primary administrator for a tenant. Key attributes include user ID, full name, email address (unique), hashed password, tenant association (foreign key), role (admin), creation timestamp. Each admin user belongs to exactly one tenant initially.

- **OTP Verification**: Represents an email verification code for tenant activation. Key attributes include OTP code (6-digit numeric), associated tenant ID, generation timestamp, expiration timestamp (10 minutes from generation), verification status (pending, verified, expired), verification attempt count, IP address of requester. Each OTP is tied to one tenant and has a limited lifespan.

- **Audit Log**: Represents security and activity events for compliance and monitoring. Key attributes include event type (registration_attempt, otp_sent, otp_verified, etc.), tenant ID, user ID, timestamp, IP address, success/failure status, additional metadata. Tracks all critical tenant lifecycle events.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: New tenants can complete the registration form in under 3 minutes
- **SC-002**: OTP emails are delivered within 60 seconds of registration for 99% of attempts
- **SC-003**: Users can successfully verify their email and activate their tenant in under 5 minutes from initial registration
- **SC-004**: 95% of valid registration attempts result in successful tenant creation
- **SC-005**: System prevents 100% of duplicate organization name registrations
- **SC-006**: System prevents 100% of duplicate admin email registrations
- **SC-007**: Zero OTP codes can be reused after successful verification
- **SC-008**: Rate limiting prevents more than 3 OTP requests per email per hour
- **SC-009**: Rate limiting prevents more than 5 verification attempts per OTP
- **SC-010**: System handles at least 100 concurrent registration requests without degradation
- **SC-011**: All registration and verification events are logged with 100% accuracy for audit purposes
- **SC-012**: 90% of users successfully complete email verification on their first attempt
- **SC-013**: Expired OTPs (older than 10 minutes) cannot be used for verification with 100% enforcement
- **SC-014**: Password validation catches 100% of weak passwords (less than 8 characters or missing required character types)
- **SC-015**: Tenant data isolation is verified with zero cross-tenant data leakage

### Assumptions

- Email service (SMTP or third-party email API like SendGrid, AWS SES, or Mailgun) is available and configured
- Email delivery is generally reliable but may have occasional delays (handled by retry logic)
- Users have access to their email inbox and can check it within reasonable time
- Organization names are case-insensitive for uniqueness checks
- Admin email addresses are case-insensitive for uniqueness checks
- Each tenant starts with exactly one admin user
- Password hashing uses industry-standard bcrypt with cost factor 12 (or argon2id)
- Multi-tenant isolation is enforced at the database level using tenant_id in all queries
- The platform uses HTTPS for all communications (SSL/TLS)
- Browser sessions are maintained during the registration-to-verification flow using secure session cookies
- OTP codes are cryptographically random and uniformly distributed across 000000-999999 range
