# Infrastructure Template

This document serves as a template for documenting the Vortyx production infrastructure.

> **Note**: Hosting environment decisions (AWS/GCP/Azure/Bare Metal) are pending.

## Hosting Architecture

### Components
-   **Load Balancer**: [Provider] (e.g., AWS ALB, Nginx, Traefik)
    -   Terminates TLS.
    -   Routes `/api/*` to Backend.
    -   Routes `/` to Frontend (Next.js).
-   **Application Server**: [Instance Type] (e.g., EC2 t3.medium, Kubernetes Pod)
    -   Running Docker containers: `backend`, `frontend`.
-   **Database**: [Managed Service] (e.g., RDS, Cloud SQL, Self-Hosted TimescaleDB)
    -   Version: PostgreSQL 16 + TimescaleDB extension.
    -   Storage: [Size] GB SSD.
    -   Backup Policy: Daily snapshots retained for 30 days.
-   **Identity**: [Zitadel Instance] (e.g., Cloud or Self-Hosted)
    -   Domain: `auth.vortyx.io`.

## Scaling Strategy

-   **Horizontal**: The backend is stateless (session state in Zitadel/Redis). Autoscaling groups can spin up additional instances based on CPU/Memory usage.
-   **Database**: Vertical scaling initially; read replicas for heavy query loads.
-   **Frontend**: Cached at CDN edge (Vercel/Cloudflare) where possible.

## Monitoring & Logging

-   **Metrics**: Prometheus scraping `/metrics` endpoint (Go).
-   **Logs**: Structured JSON logs shipped to [Log Aggregator] (e.g., Datadog, ELK, Loki).
-   **Tracing**: OpenTelemetry instrumentation for distributed tracing.

## Security Controls

-   **Network**: VPC with private subnets for Database and Backend.
-   **Firewall**: Security Groups allowing ingress only on ports 443 (LB) and restricted SSH (Bastion).
-   **Secrets**: Managed via [Secret Manager] (e.g., AWS Secrets Manager, HashiCorp Vault).
