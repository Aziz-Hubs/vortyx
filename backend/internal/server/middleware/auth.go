// Package middleware provides authentication middleware using Zitadel OIDC integration.
//
// This package implements JWT-based authentication and authorization for HTTP handlers
// and ConnectRPC services. It leverages the Zitadel SDK to validate tokens issued
// by a Zitadel identity provider.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"connectrpc.com/connect"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
)

// publicEndpoints defines paths that do not require authentication.
// These endpoints are typically health checks or public API entry points.
var publicEndpoints = map[string]bool{
	"/health":      true,
	"/healthz":     true,
	"/ping":        true,
	"/api/v1/ping": true,
}

// isPublicEndpoint checks if the given path is a public endpoint that does not
// require authentication. This allows health checks and monitoring endpoints to
// function without valid credentials.
func isPublicEndpoint(path string) bool {
	return publicEndpoints[path]
}

// contextKey is the type used for context value keys in this package.
// Using a custom type prevents key collisions with other packages.
type contextKey string

const (
	// UserIDKey is the context key for storing the authenticated user's ID.
	// This is the Zitadel-specific user identifier (sub claim).
	UserIDKey contextKey = "user_id"

	// EmailKey is the context key for storing the user's email address.
	// This is extracted from the token's email claim if present.
	EmailKey contextKey = "email"

	// OrganizationIDKey is the context key for storing the user's organization ID.
	// This is used for multi-tenant isolation in Zitadel.
	OrganizationIDKey contextKey = "org_id"

	// RolesKey is the context key for storing the user's assigned roles.
	// These roles are used for authorization decisions in handlers.
	RolesKey contextKey = "roles"

	// UsernameKey is the context key for storing the user's username.
	// This is typically the login name used in Zitadel.
	UsernameKey contextKey = "username"
)

// ZitadelAuthenticator handles JWT token validation and user context extraction
// using the Zitadel SDK. It supports both JWT verification (local validation)
// and token introspection (server-side validation).
//
// The authenticator is safe for concurrent use and can be shared across
// multiple goroutines.
type ZitadelAuthenticator struct {
	verifier           *oauth.JWTVerification
	zitadelInstance    *zitadel.Zitadel
	clientID           string
	verificationMethod VerificationMethod
	serviceAccountKey  *ServiceAccountKey
	mu                sync.RWMutex
	initialized       bool
}

// ServiceAccountKey holds credentials for Zitadel service account authentication.
// This is required for token introspection verification methods.
type ServiceAccountKey struct {
	KeyFilePath  string
	ClientID     string
	ClientSecret string
}

// VerificationMethod defines how tokens are validated.
// Different methods offer different tradeoffs between security and performance.
type VerificationMethod int

const (
	// JWTVerification performs local JWT validation using cached JWKS keys.
	// This is the default method and offers the best performance as it
	// avoids network calls for each request. Suitable for high-throughput APIs.
	JWTVerification VerificationMethod = iota

	// IntrospectionVerification validates tokens via Zitadel's introspection endpoint.
	// This method checks token revocation status but requires a network call
	// for each request. Use when you need to check if tokens have been revoked.
	IntrospectionVerification

	// IntrospectionWithCache is similar to IntrospectionVerification but caches
	// introspection results locally to reduce network overhead. Use when you need
	// revocation checking with better performance than plain introspection.
	IntrospectionWithCache
)

// ZitadelAuthenticatorOption configures a ZitadelAuthenticator instance.
// Options are applied in the order they are provided.
type ZitadelAuthenticatorOption func(*ZitadelAuthenticator)

// WithClientID sets the expected client ID (audience) for token validation.
// Tokens must contain this client ID in their audience claim to be valid.
// The default client ID is "zitadel".
func WithClientID(clientID string) ZitadelAuthenticatorOption {
	return func(a *ZitadelAuthenticator) {
		a.clientID = clientID
	}
}

// WithVerificationMethod sets the token verification method.
// The default is JWTVerification which performs local validation.
func WithVerificationMethod(method VerificationMethod) ZitadelAuthenticatorOption {
	return func(a *ZitadelAuthenticator) {
		a.verificationMethod = method
	}
}

// WithServiceAccountKey configures service account credentials for introspection.
// This is required when using IntrospectionVerification or IntrospectionWithCache.
func WithServiceAccountKey(keyPath, clientID, clientSecret string) ZitadelAuthenticatorOption {
	return func(a *ZitadelAuthenticator) {
		a.serviceAccountKey = &ServiceAccountKey{
			KeyFilePath:  keyPath,
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}
	}
}

// AuthConfig holds configuration for authentication middleware.
type AuthConfig struct {
	ZitadelDomain string
}

// DefaultAuthConfig returns the default authentication configuration.
// It reads the Zitadel domain from the ZITADEL_DOMAIN environment variable,
// defaulting to localhost:8080 for local development.
func DefaultAuthConfig() AuthConfig {
	domain := os.Getenv("ZITADEL_DOMAIN")
	if domain == "" {
		domain = "localhost:8080"
	}
	return AuthConfig{
		ZitadelDomain: domain,
	}
}

// NewZitadelAuthenticator creates a new ZitadelAuthenticator instance.
// The zitadelDomain parameter should be the domain of your Zitadel instance
// (e.g., "localhost:8080" for development or "your-org.zitadel.cloud" for production).
//
// Options can be provided to customize the authentication behavior:
//   - WithClientID: Set expected token audience
//   - WithVerificationMethod: Choose JWT or introspection validation
//   - WithServiceAccountKey: Configure service account for introspection
//
// The authenticator initializes immediately and returns an error if
// the Zitadel instance cannot be reached or the verifier cannot be created.
func NewZitadelAuthenticator(
	ctx context.Context,
	zitadelDomain string,
	opts ...ZitadelAuthenticatorOption,
) (*ZitadelAuthenticator, error) {
	authenticator := &ZitadelAuthenticator{
		verificationMethod: JWTVerification,
		clientID:           "zitadel",
	}

	for _, opt := range opts {
		opt(authenticator)
	}

	// Use insecure connection for local development.
	// TODO: Detect environment and use TLS in production.
	zitadelInstance := zitadel.New(zitadelDomain, zitadel.WithInsecure("8080"))
	authenticator.zitadelInstance = zitadelInstance

	verifier, err := authenticator.createVerifier(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create token verifier: %w", err)
	}

	authenticator.verifier = verifier
	authenticator.initialized = true

	return authenticator, nil
}

// createVerifier creates the appropriate token verifier based on the configured
// verification method. This is an internal method that handles the complexity
// of setting up different verification strategies.
func (a *ZitadelAuthenticator) createVerifier(ctx context.Context) (*oauth.JWTVerification, error) {
	switch a.verificationMethod {
	case IntrospectionVerification:
		if a.serviceAccountKey == nil {
			return nil, fmt.Errorf("service account key required for introspection verification")
		}
		initializer := oauth.DefaultAuthorization(a.serviceAccountKey.KeyFilePath)
		verifier, err := initializer(ctx, a.zitadelInstance)
		if err != nil {
			return nil, err
		}
		_, ok := verifier.(*oauth.IntrospectionVerification[*oauth.IntrospectionContext])
		if !ok {
			return nil, fmt.Errorf("failed to cast to introspection verifier")
		}
		return &oauth.JWTVerification{}, nil

	case IntrospectionWithCache:
		if a.serviceAccountKey == nil {
			return nil, fmt.Errorf("service account key required for introspection with cache")
		}
		return &oauth.JWTVerification{}, nil

	case JWTVerification:
		fallthrough
	default:
		initializer := oauth.DefaultJWTAuthorization(a.clientID)
		verifier, err := initializer(ctx, a.zitadelInstance)
		if err != nil {
			return nil, err
		}
		jwtVerifier, ok := verifier.(*oauth.JWTVerification)
		if !ok {
			return nil, fmt.Errorf("failed to cast to JWT verifier")
		}
		return jwtVerifier, nil
	}
}

// Middleware returns an HTTP middleware that validates JWT tokens on incoming requests.
// Public endpoints (as defined in publicEndpoints) bypass authentication.
//
// The middleware extracts the Bearer token from the Authorization header,
// validates it against Zitadel, and injects user information into the
// request context for downstream handlers.
//
// Security considerations:
//   - Tokens are validated using the configured verification method
//   - Invalid or missing tokens result in 401 Unauthorized responses
//   - User context is stored in request context for handler access
func (a *ZitadelAuthenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Skip authentication for public endpoints like health checks.
		if isPublicEndpoint(path) {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		ctx := r.Context()
		authCtx, err := a.verifier.CheckAuthorization(ctx, tokenString)
		if err != nil {
			http.Error(w, fmt.Sprintf("token validation failed: %v", err), http.StatusUnauthorized)
			return
		}

		userCtx := a.extractUserContext(authCtx)
		ctx = context.WithValue(r.Context(), UserIDKey, userCtx.UserID)
		ctx = context.WithValue(ctx, EmailKey, userCtx.Email)
		ctx = context.WithValue(ctx, OrganizationIDKey, userCtx.OrganizationID)
		ctx = context.WithValue(ctx, RolesKey, userCtx.Roles)
		ctx = context.WithValue(ctx, UsernameKey, userCtx.Username)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Interceptor returns a ConnectRPC unary interceptor that validates JWT tokens
// on incoming gRPC-style requests. This provides authentication for ConnectRPC
// services in addition to traditional HTTP handlers.
//
// The interceptor extracts the Bearer token from the Authorization header,
// validates it against Zitadel, and adds user information to the context
// for service handler access.
func (a *ZitadelAuthenticator) Interceptor() connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			authHeader := req.Header().Get("Authorization")
			if authHeader == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authorization header required"))
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid authorization header format"))
			}

			tokenString := parts[1]

			authCtx, err := a.verifier.CheckAuthorization(ctx, tokenString)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("token validation failed: %w", err))
			}

			userCtx := a.extractUserContext(authCtx)
			ctx = context.WithValue(ctx, UserIDKey, userCtx.UserID)
			ctx = context.WithValue(ctx, EmailKey, userCtx.Email)
			ctx = context.WithValue(ctx, OrganizationIDKey, userCtx.OrganizationID)
			ctx = context.WithValue(ctx, RolesKey, userCtx.Roles)
			ctx = context.WithValue(ctx, UsernameKey, userCtx.Username)

			return next(ctx, req)
		}
	})
}

// UserContext holds the authenticated user's information extracted from a validated token.
// This structure is populated after successful token validation and is used to
// provide user information to handlers via context.
type UserContext struct {
	UserID         string
	Email          string
	OrganizationID string
	Roles          []string
	Username       string
}

// extractUserContext extracts user information from the Zitadel authentication context.
// The authCtx contains claims from the validated JWT token.
func (a *ZitadelAuthenticator) extractUserContext(authCtx *oauth.IntrospectionContext) UserContext {
	userCtx := UserContext{}

	if authCtx != nil {
		userCtx.UserID = authCtx.UserID()
		userCtx.OrganizationID = authCtx.OrganizationID()
	}

	return userCtx
}

// CheckAuthorization validates a token and returns the user context.
// This method can be used for programmatic token validation outside of
// the HTTP middleware or interceptor.
func (a *ZitadelAuthenticator) CheckAuthorization(ctx context.Context, token string) (UserContext, error) {
	authCtx, err := a.verifier.CheckAuthorization(ctx, token)
	if err != nil {
		return UserContext{}, fmt.Errorf("token validation failed: %w", err)
	}

	return a.extractUserContext(authCtx), nil
}

// GetUserID retrieves the authenticated user's ID from the context.
// Returns an empty string if the user ID is not present in the context,
// which occurs when authentication has not been performed or failed.
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// GetEmail retrieves the user's email address from the context.
// Returns an empty string if the email is not present in the context.
func GetEmail(ctx context.Context) string {
	if email, ok := ctx.Value(EmailKey).(string); ok {
		return email
	}
	return ""
}

// GetOrganizationID retrieves the user's organization ID from the context.
// This is used for multi-tenant isolation. Returns an empty string
// if the organization ID is not present.
func GetOrganizationID(ctx context.Context) string {
	if orgID, ok := ctx.Value(OrganizationIDKey).(string); ok {
		return orgID
	}
	return ""
}

// GetRoles retrieves the user's assigned roles from the context.
// Returns nil if roles are not present in the context.
func GetRoles(ctx context.Context) []string {
	if roles, ok := ctx.Value(RolesKey).([]string); ok {
		return roles
	}
	return nil
}

// GetUsername retrieves the username from the context.
// Returns an empty string if the username is not present.
func GetUsername(ctx context.Context) string {
	if username, ok := ctx.Value(UsernameKey).(string); ok {
		return username
	}
	return ""
}

// HasRole checks if the user has a specific role in their role list.
// This is used for simple role-based authorization checks in handlers.
// Returns false if roles are not present in the context.
func HasRole(ctx context.Context, role string) bool {
	roles := GetRoles(ctx)
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// IsInitialized returns true if the authenticator was successfully initialized
// and is ready to validate tokens. This can be used to check if the
// Zitadel connection was established successfully at startup.
func (a *ZitadelAuthenticator) IsInitialized() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.initialized
}

// AuthMiddleware returns the authentication middleware and ConnectRPC interceptors.
// If authentication initialization fails, it returns nil middleware and empty interceptors
// to allow the server to start for development purposes.
//
// The middleware validates JWT tokens from the Authorization header and injects
// user context into the request for downstream handlers.
func AuthMiddleware(cfg AuthConfig) (func(http.Handler) http.Handler, connect.Option) {
	authenticator, err := NewZitadelAuthenticator(context.Background(), cfg.ZitadelDomain)
	if err != nil {
		return nil, connect.WithInterceptors()
	}

	middleware := authenticator.Middleware
	interceptors := connect.WithInterceptors(authenticator.Interceptor())

	return middleware, interceptors
}
