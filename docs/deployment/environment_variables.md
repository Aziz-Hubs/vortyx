# Environment Variables

This document lists all environment variables required to run the Vortyx platform across different environments (Development, Staging, Production).

> **WARNING**: Never commit actual secrets to version control. Use `.env` for local development and GitHub Secrets for CI/CD.

## 1. Local Development (`.env`)

These variables are used when running `task dev` or `docker-compose up`.

### Backend Service
| Variable | Description | Default / Example |
| :--- | :--- | :--- |
| `PORT` | The port the backend server listens on. | `8081` |
| `DATABASE_URL` | PostgreSQL connection string. | `postgres://postgres:password@localhost:5432/vortyx?sslmode=disable` |
| `ZITADEL_ISSUER` | URL of the Zitadel instance. | `http://localhost:8080` |
| `LOG_LEVEL` | Logging verbosity. | `debug` |

### Frontend Application
| Variable | Description | Default / Example |
| :--- | :--- | :--- |
| `NEXT_PUBLIC_API_URL` | Base URL for the Backend API. | `http://localhost:8081` |
| `NEXT_PUBLIC_ZITADEL_CLIENT_ID` | OAuth2 Client ID for Frontend. | `23487234@vortyx` |
| `NEXT_PUBLIC_ZITADEL_ISSUER` | URL of the Zitadel instance. | `http://localhost:8080` |

### Infrastructure (Docker Compose)
| Variable | Description | Default |
| :--- | :--- | :--- |
| `POSTGRES_USER` | Database superuser name. | `postgres` |
| `POSTGRES_PASSWORD` | Database superuser password. | `password` |
| `POSTGRES_DB` | Initial database name. | `vortyx` |
| `ZITADEL_MASTERKEY` | Master encryption key for Zitadel (32 chars). | `MasterkeyNeedsToBe32BytesLong123!` |

## 2. CI/CD Pipeline Configuration

These variables are managed in GitHub Repository Settings (`Settings > Secrets and variables > Actions`).

### Secrets (Encrypted)
| Secret Name | Description | Required Scope |
| :--- | :--- | :--- |
| `STAGING_DATABASE_URL` | Connection string for Staging DB. | Staging |
| `PRODUCTION_DATABASE_URL` | Connection string for Production DB. | Production |
| `STAGING_ZITADEL_URL` | URL for Staging Identity Provider. | Staging |
| `PRODUCTION_ZITADEL_URL` | URL for Production Identity Provider. | Production |
| `SNYK_TOKEN` | Token for Snyk security scanning (Optional). | All |

### Variables (Plain Text)
| Variable Name | Description | Required Scope |
| :--- | :--- | :--- |
| `STAGING_API_URL` | Public API URL for Staging environment. | Staging |
| `PRODUCTION_API_URL` | Public API URL for Production environment. | Production |

## 3. Secrets Management Best Practices

-   **Local Development**: Store secrets in `.env` file (added to `.gitignore`).
-   **CI/CD**: Use GitHub Secrets.
-   **Production Runtime**: Secrets should be injected into containers at runtime via the orchestration platform (e.g., Kubernetes Secrets, AWS Secrets Manager, or environment variables in the deployment manifest).
