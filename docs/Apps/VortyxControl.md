# **VortyxControl – SaaS Management Platform (SMP)**

**Division:** VortyxMSP (Services)  
**Architecture:** SaaS Discovery & Governance Engine  
**Status:** [GOLD_MASTER_READY]

## **1. Executive Summary**

VortyxControl is the "FinOps" and "SecOps" engine for SaaS applications. It helps organizations discover "Shadow IT" (unauthorized apps), optimize license spending, and enforce security compliance across the SaaS estate.

## **2. Core Features**

### **2.1 Managed Inventory**

- **Central Catalog:** Tracks all sanctioned SaaS applications (Microsoft 365, Google Workspace, Slack, etc.).
- **Usage Tracking:** Monitors license utilization and adoption rates.
- **Health Status:** Real-time connection status to SaaS APIs.

### **2.2 Shadow IT Discovery**

- **Network Correlation:** Correlates **VortyxGrid** network flow logs and **VortyxPulse** application inventory to identify unmanaged or high-risk SaaS applications (e.g., unauthorized cloud storage or unsanctioned AI tools).
- **Risk Scoring:** Assigns a risk score (0-100) to discovered apps based on data sensitivity and usage frequency.

### **2.3 FinOps & Optimization**

- **Automated Downgrades:** Policy engine that identifies users on expensive tiers who are not utilizing premium features, suggesting downgrades to save costs.
- **Savings Dashboard:** Visualizes the cost-saving opportunities from license reclamation.

## **3. Integration with Vortyx Ecosystem**

- **VortyxGrid:** Provides network flow logs for discovery of unauthorized cloud application usage.
- **VortyxPulse:** Reports installed desktop applications that act as gateways to SaaS services.
- **VortyxShield:** Feeds SaaS compliance data into the GRC reporting engine.
- **VortyxRadar:** Alerts on suspicious login activity or data exfiltration attempts in connected SaaS platforms.

## **4. Future Roadmap**

- **CASB Lite:** Inline blocking of high-risk SaaS apps via firewall integration.
- **Contract Renewal Alerts:** Automated alerts for upcoming SaaS renewals to prevent unexpected costs.
