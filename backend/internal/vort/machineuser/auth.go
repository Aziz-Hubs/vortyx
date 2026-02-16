// Package machineuser provides Zitadel machine user authentication using JWT Profile Grant.
//
// This package implements the OAuth 2.0 JWT Profile Grant flow (RFC 7523) for
// machine-to-machine (M2M) authentication between VORT agents and Zitadel.
// Instead of using client secrets, agents authenticate using RSA key pairs.
//
// Authentication Flow:
//
//  1. Agent sends its agent key and secret to backend /authenticate endpoint
//     (This validates the agent's credentials in our database)
//
//  2. Backend creates a JWT assertion signed with the machine user's private key
//     - Issuer: Zitadel service account ID
//     - Subject: Zitadel service account ID
//     - Audience: Zitadel token endpoint
//     - Expiry: 5 minutes (short-lived for security)
//
//  3. Backend exchanges the JWT assertion for an access token
//     POST /oauth/v2/token
//     grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer
//     assertion=<jwt_assertion>
//
//  4. Backend returns the access token to the agent
//     The agent then uses this token for subsequent API calls
//
// Security Considerations:
//
//   - JWT assertions are short-lived (5 minutes) to minimize token theft risk
//   - Private keys should be stored securely (never in version control)
//   - The token endpoint validates the JWT signature using Zitadel's JWKS
//   - Access tokens from Zitadel are typically valid for 1 hour
//
// Environment Variables:
//
//   - VORT_MACHINE_USER_KEY_PATH: Path to JSON file containing RSA private key
//   - VORT_MACHINE_USER_KEY: Base64-encoded RSA private key (alternative to PATH)
//   - VORT_MACHINE_USER_KEY_ID: Key ID for JWT header (required)
//   - ZITADEL_ISSUER: Zitadel instance URL (e.g., https://your-org.zitadel.cloud)
//   - ZITADEL_AUDIENCES: Expected audience for tokens
//   - ZITADEL_JWT_SCOPE: OAuth2 scope for JWT Profile Grant (defaults to urn:zitadel:iam:org:project:id:zitadel:aud)
//
// Example Usage:
//
//	auth, err := machineuser.GetMachineUserAuth(context.Background())
//	if err != nil {
//	    return err
//	}
//
//	// Get a Zitadel access token for the agent
//	accessToken, expiresAt, err := auth.IssueToken(agentID)
//	if err != nil {
//	    return err
//	}
//
//	// Use the token for API calls
//	req.Header.Set("Authorization", "Bearer "+accessToken)
package machineuser

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// MachineUserAuth handles JWT Profile Grant authentication with Zitadel.
// It manages the creation of JWT assertions and their exchange for access tokens.
//
// The authenticator requires an RSA private key to sign JWT assertions.
// This key must be registered in Zitadel as a service user key.
type MachineUserAuth struct {
	// issuer is the Zitadel OIDC issuer URL (e.g., https://your-org.zitadel.cloud)
	issuer string
	// audience is the expected audience for JWT tokens (typically the project ID)
	audience string
	// scope is the OAuth2 scope requested during token exchange
	scope string
	// privateKey is the RSA private key used to sign JWT assertions
	privateKey *rsa.PrivateKey
	// keyID is the ID of the key in Zitadel (used in JWT header)
	keyID string
}

// TokenResponse represents the OAuth2 token response from Zitadel.
// This is returned when exchanging a JWT assertion for an access token.
type TokenResponse struct {
	// AccessToken is the JWT token used for API authorization
	AccessToken string `json:"access_token"`
	// TokenType is typically "Bearer"
	TokenType string `json:"token_type"`
	// ExpiresIn is the number of seconds until the token expires
	ExpiresIn int `json:"expires_in"`
}

// NewMachineUserAuth creates a new MachineUserAuth instance.
//
// It reads configuration from environment variables:
//   - VORT_MACHINE_USER_KEY_PATH: Path to JSON key file
//   - VORT_MACHINE_USER_KEY: Base64-encoded private key
//   - VORT_MACHINE_USER_KEY_ID: Key ID (required)
//   - ZITADEL_ISSUER: Zitadel URL (defaults to localhost:8080)
//   - ZITADEL_AUDIENCES or ZITADEL_API_PROJECT_ID: Audience
//   - ZITADEL_JWT_SCOPE: OAuth2 scope (defaults to urn:zitadel:iam:org:project:id:zitadel:aud)
//
// Returns an error if:
//   - Neither KEY_PATH nor KEY is provided
//   - KEY_ID is missing when a key is provided
//   - The key cannot be parsed as a valid RSA private key
func NewMachineUserAuth(ctx context.Context) (*MachineUserAuth, error) {
	// Read Zitadel issuer from environment, default to localhost for development
	issuer := strings.TrimSpace(os.Getenv("ZITADEL_ISSUER"))
	if issuer == "" {
		issuer = "http://localhost:8080"
	}

	// Determine the audience for JWT tokens
	// This is typically the Zitadel project ID that represents Vortyx
	audience := strings.TrimSpace(os.Getenv("ZITADEL_AUDIENCES"))
	if audience == "" {
		audience = strings.TrimSpace(os.Getenv("ZITADEL_API_PROJECT_ID"))
	}
	if audience == "" {
		audience = "zitadel"
	}

	// Determine the scope for JWT Profile Grant
	// This controls what access is requested from Zitadel
	scope := strings.TrimSpace(os.Getenv("ZITADEL_JWT_SCOPE"))
	if scope == "" {
		scope = "urn:zitadel:iam:org:project:id:zitadel:aud"
	}

	// Load the RSA private key from file or environment
	keyPath := strings.TrimSpace(os.Getenv("VORT_MACHINE_USER_KEY_PATH"))
	var privateKey *rsa.PrivateKey

	if keyPath != "" {
		// Read key from file path
		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read machine user key from %s: %w", keyPath, err)
		}
		var errParse error
		privateKey, errParse = jwt.ParseRSAPrivateKeyFromPEM(keyData)
		if errParse != nil {
			return nil, fmt.Errorf("failed to parse private key from file: %w", errParse)
		}
	} else {
		// Try to read key from environment variable (base64 encoded)
		keyData := os.Getenv("VORT_MACHINE_USER_KEY")
		if keyData != "" {
			keyBytes, err := base64.StdEncoding.DecodeString(keyData)
			if err != nil {
				return nil, fmt.Errorf("invalid base64 key: %w", err)
			}
			var errParse error
			privateKey, errParse = jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
			if errParse != nil {
				return nil, fmt.Errorf("failed to parse private key from env: %w", errParse)
			}
		}
	}

	// Key ID is required for JWT header (Zitadel needs this to find the correct public key)
	keyID := strings.TrimSpace(os.Getenv("VORT_MACHINE_USER_KEY_ID"))
	if privateKey != nil && keyID == "" {
		return nil, fmt.Errorf("VORT_MACHINE_USER_KEY_ID is required when providing a private key")
	}

	// Validate that we have a private key
	if privateKey == nil {
		return nil, fmt.Errorf("machine user authentication requires VORT_MACHINE_USER_KEY_PATH or VORT_MACHINE_USER_KEY environment variable")
	}

	auth := &MachineUserAuth{
		issuer:     issuer,
		audience:   audience,
		scope:      scope,
		privateKey: privateKey,
		keyID:      keyID,
	}

	return auth, nil
}

// TokenGenerator is a function type that generates JWT tokens.
// This is used for custom JWT creation when needed.
type TokenGenerator func() (string, error)

// createJWTToken creates a signed JWT assertion for the JWT Profile Grant flow.
//
// The JWT contains:
//   - iss (issuer): The service account ID
//   - sub (subject): The service account ID
//   - aud (audience): The Zitadel token endpoint
//   - iat (issued at): Current timestamp
//   - exp (expiration): 5 minutes from now
//   - jti (JWT ID): Unique identifier (random UUID)
//
// This method is called internally by IssueToken to create the assertion
// that will be exchanged for an access token.
func (a *MachineUserAuth) createJWTToken() (string, error) {
	now := time.Now()

	// Generate a unique JWT ID (jti) to prevent token replay attacks
	// This must be a unique value for each token as per JWT spec
	jti, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT ID: %w", err)
	}

	// Create JWT claims for JWT Profile Grant
	// See RFC 7523 for specification details
	claims := jwt.RegisteredClaims{
		Issuer:    a.issuer,
		Subject:   a.issuer,
		Audience:  jwt.ClaimStrings{a.audience},
		ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ID:        jti.String(),
	}

	// Create and sign the token
	// RS256 is required by Zitadel for JWT Profile Grant
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = a.keyID // Key ID helps Zitadel find the correct public key

	return token.SignedString(a.privateKey)
}

// IssueToken exchanges credentials for a Zitadel access token using JWT Profile Grant.
//
// This method:
//  1. Creates a signed JWT assertion using the machine user's private key
//  2. Sends the assertion to Zitadel's token endpoint
//  3. Returns the access token and expiration time
//
// Parameters:
//   - userID: The agent's unique identifier (used for logging/tracing)
//
// Returns:
//   - accessToken: The JWT token for API authorization
//   - expiresAt: When the token expires
//   - error: If token issuance fails
//
// The returned access token is typically valid for 1 hour (43199 seconds in Zitadel).
// Agents should refresh tokens before expiration to avoid authentication failures.
func (a *MachineUserAuth) IssueToken(userID string) (string, time.Time, error) {
	// Step 1: Create the JWT assertion
	assertion, err := a.createJWTToken()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate JWT assertion: %w", err)
	}

	// Step 2: Exchange JWT for access token
	// This is the token endpoint defined in OAuth 2.0 JWT Profile Grant (RFC 7523)
	tokenURL := fmt.Sprintf("%s/oauth/v2/token", a.issuer)

	// Build the request
	req, err := http.NewRequest("POST", tokenURL, nil)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create token request: %w", err)
	}

	// Set form parameters for JWT Profile Grant
	// grant_type must be exactly: urn:ietf:params:oauth:grant-type:jwt-bearer
	q := req.URL.Query()
	q.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	q.Set("assertion", assertion)
	q.Set("scope", a.scope)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	// Timeout after 30 seconds to avoid hanging
	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for error responses
	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("token request returned status %d", resp.StatusCode)
	}

	// Parse the token response
	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to decode token response: %w", err)
	}

	// Calculate expiration time
	// Zitadel returns expires_in as seconds, so we add that to current time
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return tokenResp.AccessToken, expiresAt, nil
}

// GetMachineUserAuth is a convenience function that creates a MachineUserAuth instance.
// It wraps NewMachineUserAuth and returns the auth instance or an error.
//
// This is the recommended way to obtain a MachineUserAuth instance:
//
//	auth, err := GetMachineUserAuth(context.Background())
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	token, _, err := auth.IssueToken("agent-123")
func GetMachineUserAuth(ctx context.Context) (*MachineUserAuth, error) {
	auth, err := NewMachineUserAuth(ctx)
	if err != nil {
		return nil, err
	}
	return auth, nil
}
