# **VortyxWave – Cloud VoIP & Telephony**

**Division:** VortyxMSP (Infrastructure)  
**Architecture:** Real-time Communication (Pion WebRTC + SIP)  
**Status:** [GOLD_MASTER_READY]

## **1. Executive Summary**
VortyxWave is a modern, cloud-native Voice over IP (VoIP) phone system embedded directly into the Vortyx platform. It replaces physical desk phones and legacy PBX systems with a softphone capable of handling calls, SMS, and complex routing logic. It leverages the "Native Bridge" pattern to ensure high-quality audio processing free from browser tab throttling.

## **2. Technical Architecture**

### **2.1 The Voice Engine**
*   **Library:** **Pion WebRTC** (Pure Go implementation) for media transport.
*   **Protocol:** SIP (Session Initiation Protocol) over WebSocket for signaling.
*   **Codecs:** Opus (high fidelity) and G.711 (legacy compatibility).

## **3. Core Features**

### **3.1 Unified Softphone**
*   **Dialer:** Integrated keypad and contact list within the Vortyx dashboard.
*   **Call Handling:** Hold, Transfer (Blind & Attended), Mute, and Conference.
*   **Media Engine:** Leveraged **Pion WebRTC (v4)** for native Go-based audio processing.
*   **Adaptive Jitter Buffer:** Dynamic backend buffer that adjusts to network conditions to maintain high audio quality.

### **3.2 Advanced Routing (PBX)**
*   **IVR Editor:** Visual "Drag-and-Drop" editor for building call menus.
*   **IVR Engine:** Backend state machine that interprets call-routing DAGs stored in the database.
*   **Ring Groups:** Strategies for team ringing (Simultaneous, Round Robin, Least Recently Called).

### **3.3 Omni-Channel & Storage**
*   **Recording:** Automated capture of RTP streams directly to secure cloud storage.
*   **Visual Voicemail:** Transcribed voicemail messages available within the user's dashboard.

## **4. Integration with Vortyx Ecosystem**
*   **VortyxPilot:** Attaches call logs and recording links to active support and security tickets.
*   **VortyxNexus:** References call-routing configurations and SOPs for phone system management.
*   **VortyxRadar:** Monitors SIP signaling for suspicious activity or denial-of-service attempts.

## **5. Technical Design Document (TDD)**

### **5.1 Architectural Infrastructure**
This module is a high-performance component of the **Vortyx Unified Monolith**.
*   **Core Engine:** Compiled **Go 1.24+** backend utilizing a modular service architecture.
*   **UI Layer:** **Next.js 16** for a modern, responsive softphone interface.
*   **Time-Series Core:** Integrated **TimescaleDB** for high-volume call logs and audit trails.

### **5.2 Security, Compliance & Audit**
*   **Identity:** Unified authentication via **Zitadel** (OIDC).
*   **Protection:** Mandatory **TLS 1.3** for all signaling and media encryption.
*   **Auditability:** Detailed call records and administrative changes stored in immutable audit hypertables.

### **5.3 Dogfooding & Quality Assurance**
As part of the **VortyxCore** initiative, this app is the primary telephony system used by Vortyx's internal teams.
*   **Verification:** Validated against internal communication requirements before being deployed to clients.
