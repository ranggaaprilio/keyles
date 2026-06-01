# Keyles Frontend — SSO Provider UI & Admin Portal

React + TypeScript single-page application for the Keyles multi-tenant SSO
platform. Serves the OAuth browser consent flow (login, consent, logout,
error pages) and the admin portal for tenant, client, user, and role
management.

## Features

- **OAuth Browser Consent Flow**: Login page, consent approval/denial,
  provider-local logout, and user-facing error rendering
- **Admin Portal**: OAuth client registration, user invitation and
  lifecycle management, role assignment, session listing, and audit
  activity feeds
- **OAuth Relying Party Demo**: Callback handler and PKCE-based token
  exchange for testing the authorization code flow
- **TypeScript**: Full type safety with strict mode
- **React 18**: Functional components and hooks exclusively

## Quick Start

### Prerequisites

- Node.js 20 LTS
- Backend API running (see [backend/README.md](../backend/README.md))

### Installation

```bash
cd frontend

# Install dependencies
npm install

# Copy environment file
cp .env.example .env

# Start development server (port 5173)
npm run dev
```

Visit http://localhost:5173 (dev) or http://localhost:3000 (Docker).

## Project Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── auth/              # ConsentScreen, OAuthCallback, OAuthErrorPanel
│   │   ├── admin/             # ClientManagement, UserManagement, RoleManagement
│   │   ├── ui/                # Radix UI primitives (button, dialog, select, etc.)
│   │   ├── landing/           # Public landing page components
│   │   ├── registration/      # Tenant registration form
│   │   ├── verification/      # OTP verification form
│   │   ├── clients/           # Client form components
│   │   ├── users/             # User form components
│   │   ├── dashboard/         # Dashboard widgets
│   │   ├── ErrorBoundary.tsx  # React error boundary
│   │   └── ProtectedRoute.tsx # JWT auth guard
│   ├── pages/                 # Route-level page components
│   │   ├── OAuthLoginPage.tsx       # /oauth2/login
│   │   ├── OAuthConsentPage.tsx     # /oauth2/consent
│   │   ├── OAuthLogoutPage.tsx      # /oauth2/logout
│   │   ├── OAuthErrorPage.tsx       # /oauth2/error
│   │   ├── LoginPage.tsx            # Admin login
│   │   ├── DashboardPage.tsx        # Admin dashboard
│   │   ├── ClientManagementPage.tsx # OAuth client CRUD
│   │   ├── UserManagementPage.tsx   # User invitation & lifecycle
│   │   ├── RegisterPage.tsx         # Tenant registration
│   │   ├── VerifyOTPPage.tsx        # OTP verification
│   │   ├── AcceptInvitationPage.tsx # Invitation acceptance
│   │   └── LandingPage.tsx          # Public landing
│   ├── services/              # API clients
│   │   ├── api.ts             # Axios instance with JWT interceptor
│   │   ├── oauthService.ts    # Relying-party OAuth flow (PKCE, token exchange)
│   │   ├── oauthInteractionService.ts # Provider-side browser flow (login, consent, logout)
│   │   ├── clientService.ts   # OAuth client management
│   │   ├── roleService.ts     # Role assignment
│   │   └── api/               # Domain-specific API modules (auth, tenant, user, etc.)
│   ├── hooks/                 # Custom React hooks
│   │   ├── useAuth.ts         # Admin authentication
│   │   ├── useOAuth.ts        # Relying-party OAuth flow
│   │   ├── usePKCE.ts         # S256 PKCE challenge generation
│   │   ├── useClients.ts      # Client CRUD operations
│   │   ├── useUsers.ts        # User management operations
│   │   ├── useRoles.ts        # Role assignment operations
│   │   ├── useSessions.ts     # Session listing & revocation
│   │   ├── useInvitation.ts   # Invitation validation & acceptance
│   │   ├── useTenantRegistration.ts # Registration flow
│   │   └── useOTPVerification.ts    # OTP verification flow
│   ├── stores/                # Zustand state management
│   ├── types/                 # TypeScript type definitions
│   │   ├── oauth.ts           # OAuth interaction types
│   │   ├── api.ts             # API request/response types (OpenAPI-generated)
│   │   ├── client.ts          # OAuth client types
│   │   ├── role.ts            # Role types
│   │   ├── user.ts            # User types
│   │   └── tenant.ts          # Tenant types
│   ├── utils/                 # Utility functions
│   │   ├── pkce.ts            # PKCE code verifier & challenge
│   │   └── tokenStorage.ts    # Admin JWT token persistence
│   ├── tests/                 # Test files mirroring src structure
│   ├── App.tsx                # Route definitions
│   ├── main.tsx               # Application entry point
│   └── index.css              # Tailwind base styles
├── emails/                    # React Email templates
├── tests/                     # Legacy test directory
├── package.json
├── tsconfig.json
├── vite.config.ts
├── tailwind.config.js
└── postcss.config.js
```

## OAuth Browser Consent Flow (Feature 006)

The frontend serves the **provider-side** pages that end-users interact with
during an OAuth authorization request. The backend validates the request,
stores it in a Redis transaction, and redirects the browser to these pages
with only an opaque `transaction_id`.

### Flow Overview

```
User → /oauth2/auth (backend validates)
     → /oauth2/login?transaction_id=... (frontend login form)
     → POST /oauth2/login (backend authenticates, sets SSO cookie)
     → /oauth2/consent?transaction_id=... (frontend consent screen)
     → POST /oauth2/consent (backend issues auth code, redirects to callback)
```

### Pages

| Route | Page | Description |
|-------|------|-------------|
| `/oauth2/login` | `OAuthLoginPage` | Email/password form; reads `transaction_id` from query params |
| `/oauth2/consent` | `OAuthConsentPage` | Displays client name, scopes, user identity; allow/deny buttons |
| `/oauth2/logout` | `OAuthLogoutPage` | Submits logout via POST; renders signed-out confirmation |
| `/oauth2/error` | `OAuthErrorPage` | Displays safe, user-friendly error messages |

### Interaction API Client

```typescript
import {
  submitLogin,
  getConsentDetails,
  submitConsentDecision,
  submitLogout,
  parseInteractionError,
} from '@/services/oauthInteractionService';

// Login: sends transaction_id + credentials, receives consent redirect URL
const response = await submitLogin({ transaction_id, email, password });
// → { redirect_url: "/oauth2/consent?transaction_id=..." }

// Consent details: loads trusted info for the consent screen
const details = await getConsentDetails(transactionId);
// → { client_name, scopes, user_display_name, csrf_token }

// Consent decision: approve or deny
const result = await submitConsentDecision({
  transaction_id,
  csrf_token,
  approved: true,
});
// → { redirect_url: "https://client.example.com/callback?code=...&state=..." }

// Logout: terminates provider-local session
await submitLogout();
```

## Admin Portal

The admin portal is protected by JWT authentication and serves tenant
administrators.

### Routes

| Route | Page | Description |
|-------|------|-------------|
| `/login` | `LoginPage` | Admin email/password login |
| `/dashboard` | `DashboardPage` | Tenant dashboard (requires JWT) |
| `/admin/clients` | `ClientManagementPage` | OAuth client CRUD |
| `/admin/users` | `UserManagementPage` | User invitation & lifecycle |
| `/register` | `RegisterPage` | New tenant registration |
| `/verify-otp` | `VerifyOTPPage` | Email OTP verification |
| `/invite/:token` | `AcceptInvitationPage` | Invitation acceptance |

### Admin Service Clients

```typescript
import { clientService } from '@/services/clientService';
import { roleService } from '@/services/roleService';

// Client management
const clients = await clientService.listClients();
await clientService.createClient({ name, redirectUris, type });
await clientService.rotateSecret(clientId);

// Role management
await roleService.assignRole({ userId, clientId, role });
await roleService.revokeRole(assignmentId);
```

## OAuth Relying Party Demo

The frontend includes a demo relying party callback and PKCE helpers for
testing the authorization code flow end-to-end.

```typescript
import { useOAuth } from '@/hooks/useOAuth';
import { usePKCE } from '@/hooks/usePKCE';

// Generate S256 PKCE challenge
const { codeVerifier, codeChallenge } = generateChallenge();
sessionStorage.setItem('code_verifier', codeVerifier);

// Redirect user to provider
const authURL = buildAuthURL({
  clientId: 'dev_client_001',
  redirectUri: 'http://localhost:3000/auth/callback',
  scope: 'openid profile email',
  codeChallenge,
  state: crypto.randomUUID(),
});
window.location.href = authURL;

// After callback, exchange code for tokens
const tokens = await oauthService.exchangeCodeForTokens(
  code,
  codeVerifier,
  redirectUri
);
```

## Development

### Available Scripts

```bash
npm run dev      # Start Vite dev server (port 5173)
npm run build    # Type-check and build production assets
npm run preview  # Preview production build
npm run lint     # ESLint with zero warnings
npm run test     # Vitest with jsdom
```

### Testing

```bash
# Run all tests
npm run test

# Run tests in watch mode
npm run test -- --watch

# Run a specific test file
npm run test -- OAuthLoginPage.test.tsx

# Run with coverage
npm run test -- --coverage
```

### Code Quality

- **TypeScript strict mode**: Full type safety
- **ESLint**: Zero warnings enforced
- **Functional components only**: No class components
- **PascalCase naming**: For React components
- **Tailwind CSS**: Utility-first styling

## Environment Variables

Configure in `.env`:

```bash
# API Configuration
VITE_API_URL=http://localhost:8080

# OAuth Configuration (relying party demo)
VITE_OAUTH_ISSUER=http://localhost:8080
VITE_CLIENT_ID=dev_client_001
VITE_OAUTH_REDIRECT_URI=http://localhost:3000/auth/callback
VITE_OAUTH_SCOPES=openid profile email

# Application
VITE_APP_NAME=Keyles SSO
VITE_APP_DESCRIPTION=Multi-Tenant SSO Platform

# Feature Flags
VITE_ENABLE_DEBUG=true
```

## Security

- **Provider SSO Cookie**: The backend sets a host-only, HttpOnly,
  SameSite=Lax cookie (`keyles_sso`). The frontend never accesses it;
  all browser-flow API calls use `credentials: 'include'`
- **Admin JWT**: Stored in memory (not localStorage) and sent via
  `Authorization: Bearer` header; refreshed automatically
- **PKCE**: S256 mandatory for the relying party demo flow
- **CSRF**: Consent decisions include a per-transaction CSRF token
- **CORS**: Backend validates allowed origins; credentials enabled for
  browser-flow endpoints

## Production Build

```bash
npm run build
# Output in dist/ — deploy to any static file server
```

The Dockerfile uses `npm run build` and serves the output via a
lightweight static server on port 3000.

### Production Checklist

- [ ] Set `VITE_API_URL` to production API URL
- [ ] Set `VITE_OAUTH_ISSUER` to production issuer
- [ ] Set `VITE_OAUTH_REDIRECT_URI` to production callback URL
- [ ] Enable HTTPS for all URLs
- [ ] Configure CORS origins on backend
- [ ] Remove `VITE_ENABLE_DEBUG` or set to `false`

## Documentation

- [Feature 003: SSO Auth Provider](../specs/003-sso-auth-provider/spec.md)
- [Feature 004: Client App Registration](../specs/004-client-app-registration/spec.md)
- [Feature 005: User Management & RBAC](../specs/005-user-management-rbac/spec.md)
- [Feature 006: OAuth Consent Flow](../specs/006-oauth-consent-flow/spec.md)
- [Backend README](../backend/README.md)

## Troubleshooting

### CORS Errors

Ensure the backend `CORS_ALLOWED_ORIGINS` includes your frontend URL:

```
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
```

### 401 on Browser-Flow Endpoints

The `keyles_sso` cookie requires `credentials: 'include'` on fetch
requests. The `oauthInteractionService` already sets this; ensure
custom calls do the same.

### PKCE Validation Errors

- `code_verifier` must match the `code_challenge` sent in the
  authorization request
- Store `code_verifier` in `sessionStorage` before redirecting
- Challenge method must be S256

## License

MIT License — see [LICENSE](../LICENSE)
