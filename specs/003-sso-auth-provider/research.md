# Research: Core SSO Auth Provider

**Date**: December 26, 2025  
**Feature**: [spec.md](./spec.md)

## 1. OIDC Library Selection (Go)

### Decision: Use `github.com/ory/fosite` (Recommended)

**Rationale**:

- **Production-ready**: Battle-tested OIDC/OAuth2 framework used by Ory Hydra
- **RFC compliance**: Full compliance with OAuth 2.0 (RFC 6749), OIDC Core, PKCE (RFC 7636), and token introspection
- **Flexibility**: Highly customizable storage backends, token strategies, and validation logic
- **Clean Architecture compatible**: Interface-driven design allows custom repository implementations
- **Active maintenance**: Regular security updates and community support

**Alternatives Considered**:

| Library                     | Pros                                                             | Cons                                                    | Verdict                                            |
| --------------------------- | ---------------------------------------------------------------- | ------------------------------------------------------- | -------------------------------------------------- |
| `github.com/ory/fosite`     | Full OIDC support, extensible, RFC-compliant, multi-tenant ready | Learning curve, more complex setup                      | ✅ **Selected**                                    |
| `golang.org/x/oauth2`       | Simple, official Google library                                  | Client-side only, lacks server implementation           | ❌ Rejected - doesn't provide server functionality |
| Custom implementation       | Full control, minimal dependencies                               | High maintenance, security risks, RFC compliance burden | ❌ Rejected - too much complexity                  |
| `github.com/coreos/go-oidc` | Simple OIDC provider                                             | Client-side validation only, no server                  | ❌ Rejected - client library only                  |

**Implementation Notes**:

- Fosite provides interfaces for: `Storage`, `AccessTokenStrategy`, `RefreshTokenStrategy`
- We'll implement custom storage backends for PostgreSQL (persistent) and Redis (ephemeral)
- Supports RS256 signing out of the box with custom key management

---

## 2. JWT Signing Strategy

### Decision: RS256 with `crypto/rsa` + JWKS Endpoint

**Rationale**:

- **Asymmetric signing**: Eliminates need to share secrets with resource servers
- **JWKS distribution**: Standard mechanism for key distribution via `/.well-known/jwks.json`
- **Key rotation**: Supports multiple active keys simultaneously for zero-downtime rotation
- **Fosite integration**: Fosite has built-in RS256 support via `fosite.DefaultJWTStrategy`

**Key Generation**:

```go
// Generate 2048-bit RSA key pair
privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
publicKey := &privateKey.PublicKey
```

**Key Storage**:

- **Private keys**: Stored in PostgreSQL `signing_keys` table, encrypted at rest
- **Public keys**: Exposed via JWKS endpoint, cached in Redis for performance
- **Key rotation**: Manual process initially, automated rotation as Phase 2 enhancement

**JWKS Format** (RFC 7517):

```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "key_2025_01",
      "alg": "RS256",
      "n": "<base64url-encoded-modulus>",
      "e": "AQAB"
    }
  ]
}
```

**Libraries**:

- `crypto/rsa`: Key generation and signing
- `github.com/golang-jwt/jwt/v5`: JWT token creation and parsing
- `gopkg.in/square/go-jose.v2`: JWKS generation and management

---

## 3. Multi-Tenant Identification Strategy

### Decision: Client-based tenant lookup (from clarifications)

**Rationale** (from spec clarifications):

- Each `client_id` is associated with exactly one `tenant_id`
- When authorization request arrives with `client_id`, lookup tenant context
- No subdomain routing needed (simplifies deployment)
- Most secure approach - no user input for tenant selection

**Implementation**:

```go
// In authorization handler
client, err := storage.GetClient(ctx, clientID)
tenantID := client.TenantID
// Set tenant context for rest of request
ctx = context.WithValue(ctx, "tenant_id", tenantID)
```

**Database Schema Implication**:

```sql
CREATE TABLE clients (
    client_id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(id),
    -- ... other fields
);
CREATE INDEX idx_clients_tenant ON clients(tenant_id);
```

---

## 4. PostgreSQL Schema Design

### Migrations Tool: `golang-migrate/migrate`

**Decision**: Use `github.com/golang-migrate/migrate/v4`

**Rationale**:

- Most popular Go migration tool
- Supports both SQL and programmatic migrations
- Version control for database schema
- Compatible with PostgreSQL, supports transactions
- CLI tool for local development

**Migration File Structure**:

```
migrations/
├── 000001_create_tenants.up.sql
├── 000001_create_tenants.down.sql
├── 000002_create_clients.up.sql
├── 000002_create_clients.down.sql
└── ...
```

### Database Design Principles

**Multi-Tenancy Strategy**: Shared schema with tenant_id column (Row-Level Security)

**Rationale**:

- Simplifies deployment (single database)
- Enables cross-tenant analytics if needed
- PostgreSQL RLS provides security isolation
- Easier to maintain than separate schemas per tenant

**Entity Relationships**:

```
tenants (1) ──< (N) clients
tenants (1) ──< (N) users
clients (1) ──< (N) user_role_assignments ─> (N) users
users (1) ──< (N) refresh_tokens ─> (N) clients
```

### Required Tables (from spec entities)

1. **tenants**: Platform tenants/organizations
2. **clients**: Registered OAuth2/OIDC client applications
3. **users**: User accounts per tenant
4. **user_role_assignments**: RBAC - which users can access which clients
5. **refresh_tokens**: Long-lived tokens for token refresh flow
6. **signing_keys**: RSA keypairs for JWT signing
7. **audit_logs**: Security audit trail

### Connection Pool Configuration

**Library**: `github.com/jackc/pgx/v5` (PostgreSQL driver)

**Rationale**:

- Best performance for PostgreSQL in Go
- Native prepared statement support
- Better error handling than `database/sql`
- COPY protocol support for bulk operations

**Pool Settings** (recommended):

```go
config, _ := pgxpool.ParseConfig("postgres://...")
config.MaxConns = 25                    // Max connections
config.MinConns = 5                     // Min idle connections
config.MaxConnLifetime = time.Hour      // Connection lifetime
config.MaxConnIdleTime = 30 * time.Minute
```

---

## 5. Redis Strategy for Ephemeral Data

### Library: `github.com/redis/go-redis/v9`

**Decision**: Use go-redis v9 with context support

**Rationale**:

- Most popular Redis client for Go
- Full Redis 7.x feature support
- Context-based API for cancellation
- Connection pooling built-in
- Pub/Sub support for future features

### Data Structures & TTLs

**Authorization Codes** (5-minute lifetime):

```
Key Pattern: auth:code:{code_value}
Value: JSON {client_id, user_id, tenant_id, redirect_uri, code_challenge, scope}
TTL: 300 seconds (5 minutes)
```

**User Sessions** (8-hour lifetime):

```
Key Pattern: session:{session_id}
Value: JSON {user_id, tenant_id, created_at, expires_at, ip_address, user_agent}
TTL: 28800 seconds (8 hours)
```

**Refresh Tokens** (7-day lifetime):

```
Key Pattern: refresh:{token_value}
Value: JSON {client_id, user_id, tenant_id, created_at, expires_at, is_revoked}
TTL: 604800 seconds (7 days)
```

**JWKS Cache** (1-hour cache):

```
Key Pattern: jwks:public_keys
Value: JSON (JWKS document)
TTL: 3600 seconds (1 hour)
```

**Rate Limiting** (per-client throttling):

```
Key Pattern: ratelimit:token:{client_id}
Value: Request count
TTL: 60 seconds (1-minute window)
Command: INCR with GET to check limit
```

### Redis Configuration

**Connection**:

```go
rdb := redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    Password:     "", // from env
    DB:           0,
    PoolSize:     10,
    MinIdleConns: 5,
})
```

**High Availability** (future):

- Redis Sentinel for automatic failover
- Redis Cluster for horizontal scaling
- Consider Redis persistence (AOF) for refresh tokens

---

## 6. PKCE Implementation

### Standard: RFC 7636 - Proof Key for Code Exchange

**Decision**: Mandatory S256 (SHA-256) method

**Rationale**:

- Prevents authorization code interception attacks
- Required for public clients (SPAs, mobile apps)
- More secure than "plain" method
- Fosite has built-in PKCE support

**Flow**:

1. Client generates `code_verifier` (43-128 random characters)
2. Client computes `code_challenge = BASE64URL(SHA256(code_verifier))`
3. Authorization request includes `code_challenge` and `code_challenge_method=S256`
4. Token exchange request includes original `code_verifier`
5. Server validates: `BASE64URL(SHA256(code_verifier)) == stored_code_challenge`

**Implementation** (Fosite handles this automatically):

```go
// Fosite validates PKCE in OAuth2Provider.NewAccessRequest()
// Just need to store code_challenge with authorization code
```

**Validation**:

```go
func validatePKCE(verifier, challenge string) bool {
    hash := sha256.Sum256([]byte(verifier))
    computed := base64.RawURLEncoding.EncodeToString(hash[:])
    return subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) == 1
}
```

---

## 7. Rate Limiting Strategy

### Decision: Token bucket algorithm with Redis

**Limit**: 10 requests/minute per `client_id` (from clarifications)

**Implementation Approach**:

**Library**: `github.com/ulule/limiter/v3` with Redis driver

**Rationale**:

- Proven rate limiting library
- Multiple storage backends (Redis, memory, etc.)
- Token bucket and sliding window algorithms
- Middleware-ready for HTTP handlers

**Configuration**:

```go
import "github.com/ulule/limiter/v3"
import "github.com/ulule/limiter/v3/drivers/store/redis"

// Create rate limiter
rate := limiter.Rate{
    Period: 1 * time.Minute,
    Limit:  10,
}
store := redis.NewStoreWithOptions(redisClient, limiter.StoreOptions{
    Prefix: "ratelimit:token:",
})
instance := limiter.New(store, rate)
```

**Middleware**:

```go
func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        clientID := extractClientID(r)
        context, err := limiter.Get(r.Context(), clientID)
        if err != nil || context.Reached {
            w.Header().Set("X-RateLimit-Limit", "10")
            w.Header().Set("X-RateLimit-Remaining", "0")
            w.Header().Set("Retry-After", "60")
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

**Alternative**: Manual implementation with Redis INCR

```go
key := fmt.Sprintf("ratelimit:token:%s", clientID)
count, _ := redisClient.Incr(ctx, key).Result()
if count == 1 {
    redisClient.Expire(ctx, key, time.Minute)
}
if count > 10 {
    return ErrRateLimitExceeded
}
```

---

## 8. Secure Session Management

### Strategy: HTTP-only, Secure, SameSite cookies + Redis

**Cookie Configuration**:

```go
http.SetCookie(w, &http.Cookie{
    Name:     "session_id",
    Value:    sessionID,
    Path:     "/",
    Domain:   ".keyles.com",      // for SSO across subdomains
    MaxAge:   28800,               // 8 hours
    Secure:   true,                // HTTPS only
    HttpOnly: true,                // No JavaScript access
    SameSite: http.SameSiteLaxMode, // CSRF protection
})
```

**Session Storage**:

- Store in Redis with session_id as key
- Include user_id, tenant_id, IP address, user agent
- Invalidate on logout or timeout
- Sliding expiration: update TTL on each request

**Session Validation Middleware**:

```go
func ValidateSession(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("session_id")
        if err != nil {
            http.Error(w, "Unauthorized", 401)
            return
        }

        // Get session from Redis
        sessionKey := fmt.Sprintf("session:%s", cookie.Value)
        sessionData, err := redisClient.Get(r.Context(), sessionKey).Result()
        if err != nil {
            http.Error(w, "Session expired", 401)
            return
        }

        // Extend session TTL (sliding expiration)
        redisClient.Expire(r.Context(), sessionKey, 8*time.Hour)

        // Add session to context
        ctx := context.WithValue(r.Context(), "session", sessionData)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

## 9. CSRF Protection

### Strategy: State parameter + SameSite cookies

**OAuth State Parameter** (built into OAuth 2.0):

- Client generates random `state` value
- Includes in authorization request
- Validates returned `state` matches original
- Fosite handles this automatically

**Additional Protection**:

- `SameSite=Lax` on session cookies
- CORS configuration to restrict origins
- Referer header validation for sensitive operations

**Implementation**:

```go
// Fosite handles state validation automatically
// Just ensure state is passed through correctly

// Additional CSRF token for admin operations
func GenerateCSRFToken() string {
    b := make([]byte, 32)
    rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)
}
```

---

## 10. React Frontend Architecture

### Stack Decisions

**State Management**: React Context API + hooks

**Rationale**:

- Sufficient for auth state management
- No need for Redux complexity
- Aligns with functional component requirement
- Built-in React feature

**Routing**: React Router v6

**Form Handling**: React Hook Form

**Rationale**:

- Performance (uncontrolled components)
- Built-in validation
- TypeScript support
- Small bundle size

**HTTP Client**: Axios

**Rationale**:

- Interceptor support for auth headers
- Request/response transformation
- Better error handling than fetch
- TypeScript definitions

**UI Components**: Tailwind CSS + Headless UI

**Rationale** (observed from existing frontend):

- Already in use (tailwind.config.js exists)
- Utility-first approach
- Headless UI for accessible components
- TypeScript support

### Component Structure

```
frontend/src/
├── components/
│   ├── auth/
│   │   ├── LoginForm.tsx           # Login page (P1)
│   │   ├── AuthCallback.tsx        # OAuth callback handler
│   │   └── ProtectedRoute.tsx      # Route guard component
│   ├── admin/
│   │   ├── ClientList.tsx          # Client management (P1)
│   │   ├── ClientForm.tsx          # Create/edit client
│   │   ├── UserRoleList.tsx        # Role management (P2)
│   │   └── UserRoleForm.tsx        # Assign roles
│   └── common/
│       ├── Button.tsx
│       ├── Input.tsx
│       └── Modal.tsx
├── services/
│   ├── authService.ts              # OAuth/OIDC client logic
│   ├── apiClient.ts                # Axios instance with interceptors
│   └── types.ts                    # TypeScript interfaces
├── hooks/
│   ├── useAuth.tsx                 # Authentication state hook
│   ├── useClient.tsx               # Client management hook
│   └── useRoles.tsx                # Role management hook
├── contexts/
│   └── AuthContext.tsx             # Auth state provider
└── utils/
    ├── pkce.ts                     # PKCE code generation
    └── storage.ts                  # LocalStorage wrapper
```

### PKCE Implementation (Client-side)

```typescript
// utils/pkce.ts
export async function generateCodeVerifier(): Promise<string> {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return base64URLEncode(array);
}

export async function generateCodeChallenge(verifier: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const hash = await crypto.subtle.digest("SHA-256", data);
  return base64URLEncode(new Uint8Array(hash));
}
```

### OAuth Flow Implementation

```typescript
// services/authService.ts
export class AuthService {
  async initiateLogin() {
    const codeVerifier = await generateCodeVerifier();
    const codeChallenge = await generateCodeChallenge(codeVerifier);
    const state = generateRandomString(32);

    // Store for callback
    sessionStorage.setItem("code_verifier", codeVerifier);
    sessionStorage.setItem("oauth_state", state);

    // Redirect to authorization endpoint
    const params = new URLSearchParams({
      client_id: CLIENT_ID,
      redirect_uri: REDIRECT_URI,
      response_type: "code",
      scope: "openid profile email",
      state,
      code_challenge: codeChallenge,
      code_challenge_method: "S256",
    });

    window.location.href = `${AUTH_URL}/oauth2/auth?${params}`;
  }

  async handleCallback(code: string, state: string) {
    // Validate state
    const storedState = sessionStorage.getItem("oauth_state");
    if (state !== storedState) throw new Error("Invalid state");

    // Exchange code for tokens
    const codeVerifier = sessionStorage.getItem("code_verifier");
    const tokens = await this.exchangeCode(code, codeVerifier);

    // Store tokens
    localStorage.setItem("access_token", tokens.access_token);
    localStorage.setItem("refresh_token", tokens.refresh_token);
    localStorage.setItem("id_token", tokens.id_token);

    // Clear temporary storage
    sessionStorage.removeItem("code_verifier");
    sessionStorage.removeItem("oauth_state");

    return tokens;
  }
}
```

---

## 11. HTTP Router and Middleware

### Decision: `github.com/go-chi/chi/v5`

**Rationale**:

- Lightweight, idiomatic Go
- Standard `net/http` compatible
- Built-in middleware support
- Route grouping and sub-routers
- Context-based request values
- Better than Gin for Clean Architecture (less opinionated)

**Middleware Stack**:

```go
r := chi.NewRouter()

// Standard middleware
r.Use(middleware.RequestID)
r.Use(middleware.RealIP)
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)

// Custom middleware
r.Use(CORSMiddleware)
r.Use(TenantContextMiddleware)

// Rate-limited routes
r.Group(func(r chi.Router) {
    r.Use(RateLimitMiddleware)

    r.Post("/oauth2/token", tokenHandler)
    r.Post("/oauth2/introspect", introspectHandler)
})

// Session-protected routes
r.Group(func(r chi.Router) {
    r.Use(ValidateSession)

    r.Get("/admin/clients", listClientsHandler)
    r.Post("/admin/clients", createClientHandler)
})
```

---

## 12. Configuration Management

### Decision: Environment variables + `github.com/spf13/viper`

**Rationale**:

- 12-factor app compliance
- Supports multiple sources (env, file, remote)
- Type-safe configuration
- Easy testing with mock configs

**Configuration Structure**:

```go
type Config struct {
    Server struct {
        Port string `mapstructure:"port"`
        Host string `mapstructure:"host"`
    }
    Database struct {
        URL         string `mapstructure:"url"`
        MaxConns    int    `mapstructure:"max_conns"`
        MinConns    int    `mapstructure:"min_conns"`
    }
    Redis struct {
        URL string `mapstructure:"url"`
        DB  int    `mapstructure:"db"`
    }
    OAuth struct {
        Issuer          string `mapstructure:"issuer"`
        AccessTokenTTL  int    `mapstructure:"access_token_ttl"`  // 900 (15 min)
        RefreshTokenTTL int    `mapstructure:"refresh_token_ttl"` // 604800 (7 days)
    }
    Security struct {
        AllowedOrigins []string `mapstructure:"allowed_origins"`
        CookieDomain   string   `mapstructure:"cookie_domain"`
    }
}
```

**Environment Variables**:

```bash
# .env.example
SERVER_PORT=8080
SERVER_HOST=localhost

DATABASE_URL=postgres://user:pass@localhost:5432/keyles_sso
DATABASE_MAX_CONNS=25
DATABASE_MIN_CONNS=5

REDIS_URL=redis://localhost:6379
REDIS_DB=0

OAUTH_ISSUER=https://sso.keyles.com
OAUTH_ACCESS_TOKEN_TTL=900
OAUTH_REFRESH_TOKEN_TTL=604800

SECURITY_ALLOWED_ORIGINS=http://localhost:3000,https://app.keyles.com
SECURITY_COOKIE_DOMAIN=.keyles.com
```

---

## Summary of Technology Decisions

| Component              | Technology          | Rationale                                              |
| ---------------------- | ------------------- | ------------------------------------------------------ |
| **OIDC Framework**     | ory/fosite          | Full RFC compliance, extensible, production-ready      |
| **JWT Signing**        | crypto/rsa + RS256  | Asymmetric signing, JWKS support, industry standard    |
| **PostgreSQL Driver**  | pgx/v5              | Best performance, native features, connection pooling  |
| **Redis Client**       | go-redis/v9         | Most popular, context support, full feature set        |
| **Migrations**         | golang-migrate      | Industry standard, version control, rollback support   |
| **HTTP Router**        | chi/v5              | Lightweight, idiomatic, middleware-friendly            |
| **Rate Limiting**      | ulule/limiter/v3    | Token bucket algorithm, Redis-backed, middleware-ready |
| **Configuration**      | viper               | Environment variables, type-safe, multi-source         |
| **Frontend Framework** | React + TypeScript  | Existing stack, functional components, type safety     |
| **State Management**   | Context API + Hooks | Sufficient for auth, no Redux complexity needed        |
| **HTTP Client**        | Axios               | Interceptors, error handling, TypeScript support       |
| **UI Framework**       | Tailwind CSS        | Existing project standard, utility-first               |
| **Form Handling**      | React Hook Form     | Performance, validation, TypeScript support            |

---

## Next Steps

Phase 1 will produce:

1. **data-model.md**: Complete PostgreSQL schema definitions
2. **contracts/**: OpenAPI specifications for all endpoints
3. **quickstart.md**: Developer setup and local development guide

All decisions above will be referenced during implementation planning.
