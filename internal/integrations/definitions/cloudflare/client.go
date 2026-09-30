package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	cf "github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/option"
	"github.com/theopenlane/httpsling/httpclient"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/domainscan"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

// cloudflareRequestTimeout bounds every Cloudflare API request issued through the SDK client
const cloudflareRequestTimeout = time.Minute

// CloudflareClient wraps the Cloudflare SDK client with the account it's scoped to
type CloudflareClient struct { //nolint:revive
	*cf.Client
	// Config holds the account scope and domain scan settings this client was built with
	Config ClientConfig
}

// ClientConfig holds the account and domain scan settings a CloudflareClient is built from
type ClientConfig struct {
	// AccountID is the Cloudflare account this client is scoped to
	AccountID string
	// APIToken is the raw Cloudflare API token, needed by calls made outside the SDK client
	APIToken string
	// DomainScan configures vendor/technology classification for domain scan reports; runtime only
	DomainScan domainscan.ReportConfig
}

// Client builds Cloudflare API clients for one installation
type Client struct{}

// Build constructs the Cloudflare API client for one installation
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (*CloudflareClient, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		return nil, err
	}

	if cred.APIToken == "" {
		return nil, ErrAPITokenMissing
	}

	httpClient, err := urlx.NewHTTPClient(httpclient.Timeout(cloudflareRequestTimeout))
	if err != nil {
		return nil, err
	}

	return &CloudflareClient{
		Client: cf.NewClient(
			option.WithAPIToken(cred.APIToken),
			option.WithHTTPClient(httpClient),
		),
		Config: ClientConfig{
			AccountID: cred.AccountID,
			APIToken:  cred.APIToken,
		},
	}, nil
}

// runtimeCloudflareClientBuilder returns a build function for the operator-owned runtime client
func runtimeCloudflareClientBuilder() func(context.Context, json.RawMessage) (any, error) {
	return func(_ context.Context, config json.RawMessage) (any, error) {
		var cfg RuntimeConfig
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrRuntimeConfigDecode, err)
		}

		if !cfg.Provisioned() {
			return nil, ErrRuntimeConfigInvalid
		}

		httpClient, err := urlx.NewHTTPClient(httpclient.Timeout(cloudflareRequestTimeout))
		if err != nil {
			return nil, err
		}

		return &CloudflareClient{
			Client: cf.NewClient(
				option.WithAPIToken(cfg.APIToken),
				option.WithHTTPClient(httpClient),
			),
			Config: ClientConfig{
				AccountID:  cfg.AccountID,
				APIToken:   cfg.APIToken,
				DomainScan: cfg.DomainScan,
			},
		}, nil
	}
}

func resolveCredential(bindings types.CredentialBindings) (CredentialSchema, error) {
	cred, ok, err := cloudflareCredential.Resolve(bindings)
	if err != nil {
		return CredentialSchema{}, ErrCredentialInvalid
	}

	if !ok {
		return CredentialSchema{}, ErrCredentialMetadataRequired
	}

	return cred, nil
}
