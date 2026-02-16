package server

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/zitadel/zitadel-go/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/client/management"
	zitadel "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel"
)

func NewZitadelManagementClient(ctx context.Context) (*management.Client, error) {
	issuer := strings.TrimSpace(os.Getenv("ZITADEL_ISSUER"))
	if issuer == "" {
		return nil, fmt.Errorf("ZITADEL_ISSUER is not set")
	}

	api := strings.TrimSpace(os.Getenv("ZITADEL_DOMAIN"))
	if api == "" {
		api = strings.TrimPrefix(strings.TrimPrefix(issuer, "https://"), "http://")
		api = strings.Split(api, "/")[0]
	}

	scopes := []string{client.ScopeZitadelAPI()}

	keyPath := strings.TrimSpace(os.Getenv("ZITADEL_BACKEND_API_KEY"))
	keyID := strings.TrimSpace(os.Getenv("ZITADEL_BACKEND_API_ID"))

	var tokenInit client.TokenSourceInitializer
	if keyPath != "" && keyID != "" {
		tokenInit = client.DefaultServiceUserAuthentication(keyPath, scopes...)
	} else if pat := strings.TrimSpace(os.Getenv("ZITADEL_MANAGEMENT_PAT")); pat != "" {
		tokenInit = client.PAT(pat)
	} else {
		return nil, fmt.Errorf("missing Zitadel credentials: need ZITADEL_BACKEND_API_KEY and ZITADEL_BACKEND_API_ID, or ZITADEL_MANAGEMENT_PAT")
	}

	ts, err := tokenInit(ctx, issuer)
	if err != nil {
		return nil, err
	}

	options := []zitadel.Option{zitadel.WithTokenSource(ts)}
	if strings.HasPrefix(issuer, "http://") || strings.EqualFold(os.Getenv("ZITADEL_INSECURE"), "true") {
		options = append(options, zitadel.WithInsecure())
	}

	return management.NewClient(ctx, issuer, api, scopes, options...)
}

func NewZitadelAdminClient(ctx context.Context) (*management.Client, error) {
	issuer := strings.TrimSpace(os.Getenv("ZITADEL_ISSUER"))
	if issuer == "" {
		return nil, fmt.Errorf("ZITADEL_ISSUER is not set")
	}

	api := strings.TrimSpace(os.Getenv("ZITADEL_DOMAIN"))
	if api == "" {
		api = strings.TrimPrefix(strings.TrimPrefix(issuer, "https://"), "http://")
		api = strings.Split(api, "/")[0]
	}

	scopes := []string{client.ScopeZitadelAPI()}

	pat := strings.TrimSpace(os.Getenv("ZITADEL_MANAGEMENT_PAT"))
	if pat == "" {
		return nil, fmt.Errorf("ZITADEL_MANAGEMENT_PAT is required for admin operations")
	}

	tokenInit := client.PAT(pat)

	ts, err := tokenInit(ctx, issuer)
	if err != nil {
		return nil, err
	}

	options := []zitadel.Option{zitadel.WithTokenSource(ts)}
	if strings.HasPrefix(issuer, "http://") || strings.EqualFold(os.Getenv("ZITADEL_INSECURE"), "true") {
		options = append(options, zitadel.WithInsecure())
	}

	return management.NewClient(ctx, issuer, api, scopes, options...)
}

