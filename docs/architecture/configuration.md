# Configuration Reference

This document lists all environment variables required to run Vortyx, their default values, and validation rules.

## Core Services

### Backend (`backend/.env`)

| Variable | Description | Type | Default | Validation |
| :--- | :--- | :--- | :--- | :--- |
| `PORT` | HTTP/2 Server Port | `int` | `8081` | 1-65535 |
| `DATABASE_URL` | Postgres Connection String | `url` | (Required) | `postgres://...` |
| `ZITADEL_ISSUER` | OIDC Issuer URL | `url` | (Required) | HTTPS required in prod |
| `LOG_LEVEL` | Log Verbosity | `enum` | `info` | debug, info, warn, error |
| `ENVIRONMENT` | Deployment Environment | `enum` | `dev` | dev, staging, prod |

### Frontend (`frontend/.env.local`)

| Variable | Description | Type | Default | Validation |
| :--- | :--- | :--- | :--- | :--- |
| `NEXT_PUBLIC_API_URL` | Backend API Base URL | `url` | `http://localhost:8081` | No trailing slash |
| `NEXT_PUBLIC_ZITADEL_ISSUER` | Zitadel Issuer URL | `url` | (Required) | Matches Backend |
| `NEXT_PUBLIC_ZITADEL_CLIENT_ID` | OAuth2 Client ID | `string` | (Required) | - |
| `NEXTAUTH_SECRET` | NextAuth Encryption Key | `string` | (Required) | min 32 chars |
| `NEXTAUTH_URL` | Application Base URL | `url` | `http://localhost:3000` | - |

## Infrastructure (`docker-compose.yml`)

### Postgres

| Variable | Description | Default | Notes |
| :--- | :--- | :--- | :--- |
| `POSTGRES_USER` | Superuser | `postgres` | Change in prod |
| `POSTGRES_PASSWORD` | Password | (Required) | min 16 chars |
| `POSTGRES_DB` | Initial DB | `vortyx` | - |

### Zitadel

| Variable | Description | Default | Notes |
| :--- | :--- | :--- | :--- |
| `ZITADEL_MASTERKEY` | Master Key | (Required) | Must be exactly 32 bytes |
| `ZITADEL_EXTERNAL_DOMAIN` | Domain | `localhost` | `auth.vortyx.io` in prod |

## Feature Flags

| Flag | Description | Default |
| :--- | :--- | :--- |
| `ENABLE_RUSTDESK` | Enable Remote Desktop | `true` |
| `ENABLE_VIDEO_STREAM` | Enable CCTV Ingestion | `false` |
| `ENABLE_THREAT_INTEL` | Sync External IOCs | `true` |
