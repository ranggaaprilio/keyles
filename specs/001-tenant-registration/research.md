# Research: Multi-Tenant Registration with Email Verification

**Feature**: 001-tenant-registration  
**Date**: 2025-12-06  
**Status**: Complete

## Technology Decisions

### Backend Framework: Gin (Golang)

**Decision**: Use Gin web framework for HTTP routing and middleware

**Rationale**:
- High performance (40x faster than Martini)
- Excellent middleware support (CORS, auth, rate limiting)
- Clean API for RESTful services
- Strong community and ecosystem
- Native Go features (goroutines, channels) for concurrent request handling

**Alternatives Considered**:
- **Echo**: Similar performance, but smaller ecosystem
- **Fiber**: Express-like API, but uses fasthttp which has compatibility issues
- **Chi**: Lightweight but requires more manual setup

### ORM: GORM

**Decision**: Use GORM for database operations with PostgreSQL

**Rationale**:
- Clean Architecture friendly (interface-based repositories)
- Automatic migrations support (complemented by golang-migrate for versioning)
- Multi-tenant pattern support via scopes
- Transaction management
- Association handling for foreign keys

**Alternatives Considered**:
- **sqlx**: More control but requires more boilerplate
- **ent**: Type-safe but steeper learning curve
- Raw SQL: Maximum control but harder to maintain

### Migration Tool: golang-migrate

**Decision**: Use golang-migrate for database schema versioning

**Rationale**:
- Version-controlled migrations (up/down scripts)
- CLI and programmatic API
- Supports PostgreSQL, MySQL, SQLite
- Integrates well with CI/CD pipelines
- Separate from ORM for production safety

**Best Practices**:
- One migration per logical change
- Always provide down migrations
- Test migrations on staging before production
- Use timestamps in migration filenames

### Email Service: Brevo (formerly Sendinblue)

**Decision**: Use Brevo API for transactional email delivery

**Rationale**:
- Reliable transactional email service
- Official Go SDK available
- Free tier: 300 emails/day (sufficient for development and initial launch)
- Template management via API or web interface
- Delivery tracking and analytics
- SMTP and API options

**Configuration**:
```go
import brevo "github.com/getbrevo/brevo-go/lib"

config := brevo.NewConfiguration()
config.AddDefaultHeader("api-key", os.Getenv("BREVO_API_KEY"))
client := brevo.NewAPIClient(config)
```

**Best Practices**:
- Use API (not SMTP) for better error handling
- Implement retry logic for temporary failures
- Log all email send attempts for audit trail
- Use templates for consistent branding
- Monitor delivery rates and bounce metrics

### Email Templates: React Email

**Decision**: Use React Email for type-safe, component-based email templates

**Rationale**:
- Write emails as React components
- TypeScript support for type safety
- Preview emails during development
- Export to HTML for use with any email service
- Tailwind CSS support for styling
- Version control friendly (code-based templates)

**Integration**:
```tsx
import { render } from '@react-email/render';
import OTPEmail from './emails/OTPEmail';

const html = render(<OTPEmail otp="123456" tenantName="Acme Corp" />);
// Send via Brevo API
```

### Caching: Redis

**Decision**: Use Redis for OTP storage and rate limiting

**Rationale**:
- Built-in TTL for automatic OTP expiration (10 minutes)
- Atomic operations for rate limiting (INCR, EXPIRE)
- High performance for frequent read/write operations
- Persistence options for reliability
- Pub/Sub for future real-time features

**OTP Storage Pattern**:
```
Key: otp:{tenant_id}
Value: {otp_code}:{attempts}
TTL: 600 seconds (10 minutes)
```

**Rate Limiting Pattern**:
```
Key: rate:otp:request:{email}
Value: {count}
TTL: 3600 seconds (1 hour)

Key: rate:otp:verify:{otp_id}
Value: {attempts}
TTL: Same as OTP TTL
```

### Frontend Framework: React + TypeScript

**Decision**: React 18 with TypeScript strict mode and Vite

**Rationale**:
- Industry standard for SPAs
- TypeScript provides compile-time safety
- Vite offers fast development experience
- Large ecosystem of libraries
- Easy integration with shadcn/ui

### UI Library: shadcn/ui

**Decision**: Use shadcn/ui for component library

**Rationale**:
- Copy-paste components (no package dependency bloat)
- Built on Radix UI (accessible by default)
- Customizable via Tailwind CSS
- TypeScript native
- Modern, clean design system

**Components Needed**:
- Form (registration, OTP input)
- Input (text fields, password)
- Button (submit, resend OTP)
- Card (form containers)
- Toast (success/error notifications)
- Label (form labels)

### State Management: TanStack Query (React Query)

**Decision**: Use TanStack Query for server state management

**Rationale**:
- Automatic caching and invalidation
- Loading and error states built-in
- Optimistic updates support
- Retry logic out of the box
- DevTools for debugging
- No global state boilerplate

**Usage Pattern**:
```tsx
const { mutate, isLoading, error } = useMutation({
  mutationFn: (data) => api.registerTenant(data),
  onSuccess: () => navigate('/verify-otp'),
});
```

### Validation: Zod

**Decision**: Use Zod for runtime validation (frontend and backend)

**Rationale**:
- TypeScript-first schema validation
- Runtime type safety
- Shareable schemas between frontend/backend
- Integration with React Hook Form
- Clear error messages

**Shared Schema Example**:
```typescript
const RegistrationSchema = z.object({
  organizationName: z.string().min(2).max(100),
  adminName: z.string().min(2).max(100),
  adminEmail: z.string().email(),
  password: z.string().min(8).regex(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])/),
});
```

## Security Best Practices

### Password Hashing: Bcrypt

**Decision**: Use bcrypt with cost factor 12

**Rationale**:
- Industry standard for password hashing
- Adaptive (cost can be increased over time)
- Salt automatically generated and stored
- Resistant to brute-force attacks

**Implementation**:
```go
import "golang.org/x/crypto/bcrypt"

const bcryptCost = 12

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
    return string(bytes), err
}
```

### OTP Generation: Crypto/rand

**Decision**: Use crypto/rand for cryptographically secure OTP generation

**Rationale**:
- Cryptographically secure random number generation
- Uniform distribution
- Unpredictable (cannot be guessed)

**Implementation**:
```go
import "crypto/rand"
import "math/big"

func GenerateOTP() (string, error) {
    n, err := rand.Int(rand.Reader, big.NewInt(1000000))
    if err != nil {
        return "", err
    }
    return fmt.Sprintf("%06d", n.Int64()), nil
}
```

### Rate Limiting Strategy

**Decision**: Redis-based rate limiting with sliding window

**Rationale**:
- Prevents brute-force OTP guessing
- Prevents email spam
- Distributed (works across multiple server instances)
- Configurable limits per endpoint

**Limits**:
- OTP requests: 3 per email per hour
- OTP verification: 5 attempts per OTP
- Login attempts: 5 per email per 15 minutes (future)

### Multi-Tenant Isolation

**Decision**: Database-level isolation via tenant_id column

**Rationale**:
- Single database, multiple tenants (cost-effective)
- Row-level security via WHERE clauses
- GORM scopes enforce tenant filtering
- Horizontal scaling via sharding (future)

**Pattern**:
```go
// All queries must include tenant_id
db.Where("tenant_id = ?", tenantID).Find(&users)

// GORM scope for automatic filtering
func TenantScope(tenantID string) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("tenant_id = ?", tenantID)
    }
}
```

## Testing Strategy

### Backend Testing

**Unit Tests (≥85% coverage target)**:
- Domain entities: Business rule validation
- Use cases: Business logic flows
- Mocking: gomock for repository interfaces

**Integration Tests**:
- HTTP handlers: End-to-end API tests
- Database: testcontainers for PostgreSQL
- Redis: miniredis for in-memory testing
- Email: Mock Brevo client

**Tools**:
- Go testing package
- testify for assertions
- gomock for mocking
- testcontainers for integration tests

### Frontend Testing

**Unit Tests**:
- Components: React Testing Library
- Hooks: React Hooks Testing Library
- Utilities: Vitest

**Integration Tests (optional)**:
- User flows: Playwright
- API mocking: MSW (Mock Service Worker)

**Tools**:
- Vitest (test runner)
- React Testing Library
- user-event (user interaction simulation)

## Development Environment

### Docker Compose Setup

**Services**:
- PostgreSQL 15
- Redis 7
- Backend API (development mode with hot reload)
- Frontend dev server (Vite)

**Benefits**:
- Consistent development environment
- Easy onboarding for new developers
- Matches production infrastructure

### Environment Variables

**Backend (.env)**:
```
DATABASE_URL=postgres://user:pass@localhost:5432/keyles_sso
REDIS_URL=redis://localhost:6379/0
BREVO_API_KEY=your_api_key_here
JWT_SECRET=your_jwt_secret_here
FRONTEND_URL=http://localhost:5173
```

**Frontend (.env)**:
```
VITE_API_URL=http://localhost:8080/api/v1
VITE_APP_NAME=Keyles SSO
```

## Performance Considerations

### Database Indexing

**Critical Indexes**:
```sql
-- Tenant uniqueness and lookup
CREATE UNIQUE INDEX idx_tenants_org_name ON tenants(LOWER(organization_name));
CREATE INDEX idx_tenants_status ON tenants(status);

-- User uniqueness and lookup
CREATE UNIQUE INDEX idx_users_email ON users(LOWER(email));
CREATE INDEX idx_users_tenant_id ON users(tenant_id);

-- Audit log queries
CREATE INDEX idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
```

### Caching Strategy

**Redis Usage**:
- OTP codes: 10-minute TTL
- Rate limiting counters: 1-hour TTL
- Session tokens: JWT (stateless, no caching needed initially)

**Future Considerations**:
- Tenant metadata caching
- API response caching for public endpoints

### Connection Pooling

**PostgreSQL**:
```go
db.SetMaxOpenConns(25)    // Max concurrent connections
db.SetMaxIdleConns(5)     // Keep idle connections ready
db.SetConnMaxLifetime(5 * time.Minute)
```

**Redis**:
```go
&redis.Options{
    PoolSize: 10,
    MinIdleConns: 2,
}
```

## Deployment Considerations

### Containerization

**Backend Dockerfile**:
- Multi-stage build (builder + runtime)
- Minimal runtime image (distroless or alpine)
- Health check endpoint

**Frontend Dockerfile**:
- Build static assets
- Serve via nginx or Caddy
- Gzip compression enabled

### Environment Separation

**Development**: Local Docker Compose
**Staging**: Cloud deployment (AWS/GCP/Azure) with managed PostgreSQL and Redis
**Production**: Same as staging with auto-scaling and monitoring

### Migration Strategy

**Automated Migrations**:
- Run migrations on application startup (development)
- Manual migration step in CI/CD pipeline (production)
- Backup database before migrations

## Monitoring and Observability

### Logging

**Structured Logging** (logrus or zap):
```go
log.WithFields(log.Fields{
    "tenant_id": tenantID,
    "action": "register_tenant",
    "email": email,
}).Info("Tenant registration initiated")
```

**Log Levels**:
- ERROR: Failed operations, exceptions
- WARN: Rate limit exceeded, validation failures
- INFO: Successful operations, state changes
- DEBUG: Detailed request/response data (development only)

### Metrics (Future)

**Key Metrics**:
- Registration success rate
- OTP verification success rate
- Email delivery rate
- API response times (p50, p95, p99)
- Active tenants count

**Tools**: Prometheus + Grafana

### Health Checks

**Endpoints**:
- `/health`: Basic health check
- `/health/db`: Database connectivity
- `/health/redis`: Redis connectivity

## Future Enhancements

**Phase 2+**:
- Social login (Google, Microsoft, GitHub)
- Multi-factor authentication (TOTP)
- Tenant user invitations
- OIDC provider implementation
- SAML support
- Tenant usage analytics
- Billing integration

**Infrastructure**:
- Kubernetes deployment
- Auto-scaling based on load
- Multi-region deployment
- CDN for static assets
- Distributed tracing (OpenTelemetry)
