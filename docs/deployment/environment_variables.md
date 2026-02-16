# Environment Variables

This document lists all environment variables required to run the Vortyx platform across different environments (Development, Staging, Production).

> **WARNING**: Never commit actual secrets to version control. Use `.env` for local development and GitHub Secrets for CI/CD.

## 1. Local Development

### Single Source of Truth: Root `.env`

Vortyx uses a **single `.env` file at the workspace root** that contains all configuration for both backend and frontend. The frontend automatically syncs relevant variables to `frontend/.env.local` when running `npm run dev` or `npm run build`.

```
VORTYX_ROOT/
├── .env                    # ← Single source of truth (contains ALL secrets)
├── .env.example            # ← Template for team members (safe to commit)
├── frontend/
│   └── .env.local          # ← Auto-generated (do not edit manually)
└── backend/
```

#### Setup for New Developers

1. Copy `.env.example` to `.env`
2. Fill in the values (see sections below)
3. Run `npm run dev` (frontend) or `task dev` (full stack)

### Backend Service
| Variable | Description | Default / Example |
| :--- | :--- | :--- |
| `PORT` | The port the backend server listens on. | `8081` |
| `DATABASE_URL` | PostgreSQL connection string. | `postgres://postgres:password@localhost:5432/vortyx?sslmode=disable` |
| `ZITADEL_DOMAIN` | Zitadel gRPC/API address (host:port). | `localhost:8080` |
| `ZITADEL_ISSUER` | Zitadel OIDC issuer URL. | `http://localhost:8080` |
| `ZITADEL_PROJECT_ID` | Zitadel Project Resource ID for Vortyx Platform. | `360080154455703557` |
| `ZITADEL_INSECURE` | Use insecure connections to Zitadel (local dev only). | `true` |
| `ZITADEL_INSECURE_PORT` | Insecure port when `ZITADEL_INSECURE=true`. | `8080` |
| `ZITADEL_MANAGEMENT_PAT` | Personal Access Token for Management API (IAM OWNER). | `(secret)` |
| `ZITADEL_BACKEND_API_KEY` | Path to backend API app key file (JWT Bearer Grant). | `./json secrets/vort-api-key.json` |
| `ZITADEL_BACKEND_API_ID` | Key ID for backend API app. | `360299659698110469` |
| `LOG_LEVEL` | Logging verbosity. | `debug` |

### VORT Agent Authentication
| Variable | Description | Default / Example |
| :--- | :--- | :--- |
| `ZITADEL_VORT_SERVICE_USER_KEY_PATH` | Path to service user key for JWT Profile Grant (optional). | `./json secrets/vort-agent-key.json` |
| `ZITADEL_VORT_SERVICE_USER_KEY` | Base64-encoded RSA private key (alternative to KEY_PATH). | `(base64 string)` |
| `ZITADEL_VORT_SERVICE_USER_KEY_ID` | Key ID for JWT header (required with service user auth). | `360225578390913029` |
| `ZITADEL_VORT_AGENT_JWT_PRIVATE_KEY` | Base64-encoded RSA private key for internal token signing. | Auto-generated if not provided |
| `ZITADEL_VORT_AGENT_JWT_ISSUER` | JWT issuer for internal tokens. | `vortyx-agent-auth` |
| `ZITADEL_AGENT_JWT_AUDIENCE` | JWT audience for internal tokens. | `vortyx-api` |
| `ENV` | Environment mode (affects pprof availability). | `development` (use `production` to disable pprof) |

### Frontend Application

> **Note**: The frontend variables below are automatically synced from the root `.env` file via `frontend/scripts/sync-env.js`. Do not edit `frontend/.env.local` manually - it is regenerated on every `npm run dev` or `npm run build`.

| Variable | Description | Default / Example |
| :--- | :--- | :--- |
| `NEXT_PUBLIC_API_URL` | Base URL for the Backend API (browser). | `http://localhost:8081` |
| `ZITADEL_ISSUER` | Zitadel OIDC issuer URL (NextAuth server-side). | `http://localhost:8080` |
| `ZITADEL_CLIENT_ID` | Zitadel OIDC client ID for `vortyx-frontend`. | `360224206048264197` |
| `ZITADEL_CLIENT_SECRET` | Zitadel OIDC client secret for `vortyx-frontend`. | `(secret)` |
| `ZITADEL_PROJECT_ID` | Project ID for token audience. | `360080154455703557` |
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
