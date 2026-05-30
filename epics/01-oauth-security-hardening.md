# Epic 01: OAuth Security Hardening

## Goal

Make the current OAuth/OIDC provider safe enough to support production user management and role-based access.

## Why This Matters

The repo already has OAuth authorization, token, revoke, userinfo, discovery, and client-management flows. Before adding more admin features, close security gaps in the current auth path, especially placeholder PKCE validation in `backend/infrastructure/services/fosite_oauth_provider.go`.

## Scope

- Implement real S256 PKCE validation.
- Fix authorization URL construction.
- Add refresh token rotation or explicit refresh token reuse detection.
- Ensure token validation checks revoked clients and disabled users.
- Include role claims in access tokens and ID tokens.
- Add integration tests for PKCE, revoked clients, expired authorization codes, refresh flows, and token validation.

## Acceptance Criteria

- Invalid PKCE verifier fails token exchange.
- JWTs include tenant, client, subject, scopes, and roles.
- Deleted or disabled clients cannot issue or refresh tokens.
- OAuth integration tests pass.

## Suggested First Tasks

- Add failing tests around PKCE verifier mismatch.
- Replace placeholder PKCE logic with SHA-256 + base64url verification.
- Add token claim tests before changing token signing.
