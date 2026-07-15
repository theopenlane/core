package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	cf "github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/option"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/domainscan"
)

// CloudflareClient wraps the Cloudflare SDK client with the account ID
// it's scoped to: the customer's own account for installation-bound operations, or
// the operator-owned account for system-initiated operations run through the runtime path.
// DomainScan is only populated for the runtime (system) client, used by the domain scan
// enrichment operation
type CloudflareClient struct {
	*cf.Client
	// AccountID is the Cloudflare account this client is scoped to
	AccountID string
	// APIToken is the raw Cloudflare API token, needed by calls made outside the SDK client
	// (e.g. Browser Rendering requests issued directly by the domain scan enrichment operation)
	APIToken string
	// DomainScan configures vendor/technology classification for onboarding domain scan
	// reports; runtime client only
	DomainScan domainscan.ReportConfig
}

// Client builds Cloudflare API clients for one installation
type Client struct{}

// Build constructs the Cloudflare API client for one installation
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		return nil, err
	}

	if cred.APIToken == "" {
		return nil, ErrAPITokenMissing
	}

	return &CloudflareClient{
		Client: cf.NewClient(
			option.WithAPIToken(cred.APIToken),
			option.WithHTTPClient(&http.Client{Timeout: time.Minute}),
		),
		AccountID: cred.AccountID,
		APIToken:  cred.APIToken,
	}, nil
}

// runtimeCloudflareClientBuilder returns a build function that constructs a Cloudflare API
// client for the runtime (system) path, using the operator-owned account's API token and account
// ID plus the vendor/technology classification config used by the domain scan enrichment operation
func runtimeCloudflareClientBuilder(reportConfig domainscan.ReportConfig) func(context.Context, json.RawMessage) (any, error) {
	return func(_ context.Context, config json.RawMessage) (any, error) {
		var cfg RuntimeCloudflareConfig
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrRuntimeConfigDecode, err)
		}

		if !cfg.Provisioned() {
			return nil, ErrRuntimeConfigInvalid
		}

		return &CloudflareClient{
			Client: cf.NewClient(
				option.WithAPIToken(cfg.APIToken),
				option.WithHTTPClient(&http.Client{Timeout: time.Minute}),
			),
			AccountID:  cfg.AccountID,
			APIToken:   cfg.APIToken,
			DomainScan: reportConfig,
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
