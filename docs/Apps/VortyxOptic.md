# **VortyxOptic – Surveillance NVR & AI Inference**

**Division:** VortyxMSP (Infrastructure)  
**Architecture:** Video Management System (VMS)  
**Status:** [GOLD_MASTER_READY]

## **1. Executive Summary**

VortyxOptic is a browser-based Network Video Recorder (NVR) and AI Inference dashboard. It connects to IP cameras (ONVIF/RTSP), provides live low-latency streaming, and aggregates AI-detected events into a searchable timeline.

## **2. Core Features**

### **2.1 Live Monitoring**

- **Matrix View:** Grid layout for monitoring multiple camera streams simultaneously.
- **Low Latency:** WebRTC/MSE streaming for sub-second latency on LAN.
- **Health Monitoring:** Real-time status checks for all connected cameras.

### **2.2 PTZ Control (Pan-Tilt-Zoom)**

- **Soft Joystick:** On-screen controls for PTZ-enabled cameras.
- **Click-to-Center:** Interactive video feed targeting for calibrated PTZ units.
- **Auto-Track:** Intelligent motion tracking based on AI object detection.

### **2.3 AI & Forensics**

- **Inference Stream:** A real-time log of computer vision events (e.g., "Person Detected", "Vehicle License Plate").
- **Visual Forensics:** Interactive charts showing event density over time.
- **ROI Search:** Region-of-Interest motion searching across recorded timelines.

## **3. Integration with Vortyx Ecosystem**

- **VortyxGuard:** Trigger alerts or physical security protocols when a camera detects a breach.
- **VortyxGrid:** IoT VLAN management for isolating and securing camera traffic.
- **VortyxPulse:** Monitors the health and performance of the NVR edge hardware.
- **VortyxRadar:** Correlates physical motion events with digital security alerts.

## **4. Deployment Strategy**

- **Edge Recording:** High-bitrate loop recording on local Vortyx edge devices.
- **Archival:** Automated offloading of critical event clips to secure cloud storage for long-term retention.
