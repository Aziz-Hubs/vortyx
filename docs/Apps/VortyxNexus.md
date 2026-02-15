# **VortyxNexus – IT & Security Documentation**

**Division:** VortyxMSP (Infrastructure)  
**Architecture:** Graph Database / Wiki (Ent Graph + Zitadel Vault)  
**Status:** [GOLD_MASTER_READY]

## **1. Executive Summary**

VortyxNexus is the "Brain" of the Vortyx platform. It is a structured documentation platform designed to eliminate tribal knowledge. Unlike static wikis, VortyxNexus is a dynamic "living" system where documentation is updated automatically by the RMM (VortyxPulse) and linked deeply to the ticketing system (VortyxPilot). It handles secure credential storage, network mapping, and security standard operating procedures (SOPs).

## **2. Technical Architecture**

### **2.1 The Data Structure**

- **Graph Model:** Utilizes **Ent**'s graph capabilities to map complex dependencies (e.g., "Server A" _hosts_ "Virtual Machine B" _which runs_ "SQL Database C").
- **Security:** Field-level encryption for sensitive data (passwords, API keys) using AES-256.

### **2.2 The Security Vault**

- **Integration:** Deep integration with **Zitadel** for access control.
- **Audit:** "Break Glass" logging – every time a password is viewed, it is logged with a timestamp and user ID.

## **3. Core Features**

### **3.1 Flexible Asset Management**

- **Custom Types:** Define any asset class (e.g., SSL Certificates, Domain Names, Firewall Rules, Vendor Contacts).
- **Auto-Documentation:** Syncs with VortyxPulse to keep fields like "OS Version", "IP Address", and "Software Inventory" up to date.
- **Discovery Engine:** Reconciles VortyxPulse telemetry to maintain an accurate Configuration Management Database (CMDB).

### **3.2 Password Management**

- **Organization:** Hierarchical password storage (Client -> Site -> Device).
- **Sharing:** Secure, temporary **One-Time Links** for sharing credentials safely.
- **Masking:** Passwords are masked by default and require an explicit "Reveal" click.
- **Clipboard Security:** Automated clipboard clearing enforced via the backend.

### **3.3 Relationship Mapping**

- **Visualizer:** Interactive node-graph visualization of network and security dependencies.
- **Impact Analysis:** Interactive view showing all downstream devices and users affected by an outage or security breach.

### **3.4 Knowledge Base (SOPs)**

- **Editor:** Markdown-based editor with support for diagrams (Mermaid.js).
- **Linking:** Asset tagging within articles for instant context.
- **Governance:** Data certification module to ensure documentation remains accurate and up-to-date.

## **4. Integration with Vortyx Ecosystem**

- **VortyxPilot:** Relevant passwords and SOPs are displayed side-by-side with tickets.
- **VortyxPulse:** Provides raw telemetry to populate and reconcile asset fields automatically.
- **VortyxGuard:** Links security incidents to specific asset documentation for faster forensics.
- **VortyxRadar:** Provides context for SIEM alerts by mapping them to known infrastructure.

## **5. Technical Design Document (TDD)**

### **5.1 Architectural Infrastructure**

This module is a high-performance component of the **Vortyx Unified Monolith**, leveraging a single-binary distribution strategy.

- **Core Engine:** Compiled **Go 1.24+** backend utilizing a modular service architecture.
- **UI Layer:** **Next.js 16** with **Tailwind CSS** for a fast, modern experience.
- **Relational Model:** **PostgreSQL** with **Ent ORM** for graph-aware relationship management.
- **Time-Series Core:** Integrated **TimescaleDB** for high-volume audit logs and change history.

### **5.2 Security, Compliance & Audit**

- **Identity:** Unified authentication via **Zitadel** (OIDC).
- **Protection:** Mandatory **AES-256-GCM** encryption for all sensitive fields at rest.
- **Auditability:** Immutable audit hypertables recording every data access and modification.

### **5.3 Dogfooding & Quality Assurance**

As part of the **VortyxCore** initiative, this app is used internally to manage Vortyx's own infrastructure:
- **Operational Role:** The central repository for all internal Vortyx operational knowledge.
- **Verification:** All features are validated against internal use cases before being released to clients.
