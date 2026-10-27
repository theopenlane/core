package authentik

import (
	"context"
	"time"

	"github.com/theopenlane/httpsling/httpclient"
	authentikSDK "goauthentik.io/api/v3"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

const (
	// authentikRequestTimeout is the per-request timeout for Authentik API calls
	authentikRequestTimeout = 30 * time.Second
)

// clientBuilder returns the Authentik client constructor, allowPrivate permits non-public base urls for local development
func clientBuilder(allowPrivate bool) func(context.Context, types.ConnectionRequest[CredentialSchema]) (*authentikSDK.APIClient, error) {
	return func(_ context.Context, req types.ConnectionRequest[CredentialSchema]) (*authentikSDK.APIClient, error) {
		return buildClient(req, allowPrivate)
	}
}

// buildClient constructs the Authentik API client for one installation
func buildClient(req types.ConnectionRequest[CredentialSchema], allowPrivate bool) (*authentikSDK.APIClient, error) {
	cred := req.Credential

	switch {
	case cred.Token == "":
		return nil, ErrAPITokenMissing
	case cred.BaseURL == "":
		return nil, ErrBaseURLMissing
	}

	baseURL, err := urlx.Parse(cred.BaseURL)
	if err != nil {
		return nil, err
	}

	httpClient, err := urlx.NewCustomerHTTPClient(baseURL.String(), allowPrivate, httpclient.Timeout(authentikRequestTimeout))
	if err != nil {
		return nil, err
	}

	cfg := authentikSDK.NewConfiguration()
	cfg.Servers = authentikSDK.ServerConfigurations{{URL: baseURL.JoinPath("api", "v3").String()}}
	cfg.HTTPClient = httpClient
	cfg.AddDefaultHeader("Authorization", "Bearer "+cred.Token)

	return authentikSDK.NewAPIClient(cfg), nil
}
