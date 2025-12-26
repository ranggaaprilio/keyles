# Quickstart Guide: Core SSO Auth Provider

This guide will help you set up the Core SSO Auth Provider for local development.

## Prerequisites

- **Go**: 1.21+ ([install](https://go.dev/doc/install))
- **Node.js**: 18+ ([install](https://nodejs.org/))
- **PostgreSQL**: 14+ ([install](https://www.postgresql.org/download/))
- **Redis**: 7+ ([install](https://redis.io/download/))
- **Docker** (optional): For containerized development
- **golang-migrate**: CLI for database migrations ([install](#migrations-setup))

## Project Structure

```
keyles/
├── backend/
│   ├── domain/              # Business logic & interfaces
│   │   ├── entities/        # Domain entities
│   │   ├── repositories/    # Repository interfaces
│   │   └── services/        # Service interfaces
│   ├── usecase/             # Application business rules
│   │   ├── auth/            # Authentication use cases
│   │   ├── client/          # Client management
│   │   └── role/            # Role management
│   ├── infrastructure/      # External implementations
│   │   ├── config/          # Configuration management
│   │   ├── persistence/     # Database implementations
│   │   │   ├── postgres/    # PostgreSQL repositories
│   │   │   └── redis/       # Redis repositories
│   │   └── services/        # External service clients
│   ├── interfaces/          # HTTP handlers (outer layer)
│   │   └── http/
│   │       ├── handlers/    # Request handlers
│   │       ├── middleware/  # HTTP middleware
│   │       └── router.go    # Route definitions
│   ├── migrations/          # Database migrations
│   ├── cmd/
│   │   └── server/
│   │       └── main.go      # Application entry point
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── components/      # React components
│   │   │   ├── auth/        # Auth-related components
│   │   │   ├── admin/       # Admin portal components
│   │   │   └── common/      # Shared components
│   │   ├── services/        # API clients
│   │   ├── hooks/           # Custom React hooks
│   │   ├── contexts/        # React contexts
│   │   ├── types/           # TypeScript definitions
│   │   └── utils/           # Utility functions
│   ├── package.json
│   └── vite.config.ts
└── specs/
    └── 003-sso-auth-provider/
        ├── spec.md          # Feature specification
        ├── plan.md          # Implementation plan
        ├── research.md      # Technology decisions
        ├── data-model.md    # Database schema
        ├── quickstart.md    # This file
        └── contracts/       # API specifications
```

## Setup Instructions

### 1. Database Setup

#### Option A: Local PostgreSQL

```bash
# Create database
createdb keyles_sso

# Create user (optional)
psql postgres -c "CREATE USER keyles_user WITH PASSWORD 'keyles_pass';"
psql postgres -c "GRANT ALL PRIVILEGES ON DATABASE keyles_sso TO keyles_user;"
```

#### Option B: Docker PostgreSQL

```bash
docker run --name keyles-postgres \
  -e POSTGRES_DB=keyles_sso \
  -e POSTGRES_USER=keyles_user \
  -e POSTGRES_PASSWORD=keyles_pass \
  -p 5432:5432 \
  -d postgres:16
```

### 2. Redis Setup

#### Option A: Local Redis

```bash
# Start Redis server
redis-server

# Verify connection
redis-cli ping  # Should return "PONG"
```

#### Option B: Docker Redis

```bash
docker run --name keyles-redis \
  -p 6379:6379 \
  -d redis:7
```

### 3. Migrations Setup

Install golang-migrate CLI:

```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Windows (using scoop)
scoop install migrate
```

Run migrations:

```bash
cd backend

# Run all up migrations
migrate -path migrations -database "postgres://keyles_user:keyles_pass@localhost:5432/keyles_sso?sslmode=disable" up

# Verify migrations
migrate -path migrations -database "postgres://keyles_user:keyles_pass@localhost:5432/keyles_sso?sslmode=disable" version
```

### 4. Backend Setup

```bash
cd backend

# Install dependencies
go mod download

# Generate RSA keypair for JWT signing (development only)
go run cmd/keygen/main.go

# Copy environment template
cp .env.example .env

# Edit .env with your settings (see Configuration section below)
nano .env

# Build the application
go build -o bin/server cmd/server/main.go

# Run the server
./bin/server

# Or use air for hot reloading (optional)
go install github.com/cosmtrek/air@latest
air
```

### 5. Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Copy environment template
cp .env.example .env.local

# Edit .env.local
nano .env.local

# Start development server
npm run dev
```

## Configuration

### Backend Environment Variables

Create `backend/.env`:

```bash
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
SERVER_ENV=development

# Database Configuration
DATABASE_URL=postgres://keyles_user:keyles_pass@localhost:5432/keyles_sso?sslmode=disable
DATABASE_MAX_CONNS=25
DATABASE_MIN_CONNS=5

# Redis Configuration
REDIS_URL=redis://localhost:6379
REDIS_DB=0
REDIS_PASSWORD=

# OAuth/OIDC Configuration
OAUTH_ISSUER=http://localhost:8080
OAUTH_ACCESS_TOKEN_TTL=900          # 15 minutes
OAUTH_REFRESH_TOKEN_TTL=604800      # 7 days
OAUTH_AUTH_CODE_TTL=300             # 5 minutes

# Security Configuration
SECURITY_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
SECURITY_COOKIE_DOMAIN=localhost
SECURITY_COOKIE_SECURE=false        # Set to true in production
SECURITY_SESSION_TTL=28800          # 8 hours

# JWT Configuration (generated by keygen)
JWT_SIGNING_KEY_PATH=./keys/private.pem
JWT_PUBLIC_KEY_PATH=./keys/public.pem
JWT_KEY_ID=dev_key_001

# Rate Limiting
RATE_LIMIT_TOKEN_ENDPOINT=10        # requests per minute per client

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json
```

### Frontend Environment Variables

Create `frontend/.env.local`:

```bash
# API Configuration
VITE_API_URL=http://localhost:8080
VITE_OAUTH_ISSUER=http://localhost:8080
VITE_CLIENT_ID=dev_client_001

# OAuth Configuration
VITE_OAUTH_REDIRECT_URI=http://localhost:3000/auth/callback
VITE_OAUTH_SCOPES=openid profile email
```

## Development Workflow

### Running Both Services

#### Terminal 1: Backend
```bash
cd backend
air  # or ./bin/server
```

#### Terminal 2: Frontend
```bash
cd frontend
npm run dev
```

#### Terminal 3: Logs (optional)
```bash
# Watch backend logs
tail -f backend/logs/app.log

# Watch Redis
redis-cli MONITOR

# Watch PostgreSQL
# Connect to psql and enable query logging
```

### Creating Test Data

```bash
cd backend

# Run seeder script (creates test tenant, users, clients)
go run cmd/seed/main.go

# This creates:
# - Tenant: "dev-tenant" (ID: tenant_dev)
# - User: admin@dev-tenant.com / password: admin123
# - User: user@dev-tenant.com / password: user123
# - Client: "Dev Client" (client_id: dev_client_001)
# - Role assignments for users
```

### Testing the OAuth Flow

1. **Start both backend and frontend**

2. **Access admin portal**: http://localhost:3000/admin
   - Login with: `admin@dev-tenant.com` / `admin123`
   - View clients, users, role assignments

3. **Test OAuth flow**:
   ```bash
   # Visit authorization endpoint
   open "http://localhost:8080/oauth2/auth?client_id=dev_client_001&redirect_uri=http://localhost:3000/auth/callback&response_type=code&scope=openid%20profile%20email&state=test123&code_challenge=E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM&code_challenge_method=S256"
   ```

4. **Manual token exchange** (using curl):
   ```bash
   # After getting authorization code from callback
   curl -X POST http://localhost:8080/oauth2/token \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "grant_type=authorization_code" \
     -d "code=AUTH_CODE_HERE" \
     -d "redirect_uri=http://localhost:3000/auth/callback" \
     -d "code_verifier=dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk" \
     -d "client_id=dev_client_001" \
     -d "client_secret=CLIENT_SECRET_HERE"
   ```

## Database Management

### Running Migrations

```bash
# Create new migration
migrate create -ext sql -dir backend/migrations -seq create_new_table

# Run migrations up
make migrate-up

# Rollback last migration
make migrate-down

# Force version (if stuck)
migrate -path backend/migrations -database $DATABASE_URL force VERSION
```

### Accessing Database

```bash
# Connect to PostgreSQL
psql postgres://keyles_user:keyles_pass@localhost:5432/keyles_sso

# Common queries
# List all tables
\dt

# View tenants
SELECT * FROM tenants;

# View clients
SELECT client_id, client_name, tenant_id FROM clients;

# View users
SELECT id, email, tenant_id, is_active FROM users;

# View role assignments
SELECT u.email, c.client_name, r.role
FROM user_role_assignments r
JOIN users u ON r.user_id = u.id
JOIN clients c ON r.client_id = c.client_id
WHERE r.is_active = true;
```

### Redis Operations

```bash
# Connect to Redis
redis-cli

# View all keys
KEYS *

# View authorization codes
KEYS auth:code:*

# View sessions
KEYS session:*

# View rate limit counters
KEYS ratelimit:*

# Get specific key
GET session:sess_abc123

# Delete key
DEL session:sess_abc123

# Clear all Redis data (DANGER!)
FLUSHALL
```

## Testing

### Backend Tests

```bash
cd backend

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./domain/...
go test ./usecase/...
go test ./infrastructure/...

# Integration tests (requires DB)
go test -tags=integration ./tests/integration/...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

### Frontend Tests

```bash
cd frontend

# Run all tests
npm test

# Run with coverage
npm run test:coverage

# Run specific test file
npm test -- LoginForm.test.tsx

# Run in watch mode
npm test -- --watch

# E2E tests (if implemented)
npm run test:e2e
```

## Debugging

### Backend Debugging (VS Code)

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Backend",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/backend/cmd/server",
      "env": {
        "SERVER_PORT": "8080"
      },
      "args": []
    }
  ]
}
```

### Frontend Debugging (VS Code)

Install [Debugger for Chrome](https://marketplace.visualstudio.com/items?itemName=msjsdiag.debugger-for-chrome)

Add to `.vscode/launch.json`:

```json
{
  "name": "Debug Frontend",
  "type": "chrome",
  "request": "launch",
  "url": "http://localhost:3000",
  "webRoot": "${workspaceFolder}/frontend/src"
}
```

### Common Issues

**Issue**: `migrate: database dirty version X`
```bash
# Fix by forcing version
migrate -path backend/migrations -database $DATABASE_URL force X
```

**Issue**: `connection refused` to PostgreSQL
```bash
# Check if PostgreSQL is running
pg_isready

# Check port
lsof -i :5432
```

**Issue**: `ECONNREFUSED` to Redis
```bash
# Check if Redis is running
redis-cli ping

# Start Redis
redis-server
```

**Issue**: `CORS error` in browser
- Verify `SECURITY_ALLOWED_ORIGINS` in backend `.env`
- Check frontend is running on allowed origin

## API Documentation

### Viewing OpenAPI Spec

```bash
# Install swagger-ui locally
npm install -g swagger-ui-watcher

# Serve the API docs
swagger-ui-watcher specs/003-sso-auth-provider/contracts/openapi.yaml
```

Visit: http://localhost:8000

### Postman Collection

Import `specs/003-sso-auth-provider/contracts/openapi.yaml` into Postman:
1. Open Postman
2. Import → Upload Files → Select `openapi.yaml`
3. Collection will be created with all endpoints

## Production Deployment Checklist

Before deploying to production:

- [ ] Set `SERVER_ENV=production`
- [ ] Enable HTTPS (`SECURITY_COOKIE_SECURE=true`)
- [ ] Use strong `DATABASE_PASSWORD`
- [ ] Enable Redis password authentication
- [ ] Generate production RSA keys (4096-bit)
- [ ] Configure proper `SECURITY_ALLOWED_ORIGINS`
- [ ] Set up database connection pooling
- [ ] Configure logging aggregation
- [ ] Enable rate limiting on load balancer
- [ ] Set up database backups
- [ ] Configure Redis persistence (AOF)
- [ ] Review and test disaster recovery plan
- [ ] Set up monitoring and alerting
- [ ] Perform security audit
- [ ] Load test the system

## Useful Commands

### Makefile Targets

Create `backend/Makefile`:

```makefile
.PHONY: build run test migrate-up migrate-down seed clean

build:
	go build -o bin/server cmd/server/main.go

run:
	go run cmd/server/main.go

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

seed:
	go run cmd/seed/main.go

clean:
	rm -rf bin/ coverage.out coverage.html

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
```

Usage:
```bash
make build
make run
make test
make migrate-up
make seed
```

## Next Steps

1. **Implement User Stories** (in priority order):
   - P1: Client registration
   - P1: User authentication flow
   - P1: Token exchange
   - P2: Token validation
   - P2: Token refresh
   - P2: Role management
   - P3: Token revocation
   - P3: Multi-client management

2. **Write Tests**:
   - Unit tests for domain layer
   - Integration tests for repositories
   - API tests for handlers

3. **Documentation**:
   - API usage examples
   - Architecture decision records
   - Deployment guide

## Additional Resources

- [Feature Specification](./spec.md)
- [Implementation Plan](./plan.md)
- [Research Document](./research.md)
- [Data Model](./data-model.md)
- [API Contracts](./contracts/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [OAuth 2.0 RFC 6749](https://datatracker.ietf.org/doc/html/rfc6749)
- [OIDC Core Spec](https://openid.net/specs/openid-connect-core-1_0.html)
- [PKCE RFC 7636](https://datatracker.ietf.org/doc/html/rfc7636)

## Getting Help

- **Architecture Questions**: Review `specs/003-sso-auth-provider/plan.md`
- **API Questions**: See `specs/003-sso-auth-provider/contracts/openapi.yaml`
- **Database Schema**: Check `specs/003-sso-auth-provider/data-model.md`
- **Technology Decisions**: Read `specs/003-sso-auth-provider/research.md`

## License

[Your License Here]
