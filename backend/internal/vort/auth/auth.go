// =============================================================================
// Package: auth
// File: auth.go
// Purpose: Authentication and authorization for VORT agent system
// Created: 2026-02-15
// =============================================================================
// This package provides authentication and authorization mechanisms for secure
// agent-backend communication. It implements token-based authentication with
// agent key validation and permission management.
// =============================================================================

package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// =============================================================================
// Type: AgentInfo
// Purpose: Minimal agent info for authentication
// =============================================================================
type AgentInfo struct {
	ID          string
	AgentKeyHash string
	Status     string
}

// =============================================================================
// Errors
// Purpose: Authentication error definitions
// =============================================================================
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAgentNotFound    = errors.New("agent not found")
	ErrAgentInactive    = errors.New("agent is not active")
	ErrTokenExpired     = errors.New("token expired")
	ErrUnauthorized    = errors.New("unauthorized")
)

// =============================================================================
// Type: AgentAuthenticator
// Purpose: Main authentication handler for agent authentication
// =============================================================================
// AgentAuthenticator handles agent authentication, token generation, and
// validation. It maintains an in-memory cache of registered agents and tokens.
//
// Thread Safety:
//   - Uses sync.RWMutex for concurrent access
//   - Safe for concurrent use by multiple goroutines
//
// Usage:
//   auth := auth.NewAgentAuthenticator()
//   auth.RegisterAgent(agent)
//   token, _ := auth.Authenticate(ctx, keyHash, secret)
type AgentAuthenticator struct {
	mu           sync.RWMutex
	agents       map[string]*AgentInfo      // Agent ID -> Agent
	tokenCache   map[string]*Token         // Token string -> Token
	keyHashCache map[string]string         // Key hash -> Agent ID
	ttl          time.Duration             // Token TTL
	maxTokens    int                      // Max cached tokens
}

// =============================================================================
// Type: Token
// Purpose: Authentication token representation
// =============================================================================
type Token struct {
	AgentID     string    // Associated agent ID
	ExpiresAt   time.Time // Expiration timestamp
	Permissions []string  // Granted permissions
	CreatedAt   time.Time // Token creation time
}

// =============================================================================
// Type: AuthenticatorOption
// Purpose: Functional option for configuring authenticator
// =============================================================================
type AuthenticatorOption func(*AgentAuthenticator)

// WithTokenTTL sets the token time-to-live duration.
//
// Parameters:
//   - ttl: Token validity duration
//
// Returns:
//   - AuthenticatorOption: Configured option
func WithTokenTTL(ttl time.Duration) AuthenticatorOption {
	return func(a *AgentAuthenticator) {
		a.ttl = ttl
	}
}

// WithMaxTokens sets the maximum number of tokens to cache.
//
// Parameters:
//   - maxTokens: Maximum cached tokens
//
// Returns:
//   - AuthenticatorOption: Configured option
func WithMaxTokens(maxTokens int) AuthenticatorOption {
	return func(a *AgentAuthenticator) {
		a.maxTokens = maxTokens
	}
}

// NewAgentAuthenticator creates a new agent authenticator.
//
// Parameters:
//   - opts: Configuration options
//
// Returns:
//   - *AgentAuthenticator: Configured authenticator
func NewAgentAuthenticator(opts ...AuthenticatorOption) *AgentAuthenticator {
	auth := &AgentAuthenticator{
		agents:     make(map[string]*AgentInfo),
		tokenCache: make(map[string]*Token),
		ttl:        24 * time.Hour, // Default 24-hour token validity
		maxTokens:  10000,
	}

	// Apply options
	for _, opt := range opts {
		opt(auth)
	}

	// Start background cleanup
	go auth.cleanupExpiredTokens()

	return auth
}

// =============================================================================
// Method: RegisterAgent
// Purpose: Register an agent for authentication
// =============================================================================
// RegisterAgent adds an agent to the authenticator's cache.
// Must be called before agent can authenticate.
//
// Parameters:
//   - agent: Agent to register
func (a *AgentAuthenticator) RegisterAgent(agent *AgentInfo) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.agents[agent.ID] = agent
	a.keyHashCache[agent.AgentKeyHash] = agent.ID
}

// =============================================================================
// Method: DeregisterAgent
// Purpose: Remove an agent from authentication system
// =============================================================================
// DeregisterAgent removes an agent and all associated tokens.
// Used when decommissioning an agent.
//
// Parameters:
//   - agentID: ID of agent to remove
func (a *AgentAuthenticator) DeregisterAgent(agentID string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if agent, exists := a.agents[agentID]; exists {
		delete(a.keyHashCache, agent.AgentKeyHash)
		delete(a.agents, agentID)
	}

	// Remove all tokens for this agent
	var tokensToRemove []string
	for token, t := range a.tokenCache {
		if t.AgentID == agentID {
			tokensToRemove = append(tokensToRemove, token)
		}
	}
	for _, token := range tokensToRemove {
		delete(a.tokenCache, token)
	}
}

// =============================================================================
// Method: Authenticate
// Purpose: Authenticate an agent with credentials
// =============================================================================
// Authenticate validates agent credentials and returns an authentication token.
//
// Parameters:
//   - ctx: Context for the operation
//   - keyHash: Hashed agent key
//   - secret: Agent secret for credential validation
//
// Returns:
//   - *Token: Authentication token
//   - error: Authentication error
func (a *AgentAuthenticator) Authenticate(ctx context.Context, keyHash, secret string) (*Token, error) {
	// Look up agent by key hash
	a.mu.RLock()
	agentID, exists := a.keyHashCache[keyHash]
	a.mu.RUnlock()

	if !exists {
		return nil, ErrAgentNotFound
	}

	// Get agent
	a.mu.RLock()
	agentInfo, exists := a.agents[agentID]
	a.mu.RUnlock()

	if !exists {
		return nil, ErrAgentNotFound
	}

	// Check agent status
	if agentInfo.Status != "active" && agentInfo.Status != "pending" {
		return nil, ErrAgentInactive
	}

	// Validate credentials
	expectedHash := HashCredential(agentInfo.AgentKeyHash, secret)
	providedHash := HashCredential(keyHash, secret)

	// Constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(expectedHash), []byte(providedHash)) != 1 {
		return nil, ErrInvalidCredentials
	}

	// Generate token
	token := &Token{
		AgentID:     agentInfo.ID,
		ExpiresAt:   time.Now().Add(a.ttl),
		Permissions: []string{"agent:execute", "agent:upload", "agent:heartbeat"},
		CreatedAt:   time.Now(),
	}

	tokenString := GenerateToken(token)

	// Cache token
	a.mu.Lock()
	a.tokenCache[tokenString] = token

	// Cleanup if over limit
	if len(a.tokenCache) > a.maxTokens {
		a.cleanupOldTokens()
	}
	a.mu.Unlock()

	return token, nil
}

// =============================================================================
// Method: ValidateToken
// Purpose: Validate an authentication token
// =============================================================================
// ValidateToken checks if a token is valid and not expired.
//
// Parameters:
//   - ctx: Context for the operation
//   - tokenString: Token to validate
//
// Returns:
//   - *Token: Valid token
//   - error: Validation error
func (a *AgentAuthenticator) ValidateToken(ctx context.Context, tokenString string) (*Token, error) {
	a.mu.RLock()
	token, exists := a.tokenCache[tokenString]
	a.mu.RUnlock()

	if !exists {
		return nil, ErrUnauthorized
	}

	// Check expiration
	if time.Now().After(token.ExpiresAt) {
		a.mu.Lock()
		delete(a.tokenCache, tokenString)
		a.mu.Unlock()
		return nil, ErrTokenExpired
	}

	return token, nil
}

// =============================================================================
// Method: RevokeToken
// Purpose: Invalidate a token immediately
// =============================================================================
// RevokeToken removes a token from the cache, immediately invalidating it.
//
// Parameters:
//   - tokenString: Token to revoke
func (a *AgentAuthenticator) RevokeToken(tokenString string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.tokenCache, tokenString)
}

// =============================================================================
// Method: HasPermission
// Purpose: Check if token has a specific permission
// =============================================================================
// HasPermission checks if the token grants a specific permission.
// Supports wildcard permissions (e.g., "agent:" matches "agent:execute").
//
// Parameters:
//   - token: Token to check
//   - permission: Permission string
//
// Returns:
//   - bool: True if permission granted
func (a *AgentAuthenticator) HasPermission(token *Token, permission string) bool {
	for _, p := range token.Permissions {
		if p == permission || strings.HasPrefix(p, strings.Split(permission, ":")[0]+":") {
			return true
		}
	}
	return false
}

// =============================================================================
// Method: cleanupExpiredTokens
// Purpose: Background cleanup of expired tokens
// =============================================================================
// cleanupExpiredTokens runs periodically to remove expired tokens from cache.
// Started automatically by NewAgentAuthenticator.
func (a *AgentAuthenticator) cleanupExpiredTokens() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		a.mu.Lock()
		now := time.Now()
		var expiredKeys []string
		for key, token := range a.tokenCache {
			if now.After(token.ExpiresAt) {
				expiredKeys = append(expiredKeys, key)
			}
		}
		for _, key := range expiredKeys {
			delete(a.tokenCache, key)
		}
		a.mu.Unlock()
	}
}

// =============================================================================
// Method: cleanupOldTokens
// Purpose: Remove oldest tokens when cache is full
// =============================================================================
// cleanupOldTokens removes the oldest tokens when exceeding maxTokens limit.
func (a *AgentAuthenticator) cleanupOldTokens() {
	count := len(a.tokenCache) - a.maxTokens + 100
	var oldestKeys []string
	var oldestTime time.Time

	for key, token := range a.tokenCache {
		if len(oldestKeys) < count {
			oldestKeys = append(oldestKeys, key)
			if oldestTime.IsZero() || token.CreatedAt.Before(oldestTime) {
				oldestTime = token.CreatedAt
			}
		}
	}

	for _, key := range oldestKeys {
		delete(a.tokenCache, key)
	}
}

// =============================================================================
// Function: HashCredential
// Purpose: Create secure hash of credentials
// =============================================================================
// HashCredential combines agent key hash and secret for secure storage/comparison.
// Uses SHA-256 for hashing.
//
// Parameters:
//   - keyHash: Agent's key hash
//   - secret: Agent secret
//
// Returns:
//   - string: Hex-encoded hash
func HashCredential(keyHash, secret string) string {
	data := fmt.Sprintf("%s:%s", keyHash, secret)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// =============================================================================
// Function: GenerateToken
// Purpose: Create token string from token data
// =============================================================================
// GenerateToken creates a deterministic token string from token data.
//
// Parameters:
//   - t: Token data
//
// Returns:
//   - string: Token string
func GenerateToken(t *Token) string {
	data := fmt.Sprintf("%s:%d:%d", t.AgentID, t.CreatedAt.Unix(), time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// =============================================================================
// Function: ValidateToken (Standalone)
// Purpose: Token validation without authenticator
// =============================================================================
// Note: Standalone validation not implemented - use AgentAuthenticator
func ValidateToken(tokenString string) (*Token, error) {
	return nil, errors.New("token validation not implemented - use AgentAuthenticator")
}

// =============================================================================
// Type: Permission
// Purpose: Permission constants
// =============================================================================
type Permission string

const (
	PermissionExecute        Permission = "agent:execute"       // Execute commands
	PermissionUpload        Permission = "agent:upload"       // Upload data
	PermissionHeartbeat    Permission = "agent:heartbeat"   // Send heartbeats
	PermissionAdmin        Permission = "admin"             // Full access
)

// =============================================================================
// Type: Authorizer
// Purpose: Permission management
// =============================================================================
// Authorizer manages agent permissions and access control.
type Authorizer struct {
	permissions map[string][]Permission // Agent ID -> Permissions
	mu         sync.RWMutex
}

// NewAuthorizer creates a new authorizer instance.
//
// Returns:
//   - *Authorizer: New authorizer
func NewAuthorizer() *Authorizer {
	return &Authorizer{
		permissions: make(map[string][]Permission),
	}
}

// GrantPermission adds permissions to an agent.
//
// Parameters:
//   - agentID: Agent to modify
//   - perms: Permissions to grant
func (az *Authorizer) GrantPermission(agentID string, perms ...Permission) {
	az.mu.Lock()
	defer az.mu.Unlock()

	az.permissions[agentID] = append(az.permissions[agentID], perms...)
}

// RevokePermission removes a permission from an agent.
//
// Parameters:
//   - agentID: Agent to modify
//   - perm: Permission to revoke
func (az *Authorizer) RevokePermission(agentID string, perm Permission) {
	az.mu.Lock()
	defer az.mu.Unlock()

	var newPerms []Permission
	for _, p := range az.permissions[agentID] {
		if p != perm {
			newPerms = append(newPerms, p)
		}
	}
	az.permissions[agentID] = newPerms
}

// HasPermission checks if an agent has a specific permission.
//
// Parameters:
//   - agentID: Agent to check
//   - perm: Permission to verify
//
// Returns:
//   - bool: True if permitted
func (az *Authorizer) HasPermission(agentID string, perm Permission) bool {
	az.mu.RLock()
	defer az.mu.RUnlock()

	for _, p := range az.permissions[agentID] {
		if p == perm || p == PermissionAdmin {
			return true
		}
	}
	return false
}

// =============================================================================
// Type: Middleware
// Purpose: HTTP middleware for authentication
// =============================================================================
// Middleware is a function that authenticates requests.
// Used for chi HTTP routing.
type Middleware func(ctx context.Context, tokenString string) (context.Context, error)

// AuthMiddleware creates authentication middleware for HTTP handlers.
//
// Parameters:
//   - auth: Authenticator instance
//
// Returns:
//   - Middleware: Configured middleware function
func AuthMiddleware(auth *AgentAuthenticator) Middleware {
	return func(ctx context.Context, tokenString string) (context.Context, error) {
		if tokenString == "" {
			return ctx, ErrUnauthorized
		}

		token, err := auth.ValidateToken(ctx, tokenString)
		if err != nil {
			return ctx, err
		}

		// Add agent ID to context
		return context.WithValue(ctx, "agent_id", token.AgentID), nil
	}
}
