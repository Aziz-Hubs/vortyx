// Package interceptors provides authentication interceptors using Zitadel OIDC integration.
//
// This package implements JWT-based authentication and authorization for HTTP handlers
// and ConnectRPC services. It leverages the Zitadel SDK to validate tokens issued
// by a Zitadel identity provider.
package interceptors

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"

	"connectrpc.com/connect"
	"github.com/rs/zerolog/log"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
)

// publicEndpoints defines paths that do not require authentication.
// These endpoints are typically health checks or public API entry points.
// VORT agent endpoints (RegisterAgent, AuthenticateAgent) are public as they
// handle initial agent registration and token issuance via Zitadel machine users.
var publicEndpoints = map[string]bool{
	"/health":           true,
	"/healthz":          true,
	"/ping":             true,
	"/api/v1/ping":      true,
	"/api/vort/v1/register": true,
	"/api/vort/v1/authenticate": true,
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
	jwtVerifiers       []*oauth.JWTVerification
	zitadelInstance    *zitadel.Zitadel
	audiences          []string
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

func WithAudiences(audiences ...string) ZitadelAuthenticatorOption {
	return func(a *ZitadelAuthenticator) {
		filtered := make([]string, 0, len(audiences))
		for _, aud := range audiences {
			aud = strings.TrimSpace(aud)
			if aud == "" {
				continue
			}
			filtered = append(filtered, aud)
		}
		a.audiences = filtered
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
	Audiences    []string
}

// DefaultAuthConfig returns the default authentication configuration.
// It reads the Zitadel domain from the ZITADEL_DOMAIN environment variable,
// defaulting to localhost:8080 for local development.
func DefaultAuthConfig() AuthConfig {
	domain := os.Getenv("ZITADEL_DOMAIN")
	if domain == "" {
		domain = "localhost:8080"
	}

	audiences := splitCSV(os.Getenv("ZITADEL_PROJECT_ID"))
	if len(audiences) == 0 {
		clientID := strings.TrimSpace(os.Getenv("ZITADEL_CLIENT_ID"))
		if clientID != "" {
			audiences = []string{clientID}
		}
	}
	return AuthConfig{
		ZitadelDomain: domain,
		Audiences:    audiences,
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
		audiences:          []string{"zitadel"},
	}

	for _, opt := range opts {
		opt(authenticator)
	}

	zitadelOpts := []zitadel.Option{}
	insecure := strings.EqualFold(os.Getenv("ZITADEL_INSECURE"), "true") ||
		strings.HasPrefix(zitadelDomain, "localhost") ||
		strings.HasPrefix(zitadelDomain, "127.0.0.1")
	if insecure {
		insecurePort := os.Getenv("ZITADEL_INSECURE_PORT")
		if insecurePort == "" {
			insecurePort = "8080"
		}
		zitadelOpts = append(zitadelOpts, zitadel.WithInsecure(insecurePort))
	}

	zitadelInstance := zitadel.New(zitadelDomain, zitadelOpts...)
	authenticator.zitadelInstance = zitadelInstance

	verifiers, err := authenticator.createJWTVerifiers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create token verifier: %w", err)
	}

	authenticator.jwtVerifiers = verifiers
	authenticator.initialized = true

	return authenticator, nil
}

// createVerifier creates the appropriate token verifier based on the configured
// verification method. This is an internal method that handles the complexity
// of setting up different verification strategies.
func (a *ZitadelAuthenticator) createJWTVerifiers(ctx context.Context) ([]*oauth.JWTVerification, error) {
	switch a.verificationMethod {
	case IntrospectionVerification:
		return nil, fmt.Errorf("introspection verification not wired")

	case IntrospectionWithCache:
		return nil, fmt.Errorf("introspection with cache not wired")

	case JWTVerification:
		fallthrough
	default:
		audiences := a.audiences
		if len(audiences) == 0 {
			audiences = []string{a.clientID}
		}

		verifiers := make([]*oauth.JWTVerification, 0, len(audiences))
		for _, aud := range audiences {
			initializer := oauth.DefaultJWTAuthorization(aud)
			verifier, err := initializer(ctx, a.zitadelInstance)
			if err != nil {
				return nil, err
			}
			jwtVerifier, ok := verifier.(*oauth.JWTVerification)
			if !ok {
				return nil, fmt.Errorf("failed to cast to JWT verifier")
			}
			verifiers = append(verifiers, jwtVerifier)
		}
		return verifiers, nil
	}
}

func (a *ZitadelAuthenticator) checkAuthorization(ctx context.Context, token string) (*oauth.IntrospectionContext, error) {
	if len(a.jwtVerifiers) == 0 {
		return nil, fmt.Errorf("authenticator not initialized")
	}

	var lastErr error
	for _, verifier := range a.jwtVerifiers {
		authCtx, err := verifier.CheckAuthorization(ctx, token)
		if err == nil {
			return authCtx, nil
		}
		lastErr = err
	}

	if lastErr == nil {
		lastErr = oauth.ErrInvalidToken
	}
	return nil, lastErr
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
		authCtx, err := a.checkAuthorization(ctx, tokenString)
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

			authCtx, err := a.checkAuthorization(ctx, tokenString)
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
		userCtx.Email = authCtx.Email
		if authCtx.PreferredUsername != "" {
			userCtx.Username = authCtx.PreferredUsername
		} else if authCtx.Username != "" {
			userCtx.Username = authCtx.Username
		}
		userCtx.Roles = extractRolesFromClaims(authCtx.Claims)
	}

	return userCtx
}

// CheckAuthorization validates a token and returns the user context.
// This method can be used for programmatic token validation outside of
// the HTTP middleware or interceptor.
func (a *ZitadelAuthenticator) CheckAuthorization(ctx context.Context, token string) (UserContext, error) {
	authCtx, err := a.checkAuthorization(ctx, token)
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
// If authentication initialization fails, it returns a middleware that blocks all requests.
// This ensures the server does not run in an insecure state.
//
// The middleware validates JWT tokens from the Authorization header and injects
// user context into the request for downstream handlers.
func AuthMiddleware(cfg AuthConfig) (func(http.Handler) http.Handler, connect.Option) {
	options := []ZitadelAuthenticatorOption{}
	if len(cfg.Audiences) > 0 {
		options = append(options, WithAudiences(cfg.Audiences...))
	}
	authenticator, err := NewZitadelAuthenticator(context.Background(), cfg.ZitadelDomain, options...)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize Zitadel authenticator, running with auth disabled (INSECURE)")
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if isPublicEndpoint(r.URL.Path) {
					next.ServeHTTP(w, r)
					return
				}
				http.Error(w, "authentication unavailable", http.StatusServiceUnavailable)
			})
		}, connect.WithInterceptors()
	}

	middleware := authenticator.Middleware
	interceptors := connect.WithInterceptors(authenticator.Interceptor())

	return middleware, interceptors
}

func AuthInterceptor(cfg AuthConfig) (func(http.Handler) http.Handler, connect.Option) {
	return AuthMiddleware(cfg)
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func extractRolesFromClaims(claims map[string]any) []string {
	if len(claims) == 0 {
		return nil
	}

	roleKeys := []string{
		"urn:zitadel:iam:org:project:roles",
		"urn:zitadel:iam:org:projects:roles",
		"urn:iam:org:project:roles",
		"roles",
	}

	set := map[string]struct{}{}
	for _, key := range roleKeys {
		val, ok := claims[key]
		if !ok || val == nil {
			continue
		}
		switch t := val.(type) {
		case map[string]any:
			for k := range t {
				if k == "" {
					continue
				}
				set[k] = struct{}{}
			}
		case []any:
			for _, item := range t {
				if s, ok := item.(string); ok {
					s = strings.TrimSpace(s)
					if s != "" {
						set[s] = struct{}{}
					}
				}
			}
		case []string:
			for _, s := range t {
				s = strings.TrimSpace(s)
				if s != "" {
					set[s] = struct{}{}
				}
			}
		case string:
			s := strings.TrimSpace(t)
			if s != "" {
				set[s] = struct{}{}
			}
		}
	}

	if len(set) == 0 {
		return nil
	}

	roles := make([]string, 0, len(set))
	for r := range set {
		roles = append(roles, r)
	}
	slices.Sort(roles)
	return roles
}
