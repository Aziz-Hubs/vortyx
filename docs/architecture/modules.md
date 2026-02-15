# Module Reference

This document provides a detailed breakdown of the internal services that make up the **Vortyx** ecosystem. Each module is a self-contained domain within the unified monolith.

## Core Modules (MSP)

### 1. **Pulse (RMM)**
-   **Role**: Remote Monitoring & Management. The central nervous system of Vortyx.
-   **Responsibilities**:
    -   Collects telemetry (CPU, RAM, Disk) from endpoints.
    -   Executes remote scripts and patches.
    -   Manages RustDesk sessions.
-   **Interactions**: Uses `grid` for network context; feeds data to `radar` for security analysis.
-   **Dependencies**: `pgxpool`, `rustdesk-server`.

### 2. **Pilot (PSA)**
-   **Role**: Professional Services Automation.
-   **Responsibilities**:
    -   Ticketing system for support requests.
    -   SLA management and escalation workflows.
    -   Time tracking and billing integration.
-   **Interactions**: Linked to `pulse` for automated remediation tickets.

### 3. **Nexus (Docs)**
-   **Role**: IT Documentation & Knowledge Base.
-   **Responsibilities**:
    -   Stores structured documentation for assets and procedures.
    -   Manages encrypted credentials (password vault).
    -   Tracks dependencies between assets (Graph DB logic).

### 4. **Horizon (Strategy)**
-   **Role**: vCIO & Strategic Planning.
-   **Responsibilities**:
    -   Generates QBR (Quarterly Business Review) reports.
    -   Tracks asset lifecycle and budget forecasting.
    -   Calculates "Health Scores" for clients.

### 5. **Control (SaaS)**
-   **Role**: SaaS Management.
-   **Responsibilities**:
    -   Integrates with M365/Google Workspace APIs.
    -   Monitors license usage and unauthorized SaaS apps.
    -   Automates user onboarding/offboarding.

### 6. **Optic (CCTV)**
-   **Role**: Surveillance NVR.
-   **Responsibilities**:
    -   Ingests RTSP streams via MediaMTX.
    -   Performs object detection using TFLite.
    -   Stores video clips in object storage.

### 7. **Grid (Network)**
-   **Role**: Network Management.
-   **Responsibilities**:
    -   Scans subnets (ARP/SNMP) for topology mapping.
    -   Manages switch/router configurations (SSH/Telnet).
    -   Detects rogue devices.

---

## Security Modules (MSSP)

### 8. **Radar (SIEM)**
-   **Role**: Security Information & Event Management.
-   **Responsibilities**:
    -   Ingests logs (Syslog, Windows Event Log).
    -   Runs Sigma rules for threat detection.
    -   Provides forensic timeline search.

### 9. **Guard (EDR)**
-   **Role**: Endpoint Detection & Response.
-   **Responsibilities**:
    -   Monitors process trees and file system changes.
    -   Executes active defense (process kill, isolation).
    -   Manages "Canary" files for ransomware detection.

### 10. **Shield (GRC)**
-   **Role**: Governance, Risk, & Compliance.
-   **Responsibilities**:
    -   Automates compliance checks (ISO 27001, GDPR).
    -   Generates audit reports.
    -   Manages vendor risk assessments.

### 11. **Mind (Training)**
-   **Role**: Security Awareness.
-   **Responsibilities**:
    -   Runs phishing simulation campaigns.
    -   Delivers "Teachable Moment" training content.
    -   Tracks user risk scores.

### 12. **Probe (Scanner)**
-   **Role**: Vulnerability Scanner.
-   **Responsibilities**:
    -   Scans external IPs and internal networks (Nuclei).
    -   Identifies unpatched CVEs and misconfigurations.
    -   Reports findings to `pilot` as tickets.

### 13. **Reflex (SOAR)**
-   **Role**: Security Orchestration, Automation, & Response.
-   **Responsibilities**:
    -   Executes automated playbooks (e.g., "Isolate Host").
    -   Orchestrates actions across `radar` and `guard`.

### 14. **Sonar (NDR)**
-   **Role**: Network Detection & Response.
-   **Responsibilities**:
    -   Analyzes network traffic patterns (Flow data).
    -   Detects lateral movement and C2 beacons.

### 15. **Signal (Intel)**
-   **Role**: Threat Intelligence.
-   **Responsibilities**:
    -   Aggregates threat feeds (IPs, Hashes, Domains).
    -   Syncs IOCs to `radar` and `guard` for blocking.
