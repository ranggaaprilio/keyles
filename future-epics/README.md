# Keyles MVP Release Epics

This directory contains the epics required to take Keyles from its current development state to a production-ready MVP release.

## Current State Summary

The codebase has a solid foundation: all 5 product specs (001-005) are fully implemented with backend (Go/Clean Architecture) and frontend (React/TypeScript). Core features include tenant registration with OTP, admin auth with JWT, OAuth 2.0/OIDC provider (PKCE, refresh rotation, introspection, revocation), client management, user management with RBAC, session management, audit logging, rate limiting, and JWKS. Comprehensive unit and integration tests exist. Docker Compose and CI/CD pipelines are in place.

However, several critical pieces are missing that prevent this from being a functional, secure, and operable MVP.

## Epic Sequence

1. **01-oauth-consent-flow** — Wire the OAuth consent screen and end-user login redirect so the SSO provider actually works end-to-end
2. **02-production-security-hardening** — HTTPS/TLS, secrets management, security headers, CSRF protection, and input sanitization
3. **03-operational-readiness** — Structured logging, metrics, health checks, automated migrations, backup strategy, and deployment documentation
4. **04-user-account-lifecycle** — Password reset flow, admin logout, account lockout, and session timeout enforcement
5. **05-end-to-end-quality** — E2E tests for critical flows, edge case hardening, input validation, and production smoke tests

## How to Use

Work through epics sequentially. Each epic contains:
- **Goal**: What the epic achieves
- **Why MVP**: Why this blocks release
- **Current State**: What exists today
- **Tasks**: Specific, verifiable work items
- **Acceptance Criteria**: How to verify completion
