# Vortyx Architectural Knowledge Base

Welcome to the definitive architectural reference for **Vortyx**. This directory contains deep-dive documentation into every aspect of the system's design, from high-level module interactions to low-level configuration parameters.

## 🗺️ Architectural Map

| Document | Description |
| :--- | :--- |
| **[System Overview](./system_overview.md)** | High-level "Unified Monolith" design philosophy, service boundaries, and technology choices. |
| **[Module Reference](./modules.md)** | Detailed breakdown of the 15+ internal services (Pulse, Radar, etc.) and their responsibilities. |
| **[Data Flow](./data_flow.md)** | Sequence diagrams illustrating how data moves through the Frontend, API, and Database layers. |
| **[API Reference](./api_reference.md)** | Protocol Buffer definitions, RPC contracts, and request/response schemas. |
| **[Security Architecture](./security.md)** | Authentication (OIDC/Zitadel), Authorization (RBAC), and Data Protection standards. |
| **[Configuration](./configuration.md)** | Environment variables, validation rules, and deployment flags. |
| **[Performance](./performance.md)** | Optimization guidelines, benchmarking targets, and tuning parameters for Go/Postgres. |
| **[Troubleshooting](./troubleshooting.md)** | Common error codes, debugging workflows, and resolution steps. |

## 🔍 Directory Structure

```text
vortyx/
├── docs/
│   └── architecture/
│       ├── README.md           # You are here
│       ├── system_overview.md  # The "Big Picture"
│       ├── modules.md          # Component deep-dives
│       ├── data_flow.md        # Interaction diagrams
│       ├── api_reference.md    # Interface contracts
│       ├── security.md         # Auth & Encryption
│       ├── configuration.md    # Ops & Env Vars
│       ├── performance.md      # Speed & Scale
│       └── troubleshooting.md  # Fixes & FAQ
```

## 🔗 Quick Links

-   **[Backend Code](../../backend/)**: Go implementation of the architecture.
-   **[Frontend Code](../../frontend/)**: Next.js implementation of the client.
-   **[Proto Definitions](../../proto/)**: The source of truth for all APIs.
