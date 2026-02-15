# ZITADEL Identity & Multi-Tenancy in Vortyx

This document outlines the Identity and Access Management (IAM) architecture for the Vortyx Platform, leveraging **ZITADEL** for authentication and a **Database-per-Tenant** strategy for data isolation.

---

## 1. What is ZITADEL?

**ZITADEL** is a cloud-native Identity & Access Management system. In the Vortyx ecosystem, it serves as our primary **OIDC (OpenID Connect) Provider**.

### Key Provisions:
- **Unified Identity**: A single set of credentials works across the entire Vortyx ecosystem.
- **Organization-Based Multi-Tenancy**: Built-in support for multiple "Organizations" (Tenants), each with its own users, roles, and security policies.
- **Security First**: Native support for MFA (OTP, Passkeys), Audit Logging, and SSO.
- **Performance**: High-concurrency Go-based architecture that matches the Vortyx backend philosophy.

---

## 2. Integration with Vortyx Platform

ZITADEL does not just "log users in"; it is the foundation of our multi-tenant infrastructure.

### The "Mapping" Strategy
Vortyx uses a **Database-per-Tenant** model. The connection between a ZITADEL Identity and a physical database is managed in the `vortyx_platform` registry.

| ZITADEL Component | Vortyx Platform Component |
| :--- | :--- |
| `Organization (Org)` | `Tenant` (Entry in `vortyx_platform.tenants`) |
| `Project` | `Vortyx Product Suite` |
| `Application` | `Backend API (Go)` / `Frontend (Next.js)` |
| `Role` | `Permission Set` (e.g., `platform_admin`) |

---

## 3. Database Implementation

The platform utilizes three distinct database roles to manage scale and isolation:

1.  **`vortyx_platform` (The Registry)**: 
    - Tracks every organization registered on the platform.
    - Contains the `tenants` table which maps `org_id` (from ZITADEL) to `db_name` (the physical tenant database).
2.  **`tenant_template` (The Blueprint)**:
    - Contains the baseline schema for all modules (Pulse, MSP, etc.).
    - Never stores real data; it is used as a template for cloning when a new organization signs up.
3.  **`tenant_<id>` (The Isolated Cell)**:
    - The actual production database for a specific customer.
    - Completely isolated from other clients.

---

## 4. Organization Onboarding Workflow

When a new organization buys Vortyx services, the following "Provisioning Loop" occurs:

1.  **IAM Initialization**:
    - The Platform Admin creates a new **Organization** in the ZITADEL Console.
    - A unique `org_id` is generated.
2.  **Platform Registration**:
    - The backend inserts a record into `vortyx_platform.tenants`:
      ```sql
      INSERT INTO tenants (org_id, db_name, region, status)
      VALUES ('org_123', 'tenant_acme_corp', 'us-east-1', 'active');
      ```
3.  **Infrastructure Provisioning**:
    - The backend executes a `Cloning` script:
      ```bash
      # Pseudo-logic
      CREATE DATABASE tenant_acme_corp WITH TEMPLATE tenant_template;
      ```
4.  **Admin Activation**:
    - An invitation is sent to the client's admin via ZITADEL.
    - Upon first login, they are granted the `tenant_admin` role for their specific organization.

---

## 5. Authentication & Authorization

### Authentication Options
Organizations have granular control over how their users authenticate:
- **Standard**: Username and Password.
- **Modern**: Passkeys (Biometrics/Hardware Keys) for passwordless security.
- **Federated**: OIDC/SAML (e.g., "Login with Okta" or "Login with Azure AD") for enterprise SSO.

### Authorization
Vortyx uses **Role-Based Access Control (RBAC)**:
- **`platform_admin`**: Managed by ZITADEL at the System level. Can manage all tenants.
- **`tenant_admin`**: Can manage users and settings within their specific organization.
- **`tenant_user`**: Standard access to modules (Pulse, Alerts, etc.).

---

## 6. Current Configuration & Rationale

- **Protocol**: OIDC with Code Flow (PKCE) for maximum security.
- **External Secure**: Set to `false` for development, `true` for production.
- **Role Assertions**: Enabled. This ensures that when ZITADEL issues a token, it explicitly lists the user's role (e.g., `platform_admin`), allowing the Go backend to authorize requests instantly without a database lookup.
