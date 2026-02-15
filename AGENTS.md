# Trae Agent Context: Vortyx

This document serves as the primary context source for the Trae AI agent working on the Vortyx project. It outlines the architectural rules, coding standards, and operational workflows that **MUST** be followed.

## 1. Project Overview
**Vortyx** is a next-generation "Unified Monolith" platform for MSPs and MSSPs.
-   **Architecture**: Decoupled Monolith (Single Go Binary Backend + Next.js Frontend + Single Go Binary for the VORT agent deployed on clients machines for RMM services and more).
-   **Communication**: ConnectRPC (HTTP/2) using Protocol Buffers as the single source of truth.
-   **Database**: PostgreSQL 16 (TimescaleDB extension) accessed via `sqlc` (Type-safe Go generation).
-   **Identity**: Zitadel (OIDC/OAuth2).

### Core Service Portfolio
The platform consists of several integrated modules (Unified Monolith) and a distributed agent.

#### VORT – The Unified Machine Agent
A single-binary Go agent acting as the distributed "hands and eyes" of the platform.
-   **System Intelligence**: Real-time telemetry (CPU/RAM/Disk), event monitoring, configuration drift detection.
-   **Remote Operations**: Interactive terminal (PTY), stealth file management, script orchestration.
-   **Autonomous Security**: Active defense (process termination, network quarantine), ransomware canaries, log forwarding.
-   **Tech Stack**: Pure Go (Shared Codebase), WebSockets (Live Mode), mTLS.

#### Managed Services (MSP)
**Theme**: Visibility, Control, and Connection.
-   **VortyxPulse (RMM)**: Remote Monitoring & Management.
    -   *Tech*: Go-Sysinfo, TimescaleDB, Centrifugo, Pion (WebRTC).
-   **VortyxPilot (PSA)**: Professional Services Automation (Ticketing & Helpdesk).
    -   *Tech*: SLA Logic Engine, Automated Triage.
-   **VortyxNexus (CMDB)**: IT Documentation & Knowledge Base.
    -   *Tech*: Graph Analysis (Dependency Mapping), AES-GCM Vault.
-   **VortyxHorizon (vCIO)**: Strategic planning & Lifecycle management.
    -   *Tech*: Maroto (PDF Reports), Forecasting Algorithms.
-   **VortyxControl (SaaS)**: SaaS management & Shadow IT detection.
    -   *Tech*: Microsoft Graph SDK, Google Admin SDK.
-   **VortyxOptic (CCTV)**: NVR and AI-enhanced surveillance.
    -   *Tech*: MediaMTX, Pion, TFLite (Local Inference).
-   **VortyxGrid (Network)**: Network orchestration & Topology mapping.
    -   *Tech*: SSH Worker Pool, TextFSM (CLI Normalization).

#### Security Services (MSSP)
**Theme**: Defense, Detection, and Reflex.
-   **VortyxRadar (SIEM)**: Log aggregation and threat detection.
    -   *Tech*: Expr (Expression Engine), Sigma Rules.
-   **VortyxGuard (EDR)**: Endpoint detection and response.
    -   *Tech*: Gopacket (PCAP), Kernel Hooks (via VORT).
-   **VortyxShield (GRC)**: Governance, Risk, and Compliance.
    -   *Tech*: Automated Compliance Checks, Maroto (Reports).
-   **VortyxMind (Training)**: Phishing simulation & Awareness.
    -   *Tech*: Gophish Architecture (Go implementation).
-   **VortyxProbe (Scanner)**: Vulnerability scanning.
    -   *Tech*: Nuclei (Embedded Logic).
-   **VortyxReflex (SOAR)**: Automated security orchestration.
    -   *Tech*: Playbook Execution Engine.
-   **VortyxSonar (NDR)**: Network traffic analysis.
    -   *Tech*: Suricata Wrapper.
-   **VortyxSignal (CTI)**: Threat intelligence feed aggregation.

## 2. Tech Stack & Tools
-   **Backend**: Go 1.24+, Chi Router, `pgx/v5` (DB), `connect-go` (RPC).
-   **Frontend**: Next.js 16 (App Router), TypeScript, Tailwind CSS, ShadCN/UI, `connect-es` (RPC Client).
-   **Infrastructure**: Docker Compose, RustDesk (Remote Access).
-   **Tooling**:
    -   `task`: Task runner (replaces Makefile).
    -   `buf`: Protobuf generation and linting.
    -   `sqlc`: SQL to Go code generation.

## 3. Architecture Rules (CRITICAL)

### A. Protocol-First Development
1.  **Truth**: All APIs are defined in `proto/vortyx/<app>/v1/service.proto` first.
2.  **Generation**: Never manually write API clients or server stubs. Run `task gen` (or `task buf`) to generate them.
3.  **Versioning**: Always use semantic versioning in packages (e.g., `v1`).

### B. Database Access
1.  **No ORMs**: We do **NOT** use GORM or similar ORMs.
2.  **SQLC**: All database interactions are defined in raw SQL files:
    -   Schema: `backend/internal/<app>/db/schema/*.sql`
    -   Queries: `backend/internal/<app>/db/queries/*.sql`
3.  **Workflow**: Edit `.sql` files -> Run `task sqlc` -> Use generated Go code in `backend/internal/<app>/db/`.

### C. Backend Structure
1.  **Service Isolation**: Each domain (e.g., `pulse`, `radar`) lives in `backend/internal/<app>/`.
2.  **Dependency Injection**: Services **MUST** accept dependencies (like `*pgxpool.Pool`) via a `NewService` constructor. Global state is forbidden.
3.  **Router**: All routes are mounted in `backend/internal/server/server.go`.

### D. Frontend Structure
1.  **Client Generation**: RPC clients are generated into `frontend/src/gen/`.
2.  **Transport**: Use the singleton `transport` from `frontend/src/lib/transport.ts` for all API calls.
3.  **Components**: Use ShadCN/UI components from `frontend/src/components/ui/`.

## 4. Development Workflow

### Common Commands
-   `task up`: Start infrastructure (DB, Zitadel, etc.).
-   `task dev`: Start the full stack (Infra + Backend + Frontend).
-   `task gen`: Regenerate all code (Protobuf + SQL).
-   `task sqlc`: Regenerate only SQL code.
-   `task buf`: Regenerate only Protobuf code.

### Adding a New Feature
1.  **Define Data**: Add tables to `backend/internal/<app>/db/schema/`.
2.  **Define Queries**: Write SQL in `backend/internal/<app>/db/queries/`.
3.  **Define API**: Add RPCs to `proto/vortyx/<app>/v1/service.proto`.
4.  **Generate**: Run `task gen`.
5.  **Implement Backend**: Update `backend/internal/<app>/service.go`.
6.  **Implement Frontend**: Create/Update pages in `frontend/src/app/` using the generated client.

## 5. Documentation Maintenance Guidelines (MANDATORY)

To ensure documentation remains the single source of truth, **ALL** code changes must be accompanied by relevant documentation updates.

### A. Update Checklist (Must be verified before merging)
- [ ] **README.md**: Updated if project structure, quick start, or key tech stack changes.
- [ ] **API Docs**: `docs/api/` updated if `.proto` definitions changed (endpoints, request/response bodies).
- [ ] **User Guides**: `docs/guides/` updated if new user-facing features or workflows were added.
- [ ] **Env Config**: `docs/deployment/environment_variables.md` updated if `.env` or config flags changed.
- [ ] **Changelog**: `docs/CHANGELOG.md` updated with a summary of changes (Feature, Fix, Breaking Change).
- [ ] **Inline Comments**: Complex logic in Go/TS code must be commented explaining *why*, not just *what*.

### B. Update Criteria
| Change Type | Required Documentation Updates |
| :--- | :--- |
| **New Feature** | `CHANGELOG.md`, `docs/guides/`, Architecture diagrams if significant. |
| **Bug Fix** | `CHANGELOG.md`, Inline comments explaining the fix. |
| **API Change** | `proto/` comments, `docs/api/`, Frontend client usage examples. |
| **Config Change** | `docs/deployment/environment_variables.md`, `.env.example`. |
| **Refactor** | Architecture docs if boundaries change, inline comments. |

### C. Review Process
1.  **Self-Review**: Developer must perform a documentation pass before submitting a PR.
2.  **Code Review**: Reviewers **MUST** check documentation diffs. A PR with code changes but no doc updates (when required) should be blocked.
3.  **Approval**: Significant architectural changes require explicit sign-off from a Documentation Maintainer or Lead Architect.

### D. Templates

**Changelog Entry (`docs/CHANGELOG.md`)**
```markdown
## [1.2.0] - 2024-03-15
### Added
- Feature: Real-time CPU monitoring in Pulse dashboard.
### Fixed
- Bug: Fixed race condition in auth middleware.
### Changed
- Refactor: Moved database connection logic to `internal/db`.
```

**API Endpoint Description (in `.proto` files)**
```protobuf
// GetSystemStatus retrieves the current health and operational status of the agent.
// Required Scope: system:read
rpc GetSystemStatus(GetSystemStatusRequest) returns (GetSystemStatusResponse);
```

### E. Enforcement & Consequences
-   **Blocked Merges**: CI/CD pipelines will eventually enforce doc checks (e.g., ensuring `CHANGELOG.md` is touched).
-   **Debt Sprints**: Teams failing to maintain docs will be required to dedicate sprint time to documentation cleanup.
-   **Automation**: Use tools like `protoc-gen-doc` to auto-generate API docs from `.proto` comments to minimize manual drift.

## 6. Commit & Push Guidelines

Adherence to strict commit and push standards is required to maintain history quality and collaboration efficiency.

### A. Conventional Commits (MANDATORY)
We follow the **Conventional Commits** specification (`type(scope): description`).

| Type | Description | Example |
| :--- | :--- | :--- |
| `feat` | New feature | `feat(pulse): add cpu monitoring widget` |
| `fix` | Bug fix | `fix(auth): handle expired token error` |
| `docs` | Documentation only | `docs(readme): update setup instructions` |
| `style` | Formatting, missing semi-colons, etc. | `style(ui): run prettier on dashboard` |
| `refactor` | Code change that neither fixes a bug nor adds a feature | `refactor(db): move queries to sql files` |
| `test` | Adding or missing tests | `test(api): add integration tests for login` |
| `chore` | Build process, deps, aux tools | `chore(deps): upgrade next.js to v16` |

**Format:**
```text
type(scope): short description

[Optional] Body: detailed explanation of the change.
[Optional] Footer: references to issues (e.g., "Closes #123").
```

### B. Commit Rules
1.  **Atomic Commits**: Each commit should do **one** thing. Don't bundle a bug fix and a new feature in the same commit.
2.  **Frequency**: Commit often locally. Squash commits if necessary before pushing to shared branches.
3.  **Inclusions**:
    -   **Include**: Source code, tests, docs, config files.
    -   **Exclude**: Secrets (`.env`), build artifacts (`/dist`, `/bin`), IDE files (`.vscode`).
4.  **Verification**: Before committing, run `task gen` to ensure generated code is up-to-date.

### C. Branching Strategy
-   **`main`**: Production-ready code. Protected.
-   **`dev`**: Integration branch. Protected.
-   **Feature Branches**: `feat/feature-name` (branched from `dev`).
-   **Bugfix Branches**: `fix/bug-name` (branched from `dev`).
-   **Hotfix Branches**: `hotfix/issue-name` (branched from `main`).

### D. Push Verification Checklist
Before pushing your branch:
1.  **Build**: Run `task gen` and `go build ./...` (Backend) / `npm run build` (Frontend).
2.  **Test**: Run `go test ./...` and `npm run test`.
3.  **Lint**: Ensure no linter errors (`golangci-lint`, `eslint`).
4.  **Pull**: Rebase on the latest `dev` to resolve conflicts locally (`git pull --rebase origin dev`).

### E. Handling Merge Conflicts
1.  **Don't Panic**: Conflicts are normal.
2.  **Rebase**: Prefer `git rebase` over `git merge` to keep history linear.
3.  **Communication**: If a conflict is complex, talk to the author of the conflicting change.
4.  **Verification**: After resolving, **re-run tests** to ensure the resolution didn't break logic.

### F. Documentation Requirements
-   If your commit involves a **Feature** or **Breaking Change**, you **MUST** include updates to `docs/` or `CHANGELOG.md` in the **same PR**.
-   **Bad Practice**: "I'll update the docs in a later PR." (This never happens).
-   **Good Practice**: The PR includes `src/` changes AND `docs/` changes.

## 7. Directory Map

```text
vortyx/
├── backend/
│   ├── cmd/                    # Main entry point (unused, see main.go)
│   ├── gen/proto/go/           # Generated Go Proto code (DO NOT EDIT)
│   ├── internal/
│   │   ├── auth/               # Auth middleware & interceptors
│   │   ├── server/             # Router & service wiring
│   │   ├── <app>/              # Domain Services (Repeated for: pulse, radar, pilot, etc.)
│   │   │   ├── db/
│   │   │   │   ├── queries/    # Raw SQL Queries (*.sql)
│   │   │   │   ├── schema/     # Raw SQL Schema (*.sql)
│   │   │   │   ├── db.go       # Generated SQLC Interface
│   │   │   │   ├── models.go   # Generated SQLC Structs
│   │   │   │   └── *.sql.go    # Generated SQLC Methods
│   │   │   └── service.go      # Service Implementation
│   └── main.go                 # App initialization (Env, DB, Server)
├── frontend/
│   ├── public/                 # Static assets
│   ├── src/
│   │   ├── app/
│   │   │   ├── (dashboard)/    # Dashboard Layout Routes
│   │   │   │   ├── control/    # VortyxControl Page
│   │   │   │   ├── grid/       # VortyxGrid Page
│   │   │   │   ├── guard/      # VortyxGuard Page
│   │   │   │   ├── horizon/    # VortyxHorizon Page
│   │   │   │   ├── mind/       # VortyxMind Page
│   │   │   │   ├── nexus/      # VortyxNexus Page
│   │   │   │   ├── optic/      # VortyxOptic Page
│   │   │   │   ├── pilot/      # VortyxPilot Page
│   │   │   │   ├── probe/      # VortyxProbe Page
│   │   │   │   ├── pulse/      # VortyxPulse Page
│   │   │   │   ├── radar/      # VortyxRadar Page
│   │   │   │   ├── reflex/     # VortyxReflex Page
│   │   │   │   ├── shield/     # VortyxShield Page
│   │   │   │   ├── signal/     # VortyxSignal Page
│   │   │   │   ├── sonar/      # VortyxSonar Page
│   │   │   │   └── layout.tsx  # Dashboard Sidebar/Header
│   │   │   └── page.tsx        # Landing Page
│   │   ├── components/ui/      # ShadCN UI Components
│   │   ├── gen/proto/ts/       # Generated TS Proto code (DO NOT EDIT)
│   │   └── lib/                # Utilities (transport.ts)
├── proto/vortyx/               # Protocol Buffer Definitions
│   ├── <app>/v1/service.proto  # Service Contracts
├── scripts/                    # Helper scripts (init-db.sh)
├── docs/                       # Documentation
│   ├── api/                    # API & Auth Docs
│   ├── architecture/           # System Design Docs
│   ├── deployment/             # Infra & CI/CD Docs
│   └── development/            # Setup & Standards Docs
├── .env                        # Environment Variables (Ignored)
├── docker-compose.yml          # Infrastructure Definition
├── sqlc.yaml                   # SQLC Configuration
└── Taskfile.yml                # Task Runner Configuration
```

## 8. Environment & Configuration
-   **Secrets**: Stored in `.env` (not committed).
-   **Ports**:
    -   Frontend: `3000`
    -   Backend: `8081`
    -   Zitadel: `8080`
    -   Postgres: `5432`

## 9. Operational Constraints
-   **No Magic**: If you add a library, document it.
-   **Port Conflicts**: Backend runs on `8081` to avoid conflict with Zitadel on `8080`.
-   **Auth**: All API requests require a valid OIDC Bearer token (handled by `auth` middleware).

## 10. Intelligent Documentation Synchronization (AI AGENT PROTOCOL)

This section mandates the protocol for AI agents to ensure documentation (`docs/`) stays perfectly synchronized with the codebase (`backend/`, `frontend/`, `proto/`, `infra/`).

### A. Pre-Processing Routine
Before executing any task, the Agent MUST:
1.  **Scan Context**: Read `AGENTS.md` and `README.md` to understand the current state.
2.  **Identify Targets**: Determine which documentation files are relevant to the planned changes (e.g., if changing API, check `docs/api/`).
3.  **Baseline Check**: Verify the current documentation reflects the existing code before making changes.

### B. Execution & Tracking
During task execution, the Agent MUST:
1.  **Track Modifications**: Log every file created, modified, or deleted.
2.  **Categorize Changes**:
    -   *Structural*: Moving files, renaming directories.
    -   *Functional*: Changing logic, adding features.
    -   *Interface*: Changing APIs, CLI flags, Environment Variables.
3.  **Detect Impact**: Flag any change that invalidates existing documentation (e.g., "I changed the port in `docker-compose.yml`, so `docs/deployment/environment_variables.md` is now wrong").

### C. Post-Processing Verification
After completing code changes, the Agent MUST:
1.  **Re-examine `docs/`**: Scan the documentation folder.
2.  **Identify Discrepancies**: Compare the new code state against the documentation.
    -   *Did I add a new env var?* -> Check `environment_variables.md`.
    -   *Did I add a new service?* -> Check `architecture/system_overview.md`.
    -   *Did I change the build process?* -> Check `deployment/ci_cd_pipeline.md`.
3.  **Validate References**: Ensure all links (e.g., `[link](./file.md)`) and code references remain valid.

### D. Automatic Updates & Task Generation
If discrepancies are found:
1.  **Auto-Correction**: If the Agent has sufficient context, it MUST update the documentation immediately.
    -   *Rule*: Preserve existing valid content. Do not overwrite manual explanations with generic AI text unless necessary.
    -   *Rule*: Use consistent terminology defined in `AGENTS.md`.
2.  **Task Generation**: If the update is ambiguous or requires human input, the Agent MUST generate a specific **TODO** item.
    -   *Example*: "TODO: Update `docs/architecture/data_flow.md` to reflect the new caching layer."
3.  **Change Logging**: Update `docs/CHANGELOG.md` with the technical details of the changes.

### E. Error Handling & Notification
1.  **Critical Mismatch**: If code changes contradict fundamental architectural rules (Section 3), the Agent MUST stop and warn the user.
2.  **Notification**: When documentation is updated, explicitly mention it in the final response: "I have updated `docs/deployment/ci_cd_pipeline.md` to reflect the new CI workflow."
