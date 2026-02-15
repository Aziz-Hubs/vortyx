# Data Flow

This document visualizes the data flow through the Vortyx architecture.

## 1. Authentication Flow (Frontend -> Zitadel -> Backend)

```mermaid
sequenceDiagram
    participant User
    participant Frontend as Next.js Client
    participant Zitadel as Identity Provider
    participant Backend as Go Service
    
    User->>Frontend: Click "Login"
    Frontend->>Zitadel: Redirect (OIDC Auth Code Flow)
    User->>Zitadel: Enter Credentials (MFA)
    Zitadel->>Frontend: Callback with Auth Code
    Frontend->>Zitadel: Exchange Code for Tokens
    Zitadel-->>Frontend: Returns {access_token, id_token}
    Frontend->>Backend: API Request + Bearer Token
    Backend->>Backend: Validate Token Signature (JWKS)
    Backend-->>Frontend: Return Data
```

## 2. Telemetry Ingestion (Agent -> Pulse -> TimescaleDB)

```mermaid
sequenceDiagram
    participant Agent as Vort Agent
    participant Pulse as Pulse Service
    participant DB as TimescaleDB
    
    loop Every 30s
        Agent->>Pulse: gRPC Stream (CPU, RAM, Disk)
        Pulse->>Pulse: Validate & Buffer
        Pulse->>DB: Batch Insert (COPY)
    end
```

## 3. Remote Desktop (User -> RustDesk -> Agent)

```mermaid
sequenceDiagram
    participant Admin
    participant Frontend
    participant Server as RustDesk Server (hbbs/hbbr)
    participant Agent
    
    Admin->>Frontend: Request Remote Session
    Frontend->>Server: Get Connection Token
    Frontend->>Agent: Send "Connect" Command (via Pulse)
    Agent->>Server: Establish Tunnel
    Server-->>Frontend: Relay Stream
    Frontend-->>Admin: Display Remote Screen
```

## 4. Threat Detection (Log -> Radar -> Reflex)

```mermaid
sequenceDiagram
    participant Firewall
    participant Radar as Radar Service
    participant Engine as Sigma Engine
    participant Reflex as Reflex Service
    participant Guard as Guard Service
    
    Firewall->>Radar: Syslog Stream
    Radar->>Engine: Parse & Match Rules
    Engine->>Radar: Alert: "Brute Force Detected"
    Radar->>Reflex: Trigger Incident
    Reflex->>Guard: Execute Playbook: "Block IP"
    Guard->>Firewall: Update ACL
```
