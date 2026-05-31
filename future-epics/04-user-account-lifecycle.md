# Epic 04: User Account Lifecycle

## Goal

Complete the user account lifecycle: password reset for admins, a proper logout mechanism, account lockout after repeated failed attempts, session timeout enforcement, and email verification for admin accounts. This ensures basic account security expectations are met for both tenant admins and end-users.

## Why MVP

An SSO platform without password reset means locked-out admins have no recovery path — a critical support issue on day one. Without logout, sessions cannot be terminated client-side. Without account lockout, brute-force attacks are only gated by rate limiting. Without session timeouts, stolen tokens have unlimited validity windows. These are table-stakes account security features.

## Current State

- **Password reset**: The OTP domain entity supports `purpose: "password_reset"` but no use case, handler, route, or frontend page exists for it. Only `email_verification` purpose is implemented.
- **Logout**: No logout endpoint exists. The admin JWT is stored in `localStorage` on the frontend and cleared on logout UI action, but the token itself remains valid until expiry. No server-side token invalidation.
- **Account lockout**: Rate limiting restricts requests per time window, but there is no progressive lockout (temporary lock after N failures, longer lock after more failures).
- **Session timeout**: Admin JWT expiration is configurable (`JWT_EXPIRATION_HOURS`, default 24h) but there is no idle timeout — a token stays valid for its full lifetime regardless of activity.
- **Email verification for admins**: Admin email is not verified after registration (tenant OTP verifies the organization email, but the admin can log in without their own email being verified separately).

## Tasks

### 4.1 — Password reset flow (backend)
- Create `usecase/auth/request_password_reset.go` — generates OTP with `purpose: "password_reset"`, sends email with reset link/code
- Create `usecase/auth/reset_password.go` — verifies OTP, validates new password, updates user password, invalidates all existing sessions
- Create `POST /api/v1/password-reset/request` handler — accepts email, triggers reset email
- Create `POST /api/v1/password-reset/confirm` handler — accepts token + new password, completes reset
- Rate limit password reset requests (prevent email abuse)
- OTP for password reset should have shorter expiry (e.g., 5 minutes)

### 4.2 — Password reset flow (frontend)
- Create `ForgotPasswordPage.tsx` — email input form that calls reset request endpoint
- Create `ResetPasswordPage.tsx` — new password form that calls reset confirm endpoint
- Add routes `/forgot-password` and `/reset-password` in `App.tsx`
- Show success/error states appropriately
- Add "Forgot password?" link on login page

### 4.3 — Server-side logout with token invalidation
- Create `POST /api/v1/logout` endpoint
- Maintain a Redis-based token blacklist for revoked admin JWTs (similar to existing `userBlacklist` for end-users)
- On logout: add JWT `jti` to blacklist with TTL matching the token's remaining lifetime
- Update auth middleware to check blacklist on every request
- Frontend: call `/api/v1/logout` on logout action, clear localStorage

### 4.4 — Account lockout after repeated failures
- Track failed login attempts per account in Redis with a sliding window
- After N consecutive failures (configurable, default 5): temporarily lock account for M minutes (configurable, default 15)
- After 3x lockouts in 24 hours: lock for 1 hour
- Successful login resets the failure counter
- Return appropriate error message (don't reveal if account exists vs locked)
- Add admin unlock endpoint for support (optional, can be manual DB operation initially)

### 4.5 — Session idle timeout
- Track last activity timestamp per admin session in Redis
- On each authenticated request: update last activity timestamp
- On authentication: check if session has been idle beyond `SESSION_IDLE_TIMEOUT` (configurable, default 30 min)
- If idle timeout exceeded: reject with 401, require re-login
- Add `SESSION_ABSOLUTE_TIMEOUT` (configurable) — session invalid regardless of activity after this duration
- Frontend: handle 401 responses by redirecting to login

### 4.6 — Admin email verification on registration
- After tenant registration + OTP verification, send a separate verification email to the admin's email address
- Admin cannot log in until their email is verified
- Add `POST /api/v1/verify-admin-email` endpoint
- Frontend: show "Verify your email" prompt after registration, handle verification link

## Acceptance Criteria

1. Admin can request a password reset via email, receive a reset link, and set a new password
2. After password reset, all existing sessions are invalidated
3. Logout clears the session server-side; the JWT cannot be reused after logout
4. After 5 consecutive failed login attempts, the account is temporarily locked (15 min)
5. After 15 minutes of inactivity, the admin session is invalidated and requires re-login
6. Admin email must be verified before first login after registration
7. Rate limiting prevents abuse of password reset emails
8. Error messages during login do not distinguish between "account not found", "account locked", and "wrong password"
