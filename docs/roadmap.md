# Vortyx Project Roadmap

This document outlines the strategic plan for developing the Vortyx platform from its current architectural foundation to a production-ready MSP/MSSP ecosystem.

## 📅 Roadmap Overview

| Phase | Focus | Duration (Est.) | Key Deliverables |
| :--- | :--- | :--- | :--- |
| **Phase 1** | **Foundation** | Weeks 1-4 | Auth, CI/CD, Base UI, Telemetry Pipeline |
| **Phase 2** | **MSP Core** | Weeks 5-12 | Pulse (RMM), Pilot (PSA), Nexus (Docs) |
| **Phase 3** | **MSSP Security** | Weeks 13-20 | Radar (SIEM), Guard (EDR), Shield (GRC) |
| **Phase 4** | **Advanced Ops** | Weeks 21-28 | Horizon (vCIO), Control (SaaS), Grid (Net) |
| **Phase 5** | **Launch Readiness** | Weeks 29-32 | Load Testing, Security Audit, Documentation |

---

## 🏗️ Phase 1: Foundation & Infrastructure
**Goal**: Establish a stable, secure, and automated development environment.

### 1.1 Authentication & Authorization
- [ ] **Zitadel Integration**: Complete OIDC flow with MFA enforcement.
- [ ] **RBAC**: Implement role-based middleware (`admin`, `tech`, `viewer`).
- [ ] **Audit Logs**: Track all login/logout and sensitive actions.

### 1.2 Base UI & UX
- [ ] **Dashboard Layout**: Responsive sidebar, header with user profile.
- [ ] **Component Library**: Complete ShadCN/UI implementation (Tables, Forms, Dialogs).
- [ ] **Theme Support**: Dark/Light mode toggle.

### 1.3 Telemetry Pipeline
- [ ] **TimescaleDB**: Optimize schema for high-volume ingestion.
- [ ] **Agent Proto**: Define `AgentTelemetry` message (CPU, RAM, Disk).
- [ ] **Ingest Service**: Create high-performance gRPC streaming endpoint in `pulse`.

### ✅ Success Criteria
-   User can log in/out securely.
-   Protected routes reject unauthenticated requests.
-   Backend can ingest 10k dummy telemetry points/sec.

---

## 🛠️ Phase 2: MSP Core Services
**Goal**: Enable basic IT management capabilities.

### 2.1 Pulse (RMM) - *Critical Path*
- [ ] **Agent Registration**: Handshake protocol for new endpoints.
- [ ] **Live Monitoring**: Real-time websocket feed of system stats.
- [ ] **Remote Shell**: Web-based terminal via RustDesk/WebSocket.
- [ ] **Script Runner**: Execute PowerShell/Bash scripts on agents.

### 2.2 Pilot (PSA)
- [ ] **Ticketing**: CRUD for tickets, comments, and status workflow.
- [ ] **SLA Engine**: Auto-calculate due dates based on priority.
- [ ] **Time Tracking**: Log work hours against tickets.

### 2.3 Nexus (Documentation)
- [ ] **Asset CMDB**: Store hardware/software inventory.
- [ ] **Password Vault**: Encrypted storage for credentials (AES-256).
- [ ] **Relationship Graph**: Link Assets <-> Users <-> Passwords.

### ✅ Success Criteria
-   Agent can register and send live data.
-   Technician can open a remote shell.
-   Tickets can be created and linked to assets.

---

## 🛡️ Phase 3: MSSP Security Services
**Goal**: Add threat detection and compliance features.

### 3.1 Radar (SIEM)
- [ ] **Log Ingestion**: Syslog/EventLog gRPC endpoints.
- [ ] **Sigma Engine**: Implement Go-based Sigma rule matcher.
- [ ] **Alerting**: Trigger Pilot tickets on high-severity matches.

### 3.2 Guard (EDR)
- [ ] **Process Monitor**: Track process start/stop events.
- [ ] **Isolation**: "Panic Button" to cut network access for an agent.
- [ ] **Canary Files**: Detect ransomware encryption attempts.

### 3.3 Shield (GRC)
- [ ] **Compliance Checks**: Automated checks for BitLocker, Firewall, etc.
- [ ] **Reporting**: Generate PDF compliance reports (ISO/GDPR).

### ✅ Success Criteria
-   Simulated attack (e.g., EICAR) triggers an alert in Radar.
-   "Isolate Host" command successfully blocks network traffic on agent.
-   Compliance report accurately reflects system state.

---

## 🚀 Phase 4: Advanced Operations
**Goal**: Enhance strategic value and network visibility.

### 4.1 Horizon (vCIO)
- [ ] **Health Score**: Algorithm to score client infrastructure.
- [ ] **Budgeting**: Hardware refresh forecasting.
- [ ] **QBR Generator**: Automated PDF report for client meetings.

### 4.2 Control (SaaS)
- [ ] **M365 Integration**: Sync users and licenses via Graph API.
- [ ] **Shadow IT**: Detect unauthorized SaaS usage via browser extensions/logs.

### 4.3 Grid (Network)
- [ ] **Discovery**: SNMP/ARP scanning for topology mapping.
- [ ] **Config Backup**: SSH fetch of switch/router configs.

### ✅ Success Criteria
-   M365 users are synced to Nexus.
-   Network topology map is generated automatically.
-   QBR report is downloadable.

---

## 🏁 Phase 5: Polish, QA & Launch
**Goal**: Ensure stability, security, and performance for production.

### 5.1 Quality Assurance
- [ ] **E2E Testing**: Cypress/Playwright flows for critical paths.
- [ ] **Load Testing**: k6 tests simulating 50k agents.
- [ ] **Security Audit**: Penetration testing and code review.

### 5.2 Documentation & Training
- [ ] **User Manuals**: Guides for Technicians and Admins.
- [ ] **API Docs**: Finalize public API reference.
- [ ] **Video Tutorials**: "How-to" series for common tasks.

### 5.3 Deployment
- [ ] **Infrastructure**: Terraform/Ansible scripts for prod environment.
- [ ] **CI/CD**: Fully automated pipeline to Staging/Prod.
- [ ] **Disaster Recovery**: Tested backup/restore procedures.

### ✅ Success Criteria
-   Zero critical bugs.
-   System handles 50k concurrent agents with <100ms latency.
-   All documentation is complete and reviewed.

---

## ⚠️ Risk Management

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| **Performance Bottleneck** | Slow dashboards, laggy remote control. | Use TimescaleDB compression, Redis caching, and WebSocket optimization. |
| **Agent Stability** | Crashes on endpoints disrupt operations. | Extensive testing on VMs, auto-restart services, rigorous Go panic handling. |
| **Security Breach** | Data leak or unauthorized access. | Regular audits, minimal agent privileges, strict mTLS/OIDC enforcement. |
| **Scope Creep** | Delayed delivery. | Strict adherence to MVP features; move "nice-to-haves" to post-launch backlog. |
