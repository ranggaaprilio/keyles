# API Contracts: End-User Management with RBAC

**Feature**: [spec.md](../spec.md)  
**OpenAPI Spec**: [openapi.yaml](./openapi.yaml)

## Overview

This directory contains the OpenAPI 3.0 contract for the End-User Management & RBAC API.

## Endpoint Summary

### User Management (Admin only)

| Method   | Path                                          | Description                              |
| -------- | --------------------------------------------- | ---------------------------------------- |
| `GET`    | `/api/v1/admin/users`                         | List users (paginated, search, filter)   |
| `POST`   | `/api/v1/admin/users`                         | Invite a new user                        |
| `GET`    | `/api/v1/admin/users/{user_id}`               | Get user details + role assignments      |
| `PATCH`  | `/api/v1/admin/users/{user_id}`               | Update user display name                 |
| `DELETE` | `/api/v1/admin/users/{user_id}`               | Permanently delete a user                |
| `PUT`    | `/api/v1/admin/users/{user_id}/status`        | Enable or disable a user account         |
| `POST`   | `/api/v1/admin/users/{user_id}/resend-invitation` | Resend invitation for pending user  |

### Invitation Acceptance (Public)

| Method | Path                                  | Description                            |
| ------ | ------------------------------------- | -------------------------------------- |
| `POST` | `/api/v1/invitations/{token}/accept`  | Accept invitation & set password       |

### Role Management (Admin only)

| Method   | Path                                                    | Description                        |
| -------- | ------------------------------------------------------- | ---------------------------------- |
| `GET`    | `/api/v1/admin/users/{user_id}/roles`                   | List all role assignments for user |
| `POST`   | `/api/v1/admin/users/{user_id}/roles`                   | Assign a role to a user            |
| `DELETE` | `/api/v1/admin/users/{user_id}/roles/{assignment_id}`   | Revoke a role assignment           |

### Session Management (Admin only)

| Method   | Path                                                       | Description                       |
| -------- | ---------------------------------------------------------- | --------------------------------- |
| `GET`    | `/api/v1/admin/users/{user_id}/sessions`                   | List active sessions              |
| `DELETE` | `/api/v1/admin/users/{user_id}/sessions/{session_id}`      | Terminate a specific session      |

### User Activity (Admin only)

| Method | Path                                        | Description                          |
| ------ | ------------------------------------------- | ------------------------------------ |
| `GET`  | `/api/v1/admin/users/{user_id}/activity`    | Paginated activity log for user      |

## JWT Claim Changes

When issuing access tokens and ID tokens for a user authenticating via a client application,
the token payload is extended with:

```json
{
  "sub": "usr_aB3dEfGhIjKlMn",
  "tenant_id": "acme-corp",
  "client_id": "aBcDeFgHiJkLmNo",
  "roles": ["analyst", "report-exporter"],
  "exp": 1749636600,
  "iat": 1749635700
}
```

**Rules**:
- `roles` contains only the active role assignments for the **authenticating client** (`client_id`).
- Users with **no active roles** for the client are denied authorization before a token is issued.
- The userinfo endpoint also returns `roles` in its JSON response.

## Authentication

All admin endpoints require a valid JWT Bearer token from a **tenant administrator**.

```
Authorization: Bearer <admin_access_token>
```

The `/api/v1/invitations/{token}/accept` endpoint is **public** (no authentication required).

## Error Codes

| HTTP Status | `error` field       | Meaning                                             |
| ----------- | ------------------- | --------------------------------------------------- |
| 400         | `validation_error`  | Request body failed field-level validation          |
| 400         | `token_expired`     | Invitation token has passed its 72-hour window      |
| 400         | `token_used`        | Invitation token has already been accepted          |
| 400         | `invalid_state`     | Operation not valid for the resource's current state|
| 401         | `unauthorized`      | Missing or invalid Bearer token                     |
| 403         | `forbidden`         | Valid token but insufficient permissions            |
| 404         | `not_found`         | Resource does not exist in this tenant              |
| 409         | `conflict`          | Duplicate resource (email, role assignment)         |
| 409         | `quota_exceeded`    | Tenant has reached the user or role quota limit     |
