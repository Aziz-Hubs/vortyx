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

type AgentTokenClaims struct {
	jwt.RegisteredClaims
	AgentID   string `json:"agent_id"`
	AgentName string `json:"agent_name"`
}

type AgentTokenService struct {
	privateKey *rsa.PrivateKey
	publicKey any
	issuer    string
	audience  string
}

func NewAgentTokenService() (*AgentTokenService, error) {
	keyData := os.Getenv("VORT_AGENT_JWT_PRIVATE_KEY")
	var privateKey *rsa.PrivateKey

	if keyData == "" {
		var err error
		privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key: %w", err)
		}
	} else {
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
		audience:  os.Getenv("VORT_AGENT_JWT_AUDIENCE"),
	}, nil
}

func (s *AgentTokenService) IssueToken(agentID, agentName string, expiry time.Duration) (string, error) {
	now := time.Now()
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

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

func (s *AgentTokenService) ValidateToken(tokenString string) (*AgentTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AgentTokenClaims{}, func(token *jwt.Token) (any, error) {
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

func (s *AgentTokenService) GetPublicKeyPEM() ([]byte, error) {
	pubBytes, err := json.Marshal(&s.privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	return pubBytes, nil
}

func generateTokenID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

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

type AgentTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Type      string    `json:"type"`
}

func FormatTokenResponse(token string, expiresAt time.Time) AgentTokenResponse {
	return AgentTokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		Type:      "Bearer",
	}
}

func TokenFromJSON(jsonStr string) (string, time.Time, error) {
	var resp AgentTokenResponse
	if err := json.NewDecoder(strings.NewReader(jsonStr)).Decode(&resp); err != nil {
		return "", time.Time{}, err
	}
	return resp.Token, resp.ExpiresAt, nil
}
