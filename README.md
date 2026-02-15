# Vortyx

[![CI](https://github.com/Aziz-Hubs/vortyx/actions/workflows/ci.yml/badge.svg)](https://github.com/Aziz-Hubs/vortyx/actions/workflows/ci.yml)
[![Deploy Staging](https://github.com/Aziz-Hubs/vortyx/actions/workflows/deploy-staging.yml/badge.svg)](https://github.com/Aziz-Hubs/vortyx/actions/workflows/deploy-staging.yml)
[![Deploy Production](https://github.com/Aziz-Hubs/vortyx/actions/workflows/deploy-production.yml/badge.svg)](https://github.com/Aziz-Hubs/vortyx/actions/workflows/deploy-production.yml)
[![Security Scan](https://github.com/Aziz-Hubs/vortyx/actions/workflows/security.yml/badge.svg)](https://github.com/Aziz-Hubs/vortyx/actions/workflows/security.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Vortyx** is a next-generation **"Unified Monolith"** platform designed specifically for Managed Service Providers (MSPs) and Managed Security Service Providers (MSSPs). It combines high-performance remote monitoring, comprehensive security operations, and business automation into a single, cohesive ecosystem.

## 🚀 Project Overview

Vortyx adopts a **Decoupled Monolith** architecture:
-   **Backend**: A single, high-performance Go binary serving all modules via ConnectRPC (HTTP/2).
-   **Frontend**: A modern Next.js 16 application providing a unified dashboard.
-   **Agent (VORT)**: A lightweight, single-binary Go agent deployed on endpoints for telemetry and execution.
-   **Data Layer**: PostgreSQL 16 with TimescaleDB for time-series telemetry, accessed via type-safe `sqlc`.
-   **Identity**: Zitadel for robust OIDC/OAuth2 authentication and authorization.

## ✨ Key Features

The platform is divided into three core pillars:

### 1. VORT – The Unified Machine Agent
The distributed "hands and eyes" of the platform.
-   **Real-Time Telemetry**: Sub-second streaming of CPU, RAM, Disk, and Network metrics.
-   **Remote Operations**: Interactive web-based terminal (PTY), stealth file browser, and script orchestration.
-   **Active Defense**: Autonomous process termination, network quarantine, and ransomware canary monitoring.

### 2. Managed Services (MSP Suite)
*Theme: Visibility, Control, and Connection.*
-   **VortyxPulse (RMM)**: Real-time health monitoring, patch management, and remote control (RustDesk integration).
-   **VortyxPilot (PSA)**: Intelligent ticketing, helpdesk automation, and SLA management.
-   **VortyxNexus (CMDB)**: IT documentation, asset tracking, and dependency graph analysis.
-   **VortyxHorizon (vCIO)**: Strategic IT planning, lifecycle management, and automated QBRs.
-   **VortyxControl (SaaS)**: Cloud license optimization and shadow IT detection (M365/Google Workspace).
-   **VortyxOptic (CCTV)**: AI-enhanced surveillance with local object detection.
-   **VortyxGrid (Network)**: Network topology mapping and infrastructure orchestration.

### 3. Security Services (MSSP Suite)
*Theme: Defense, Detection, and Reflex.*
-   **VortyxRadar (SIEM)**: Centralized log aggregation with Sigma rule-based threat detection.
-   **VortyxGuard (EDR)**: Endpoint detection and response with kernel-level hooking.
-   **VortyxShield (GRC)**: Automated governance, risk, and compliance reporting (ISO 27001, GDPR).
-   **VortyxMind (Training)**: Phishing simulation and security awareness training.
-   **VortyxProbe (Scanner)**: Internal and external vulnerability scanning.
-   **VortyxReflex (SOAR)**: Automated incident response playbooks.

## 🛠️ Tech Stack

-   **Language**: Go 1.24+ (Backend/Agent), TypeScript (Frontend).
-   **Frameworks**: Chi Router, ConnectRPC, Next.js 16, Tailwind CSS, ShadCN/UI.
-   **Database**: PostgreSQL 16 + TimescaleDB.
-   **Infrastructure**: Docker Compose, GitHub Actions.
-   **Tooling**: `task` (Build), `buf` (Protobuf), `sqlc` (Database).

## 📦 Getting Started

### Prerequisites
-   [Go 1.24+](https://go.dev/dl/)
-   [Node.js 20+](https://nodejs.org/)
-   [Docker & Docker Compose](https://www.docker.com/)
-   [Task](https://taskfile.dev) (Recommended for running commands)
-   [Buf](https://buf.build) (For Protobuf generation)

### Installation

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/Aziz-Hubs/vortyx.git
    cd vortyx
    ```

2.  **Install Backend Dependencies**:
    ```bash
    cd backend
    go mod download
    ```

3.  **Install Frontend Dependencies**:
    ```bash
    cd frontend
    npm install
    ```

4.  **Start Infrastructure**:
    ```bash
    task up
    # OR
    docker-compose up -d
    ```

### Running Development Environment

Start the full stack (Database, Backend, Frontend):
```bash
task dev
```
-   **Frontend**: [http://localhost:3000](http://localhost:3000)
-   **Backend API**: [http://localhost:8081](http://localhost:8081)
-   **Zitadel Auth**: [http://localhost:8080](http://localhost:8080)

## 📂 Directory Structure

```text
vortyx/
├── backend/            # Go backend service (Unified Monolith)
│   ├── internal/       # Domain logic (pulse, radar, etc.)
│   └── cmd/            # Entry points
├── frontend/           # Next.js 16 application
├── proto/              # Protocol Buffer definitions (Single Source of Truth)
├── docs/               # Comprehensive documentation
└── scripts/            # Setup and maintenance scripts
```

## 🤝 Contributing

We welcome contributions! Please follow these steps:

1.  **Fork the repository**.
2.  Create a feature branch: `git checkout -b feat/amazing-feature`.
3.  Commit your changes following [Conventional Commits](https://www.conventionalcommits.org/): `feat(pulse): add new widget`.
4.  **Verify your changes**: Run `task gen` and ensure tests pass.
5.  Push to the branch: `git push origin feat/amazing-feature`.
6.  Open a **Pull Request**.

For detailed guidelines, please refer to [CONTRIBUTING.md](docs/CONTRIBUTING.md) (if available) or the `docs/` directory.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Support

For support, questions, or feedback:
-   **Issues**: [GitHub Issues](https://github.com/Aziz-Hubs/vortyx/issues)
-   **Email**: support@vortyx.io (Placeholder)

---
*Built with ❤️ by the Vortyx Team*
