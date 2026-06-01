# Tasks: OAuth Consent Flow and End-User Authentication

**Input**: Design documents from `/specs/006-oauth-consent-flow/`  
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓, quickstart.md ✓  
**Feature Branch**: `006-oauth-consent-flow`

**Tests**: Per the constitution and NFR-005, tests are mandatory. Write tests first,
confirm they fail, then implement. Maintain at least 85% coverage for domain and
business logic and add integration coverage for each browser-flow handler.

**Organization**: Tasks are grouped by user story. Shared Redis, configuration,
audit, and interface work is completed first so each story phase can be exercised
independently.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel because it touches different files and has no
  dependency on incomplete tasks in the same phase.
- **[Story]**: Maps implementation work to US1 through US6 from spec.md.
- Every checklist item includes an exact repository path.

## User Story Index

| ID | Story | Priority |
| --- | --- | --- |
| US1 | Authenticate an End-User During OAuth | P1 |
| US2 | Approve or Deny Consent | P1 |
| US3 | Reuse and Refresh End-User Sessions | P1 |
| US5 | Complete the Existing PKCE Token Flow | P1 |
| US4 | Handle OAuth Errors Safely | P2 |
| US6 | End the Keyles Browser SSO Session | P2 |

---

## Phase 1: Setup and Configuration

**Purpose**: Add browser-flow configuration and document local runtime values.

- [X] T001 [P] Add configuration tests for `FRONTEND_URL`, `SECURITY_COOKIE_SECURE`, `SECURITY_SESSION_TTL`, `OAUTH_AUTH_TRANSACTION_TTL`, `RATE_LIMIT_OAUTH_LOGIN_FAILURES`, and `RATE_LIMIT_OAUTH_LOGIN_WINDOW_SECONDS` defaults and invalid values in `backend/infrastructure/config/config_test.go`
- [X] T002 Implement the tested OAuth browser-flow configuration fields, parsing, defaults, and production URL validation in `backend/infrastructure/config/config.go`
- [X] T003 [P] Add documented local values for frontend URL, secure-cookie toggle, session TTL, transaction TTL, and OAuth login failure limits in `backend/.env.example`
- [X] T004 [P] Add Docker Compose backend environment values for frontend URL, secure-cookie toggle, session TTL, transaction TTL, and OAuth login failure limits in `docker-compose.yml`

**Checkpoint**: Browser-flow configuration is explicit and does not include a
cookie-domain setting.

---

## Phase 2: Foundational Security State

**Purpose**: Define and implement shared domain abstractions, Redis security state,
audit helpers, and mocks required by every browser-flow story.

**⚠️ CRITICAL**: Complete this phase before starting user-story implementation.

### Tests First

- [X] T005 [P] Add Redis authorization-transaction repository tests for TTL storage, lookup, binding updates, atomic one-time completion, expiry, replay rejection, and Redis errors in `backend/infrastructure/persistence/redis/authorization_transaction_repository_test.go`
- [X] T006 [P] Add Redis OAuth login-throttler tests for source-IP counters, tenant-email counters, five-failure blocking, fixed 15-minute TTL created only on first failure, atomic increments without TTL extension, email-bucket clearing, and Redis errors in `backend/infrastructure/persistence/redis/login_throttler_test.go`
- [X] T007 [P] Add domain audit-event tests covering all OAuth browser-flow event constants and sanitized event-data construction in `backend/tests/unit/domain/audit_log_test.go`
- [X] T008 [P] Add OAuth audit-helper tests for required identifiers, secret exclusion, repository writes, and sanitized structured-log fallback on audit persistence failure in `backend/tests/unit/usecase/oauth_audit_test.go`

### Domain Interfaces and Entities

- [X] T009 Create `AuthorizationTransaction`, stage constants, `AuthorizationTransactionRepository`, and atomic completion contract in `backend/domain/repositories/authorization_transaction_repository.go`
- [X] T010 [P] Make `Session` client-agnostic by removing the legacy `ClientID` field, adding `AuthenticatedAt`, and documenting cross-client SSO reuse semantics in `backend/domain/repositories/session_repository.go`
- [X] T011 [P] Define the Redis-backed dual-key `LoginThrottler` abstraction for check, record-failure, and clear-email operations in `backend/domain/services/login_throttler.go`
- [X] T012 [P] Add OAuth login success/failure/throttled, consent approved/denied, logout, and invalid-callback event constants in `backend/domain/entities/audit_log.go`

### Infrastructure and Shared Use Cases

- [X] T013 Implement Redis storage, binding updates, and atomic transaction completion for `AuthorizationTransactionRepository` in `backend/infrastructure/persistence/redis/authorization_transaction_repository.go`
- [X] T014 [P] Implement Redis source-IP and tenant-email failure buckets for `LoginThrottler` in `backend/infrastructure/persistence/redis/login_throttler.go`
- [X] T015 [P] Update Redis session serialization tests for the client-agnostic shape, absence of legacy `ClientID`, and `AuthenticatedAt` round trips in `backend/infrastructure/persistence/redis/session_repository_test.go`
- [X] T016 [P] Add a testify mock for `AuthorizationTransactionRepository` in `backend/tests/mocks/authorization_transaction_repository.go`
- [X] T017 [P] Add a testify mock for `LoginThrottler` in `backend/tests/mocks/login_throttler.go`
- [X] T018 [P] Extend the session mock fixtures for `AuthenticatedAt` assertions in `backend/tests/mocks/session_repository.go`
- [X] T019 Implement sanitized OAuth audit-event creation and structured-log fallback around the existing audit repository in `backend/usecase/auth/oauth_audit.go`

**Clean Architecture Checkpoint**:

- Domain files contain interfaces and entities only.
- Redis implementations remain under `backend/infrastructure/persistence/redis/`.
- Shared audit logic depends on `repositories.AuditRepository`.

---

## Phase 3: User Story 1 - Authenticate an End-User During OAuth (Priority: P1) 🎯 MVP

**Goal**: Validate the initial OAuth request, route unauthenticated users to the
frontend login page, authenticate a tenant-scoped end-user, set a host-only SSO
cookie, and continue to consent.

**Independent Test**: Start a valid `/oauth2/auth` request without a session, log in
with an active end-user who has a client role, verify a host-only HttpOnly cookie is
set, and verify redirect to `/oauth2/consent?transaction_id=...`.

### Tests for US1

- [X] T020 [P] [US1] Add unit tests for initial request validation, trusted callback validation, `crypto/rand`-backed opaque transaction and interaction-CSRF generation, invalid-callback audit events, and Redis fail-closed behavior in `backend/tests/unit/usecase/oauth_interaction_test.go`
- [X] T021 [P] [US1] Add unit tests for tenant-scoped email lookup, generic credential failures, active-user and role checks, fixed-window dual-key throttle checks, failure increments, email-bucket clearing, `crypto/rand`-backed client-agnostic session creation, transaction binding, and login audit events in `backend/tests/unit/usecase/authenticate_end_user_test.go`
- [X] T022 [P] [US1] Replace header-based authorization integration fixtures with browser-flow redirect, login, cookie, fixed-window throttle, direct-peer source-IP behavior despite spoofed forwarded headers, credentialed frontend-origin CORS, invalid callback, and local Redis-outage redirect tests in `backend/tests/integration/oauth_auth_test.go`
- [X] T023 [P] [US1] Add frontend service tests for credentialed login requests and interaction redirects in `frontend/src/tests/services/oauthInteractionService.test.ts`
- [X] T024 [P] [US1] Add frontend login-page tests for transaction parsing, email/password submission, loading state, generic errors, throttle errors, and navigation to consent in `frontend/src/tests/OAuthLoginPage.test.tsx`

### Implementation for US1

- [X] T025 [US1] Implement validated authorization-transaction initialization with `crypto/rand`-backed opaque transaction and interaction-CSRF identifiers, local invalid-callback handling, local Redis fail-closed error redirects, and login redirect selection in `backend/usecase/auth/oauth_interaction.go`
- [X] T026 [US1] Implement OAuth end-user credential authentication, fixed-window dual-key throttling, active-role validation, `crypto/rand`-backed client-agnostic session creation, email-bucket clearing, last-login update, transaction binding, and audit emission in `backend/usecase/auth/authenticate_end_user.go`
- [X] T027 [US1] Replace `X-User-ID` handling with transaction initialization, derive OAuth throttle and audit source IP from the direct TCP peer without trusting forwarded headers, and add `POST /oauth2/login` JSON handling plus host-only HttpOnly `SameSite=Lax; Path=/` cookie creation in `backend/interfaces/http/handlers/oauth_handler.go`
- [X] T028 [US1] Register `POST /oauth2/login` alongside the existing authorization endpoint in `backend/interfaces/http/router.go`
- [X] T029 [US1] Construct the transaction repository, session repository, login throttler, OAuth audit helper, interaction use case, and end-user login use case in `backend/cmd/server/main.go`
- [X] T030 [P] [US1] Add authorization-transaction, OAuth login, interaction redirect, and interaction-error TypeScript contracts in `frontend/src/types/oauth.ts`
- [X] T031 [US1] Implement the credentialed OAuth login API client and shared interaction-error parsing in `frontend/src/services/oauthInteractionService.ts`
- [X] T032 [US1] Create the end-user OAuth login form with transaction-aware navigation and generic credential errors in `frontend/src/pages/OAuthLoginPage.tsx`
- [X] T033 [US1] Register the public `/oauth2/login` route in `frontend/src/App.tsx`

**Checkpoint**: A valid OAuth request reaches frontend login, authenticates an
eligible user, and establishes an SSO session without trusting browser-supplied
OAuth fields.

---

## Phase 4: User Story 2 - Approve or Deny Consent (Priority: P1)

**Goal**: Show trusted consent details and atomically approve or deny the pending
authorization transaction.

**Independent Test**: Load consent for a valid bound session, approve once and
verify callback `code` plus original `state`; repeat with deny and verify
`error=access_denied`.

### Tests for US2

- [X] T034 [P] [US2] Add unit tests for consent-detail ownership, display-safe response fields, expired transactions, missing sessions, and Redis fail-closed behavior in `backend/tests/unit/usecase/get_consent_details_test.go`
- [X] T035 [P] [US2] Add unit tests for interaction CSRF validation, active-user and role re-check, atomic approval, denial, state preservation, replay rejection, Redis failure, and consent audit events in `backend/tests/unit/usecase/consent_decision_test.go`
- [X] T036 [P] [US2] Add integration tests for consent detail reads, approve redirects, deny redirects, callback query preservation, replay rejection, and Redis outages in `backend/tests/integration/oauth_consent_test.go`
- [X] T037 [P] [US2] Add frontend consent-page tests for detail loading, current-user display, scope rendering, allow, deny, expired interactions, and callback navigation in `frontend/src/tests/OAuthConsentPage.test.tsx`

### Implementation for US2

- [X] T038 [US2] Implement bound-session consent-detail reads with display-safe client and user fields in `backend/usecase/auth/get_consent_details.go`
- [X] T039 [US2] Refactor consent finalization into one atomic approval-or-denial path with interaction CSRF checks, active-user and role revalidation, existing five-minute authorization-code storage, callback construction, and audit emission in `backend/usecase/auth/consent_decision.go`
- [X] T040 [US2] Add consent-detail and consent-decision handlers with local Redis-outage responses in `backend/interfaces/http/handlers/oauth_handler.go`
- [X] T041 [US2] Register `GET /oauth2/consent/:transactionId` and `POST /oauth2/consent` in `backend/interfaces/http/router.go`
- [X] T042 [US2] Extend the interaction API client with credentialed consent-detail and consent-decision methods in `frontend/src/services/oauthInteractionService.ts`
- [X] T043 [US2] Create the consent container page that loads trusted details and composes the existing `ConsentScreen` component in `frontend/src/pages/OAuthConsentPage.tsx`
- [X] T044 [US2] Register the public `/oauth2/consent` route in `frontend/src/App.tsx`

**Checkpoint**: An authenticated user can approve or deny once, and callback
redirects preserve the original client state safely.

---

## Phase 5: User Story 3 - Reuse and Refresh End-User Sessions (Priority: P1)

**Goal**: Reuse an eligible SSO session while honoring OIDC `prompt`, `max_age`,
`nonce`, and authentication-time behavior.

**Independent Test**: Complete one login, start another authorization request and
verify direct consent; repeat with `prompt=login`, `max_age=0`, exceeded session age,
and `prompt=none`.

### Tests for US3

- [X] T045 [P] [US3] Add unit cases for cross-client client-agnostic session reuse, active-user and current-role eligibility, prompt parsing, `prompt=none` conflicts, unsupported `select_account`, `prompt=login`, `prompt=consent`, non-negative `max_age`, expired session age, and silent `login_required` or `consent_required` callbacks in `backend/tests/unit/usecase/oauth_interaction_test.go`
- [X] T046 [P] [US3] Add integration tests for cross-client session reuse in one tenant, forced login after user disable or role removal, prompt-driven login, silent OIDC errors, and absent `Domain` cookie attribute in `backend/tests/integration/oauth_session_test.go`
- [X] T047 [P] [US3] Add token-issuance unit cases asserting approved browser flows propagate `nonce` and `auth_time` into ID-token claims in `backend/tests/unit/usecase/issue_token_test.go`

### Implementation for US3

- [X] T048 [P] [US3] Extend authorization-code fields with `Nonce` and `AuthenticatedAt` in `backend/domain/entities/authorization_code.go`
- [X] T049 [P] [US3] Extend shared token claims with optional `nonce` and `auth_time` JSON fields in `backend/domain/services/token_service.go`
- [X] T050 [US3] Add prompt parsing, max-age evaluation, active-user and current-role revalidation for cross-client eligible-session binding, and silent callback error selection to `backend/usecase/auth/oauth_interaction.go`
- [X] T051 [US3] Resolve the host-only SSO cookie during authorization and route eligible sessions directly to consent in `backend/interfaces/http/handlers/oauth_handler.go`
- [X] T052 [US3] Copy session `AuthenticatedAt` and transaction `Nonce` into approved authorization codes in `backend/usecase/auth/consent_decision.go`
- [X] T053 [US3] Include authorization-code `Nonce` and `AuthenticatedAt` in ID-token claims during code exchange in `backend/usecase/auth/issue_token.go`
- [X] T054 [US3] Preserve `nonce` and `auth_time` while signing and validating JWT claims in `backend/infrastructure/services/rsa_token_service.go`

**Checkpoint**: Existing sessions reduce login prompts correctly and ID tokens
preserve OIDC authentication context.

---

## Phase 6: User Story 5 - Complete the Existing PKCE Token Flow (Priority: P1)

**Goal**: Prove that an authorization code issued after browser login and consent
still exchanges once through the existing S256 PKCE token endpoint.

**Independent Test**: Complete login and consent, exchange the returned code with
the matching verifier, then verify access, ID, and refresh tokens; retry with wrong
and reused codes.

### Tests and Compatibility Work for US5

- [X] T055 [P] [US5] Update token integration fixtures to obtain codes through browser login and consent, then verify matching-verifier success, wrong-verifier rejection, expiry rejection, and one-time consumption in `backend/tests/integration/oauth_token_test.go`
- [X] T056 [P] [US5] Add frontend OAuth callback regression tests for state validation and approved-code token exchange compatibility in `frontend/src/tests/OAuthCallback.test.tsx`
- [X] T057 [US5] Adapt existing token-exchange test helpers to the transaction-backed authorization path while preserving public and confidential client behavior in `backend/tests/integration/oauth_token_test.go`

**Checkpoint**: The full browser login, consent, and token-exchange chain works
without regressing the existing PKCE endpoint.

---

## Phase 7: User Story 4 - Handle OAuth Errors Safely (Priority: P2)

**Goal**: Render user-friendly local failures without redirecting untrusted callback
URIs or exposing sensitive implementation details.

**Independent Test**: Submit invalid clients, mismatched callbacks, expired
transactions, throttle errors, and Redis outages; verify local error rendering and
safe callback behavior.

### Tests for US4

- [X] T058 [P] [US4] Add handler integration cases proving invalid clients and invalid callbacks stay local, authorization-init Redis outages redirect to the local frontend error page, JSON interaction Redis outages return local `503` responses, and validated callback errors preserve `state` in `backend/tests/integration/oauth_error_test.go`
- [X] T059 [P] [US4] Add frontend error-page tests for `invalid_client`, `invalid_request`, `access_denied`, `temporarily_unavailable`, expired transactions, and safe default messaging in `frontend/src/tests/OAuthErrorPage.test.tsx`

### Implementation for US4

- [X] T060 [US4] Centralize local frontend-error URL construction and callback-safe OAuth error redirects in `backend/interfaces/http/handlers/oauth_handler.go`
- [X] T061 [P] [US4] Create a reusable user-friendly OAuth error panel with safe error-code mapping in `frontend/src/components/auth/OAuthErrorPanel.tsx`
- [X] T062 [US4] Create the local OAuth error page and parse only display-safe query parameters in `frontend/src/pages/OAuthErrorPage.tsx`
- [X] T063 [US4] Register the public `/oauth2/error` route in `frontend/src/App.tsx`

**Checkpoint**: Unsafe callbacks never receive redirects and browser-flow errors are
clear without leaking internal details.

---

## Phase 8: User Story 6 - End the Keyles Browser SSO Session (Priority: P2)

**Goal**: Let an end-user terminate the provider-local browser session without
revoking external-client tokens or accepting client-controlled redirects.

**Independent Test**: Log in, submit `POST /oauth2/logout`, verify Redis deletion
and expired cookie, then start authorization again and verify login is required.

### Tests for US6

- [X] T064 [P] [US6] Add unit tests for idempotent session deletion, logout audit emission, Redis-delete failure tolerance, and unconditional cookie-expiry intent in `backend/tests/unit/usecase/logout_end_user_test.go`
- [X] T065 [P] [US6] Add handler integration tests for active-session logout, missing-session logout, Redis-outage logout, expired host-only cookie response, and forced login afterward in `backend/tests/integration/oauth_logout_test.go`
- [X] T066 [P] [US6] Add frontend logout-page tests for credentialed logout submission and signed-out rendering in `frontend/src/tests/OAuthLogoutPage.test.tsx`

### Implementation for US6

- [X] T067 [US6] Implement idempotent provider-local session deletion and logout audit emission while tolerating Redis deletion failure in `backend/usecase/auth/logout_end_user.go`
- [X] T068 [US6] Add `POST /oauth2/logout` handling that always expires the host-only `Path=/` cookie and accepts no redirect parameter in `backend/interfaces/http/handlers/oauth_handler.go`
- [X] T069 [US6] Register `POST /oauth2/logout` in `backend/interfaces/http/router.go`
- [X] T070 [US6] Extend the interaction API client with credentialed provider-local logout in `frontend/src/services/oauthInteractionService.ts`
- [X] T071 [US6] Create the signed-out provider page that submits logout once and renders completion state in `frontend/src/pages/OAuthLogoutPage.tsx`
- [X] T072 [US6] Register the public `/oauth2/logout` route in `frontend/src/App.tsx`

**Checkpoint**: Provider-local logout reliably expires the browser session cookie,
including during Redis outages.

---

## Phase 9: Polish and Cross-Cutting Verification

**Purpose**: Validate architecture, documentation, and the complete Docker Compose
flow after all desired stories are implemented.

- [X] T073 [P] Update backend OAuth browser-flow configuration and security documentation in `backend/README.md`
- [X] T074 [P] Add a Docker Compose end-to-end OAuth integration test suite covering approve, deny, cross-client session reuse, forced login after disable or role removal, fixed-window throttle behavior, spoofed forwarded-header rejection, credentialed frontend-origin CORS, logout, Redis outage, PKCE rejection, and compatibility smoke checks for refresh, revocation, introspection, discovery, JWKS, and userinfo in `backend/tests/integration/oauth_browser_flow_test.go`
- [X] T075 Run `gofmt` on changed Go files and execute `make test` plus `make test-coverage` from `backend/`, explicitly verifying the existing refresh, revocation, introspection, discovery, JWKS, and userinfo regression suites remain green
- [X] T076 Run `npm run test -- --run`, `npm run lint`, and `npm run build` from `frontend/`, then complete the manual matrix in `specs/006-oauth-consent-flow/quickstart.md`

---

## Dependencies and Execution Order

### Phase Dependencies

- **Phase 1 Setup**: Starts immediately.
- **Phase 2 Foundational Security State**: Depends on Phase 1 and blocks all stories.
- **US1**: Starts after Phase 2 and establishes the browser login session.
- **US2**: Depends on US1 because consent requires a bound authenticated session.
- **US3**: Depends on US1 and US2 because it reuses sessions and propagates approved
  interaction context into codes.
- **US5**: Depends on US2 and US3 because it verifies codes issued by the completed
  browser flow.
- **US4**: Can start after US1; finish after US2 so consent errors are covered.
- **US6**: Can start after US1 and proceeds independently of US2, US3, and US5.
- **Polish**: Depends on all selected stories.

### User Story Dependency Graph

```text
Setup -> Foundation -> US1 -> US2 -> US3 -> US5
                         |      |
                         +----> US4
                         |
                         +----> US6
```

### Within Each User Story

- Write tests first and confirm they fail before implementation.
- Define domain changes before infrastructure implementations.
- Implement use cases before handlers and frontend containers.
- Keep browser OAuth parameters trusted only after server-side validation.
- Stop at each checkpoint and run the story's focused tests.

## Parallel Opportunities

- Setup documentation tasks T003 and T004 can run while configuration tests and
  implementation proceed.
- Foundational test tasks T005 through T008 can run in parallel.
- Domain tasks T010 through T012 and mock tasks T016 through T018 can run in
  parallel after their interfaces exist.
- Frontend tests can run in parallel with backend use-case tests within each story.
- After US1 completes, US6 can proceed alongside US2 and US4.
- Error-page frontend work in US4 can proceed while backend error integration tests
  are written.

## Parallel Example: User Story 2

```text
Task T034: Add consent-detail unit tests in backend/tests/unit/usecase/get_consent_details_test.go
Task T035: Add consent-decision unit tests in backend/tests/unit/usecase/consent_decision_test.go
Task T036: Add consent integration tests in backend/tests/integration/oauth_consent_test.go
Task T037: Add frontend consent-page tests in frontend/src/tests/OAuthConsentPage.test.tsx
```

## Implementation Strategy

### MVP First

1. Complete Setup and Foundational Security State.
2. Complete US1 end-user login.
3. Complete US2 consent approval and denial.
4. Complete US3 session semantics and OIDC context propagation.
5. Complete US5 browser-to-token PKCE regression coverage.
6. Validate the P1 flow end-to-end before adding P2 UX and logout work.

### Incremental Delivery

1. Foundation provides typed Redis security state and audit support.
2. US1 establishes end-user login and host-only sessions.
3. US2 turns login into a usable authorization-code browser flow.
4. US3 completes SSO reuse and OIDC prompt semantics.
5. US5 proves compatibility with the existing token endpoint.
6. US4 adds polished safe error pages.
7. US6 adds explicit provider-local logout.

## Notes

- `[P]` tasks touch different files and can be scheduled concurrently.
- The full P1 MVP is US1, US2, US3, and US5 because an OAuth provider is only
  useful once a client can exchange the approved code.
- No PostgreSQL migration is expected for this feature.
- Never log passwords, cookies, authorization codes, PKCE values, or other secrets.
- Keep the SSO cookie host-only: do not add a `Domain` attribute.
