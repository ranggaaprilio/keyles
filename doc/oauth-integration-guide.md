# Keyles — OAuth 2.0 / OpenID Connect Integration Guide

How to integrate your application with the Keyles SSO platform as an
OAuth 2.0 relying party (client application).

## Overview

Keyles is a multi-tenant OAuth 2.0 Authorization Server with OpenID
Connect support. It implements the Authorization Code Flow with PKCE
(RFC 7636), issues RS256-signed JWTs, and exposes OIDC discovery
and JWKS endpoints for standards-compliant integration.

**Endpoints** (development):

| Endpoint | URL |
|---|---|
| Authorization | `http://localhost:8080/oauth2/auth` |
| Token | `http://localhost:8080/oauth2/token` |
| Revocation | `http://localhost:8080/oauth2/revoke` |
| Introspection | `http://localhost:8080/oauth2/introspect` |
| UserInfo | `http://localhost:8080/oauth2/userinfo` |
| OIDC Discovery | `http://localhost:8080/.well-known/openid-configuration` |
| JWKS | `http://localhost:8080/.well-known/jwks.json` |

## 1. Register Your OAuth Client

Before integrating, register your application as an OAuth client through
the Keyles admin portal (`/admin/clients`) or the admin API:

```bash
curl -X POST http://localhost:8080/api/v1/admin/clients \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin_jwt>" \
  -d '{
    "client_name": "My Application",
    "redirect_uris": ["https://myapp.example.com/auth/callback"],
    "type": "public"
  }'
```

You will receive a `client_id` and `client_secret`. Confidential clients
(server-side web apps) should store the secret securely and never expose
it in client-side code. Public clients (SPAs, mobile apps) must use PKCE
and do not authenticate with a secret at the token endpoint.

## 2. Authorization Code Flow with PKCE

Keyles requires S256 PKCE for all authorization code grants. The flow
has six steps:

```
┌──────────┐     ┌──────────┐     ┌──────────┐
│  Your App │     │  Browser │     │  Keyles  │
└────┬─────┘     └────┬─────┘     └────┬─────┘
     │                 │               │
     │ 1. Generate PKCE verifier & challenge
     │                 │               │
     │ 2. Redirect to /oauth2/auth     │
     │────────────────────────────────>│
     │                 │               │
     │                 │ 3. User logs  │
     │                 │    in &       │
     │                 │    consents   │
     │                 │               │
     │ 4. Redirect to callback with code│
     │<────────────────────────────────│
     │                 │               │
     │ 5. POST /oauth2/token           │
     │   (code + code_verifier)        │
     │────────────────────────────────>│
     │                 │               │
     │ 6. Receive access_token,        │
     │    refresh_token, id_token      │
     │<────────────────────────────────│
     │                 │               │
```

### Step 1: Generate PKCE Parameters

Generate a cryptographically random code verifier (43–128 characters)
and compute its S256 challenge:

```typescript
// Generate code_verifier (64 chars, URL-safe random)
function generateCodeVerifier(): string {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~';
  const array = new Uint8Array(64);
  crypto.getRandomValues(array);
  return Array.from(array, b => chars[b % chars.length]).join('');
}

// Compute S256 code_challenge
async function generateCodeChallenge(verifier: string): Promise<string> {
  const encoder = new TextEncoder();
  const digest = await crypto.subtle.digest('SHA-256', encoder.encode(verifier));
  return btoa(String.fromCharCode(...new Uint8Array(digest)))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '');
}

const codeVerifier = generateCodeVerifier();
const codeChallenge = await generateCodeChallenge(codeVerifier);
const state = crypto.randomUUID(); // CSRF protection

// Store code_verifier and state for later (never in the URL)
sessionStorage.setItem('pkce_verifier', codeVerifier);
sessionStorage.setItem('oauth_state', state);
```

### Step 2: Redirect to Authorization Endpoint

Build the authorization URL and redirect the user's browser:

```
GET /oauth2/auth
  ?client_id=<your_client_id>
  &redirect_uri=<your_registered_callback>
  &response_type=code
  &scope=openid%20profile%20email
  &state=<random_state>
  &code_challenge=<s256_challenge>
  &code_challenge_method=S256
```

```typescript
const authURL = new URL('http://localhost:8080/oauth2/auth');
authURL.searchParams.set('client_id', clientId);
authURL.searchParams.set('redirect_uri', redirectUri);
authURL.searchParams.set('response_type', 'code');
authURL.searchParams.set('scope', 'openid profile email');
authURL.searchParams.set('state', state);
authURL.searchParams.set('code_challenge', codeChallenge);
authURL.searchParams.set('code_challenge_method', 'S256');

window.location.href = authURL.toString();
```

**Optional OIDC parameters**:

| Parameter | Description |
|---|---|
| `nonce` | Opaque value bound to the ID token for replay protection |
| `prompt` | Space-delimited: `login`, `consent`, `none`, `select_account` |
| `max_age` | Maximum allowable authentication age in seconds |

### Step 3: User Authenticates and Consents

Keyles handles this step entirely. The user:

1. Logs in with their Keyles credentials (or reuses an existing session)
2. Reviews the consent screen showing your application name and requested scopes
3. Approves or denies the request

Your application does not interact with this step.

### Step 4: Handle the Callback

Keyles redirects the browser back to your registered `redirect_uri`:

```
GET https://myapp.example.com/auth/callback
  ?code=<authorization_code>
  &state=<your_original_state>
```

**Validate the state** to prevent CSRF attacks:

```typescript
const url = new URL(window.location.href);
const code = url.searchParams.get('code');
const returnedState = url.searchParams.get('state');

const storedState = sessionStorage.getItem('oauth_state');
if (returnedState !== storedState) {
  throw new Error('State mismatch — possible CSRF attack');
}

// Clear state now
sessionStorage.removeItem('oauth_state');
```

**Error callbacks** — if the user denies consent or an error occurs:

```
GET https://myapp.example.com/auth/callback
  ?error=access_denied
  &error_description=The+user+denied+the+request
  &state=<your_original_state>
```

### Step 5: Exchange the Authorization Code

Send the code and PKCE verifier to the token endpoint. The code is
single-use and expires after 5 minutes.

```
POST /oauth2/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code
&code=<authorization_code>
&redirect_uri=<same_redirect_uri>
&client_id=<your_client_id>
&code_verifier=<pkce_verifier>
&client_secret=<client_secret>   # For confidential clients only
```

```typescript
const codeVerifier = sessionStorage.getItem('pkce_verifier')!;

const params = new URLSearchParams();
params.set('grant_type', 'authorization_code');
params.set('code', code);
params.set('redirect_uri', redirectUri);
params.set('client_id', clientId);
params.set('code_verifier', codeVerifier);
// For confidential clients:
// params.set('client_secret', clientSecret);

const response = await fetch('http://localhost:8080/oauth2/token', {
  method: 'POST',
  headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
  body: params,
});

const tokens = await response.json();
sessionStorage.removeItem('pkce_verifier');
```

### Step 6: Receive Tokens

A successful token response:

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 900,
  "refresh_token": "8xJ3kL9mN2pQ...",
  "id_token": "eyJhbGciOiJSUzI1NiIs...",
  "scope": "openid profile email"
}
```

| Field | Description |
|---|---|
| `access_token` | RS256 JWT; use as `Authorization: Bearer <token>` |
| `token_type` | Always `Bearer` |
| `expires_in` | Seconds until expiry (default: 900 / 15 minutes) |
| `refresh_token` | Long-lived token for obtaining new access tokens (7 days) |
| `id_token` | OIDC identity token with user claims (only when `openid` scope is requested) |

## 3. Validate the ID Token

ID tokens are RS256-signed JWTs. Validate them before trusting claims:

1. Fetch the JWKS document from `/.well-known/jwks.json`
2. Find the key matching the token's `kid` header
3. Verify the RS256 signature using the public key
4. Validate claims:

| Claim | Rule |
|---|---|
| `iss` | Must match the issuer URL exactly |
| `aud` | Must include your `client_id` |
| `exp` | Must be in the future |
| `iat` | Should be in the recent past |
| `nonce` | If sent in the auth request, must match exactly |
| `auth_time` | If present, should be recent |

```typescript
// Fetch JWKS
const discovery = await fetch(
  'http://localhost:8080/.well-known/openid-configuration'
).then(r => r.json());

const jwks = await fetch(discovery.jwks_uri).then(r => r.json());

// Decode JWT header to find kid, then locate the matching key
const header = JSON.parse(atob(idToken.split('.')[0]));
const key = jwks.keys.find((k: any) => k.kid === header.kid);

// Verify signature with Web Crypto API
const publicKey = await crypto.subtle.importKey(
  'jwk', key,
  { name: 'RSASSA-PKCS1-v1_5', hash: 'SHA-256' },
  false, ['verify']
);

const [rawHeader, rawPayload, rawSignature] = idToken.split('.');
const signedData = new TextEncoder().encode(`${rawHeader}.${rawPayload}`);
const signature = base64URLDecode(rawSignature);

const valid = await crypto.subtle.verify(
  'RSASSA-PKCS1-v1_5', publicKey, signature, signedData
);

// Then validate claims from the decoded payload
const payload = JSON.parse(atob(rawPayload));
```

## 4. Use the Access Token

Include the access token in API requests via the `Authorization` header:

```bash
curl -H "Authorization: Bearer <access_token>" \
  http://localhost:8080/oauth2/userinfo
```

### UserInfo Endpoint

Returns claims about the authenticated end-user:

```
GET /oauth2/userinfo
Authorization: Bearer <access_token>
```

Response:

```json
{
  "sub": "user-uuid",
  "email": "user@example.com",
  "email_verified": true,
  "name": "Jane Doe",
  "preferred_username": "jane.doe"
}
```

## 5. Refresh Access Tokens

When an access token expires, use the refresh token to obtain a new one
without requiring the user to re-authenticate:

```
POST /oauth2/token
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token
&refresh_token=<refresh_token>
&client_id=<your_client_id>
&client_secret=<client_secret>   # For confidential clients only
```

```typescript
async function refreshAccessToken(refreshToken: string): Promise<TokenResponse> {
  const params = new URLSearchParams();
  params.set('grant_type', 'refresh_token');
  params.set('refresh_token', refreshToken);
  params.set('client_id', clientId);

  const response = await fetch('http://localhost:8080/oauth2/token', {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: params,
  });

  if (!response.ok) {
    // Refresh token expired or revoked — redirect to login
    throw new Error('Token refresh failed');
  }

  return response.json();
}
```

The response includes a new access token and a **new refresh token**.
Replace both stored tokens. The old refresh token is invalidated
(rotation).

## 6. Revoke Tokens

Revoke a refresh token or access token when the user logs out:

```
POST /oauth2/revoke
Content-Type: application/x-www-form-urlencoded

token=<token_to_revoke>
&token_type_hint=refresh_token
&client_id=<your_client_id>
&client_secret=<client_secret>
```

The endpoint returns `200 OK` regardless of whether the token was valid
(RFC 7009). This prevents information leakage.

```typescript
await fetch('http://localhost:8080/oauth2/revoke', {
  method: 'POST',
  headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
  body: new URLSearchParams({
    token: refreshToken,
    token_type_hint: 'refresh_token',
    client_id: clientId,
  }),
});
```

## 7. Introspect Tokens (Server-Side)

For resource servers that need to validate an access token server-side,
use the introspection endpoint. This requires client authentication:

```
POST /oauth2/introspect
Content-Type: application/x-www-form-urlencoded

token=<access_token>
&client_id=<your_client_id>
&client_secret=<your_client_secret>
```

Response (active token):

```json
{
  "active": true,
  "client_id": "my_client",
  "sub": "user-uuid",
  "scope": "openid profile email",
  "exp": 1717200000,
  "iat": 1717199100,
  "token_type": "Bearer"
}
```

Response (inactive/revoked/expired):

```json
{
  "active": false
}
```

## 8. OIDC Discovery

Fetch the OIDC discovery document for dynamic endpoint resolution:

```
GET /.well-known/openid-configuration
```

```json
{
  "issuer": "http://localhost:8080",
  "authorization_endpoint": "http://localhost:8080/oauth2/auth",
  "token_endpoint": "http://localhost:8080/oauth2/token",
  "jwks_uri": "http://localhost:8080/.well-known/jwks.json",
  "userinfo_endpoint": "http://localhost:8080/oauth2/userinfo",
  "revocation_endpoint": "http://localhost:8080/oauth2/revoke",
  "introspection_endpoint": "http://localhost:8080/oauth2/introspect",
  "scopes_supported": ["openid", "profile", "email"],
  "response_types_supported": ["code"],
  "grant_types_supported": ["authorization_code", "refresh_token"],
  "code_challenge_methods_supported": ["S256"],
  "token_endpoint_auth_methods_supported": ["client_secret_post"]
}
```

## 9. OAuth Client Types

Keyles supports two client types:

| Type | Secret | Token Auth | Use Case |
|---|---|---|---|
| **Public** | No secret | `client_id` only (PKCE required) | SPAs, mobile apps |
| **Confidential** | Has secret | `client_id` + `client_secret` | Server-side web apps, backend services |

Public clients must always use PKCE. Confidential clients should also use
PKCE for defense in depth.

## 10. Error Handling

### Authorization Errors

Returned to your `redirect_uri` as query parameters:

| Error Code | Meaning |
|---|---|
| `invalid_request` | Malformed or missing required parameters |
| `invalid_client` | Client ID is unknown |
| `unauthorized_client` | Client is not authorized for this grant type |
| `access_denied` | User denied the consent request |
| `unsupported_response_type` | Only `code` is supported |
| `login_required` | `prompt=none` was used but no session exists |
| `consent_required` | `prompt=none` was used but consent is required |
| `temporarily_unavailable` | Internal infrastructure error (Redis down) |

### Token Endpoint Errors

Returned as `400` JSON response:

```json
{
  "error": "invalid_grant",
  "error_description": "Authorization code has expired or already been used"
}
```

| Error Code | Meaning |
|---|---|
| `invalid_request` | Missing `grant_type`, `code`, or `code_verifier` |
| `invalid_grant` | Code expired, already used, or PKCE verifier mismatch |
| `invalid_client` | Client authentication failed |
| `unsupported_grant_type` | Grant type other than `authorization_code` or `refresh_token` |

## 11. Complete Integration Example (TypeScript)

A minimal OAuth client class:

```typescript
interface TokenSet {
  accessToken: string;
  refreshToken: string;
  expiresAt: number; // epoch seconds
}

class KeylesOAuthClient {
  private clientId: string;
  private redirectUri: string;
  private issuer: string;

  constructor(config: { clientId: string; redirectUri: string; issuer: string }) {
    this.clientId = config.clientId;
    this.redirectUri = config.redirectUri;
    this.issuer = config.issuer;
  }

  // Step 1: Generate PKCE and redirect to authorization
  async login(): Promise<void> {
    const verifier = this.generateCodeVerifier();
    const challenge = await this.generateCodeChallenge(verifier);
    const state = crypto.randomUUID();

    sessionStorage.setItem('pkce_verifier', verifier);
    sessionStorage.setItem('oauth_state', state);

    const url = new URL(`${this.issuer}/oauth2/auth`);
    url.searchParams.set('client_id', this.clientId);
    url.searchParams.set('redirect_uri', this.redirectUri);
    url.searchParams.set('response_type', 'code');
    url.searchParams.set('scope', 'openid profile email');
    url.searchParams.set('state', state);
    url.searchParams.set('code_challenge', challenge);
    url.searchParams.set('code_challenge_method', 'S256');

    window.location.href = url.toString();
  }

  // Step 2: Handle callback and exchange code
  async handleCallback(): Promise<TokenSet> {
    const url = new URL(window.location.href);
    const code = url.searchParams.get('code');
    const state = url.searchParams.get('state');

    // Check for errors
    const error = url.searchParams.get('error');
    if (error) throw new Error(`Authorization error: ${error}`);

    // Validate state
    const storedState = sessionStorage.getItem('oauth_state');
    if (state !== storedState) throw new Error('State mismatch');

    // Exchange code
    const verifier = sessionStorage.getItem('pkce_verifier');
    const params = new URLSearchParams();
    params.set('grant_type', 'authorization_code');
    params.set('code', code!);
    params.set('redirect_uri', this.redirectUri);
    params.set('client_id', this.clientId);
    params.set('code_verifier', verifier!);

    const response = await fetch(`${this.issuer}/oauth2/token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: params,
    });

    if (!response.ok) {
      const err = await response.json();
      throw new Error(err.error_description || 'Token exchange failed');
    }

    const data = await response.json();

    // Clean up
    sessionStorage.removeItem('pkce_verifier');
    sessionStorage.removeItem('oauth_state');

    return {
      accessToken: data.access_token,
      refreshToken: data.refresh_token,
      expiresAt: Math.floor(Date.now() / 1000) + data.expires_in,
    };
  }

  // Step 3: Refresh tokens
  async refresh(tokenSet: TokenSet): Promise<TokenSet> {
    const params = new URLSearchParams();
    params.set('grant_type', 'refresh_token');
    params.set('refresh_token', tokenSet.refreshToken);
    params.set('client_id', this.clientId);

    const response = await fetch(`${this.issuer}/oauth2/token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: params,
    });

    if (!response.ok) throw new Error('Token refresh failed');
    const data = await response.json();

    return {
      accessToken: data.access_token,
      refreshToken: data.refresh_token,
      expiresAt: Math.floor(Date.now() / 1000) + data.expires_in,
    };
  }

  // Step 4: Logout (revoke tokens)
  async logout(tokenSet: TokenSet): Promise<void> {
    await fetch(`${this.issuer}/oauth2/revoke`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        token: tokenSet.refreshToken,
        token_type_hint: 'refresh_token',
        client_id: this.clientId,
      }),
    });
  }

  // Utility: generate cryptographically random code verifier
  private generateCodeVerifier(length = 64): string {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~';
    const array = new Uint8Array(length);
    crypto.getRandomValues(array);
    return Array.from(array, b => chars[b % chars.length]).join('');
  }

  // Utility: compute S256 challenge
  private async generateCodeChallenge(verifier: string): Promise<string> {
    const digest = await crypto.subtle.digest(
      'SHA-256',
      new TextEncoder().encode(verifier)
    );
    return btoa(String.fromCharCode(...new Uint8Array(digest)))
      .replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
  }
}

// Usage:
const keyles = new KeylesOAuthClient({
  clientId: 'your_client_id',
  redirectUri: 'https://myapp.example.com/auth/callback',
  issuer: 'http://localhost:8080',
});

// Start login flow
await keyles.login();

// In your callback handler:
const tokens = await keyles.handleCallback();

// Later, when access token expires:
const newTokens = await keyles.refresh(tokens);

// On user logout:
await keyles.logout(tokens);
```

## 12. Token Lifetimes

| Token | Default TTL | Configurable |
|---|---|---|
| Authorization Code | 5 minutes | `OAUTH_AUTH_CODE_TTL` |
| Access Token | 15 minutes | `OAUTH_ACCESS_TOKEN_TTL` |
| Refresh Token | 7 days | `OAUTH_REFRESH_TOKEN_TTL` |
| Authorization Transaction | 10 minutes | `OAUTH_AUTH_TRANSACTION_TTL` |
| End-User SSO Session | 8 hours | `SECURITY_SESSION_TTL` |

## 13. Security Best Practices

1. **Always use PKCE (S256)** — mandatory for all authorization code flows.
2. **Always validate `state`** — prevents CSRF attacks on your callback.
3. **Use `nonce`** — binds the ID token to your session, prevents replay.
4. **Store tokens securely**:
   - SPAs: Keep access tokens in memory only; never in `localStorage`.
   - Server-side: Store encrypted in an HttpOnly session cookie.
5. **Validate ID tokens** — verify `iss`, `aud`, `exp`, `nonce` before
   trusting claims. Never use the access token as proof of authentication.
6. **Use HTTPS** — all communication with the authorization server must
   be over TLS in production.
7. **Rotate refresh tokens** — each refresh returns a new refresh token;
   discard the old one.
8. **Register exact redirect URIs** — no wildcards; Keyles performs exact
   string matching on callback URIs.
9. **Handle token expiry** — refresh proactively before the access token
   expires; handle 401 responses gracefully.
10. **Revoke tokens on logout** — call `/oauth2/revoke` for the refresh
    token when the user signs out of your application.

## Quick Reference: curl Walkthrough

```bash
# 1. Generate PKCE (do this in your app, not curl)
CODE_VERIFIER=$(node -e "
  const c='ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~';
  const a=new Uint8Array(64); crypto.getRandomValues(a);
  console.log(Array.from(a,b=>c[b%c.length]).join(''))
")
CODE_CHALLENGE=$(echo -n "$CODE_VERIFIER" | openssl dgst -sha256 -binary | base64 | tr '/+' '_-' | tr -d '=')
STATE=$(uuidgen)

# 2. Open authorization URL in browser
echo "http://localhost:8080/oauth2/auth?client_id=dev_client_001\
&redirect_uri=http://localhost:3000/auth/callback\
&response_type=code\
&scope=openid%20profile%20email\
&state=$STATE\
&code_challenge=$CODE_CHALLENGE\
&code_challenge_method=S256"

# 3. After callback, exchange the code (replace AUTH_CODE with actual code)
curl -X POST http://localhost:8080/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d "grant_type=authorization_code" \
  -d "code=AUTH_CODE_HERE" \
  -d "redirect_uri=http://localhost:3000/auth/callback" \
  -d "client_id=dev_client_001" \
  -d "code_verifier=$CODE_VERIFIER" \
  -d "client_secret=dev_client_secret_change_in_production"

# 4. Get user info with access token
curl -H 'Authorization: Bearer ACCESS_TOKEN' \
  http://localhost:8080/oauth2/userinfo

# 5. Refresh tokens
curl -X POST http://localhost:8080/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d "grant_type=refresh_token" \
  -d "refresh_token=REFRESH_TOKEN" \
  -d "client_id=dev_client_001" \
  -d "client_secret=dev_client_secret_change_in_production"

# 6. Revoke tokens
curl -X POST http://localhost:8080/oauth2/revoke \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d "token=REFRESH_TOKEN" \
  -d "token_type_hint=refresh_token" \
  -d "client_id=dev_client_001" \
  -d "client_secret=dev_client_secret_change_in_production"
```
