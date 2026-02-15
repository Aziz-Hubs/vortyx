# **VortyxPilot – Unified Service Desk (PSA)**

**Division:** VortyxMSP (Services)  
**Architecture:** PSA / Helpdesk Engine  
**Status:** [GOLD_MASTER_READY]

## **1. Executive Summary**

VortyxPilot is the operational cockpit for the Vortyx platform. It handles the "Human" side of IT and Security management: Tickets, Projects, and Service Level Agreements (SLAs). VortyxPilot is uniquely unified—handling both traditional MSP helpdesk tickets and MSSP incident response (IR) workflows in a single pane of glass.

## **2. Core Features**

### **2.1 Service Desk (Ticketing)**

- **Ticket Management:** Kanban and List views for managing support incidents.
- **SLA Tracking:** Real-time countdown timers for Response and Resolution SLAs.
- **Priority Queues:** Automated routing of Critical/High priority tickets to tier-3 engineers or SOC analysts.
- **Incident Response:** Dedicated workflows for security incidents (mapped to NIST/SANS frameworks).

### **2.2 Project Management**

- **Project Tracking:** Manage long-term initiatives (e.g., "SOC Onboarding", "Cloud Migration").
- **Gantt/Timeline:** Visual progress tracking against deadlines.
- **Resource Allocation:** Assign technicians/analysts to projects and track utilization.

### **2.3 Chained Remediation (The "Magic" Button)**

- **Remediation Steps:** Link tickets directly to VortyxPulse or VortyxReflex actions.
- **One-Click Fixes:** Resolve issues directly from the ticket UI (e.g., "Restart Print Spooler", "Isolate Infected Host").
- **Playbooks:** Sequence multiple actions across multiple devices to resolve complex outages or security threats.

## **3. Integration with Vortyx Ecosystem**

- **VortyxPulse:** Deep 2-way sync. Alerts (e.g., High RAM, Server Offline) automatically create tickets with direct links to the remote management tools.
- **VortyxRadar:** Security alerts from the SIEM are escalated into VortyxPilot for investigation and resolution.
- **VortyxNexus:** Documentation and SOPs are linked directly to tickets to provide technicians with instant context.
- **VortyxReflex:** SOAR playbooks are triggered and tracked within the ticket lifecycle.

## **4. Future Roadmap**

- **AI Triage:** Auto-categorize tickets and suggest solutions based on historical SIEM/RMM data.
- **Customer Satisfaction (CSAT):** Automated surveys sent upon ticket closure.
- **Integrated Knowledge Base:** Solution articles directly accessible within the ticket view.
