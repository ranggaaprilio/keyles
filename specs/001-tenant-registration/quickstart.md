# Quickstart Guide: Multi-Tenant Registration with Email Verification

**Feature**: 001-tenant-registration  
**Last Updated**: 2025-12-06

## Prerequisites

- **Docker** and **Docker Compose** installed
- **Go** 1.22+ (for local development without Docker)
- **Node.js** 20.x LTS (for frontend development)
- **PostgreSQL** 15+ (or use Docker)
- **Redis** 7+ (or use Docker)
- **Brevo API Key** (sign up at [brevo.com](https://www.brevo.com))

## Quick Start (Docker Compose)

### 1. Clone and Setup

```bash
# Clone repository
git clone <repository-url>
cd keyles

# Checkout feature branch
git checkout 001-tenant-registration

# Copy environment files
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env
```

### 2. Configure Environment Variables

**backend/.env**:
```env
# Database
DATABASE_URL=postgres://keyles:keyles123@postgres:5432/keyles_sso?sslmode=disable

# Redis
REDIS_URL=redis://redis:6379/0

# Email Service (Brevo)
BREVO_API_KEY=your_brevo_api_key_here

# JWT
JWT_SECRET=your_super_secret_jwt_key_change_in_production

# Server
PORT=8080
GIN_MODE=debug

# CORS
FRONTEND_URL=http://localhost:5173

# Environment
ENVIRONMENT=development
```

**frontend/.env**:
```env
VITE_API_URL=http://localhost:8080/api/v1
VITE_APP_NAME=Keyles SSO
```

### 3. Get Brevo API Key

1. Sign up at [brevo.com](https://www.brevo.com) (free tier: 300 emails/day)
2. Go to **Settings** → **SMTP & API** → **API Keys**
3. Create a new API key
4. Copy the key and paste it in `backend/.env` as `BREVO_API_KEY`

### 4. Start Services

```bash
# Start all services (backend, frontend, PostgreSQL, Redis)
docker-compose up -d

# View logs
docker-compose logs -f

# Check service health
docker-compose ps
```

Services will be available at:
- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080/api/v1
- **API Docs**: http://localhost:8080/swagger (if implemented)
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

### 5. Run Database Migrations

```bash
# Migrations run automatically on backend startup
# Or manually:
docker-compose exec backend migrate -path=/app/migrations -database="$DATABASE_URL" up
```

### 6. Test the Application

**Option A: Using the UI (http://localhost:5173)**

1. Navigate to http://localhost:5173/register
2. Fill in the registration form:
   - Organization Name: "Test Corp"
   - Admin Name: "Admin User"
   - Admin Email: your-email@example.com
   - Password: "TestP@ss123"
3. Click "Register"
4. Check your email for the 6-digit OTP
5. Enter the OTP on the verification page
6. You should be redirected to the dashboard

**Option B: Using cURL**

```bash
# 1. Register a tenant
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "organizationName": "Test Corporation",
    "adminName": "Admin User",
    "adminEmail": "admin@testcorp.com",
    "password": "SecureP@ss123"
  }'

# Response will include tenantId and otpExpiresAt
# Check your email for the OTP code

# 2. Verify OTP
curl -X POST http://localhost:8080/api/v1/verify-otp \
  -H "Content-Type: application/json" \
  -d '{
    "tenantId": "<tenant-id-from-step-1>",
    "otpCode": "<code-from-email>"
  }'

# 3. Login
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@testcorp.com",
    "password": "SecureP@ss123"
  }'

# Response will include JWT token

# 4. Access dashboard (use token from login)
curl http://localhost:8080/api/v1/dashboard \
  -H "Authorization: Bearer <jwt-token-from-login>"
```

### 7. Stop Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (clears database)
docker-compose down -v
```

## Local Development (Without Docker)

### Backend

```bash
cd backend

# Install dependencies
go mod download

# Create database
createdb keyles_sso

# Run migrations
migrate -path=./migrations -database="postgres://localhost:5432/keyles_sso?sslmode=disable" up

# Start Redis (in separate terminal)
redis-server

# Set environment variables
export DATABASE_URL="postgres://localhost:5432/keyles_sso?sslmode=disable"
export REDIS_URL="redis://localhost:6379/0"
export BREVO_API_KEY="your_brevo_api_key"
export JWT_SECRET="your_jwt_secret"
export PORT=8080
export FRONTEND_URL="http://localhost:5173"

# Run tests
go test ./... -v

# Run backend
go run cmd/server/main.go

# Or with hot reload (install air: go install github.com/cosmtrek/air@latest)
air
```

### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev

# Run tests
npm test

# Run tests with coverage
npm test -- --coverage

# Build for production
npm run build

# Preview production build
npm run preview
```

## Testing

### Backend Tests

```bash
cd backend

# Run all tests
go test ./... -v

# Run tests with coverage
go test ./... -cover -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out

# Run specific package tests
go test ./domain/entities/... -v
go test ./usecase/tenant/... -v
go test ./interfaces/http/handlers/... -v

# Run integration tests only
go test ./tests/integration/... -v

# Run with race detection
go test ./... -race
```

### Frontend Tests

```bash
cd frontend

# Run unit tests
npm test

# Run with watch mode
npm test -- --watch

# Run with coverage
npm test -- --coverage

# Run specific test file
npm test -- RegistrationForm.test.tsx

# Run integration tests
npm run test:integration
```

### End-to-End Testing

```bash
# Install Playwright (if using E2E tests)
npm install -D @playwright/test

# Run E2E tests (requires services running)
npm run test:e2e

# Run in headless mode
npm run test:e2e:headless

# Open Playwright UI
npm run test:e2e:ui
```

## Manual Testing Scenarios

### Scenario 1: Happy Path (Successful Registration)

1. Navigate to registration page
2. Fill valid data:
   - Organization: "Acme Corp"
   - Name: "John Doe"
   - Email: valid email
   - Password: "Test@12345"
3. Submit form
4. Verify OTP email received
5. Enter OTP code
6. Verify redirect to dashboard
7. Verify tenant status is "active"

### Scenario 2: Duplicate Organization Name

1. Register first tenant with "Acme Corp"
2. Try to register second tenant with "Acme Corp"
3. Verify error: "Organization name already exists"

### Scenario 3: Duplicate Email

1. Register first tenant with "admin@test.com"
2. Try to register second tenant with "admin@test.com"
3. Verify error: "Email address already registered"

### Scenario 4: Weak Password

1. Try passwords:
   - "short" → Error: "minimum 8 characters"
   - "alllowercase" → Error: "must contain uppercase"
   - "ALLUPPERCASE" → Error: "must contain lowercase"
   - "NoNumbers!" → Error: "must contain number"
   - "NoSpecial1" → Error: "must contain special character"
2. Verify proper validation

### Scenario 5: Invalid OTP

1. Register tenant
2. Enter wrong OTP code (e.g., "000000")
3. Verify error and remaining attempts shown
4. Try 5 times
5. Verify max attempts error

### Scenario 6: Expired OTP

1. Register tenant
2. Wait 11 minutes (or manually expire in Redis)
3. Try to verify OTP
4. Verify error: "OTP expired"
5. Click "Resend OTP"
6. Verify new OTP sent

### Scenario 7: Rate Limiting (OTP Requests)

1. Register tenant
2. Request OTP resend 3 times within 1 hour
3. Try 4th request
4. Verify error: "Rate limit exceeded"

### Scenario 8: Login Before Verification

1. Register tenant (status: pending_verification)
2. Try to login
3. Verify error: "Please verify your email"

### Scenario 9: Successful Login After Verification

1. Register and verify tenant
2. Login with credentials
3. Verify JWT token returned
4. Access dashboard with token
5. Verify tenant and user data displayed

## Database Inspection

```bash
# Connect to PostgreSQL (Docker)
docker-compose exec postgres psql -U keyles -d keyles_sso

# Connect to PostgreSQL (Local)
psql keyles_sso

# Useful queries:
SELECT * FROM tenants;
SELECT * FROM users;
SELECT * FROM audit_logs ORDER BY created_at DESC LIMIT 10;

# Check tenant status
SELECT id, organization_name, status, verified_at FROM tenants;

# Check users by tenant
SELECT u.email, u.full_name, t.organization_name 
FROM users u 
JOIN tenants t ON u.tenant_id = t.id;

# View audit trail for a tenant
SELECT event_type, event_data, created_at 
FROM audit_logs 
WHERE tenant_id = '<tenant-uuid>'
ORDER BY created_at DESC;
```

## Redis Inspection

```bash
# Connect to Redis (Docker)
docker-compose exec redis redis-cli

# Connect to Redis (Local)
redis-cli

# Useful commands:
KEYS otp:*                              # List all OTP keys
GET otp:<tenant-id>                     # Get OTP for tenant
TTL otp:<tenant-id>                     # Check OTP expiration time
KEYS rate:*                             # List rate limit keys
GET rate:otp:request:<email>            # Check OTP request count
```

## Troubleshooting

### Backend won't start

**Problem**: `panic: dial tcp: connect: connection refused`

**Solution**: Check PostgreSQL and Redis are running
```bash
docker-compose ps
# Restart services if needed
docker-compose restart postgres redis
```

### Emails not sending

**Problem**: OTP emails not arriving

**Solutions**:
1. Check Brevo API key is valid in `.env`
2. Check Brevo dashboard for delivery status
3. Check backend logs for email sending errors:
   ```bash
   docker-compose logs -f backend | grep -i email
   ```
4. Verify email is not in spam folder
5. Check Brevo free tier limit (300 emails/day)

### Database migrations failed

**Problem**: `migration failed: relation already exists`

**Solution**: Reset database
```bash
# Docker
docker-compose down -v
docker-compose up -d

# Local
dropdb keyles_sso
createdb keyles_sso
migrate -path=./migrations -database="$DATABASE_URL" up
```

### CORS errors in browser

**Problem**: `Access-Control-Allow-Origin` errors

**Solution**: Check `FRONTEND_URL` in backend `.env` matches frontend URL
```bash
# backend/.env
FRONTEND_URL=http://localhost:5173
```

### Frontend can't connect to API

**Problem**: `Network Error` or `ERR_CONNECTION_REFUSED`

**Solution**: 
1. Verify backend is running: `curl http://localhost:8080/api/v1/health`
2. Check `VITE_API_URL` in frontend `.env`:
   ```env
   VITE_API_URL=http://localhost:8080/api/v1
   ```
3. Restart frontend: `npm run dev`

### Tests failing

**Problem**: Unit tests fail with database errors

**Solution**: Use test database or in-memory mock
```bash
# Set test database URL
export DATABASE_URL_TEST="postgres://localhost:5432/keyles_sso_test?sslmode=disable"

# Create test database
createdb keyles_sso_test

# Run migrations
migrate -path=./migrations -database="$DATABASE_URL_TEST" up

# Run tests
go test ./... -v
```

## Performance Testing

### Load Testing with `wrk`

```bash
# Install wrk (if not installed)
# macOS: brew install wrk
# Ubuntu: sudo apt-get install wrk

# Test registration endpoint
wrk -t4 -c100 -d30s -s test/load/register.lua http://localhost:8080/api/v1/register

# Test login endpoint
wrk -t4 -c100 -d30s -s test/load/login.lua http://localhost:8080/api/v1/login
```

### Memory and CPU Profiling

```bash
# Run backend with profiling
go run cmd/server/main.go -cpuprofile=cpu.prof -memprofile=mem.prof

# Analyze CPU profile
go tool pprof cpu.prof

# Analyze memory profile
go tool pprof mem.prof
```

## Security Testing

### Check Password Hashing

```bash
# Connect to database
psql keyles_sso

# Verify passwords are hashed (should NOT see plain text)
SELECT email, password_hash FROM users LIMIT 1;
# password_hash should start with $2a$ (bcrypt)
```

### Check JWT Token Expiration

```bash
# Login and get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"Test@12345"}' \
  | jq -r '.data.token')

# Decode JWT (use jwt.io or jwt-cli)
echo $TOKEN | jwt decode -

# Verify exp claim is set (should be 24 hours from now)
```

### Test Rate Limiting

```bash
# Send 4 OTP requests quickly (should block 4th)
for i in {1..4}; do
  curl -X POST http://localhost:8080/api/v1/resend-otp \
    -H "Content-Type: application/json" \
    -d '{"tenantId":"<tenant-id>"}'
  echo ""
done
```

## Clean Up

### Remove All Data

```bash
# Stop and remove containers, volumes
docker-compose down -v

# Remove built images
docker-compose down --rmi all

# Remove local database
dropdb keyles_sso
dropdb keyles_sso_test
```

### Reset Redis

```bash
docker-compose exec redis redis-cli FLUSHALL
```

## Next Steps

After successfully testing this feature:

1. **Review code for Clean Architecture compliance**
2. **Run full test suite and verify ≥85% coverage**
3. **Test all edge cases from spec.md**
4. **Update documentation with any learnings**
5. **Create pull request for review**
6. **Plan next feature (OIDC provider implementation)**

## Support

For issues or questions:
- Check logs: `docker-compose logs -f`
- Review [spec.md](./spec.md) for requirements
- Review [data-model.md](./data-model.md) for database schema
- Review [api.yaml](./contracts/api.yaml) for API documentation
- Check [research.md](./research.md) for technology decisions
