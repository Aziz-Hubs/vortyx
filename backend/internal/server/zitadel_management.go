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

	var tokenInit client.TokenSourceInitializer
	if pat := strings.TrimSpace(os.Getenv("ZITADEL_PAT")); pat != "" {
		tokenInit = client.PAT(pat)
	} else if keyPath := strings.TrimSpace(os.Getenv("ZITADEL_SERVICE_USER_KEY_PATH")); keyPath != "" {
		tokenInit = client.DefaultServiceUserAuthentication(keyPath, scopes...)
	} else {
		return nil, fmt.Errorf("missing Zitadel credentials")
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

