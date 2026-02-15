# System Overview

## Architectural Philosophy

Vortyx adopts a **Unified Monolith** architecture. This strategic choice balances the simplicity of deployment and shared infrastructure with the modularity of microservices.

### Core Principles
- **Decoupled Logic**: Although deployed as a single binary, business domains (e.g., `pulse`, `radar`) are strictly isolated within `internal/` packages.
- **Protocol-First**: All service contracts are defined in **Protobuf** (`proto/vortyx/`), serving as the single source of truth for both Backend (Go) and Frontend (TypeScript).
- **Type Safety**: End-to-end type safety from database query (sqlc) to API response (ConnectRPC).

## High-Level Diagram

```mermaid
graph TD
    Client[Browser / Next.js] -->|ConnectRPC / HTTP2| LB[Load Balancer / Reverse Proxy]
    LB -->|Port 8081| Backend[Go Monolith]
    LB -->|Port 8080| Identity[Zitadel OIDC]
    
    subgraph "Vortyx Backend"
        Router[Chi Router]
        Auth[Auth Middleware]
        
        subgraph "MSP Services"
            Pulse[Pulse (RMM)]
            Pilot[Pilot (PSA)]
            Nexus[Nexus (Docs)]
        end
        
        subgraph "MSSP Services"
            Radar[Radar (SIEM)]
            Guard[Guard (EDR)]
            Shield[Shield (GRC)]
        end
        
        Router --> Auth
        Auth --> Pulse & Pilot & Nexus
        Auth --> Radar & Guard & Shield
    end
    
    Backend -->|pgx/v5| DB[(PostgreSQL / TimescaleDB)]
    Backend -->|OIDC Check| Identity
```

## Component Breakdown

### 1. Frontend Layer
- **Technology**: Next.js 16 (App Router), React, TypeScript.
- **Styling**: Tailwind CSS, ShadCN/UI.
- **Communication**: `connect-es` (ConnectRPC client) generated from Protobuf.
- **State**: React Query (TanStack Query) recommended for server state management.

### 2. Backend Layer
- **Technology**: Go 1.22+, ConnectRPC (Go).
- **Routing**: `go-chi` for HTTP routing and middleware.
- **Service Layer**: 15+ isolated services (Pulse, Radar, etc.) implementing generated interfaces.
- **Dependency Injection**: Services receive dependencies (like `pgxpool.Pool`) via constructor injection.

### 3. Data Layer
- **Database**: PostgreSQL 16 extended with TimescaleDB for time-series telemetry.
- **Access Pattern**: `sqlc` generates type-safe Go code from raw SQL queries (`schema.sql`, `query.sql`).
- **Migration**: Schema-first approach managed via SQL files in `backend/internal/<app>/db/schema/`.

### 4. Infrastructure & Security
- **Identity Provider**: Zitadel (Self-hosted) handling OIDC, MFA, and User Management.
- **Remote Access**: RustDesk (Self-hosted) for remote desktop capabilities.
- **Containerization**: Docker Compose for local development; Kubernetes-ready architecture.

## Data Flow
1.  **Request**: User interacts with Next.js frontend.
2.  **Transport**: Frontend client sends a strongly-typed `POST` request to `https://api.vortyx.io/<service>/<method>`.
3.  **Auth**: Backend middleware intercepts request, validates JWT Bearer token against Zitadel's public keys.
4.  **Routing**: Chi router directs request to the specific service handler (e.g., `pulse.GetStatus`).
5.  **Logic**: Service executes business logic, invoking the Data Access Layer (DAL).
6.  **Query**: `sqlc`-generated code executes optimized SQL against the TimescaleDB instance.
7.  **Response**: Data flows back up the stack, marshaled into Protobuf, and returned to the client.
