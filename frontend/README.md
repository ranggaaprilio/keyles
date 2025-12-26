# Keyles Frontend - OAuth 2.0 SSO Integration

React + TypeScript frontend for the Keyles multi-tenant SSO platform with OAuth 2.0 + OIDC integration.

## Features

- **OAuth 2.0 Client**: Full Authorization Code Flow with PKCE support
- **Admin Portal**: Client management, role assignments, user administration
- **Consent UI**: OAuth consent screen for user authorization
- **JWT Handling**: Secure token storage and automatic refresh
- **TypeScript**: Full type safety with strict mode
- **React 18**: Functional components with hooks only

## Quick Start

### Prerequisites

- Node.js 18+
- Backend API running (see `backend/README.md`)

### Installation

```bash
cd frontend

# Install dependencies
npm install

# Copy environment file
cp .env.example .env.local

# Edit .env.local with your configuration
nano .env.local

# Start development server
npm run dev
```

Visit http://localhost:5173

## OAuth Configuration

### Environment Variables

Configure OAuth in `.env.local`:

```bash
# API Configuration
VITE_API_URL=http://localhost:8080

# OAuth Configuration
VITE_OAUTH_ISSUER=http://localhost:8080
VITE_CLIENT_ID=dev_client_001
VITE_OAUTH_REDIRECT_URI=http://localhost:3000/auth/callback
VITE_OAUTH_SCOPES=openid profile email

# Application
VITE_APP_NAME=Keyles SSO
```

## Project Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── auth/              # OAuth authentication components
│   │   │   ├── ConsentScreen.tsx
│   │   │   ├── OAuthCallback.tsx
│   │   │   └── LoginForm.tsx
│   │   ├── admin/             # Admin portal components
│   │   │   ├── ClientManagement.tsx
│   │   │   ├── ClientForm.tsx
│   │   │   ├── RoleManagement.tsx
│   │   │   └── UserRoles.tsx
│   │   └── common/            # Shared components
│   ├── services/              # API clients
│   │   ├── api.ts             # Base API client
│   │   ├── oauthService.ts    # OAuth flow API
│   │   ├── clientService.ts   # Client management API
│   │   └── roleService.ts     # Role management API
│   ├── hooks/                 # Custom React hooks
│   │   ├── useAuth.ts
│   │   ├── useOAuth.ts        # OAuth flow hook
│   │   └── usePKCE.ts         # PKCE generation hook
│   ├── types/                 # TypeScript definitions
│   │   ├── oauth.ts           # OAuth types
│   │   ├── client.ts          # Client types
│   │   └── role.ts            # Role types
│   ├── utils/                 # Utility functions
│   │   ├── pkce.ts            # PKCE helper
│   │   └── tokenStorage.ts    # Token storage helper
│   └── tests/                 # Unit and integration tests
├── package.json
├── tsconfig.json
├── vite.config.ts
└── tailwind.config.js
```

## OAuth Integration Guide

### 1. Implementing OAuth Login

```typescript
import { useOAuth } from '@/hooks/useOAuth';

function LoginButton() {
  const { initiateLogin } = useOAuth();

  const handleLogin = async () => {
    // Generates PKCE challenge and redirects to authorization endpoint
    await initiateLogin({
      clientId: import.meta.env.VITE_CLIENT_ID,
      redirectUri: import.meta.env.VITE_OAUTH_REDIRECT_URI,
      scope: 'openid profile email',
    });
  };

  return <button onClick={handleLogin}>Sign in with SSO</button>;
}
```

### 2. Handling OAuth Callback

```typescript
import { OAuthCallback } from '@/components/auth/OAuthCallback';

// In your router configuration
<Route path="/auth/callback" element={<OAuthCallback />} />
```

### 3. Using PKCE

```typescript
import { usePKCE } from "@/hooks/usePKCE";

function Component() {
  const { generateChallenge, verifyChallenge } = usePKCE();

  // Generate code_verifier and code_challenge
  const { codeVerifier, codeChallenge } = await generateChallenge();

  // Store code_verifier securely
  sessionStorage.setItem("code_verifier", codeVerifier);

  // Use code_challenge in authorization request
  // Later, retrieve code_verifier for token exchange
}
```

### 4. Managing Tokens

```typescript
import { tokenStorage } from "@/utils/tokenStorage";

// Store tokens after successful authentication
tokenStorage.setAccessToken(accessToken);
tokenStorage.setRefreshToken(refreshToken);

// Retrieve tokens
const accessToken = tokenStorage.getAccessToken();

// Clear tokens on logout
tokenStorage.clearTokens();
```

### 5. Automatic Token Refresh

The API client automatically refreshes expired access tokens using refresh tokens:

```typescript
// Configured in src/services/api.ts
axios.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Attempt to refresh token
      const newAccessToken = await refreshAccessToken();
      // Retry original request with new token
    }
    return Promise.reject(error);
  }
);
```

## Admin Portal Features

### Client Management

Admin users can:

- Register new OAuth clients
- View all clients for their tenant
- Update client configurations (name, redirect URIs)
- Rotate client secrets
- Delete clients

```typescript
import { ClientManagement } from '@/components/admin/ClientManagement';

<Route path="/admin/clients" element={<ClientManagement />} />
```

### Role Management

Admin users can:

- Assign roles to users for specific clients
- Revoke user roles
- View role assignments
- Control user access to client applications

```typescript
import { RoleManagement } from '@/components/admin/RoleManagement';

<Route path="/admin/roles" element={<RoleManagement />} />
```

## Development

### Available Scripts

```bash
npm run dev        # Start development server
npm run build      # Build for production
npm run preview    # Preview production build
npm run lint       # Run ESLint
npm test           # Run tests
npm run test:coverage  # Generate coverage report
```

### Testing

```bash
# Run all tests
npm test

# Run tests in watch mode
npm test -- --watch

# Generate coverage report
npm run test:coverage

# Run specific test file
npm test -- LoginForm.test.tsx
```

### Code Quality

The project enforces:

- **TypeScript strict mode**: Full type safety
- **ESLint**: Code linting and style checking
- **Prettier**: Code formatting (if configured)
- **Functional components only**: No class components (per constitution)
- **PascalCase naming**: For all React components

## API Integration

### Base API Configuration

```typescript
// src/services/api.ts
import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

// Add authorization header
api.interceptors.request.use((config) => {
  const token = tokenStorage.getAccessToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
```

### OAuth Service

```typescript
// src/services/oauthService.ts
export const oauthService = {
  // Build authorization URL with PKCE
  buildAuthURL: (params: AuthParams) => string,

  // Exchange authorization code for tokens
  exchangeCodeForTokens: (code: string, codeVerifier: string) =>
    Promise<TokenResponse>,

  // Refresh access token
  refreshAccessToken: (refreshToken: string) => Promise<TokenResponse>,

  // Revoke token
  revokeToken: (token: string) => Promise<void>,
};
```

### Client Service

```typescript
// src/services/clientService.ts
export const clientService = {
  createClient: (data: CreateClientRequest) => Promise<Client>,
  listClients: () => Promise<Client[]>,
  getClient: (clientId: string) => Promise<Client>,
  updateClient: (clientId: string, data: UpdateClientRequest) =>
    Promise<Client>,
  deleteClient: (clientId: string) => Promise<void>,
  rotateSecret: (clientId: string) => Promise<{ clientSecret: string }>,
};
```

## Testing OAuth Flow

### 1. Start Backend

```bash
cd backend
make dev  # Starts Docker, runs migrations, seeds data
```

### 2. Start Frontend

```bash
cd frontend
npm run dev
```

### 3. Test Login Flow

1. Click "Sign in with SSO" button
2. Redirected to authorization endpoint
3. Enter credentials: `admin@dev-tenant.com` / `admin123`
4. Consent screen (if required)
5. Redirected back with authorization code
6. Code exchanged for tokens automatically
7. Access token stored and used for API requests

### 4. Admin Portal Access

Visit http://localhost:3000/admin (requires admin role)

## Security Considerations

- **PKCE**: Always use PKCE for OAuth flows (prevents authorization code interception)
- **Token Storage**: Store tokens in memory or sessionStorage, never in localStorage
- **HTTPS**: Use HTTPS in production (`SECURITY_COOKIE_SECURE=true`)
- **CSP**: Configure Content Security Policy headers
- **CORS**: Backend validates allowed origins
- **XSS Protection**: React escapes content by default, avoid `dangerouslySetInnerHTML`

## Production Build

```bash
# Build for production
npm run build

# Preview production build
npm run preview

# The build output will be in dist/
# Deploy dist/ to your hosting provider
```

### Production Checklist

- [ ] Set `VITE_API_URL` to production API URL
- [ ] Set `VITE_OAUTH_ISSUER` to production issuer
- [ ] Set `VITE_OAUTH_REDIRECT_URI` to production callback URL
- [ ] Enable HTTPS for all URLs
- [ ] Configure proper CORS on backend
- [ ] Set up CDN for static assets
- [ ] Enable error tracking (Sentry, etc.)
- [ ] Configure monitoring and analytics

## Documentation

- [Feature Specification](../specs/003-sso-auth-provider/spec.md)
- [Quickstart Guide](../specs/003-sso-auth-provider/quickstart.md)
- [API Contracts](../specs/003-sso-auth-provider/contracts/openapi.yaml)
- [Backend README](../backend/README.md)

## Troubleshooting

### CORS Errors

Ensure backend `CORS_ALLOWED_ORIGINS` includes your frontend URL:

```bash
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
```

### Token Refresh Failures

Check that:

- Refresh token is valid and not expired (7 days)
- Refresh token hasn't been revoked
- Client credentials are correct

### PKCE Validation Errors

Verify that:

- `code_verifier` matches the `code_challenge` used in authorization request
- `code_verifier` is stored securely in sessionStorage
- Challenge method is S256 (SHA-256)

## License

[Your License Here]
