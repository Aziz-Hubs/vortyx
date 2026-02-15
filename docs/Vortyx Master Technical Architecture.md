# **Vortyx - Master Technical Architecture**

**Version:** 5.3 (Unified MSSP + MSP Specification)
**Architecture Style:** Decoupled Fullstack Web Application
**Repository Strategy:** Monorepo (Next.js 16 + Go API)
**Target Launch:** Q3 2026

---

## **1. Architectural Vision: "The Decoupled Core"**

The Vortyx ecosystem is a **Modern Fullstack Platform** built on a decoupled architecture. This ensures that the user interface (Frontend) and the business logic (Backend) are isolated, allowing for independent scaling, specialized hosting, and a cleaner separation of concerns.

### **1.1 The "Vort" Unified Agent**
Despite the decoupled web architecture, the endpoint strategy remains unified.
- **Vort Agent:** A single, modular Go binary deployed to endpoints.
- **Communication:** Connects directly to the **Go Backend API** via gRPC (Heartbeats) and WebSockets (Live Mode).

### **1.2 The API Protocol (ConnectRPC)**
**ConnectRPC** serves as the type-safe bridge between the decoupled layers.
- **Go Backend:** Serves as the authoritative RPC server.
- **Next.js Frontend:** Acts as an RPC client, calling the Go API via type-safe TypeScript hooks. This eliminates the need for manual JSON mapping or REST endpoint management.

---

## **2. Core Technology Stack**

| Layer            | Technology                  | Implementation Detail                                                               |
| :--------------- | :-------------------------- | :---------------------------------------------------------------------------------- |
| **Frontend UI**  | **Next.js 16 (App Router)** | Deployed as a **Node.js** application. Handles SSR, RSC, and Client-side UI.        |
| **Core Backend** | **Go (Golang) 1.24+**       | Deployed as a **Static Binary**. Handles all Business Logic, RPCs, and DB access.   |
| **Identity / IAM**| **Zitadel**                 | OIDC Provider sidecar. Manages Auth for both the Frontend and the Backend.          |
| **Database**     | **PostgreSQL 16**           | Relational store for metadata and platform state.                                   |
| **Time-Series**  | **TimescaleDB**             | Sidecar for high-volume logs (Radar) and metrics (Pulse).                           |

---

## **3. Deployment & Infrastructure (The Decoupled Stack)**

Vortyx is orchestrated via Docker Compose or Kubernetes, with traffic typically routed through a **Reverse Proxy** (e.g., Nginx, Caddy, or Traefik).

### **3.1 The Multi-Service Model**
1.  **Frontend Container (Node.js):** Hosts the Next.js application. Optimized for UI rendering and server-side components.
2.  **Backend Container (Go):** Hosts the high-concurrency API. It handles all Agent connections and data processing.
3.  **Identity Sidecar (Zitadel):** Manages the shared session state across the domain.
4.  **Data Sidecar (TimescaleDB):** Provides the high-performance data engine for the entire stack.

### **3.2 Reverse Proxy Logic**
The Reverse Proxy handles the unified domain entry point:
- `vortyx.app/*` → Routed to the **Next.js Frontend**.
- `vortyx.app/rpc/*` → Routed to the **Go Backend API**.
- `vortyx.app/auth/*` → Routed to the **Zitadel Sidecar**.
and so on.
