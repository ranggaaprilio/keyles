# Epic 03: User Invitation & Activation Flow

## Goal

Let tenant administrators invite users by email, and let invited users activate their account by setting a password.

## Why This Matters

This is the first usable end-user management workflow. It creates a path from admin action to active tenant user without manual database work.

## Scope

- Implement backend use cases:
  - `InviteUser`
  - `AcceptInvitation`
  - `ResendInvitation`
- Add admin route `POST /api/v1/admin/users/invite`.
- Add public route `POST /api/v1/invitations/:token/accept`.
- Add invitation email support with a 72-hour token.
- Add frontend invite dialog.
- Add public invitation acceptance page.
- Add password setup form with strength validation.

## Acceptance Criteria

- Admin can invite a user by email and display name.
- Invitation creates a pending user and a hashed token record.
- Invited user can activate the account with a strong password.
- Expired or already-used invitations return `410 Gone`.
- Duplicate pending invitations return a conflict response.

## Suggested First Tasks

- Write use case tests for invite, accept, and resend.
- Implement backend use cases before handlers.
- Build the frontend flow against typed API functions.
