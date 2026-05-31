# Epic 01: OAuth Consent Flow & End-User Authentication

## Goal

Make the OAuth 2.0 / OIDC authorization flow work end-to-end by wiring the consent screen into the frontend, adding an end-user login page for the SSO provider, and connecting the full redirect chain so external applications can actually authenticate users through Keyles.

## Why MVP

The SSO provider is the core product. Without a working consent and login flow, no external application can authenticate users through Keyles. This is the single biggest gap — the OAuth endpoints exist on the backend and the ConsentScreen component exists on the frontend, but they are not connected. When an application redirects to `/oauth2/auth`, there is no visible UI for the end-user to log in and grant consent.

## Current State

- **Backend**: `/oauth2/auth` (GET/POST) handler exists in `oauth_handler.go`. `AuthorizeClient` use case validates the request. `ConsentDecision` use case processes approve/deny. OIDC discovery and JWKS endpoints work. Token exchange, refresh, revocation, and introspection all work.
- **Frontend**: `ConsentScreen.tsx` component is fully built with scope descriptions, approve/deny buttons, and client info display. `OAuthCallback.tsx` handles the post-authorization redirect. `OAuthService` class manages PKCE, state, and nonce generation. Types are defined in `types/oauth.ts`.
- **Missing**: No `/oauth2/consent` or `/oauth2/login` route in `App.tsx`. The backend's `Authorize` handler redirects or returns JSON but there is no frontend route to intercept and render the consent UI. No end-user (non-admin) login page exists — the only login page is for tenant admins.

## Tasks

### 1.1 — Add end-user login page for OAuth flow
- Create `frontend/src/pages/OAuthLoginPage.tsx` — a login form for end-users (not admins) that authenticates against the SSO provider during OAuth flow
- End-user credentials are email + password, validated via an existing or new backend endpoint
- After successful login, redirect to consent screen with the original OAuth params preserved

### 1.2 — Add OAuth consent route to frontend
- Add route `/oauth2/consent` in `App.tsx` that renders `ConsentScreen`
- Extract OAuth params (client_id, redirect_uri, scope, state, code_challenge, etc.) from URL query string
- Display client info (name, logo) and requested scopes with descriptions
- On approve: POST consent decision to backend, redirect to client's redirect_uri with auth code
- On deny: POST deny decision to backend, redirect to client's redirect_uri with error

### 1.3 — Add OAuth login redirect route
- Add route `/oauth2/login` in `App.tsx` that renders `OAuthLoginPage`
- When an unauthenticated end-user hits `/oauth2/auth`, backend redirects to `/oauth2/login` with all OAuth params preserved
- After login, redirect to `/oauth2/consent` with the same params

### 1.4 — Wire backend authorize redirect to frontend routes
- Backend `Authorize` handler: when user is not authenticated, redirect to `FRONTEND_URL/oauth2/login?<params>` instead of returning an error
- Backend `Authorize` handler: when user is authenticated but consent not given, redirect to `FRONTEND_URL/oauth2/consent?<params>`
- Add `FRONTEND_URL` environment variable to backend config

### 1.5 — Add consent-decision API endpoint
- Create backend endpoint `POST /oauth2/consent` that processes user's approve/deny decision
- On approve: generate authorization code, redirect to client's redirect_uri
- On deny: redirect to client's redirect_uri with `error=access_denied`
- Validate CSRF state param

### 1.6 — Add end-user session management for OAuth
- End-users get a session cookie (not JWT) after login on the SSO provider
- Session is separate from admin JWT auth
- Support `prompt=login` to force re-authentication
- Support `prompt=consent` to force re-consent
- Support `max_age` parameter to enforce session age limits

### 1.7 — Add OAuth error pages
- Create error display pages for common OAuth errors (invalid_client, invalid_request, access_denied, etc.)
- Show user-friendly messages with the error description

## Acceptance Criteria

1. An external OAuth client (e.g., a test app) can redirect to Keyles `/oauth2/auth` with valid params
2. An unauthenticated end-user is redirected to a login page, logs in, and lands on the consent screen
3. The consent screen shows the client name and requested scopes with descriptions
4. Approving redirects back to the client with a valid authorization code
5. Denying redirects back to the client with `error=access_denied`
6. The client can exchange the auth code for tokens at `/oauth2/token`
7. PKCE is enforced (S256 required)
8. `prompt`, `max_age`, and `state` params are handled correctly
9. The full flow works end-to-end in Docker Compose
