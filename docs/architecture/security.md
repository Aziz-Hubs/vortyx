# Security Architecture

## 1. Authentication (OIDC/OAuth2)
Vortyx relies on **Zitadel** as the central Identity Provider (IdP).
-   **Protocol**: OpenID Connect (OIDC).
-   **Flow**: Authorization Code Flow with PKCE for SPAs (Next.js).
-   **Tokens**: Short-lived Access Tokens (JWT) + Refresh Tokens.
-   **Keys**: JWKS (JSON Web Key Set) hosted by Zitadel for signature verification.

## 2. Authorization (RBAC)
Role-Based Access Control is enforced at the service level.
-   **Roles**: `admin`, `technician`, `viewer`.
-   **Scopes**: `pulse:read`, `radar:write`, `system:admin`.
-   **Enforcement**: Middleware checks claims in the JWT `scope` field.

## 3. Data Protection
-   **In-Transit**: TLS 1.3 for all external communication. mTLS for internal service-to-service calls (future).
-   **At-Rest**:
    -   **Database**: Volume encryption (LUKS/dm-crypt) on the host.
    -   **Vault**: Application-level encryption (AES-256-GCM) for sensitive fields (e.g., Nexus credentials).
    -   **Backups**: Encrypted before upload to S3.

## 4. API Security
-   **Rate Limiting**: Per-IP and per-User limits (Redis-backed).
-   **Input Validation**: Strict validation of all Protobuf messages.
-   **CORS**: Restricted to trusted origins (`https://app.vortyx.io`).
-   **Headers**: HSTS, CSP, X-Frame-Options enforced by middleware.

## 5. Secret Management
-   **Development**: `.env` files (git-ignored).
-   **Production**: Secrets injected via Kubernetes Secrets or HashiCorp Vault.
-   **Rotation**: Database credentials and Zitadel keys rotated quarterly.
