# **VortyxGuard – Endpoint Detection & Response (EDR)**

**Division:** VortyxSec (Security)  
**Architecture:** Unified Agent Module (**Vort**)  
**Status:** [DESIGN]

## **1. Executive Summary**

VortyxGuard is the security-hardened module of the unified **Vort** agent. While the base **Vort** agent handles Pulse (RMM) functionality, toggling the **Guard** module activates real-time kernel-level monitoring, automated threat defense, and ransomware containment—all without installing additional software.

## **2. Technical Architecture**

### **2.1 The Defender**

- **Kernel Integration:** Utilizes kernel drivers (Sysmon / eBPF on Linux) for deep visibility.
- **Engine:** **Gopacket** for local packet inspection and process tree analysis.

## **3. Core Features**

### **3.1 Active Protection**

- **Process Blocking:** Prevents execution of unsigned or blacklisted binaries.
- **Ransomware Canary:** Monitors "Honeypot" files. If a process tries to modify them, the process is instantly killed and the device isolated.
- **USB Control:** Blocks or Read-Only mounts unauthorized USB storage devices.

### **3.2 Investigation**

- **Flight Recorder:** Continuous recording of process executions, file mods, and network conns (metadata sent to Cloud).
- **Live Shell:** Secure remote shell for analysts to perform forensic cleanup.

### **3.3 Isolation**

- **Network Quarantine:** Can cut off all network access except to the Vortyx management server, preventing lateral movement.

## **4. Integration with Vortyx Ecosystem**

- **VortyxPulse:** Deploys and manages the VortyxGuard module.
- **VortyxRadar:** Streams EDR telemetry to the SIEM.
- **VortyxSignal:** Consumes hash blocklists to stop known malware.
- **VortyxNexus:** Integrates with the application inventory for documentation.
- **VortyxReflex:** Triggers automated isolation playbooks.

## **5. Technical Design Document (TDD)**

### **5.1 Architectural Infrastructure**

This module is a high-performance component of the **Vortyx Unified Monolith**.

- **Core Engine:** Compiled **Go 1.24+** backend utilizing a modular service architecture.
- **UI Layer:** **Next.js 16** for a responsive security management interface.
- **Communication:** Internal function-call routing for inter-app telemetry and **Centrifugo** for real-time security events.

### **5.2 Security, Compliance & Audit**

- **Identity:** Unified authentication via **Zitadel** (OIDC).
- **Protection:** Mandatory **AES-256-GCM** encryption for all sensitive fields at rest.
- **Auditability:** Every transaction is recorded in an immutable audit hypertable.

### **5.3 Dogfooding & Quality Assurance**

As part of the **VortyxCore** initiative, this app is the primary defense layer on every Vortyx internal workstation.
- **Verification:** Features are validated against internal security operations before deployment to clients.
