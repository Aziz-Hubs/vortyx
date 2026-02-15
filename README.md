# Vortyx

[![CI](https://github.com/${{ github.repository }}/actions/workflows/ci.yml/badge.svg)](https://github.com/${{ github.repository }}/actions/workflows/ci.yml)
[![Deploy Staging](https://github.com/${{ github.repository }}/actions/workflows/deploy-staging.yml/badge.svg)](https://github.com/${{ github.repository }}/actions/workflows/deploy-staging.yml)
[![Deploy Production](https://github.com/${{ github.repository }}/actions/workflows/deploy-production.yml/badge.svg)](https://github.com/${{ github.repository }}/actions/workflows/deploy-production.yml)
[![Security Scan](https://github.com/${{ github.repository }}/actions/workflows/security.yml/badge.svg)](https://github.com/${{ github.repository }}/actions/workflows/security.yml)
[![Dependency Updates](https://github.com/${{ github.repository }}/actions/workflows/dependency-updates.yml/badge.svg)](https://github.com/${{ github.repository }}/actions/workflows/dependency-updates.yml)

Vortyx is a Modern Fullstack Platform built on a decoupled architecture.

## Architecture

- **Frontend**: Next.js 16 (App Router)
- **Backend**: Go (Golang) 1.24+ (ConnectRPC)
- **Database**: PostgreSQL 16 / TimescaleDB
- **Identity**: Zitadel

## Prerequisites

- Go 1.24+
- Node.js 20+
- Docker & Docker Compose
- [Task](https://taskfile.dev) (Optional, for easy commands)
- [Buf](https://buf.build) (For Protobuf generation)

## Getting Started

1.  **Install dependencies**:
    ```bash
    # Backend
    cd backend
    go mod download

    # Frontend
    cd frontend
    npm install
    ```

2.  **Start Infrastructure**:
    ```bash
    task up
    # OR
    docker-compose up -d
    ```

3.  **Generate Code (if you changed proto files)**:
    ```bash
    task gen
    # OR
    buf generate
    ```

4.  **Run Development Environment**:
    ```bash
    task dev
    ```
    This will start:
    - Infrastructure (Docker)
    - Backend (Go) on port 8080
    - Frontend (Next.js) on port 3000

## Directory Structure

- `backend/`: Go backend service
- `frontend/`: Next.js frontend application
- `proto/`: Protobuf definitions
- `scripts/`: Helper scripts (e.g., DB init)
- `docker-compose.yml`: Infrastructure configuration

## Branching Strategy

This project follows the **Git Flow** methodology for branch management:

### Branch Types

| Branch | Purpose | Base | Merges To |
|--------|---------|------|-----------|
| `main` | Production-ready code | - | - |
| `dev` | Integration branch for development | `main` | `main` |
| `feature/*` | New features | `dev` | `dev` |
| `fix/*` | Bug fixes | `dev` | `dev` |
| `release/*` | Release preparation | `dev` | `main`, `dev` |
| `hotfix/*` | Critical production fixes | `main` | `main`, `dev` |

### Branch Rules

- **main**: Protected branch requiring pull request reviews and passing CI checks
- **dev**: Protected branch requiring pull request reviews and passing CI checks
- **feature/***: Created from `dev`, merged back to `dev` via PR
- **fix/***: Created from `dev`, merged back to `dev` via PR
- **release/***: Created from `dev`, merged to both `main` and `dev`
- **hotfix/***: Created from `main`, merged to both `main` and `dev`

### Creating Feature Branches

```bash
# Switch to dev branch
git checkout dev

# Update dev with latest changes
git pull origin dev

# Create feature branch
git checkout -b feature/your-feature-name

# Work on your feature
git add .
git commit -m "feat: add new feature"

# Push to remote
git push origin feature/your-feature-name

# Create Pull Request to dev branch
```

## CI/CD Pipeline

The project uses GitHub Actions for continuous integration and deployment:

### Workflows

| Workflow | Trigger | Description |
|----------|---------|-------------|
| **CI** | PR to `main`, `dev` | Linting, testing, type checking |
| **Deploy Staging** | Push to `main` | Deploys to staging environment |
| **Deploy Production** | Release published | Deploys to production with approval |
| **Security Scan** | Weekly + push to `main` | Vulnerability scanning |
| **Dependency Updates** | Weekly | Automated dependency updates |

### Pipeline Stages

1. **Lint & Test**: ESLint, TypeScript checks, Go linting, unit tests
2. **Build**: Docker image creation with caching
3. **Security Scan**: Trivy vulnerability scanning, CodeQL analysis
4. **Deploy**: Environment-specific deployment with health checks

### Environment Variables

Required secrets and variables for GitHub Actions:

**Repository Secrets:**
- `SNYK_TOKEN`: Snyk security scanning token
- `PRODUCTION_DATABASE_URL`: Production database connection
- `STAGING_DATABASE_URL`: Staging database connection
- `PRODUCTION_ZITADEL_URL`: Production Zitadel instance URL
- `STAGING_ZITADEL_URL`: Staging Zitadel instance URL
- `DOCKER_REGISTRY_TOKEN`: Container registry authentication

**Repository Variables:**
- `PRODUCTION_API_URL`: Production API endpoint
- `STAGING_API_URL`: Staging API endpoint

### Status Badges

![CI](https://github.com/${{ github.repository }}/actions/workflows/ci.yml/badge.svg)
![Deploy Staging](https://github.com/${{ github.repository }}/actions/workflows/deploy-staging.yml/badge.svg)
![Deploy Production](https://github.com/${{ github.repository }}/actions/workflows/deploy-production.yml/badge.svg)
![Security Scan](https://github.com/${{ github.repository }}/actions/workflows/security.yml/badge.svg)
![Dependency Updates](https://github.com/${{ github.repository }}/actions/workflows/dependency-updates.yml/badge.svg)

## Contributing

1. Create a feature branch from `dev`
2. Make your changes following the [commit guidelines](docs/development/coding_standards.md)
3. Run tests locally: `task gen && go test ./... && npm test`
4. Submit a pull request to `dev`
5. Wait for CI checks and code review
6. Squash and merge after approval

## License

MIT
