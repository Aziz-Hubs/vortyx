# **VortyxRadar – SIEM & Log Analysis**

**Division:** VortyxSec (Security)  
**Architecture:** Log Aggregator / **Vort** Agent Module  
**Status:** [DESIGN]

## **1. Executive Summary**
VortyxRadar is the central nervous system of the Vortyx security suite. It is a Security Information and Event Management (SIEM) platform built for speed and simplicity. Unlike traditional SIEMs that require massive clusters, VortyxRadar uses a highly optimized Go-based ingestion engine and "Sigma" rules to detect threats in real-time with minimal latency.

## **2. Technical Architecture**

### **2.1 The Collector**
*   **Ingestion:** Acts as a Syslog server (UDP/TCP 514) and Windows Event Collector.
*   **Normalization:** Converts disparate log formats (JSON, CSV, raw text) into a unified "Vortyx Event Schema".

### **2.2 The Detection Engine**
*   **Logic:** Uses **Expr** (Expression Language for Go) to evaluate logs against detection rules.
*   **Standard:** Native support for **Sigma Rules**, allowing immediate usage of thousands of open-source threat detection rules.

### **2.3 Data Store**
*   **Hot Storage:** TimescaleDB for recent logs (searchable < 10ms).
*   **Cold Storage:** Compressed Parquet files for long-term compliance retention.

## **3. Core Features**

### **3.1 Real-Time Threat Detection**
*   **Correlation:** Detects patterns across multiple sources (e.g., "Failed Logins" + "Successful Admin Login").
*   **Rule Management:** Git-integrated rule repository with automated Sigma rule updates.

### **3.2 Forensic Search**
*   **Query Language:** A simplified SQL-like syntax for threat hunting.
*   **Timeline:** Interactive histogram showing event volume over time to identify anomalies.

### **3.3 Entity Behavior Analytics (UEBA)**
*   **Baselines:** Learns "normal" behavior patterns for users and entities.
*   **Anomalies:** Flags deviations from established baselines for immediate investigation.

## **4. Integration with Vortyx Ecosystem**
*   **VortyxReflex:** Triggers automated playbooks when high-severity security alerts are fired.
*   **VortyxPilot:** Creates security incident tickets for analyst investigation and tracking.
*   **VortyxPulse:** Deploys log shipping agents to endpoints and provides system context.
*   **VortyxGuard:** Correlates endpoint telemetry with network and system logs.
*   **VortyxShield:** Feeds security event data into GRC and compliance reports.

## **5. Technical Design Document (TDD)**

### **5.1 Architectural Infrastructure**
This module is a high-performance component of the **Vortyx Unified Monolith**.
*   **Core Engine:** Compiled **Go 1.24+** backend utilizing a modular service architecture.
*   **UI Layer:** **Next.js 16** for a responsive, high-fidelity security dashboard.
*   **Time-Series Core:** Integrated **TimescaleDB** hypertables for high-velocity log ingestion.

### **5.2 Security, Compliance & Audit**
*   **Identity:** Unified authentication via **Zitadel** (OIDC).
*   **Protection:** Mandatory **AES-256-GCM** encryption for all sensitive fields at rest.
*   **Auditability:** Immutable audit hypertables recording every security-critical action.

### **5.3 Dogfooding & Quality Assurance**
As part of the **VortyxCore** initiative, this app monitors all internal Vortyx security logs:
*   **Operational Role:** The primary monitoring tool for the Vortyx internal security posture.
*   **Verification:** Validated against internal security operations before deployment to clients.
