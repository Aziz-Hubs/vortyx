# API Reference

This document outlines the API contracts defined in our Protocol Buffers.

## Authentication
All RPCs require the following header:
`Authorization: Bearer <oidc_access_token>`

## Base Service (`vortyx.v1`)

### `Ping`
Check system health and version.
-   **Request**: `PingRequest { message: string }`
-   **Response**: `PingResponse { message: string }` (Includes "Vortyx-Version" header)

---

## MSP Services

### Pulse (`vortyx.pulse.v1`)
-   **`GetStatus`**
    -   **Request**: `GetStatusRequest {}`
    -   **Response**: `GetStatusResponse { status: string, version: string }`
    -   **Description**: Returns the operational status of the RMM engine.

### Pilot (`vortyx.pilot.v1`)
-   **`GetStatus`**
    -   **Request**: `GetStatusRequest {}`
    -   **Response**: `GetStatusResponse { status: string, version: string }`
    -   **Description**: Returns the status of the ticketing system.

*(Repeat for Nexus, Horizon, Control, Optic, Grid)*

---

## MSSP Services

### Radar (`vortyx.radar.v1`)
-   **`GetStatus`**
    -   **Request**: `GetStatusRequest {}`
    -   **Response**: `GetStatusResponse { status: string, version: string }`
    -   **Description**: Returns the status of the SIEM ingestion pipeline.

### Guard (`vortyx.guard.v1`)
-   **`GetStatus`**
    -   **Request**: `GetStatusRequest {}`
    -   **Response**: `GetStatusResponse { status: string, version: string }`
    -   **Description**: Returns the status of the EDR agent manager.

*(Repeat for Shield, Mind, Probe, Reflex, Sonar, Signal)*

## Error Handling

Vortyx uses standard gRPC status codes mapped to HTTP status codes:

| gRPC Code | HTTP Code | Meaning |
| :--- | :--- | :--- |
| `OK` (0) | 200 | Success. |
| `INVALID_ARGUMENT` (3) | 400 | Client specified an invalid argument. |
| `UNAUTHENTICATED` (16) | 401 | Request does not have valid authentication credentials. |
| `PERMISSION_DENIED` (7) | 403 | The caller does not have permission to execute the specified operation. |
| `NOT_FOUND` (5) | 404 | Some requested entity (e.g., file or directory) was not found. |
| `INTERNAL` (13) | 500 | Internal server error. |
