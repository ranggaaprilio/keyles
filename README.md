# Keyles - Multi-Tenant SSO Platform

A modern, secure, and scalable Single Sign-On (SSO) platform built with Go and React, featuring multi-tenant architecture, email-based OTP verification, and JWT authentication.

[![CI/CD](https://github.com/ranggaaprilio/keyles/workflows/Backend%20CI/badge.svg)](https://github.com/ranggaaprilio/keyles/actions)
[![Coverage](https://codecov.io/gh/ranggaaprilio/keyles/branch/main/graph/badge.svg)](https://codecov.io/gh/ranggaaprilio/keyles)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## 🎯 Features

### ✅ Implemented (v1.0.0)

- **Multi-Tenant Architecture**: Complete tenant isolation with organization-level management
- **Email OTP Verification**: Secure 6-digit OTP with 10-minute expiration  
- **JWT Authentication**: Stateless auth with 24-hour token expiration
- **Admin Dashboard**: Post-verification access to tenant management console
- **Health Monitoring**: Comprehensive health check endpoints for all services
- **Production Ready**: Docker images, CI/CD pipelines, error boundaries
- **Accessibility**: WCAG 2.1 AA compliant forms with ARIA labels
- **RESTful API**: Clean, documented API endpoints

### 🚧 Roadmap

- [ ] OAuth 2.0 / OpenID Connect provider
- [ ] SAML 2.0 Support
- [ ] Multi-factor authentication (TOTP)
- [ ] User Management with RBAC
## Architecture

This project follows **Clean Architecture** (Hexagonal Architecture) with strict separation of concerns:

```
backend/
├── cmd/server/           # Application entry point
├── domain/               # Business entities and interfaces (NO external dependencies)
│   ├── entities/         # Core business entities (Tenant, User, OTP)
│   ├── repositories/     # Repository interfaces
│   └── services/         # Domain services interfaces
├── usecase/              # Application business rules
## Technology Stack

### Backend
- **Language**: Go 1.23+
- **Web Framework**: Gin (HTTP router and middleware)
- **ORM**: GORM with PostgreSQL driver
## Quick Start

### Prerequisites

- **Docker** 20+ and **Docker Compose** 2+ (recommended)
- **Go** 1.23+ (for local backend development)
- **Node.js** 20 LTS (for local frontend development)
- **PostgreSQL** 15+ (if running without Docker)
- **Redis** 7+ (if running without Docker)

### Running with Docker Compose (Recommended)

1. **Clone the repository**
```bash
git clone https://github.com/ranggaaprilio/keyles.git
cd keyles
```

2. **Set up environment variables**
```bash
# Create backend .env (or use defaults in docker-compose.yml)
echo "BREVO_API_KEY=your_brevo_api_key" > backend/.env
echo "JWT_SECRET=$(openssl rand -base64 32)" >> backend/.env

# Frontend environment variables are in docker-compose.yml
```

3. **Start all services**
```bash
# Start PostgreSQL, Redis, Backend, and Frontend
docker-compose up -d

# Check service health
docker-compose ps

# View logs
docker-compose logs -f backend
docker-compose logs -f frontend
```

4. **Access the application**
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **Health Checks**: 
  - http://localhost:8080/health
### Running Locally (Development)

#### Backend Setup

```bash
cd backend

# Install Go dependencies
go mod download

# Set up environment variables
cp .env.example .env
# Edit .env with your configuration

# Ensure PostgreSQL and Redis are running
# Update .env with connection details

# Run the server
go run cmd/server/main.go

# Or build and run
go build -o bin/server cmd/server/main.go
./bin/server
```

#### Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Set up environment variables  
cp .env.example .env
# Edit .env: VITE_API_URL=http://localhost:8080

# Run development server with hot reload
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

### Environment Variables

#### Backend (.env)
```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=keyles
DB_USER=keyles
DB_PASSWORD=secure_password

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=your_jwt_secret_minimum_32_characters

# Email (Brevo)
BREVO_API_KEY=your_brevo_api_key

# Server
GIN_MODE=release  # or debug for development
PORT=8080
```

#### Frontend (.env)
```env
VITE_API_URL=http://localhost:8080
VITE_APP_NAME=Keyles SSO
```backend/.env.example backend/.env
# Edit backend/.env and set BREVO_API_KEY

# Frontend
cp frontend/.env.example frontend/.env
```

### 3. Start Services

```bash
# Start all services (PostgreSQL, Redis, Backend, Frontend)
docker-compose up -d

# Check logs
docker-compose logs -f backend

# Run database migrations
cd backend
go run cmd/server/main.go migrate
```

### 4. Access Application

- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080
- **API Health Check**: http://localhost:8080/health

## Development

### Backend Development

```bash
cd backend

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run server locally
go run cmd/server/main.go
```

### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Run development server
npm run dev

# Run tests
npm test

# Build for production
npm run build
```

### Database Migrations

```bash
cd backend

# Create new migration
migrate create -ext sql -dir migrations -seq <migration_name>

# Run migrations
migrate -path migrations -database "postgres://keyles:password@localhost:5432/keyles?sslmode=disable" up

# Rollback migration
migrate -path migrations -database "postgres://keyles:password@localhost:5432/keyles?sslmode=disable" down 1
```

## Testing

### Backend Tests

```bash
cd backend

# Run all tests
go test ./...

# Run specific test packages
go test ./tests/unit/domain/...
go test ./tests/unit/usecase/...
go test ./tests/integration/...

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test ./tests/integration/ -run TestLogin -v

# Run tests with race detection
go test -race ./...
```

### Frontend Tests

```bash
cd frontend

# Run all tests
npm test

# Run with coverage
npm test -- --coverage

# Run in watch mode
npm test -- --watch

# Run specific test file
npm test -- LoginPage.test.tsx

# Run UI tests (if configured)
npm run test:ui
```

### Test Coverage Goals

- **Domain Layer**: ≥90% statement coverage ✅
- **Use Case Layer**: ≥85% statement coverage ✅  
- **Handler Layer**: ≥80% statement coverage ✅
- **Frontend**: ≥70% statement coverage ✅

### Current Test Stats

- **Backend Unit Tests**: 15+ test files
- **Backend Integration Tests**: 7+ test files
- **Frontend Unit Tests**: 3+ test files
- **Total Test Cases**: 100+ assertions

## Security

### Authentication & Authorization
- **Password Hashing**: bcrypt with cost factor 10
- **JWT Tokens**: HS256 signing algorithm, 24-hour expiration
- **Token Storage**: localStorage (frontend), memory (backend validation)
- **Protected Routes**: JWT middleware validates all dashboard endpoints

### OTP Security
- **Format**: 6-digit numeric codes
- **Expiration**: 10 minutes from generation
- **One-Time Use**: OTPs invalidated after successful verification
## Deployment

### Docker Production Deployment

```bash
# Build images
docker-compose build

# Start in production mode
## Contributing

We welcome contributions! Please follow these guidelines:

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Authors

- **Rangga Aprilio** - *Initial work* - [GitHub](https://github.com/ranggaaprilio)

## Acknowledgments

- Clean Architecture by Robert C. Martin
- Domain-Driven Design by Eric Evans
- The Go and React communities
- All contributors and supporters
   - Domain/Use Case layers: ≥85% coverage
   - All new features must have tests
5. **Lint**: Ensure code passes linting
   ```bash
   # Backend
   cd backend && golangci-lint run
   
   # Frontend
   cd frontend && npm run lint
   ```
6. **Commit**: Write clear commit messages
7. **Push**: Push to your fork
8. **PR**: Open a Pull Request with description

### Architecture Rules (NON-NEGOTIABLE)

✅ **Domain Independence**: Domain layer has ZERO infrastructure dependencies  
✅ **Dependency Direction**: All dependencies point inward  
✅ **Interface Definition**: Repositories defined in domain, implemented in infrastructure  
✅ **Test Coverage**: ≥85% for business logic layers  
✅ **Clean Code**: Single Responsibility, DRY, KISS principles

### Code Style

**Go (Backend)**:
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `go fmt` for formatting
- Use meaningful variable names
- Document exported functions and types
- Handle errors explicitly

**TypeScript (Frontend)**:
- Follow [TypeScript Style Guide](https://google.github.io/styleguide/tsguide.html)
- Use Prettier for formatting
- Use functional components and hooks
- Prefer composition over inheritance
- Type everything (avoid `any`)

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Examples:
- `feat(auth): add JWT token refresh endpoint`
- `fix(otp): resolve expiration check bug`
- `docs(readme): update installation instructions`
docker-compose ps

# View logs
docker-compose logs -f

# Scale services (if needed)
docker-compose up -d --scale backend=3
```

### Manual Deployment

#### Backend
```bash
cd backend

# Build binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-w -s" \
  -o bin/server \
  ./cmd/server

# Run with environment variables
export DB_HOST=your-db-host
export JWT_SECRET=your-secret
./bin/server
```

#### Frontend
```bash
cd frontend

# Build for production
npm run build

# Serve with nginx or any static file server
# Output is in dist/ directory
```

### Environment Configuration

Production environment checklist:
- [ ] Set strong `JWT_SECRET` (minimum 32 characters)
- [ ] Configure `BREVO_API_KEY` for email sending
- [ ] Set `GIN_MODE=release` for backend
- [ ] Use strong database passwords
- [ ] Configure proper CORS origins
- [ ] Enable HTTPS (via reverse proxy)
- [ ] Set up database backups
- [ ] Configure log aggregation
- [ ] Set up monitoring and alerting
- [ ] Enable rate limiting
- [ ] Configure firewall rules

### Health Checks

Monitor these endpoints:
- `GET /health` - Basic application health
- `GET /health/db` - Database connectivity
- `GET /health/redis` - Cache connectivity

All health endpoints return:
- `200 OK` when healthy
- `503 Service Unavailable` when unhealthy endpoints (login, registration)
- **HTTPS**: Required in production (handled by reverse proxy)

### HTTP Security Headers
```
X-Frame-Options: SAMEORIGIN
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Referrer-Policy: no-referrer-when-downgrade
```

### Best Practices
✅ Non-root containers (user 1000:appuser)
✅ Multi-stage Docker builds (minimal attack surface)
✅ Secret management via environment variables
✅ Health checks for monitoring
✅ Structured error messages (no stack traces to clients)
✅ CSRF protection via JWT in Authorization header

## API Documentation

### Core Endpoints

#### 1. Health Check Endpoints

```http
# Basic health check
GET /health
Response: { "status": "healthy", "service": "keyles-api", "version": "1.0.0" }

# Database health check  
GET /health/db
Response: { "status": "healthy", "checks": { "database": "healthy" } }

# Redis health check
GET /health/redis  
Response: { "status": "healthy", "checks": { "redis": "healthy" } }
```

#### 2. Registration Flow

**Step 1: Check Availability**
```http
GET /api/v1/check-availability?email=admin@acme.com
GET /api/v1/check-availability?organization_name=Acme

Response:
{
  "available": true,
  "field": "email"
}
```

**Step 2: Register Tenant**
```http
POST /api/v1/register
Content-Type: application/json

{
  "organization_name": "Acme Corporation",
  "admin_email": "admin@acme.com",
  "admin_full_name": "John Doe",
  "admin_password": "SecurePassword123!"
}

Response: 201 Created
{
  "message": "Registration successful. Please check your email for verification code.",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "admin@acme.com"
}
```

**Step 3: Verify OTP**
```http
POST /api/v1/verify-otp
Content-Type: application/json

{
  "email": "admin@acme.com",
  "otp_code": "123456"
}

Response: 200 OK
{
  "message": "Email verified successfully",
  "tenant_status": "active"
}
```

**Step 3b: Resend OTP (if expired)**
```http
POST /api/v1/resend-otp
Content-Type: application/json

{
  "email": "admin@acme.com"
}

Response: 200 OK
{
  "message": "New verification code sent to your email"
}
```

#### 3. Authentication Flow

**Login**
```http
POST /api/v1/login
Content-Type: application/json

{
  "email": "admin@acme.com",
  "password": "SecurePassword123!"
}

Response: 200 OK
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 86400,
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "email": "admin@acme.com",
    "full_name": "John Doe",
    "role": "admin"
  },
  "tenant": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "organization_name": "Acme Corporation",
    "status": "active"
  }
}
```

**Access Dashboard (Protected)**
```http
GET /api/v1/dashboard
Authorization: Bearer <jwt_token>

Response: 200 OK
{
  "tenant": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "organization_name": "Acme Corporation",
    "status": "active",
    "created_at": "2024-01-15T10:00:00Z",
    "verified_at": "2024-01-15T10:05:00Z"
  },
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "full_name": "John Doe",
    "email": "admin@acme.com",
    "role": "admin"
  }
}
```

### Error Responses

All endpoints return consistent error responses:

```json
{
  "error": "Error message description",
  "code": "ERROR_CODE",
  "details": {} // Optional additional context
}
```

Common HTTP status codes:
- `200` - Success
- `201` - Created
- `400` - Bad Request (validation errors)
- `401` - Unauthorized (invalid credentials or token)
- `403` - Forbidden (tenant not verified)
- `404` - Not Found (resource doesn't exist)
- `409` - Conflict (duplicate email/organization)
- `429` - Too Many Requests (rate limit exceeded)
- `500` - Internal Server Error

## Contributing

1. Read the [Constitution](.specify/memory/constitution.md) (NON-NEGOTIABLE principles)
2. Create feature branch from `main`
3. Follow Clean Architecture - domain layer has ZERO infrastructure dependencies
4. Write tests FIRST (TDD approach, ≥85% coverage for domain/usecase layers)
5. Run all tests before submitting PR
6. Ensure no lint errors (`go fmt`, `eslint`)

## License

MIT License - see [LICENSE](LICENSE) file

## Support

- **Documentation**: See `/specs/` directory for feature specifications
- **Issues**: https://github.com/yourusername/keyles/issues
- **Email**: support@keyles.com

---

**Version**: 0.1.0 | **Status**: Alpha | **Last Updated**: 2025-12-06
