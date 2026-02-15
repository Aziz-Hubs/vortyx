# **VortyxPulse – Remote Monitoring & Management (RMM)**

**Division:** VortyxMSP (Infrastructure)  
**Status:** [PLANNING]

## **1. Executive Summary**

VortyxPulse is the heartbeat of the Vortyx platform. It is a high-performance, low-latency Remote Monitoring and Management (RMM) platform designed to provide real-time visibility into the health and performance of servers, workstations, and network devices. Unlike traditional RMMs that rely on heavy Java or .NET agents, VortyxPulse utilizes a lightweight, native Go binary that runs silently in the background with minimal resource overhead.

## **2. Technical Architecture**

### **2.1 The Agent (Vort)**

- **Name:** **Vort**
- **Language:** Go (Golang) 1.24+
- **Architecture:** Unified Modular Binary. All Vortyx features (Pulse, Guard, Radar) are included in this single binary.
- **Deployment:** Native binary service (Systemd on Linux, Windows Service on Windows, LaunchDaemon on macOS).
- **Footprint:** <1% CPU usage, <20MB RAM idle (base module).
- **Communication:** Secure WebSocket (WSS) + gRPC to the VortyxCore central server.
- **Feature Toggling:** Modules are activated/deactivated in real-time via the central dashboard without requiring a restart or re-installation of the binary.

### **2.2 The Backend (The Heart)**

- **Ingestion:** TimescaleDB (PostgreSQL Extension) for high-velocity metrics (CPU, RAM, Network I/O).
- **Processing:** Real-time stream processing for alert triggering (e.g., "CPU > 90% for 5 mins").

### **2.3 The Frontend (The Monitor)**

- **Framework:** NextJS 16
- **Visualization:** Real-time graphs using `ShadCN` UI components for high-density time-series data.

## **3. Core Features**

### **3.1 Real-Time Telemetry**

- **Hardware Monitoring:** Live streaming of CPU temperature, fan speeds, voltage, disk S.M.A.R.T status, and RAM utilization.
- **Process Management:** Live task manager allowing the technician to kill runaway processes remotely without RDP.
- **Service Control:** Start, stop, and restart system services (Windows Services / Systemd units).
- **Antivirus Manager:** Real-time status reporting of Anti-virus software (e.g., Windows Defender) ensuring definitions are up-to-date and protection is active.

### **3.2 Automated Self-Healing (Automation Engine)**

- **Script Repository:** Centralized library for managing PowerShell (.ps1) and Bash (.sh) scripts.
- **Job Scheduler:** Allows technicians to schedule scripts to run once or on a recurring Cron schedule.
- **Execution History:** Full audit log of script runs, output (stdout/stderr), and exit codes.

### **3.3 Patch Management**

- **Windows Updates:** Native integration with Windows Update Agent (WUA) via COM interface.
- **Linux Updates:** Wrapper for `apt`, `dnf`, and `yum` package managers.
- **Patch Dashboard:** Unified view of missing updates (Critical, Security, Feature) with one-click install capability.

### **3.4 Remote Access**

- **Graphical Remote Desktop:** Powered by a self-hosted RustDesk infrastructure. The VortyxPulse agent automatically deploys and configures the RustDesk service on endpoints, peering them directly with the private Vortyx ID/Relay servers.
- **Terminal:** Fully interactive web-based terminal using `xterm.js` and a custom Go PTY backend.
- **File Browser:** Dedicated remote file system explorer for background file management.

### **3.5 Network Monitoring (SNMP)**

- **Device Discovery:** Polls and monitors network devices (Switches, Firewalls, Printers).
- **Metrics:** Tracks Bandwidth (In/Out), Interface Status, and Device Uptime via SNMPv2c/v3.

### **3.6 System Power Actions**

- **Reboot:** One-click device restart directly from the dashboard.
- **Shutdown:** Graceful OS shutdown for maintenance windows.

### **3.7 Event Log Viewer**

- **Windows:** Pulls Error/Warning logs from the System Event Log.
- **Linux:** Fetches logs from `journalctl` with priority filter.

### **3.8 Availability & Heartbeat Logic**

- **Pulse Mechanism:** Persistent WebSocket connection ("The Pulse") for real-time status. Fallback to HTTP/2 heartbeat beacon every 60 seconds.
- **State Logic:**
  - **Online (Green):** WebSocket is active; telemetry < 30s old.
  - **Unresponsive (Amber):** WebSocket dropped, but HTTP beacon received within 5 mins.
  - **Offline (Red):** No beacons received for > 3 minutes.
- **Wake-on-LAN Proxy:** Target offline devices via a neighboring online agent within the same subnet.

### **3.9 Asset Grouping & Tagging**

- **Hierarchical Structure:** Assets organized by Client > Site > Device.
- **Dynamic Tagging:** Flexible key-value tagging system (e.g., role:server, os:windows10).
- **Smart Groups:** Virtual device groups defined by boolean logic.

## **4. Data Strategy**

- **Metric Storage:** All numerical telemetry stored in **TimescaleDB** hypertables.
- **Retention:** Raw data kept for 7 days; 1-hour rollups kept for 1 year.

## **5. Integration with Vortyx Ecosystem**

- **VortyxPilot:** Automatically generates tickets when critical alerts (e.g., "Server Offline") are triggered.
- **VortyxNexus:** Updates asset data (RAM, CPU model, Serial Number) in the documentation wiki automatically.
- **VortyxGuard:** Deploys and manages the EDR module.
- **VortyxRadar:** Streams telemetry data to the SIEM for security analysis.
- **VortyxReflex:** Triggers isolation playbooks when malicious behavior is detected.

---

# **VortyxPulse – MVP Technical Design Document (TDD)**

**Version:** 1.0  
**Status:** [DRAFT]  
**Target Delivery:** Q2 2026

---

## **6. MVP Scope Definition**

The MVP focuses on the **core heartbeat loop**: real-time device visibility, basic remote actions, and automated alerting.

### **6.1 MVP Feature Matrix**

| Feature                     | MVP (Phase 1) | Phase 2 | Phase 3 |
| :-------------------------- | :-----------: | :-----: | :-----: |
| Agent Deployment (Go)       |      ✅       |         |         |
| Heartbeat & Online Status   |      ✅       |         |         |
| Real-Time Telemetry         |      ✅       |         |         |
| Alert Engine                |      ✅       |         |         |
| Web Dashboard               |      ✅       |         |         |
| Device Detail Panel         |      ✅       |         |         |
| Reboot / Shutdown Actions   |      ✅       |         |         |
| Live Terminal (PTY)         |      ✅       |         |         |
| Script Execution (Ad-hoc)   |      ✅       |         |         |
| Patch Management            |               |   ✅    |         |
| Service Control             |               |   ✅    |         |
| Process Manager             |               |   ✅    |         |
| SNMP Network Monitoring     |               |   ✅    |         |
| Background File Explorer    |               |   ✅    |         |
| RustDesk Remote Desktop     |               |         |   ✅    |
| Wake-on-LAN Proxy           |               |         |   ✅    |
| Smart Groups & Tagging      |               |         |   ✅    |

---

## **7. System Architecture (MVP)**

### **7.1 Component Diagram**

```
┌─────────────────────────────────────────────────────────────────────┐
│                          Vortyx Platform                            │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │                    Next.js Frontend (App Router)               │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │ │
│  │  │ Device List  │  │ Device Panel │  │ Terminal Component   │  │ │
│  │  └──────────────┘  └──────────────┘  └──────────────────────┘  │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                              │ ConnectRPC (HTTP/JSON)               │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │                      Go Backend (Echo)                         │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │ │
│  │  │ AgentService │  │ AlertService │  │ TerminalService      │  │ │
│  │  │ (ConnectRPC) │  │  (Internal)  │  │ (WebSocket Handler)  │  │ │
│  │  └──────────────┘  └──────────────┘  └──────────────────────┘  │ │
│  │                                                                │ │
│  │  ┌──────────────────────────────────────────────────────────┐  │ │
│  │  │              TimescaleDB (PostgreSQL + Hypertables)      │  │ │
│  │  └──────────────────────────────────────────────────────────┘  │ │
│  └────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
                              ▲
                              │ gRPC (Protobuf) / HTTPS Heartbeat
                              │
       ┌──────────────────────┴──────────────────────┐
       │                                              │
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│   Agent 1    │  │   Agent 2    │  │   Agent N    │
│  (Windows)   │  │   (Linux)    │  │   (macOS)    │
│   Go Binary  │  │   Go Binary  │  │   Go Binary  │
└──────────────┘  └──────────────┘  └──────────────┘
```

## **8. API Contracts (ConnectRPC / Protobuf)**

**File:** `proto/vortyxpulse/v1/pulse.proto`

```protobuf
syntax = "proto3";

package vortyxpulse.v1;

import "google/protobuf/timestamp.proto";

option go_package = "github.com/vortyx/vortyx-platform/pkg/proto/vortyxpulse/v1;vortyxpulsev1";

// ============ AGENT SERVICE (Agent -> Server) ============

service AgentService {
  rpc CheckIn(CheckInRequest) returns (CheckInResponse);
  rpc StreamTerminal(stream TerminalInput) returns (stream TerminalOutput);
}

// ... (Rest of proto definitions)
```

---

## **9. Implementation Milestones**

### **Phase 1A: Foundation (Weeks 1-2)**

- [ ] Define Protobuf schemas (`pulse.proto`)
- [ ] Generate Go and TypeScript clients via `buf generate`
- [ ] Create database migrations (SQLc)
- [ ] Implement `AgentService.CheckIn` RPC
- [ ] Build basic agent telemetry collector

### **Phase 1B: Agent Core (Weeks 3-4)**

- [ ] Implement agent heartbeat loop with gRPC
- [ ] Add HTTP fallback for unstable networks
- [ ] Build command executor (Reboot, Shutdown)
- [ ] Create OS-specific installers (MSI, DEB)

### **Phase 1C: Backend Services (Weeks 5-6)**

- [ ] Implement `DashboardService` RPCs
- [ ] Build Alert Engine background worker
- [ ] Integrate TimescaleDB for telemetry storage

### **Phase 1D: Frontend (Weeks 7-8)**

- [ ] Build Device List page
- [ ] Build Device Detail panel with charts
- [ ] Integrate Terminal component (xterm.js)

### **Phase 1E: Polish & Launch (Weeks 9-10)**

- [ ] End-to-end testing
- [ ] Security audit (mTLS, RBAC)
- [ ] MVP Launch 🚀
