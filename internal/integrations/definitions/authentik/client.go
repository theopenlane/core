package authentik

import (
	"context"
	"time"

	"github.com/theopenlane/httpsling/httpclient"
	authentikSDK "goauthentik.io/api/v3"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/urlx"
)

const (
	// authentikRequestTimeout is the per-request timeout for Authentik API calls
	authentikRequestTimeout = 30 * time.Second
)

// Client builds Authentik API clients for one installation
type Client struct{}

// Build constructs the Authentik API client for one installation
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		return nil, err
	}

	if cred.Token == "" {
		return nil, ErrAPITokenMissing
	}

	if cred.BaseURL == "" {
		return nil, ErrBaseURLMissing
	}

	baseURL, err := urlx.Parse(cred.BaseURL)
	if err != nil {
		return nil, err
	}

	httpClient, err := urlx.NewHTTPClient(httpclient.Timeout(authentikRequestTimeout))
	if err != nil {
		return nil, err
	}

	cfg := authentikSDK.NewConfiguration()
	cfg.Servers = authentikSDK.ServerConfigurations{{URL: baseURL.JoinPath("api", "v3").String()}}
	cfg.HTTPClient = httpClient
	cfg.AddDefaultHeader("Authorization", "Bearer "+cred.Token)

	return authentikSDK.NewAPIClient(cfg), nil
}

// resolveCredential extracts the CredentialSchema from the provided credential bindings
func resolveCredential(bindings types.CredentialBindings) (CredentialSchema, error) {
	cred, ok, err := authentikCredential.Resolve(bindings)
	if err != nil {
		return CredentialSchema{}, ErrCredentialDecode
	}

	if !ok {
		return CredentialSchema{}, ErrCredentialDecode
	}

	return cred, nil
}
