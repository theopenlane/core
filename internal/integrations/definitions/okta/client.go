package okta

import (
	"context"

	oktagosdk "github.com/okta/okta-sdk-golang/v6/okta"

	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// oktaRateLimitMaxRetries is the maximum number of retries when an Okta API call is rate-limited
	oktaRateLimitMaxRetries = 3
	// oktaRequestTimeout is the per-request timeout in seconds for Okta API calls
	oktaRequestTimeout = 30
)

// Client builds Okta API clients for one installation
type Client struct{}

// Build constructs the Okta API client for one installation
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	cred, _, err := oktaCredential.Resolve(req.Credentials)
	if err != nil {
		return nil, ErrCredentialInvalid
	}

	if cred.APIToken == "" {
		return nil, ErrAPITokenMissing
	}

	if cred.OrgURL == "" {
		return nil, ErrOrgURLMissing
	}

	cfg, err := oktagosdk.NewConfiguration(oktagosdk.WithOrgUrl(cred.OrgURL), oktagosdk.WithToken(cred.APIToken), oktagosdk.WithRateLimitMaxRetries(oktaRateLimitMaxRetries), oktagosdk.WithRequestTimeout(oktaRequestTimeout))
	if err != nil {
		return nil, ErrClientConfigInvalid
	}

	return oktagosdk.NewAPIClient(cfg), nil
}
