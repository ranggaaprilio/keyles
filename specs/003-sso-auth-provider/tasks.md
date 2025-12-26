# Tasks: Core SSO Auth Provider (OAuth 2.0 + OIDC)

**Input**: Design documents from `/specs/003-sso-auth-provider/`
**Prerequisites**: plan.md (complete), spec.md (complete), research.md (complete), data-model.md (complete), contracts/openapi.yaml (complete), quickstart.md (complete)

**Tests**: Per constitution, tests are MANDATORY. All business logic requires unit tests (≥85% coverage), and all handlers require integration tests.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Web app (Clean Architecture)**:
  - Backend: `backend/domain/`, `backend/usecase/`, `backend/infrastructure/`, `backend/interfaces/`
  - Frontend: `frontend/src/components/`, `frontend/src/services/`
  - Tests: `backend/tests/unit/`, `backend/tests/integration/`, `frontend/tests/`
- Migrations: `backend/migrations/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Install backend OAuth dependencies (go get github.com/ory/fosite, github.com/jackc/pgx/v5, github.com/redis/go-redis/v9, github.com/go-chi/chi/v5, github.com/ulule/limiter/v3)
- [x] T002 [P] Install frontend OAuth UI dependencies (npm install axios react-router-dom, update package.json)
- [x] T003 Create backend/.env.example with OAuth-specific variables (OAUTH_ISSUER, OAUTH_ACCESS_TOKEN_TTL=900, OAUTH_REFRESH_TOKEN_TTL=604800, JWT_SIGNING_KEY_PATH, JWT_PUBLIC_KEY_PATH, JWT_KEY_ID, RATE_LIMIT_TOKEN_ENDPOINT=10)
- [x] T004 [P] Create frontend/.env.example with OAuth variables (VITE_OAUTH_ISSUER, VITE_CLIENT_ID, VITE_OAUTH_REDIRECT_URI, VITE_OAUTH_SCOPES)
- [x] T005 Create backend/cmd/keygen/main.go for RSA keypair generation utility (generates 2048-bit keys to backend/keys/)
- [x] T006 [P] Create backend/cmd/seed/main.go for test data seeder (creates dev tenant, users, clients, role assignments)
- [x] T007 Create backend/Makefile with targets: build, run, test, test-coverage, migrate-up, migrate-down, seed, clean, docker-up, docker-down
- [x] T008 Update docker-compose.yml to include Redis service if not already present
- [x] T009 Update backend/README.md with OAuth setup instructions and quickstart reference
- [x] T010 [P] Update frontend/README.md with OAuth integration guide

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Database Migrations

- [x] T011 Create migration 000004_create_clients_table.up.sql with columns: client_id (PK), tenant_id (FK), client_name, client_secret_hash, allowed_redirect_uris (TEXT[]), created_at, updated_at, is_active
- [x] T012 Create migration 000004_create_clients_table.down.sql
- [x] T013 [P] Create migration 000005_create_user_role_assignments_table.up.sql with columns: id (PK), user_id (FK), client_id (FK), role (VARCHAR 50), assigned_at, assigned_by, is_active, UNIQUE(user_id, client_id, role)
- [x] T014 [P] Create migration 000005_create_user_role_assignments_table.down.sql
- [x] T015 [P] Create migration 000006_create_refresh_tokens_table.up.sql with columns: token_id (PK), user_id (FK), client_id (FK), tenant_id (FK), token_hash, expires_at, revoked_at, created_at, INDEX(user_id, client_id)
- [x] T016 [P] Create migration 000006_create_refresh_tokens_table.down.sql
- [x] T017 [P] Create migration 000007_create_signing_keys_table.up.sql with columns: key_id (PK), algorithm (VARCHAR 10), public_key (TEXT), private_key_encrypted (TEXT), created_at, expires_at, is_active
- [x] T018 [P] Create migration 000007_create_signing_keys_table.down.sql

### Domain Entities (Clean Architecture - Innermost Layer)

- [x] T019 Define Client entity in backend/domain/entities/client.go with fields: ClientID, TenantID, ClientName, ClientSecretHash, AllowedRedirectURIs, IsActive, CreatedAt, UpdatedAt, validation methods (ValidateRedirectURI, IsURIAllowed)
- [x] T020 [P] Define UserRole entity in backend/domain/entities/user_role.go with fields: ID, UserID, ClientID, Role, AssignedAt, AssignedBy, IsActive, validation methods (ValidateRole)
- [x] T021 [P] Define RefreshToken entity in backend/domain/entities/refresh_token.go with fields: TokenID, UserID, ClientID, TenantID, TokenHash, ExpiresAt, RevokedAt, CreatedAt, methods (IsExpired, IsRevoked, IsValid)
- [x] T022 [P] Define SigningKey entity in backend/domain/entities/signing_key.go with fields: KeyID, Algorithm, PublicKey, PrivateKeyEncrypted, CreatedAt, ExpiresAt, IsActive
- [x] T023 [P] Define AuthorizationCode entity (in-memory/Redis) in backend/domain/entities/authorization_code.go with fields: Code, ClientID, UserID, TenantID, RedirectURI, CodeChallenge, CodeChallengeMethod, Scope, ExpiresAt, IsUsed

### Domain Repository Interfaces (Clean Architecture - Domain Layer)

- [x] T024 Define ClientRepository interface in backend/domain/repositories/client_repository.go with methods: Create(client), GetByID(clientID), GetByClientID(clientID, tenantID), Update(client), Delete(clientID), ListByTenant(tenantID), ValidateCredentials(clientID, clientSecret)
- [x] T025 [P] Define RoleRepository interface in backend/domain/repositories/role_repository.go with methods: AssignRole(userID, clientID, role, assignedBy), RevokeRole(userID, clientID, role), GetUserRoles(userID, clientID), HasRole(userID, clientID), ListRolesByClient(clientID), ListRolesByUser(userID)
- [x] T026 [P] Define RefreshTokenRepository interface in backend/domain/repositories/refresh_token_repository.go with methods: Create(token), GetByToken(tokenHash), Revoke(tokenID), RevokeAllForUser(userID, clientID), DeleteExpired(), IsRevoked(tokenHash)
- [x] T027 [P] Define AuthCodeRepository interface in backend/domain/repositories/auth_code_repository.go with methods: Store(code, metadata, ttl), Get(code), MarkAsUsed(code), Delete(code)
- [x] T028 [P] Define SessionRepository interface in backend/domain/repositories/session_repository.go with methods: Create(sessionID, userID, metadata, ttl), Get(sessionID), Delete(sessionID), Exists(sessionID)
- [x] T029 [P] Define SigningKeyRepository interface in backend/domain/repositories/signing_key_repository.go with methods: Create(key), GetActive(), GetByKeyID(keyID), Deactivate(keyID), DeleteExpired()

### Domain Service Interfaces (Clean Architecture - Domain Layer)

- [x] T030 Define OAuthProvider interface in backend/domain/services/oauth_provider.go with methods: GenerateAuthCode(req AuthRequest), ExchangeCodeForTokens(code, verifier, clientID, clientSecret), RefreshAccessToken(refreshToken, clientID), ValidateToken(token), RevokeToken(token)
- [x] T031 [P] Define TokenService interface in backend/domain/services/token_service.go with methods: SignIDToken(claims), SignAccessToken(claims), ValidateTokenSignature(token), GetPublicKey(keyID), GetJWKS()

### Infrastructure - Repository Implementations (Clean Architecture - Infrastructure Layer)

- [x] T032 Implement PostgresClientRepository in backend/infrastructure/persistence/postgres/client_repository.go using pgx connection pool, implements ClientRepository interface
- [x] T033 [P] Implement PostgresRoleRepository in backend/infrastructure/persistence/postgres/role_repository.go using pgx, implements RoleRepository interface
- [x] T034 [P] Implement PostgresRefreshTokenRepository in backend/infrastructure/persistence/postgres/refresh_token_repository.go using pgx, implements RefreshTokenRepository interface, stores bcrypt-hashed tokens
- [x] T035 [P] Implement PostgresSigningKeyRepository in backend/infrastructure/persistence/postgres/signing_key_repository.go using pgx, implements SigningKeyRepository interface
- [x] T036 Implement RedisAuthCodeRepository in backend/infrastructure/persistence/redis/auth_code_repository.go using go-redis client, implements AuthCodeRepository interface, 5-minute TTL
- [x] T037 [P] Implement RedisSessionRepository in backend/infrastructure/persistence/redis/session_repository.go using go-redis, implements SessionRepository interface, 8-hour TTL
- [x] T038 [P] Implement RedisTokenCacheRepository in backend/infrastructure/persistence/redis/token_cache_repository.go for refresh token caching (7-day TTL), uses SET with EX for automatic expiration
- [x] T039 [P] Implement RedisRateLimiter in backend/infrastructure/persistence/redis/rate_limiter.go using ulule/limiter, supports per-client_id rate limiting (10 req/min)

### Infrastructure - Service Implementations (Clean Architecture - Infrastructure Layer)

- [x] T040 Implement FositeOAuthProvider in backend/infrastructure/services/fosite_oauth_provider.go wrapping github.com/ory/fosite, implements OAuthProvider interface, handles PKCE validation
- [x] T041 [P] Implement RSATokenService in backend/infrastructure/services/rsa_token_service.go using crypto/rsa and crypto/sha256, implements TokenService interface, RS256 signing, JWKS generation

### HTTP Middleware (Clean Architecture - Interfaces Layer)

- [x] T042 Create rate limiting middleware in backend/interfaces/http/middleware/rate_limiter.go using RedisRateLimiter (implements FR-057: 10 req/min per client_id), extracts client_id from request, returns HTTP 429 when exceeded
- [x] T043 [P] Create tenant context middleware in backend/interfaces/http/middleware/tenant_context.go extracts tenant_id from client_id lookup and adds to request context
- [x] T044 [P] Update error handler middleware in backend/interfaces/http/middleware/error_handler.go to handle OAuth-specific errors (invalid_request, invalid_grant, unauthorized_client, access_denied, unsupported_grant_type, invalid_scope)

**Clean Architecture Checkpoint**:

- Domain layer established with entities and interfaces only
- Infrastructure layer provides concrete implementations
- No domain-to-infrastructure dependencies
- Dependency arrows point inward

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Client App Registration (Priority: P1) 🎯 MVP

**Goal**: Tenant administrators can register client applications with credentials and redirect URIs

**Independent Test**: Create tenant, register client with redirect URIs, verify client_id/client_secret generated

### Tests for User Story 1 (MANDATORY per constitution) ⚠️

> **CONSTITUTION REQUIREMENT: Write these tests FIRST, ensure they FAIL before implementation** > **Target: ≥85% coverage for domain/business logic, integration tests for all handlers**

- [ ] T045 [P] [US1] Unit test for Client entity validation in backend/tests/unit/domain/client_test.go (test ValidateRedirectURI, IsURIAllowed, HTTPS validation)
- [ ] T046 [P] [US1] Unit test for CreateClient use case in backend/tests/unit/usecase/create_client_test.go (mock ClientRepository, test tenant isolation, duplicate client_id handling)
- [ ] T047 [P] [US1] Unit test for GetClient use case in backend/tests/unit/usecase/get_client_test.go (mock ClientRepository, test not found scenarios)
- [ ] T048 [P] [US1] Unit test for UpdateClient use case in backend/tests/unit/usecase/update_client_test.go (mock ClientRepository, test redirect URI updates)
- [ ] T049 [P] [US1] Unit test for DeleteClient use case in backend/tests/unit/usecase/delete_client_test.go (mock ClientRepository, test soft delete)
- [ ] T050 [US1] Integration test for client registration in backend/tests/integration/client_management_test.go (real PostgreSQL via testcontainers, test POST /api/admin/clients, verify DB state, test redirect URI validation, test tenant isolation)
- [ ] T051 [US1] Frontend unit test for ClientManagement component in frontend/tests/unit/components/ClientManagement.test.tsx (test form submission, validation, client list rendering)

### Implementation for User Story 1

> **Clean Architecture Task Order**: Domain → Use Cases → Infrastructure → Interfaces (Handlers)

- [ ] T052 [US1] Implement CreateClient use case in backend/usecase/client/create_client.go (depends on T019, T024, accepts ClientRepository interface, generates client_id/client_secret, hashes secret with bcrypt, validates redirect URIs, enforces tenant isolation per FR-006)
- [ ] T053 [US1] Implement GetClient use case in backend/usecase/client/get_client.go (depends on T024, fetches by client_id, enforces tenant context)
- [ ] T054 [US1] Implement UpdateClient use case in backend/usecase/client/update_client.go (depends on T024, allows updating name and redirect URIs, prevents changing client_id)
- [ ] T055 [US1] Implement DeleteClient use case in backend/usecase/client/delete_client.go (depends on T024, soft delete by setting is_active=false)
- [ ] T056 [US1] Implement ListClients use case in backend/usecase/client/list_clients.go (depends on T024, lists all clients for tenant with pagination)
- [ ] T057 [US1] Implement RotateSecret use case in backend/usecase/client/rotate_secret.go (depends on T024, generates new client_secret, invalidates old one)
- [ ] T058 [US1] Implement ClientHandler in backend/interfaces/http/handlers/client_handler.go with routes: POST /api/admin/clients (create), GET /api/admin/clients (list), GET /api/admin/clients/:id (get), PUT /api/admin/clients/:id (update), DELETE /api/admin/clients/:id (delete), POST /api/admin/clients/:id/rotate-secret (rotate), requires admin authentication, extracts tenant from JWT
- [ ] T059 [US1] Add client management routes to backend/interfaces/http/router.go in /api/admin/\* group with auth middleware
- [ ] T060 [P] [US1] Create ClientManagement component in frontend/src/components/admin/ClientManagement.tsx (table view of clients, create/edit/delete buttons)
- [ ] T061 [P] [US1] Create ClientForm component in frontend/src/components/admin/ClientForm.tsx (form for name, redirect URIs with validation, displays generated credentials)
- [ ] T062 [US1] Create clientService API client in frontend/src/services/clientService.ts (axios wrapper for client CRUD endpoints)
- [ ] T063 [P] [US1] Create Client types in frontend/src/types/client.ts (TypeScript interfaces for Client, CreateClientRequest, UpdateClientRequest)

**Architecture Verification**:

- Domain (Client entity) has no infrastructure imports ✓
- Use cases depend only on ClientRepository interface ✓
- Handlers depend on use cases, not domain directly ✓

**Checkpoint**: At this point, User Story 1 should be fully functional - admins can register/manage clients

---

## Phase 4: User Story 2 - User Authentication Flow (Priority: P1) 🎯 MVP

**Goal**: End users authenticate via SSO, receive authorization code after successful login

**Independent Test**: Initiate OAuth authorization request, complete login, verify auth code returned to redirect URI

### Tests for User Story 2 (MANDATORY per constitution) ⚠️

- [ ] T064 [P] [US2] Unit test for AuthorizationCode entity in backend/tests/unit/domain/authorization_code_test.go (test expiration, single-use validation)
- [ ] T065 [P] [US2] Unit test for AuthorizeClient use case in backend/tests/unit/usecase/authorize_client_test.go (mock ClientRepository, UserRepository, RoleRepository, AuthCodeRepository, test PKCE validation, test redirect URI matching per FR-010, test role-based access control per FR-006d, FR-012)
- [ ] T066 [US2] Integration test for OAuth authorization flow in backend/tests/integration/oauth_auth_test.go (test GET /oauth2/auth with all required parameters per FR-009, test login form rendering, test POST with credentials, verify auth code generation, test redirect with code and state, test invalid client_id/redirect_uri per FR-017, FR-018, test user without role assignment denied per FR-006d)
- [ ] T067 [P] [US2] Frontend unit test for ConsentScreen component in frontend/tests/unit/components/ConsentScreen.test.tsx (test consent approval/denial, test scope display)
- [ ] T068 [P] [US2] Frontend unit test for useOAuth hook in frontend/tests/unit/hooks/useOAuth.test.ts (test authorization URL generation, test callback handling)

### Implementation for User Story 2

- [ ] T069 [US2] Implement AuthorizeClient use case in backend/usecase/auth/authorize_client.go (validates client_id per FR-017, validates redirect_uri per FR-010, checks user has role for client per FR-006d, FR-012, validates PKCE parameters per FR-009, creates session per FR-013, generates authorization code per FR-014, stores code_challenge per FR-015, returns redirect URI with code and state per FR-016)
- [ ] T070 [US2] Implement ConsentDecision use case in backend/usecase/auth/consent_decision.go (handles user consent approval/denial, generates auth code on approval)
- [ ] T071 [US2] Implement OAuthHandler in backend/interfaces/http/handlers/oauth_handler.go with GET /oauth2/auth endpoint (validates query params per FR-009, determines tenant from client_id per FR-010a, renders login form if unauthenticated, handles authentication, calls AuthorizeClient use case, returns 302 redirect with auth code per FR-016)
- [ ] T072 [US2] Add OAuth routes to backend/interfaces/http/router.go: GET /oauth2/auth, POST /oauth2/auth (for login submission), with tenant context middleware
- [ ] T073 [P] [US2] Create ConsentScreen component in frontend/src/components/auth/ConsentScreen.tsx (displays client name, requested scopes, approve/deny buttons, functional component with hooks)
- [ ] T074 [P] [US2] Create OAuthCallback component in frontend/src/components/auth/OAuthCallback.tsx (extracts code and state from URL, sends to parent app via postMessage or callback)
- [ ] T075 [P] [US2] Create useOAuth hook in frontend/src/hooks/useOAuth.ts (generates authorization URL with PKCE challenge, handles callback, stores code_verifier in sessionStorage)
- [ ] T076 [P] [US2] Create usePKCE hook in frontend/src/hooks/usePKCE.ts (generates code_verifier and code_challenge using Web Crypto API, implements S256 hashing)
- [ ] T077 [US2] Create oauthService API client in frontend/src/services/oauthService.ts (functions for buildAuthURL, handleCallback)
- [ ] T078 [P] [US2] Create OAuth types in frontend/src/types/oauth.ts (AuthorizationRequest, AuthorizationResponse, TokenResponse)
- [ ] T079 [P] [US2] Create PKCE utility in frontend/src/utils/pkce.ts (generateCodeVerifier, generateCodeChallenge, base64URLEncode)

**Checkpoint**: Users can successfully authenticate and receive authorization codes

---

## Phase 5: User Story 3 - Token Exchange (Priority: P1) 🎯 MVP

**Goal**: Client apps exchange authorization codes for JWT tokens with PKCE validation

**Independent Test**: Obtain auth code, exchange with code_verifier and client credentials, verify JWT tokens returned

### Tests for User Story 3 (MANDATORY per constitution) ⚠️

- [ ] T080 [P] [US3] Unit test for IssueToken use case in backend/tests/unit/usecase/issue_token_test.go (mock AuthCodeRepository, ClientRepository, RefreshTokenRepository, TokenService, test PKCE validation per FR-021, test authorization code expiration per FR-023, test one-time use per FR-024, test client credential validation per FR-025, test redirect_uri matching per FR-022)
- [ ] T081 [US3] Integration test for token exchange in backend/tests/integration/oauth_token_test.go (test POST /oauth2/token with grant_type=authorization_code, verify JWT structure per FR-029-FR-032, test PKCE failure returns invalid_grant per FR-028, test expired code returns invalid_grant, test code reuse returns invalid_grant per FR-024, verify access token 15min expiry per FR-033, verify refresh token 7day expiry per FR-035, verify RS256 signature per FR-036)

### Implementation for User Story 3

- [ ] T082 [US3] Implement IssueToken use case in backend/usecase/auth/issue_token.go (validates grant_type per FR-020, retrieves and validates auth code per FR-023, FR-024, validates PKCE code_verifier per FR-021, validates redirect_uri per FR-022, validates client credentials per FR-025, revokes auth code per FR-026, generates ID token per FR-030-FR-031, generates access token per FR-032-FR-033, generates refresh token per FR-034-FR-035, returns all three tokens per FR-027, returns invalid_grant on failure per FR-028)
- [ ] T083 [US3] Add POST /oauth2/token endpoint to backend/interfaces/http/handlers/oauth_handler.go (handles token exchange requests, applies rate limiting middleware per FR-051, calls IssueToken use case, returns JSON with id_token, access_token, refresh_token, token_type=Bearer, expires_in)
- [ ] T084 [US3] Update backend/interfaces/http/router.go to add POST /oauth2/token route with rate limiter middleware (10 req/min per client_id per FR-051)
- [ ] T085 [P] [US3] Add exchangeCodeForTokens function to frontend/src/services/oauthService.ts (POST to /oauth2/token with code, code_verifier, client_id, client_secret, redirect_uri)
- [ ] T086 [P] [US3] Create tokenStorage utility in frontend/src/utils/tokenStorage.ts (securely stores tokens in memory or secure storage, getAccessToken, getRefreshToken, clearTokens)

**Checkpoint**: Complete OAuth flow works end-to-end - users can authenticate and clients receive JWT tokens

---

## Phase 6: User Story 4 - Token Validation & Signature Verification (Priority: P2)

**Goal**: Clients can validate JWT signatures using public keys from JWKS endpoint

**Independent Test**: Obtain JWT, fetch JWKS, verify signature using public key

### Tests for User Story 4 (MANDATORY per constitution) ⚠️

- [ ] T087 [P] [US4] Unit test for ValidateToken use case in backend/tests/unit/usecase/validate_token_test.go (mock TokenService, test signature validation, test expiration validation, test tenant_id validation)
- [ ] T088 [US4] Integration test for JWKS discovery in backend/tests/integration/discovery_test.go (test GET /.well-known/openid-configuration returns OIDC metadata per FR-041, test GET /.well-known/jwks.json returns JWKS per FR-039, verify public key format, test signature validation using fetched public key)

### Implementation for User Story 4

- [ ] T089 [US4] Implement ValidateToken use case in backend/usecase/auth/validate_token.go (parses JWT, validates signature using TokenService per FR-039, validates expiration, validates tenant_id, validates audience)
- [ ] T090 [US4] Implement DiscoveryHandler in backend/interfaces/http/handlers/discovery_handler.go with GET /.well-known/openid-configuration endpoint (returns OIDC discovery metadata per FR-041: issuer, authorization_endpoint, token_endpoint, jwks_uri, userinfo_endpoint, supported scopes, response_types, grant_types, token_endpoint_auth_methods, code_challenge_methods_supported=[S256])
- [ ] T091 [US4] Add GET /.well-known/jwks.json endpoint to DiscoveryHandler (returns JWKS per FR-039, includes all active public keys per FR-040, format: {"keys": [{"kty": "RSA", "kid": "...", "use": "sig", "alg": "RS256", "n": "...", "e": "..."}]})
- [ ] T092 [US4] Add discovery routes to backend/interfaces/http/router.go: GET /.well-known/openid-configuration, GET /.well-known/jwks.json (public routes, no auth required)

**Checkpoint**: OIDC discovery works, clients can validate JWT signatures offline

---

## Phase 7: User Story 5 - Token Refresh (Priority: P2)

**Goal**: Clients refresh expired access tokens using refresh tokens without re-authentication

**Independent Test**: Obtain refresh token, exchange for new access token, verify old refresh token optionally rotated

### Tests for User Story 5 (MANDATORY per constitution) ⚠️

- [ ] T093 [P] [US5] Unit test for RefreshToken use case in backend/tests/unit/usecase/refresh_token_test.go (mock RefreshTokenRepository, ClientRepository, TokenService, test revoked token rejection per FR-046, test client_id mismatch rejection per FR-047, test expired token handling)
- [ ] T094 [US5] Integration test for token refresh in backend/tests/integration/oauth_refresh_test.go (test POST /oauth2/token with grant_type=refresh_token, verify new access token returned per FR-043, test invalid_grant on revoked token per FR-046, test client_id validation per FR-047, verify optional refresh token rotation)

### Implementation for User Story 5

- [ ] T095 [US5] Implement RefreshToken use case in backend/usecase/auth/refresh_token.go (validates grant_type=refresh_token, retrieves refresh token from database, validates not revoked per FR-046, validates not expired, validates client_id matches per FR-047, generates new access token per FR-043, optionally rotates refresh token, returns new tokens)
- [ ] T096 [US5] Update POST /oauth2/token endpoint in backend/interfaces/http/handlers/oauth_handler.go to handle grant_type=refresh_token (calls RefreshToken use case, returns new access token and optionally new refresh token)
- [ ] T097 [P] [US5] Add refreshAccessToken function to frontend/src/services/oauthService.ts (POST to /oauth2/token with grant_type=refresh_token, refresh_token, client_id, client_secret)
- [ ] T098 [P] [US5] Update tokenStorage utility to implement automatic token refresh logic (intercepts 401 responses, refreshes token, retries original request)

**Checkpoint**: Token refresh works, sessions can be maintained without re-authentication

---

## Phase 8: User Story 6 - Token Revocation (Priority: P3)

**Goal**: Administrators can revoke refresh tokens to terminate sessions

**Independent Test**: Revoke refresh token via admin portal or revocation endpoint, verify subsequent use fails

### Tests for User Story 6 (MANDATORY per constitution) ⚠️

- [ ] T099 [P] [US6] Unit test for RevokeToken use case in backend/tests/unit/usecase/revoke_token_test.go (mock RefreshTokenRepository, test token revocation, test cascade revocation for user-client pair)
- [ ] T100 [US6] Integration test for token revocation in backend/tests/integration/oauth_revoke_test.go (test POST /oauth2/revoke, verify token marked as revoked per FR-048, verify subsequent refresh fails per FR-050, test admin revocation via admin portal per FR-049)

### Implementation for User Story 6

- [ ] T101 [US6] Implement RevokeToken use case in backend/usecase/auth/revoke_token.go (marks refresh token as revoked per FR-048, optionally revokes all tokens for user-client pair, records revocation timestamp)
- [ ] T102 [US6] Add POST /oauth2/revoke endpoint to backend/interfaces/http/handlers/oauth_handler.go (implements FR-051: revocation endpoint, accepts token and token_type_hint parameters, calls RevokeToken use case, returns 200 OK per RFC 7009)
- [ ] T103 [US6] Add token revocation to admin portal in existing or new handler (implements FR-052: admin revocation, allows admins to revoke user sessions, lists active refresh tokens, provides revoke button)
- [ ] T104 [US6] Update backend/interfaces/http/router.go to add POST /oauth2/revoke route
- [ ] T105 [P] [US6] Add revokeToken function to frontend/src/services/oauthService.ts (POST to /oauth2/revoke with token parameter)

**Checkpoint**: Token revocation works, administrators can terminate sessions remotely

---

## Phase 9: User Story 8 - User Role Management (Priority: P2)

**Goal**: Administrators control user access to client apps via role assignments

**Independent Test**: Assign role to user for client, verify authentication succeeds; revoke role, verify authentication denied

### Tests for User Story 8 (MANDATORY per constitution) ⚠️

- [ ] T106 [P] [US8] Unit test for UserRole entity in backend/tests/unit/domain/user_role_test.go (test role validation, test active/inactive states)
- [ ] T107 [P] [US8] Unit test for AssignRole use case in backend/tests/unit/usecase/assign_role_test.go (mock RoleRepository, test role assignment, test duplicate prevention)
- [ ] T108 [P] [US8] Unit test for RevokeRole use case in backend/tests/unit/usecase/revoke_role_test.go (mock RoleRepository, RefreshTokenRepository, test role revocation, test cascade refresh token revocation per FR-006e)
- [ ] T109 [US8] Integration test for role management in backend/tests/integration/role_management_test.go (test POST /api/admin/roles/assign, test POST /api/admin/roles/revoke, test GET /api/admin/roles/users/:userId, verify authentication denied without role per FR-006d, verify refresh tokens revoked on role revocation per FR-006e)
- [ ] T110 [P] [US8] Frontend unit test for RoleManagement component in frontend/tests/unit/components/RoleManagement.test.tsx

### Implementation for User Story 8

- [ ] T111 [US8] Implement AssignRole use case in backend/usecase/role/assign_role.go (validates user exists, validates client exists, creates role assignment per FR-006a, prevents duplicates, records assigned_by)
- [ ] T112 [US8] Implement RevokeRole use case in backend/usecase/role/revoke_role.go (marks role as inactive per FR-006b, revokes all refresh tokens for user-client pair per FR-006e, records revocation)
- [ ] T113 [US8] Implement ListUserRoles use case in backend/usecase/role/list_user_roles.go (returns all active roles for user per FR-006b, supports filtering by client)
- [ ] T114 [US8] Implement RoleHandler in backend/interfaces/http/handlers/role_handler.go with routes: POST /api/admin/roles/assign (assign role per FR-006a), POST /api/admin/roles/revoke (revoke role per FR-006b), GET /api/admin/roles/users/:userId (list user roles), GET /api/admin/roles/clients/:clientId (list client roles), requires admin authentication
- [ ] T115 [US8] Add role management routes to backend/interfaces/http/router.go in /api/admin/\* group
- [ ] T116 [US8] Update AuthorizeClient use case (T069) to check user has active role for client per FR-006d, FR-012 (if not already implemented)
- [ ] T117 [P] [US8] Create RoleManagement component in frontend/src/components/admin/RoleManagement.tsx (displays user-client role matrix, assign/revoke buttons)
- [ ] T118 [P] [US8] Create UserRoles component in frontend/src/components/admin/UserRoles.tsx (shows roles for specific user across all clients)
- [ ] T119 [US8] Create roleService API client in frontend/src/services/roleService.ts (assignRole, revokeRole, listUserRoles, listClientRoles)
- [ ] T120 [P] [US8] Create Role types in frontend/src/types/role.ts (UserRole, AssignRoleRequest, RevokeRoleRequest)

**Checkpoint**: RBAC works, administrators can control user access to clients

---

## Phase 10: User Story 7 - Multi-Client Management (Priority: P3)

**Goal**: Tenants manage multiple independent client applications with separate configurations

**Independent Test**: Register multiple clients for one tenant, verify independent credentials and configs

### Tests for User Story 7 (MANDATORY per constitution) ⚠️

- [ ] T121 [US7] Integration test for multi-client scenarios in backend/tests/integration/multi_client_test.go (create multiple clients for one tenant, verify independent credentials per FR-003, verify redirect URI updates affect only one client, test credential rotation per FR-005, verify tokens issued for Client A cannot be used with Client B, test client isolation)

### Implementation for User Story 7

- [ ] T122 [US7] Enhance ClientManagement component in frontend/src/components/admin/ClientManagement.tsx to support bulk operations (multi-select, bulk activate/deactivate)
- [ ] T123 [US7] Add client cloning functionality to CreateClient use case (copy configuration from existing client, generate new credentials)
- [ ] T124 [US7] Add client usage analytics to backend (track authentication attempts, token issuances per client)
- [ ] T125 [P] [US7] Add client metrics display to ClientManagement component (show active users, recent authentications per client)

**Checkpoint**: Multi-client management fully functional, tenants can manage complex app ecosystems

---

## Phase 11: Additional Endpoints & Features

**Purpose**: Complete remaining endpoints and cross-cutting concerns

### UserInfo Endpoint

- [ ] T126 [P] Unit test for GetUserInfo use case in backend/tests/unit/usecase/get_userinfo_test.go
- [ ] T127 Implement GetUserInfo use case in backend/usecase/auth/get_userinfo.go (extracts user ID from access token, returns user profile claims)
- [ ] T128 Implement UserinfoHandler in backend/interfaces/http/handlers/userinfo_handler.go with GET /oauth2/userinfo endpoint (requires valid access token, returns user profile per FR-052)
- [ ] T129 Add GET /oauth2/userinfo route to backend/interfaces/http/router.go with JWT validation middleware

### Session Management

- [ ] T130 Implement session cleanup job in backend/cmd/cleanup/main.go (cron job to delete expired sessions, expired refresh tokens, used authorization codes)

### Rate Limiting

- [ ] T131 Integration test for rate limiting in backend/tests/integration/rate_limit_test.go (send >10 requests per minute to token endpoint per FR-051, verify HTTP 429 returned, verify rate limit headers)

---

## Phase 12: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T132 [P] Add API documentation comments to all exported functions per Effective Go (domain entities, use cases, handlers)
- [ ] T133 [P] Add TypeScript JSDoc comments to all exported functions and components (services, hooks, components)
- [ ] T134 [P] Run golangci-lint on backend code and fix issues
- [ ] T135 [P] Run ESLint on frontend code and fix issues
- [ ] T136 Verify ≥85% test coverage for backend domain + usecase layers (go test -coverprofile=coverage.out ./domain/... ./usecase/...; go tool cover -func=coverage.out)
- [ ] T137 [P] Verify frontend component test coverage (npm run test:coverage)
- [ ] T138 Run security audit: verify no plain-text secrets in database per FR-056, verify HTTPS-only cookies per FR-058, verify PKCE mandatory per FR-008
- [ ] T139 [P] Architecture compliance review: verify no domain-to-infrastructure imports (grep -r "infrastructure" backend/domain/ backend/usecase/)
- [ ] T140 Performance testing: verify authorization endpoint <100ms p95, token endpoint <200ms p95 (use Apache Bench or k6)
- [ ] T141 [P] Run quickstart.md validation: follow setup instructions, verify all components start successfully
- [ ] T142 Create production deployment checklist based on quickstart.md "Production Deployment Checklist" section
- [ ] T143 [P] Update repository README.md with OAuth feature overview, links to spec and quickstart
- [ ] T144 Generate OpenAPI client code for frontend (npm install -D openapi-typescript; openapi-typescript specs/003-sso-auth-provider/contracts/openapi.yaml -o frontend/src/types/api.ts)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-10)**: All depend on Foundational phase completion
  - US1 (Client Registration): Can start after Foundational - No dependencies on other stories
  - US2 (Authentication Flow): Can start after Foundational - No dependencies on other stories
  - US3 (Token Exchange): Depends on US2 (needs authorization codes)
  - US4 (Token Validation): Depends on US3 (needs JWT tokens)
  - US5 (Token Refresh): Depends on US3 (needs refresh tokens)
  - US6 (Token Revocation): Depends on US5 (revokes refresh tokens)
  - US8 (Role Management): Can start after Foundational - integrates with US2 authentication checks
  - US7 (Multi-Client): Extends US1, can start after US1 complete
- **Additional Features (Phase 11)**: Depends on US3 (needs tokens)
- **Polish (Phase 12)**: Depends on all desired user stories being complete

### Critical Path for MVP

**Minimum viable product** (most basic OAuth flow):

1. Phase 1: Setup → Phase 2: Foundational → Phase 3: US1 (Client Registration) → Phase 4: US2 (Authentication) → Phase 5: US3 (Token Exchange)

This delivers: Client registration + user authentication + token issuance

**Production-ready** adds:

- Phase 6: US4 (Token Validation) - for distributed validation
- Phase 7: US5 (Token Refresh) - for long sessions
- Phase 9: US8 (Role Management) - for access control
- Phase 8: US6 (Token Revocation) - for security

### Parallel Opportunities

- Within Phase 2 (Foundational): All database migrations [P], all domain entities [P], all repository interfaces [P], all repository implementations [P]
- Within each User Story: All unit tests [P] can run in parallel
- User Stories can proceed in parallel IF teams work on different stories:
  - US1 (Client Registration) + US2 (Authentication) + US8 (Role Management) can run in parallel (different files)
  - US3-US6 must be sequential (token dependencies)

---

## Implementation Strategy

### MVP First (US1 + US2 + US3 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (**CRITICAL** - blocks all stories)
3. Complete Phase 3: US1 (Client Registration)
4. Complete Phase 4: US2 (Authentication Flow)
5. Complete Phase 5: US3 (Token Exchange)
6. **STOP and VALIDATE**: Test complete OAuth flow end-to-end
7. Deploy/demo if ready

### Incremental Delivery (Recommended)

1. MVP (US1 + US2 + US3) → Test → Deploy
2. Add US4 (Token Validation) → Test → Deploy (enables distributed validation)
3. Add US8 (Role Management) → Test → Deploy (enables RBAC)
4. Add US5 (Token Refresh) → Test → Deploy (enables long sessions)
5. Add US6 (Token Revocation) → Test → Deploy (security hardening)
6. Add US7 (Multi-Client) → Test → Deploy (advanced scenarios)

Each increment delivers production-usable value.

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: US1 (Client Registration)
   - Developer B: US2 (Authentication Flow)
   - Developer C: US8 (Role Management)
3. After US2 complete:
   - Developer D: US3 (Token Exchange) - depends on US2
4. After US3 complete:
   - Developer E: US4, US5, US6 in sequence

---

## Notes

- **PKCE is mandatory** per FR-008 - all authorization requests require code_challenge
- **RS256 signing** per FR-036 - asymmetric keys for signature verification
- **Rate limiting** per FR-051 - 10 requests/minute per client_id on token endpoint
- **Token expiration**: Access tokens 15min per FR-033, Refresh tokens 7 days per FR-035
- **Clean Architecture**: All use cases accept repository interfaces, never concrete types
- **Test-first**: Write tests, verify they fail, implement code, verify tests pass
- **Constitution compliance**: ≥85% coverage for domain/usecase, integration tests for all handlers
- [P] tasks = different files, no dependencies, can run in parallel
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
