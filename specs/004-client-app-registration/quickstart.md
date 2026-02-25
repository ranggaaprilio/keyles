# Quickstart: OAuth Client Application Registration Portal

**Date**: 2026-02-25  
**Feature**: [spec.md](./spec.md)  
**Plan**: [plan.md](./plan.md)

## Prerequisites

- Go 1.23+ installed
- Node.js 18+ and npm installed
- PostgreSQL running (with `pg_trgm` extension)
- Redis running
- Docker + Docker Compose (for local infra)

## Quick Setup

### 1. Start Infrastructure

```bash
cd /path/to/keyles
docker-compose up -d postgres redis
```

### 2. Run Database Migration

```bash
cd backend
# Apply the new migration for client_type and description
migrate -path migrations -database "postgres://user:pass@localhost:5432/keyles?sslmode=disable" up
```

### 3. Start Backend

```bash
cd backend
go run cmd/server/main.go
# Server starts on http://localhost:8080
```

### 4. Start Frontend

```bash
cd frontend
npm install
npm run dev
# Frontend starts on http://localhost:5173
```

## API Usage Examples

### Register a Confidential Client

```bash
# Login as tenant admin first
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"your-password"}' | jq -r '.token')

# Register a confidential client
curl -X POST http://localhost:8080/api/v1/admin/clients \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "client_name": "My Web App",
    "description": "Production web application",
    "client_type": "confidential",
    "redirect_uris": ["https://myapp.example.com/callback"]
  }'

# Response includes client_id and client_secret (one-time display)
# {
#   "client_id": "aBcDeFg...",
#   "client_secret": "sEcReT...",   <-- SAVE THIS! Cannot be retrieved later
#   "client_name": "My Web App",
#   "client_type": "confidential",
#   ...
# }
```

### Register a Public Client (SPA/Mobile)

```bash
curl -X POST http://localhost:8080/api/v1/admin/clients \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "client_name": "My Mobile App",
    "description": "iOS mobile application",
    "client_type": "public",
    "redirect_uris": [
      "https://myapp.example.com/callback",
      "http://localhost:3000/callback"
    ]
  }'

# Response has NO client_secret (public clients use PKCE only)
# {
#   "client_id": "xYzAbCd...",
#   "client_name": "My Mobile App",
#   "client_type": "public",
#   ...
# }
```

### List Clients (Paginated)

```bash
# List all clients
curl http://localhost:8080/api/v1/admin/clients \
  -H "Authorization: Bearer $TOKEN"

# Search by name with pagination
curl "http://localhost:8080/api/v1/admin/clients?search=mobile&page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN"
```

### Update Client Configuration

```bash
curl -X PUT http://localhost:8080/api/v1/admin/clients/aBcDeFg... \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated description",
    "redirect_uris": [
      "https://myapp.example.com/callback",
      "https://staging.myapp.example.com/callback"
    ]
  }'
```

### Rotate Client Secret

```bash
curl -X POST http://localhost:8080/api/v1/admin/clients/aBcDeFg.../rotate-secret \
  -H "Authorization: Bearer $TOKEN"

# Response includes new secret (one-time display)
# Old secret is immediately invalidated
```

### Delete Client

```bash
curl -X DELETE http://localhost:8080/api/v1/admin/clients/aBcDeFg... \
  -H "Authorization: Bearer $TOKEN"

# Returns 204 No Content
# All tokens for this client are immediately revoked
```

## Using Client Credentials in OAuth Flows

### Authorization Code Flow (Confidential Client)

```bash
# 1. Redirect user to authorize endpoint
# https://sso.keyles.com/oauth2/auth?
#   client_id=YOUR_CLIENT_ID&
#   redirect_uri=https://myapp.example.com/callback&
#   response_type=code&
#   scope=openid%20profile%20email&
#   state=random-csrf-token&
#   code_challenge=BASE64URL(SHA256(code_verifier))&
#   code_challenge_method=S256

# 2. Exchange authorization code for tokens
curl -X POST http://localhost:8080/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&\
code=AUTH_CODE_FROM_CALLBACK&\
redirect_uri=https://myapp.example.com/callback&\
code_verifier=YOUR_CODE_VERIFIER&\
client_id=YOUR_CLIENT_ID&\
client_secret=YOUR_CLIENT_SECRET"
```

### Authorization Code Flow with PKCE (Public Client)

```bash
# Same as above but WITHOUT client_secret
curl -X POST http://localhost:8080/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&\
code=AUTH_CODE_FROM_CALLBACK&\
redirect_uri=https://myapp.example.com/callback&\
code_verifier=YOUR_CODE_VERIFIER&\
client_id=YOUR_CLIENT_ID"
```

## Running Tests

### Backend

```bash
cd backend
go test ./usecase/client/... -v -cover
go test ./tests/unit/... -v
go test ./tests/integration/... -v
```

### Frontend

```bash
cd frontend
npm run test
```

## Key Implementation Files

| Layer          | File                                                               | Purpose                     |
| -------------- | ------------------------------------------------------------------ | --------------------------- |
| Domain         | `backend/domain/entities/client.go`                                | Client entity + validation  |
| Domain         | `backend/domain/repositories/client_repository.go`                 | Repository interface        |
| Use Case       | `backend/usecase/client/*.go`                                      | 6 use cases (CRUD + rotate) |
| Infrastructure | `backend/infrastructure/persistence/postgres/client_repository.go` | PostgreSQL implementation   |
| Infrastructure | `backend/infrastructure/persistence/redis/client_count_cache.go`   | Redis count cache           |
| Interface      | `backend/interfaces/http/handlers/client_handler.go`               | HTTP handlers               |
| Interface      | `backend/interfaces/http/router.go`                                | Route definitions           |
| Migration      | `backend/migrations/000008_add_client_type_and_description.*`      | Schema migration            |
| Frontend       | `frontend/src/pages/ClientManagementPage.tsx`                      | Main page                   |
| Frontend       | `frontend/src/components/clients/*.tsx`                            | UI components               |
| Frontend       | `frontend/src/hooks/useClients.ts`                                 | Data fetching hooks         |
| Frontend       | `frontend/src/services/clientService.ts`                           | API client                  |
