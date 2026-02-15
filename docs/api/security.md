# Security Documentation

## Overview

Vortyx uses the **Zitadel SDK** for authentication and authorization. The security layer is implemented in `backend/internal/auth/` using the official Zitadel Go SDK (`zitadel-go/v3`).

## Table of Contents

1. [Architecture](#architecture)
2. [Authentication Middleware](#authentication-middleware)
3. [Configuration](#configuration)
4. [Environment Variables](#environment-variables)
5. [Testing](#testing)

---

## Architecture

The authentication system is implemented in a single package:

```
backend/internal/auth/
├── middleware.go        # Zitadel SDK authentication
└── middleware_test.go   # Unit tests
```

### Key Components

- **ZitadelAuthenticator**: Main authenticator using Zitadel SDK
- **HTTP Middleware**: Validates JWT tokens for chi router
- **ConnectRPC Interceptor**: Validates tokens for gRPC-style APIs

---

## Authentication Middleware

**File:** `middleware.go`

The authentication middleware verifies JWT tokens issued by Zitadel using the official Zitadel Go SDK.

### Features

- JWT token verification using Zitadel SDK (`zitadel-go/v3`)
- Claims extraction (user ID, email, roles, org ID)
- Context injection for downstream handlers
- Support for JWT verification (default) and token introspection
- ConnectRPC interceptor support for gRPC-style APIs

### Usage

```go
import (
    "context"
    "log"
    "os"

    "github.com/abdul/vortyx/backend/internal/auth"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
    "connectrpc.com/connect"
)

func NewRouter() http.Handler {
    r := chi.NewRouter()

    // Get Zitadel domain from environment
    zitadelDomain := os.Getenv("ZITADEL_DOMAIN")
    if zitadelDomain == "" {
        zitadelDomain = "localhost:8080"
    }

    // Create authenticator with Zitadel SDK
    authenticator, err := auth.NewZitadelAuthenticator(context.Background(), zitadelDomain)
    var interceptors connect.Option
    if err != nil {
        log.Printf("Warning: Failed to initialize authenticator: %v", err)
        interceptors = connect.WithInterceptors()
    } else {
        // Use HTTP middleware
        r.Use(authenticator.Middleware)
        
        // Use ConnectRPC interceptor
        interceptors = connect.WithInterceptors(authenticator.Interceptor())
    }

    // Basic middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    // CORS (using chi/cors)
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
        AllowCredentials: true,
        MaxAge:           300,
    }))

    // Mount services with interceptors
    // ...

    return r
}
```

### Configuration Options

The authenticator supports several configuration options:

```go
import "github.com/abdul/vortyx/backend/internal/auth"

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

| Option | Description |
|--------|-------------|
| `WithClientID(clientID)` | Set the client ID for token validation |
| `WithVerificationMethod(method)` | Set verification method: `JWTVerification`, `IntrospectionVerification`, `IntrospectionWithCache` |
| `WithServiceAccountKey(path, clientID, secret)` | Configure service account key for introspection |

### Token Verification Methods

1. **JWT Verification (Default)**: Local JWT validation using cached JWKS keys. Recommended for high-throughput APIs as it performs validation locally without network calls.

2. **Token Introspection**: Validates tokens by calling Zitadel's introspection endpoint. This checks if the token has been revoked and is more secure but requires a network call for each request.

3. **Introspection with Cache**: Same as introspection but caches results to improve performance while still checking for token revocation.

### Context Keys

The middleware injects the following values into the request context:

| Key | Type | Description |
|-----|------|-------------|
| `user_id` | string | Zitadel user ID |
| `email` | string | User email address |
| `roles` | []string | User roles |
| `org_id` | string | Organization ID |
| `username` | string | Username |

### Helper Functions

The auth package provides utility functions to extract user information:

```go
import "github.com/abdul/vortyx/backend/internal/auth"

// In your handler
func HandleGetUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Get user ID from context
    userID := auth.GetUserID(ctx)

    // Get user email
    email := auth.GetEmail(ctx)

    // Get organization ID
    orgID := auth.GetOrganizationID(ctx)

    // Get user roles
    roles := auth.GetRoles(ctx)

    // Get username
    username := auth.GetUsername(ctx)

    // Check if user has specific role
    if auth.HasRole(ctx, "admin") {
        // User is admin
    }
}
```

### ConnectRPC Integration

For ConnectRPC services, use the interceptor:

```go
import (
    "connectrpc.com/connect"
    "github.com/abdul/vortyx/backend/internal/auth"
)

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

---

## Configuration

### Quick Start

```go
func NewRouter() http.Handler {
    r := chi.NewRouter()

    zitadelDomain := os.Getenv("ZITADEL_DOMAIN")
    if zitadelDomain == "" {
        zitadelDomain = "localhost:8080"
    }

    authenticator, err := auth.NewZitadelAuthenticator(context.Background(), zitadelDomain)
    var interceptors connect.Option
    if err != nil {
        log.Printf("Warning: Failed to initialize authenticator: %v", err)
        interceptors = connect.WithInterceptors()
    } else {
        r.Use(authenticator.Middleware)
        interceptors = connect.WithInterceptors(authenticator.Interceptor())
    }

    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
        AllowCredentials: true,
        MaxAge:           300,
    }))

    // Mount services...

    return r
}
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ZITADEL_DOMAIN` | Zitadel instance domain | `localhost:8080` |
| `ZITADEL_ISSUER` | Zitadel OIDC issuer URL | `http://localhost:8080` |
| `ZITADEL_CLIENT_ID` | Zitadel client ID | - |
| `ZITADEL_PAT` | Zitadel Personal Access Token | - |

---

## Production Considerations

### Security Recommendations

1. **HTTPS**: All traffic MUST be encrypted in production. Remove any insecure configurations.

2. **Token Storage**: Tokens should be stored securely (e.g., `httpOnly` cookies or in memory), never in `localStorage` if possible to avoid XSS.

3. **CORS**: Configure appropriate allowed origins for your production frontend domain.

4. **Token Validation**: Use JWT verification (default) for high-performance APIs. Use introspection if you need to check token revocation.

### Performance

- JWT verification performs validation locally using cached JWKS keys
- For high-throughput APIs, JWT verification is recommended
- Token introspection requires network calls and may impact performance

---

## Testing

Run the authentication middleware tests:

```bash
cd backend
go test ./internal/auth/... -v
```

### Test Coverage

- Public endpoint detection
- Context value extraction (GetUserID, GetEmail, GetOrganizationID, GetRoles, GetUsername)
- Role checking (HasRole)
- Authorization header validation
- Middleware options configuration
- User context handling

---

## Dependencies

The auth middleware uses the following packages:

- `github.com/zitadel/zitadel-go/v3` - Zitadel SDK for authentication
- `github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth` - Token verification
- `github.com/go-chi/chi/v5` - HTTP routing
- `github.com/go-chi/cors` - CORS handling
- `connectrpc.com/connect` - ConnectRPC for gRPC-style APIs
