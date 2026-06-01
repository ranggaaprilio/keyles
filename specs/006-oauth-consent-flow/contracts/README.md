# API Contracts: OAuth Consent Flow and End-User Authentication

**Feature**: [spec.md](../spec.md)  
**Date**: June 1, 2026

## Overview

This directory documents the browser-facing OAuth interaction API added around the
existing OAuth token flow.

## Files

| File | Description |
| --- | --- |
| `openapi.yaml` | OpenAPI 3.0.3 contract for authorization initialization, end-user login, consent details, consent decision, and provider-local logout |

## Endpoints Summary

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/oauth2/auth` | Validate OAuth request and redirect browser to login, consent, local error, or validated callback error |
| `POST` | `/oauth2/login` | Authenticate end-user and create HttpOnly SSO session |
| `GET` | `/oauth2/consent/{transactionId}` | Return display-safe consent details for the bound session |
| `POST` | `/oauth2/consent` | Approve or deny consent and return the callback redirect |
| `POST` | `/oauth2/logout` | Delete the provider session and expire its host-only cookie |

## Notes

- Frontend routes receive only opaque transaction identifiers.
- OAuth `state` is stored server-side and returned unchanged to the external client.
- Invalid client and callback URI failures never redirect to an untrusted URI.
- Existing `/oauth2/token`, `/oauth2/revoke`, `/oauth2/introspect`, discovery,
  JWKS, and userinfo contracts remain compatible.
- Browser-flow endpoints fail closed with local `temporarily_unavailable` errors
  when Redis security state cannot be read or written. Logout still expires the
  cookie locally.
- Authorization initialization redirects Redis failures to the local frontend error
  page. JSON interaction endpoints return a local `503` response instead.
- OAuth login throttling and audit events use the direct TCP peer address. Forwarded
  IP headers remain untrusted until a trusted-proxy allowlist is designed.
