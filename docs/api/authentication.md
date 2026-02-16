# Authentication

Vortyx uses **OIDC (OpenID Connect)** and **Zitadel** for authentication and authorization.

## Overview

The authentication flow is based on standard OIDC/OAuth2 protocols:
1.  **Login**: User logs in via the Frontend (redirects to Zitadel).
2.  **Token**: Zitadel issues an `access_token` (JWT) and `id_token`.
3.  **Request**: Frontend sends the `access_token` in the `Authorization: Bearer <token>` header to the Backend.
4.  **Validation**: Backend validates the token using Zitadel SDK.
5.  **Context**: Validated user identity (sub, email, roles, org) is injected into the request context.

## Frontend Implementation

We use `next-auth` (or equivalent OIDC client library) to handle the redirect flow.

```tsx
// Example using NextAuth
import NextAuth from "next-auth";
import ZitadelProvider from "next-auth/providers/zitadel";

export const authOptions = {
  providers: [
    ZitadelProvider({
      clientId: process.env.ZITADEL_CLIENT_ID,
      clientSecret: process.env.ZITADEL_CLIENT_SECRET,
      issuer: process.env.ZITADEL_ISSUER,
    }),
  ],
  callbacks: {
    async jwt({ token, account }) {
      if (account) {
        token.accessToken = account.access_token;
      }
      return token;
    },
    async session({ session, token }) {
      session.accessToken = token.accessToken;
      return session;
    },
  },
};
```

## Backend Validation

The backend uses a middleware and ConnectRPC interceptor to validate tokens on every request using the Zitadel Go SDK.

### Creating the Authenticator

```go
import (
	"github.com/abdul/vortyx/backend/internal/server/interceptors"
)

authMw, connectOpts := interceptors.AuthMiddleware(interceptors.DefaultAuthConfig())
if authMw != nil {
	r.Use(authMw)
}

path, handler := someconnect.NewSomeServiceHandler(svc, connectOpts)
r.Mount(path, handler)
```

### Configuration Options

The authenticator supports several configuration options:

```go
// Set custom client ID
auth.NewZitadelAuthenticator(
    ctx,
    "localhost:8080",
    auth.WithClientID("your-client-id"),
)

// Use introspection instead of JWT verification
auth.NewZitadelAuthenticator(
    ctx,
    "localhost:8080",
    auth.WithVerificationMethod(auth.IntrospectionVerification),
)

// Configure service account key for introspection
auth.NewZitadelAuthenticator(
    ctx,
    "localhost:8080",
    auth.WithServiceAccountKey("/path/to/key.json", "client-id", "client-secret"),
)
```

### Token Verification Methods

1. **JWT Verification (Default)**: Local JWT validation using cached JWKS keys. This is the recommended method for high-throughput APIs as it performs validation locally without network calls.

2. **Token Introspection**: Validates tokens by calling Zitadel's introspection endpoint. This checks if the token has been revoked and is more secure but requires a network call for each request.

3. **Introspection with Cache**: Same as introspection but caches results to improve performance while still checking for token revocation.

### Extracting User Information

The middleware injects user information into the request context:

```go
// In your handler
func HandleGetUser(w http.ResponseWriter, r *http.Request) {
    userID := auth.GetUserID(r.Context())
    email := auth.GetEmail(r.Context())
    orgID := auth.GetOrganizationID(r.Context())
    roles := auth.GetRoles(r.Context())
    
    // Check if user has specific role
    if auth.HasRole(r.Context(), "admin") {
        // User is admin
    }
}
```

### ConnectRPC Integration

For ConnectRPC services, use the interceptor:

```go
import "connectrpc.com/connect"

// Create authenticator
authenticator, _ := auth.NewZitadelAuthenticator(ctx, "localhost:8080")

// Use interceptor with service
interceptors := connect.WithInterceptors(authenticator.Interceptor())

// Mount service with interceptor
path, handler := examplev1connect.NewExampleServiceHandler(&ExampleService{}, interceptors)
r.Mount(path, handler)
```

### Middleware Security Features

The Zitadel SDK middleware provides:

- **Token Signature Validation**: Verifies JWT signatures using Zitadel's public keys
- **Expiration Check**: Automatically validates token expiration
- **Audience Validation**: Ensures token was issued for the correct audience
- **Issuer Validation**: Verifies token was issued by trusted Zitadel instance
- **Claims Extraction**: Extracts user ID, email, organization, and roles from token

## Security Considerations

-   **HTTPS**: All traffic MUST be encrypted in production.
-   **Token Storage**: Tokens should be stored securely (e.g., `httpOnly` cookies or in memory), never in `localStorage` if possible to avoid XSS.
-   **Scopes**: Always request the minimum required scopes (`openid profile email`).
-   **Rotation**: Use refresh tokens to maintain sessions securely without long-lived access tokens.
-   **Insecure Mode**: Only use `zitadel.WithInsecure()` in development. Production must use TLS.
-   **Client ID**: Configure appropriate client ID for your application in production.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ZITADEL_DOMAIN` | Zitadel instance domain | `localhost:8080` |
| `ZITADEL_ISSUER` | Zitadel OIDC issuer URL | `http://localhost:8080` |
| `ZITADEL_CLIENT_ID` | Zitadel client ID | - |

## Testing

Run backend tests:

```bash
cd backend
go test ./... 
```
