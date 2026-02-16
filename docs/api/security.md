# Security Documentation

## Overview

Vortyx uses the **Zitadel SDK** for authentication and authorization. The security layer is implemented in `backend/internal/server/interceptors/` using the official Zitadel Go SDK (`zitadel-go/v3`).

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
backend/internal/server/interceptors/
└── auth.go
```

### Key Components

- **ZitadelAuthenticator**: Main authenticator using Zitadel SDK
- **HTTP Middleware**: Validates JWT tokens for chi router
- **ConnectRPC Interceptor**: Validates tokens for gRPC-style APIs

---

## Authentication Middleware

**File:** `backend/internal/server/interceptors/auth.go`

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
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/abdul/vortyx/backend/internal/server/interceptors"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	authMw, connectOpts := interceptors.AuthMiddleware(interceptors.DefaultAuthConfig())
	if authMw != nil {
		r.Use(authMw)
	}
	_ = connectOpts
	return r
}
```

### Configuration Options

The authenticator supports several configuration options:

```go
import "github.com/abdul/vortyx/backend/internal/server/interceptors"

// Set custom client ID
interceptors.NewZitadelAuthenticator(
    ctx,
    "localhost:8080",
    interceptors.WithClientID("your-client-id"),
)

// Use introspection instead of JWT verification
interceptors.NewZitadelAuthenticator(
    ctx,
    "localhost:8080",
    interceptors.WithVerificationMethod(interceptors.IntrospectionVerification),
)

// Configure service account key for introspection
interceptors.NewZitadelAuthenticator(
    ctx,
    "localhost:8080",
    interceptors.WithServiceAccountKey("/path/to/key.json", "client-id", "client-secret"),
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

The interceptors package provides utility functions to extract user information:

```go
import "github.com/abdul/vortyx/backend/internal/server/interceptors"

// In your handler
func HandleGetUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Get user ID from context
    userID := interceptors.GetUserID(ctx)

    // Get user email
    email := interceptors.GetEmail(ctx)

    // Get organization ID
    orgID := interceptors.GetOrganizationID(ctx)

    // Get user roles
    roles := interceptors.GetRoles(ctx)

    // Get username
    username := interceptors.GetUsername(ctx)

    // Check if user has specific role
    if interceptors.HasRole(ctx, "admin") {
        // User is admin
    }
}
```

### ConnectRPC Integration

For ConnectRPC services, use the interceptor:

```go
import (
    "connectrpc.com/connect"
    "github.com/abdul/vortyx/backend/internal/server/interceptors"
)

// Create authenticator
authenticator, _ := interceptors.NewZitadelAuthenticator(ctx, "localhost:8080")

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
- **Fail-Closed Security**: Returns 503 if authentication fails (not fail-open)

---

## VORT Agent Authentication

The VORT agent system has additional authentication components:

```
backend/internal/vort/
├── machineuser/
│   └── auth.go          # Zitadel JWT Profile Grant authentication
├── token/
│   └── agent_token.go  # Internal JWT token service
└── service/
    └── service.go      # Agent service with bcrypt credential hashing
```

### Authentication Flow

1. **RegisterAgent** (Public): Agent registers and receives encrypted credentials
2. **AuthenticateAgent** (Public): Agent validates credentials, receives JWT token
3. **Protected Endpoints**: Require valid JWT (Zitadel or internal)

### Security Features

- **bcrypt Hashing**: Agent secrets stored with bcrypt (cost factor 10)
- **JWT Profile Grant**: Zitadel machine user authentication supported
- **Internal Token Fallback**: Automatic fallback if Zitadel unavailable
- **Token Expiry**: 24-hour token validity

### Environment Variables

| Variable | Description |
|----------|-------------|
| `ZITADEL_VORT_SERVICE_USER_KEY_PATH` | Path to RSA private key for Zitadel auth |
| `ZITADEL_VORT_SERVICE_USER_KEY_ID` | Key ID for JWT header |
| `ZITADEL_VORT_AGENT_JWT_PRIVATE_KEY` | Private key for internal tokens (optional) |

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

    authMw, connectOpts := interceptors.AuthMiddleware(interceptors.DefaultAuthConfig())
    r.Use(authMw)
    _ = connectOpts

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

> **Note**: If Zitadel authentication fails to initialize, the middleware enters fail-closed mode and returns 503 for all non-public endpoints. This prevents unauthorized access when the identity provider is unavailable.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ZITADEL_DOMAIN` | Zitadel instance domain | `localhost:8080` |
| `ZITADEL_ISSUER` | Zitadel OIDC issuer URL | `http://localhost:8080` |
| `ZITADEL_CLIENT_ID` | Zitadel client ID | - |
| `ZITADEL_MANAGEMENT_PAT` | Zitadel Management PAT (IAM OWNER) | - |

---

## Production Considerations

### Security Recommendations

1. **HTTPS**: All traffic MUST be encrypted in production. Remove any insecure configurations.

2. **Token Storage**: Tokens should be stored securely (e.g., `httpOnly` cookies or in memory), never in `localStorage` if possible to avoid XSS.

3. **CORS**: Configure appropriate allowed origins for your production frontend domain.

4. **Token Validation**: Use JWT verification (default) for high-performance APIs. Use introspection if you need to check token revocation.

5. **pprof Protection**: Profiling endpoints (`/debug/pprof/*`) are automatically disabled when `ENV=production`. Never expose pprof in production.

### Performance

- JWT verification performs validation locally using cached JWKS keys
- For high-throughput APIs, JWT verification is recommended
- Token introspection requires network calls and may impact performance

---

## Testing

Run backend tests:

```bash
cd backend
go test ./... 
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
- `github.com/golang-jwt/jwt/v5` - JWT token handling
- `golang.org/x/crypto/bcrypt` - Secure password hashing
- `github.com/rs/zerolog` - Structured logging
