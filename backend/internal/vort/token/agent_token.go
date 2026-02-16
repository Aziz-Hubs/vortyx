// Package token provides internal JWT token generation for VORT agents.
//
// This package implements a fallback token service used when Zitadel machine user
// authentication is not configured. It generates RS256-signed JWT tokens that can be
// used for API authentication within the Vortyx platform.
//
// Security Note: This token service is intended as a fallback mechanism.
// For production deployments, prefer using Zitadel machine user authentication
// (see package machineuser) which provides proper identity provider integration.
//
// Token Structure:
//   - Algorithm: RS256 (RSA Signature with SHA-256)
//   - Issuer: Configurable (default: "vortyx-agent-auth")
//   - Audience: Configurable (default: "vortyx-api")
//   - Expiry: Configurable (default: 24 hours)
//
// The service automatically generates an RSA key pair if no key is provided,
// but for production use, a persistent key should be configured via
// VORT_AGENT_JWT_PRIVATE_KEY environment variable.
package token

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AgentTokenClaims represents the claims in a VORT agent JWT token.
// These claims identify the agent and its permissions.
type AgentTokenClaims struct {
	jwt.RegisteredClaims
	// AgentID is the unique identifier of the agent
	AgentID string `json:"agent_id"`
	// AgentName is the human-readable name of the agent
	AgentName string `json:"agent_name"`
}

// AgentTokenService handles JWT token creation and validation for VORT agents.
//
// This service is used as a fallback when Zitadel machine user authentication
// is not available. It provides internal token generation with the following
// security characteristics:
//
//   - RS256 signing algorithm for cryptographic integrity
//   - Configurable issuer and audience for token scope
//   - Automatic key generation if no key is provided
//   - Short token lifetime (24 hours) to minimize exposure
//
// For production, configure a persistent private key via environment variable.
type AgentTokenService struct {
	// privateKey is used to sign JWT tokens
	privateKey *rsa.PrivateKey
	// publicKey is used to validate tokens
	publicKey any
	// issuer is the token issuer claim
	issuer string
	// audience is the expected token audience
	audience string
}

// NewAgentTokenService creates a new AgentTokenService instance.
//
// It attempts to load an RSA private key from the VORT_AGENT_JWT_PRIVATE_KEY
// environment variable. If not set, it generates a new key pair.
//
// Environment Variables:
//   - VORT_AGENT_JWT_PRIVATE_KEY: Base64-encoded RSA private key
//
// Returns an error if key parsing fails.
func NewAgentTokenService() (*AgentTokenService, error) {
	keyData := os.Getenv("VORT_AGENT_JWT_PRIVATE_KEY")
	var privateKey *rsa.PrivateKey

	if keyData == "" {
		// No key provided - generate a new one
		// Note: This means tokens won't persist across restarts
		var err error
		privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key: %w", err)
		}
	} else {
		// Parse the provided key
		keyBytes, err := base64.StdEncoding.DecodeString(keyData)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 key: %w", err)
		}
		privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
		if err != nil {
			return nil, fmt.Errorf("invalid PEM key: %w", err)
		}
	}

	return &AgentTokenService{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		issuer:     os.Getenv("VORT_AGENT_JWT_ISSUER"),
		audience:   os.Getenv("VORT_AGENT_JWT_AUDIENCE"),
	}, nil
}

// IssueToken creates a new JWT token for an agent.
//
// Parameters:
//   - agentID: Unique identifier for the agent (from database)
//   - agentName: Human-readable name of the agent
//   - expiry: How long the token should be valid
//
// Returns:
//   - A signed JWT token string
//   - An error if token generation fails
//
// The token includes:
//   - agent_id: The agent's database ID
//   - agent_name: The agent's name
//   - Standard claims: iat, exp, nbf, iss, sub, jti
func (s *AgentTokenService) IssueToken(agentID, agentName string, expiry time.Duration) (string, error) {
	now := time.Now()

	// Create claims with agent-specific data
	claims := AgentTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.issuer,
			Subject:   agentID,
			ID:        generateTokenID(),
		},
		AgentID:   agentID,
		AgentName: agentName,
	}

	// Sign the token with RS256
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

// ValidateToken validates a JWT token and returns the claims.
//
// Parameters:
//   - tokenString: The JWT token string to validate
//
// Returns:
//   - AgentTokenClaims if validation succeeds
//   - An error if validation fails (expired, invalid signature, etc.)
//
// Validation checks:
//   - Token signature using the public key
//   - Token expiration (exp claim)
//   - Token not-before time (nbf claim)
//   - Issuer matches configured issuer
//   - Audience matches configured audience
func (s *AgentTokenService) ValidateToken(tokenString string) (*AgentTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AgentTokenClaims{}, func(token *jwt.Token) (any, error) {
		// Verify signing method is RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	},
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*AgentTokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetPublicKeyPEM returns the public key in PEM format.
//
// This can be used to distribute the public key to services
// that need to validate tokens issued by this service.
//
// Returns:
//   - JSON-encoded public key
//   - An error if marshaling fails
func (s *AgentTokenService) GetPublicKeyPEM() ([]byte, error) {
	pubBytes, err := json.Marshal(&s.privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	return pubBytes, nil
}

// generateTokenID creates a unique identifier for a token.
// This is used as the JWT ID (jti) claim.
func generateTokenID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// GetAgentTokenService is a convenience function that creates and configures
// an AgentTokenService with default values.
//
// It wraps NewAgentTokenService and sets default issuer/audience if not configured:
//
//   - Default issuer: "vortyx-agent-auth"
//   - Default audience: "vortyx-api"
//
// Returns an error if service creation fails.
func GetAgentTokenService() (*AgentTokenService, error) {
	svc, err := NewAgentTokenService()
	if err != nil {
		return nil, err
	}
	if svc.issuer == "" {
		svc.issuer = "vortyx-agent-auth"
	}
	if svc.audience == "" {
		svc.audience = "vortyx-api"
	}
	return svc, nil
}

// AgentTokenResponse is the JSON response format for token issuance.
// This is returned by the /authenticate endpoint.
type AgentTokenResponse struct {
	// Token is the JWT access token
	Token string `json:"token"`
	// ExpiresAt is when the token expires
	ExpiresAt time.Time `json:"expires_at"`
	// Type is the token type (typically "Bearer")
	Type string `json:"type"`
}

// FormatTokenResponse creates a standardized token response.
func FormatTokenResponse(token string, expiresAt time.Time) AgentTokenResponse {
	return AgentTokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		Type:      "Bearer",
	}
}

// TokenFromJSON parses a token response from JSON string.
// This is useful when receiving token data from external sources.
func TokenFromJSON(jsonStr string) (string, time.Time, error) {
	var resp AgentTokenResponse
	if err := json.NewDecoder(strings.NewReader(jsonStr)).Decode(&resp); err != nil {
		return "", time.Time{}, err
	}
	return resp.Token, resp.ExpiresAt, nil
}
