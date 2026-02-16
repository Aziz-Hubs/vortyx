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
| `ZITADEL_DOMAIN` | Zitadel gRPC/API address (host:port). | `localhost:8080` |
| `ZITADEL_ISSUER` | Zitadel OIDC issuer URL. | `http://localhost:8080` |
| `ZITADEL_AUDIENCES` | Comma-separated accepted `aud` values for JWT access tokens. | `3600801544557033557,360224206048264197` |
| `ZITADEL_API_PROJECT_ID` | Zitadel Project Resource ID for Vortyx Platform. | `3600801544557033557` |
| `ZITADEL_INSECURE` | Use insecure connections to Zitadel (local dev only). | `true` |
| `ZITADEL_INSECURE_PORT` | Insecure port when `ZITADEL_INSECURE=true`. | `8080` |
| `ZITADEL_PAT` | Personal Access Token for Management API (optional). | `(secret)` |
| `ZITADEL_SERVICE_USER_KEY_PATH` | Service-user key file path for Management API (optional). | `./json secrets/<file>.json` |
| `LOG_LEVEL` | Logging verbosity. | `debug` |

### VORT Agent Authentication
| Variable | Description | Default / Example |
| :--- | :--- | :--- |
| `VORT_MACHINE_USER_KEY_PATH` | Path to RSA private key for JWT Profile Grant (optional). | `./json secrets/vort-agent-secret.json` |
| `VORT_MACHINE_USER_KEY` | Base64-encoded RSA private key (alternative to KEY_PATH). | `(base64 string)` |
| `VORT_MACHINE_USER_KEY_ID` | Key ID for JWT header (required with machine user auth). | `360225578390913029` |
| `VORT_AGENT_JWT_PRIVATE_KEY` | Base64-encoded RSA private key for internal token signing. | Auto-generated if not provided |
| `VORT_AGENT_JWT_ISSUER` | JWT issuer for internal tokens. | `vortyx-agent-auth` |
| `VORT_AGENT_JWT_AUDIENCE` | JWT audience for internal tokens. | `vortyx-api` |
| `ENV` | Environment mode (affects pprof availability). | `development` (use `production` to disable pprof) |

### Frontend Application
| Variable | Description | Default / Example |
| :--- | :--- | :--- |
| `NEXT_PUBLIC_API_URL` | Base URL for the Backend API (browser). | `http://localhost:8081` |
| `ZITADEL_ISSUER` | Zitadel OIDC issuer URL (NextAuth server-side). | `http://localhost:8080` |
| `ZITADEL_CLIENT_ID` | Zitadel OIDC client ID for `vortyx-frontend`. | `360224206048264197` |
| `ZITADEL_CLIENT_SECRET` | Zitadel OIDC client secret for `vortyx-frontend`. | `(secret)` |
| `ZITADEL_API_PROJECT_ID` | Adds API audience to access tokens via reserved scope. | `3600801544557033557` |
| `ZITADEL_ENABLE_PASSWORD_GRANT` | Enables legacy password grant login (not recommended). | `false` |
| `NEXTAUTH_URL` | Public base URL of the frontend. | `http://localhost:3000` |
| `NEXTAUTH_SECRET` | NextAuth secret. | `(secret)` |

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
