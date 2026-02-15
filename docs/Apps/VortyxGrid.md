# **VortyxGrid – Network Automation & Topology**

**Division:** VortyxMSP (Infrastructure)  
**Architecture:** NetOps & Automation Engine  
**Status:** [GOLD_MASTER_READY]

## **1. Executive Summary**

VortyxGrid is the "Map of the Territory". It visualizes the physical and logical network topology, detecting devices via LLDP/CDP crawling. It goes beyond monitoring by offering "Configuration Drift Detection" – identifying when a device's config changes from the known-good baseline.

## **2. Core Features**

### **2.1 Topology Mapping**

- **Visual Map:** Interactive graph of switches, routers, and endpoints.
- **Status Indicators:** Color-coded badges for device health (Online, Offline, Drift).
- **Port Grid:** Visual switch faceplate to see port status (Up/Down/PoE) at a glance.

### **2.2 Configuration Governance**

- **Drift Engine:** Compares the live running-config vs. the stored startup-config.
- **Diff Viewer:** Highlights changes in configuration files (e.g., changed routes or VLANs).

### **2.3 Security Integration**

- **Lateral Movement Detection:** Visual alerts when suspicious MAC address movement is detected between switch ports.
- **Quarantine:** One-click execution to shut down a switch port hosting a compromised device via VortyxReflex.

## **3. Integration with Vortyx Ecosystem**

- **VortyxReflex:** The execution arm for network-level threat containment.
- **VortyxControl:** Uses network flow logs to find Shadow IT applications.
- **VortyxPulse:** Maps endpoints (Agents) to specific switch ports in the Grid.
- **VortyxSonar:** Provides the network visibility needed for deep packet analysis.
