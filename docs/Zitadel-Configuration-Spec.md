# Zitadel Configuration & Architecture Specification for Vortyx

## 1. Architecture Analysis

### 1.1. Organizational Structure
Vortyx operates as a multi-tenant MSP/MSSP platform. The Zitadel architecture must reflect this hierarchy to ensure data isolation and proper access control.

*   **Zitadel Instance**: One instance hosting the entire platform.
*   **Default Organization ("Vortyx Provider")**:
    *   Owns the **Projects** (the software definitions).
    *   Manages "System Administrators" (Vortyx employees).
*   **Customer Organizations (Tenants)**:
    *   Each MSP Client (or the MSP itself if self-hosted) gets a distinct **Zitadel Organization**.
    *   Users (Technicians, Admins) belong to these organizations.
    *   **Project Grants** are used to share the "Vortyx Platform" project from the Provider Org to Customer Orgs.

### 1.2. Project Structure
To maintain a "Unified Monolith" experience while allowing granular control:

*   **Project Name**: `Vortyx Platform` (Single Project Strategy)
    *   **Rationale**: Vortyx is a unified suite. Splitting Pulse, Radar, etc., into separate *Zitadel Projects* would require users to authorize multiple times or manage complex token exchanges.
    *   **Module Control**: Access to modules (e.g., `pulse:read`, `radar:admin`) is managed via **Roles** within the single project.

### 1.3. Application Architecture
The architecture defines three primary application types interacting with Zitadel:

| Application | Type | Auth Method | Grant Type | Token Type |
| :--- | :--- | :--- | :--- | :--- |
| **Vortyx Frontend** | User Agent (Web) | PKCE | Authorization Code | JWT |
| **Vortyx Backend** | API (Resource Server) | Basic / Private Key | Introspection | JWT |
| **VORT Agent** | Machine User | JWT Profile | JWT Bearer | JWT |

### 1.4. Current Gaps & Recommendations
1.  **Legacy Grant Type**: The current `frontend/src/auth.ts` includes a `Credentials` provider (Password Grant).
    *   **Risk**: High. Exposes user credentials to the frontend server.
    *   **Recommendation**: Remove `Credentials` provider. Enforce standard OIDC (PKCE) flow via the `Zitadel` provider.
2.  **Role Assertion**: Backend relies on roles in the token.
    *   **Fix**: "Assert Roles on Authentication" must be enabled in the Project settings.
3.  **Machine Identity**: VORT agents currently have public registration endpoints.
    *   **Recommendation**: Transition to strict Machine User provisioning during agent install.

---

## 2. Zitadel Console Configuration Guide

### 2.1. Organization & Project Setup

1.  **Create/Select Organization**: `Vortyx Dev` (or your MSP name).
2.  **Create Project**:
    *   **Name**: `Vortyx Platform`
    *   **Settings**:
        *   Enable **"Assert Roles on Authentication"** (Critical for Backend RBAC).
        *   Enable **"Check for Project on Authentication"**.

### 2.2. Application Configuration

#### A. Frontend Application (Next.js)
*   **Name**: `vortyx-frontend`
*   **Type**: **Web Application**
*   **Authentication Method**: **PKCE** (Recommended)
*   **Redirect URIs**:
    *   Dev: `http://localhost:3000/api/auth/callback/zitadel`
    *   Prod: `https://vortyx.yourdomain.com/api/auth/callback/zitadel`
*   **Post Logout URIs**:
    *   Dev: `http://localhost:3000`
    *   Prod: `https://vortyx.yourdomain.com`
*   **Access Token Type**: **JWT** (Allows backend to validate locally without extra round-trips).
*   **ID Token**:
    *   Add **User Roles** to ID Token.
    *   Add **User Info** to ID Token.

#### B. Backend Application (API)
*   **Name**: `vortyx-backend`
*   **Type**: **API**
*   **Authentication Method**: **Basic** or **Private Key JWT** (for Introspection).
*   **Notes**: This app entry is mainly for the backend to perform introspection or administrative tasks if needed. The backend primarily *validates* tokens issued to the Frontend.

#### C. VORT Agent (Machine User)
*   **Name**: `vort-agent-{id}`
*   **Type**: **Machine User** (Service User)
*   **Access**:
    *   **Token Type**: JWT
    *   **Auth Method**: Private Key (JWT Profile)
*   **Key Expiration**: 90 Days (Rotated via orchestration).

### 2.3. Role-Based Access Control (RBAC)
Define these roles in the **Vortyx Platform** project:

| Role Key | Display Name | Description |
| :--- | :--- | :--- |
| `platform.admin` | Platform Admin | Full access to all Vortyx modules. |
| `msp.admin` | Tenant Admin | Admin access to own organization's data. |
| `pulse.viewer` | Pulse Viewer | Read-only access to RMM data. |
| `pulse.technician`| Pulse Tech | RMM operations (Scripts, Terminal). |
| `radar.viewer` | Radar Viewer | Read-only access to Security Events. |
| `radar.analyst` | Radar Analyst | Manage security incidents. |

---

## 3. OIDC Configuration Specifications

### 3.1. Endpoints (Standard)
*   **Issuer**: `http://localhost:8080` (Dev) / `https://auth.yourdomain.com` (Prod)
*   **Authorization**: `/oauth/v2/authorize`
*   **Token**: `/oauth/v2/token`
*   **User Info**: `/oidc/v1/userinfo`
*   **Introspection**: `/oauth/v2/introspect`
*   **End Session**: `/oidc/v1/end_session`
*   **JWKS**: `/oauth/v2/keys`

### 3.2. Scopes
*   **Frontend Request**: `openid email profile offline_access urn:zitadel:iam:org:project:id:vortyx:aud`
    *   `offline_access`: Required for Refresh Tokens.
    *   `urn:zitadel:iam:org:project:id:vortyx:aud`: Ensures the token is audienced for the Vortyx project (optional if "Project on Auth" is enabled).

---

## 4. Security & Compliance Checklist

### 4.1. Access Control
- [ ] **MFA Enforcement**:
    -   Go to **Organization Settings** -> **Login Policy**.
    -   Enable **"Force MFA"** (or at least for `platform.admin` roles).
    -   Allowed Second Factors: **OTP**, **U2F/FIDO2**.
- [ ] **Password Policy**:
    -   Minimum Length: **12 characters**.
    -   Require: Uppercase, Lowercase, Number, Symbol.
- [ ] **Lockout Policy**:
    -   Max Attempts: **5**.
    -   Lockout Duration: **10 minutes**.

### 4.2. Token Policies
- [ ] **Access Token Lifetime**: **15 minutes** (Short-lived for security).
- [ ] **Refresh Token Lifetime**: **7 days** (Idle expiration).
- [ ] **RefreshToken Rotation**: **Enabled** (New refresh token issued on every use).

### 4.3. CORS
- [ ] Ensure **CORS** is configured in Zitadel to allow requests from the Frontend Origin (`http://localhost:3000`).

---

## 5. Testing Procedures

### 5.1. User Login Flow
1.  Navigate to `http://localhost:3000`.
2.  Click "Sign In".
3.  Verify redirection to Zitadel Login UI.
4.  Enter credentials.
5.  Verify redirection back to `http://localhost:3000`.
6.  **Validation**: Check browser storage/cookies for `next-auth.session-token`.

### 5.2. Role Enforcement
1.  Assign `pulse.viewer` to User A.
2.  Login as User A.
3.  Attempt to access `VortyxPulse` Dashboard -> **Success**.
4.  Attempt to "Execute Script" (Requires `pulse.technician`) -> **Fail (403 Forbidden)**.
    *   *Note*: Ensure Backend logs show "Missing Role" error.

### 5.3. Token Refresh
1.  Wait 15 minutes (or manually expire token).
2.  Perform an action on the Frontend.
3.  **Success**: NextAuth should seamlessly refresh the token using `offline_access`.
4.  **Failure**: User is forced to log in again (Check Refresh Token Rotation settings).

### 5.4. Logout
1.  Click "Sign Out".
2.  Verify redirection to Zitadel End Session endpoint.
3.  Verify redirection back to Login screen.
4.  **Validation**: Try to use the old Access Token against the API -> **Fail (401 Unauthorized)**.

