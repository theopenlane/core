package okta

import (
	"context"

	oktagosdk "github.com/okta/okta-sdk-golang/v6/okta"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

const (
	// oktaRateLimitMaxRetries is the maximum number of retries when an Okta API call is rate-limited
	oktaRateLimitMaxRetries = 3
	// oktaRequestTimeout is the per-request timeout in seconds for Okta API calls
	oktaRequestTimeout = 30
)

// buildClient constructs the Okta API client for one installation
func buildClient(_ context.Context, req types.ConnectionRequest[CredentialSchema]) (*oktagosdk.APIClient, error) {
	cred := req.Credential

	switch {
	case cred.APIToken == "":
		return nil, ErrAPITokenMissing
	case cred.OrgURL == "":
		return nil, ErrOrgURLMissing
	}

	cfg, err := oktagosdk.NewConfiguration(oktagosdk.WithOrgUrl(cred.OrgURL), oktagosdk.WithToken(cred.APIToken), oktagosdk.WithRateLimitMaxRetries(oktaRateLimitMaxRetries), oktagosdk.WithRequestTimeout(oktaRequestTimeout))
	if err != nil {
		return nil, ErrClientConfigInvalid
	}

	return oktagosdk.NewAPIClient(cfg), nil
}
