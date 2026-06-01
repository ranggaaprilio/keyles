# Quickstart: OAuth Consent Flow and End-User Authentication

**Feature**: [spec.md](spec.md)  
**Date**: June 1, 2026

## Prerequisites

- Docker and Docker Compose
- A seeded tenant, active end-user, client application, and user-client role
- A PKCE-capable test client with a registered callback URI

## Environment

Add the browser-flow configuration to the backend environment:

```dotenv
FRONTEND_URL=http://localhost:3000
SECURITY_COOKIE_SECURE=false
SECURITY_SESSION_TTL=28800
OAUTH_AUTH_TRANSACTION_TTL=600
RATE_LIMIT_OAUTH_LOGIN_FAILURES=5
RATE_LIMIT_OAUTH_LOGIN_WINDOW_SECONDS=900
```

Keep `SECURITY_COOKIE_SECURE=true` in deployed HTTPS environments.
The SSO cookie is host-only with `Path=/`; do not configure a `Domain` attribute.
OAuth login throttling and audit events use the backend's direct TCP peer address.
Do not expose the backend through a reverse proxy that rewrites client identity until
an explicit trusted-proxy allowlist is implemented.

## Start Services

```bash
docker compose up --build
```

The backend is served at `http://localhost:8080` and the frontend at
`http://localhost:3000`.

## Manual PKCE Flow

1. Generate a PKCE verifier and S256 challenge in the test client.
2. Navigate the browser to:

```text
http://localhost:8080/oauth2/auth?client_id={client_id}&redirect_uri={encoded_callback}&response_type=code&scope=openid%20profile%20email&state={random_state}&code_challenge={s256_challenge}&code_challenge_method=S256
```

3. Verify redirect to:

```text
http://localhost:3000/oauth2/login?transaction_id=...
```

4. Log in as the seeded end-user and verify redirect to:

```text
http://localhost:3000/oauth2/consent?transaction_id=...
```

5. Approve consent and verify the registered callback receives `code` and the
   original `state`.
6. Exchange the code:

```bash
curl -X POST http://localhost:8080/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'grant_type=authorization_code' \
  --data-urlencode 'client_id={client_id}' \
  --data-urlencode 'code={authorization_code}' \
  --data-urlencode 'redirect_uri={registered_callback}' \
  --data-urlencode 'code_verifier={pkce_verifier}'
```

For confidential clients, also provide `client_secret`.

## Verification Matrix

| Scenario | Expected result |
| --- | --- |
| Approve consent | Callback receives `code` and original `state` |
| Deny consent | Callback receives `error=access_denied` and original `state` |
| Invalid callback URI | Keyles local error page; no callback redirect |
| Reuse valid session across clients in the same tenant | Skip login and display consent when the user remains active and authorized |
| Reuse session after user disable or role removal | Require login; never continue silently to consent |
| `prompt=login` | Force login despite valid session |
| `max_age=0` | Force login |
| `prompt=none` without session | Callback receives `login_required` |
| `prompt=none` with session and no saved grant | Callback receives `consent_required` |
| Replay consent POST | Rejected; no second code |
| Wrong PKCE verifier | Token endpoint returns `invalid_grant` |
| Five failed logins for either source IP or tenant email | Further verification is throttled for that fixed 15-minute bucket |
| Spoofed forwarded-IP headers | Direct TCP peer remains the throttle and audit source IP |
| Credentialed frontend-origin request | CORS returns the configured origin and allows credentials without wildcard origin |
| `POST /oauth2/logout` | Redis session is deleted and browser cookie expires |
| Redis unavailable during auth, login, or consent | Local `temporarily_unavailable` error; no external callback redirect or code |
| Redis unavailable during logout | Browser cookie still expires |

## Test Commands

```bash
cd backend
make test
make test-compose-e2e

cd ../frontend
npm run test -- --run
npm run lint
npm run build
```
