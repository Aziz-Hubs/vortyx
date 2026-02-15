# CI/CD Pipeline Template

This document serves as a template for documenting the Vortyx CI/CD workflows.

> **Note**: Specific CI tool (GitHub Actions, GitLab CI, Jenkins) is pending.

## Overview

The CI/CD pipeline automates the testing, building, and deployment of Vortyx services.

### Triggers
-   **Push to `main`**: Runs full test suite + deploys to Staging.
-   **Push to `release/*`**: Deploys to Production (Manual Approval).
-   **Pull Request**: Runs linting, unit tests, and security scans.

## Pipeline Stages

### 1. **Lint & Test** (Parallel)
-   **Lint**:
    -   `golangci-lint` (Backend)
    -   `npm run lint` (Frontend)
    -   `buf lint` (Protobuf)
-   **Test**:
    -   `go test ./...` (Unit Tests)
    -   `npm run test` (Jest/Vitest)
    -   `sqlc check` (SQL Validity)

### 2. **Build**
-   **Backend**: `go build` -> Docker Image (`vortyx-backend:sha`).
-   **Frontend**: `npm run build` -> Docker Image (`vortyx-frontend:sha`).
-   **Artifacts**: Push images to [Container Registry] (e.g., GHCR, ECR, Docker Hub).

### 3. **Deploy**
-   **Staging**: Update Kubernetes manifests / Compose file with new image tag.
    -   Run database migrations (`migrate up`).
    -   Wait for health checks.
-   **Production**: Same as Staging, but gated by manual approval.

## Rollback Strategy

-   **Database**: Automated down migrations (use with caution) or manual restore from snapshot.
-   **Application**: Revert Docker image tag to previous stable version (`vortyx-backend:prev`).

## Quality Gates

-   **Code Coverage**: > 80% (Backend), > 70% (Frontend).
-   **Security**: Snyk/Trivy scan for vulnerabilities (CVEs) in dependencies.
-   **Performance**: Lighthouse score > 90 (Frontend).
