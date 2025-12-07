# Keyles - Multi-Tenant SSO Platform

A modern, secure Single Sign-On (SSO) platform built with Clean Architecture principles, supporting multi-tenant organizations with OIDC authentication flows.

## Features

### ✅ Implemented (v0.1.0)
- **Multi-Tenant Registration**: Organizations can register via self-service form
- **Email OTP Verification**: Secure email verification with 10-minute expiration
- **Admin Dashboard**: Post-verification access to tenant management

### 🚧 In Development
- OIDC Provider Configuration
- SAML 2.0 Support
- User Management
- Role-Based Access Control (RBAC)

## Architecture

This project follows **Clean Architecture** (Hexagonal Architecture) principles:

```
backend/
├── domain/          # Business entities and interfaces (no external dependencies)
├── usecase/         # Application business rules
├── infrastructure/  # External implementations (PostgreSQL, Redis, Brevo)
└── interfaces/      # HTTP handlers, controllers
```

**Technology Stack**:
- **Backend**: Go 1.22+, Gin Framework, GORM, PostgreSQL 15+, Redis 7+
- **Frontend**: React 18+, TypeScript (strict mode), shadcn/ui, TanStack Query
- **Email**: Brevo API for transactional emails
- **Testing**: Go testing + testify + gomock, Vitest + React Testing Library

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.22+ (for local development)
- Node.js 20.x LTS (for frontend development)

### 1. Clone and Setup

```bash
git clone https://github.com/yourusername/keyles.git
cd keyles
```

### 2. Configure Environment

```bash
# Backend
cp backend/.env.example backend/.env
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

### Run All Tests

```bash
# Backend unit tests
cd backend && go test ./tests/unit/...

# Backend integration tests
cd backend && go test ./tests/integration/...

# Frontend tests
cd frontend && npm test

# Check test coverage
cd backend && go test -cover ./domain/... ./usecase/...
```

## Project Structure

See [specs/001-tenant-registration/plan.md](specs/001-tenant-registration/plan.md) for complete directory structure and architectural details.

## API Documentation

API contracts are defined in OpenAPI 3.0 format:
- [REST API Specification](specs/001-tenant-registration/contracts/api.yaml)

Access Swagger UI when backend is running:
- http://localhost:8080/swagger/index.html

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
