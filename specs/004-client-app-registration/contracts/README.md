# API Contracts: OAuth Client Application Registration

**Feature**: [spec.md](../spec.md)  
**Date**: 2026-02-25

## Overview

This directory contains the API contract for the OAuth Client Application Registration Portal.

## Files

| File           | Description                                                 |
| -------------- | ----------------------------------------------------------- |
| `openapi.yaml` | OpenAPI 3.0.3 specification for client management endpoints |

## Endpoints Summary

| Method   | Path                                             | Description                       |
| -------- | ------------------------------------------------ | --------------------------------- |
| `POST`   | `/api/v1/admin/clients`                          | Register new OAuth client app     |
| `GET`    | `/api/v1/admin/clients`                          | List clients (paginated + search) |
| `GET`    | `/api/v1/admin/clients/{clientId}`               | Get client details                |
| `PUT`    | `/api/v1/admin/clients/{clientId}`               | Update client config              |
| `DELETE` | `/api/v1/admin/clients/{clientId}`               | Delete client + revoke tokens     |
| `POST`   | `/api/v1/admin/clients/{clientId}/rotate-secret` | Rotate client secret              |

## Authentication

All endpoints require JWT Bearer token from tenant administrator login (`POST /api/v1/login`).

## Notes

- These endpoints already exist in the codebase from feature 003-sso-auth-provider
- This contract documents the **extended** behavior: client_type, description, pagination, quota enforcement, audit logging, and token revocation on deletion
- Routes do not change — only request/response payloads are extended
