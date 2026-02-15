# Environment Variables

This document lists all environment variables required to run the Vortyx platform.

> **WARNING**: Never commit actual secrets to version control. Use `.env` for local development and a secrets manager for production.

## Backend Configuration

| Variable | Description | Type | Default | Example |
| :--- | :--- | :--- | :--- | :--- |
| `PORT` | The port the backend server listens on. | `int` | `8081` | `8081` |
| `DATABASE_URL` | Connection string for PostgreSQL/TimescaleDB. | `string` | (Required) | `postgres://user:pass@host:5432/db?sslmode=disable` |
| `ZITADEL_ISSUER` | URL of the Zitadel instance for OIDC discovery. | `string` | (Required) | `https://auth.vortyx.io` |
| `LOG_LEVEL` | Logging verbosity (debug, info, warn, error). | `string` | `info` | `debug` |

## Frontend Configuration

| Variable | Description | Type | Default | Example |
| :--- | :--- | :--- | :--- | :--- |
| `NEXT_PUBLIC_API_URL` | Base URL for the Backend API (ConnectRPC). | `string` | `http://localhost:8081` | `https://api.vortyx.io` |
| `NEXT_PUBLIC_ZITADEL_CLIENT_ID` | OAuth2 Client ID for Frontend login. | `string` | (Required) | `23487234@vortyx` |
| `NEXT_PUBLIC_ZITADEL_ISSUER` | URL of the Zitadel instance. | `string` | (Required) | `https://auth.vortyx.io` |

## Infrastructure (Docker Compose)

| Variable | Description | Type | Default | Example |
| :--- | :--- | :--- | :--- | :--- |
| `POSTGRES_USER` | Database superuser name. | `string` | `postgres` | `postgres` |
| `POSTGRES_PASSWORD` | Database superuser password. | `string` | (Required) | `secret123` |
| `POSTGRES_DB` | Initial database name. | `string` | `vortyx` | `vortyx` |
| `ZITADEL_MASTERKEY` | Master encryption key for Zitadel (32 bytes). | `string` | (Required) | `MasterkeyNeedsToBe32BytesLong123!` |

## Secrets Management

In production, these values should be injected via:
-   **Kubernetes Secrets**
-   **AWS Secrets Manager**
-   **HashiCorp Vault**
-   **GitHub Secrets** (for CI/CD)
