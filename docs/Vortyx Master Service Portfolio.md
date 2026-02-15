# **Vortyx \- Master Service Portfolio**

## **0\. VORT – The Unified Machine Agent**
  ### **Description:** 
  **Vort** is a high-performance, single-binary agent written in **Go (Golang)** that serves as the distributed "hands and eyes" of the Vortyx platform. Deployed as a native system service, it acts as the local executor for all platform requests—ranging from routine health monitoring to high-stakes security containment.

  ### **Key Capabilities**
  **System Intelligence & Telemetry**
  *   **Real-Time Streaming:** Continuous telemetry of CPU, RAM, Disk, and Network I/O metrics to the TimescaleDB backend.
  *   **System Event Monitoring:** Real-time tracking of OS-level events including user logins/logouts, system shutdowns, and reboots.
  *   **Configuration Drift Detection:** Monitoring of the local registry and system configuration files to detect unauthorized changes.

  **Remote Operations**
  *   **Interactive Terminal:** A web-based PTY (Pseudo-Terminal) supporting `cmd.exe`, `PowerShell`, and `bash` via secure WebSocket.
  *   **Stealth File Management:** Background file explorer for non-intrusive CRUD operations, chunked uploads, and log retrieval.
  *   **Script Orchestration:** Native execution engine for PowerShell, Bash, and Python scripts with full stdout/stderr capture and exit code reporting.

  **Autonomous Security**
  *   **Active Defense:** Real-time process termination and automated network quarantine (cutting all traffic except management tunnels).
  *   **Ransomware Canaries:** Placement and monitoring of hidden "canary" files to trigger instant process freezes upon encryption attempts.
  *   **Log Forwarding:** Secure collection, compression, and encrypted shipping of Windows Event Logs and Linux Syslogs.

  ### **Ecosystem Integrations**

  This section describes the specific role the **Vort Agent** plays for each module within the Vortyx ecosystem.

  | Module | Agent's Specific Role |
  | :--- | :--- |
  | **VortyxPulse (RMM)** | Acts as the primary data source for the RMM dashboard. It manages **Patch Management** (interfacing with WUA/Apt), executes maintenance scripts, and provides the pipe for **RustDesk** graphical remote sessions and terminal access. |
  | **VortyxRadar (SIEM)** | Serves as the **Lightweight Log Shipper**. It filters, compresses, and forwards local security logs to the SIEM ingestion engine for real-time Sigma rule matching. |
  | **VortyxGuard (EDR)** | Provides **Kernel-Level Hooking**. It monitors the process tree for malicious behavior (e.g., suspicious child-process spawning) and executes the **Active Defense** protocols (isolation/termination). |
  | **VortyxNexus (CMDB)** | Acts as the **Inventory Discovery Engine**. It performs deep hardware probes (Serial Numbers, BIOS, RAM slots) and software inventory audits, feeding the data into the CMDB to maintain an authoritative asset record. |
  | **VortyxReflex (SOAR)** | Serves as the **Response Arm**. When an automation playbook is triggered, the agent executes the containment actions—such as blocking a MAC address, resetting a local credential, or killing a specific network thread. |
  | **VortyxProbe (Scanner)** | Operates as an **Authenticated Internal Scanner**. It runs local, template-based vulnerability checks (via embedded Nuclei logic) to find unpatched CVEs or weak security configurations that are invisible to external scanners. |
  | **VortyxShield (GRC)** | Functions as the **Evidence Collector**. It automatically verifies and reports on compliance-critical settings, such as BitLocker/FileVault encryption status, firewall state, and MFA enrollment presence. |
  | **VortyxControl (SaaS)** | Provides **Application Correlation**. It scans the local process list to identify desktop applications used to access SaaS platforms, helping to bridge the gap between local software usage and cloud account management. |
  | **VortyxMind (Training)** | Acts as the **Desktop Interaction Layer**. It can trigger educational pop-ups or "Teachable Moment" notifications directly on the user's screen when a high-risk action is detected or blocked. |
  | **VortyxGrid (Network)** | Serves as a **Subnet Proxy**. It broadcasts Wake-on-LAN packets for neighboring devices and can act as a local cache for large software deployments, reducing WAN bandwidth usage for the entire site. |
  | **Partner Modules** | Managed **Orchestration**. Vort manages the silent installation, health monitoring, and policy enforcement of third-party tools like **Bitdefender** (Antivirus) and **Acronis** (Backup). |

  ### **4. Self-Management & Reliability**
  *   **Unified Binary:** One binary for all features; toggled via dashboard.
  *   **Auto-Update:** Self-versioning engine allows for silent, 1-click agent updates across the entire fleet.
  *   **Heartbeat Resilience:** Persistent WebSocket connection for "Live Mode," falling back to an HTTP/2 beacon if the socket is severed.
  *   **Security:** Mutual TLS (mTLS) authentication for all server communication and AES-256-GCM encryption for all sensitive fields at rest.

## **1\. Managed Services (MSP)**

**Theme:** Visibility, Control, and Connection.

### **VORTpulse (Remote Monitoring & Management)**

- **Core Function:** A comprehensive RMM platform for real-time health monitoring and management of servers and workstations.
- **Key Features:**
  - **Unified Agent (Vort):** Lightweight `vort` binary (Go-Sysinfo) that acts as the platform's endpoint engine.
  - **Remote Support Toolkit:** Built-in **Graphical RDP**, Interactive Terminal, and Remote File Browser (powered by Vort).
  - **Auto-Healing & Control:** Remote Service Manager, Process Killer, and System Power Actions (Reboot/Shutdown).
  - **Patch Management:** Integrated Windows Update (WUA) and Linux Package Manager (APT/DNF) control.
  - **TimescaleDB Ingest:** High-performance telemetry ingestion using `pgx` CopyFrom for 100k+ events/sec.
  - **Hybrid Caching:** Intelligent SQLite/Memory cache for offline telemetry buffering.
- **Tech Engine:** Go-Sysinfo + TimescaleDB + Centrifugo + Pion (WebRTC).

### **VortyxPilot (Professional Services Automation)**

- **Core Function:** Intelligent ticketing and helpdesk management.
- **Key Features:**
  - **Chained Remediation:** One-click "Playbooks" that execute multi-step fixes via VortyxPulse agents.
  - **SLA & Seniority Logic:** WIP limits and escalation workflows based on agent seniority and contract terms.
  - **Grace-Period Billing:** Automated time-entry reconciliation with flexible grace-period rounding.
- **Business Value:** Reduces "Ping-Pong" support conversations by 40% through automated triage.

### **VortyxNexus (IT Documentation)**

- **Core Function:** The "Brain" of the IT department (Wiki/Knowledge Base).
- **Key Features:**
  - **Recursive Impact Engine:** 3-hop graph analysis to identify downstream dependencies (e.g., "If Switch X fails, which VMs go down?").
  - **AES-GCM Vault:** Hardware-encrypted credential storage with mandatory "Justification" logs for access.
  - **Data Certification:** Automated worker that flags stale documentation for review.
- **Business Value:** Eliminates tribal knowledge; makes technician onboarding instant.

### **VortyxHorizon (vCIO & Strategy)**

- **Core Function:** Strategic IT planning and lifecycle management.
- **Key Features:**
  - **Holistic Health Score:** 0-100 algorithm weighted by Performance (40%), Security (40%), and Lifecycle (20%).
  - **Maroto QBR Reports:** Automated generation of pixel-perfect PDF Quarterly Business Reviews.
  - **Budget Forecasting:** 3-year refresh roadmaps based on asset lifecycle telemetry.

### **VortyxControl (SaaS Management)**

- **Core Function:** Management of cloud subscriptions (M365, Workspace).
- **Key Features:**
  - **Native SDK Integration:** Direct connectors for Microsoft Graph and Google Admin SDKs.
  - **SNI-Based Filtering:** TLS-level blocking of unauthorized SaaS applications.
  - **Automated Downgrade:** Identifies and downgrades underutilized premium licenses (e.g., E5 to E3).

### **VortyxOptic (CCTV Surveillance)**

- **Core Function:** Network Video Recorder (NVR) and AI-enhanced surveillance.
- **Key Features:**
  - **Pion/WebRTC Signaling:** Low-latency 4K streaming directly to the app via MediaMTX.
  - **TFLite Local Inference:** On-device person/object detection using quantized INT8 models (no cloud required).
  - **ROI Smart Search:** Search for motion events within specific "Regions of Interest" across the timeline.
- **Tech Engine:** MediaMTX \+ Pion \+ TFLite.

### **VortyxGrid (Network Management)**

- **Core Function:** Network infrastructure orchestration and configuration.
- **Key Features:**
  - **SSH Worker Pool:** High-concurrency SSH/Telnet pool for mass configuration deployment.
  - **TextFSM Normalization:** Converts raw CLI output from 50+ vendors into structured JSON.
  - **Recursive Topology:** Automated L2/L3 map generation using LLDP/CDP/ARP table analysis.

## **2\. Security Services (MSSP)**

**Theme:** Defense, Detection, and Reflex.

### **VortyxRadar (SIEM & Log Analysis)**

- **Core Function:** Centralized log aggregation and threat detection.
- **Key Features:**
  - Ingests logs from Windows Event Viewer, Syslog, and Firewalls.
  - Real-time Sigma rule matching for anomaly detection.
  - Forensic timeline search.
- **Tech Engine:** Expr \+ Sigma (Isolated Worker Mode).

### **VortyxGuard (Endpoint Detection & Response)**

- **Core Function:** Active endpoint protection and threat hunting.
- **Key Features:**
  - Process killing and network isolation for infected hosts.
  - File integrity monitoring.
  - Ransomware canary files.
- **Tech Engine:** Gopacket (Packet Capture).

### **VortyxShield (GRC & Compliance)**

- **Core Function:** Automated governance, risk, and compliance reporting.
- **Key Features:**
  - One-click ISO 27001 / GDPR gap analysis.
  - Automated PDF report generation.
  - Vendor risk assessment questionnaires.
- **Tech Engine:** Maroto (PDF Generation).

### **VortyxMind (Phishing Simulation)**

- **Core Function:** Employee security awareness training.
- **Key Features:**
  - Simulated phishing email campaigns.
  - "Teachable Moments" landing pages for failed tests.
  - Training progress tracking.
- **Tech Engine:** Gophish (Architecture Reference).

### **VortyxProbe (Vulnerability Scanner)**

- **Core Function:** Continuous identification of network weaknesses.
- **Key Features:**
  - External IP scanning (Ports, CVEs).
  - Internal network scanning (Weak passwords, unpatched OS).
  - Automated remediation ticket creation.
- **Tech Engine:** Nuclei.

### **VortyxReflex (SOAR / Automation)**

- **Core Function:** Automated security orchestration and response.
- **Key Features:**
  - Automated containment of threats based on Radar/Guard alerts.
  - API-driven workflows for incident remediation.
  - Integration with third-party security tools.

### **VortyxSonar (Network Detection)**

- **Core Function:** Real-time network traffic analysis and threat detection.
- **Tech Engine:** Suricata Wrapper.

### **VortyxSignal (Threat Intelligence)**

- **Core Function:** Aggregated threat intelligence feed synchronization.
