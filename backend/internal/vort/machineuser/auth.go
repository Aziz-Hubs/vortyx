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
)

type MachineUserAuth struct {
	issuer     string
	audience   string
	privateKey *rsa.PrivateKey
	keyID      string
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func NewMachineUserAuth(ctx context.Context) (*MachineUserAuth, error) {
	issuer := strings.TrimSpace(os.Getenv("ZITADEL_ISSUER"))
	if issuer == "" {
		issuer = "http://localhost:8080"
	}

	audience := strings.TrimSpace(os.Getenv("ZITADEL_AUDIENCES"))
	if audience == "" {
		audience = strings.TrimSpace(os.Getenv("ZITADEL_API_PROJECT_ID"))
	}
	if audience == "" {
		audience = "zitadel"
	}

	keyPath := strings.TrimSpace(os.Getenv("VORT_MACHINE_USER_KEY_PATH"))
	var privateKey *rsa.PrivateKey

	if keyPath != "" {
		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read machine user key: %w", err)
		}
		var errParse error
		privateKey, errParse = jwt.ParseRSAPrivateKeyFromPEM(keyData)
		if errParse != nil {
			return nil, fmt.Errorf("failed to parse machine user key: %w", errParse)
		}
	} else {
		keyData := os.Getenv("VORT_MACHINE_USER_KEY")
		if keyData != "" {
			keyBytes, err := base64.StdEncoding.DecodeString(keyData)
			if err != nil {
				return nil, fmt.Errorf("invalid base64 key: %w", err)
			}
			var errParse error
			privateKey, errParse = jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
			if errParse != nil {
				return nil, fmt.Errorf("failed to parse machine user key: %w", errParse)
			}
		}
	}

	keyID := strings.TrimSpace(os.Getenv("VORT_MACHINE_USER_KEY_ID"))
	if privateKey != nil && keyID == "" {
		return nil, fmt.Errorf("VORT_MACHINE_USER_KEY_ID is required when providing a private key")
	}

	if privateKey == nil {
		return nil, fmt.Errorf("machine user authentication requires VORT_MACHINE_USER_KEY_PATH or VORT_MACHINE_USER_KEY")
	}

	auth := &MachineUserAuth{
		issuer:     issuer,
		audience:   audience,
		privateKey: privateKey,
		keyID:      keyID,
	}

	return auth, nil
}

type TokenGenerator func() (string, error)

func (a *MachineUserAuth) createJWTToken() (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    a.issuer,
		Subject:   a.audience,
		Audience:  jwt.ClaimStrings{a.audience},
		ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ID:        a.keyID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = a.keyID

	return token.SignedString(a.privateKey)
}

func (a *MachineUserAuth) IssueToken(userID string) (string, time.Time, error) {
	assertion, err := a.createJWTToken()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate JWT: %w", err)
	}

	tokenURL := fmt.Sprintf("%s/oauth/v2/token", a.issuer)

	req, err := http.NewRequest("POST", tokenURL, nil)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	q.Set("assertion", assertion)
	q.Set("scope", "urn:zitadel:iam:org:project:id:zitadel:aud")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("token request returned status %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to decode token response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return tokenResp.AccessToken, expiresAt, nil
}

func GetMachineUserAuth(ctx context.Context) (*MachineUserAuth, error) {
	auth, err := NewMachineUserAuth(ctx)
	if err != nil {
		return nil, err
	}
	return auth, nil
}
