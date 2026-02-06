# API Contracts

This directory contains API specifications for the Core SSO Auth Provider.

## Files

- **[openapi.yaml](./openapi.yaml)**: Complete OpenAPI 3.0 specification for all endpoints

## Endpoint Categories

### OAuth 2.0 / OIDC Endpoints

- `GET /oauth2/auth` - Authorization endpoint (redirect to login)
- `POST /oauth2/token` - Token endpoint (code exchange, token refresh)
- `POST /oauth2/revoke` - Token revocation
- `GET /oauth2/userinfo` - User profile information

### Discovery Endpoints

- `GET /.well-known/openid-configuration` - OIDC discovery document
- `GET /.well-known/jwks.json` - Public keys for JWT verification

### Admin Endpoints (Session-protected)

**Client Management**:

- `GET /admin/clients` - List clients
- `POST /admin/clients` - Create client
- `GET /admin/clients/{client_id}` - Get client details
- `PUT /admin/clients/{client_id}` - Update client
- `DELETE /admin/clients/{client_id}` - Delete client
- `POST /admin/clients/{client_id}/secret` - Regenerate secret

**Role Management**:

- `GET /admin/roles` - List role assignments
- `POST /admin/roles` - Assign role to user
- `DELETE /admin/roles/{assignment_id}` - Revoke role

## Authentication

### OAuth Endpoints

- **No auth** required for `/oauth2/auth` (redirects to login if needed)
- **Client credentials** (`client_id` + `client_secret`) for `/oauth2/token` and `/oauth2/revoke`
- **Bearer token** for `/oauth2/userinfo`

### Admin Endpoints

- **Session cookie** (`session_id`) required for all `/admin/*` endpoints
- Session obtained by logging into admin portal

## Rate Limiting

- **Token endpoint**: 10 requests/minute per `client_id`
- Returns `429 Too Many Requests` with `Retry-After` header when exceeded

## Viewing the Spec

### Online Viewer

Upload `openapi.yaml` to [Swagger Editor](https://editor.swagger.io/)

### Local Viewer

```bash
# Install swagger-ui
npm install -g swagger-ui-watcher

# Serve the spec
swagger-ui-watcher contracts/openapi.yaml
```

### VS Code Extension

Install [OpenAPI (Swagger) Editor](https://marketplace.visualstudio.com/items?itemName=42Crunch.vscode-openapi) extension

## Code Generation

### Go Server Stubs

```bash
# Install oapi-codegen
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest

# Generate server code
oapi-codegen -package handlers -generate types,chi-server contracts/openapi.yaml > interfaces/http/handlers/generated.go
```

### TypeScript Client

```bash
# Install openapi-typescript
npm install -D openapi-typescript

# Generate TypeScript types
npx openapi-typescript contracts/openapi.yaml -o frontend/src/types/api.ts
```

## Standards Compliance

This API adheres to:

- **RFC 6749**: OAuth 2.0 Authorization Framework
- **RFC 6750**: OAuth 2.0 Bearer Token Usage
- **RFC 7009**: OAuth 2.0 Token Revocation
- **RFC 7517**: JSON Web Key (JWK)
- **RFC 7519**: JSON Web Token (JWT)
- **RFC 7636**: Proof Key for Code Exchange (PKCE)
- **RFC 8414**: OAuth 2.0 Authorization Server Metadata
- **OpenID Connect Core 1.0**: Authentication layer on top of OAuth 2.0
