# Zitadel Integration Guide

This document provides comprehensive documentation on Vortyx's integration with Zitadel for identity and access management.

---

## Table of Contents

1. [What is Zitadel?](#1-what-is-zitadel)
2. [Zitadel Architecture in Vortyx](#2-zitadel-architecture-in-vortyx)
3. [Current Implementation](#3-current-implementation)
4. [Features Currently Used](#4-features-currently-used)
5. [Future Considerations](#5-future-considerations)
6. [Zitadel Console Guide - Human Users](#6-zitadel-console-guide---human-users)
7. [Zitadel Console Guide - Service Users (Machine Authentication)](#7-zitadel-console-guide---service-users-machine-authentication)
8. [JWT Profile Grant for M2M Communication](#8-jwt-profile-grant-for-m2m-communication)

---

## 1. What is Zitadel?

**Zitadel** is a cloud-native Identity & Access Management (IAM) system built on Go. It provides:
- **OIDC/OAuth2 Provider**: Standards-based authentication
- **Multi-Tenancy**: Built-in organization management
- **MFA Support**: OTP, Passkeys, Biometrics
- **Machine Users**: Service account support for M2M authentication
- **Audit Logging**: Comprehensive action tracking
- **Role-Based Access Control (RBAC)**: Granular permissions

### Key Zitadel Concepts

| Concept | Description |
|---------|-------------|
| **Instance** | Single Zitadel deployment (one per environment) |
| **Organization** | Top-level tenant (maps to Vortyx customer) |
| **Project** | Grouping of applications (maps to Vortyx product suite) |
| **Application** | Individual service (maps to Vortyx API/Frontend) |
| **User** | Identity (Human or Machine) |
| **Role** | Permission set attached to user/workspace |

---

## 2. Zitadel Architecture in Vortyx

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           VORTYX ARCHITECTURE                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────┐     ┌──────────────────────────────────────────────┐    │
│  │   HUMAN      │     │              ZITADEL                        │    │
│  │   USERS      │     │                                              │    │
│  │              │     │  ┌────────────┐  ┌────────────┐           │    │
│  │  - Admin     │◄───►│  │  Human     │  │  Machine   │           │    │
│  │  - Technician│     │  │  Users     │  │  Users     │           │    │
│  │  - Viewer    │     │  │  (OIDC)    │  │  (JWT)     │           │    │
│  └──────────────┘     │  └────────────┘  └────────────┘           │    │
│                       │                                              │    │
│  ┌──────────────┐     │  ┌────────────┐  ┌────────────┐           │    │
│  │   VORT       │     │  │  Projects  │  │   Roles    │           │    │
│  │   AGENTS     │◄───►│  │            │  │            │           │    │
│  │  (Machine)   │     │  └────────────┘  └────────────┘           │    │
│  └──────────────┘     │                                              │    │
│                       └──────────────────────────────────────────────┘    │
│                                       │                                   │
│                                       ▼                                   │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                      VORTYX BACKEND                               │  │
│  │                                                                    │  │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐   │  │
│  │  │  Interceptors  │  │  Platform       │  │   Services      │   │  │
│  │  │  (Auth)        │  │  Service        │  │   (MSP/MSSP)   │   │  │
│  │  │                │  │  (User Mgmt)    │  │                 │   │  │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘   │  │
│  │                                                                    │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                       │                                   │
│                                       ▼                                   │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                      DATABASE (PostgreSQL)                        │  │
│  │                                                                    │  │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐ │  │
│  │  │  Platform  │  │   VORT     │  │    MSP     │  │   MSSP     │ │  │
│  │  │  (Audit)  │  │  (Agents)  │  │   DBs      │  │    DBs     │ │  │
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘ │  │
│  └───────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Authentication Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     AUTHENTICATION FLOWS                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  HUMAN USER (OIDC)                                                         │
│  ────────────────                                                          │
│                                                                             │
│    ┌─────────┐      ┌──────────┐      ┌───────────┐      ┌──────────┐    │
│    │ Browser │ ───► │ Frontend │ ───► │  Zitadel  │ ───► │ Backend  │    │
│    │         │      │          │      │  (Login)  │      │ (Validate│    │
│    └─────────┘      └──────────┘      └───────────┘      │  Token)  │    │
│        │                                    │             └──────────┘    │
│        │  1. Redirect to Zitadel          │                              │
│        │◄─────────────────────────────────┘                              │
│        │                                                            │
│        │  2. Login + MFA (if enabled)                                 │
│        │─────────────────────────────────────►                          │
│        │                                                            │
│        │  3. Return with OAuth code                                      │
│        │◄─────────────────────────────────────                          │
│        │                                                            │
│        │  4. Exchange for tokens                                        │
│        │─────────────────────────────────────►                          │
│        │                                                            │
│        │  5. Access Token (JWT)                                         │
│        │◄─────────────────────────────────────                          │
│        │                                                            │
│        │  6. API Request with Bearer Token                              │
│        │─────────────────────────────────────►                          │
│        │                                                            │
│                                                                             │
│  MACHINE AGENT (JWT Profile)                                                │
│  ────────────────────────                                                   │
│                                                                             │
│    ┌──────────┐      ┌───────────┐      ┌───────────┐      ┌──────────┐  │
│    │  VORT    │ ───► │  Zitadel  │ ───► │  Backend  │ ───► │ Database │  │
│    │  Agent   │      │ (Token    │      │ (Validate │      │          │  │
│    │          │      │  Endpoint)│      │  JWT)     │      │          │  │
│    └──────────┘      └───────────┘      └───────────┘      └──────────┘  │
│        │                                                            │
│        │  1. Get service user credentials                               │
│        │     (client_id + private_key)                                  │
│        │                                                            │
│        │  2. Request token from Zitadel                                │
│        │     (JWT Profile Grant)                                        │
│        │─────────────────────────────────────►                          │
│        │                                                            │
│        │  3. Receive JWT access token                                   │
│        │◄─────────────────────────────────────                          │
│        │                                                            │
│        │  4. API Request with Bearer Token                              │
│        │─────────────────────────────────────►                          │
│        │                                                            │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Current Implementation

### Backend Integration

The Vortyx backend uses Zitadel for authentication via interceptors:

**Location**: `backend/internal/server/interceptors/auth.go`

```go
// Key components:
- ZitadelAuthenticator: Handles JWT token validation
- AuthMiddleware: HTTP middleware for request authentication
- AuthInterceptor: ConnectRPC interceptor for gRPC-style APIs

// Features:
- JWT verification using Zitadel's JWKS
- Context extraction (UserID, Email, OrgID, Roles, Username)
- Public endpoint configuration
- Multiple verification methods (JWT, Introspection, Cached Introspection)
```

### Platform Service

**Location**: `backend/internal/platform/service.go`

```go
// Uses Zitadel Management SDK for user operations:
// - CreateUser: Creates human users in Zitadel
// - ListUsers: Lists users from Zitadel
// - DeleteUser: Removes users from Zitadel
```

### Public Endpoints

The following endpoints do not require authentication:

```go
var publicEndpoints = map[string]bool{
    "/health":              true,
    "/healthz":             true,
    "/ping":                true,
    "/api/v1/ping":         true,
    "/api/vort/v1/register":      true,  // VORT agent registration
    "/api/vort/v1/authenticate":   true,  // VORT agent authentication
}
```

---

## 4. Features Currently Used

### Human Users (OIDC)

| Feature | Status | Description |
|---------|--------|-------------|
| OIDC Authentication | ✅ Implemented | Standard OIDC code flow with PKCE |
| JWT Token Validation | ✅ Implemented | Local JWT verification using JWKS |
| Multi-Tenancy | ✅ Implemented | Organization-based isolation |
| Role Extraction | ✅ Implemented | Roles extracted from `urn:zitadel:iam:org:project:roles` claim |
| MFA | 🔲 Not Used | Not yet enabled in Vortyx |
| Passkeys | 🔲 Not Used | Not yet enabled in Vortyx |

### Frontend Security

| Feature | Status | Description |
|---------|--------|-------------|
| Route Protection | ✅ Implemented | Middleware enforces authentication |
| Session Cookie Security | ✅ Implemented | NextAuth session cookies |
| XSS Protection | ✅ Implemented | Error messages sanitized |
| Hardcoded Secrets | ✅ Fixed | PAT removed, requires env var |

### Machine Users (Service Accounts)

| Feature | Status | Description |
|---------|--------|-------------|
| Machine Users | ✅ Implemented | Each VORT agent can authenticate via Zitadel machine user |
| JWT Profile Grant | ✅ Implemented | JWT Profile Grant (`urn:ietf:params:oauth:grant-type:jwt-bearer`) |
| Service User Keys | ✅ Implemented | RSA key-based authentication |
| Agent Token Service | ✅ Implemented | Internal token fallback if Zitadel unavailable |
| bcrypt Credential Hashing | ✅ Implemented | Agent secrets stored with bcrypt (not SHA256) |

---

## 5. Implemented Features

### Phase 1: Machine User Integration (COMPLETED)

1. **Machine User Authentication**
   - VORT agents authenticate using Zitadel JWT Profile Grant
   - Located in: `backend/internal/vort/machineuser/auth.go`

2. **Agent Token Service**
   - Internal token generation as fallback if Zitadel unavailable
   - Located in: `backend/internal/vort/token/agent_token.go`
   - Uses RS256 signed JWTs with configurable expiry

3. **Secure Credential Storage**
   - Agent secrets hashed with bcrypt (cost factor 10)
   - Located in: `backend/internal/vort/service/service.go`

4. **Authentication Flow**
   - RegisterAgent: Public endpoint for initial agent registration
   - AuthenticateAgent: Validates credentials + issues JWT token
   - Heartbeat/Data endpoints: Require valid JWT (Zitadel or internal)

### Phase 2: Enhanced Security (IN PROGRESS)

1. **Fail-Open Authentication Fixed**
   - Backend now returns 503 if Zitadel unavailable
   - Non-public endpoints blocked when auth fails

2. **Frontend Route Protection**
   - Middleware at `frontend/src/middleware.ts`
   - Redirects unauthenticated users to login

3. **Production pprof Protection**
   - Profiling endpoints only enabled in non-production environments

---

## 6. Zitadel Console Guide - Human Users

### Accessing the Console

1. Navigate to your Zitadel instance: `https://your-zitadel-instance.com`
2. Login with admin credentials
3. You'll see the dashboard:

```
┌─────────────────────────────────────────────────────────────────┐
│  ZITADEL DASHBOARD                                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                 │
│  │  Users     │  │  Projects  │  │  Analytics │                 │
│  └────────────┘  └────────────┘  └────────────┘                 │
│                                                                  │
│  [Organization Name]                                            │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Users    │  Projects  │  Roles  │  Settings  │  Audit    │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Creating a Human User

1. **Navigate to Users**
   - Click "Users" in the left sidebar
   - Click "New" button

2. **Fill User Details**
   ```
   Username: john.doe
   First Name: John
   Last Name: Doe
   Email: john.doe@company.com
   ```

3. **Set Initial Password**
   - Check "Set initial password"
   - Enter temporary password
   - User must change on first login

4. **Assign Roles** (Optional)
   - Click "Add Role"
   - Select from available roles

5. **Create User**
   - Click "Create"
   - User receives email with login instructions

### Managing User Roles

1. **Navigate to User**
   - Go to Users → Select User

2. **View/Edit Roles**
   - Click "Assignments" tab
   - Add or remove role memberships

3. **Common Roles in Vortyx**
   - `platform_admin`: Full platform access
   - `tenant_admin`: Organization admin
   - `tenant_user`: Standard access

### Enabling MFA for Users

1. **For Individual Users**
   - Go to Users → Select User
   - Click "MFA" tab
   - Enroll OTP or Passkey

2. **Organization-Wide Policy**
   - Go to Settings → Security Policies
   - Enable "Multi-Factor Authentication"
   - Choose MFA method (OTP, Passkey, or Both)

---

## 7. Zitadel Console Guide - Service Users (Machine Authentication)

Service users provide machine-to-machine (M2M) authentication for VORT agents.

### Creating a Service User

1. **Navigate to Service Users**
   - Go to Users → Switch to "Service Users" tab
   - Click "New"

2. **Fill Service User Details**
   ```
   Username: vort-agent-001
   Name: VORT Agent 001
   Description: Production monitoring agent
   ```

3. **Select Authentication**
   - Choose "JWT" (recommended for machine users)
   - Or choose "Static" (client secret - less secure)

4. **Create Service User**
   - Click "Create"
   - **Important**: Copy and securely store the client ID

### Generating API Keys (for JWT Authentication)

1. **Navigate to Service User**
   - Go to Users → Service Users
   - Select your service user

2. **Add Key**
   - Click "Add Key" button

3. **Configure Key**
   ```
   Key Type: JSON
   Expiration: 1 year (or your preference)
   ```

4. **Download Key**
   - Click "Download"
   - **Important**: This is the only time you'll see the private key!
   - Store securely (never commit to version control)

   ```json
   {
     "type": "serviceaccount",
     "key_id": "xxxxx",
     "key": "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
     "client_id": "service-xxxxx@project-xxxxx"
   }
   ```

### Assigning Roles to Service Users

1. **Navigate to Service User**
   - Go to Users → Service Users
   - Select your service user

2. **Add Role Assignment**
   - Click "Assignments"
   - Add project and roles

3. **Common Roles for Agents**
   - `agent`: Basic agent operations
   - `agent:execute`: Command execution
   - `agent:upload`: Data upload
   - `agent:heartbeat`: Health reporting

---

## 8. JWT Profile Grant for M2M Communication

The JWT Profile Grant allows service users to authenticate without a client secret.

### Token Request Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                 JWT PROFILE GRANT FLOW                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. PREPARE JWT ASSERTION                                       │
│     ────────────────────                                         │
│     Header:                                                     │
│     {                                                           │
│       "alg": "RS256",                                           │
│       "typ": "JWT",                                             │
│       "kid": "<key_id>"                                         │
│     }                                                           │
│                                                                  │
│     Payload:                                                    │
│     {                                                           │
│       "iss": "<client_id>",                                     │
│       "sub": "<client_id>",                                     │
│       "aud": "https://<zitadel-domain>/oauth/token",          │
│       "iat": <current_timestamp>,                               │
│       "exp": <current_timestamp + 300>                          │
│     }                                                           │
│                                                                  │
│     Signature: RS256 with private key                           │
│                                                                  │
│  2. EXCHANGE FOR TOKEN                                          │
│     ────────────────────                                         │
│     POST https://<zitadel-domain>/oauth/token                  │
│     Content-Type: application/x-www-form-urlencoded             │
│                                                                  │
│     grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer     │
│     assertion=<jwt_assertion>                                   │
│     client_id=<client_id>                                       │
│                                                                  │
│  3. RECEIVE ACCESS TOKEN                                        │
│     ────────────────────                                         │
│     {                                                           │
│       "access_token": "eyJ...",                                 │
│       "token_type": "Bearer",                                   │
│       "expires_in": 43199                                        │
│     }                                                           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Implementation Example (Go)

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/url"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type ZitadelClient struct {
    domain     string
    clientID   string
    privateKey string
    keyID      string
}

type TokenResponse struct {
    AccessToken string `json:"access_token"`
    TokenType   string `json:"token_type"`
    ExpiresIn  int    `json:"expires_in"`
}

func (z *ZitadelClient) GetToken(ctx context.Context) (string, error) {
    // Create JWT assertion
    now := time.Now()
    claims := jwt.MapClaims{
        "iss": z.clientID,
        "sub": z.clientID,
        "aud": fmt.Sprintf("https://%s/oauth/token", z.domain),
        "iat": now.Unix(),
        "exp": now.Add(5 * time.Minute).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    token.Header["kid"] = z.keyID

    signedToken, err := token.SignedString([]byte(z.privateKey))
    if err != nil {
        return "", fmt.Errorf("failed to sign token: %w", err)
    }

    // Exchange for access token
    data := url.Values{}
    data.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
    data.Set("assertion", signedToken)
    data.Set("client_id", z.clientID)

    req, err := http.NewRequestWithContext(
        ctx,
        "POST",
        fmt.Sprintf("https://%s/oauth/token", z.domain),
        strings.NewReader(data.Encode()),
    )
    if err != nil {
        return "", err
    }

    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("token request failed: %s", body)
    }

    var tokenResp TokenResponse
    if err := json.Unmarshal(body, &tokenResp); err != nil {
        return "", err
    }

    return tokenResp.AccessToken, nil
}
```

### Using the Token

```go
// Create client
client := &ZitadelClient{
    domain:     "your-zitadel-instance.com",
    clientID:   "service-xxxxx@project-xxxxx",
    privateKey: "...pem content...",
    keyID:      "xxxxx",
}

// Get token
token, err := client.GetToken(context.Background())
if err != nil {
    log.Fatal(err)
}

// Use in API request
req, _ := http.NewRequest("GET", "https://api.vortyx.com/v1/agents", nil)
req.Header.Set("Authorization", "Bearer "+token)

client := &http.Client{}
resp, err := client.Do(req)
```

---

## Appendix: Environment Variables

### Core Zitadel Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `ZITADEL_DOMAIN` | Zitadel instance domain | Yes |
| `ZITADEL_ISSUER` | Zitadel OIDC issuer URL | Yes |
| `ZITADEL_CLIENT_ID` | OIDC client ID for `vortyx-frontend` | Yes |
| `ZITADEL_CLIENT_SECRET` | OIDC client secret for `vortyx-frontend` | Yes |
| `ZITADEL_API_PROJECT_ID` | Vortyx Platform project resource ID (reserved `aud` scope + grants) | Recommended |
| `ZITADEL_AUDIENCES` | Comma-separated accepted `aud` values for backend token validation | Recommended |
| `ZITADEL_INSECURE` | Use insecure connections to Zitadel (local dev only) | Dev only |
| `ZITADEL_PAT` | PAT for Management API (used by PlatformService) | Optional |
| `ZITADEL_SERVICE_USER_KEY_PATH` | Service-user key file path for Management API | Optional |

### VORT Agent Machine User Authentication

| Variable | Description | Required |
|----------|-------------|----------|
| `VORT_MACHINE_USER_KEY_PATH` | Path to RSA private key file for JWT Profile Grant | No* |
| `VORT_MACHINE_USER_KEY` | Base64-encoded RSA private key (alternative to KEY_PATH) | No* |
| `VORT_MACHINE_USER_KEY_ID` | Key ID for JWT header (required when using machine user auth) | No* |

*Either `VORT_MACHINE_USER_KEY_PATH` or `VORT_MACHINE_USER_KEY` is required for Zitadel machine user auth. `VORT_MACHINE_USER_KEY_ID` is required when providing a private key. If not configured, the system falls back to internal token generation.

### Internal Agent Token Service

| Variable | Description | Default |
|----------|-------------|---------|
| `VORT_AGENT_JWT_PRIVATE_KEY` | Base64-encoded RSA private key for internal token signing | Auto-generated |
| `VORT_AGENT_JWT_ISSUER` | JWT issuer for internal tokens | `vortyx-agent-auth` |
| `VORT_AGENT_JWT_AUDIENCE` | JWT audience for internal tokens | `vortyx-api` |

## Appendix: API Endpoints

| Endpoint | Auth Required | Description |
|----------|---------------|-------------|
| `GET /health` | No | Health check |
| `POST /api/vort/v1/register` | No | Register new agent |
| `POST /api/vort/v1/authenticate` | No | Authenticate agent |
| `POST /api/vort/v1/heartbeat` | Yes (Zitadel JWT) | Agent heartbeat |
| `GET /api/vort/v1/agents` | Yes (Zitadel JWT) | List agents |
| `POST /api/vort/v1/data` | Yes (Zitadel JWT) | Submit agent data |

---

*Last Updated: 2026-02-16*
