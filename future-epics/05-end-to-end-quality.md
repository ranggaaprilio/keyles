# Epic 05: End-to-End Quality & Production Smoke Tests

## Goal

Ensure the MVP is reliable, correct, and resilient through end-to-end tests of critical user journeys, edge case hardening, input validation at all layers, and production smoke tests that validate the deployed system.

## Why MVP

Unit and integration tests exist but they test components in isolation. No test exercises the full OAuth flow from redirect through consent to token exchange. Edge cases like concurrent requests, slow networks, invalid inputs, and service failures have unknown behavior. Without E2E tests and smoke tests, regressions in critical paths will be discovered by users, not by CI.

## Current State

- **Unit tests**: 34+ test files across domain and use case layers, using testify and mocks.
- **Integration tests**: 20+ test files for HTTP endpoints with full router setup, using test fixtures.
- **Frontend tests**: Component tests (RegistrationForm, OTPVerificationForm, ClientManagement, UserList, etc.) and page tests (DashboardPage, LoginPage) using Vitest + Testing Library.
- **Missing**: No end-to-end test that drives a browser through the full OAuth flow. No performance/load tests. No chaos/ resilience tests. No production smoke test suite.

## Tasks

### 5.1 — E2E test for OAuth authorization code flow
- Set up Playwright (or Cypress) for browser-based E2E tests
- Write test: "Full OAuth Authorization Code + PKCE flow"
  - Register a new tenant (or use seed data)
  - Create an OAuth client
  - As an end-user, navigate to `/oauth2/auth?client_id=...&redirect_uri=...&...`
  - Log in as end-user
  - See consent screen with correct client name and scopes
  - Approve consent
  - Verify redirect to client's redirect_uri with `code` and `state`
  - Exchange code for tokens at `/oauth2/token`
  - Verify access_token, id_token, refresh_token in response
- Write test: "Consent denial flow"
- Write test: "PKCE enforcement — reject missing code_challenge"
- Write test: "Invalid client_id shows error page"
- Write test: "Expired authorization code rejected at token endpoint"

### 5.2 — E2E test for tenant registration and admin flow
- Write test: "Tenant registration → OTP verification → admin login → dashboard"
- Write test: "Admin creates OAuth client → views client list → rotates secret"
- Write test: "Admin invites end-user → end-user accepts invitation → end-user can authenticate"

### 5.3 — Input validation hardening
- Audit all handler input structs: add missing validation tags (min, max, email format, URL format)
- Add Gin binding validation error handling that returns user-friendly messages
- Test with malicious inputs: SQL injection strings, XSS payloads, oversized inputs, Unicode homoglyphs, null bytes
- Ensure redirect_uri validation prevents open redirect attacks (exact match or registered prefix only)
- Validate client_id, tenant_id, user_id formats everywhere (UUID validation)

### 5.4 — Concurrent request safety
- Test: concurrent token exchanges with the same auth code (only one should succeed)
- Test: concurrent refresh token rotations (should not produce duplicate valid tokens)
- Test: concurrent user status changes (enable/disable race)
- Test: concurrent role assignment/revocation
- Verify database transaction isolation where needed

### 5.5 — Error handling and edge cases
- Test: email service unavailable during registration (should not crash, should return appropriate error)
- Test: Redis unavailable during token operations (degraded mode)
- Test: database connection pool exhausted (connection timeout, retry)
- Test: very large request bodies (rejected with 413)
- Test: malformed JSON (rejected with 400)
- Test: missing required fields (rejected with 422 and field-level errors)
- Test: invalid JWT signatures (rejected with 401)
- Test: expired JWT (rejected with 401)

### 5.6 — Token and session edge cases
- Test: refresh token reuse detection (refresh token rotation — old token invalidated after use)
- Test: token introspection with valid, expired, and revoked tokens
- Test: token revocation cascades (revoking refresh token invalidates access tokens)
- Test: signing key rotation does not invalidate existing tokens

### 5.7 — Production smoke test suite
- Create a script `scripts/smoke-test.sh` that:
  - Hits `/health`, `/health/db`, `/health/redis` — all must return 200
  - Hits `/.well-known/openid-configuration` — must return valid OIDC discovery doc
  - Hits `/.well-known/jwks.json` — must return valid JWKS with at least one key
  - Performs a full OAuth flow end-to-end (using curl)
  - Validates that rate limiting headers are present
  - Validates that security headers are present
- This script can be run as a post-deployment validation or as a CronJob for continuous validation

### 5.8 — Load test baseline
- Create a basic load test script (using k6 or vegeta) for:
  - Health endpoint (high throughput baseline)
  - Token endpoint (authenticated, moderate throughput)
  - Registration endpoint (rate-limited, low throughput)
- Document baseline performance numbers (requests/second, p95 latency)
- These are not for performance tuning but to catch severe regressions

## Acceptance Criteria

1. E2E OAuth flow passes in CI: registration → client creation → authorization → consent → token exchange
2. All handler inputs reject invalid data with clear error messages
3. Concurrent requests for single-use resources (auth codes, refresh tokens) produce exactly one success
4. Service returns appropriate errors (not crashes) when dependencies are unavailable
5. Refresh token reuse is detected and results in token family invalidation
6. Smoke test script passes against a freshly deployed instance
7. Load test runs without errors or crashes at documented baseline throughput
