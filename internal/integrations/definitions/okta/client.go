package okta

import (
	"context"

	oktagosdk "github.com/okta/okta-sdk-golang/v5/okta"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// Client builds Okta API clients for one installation
type Client struct{}

// Build constructs the Okta API client for one installation
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	var cred CredentialSchema
	if err := jsonx.UnmarshalIfPresent(req.Credential.ProviderData, &cred); err != nil {
		return nil, ErrCredentialInvalid
	}

	if cred.APIToken == "" {
		return nil, ErrAPITokenMissing
	}

	if cred.OrgURL == "" {
		return nil, ErrOrgURLMissing
	}

	cfg, err := oktagosdk.NewConfiguration(
		oktagosdk.WithOrgUrl(cred.OrgURL),
		oktagosdk.WithToken(cred.APIToken),
		oktagosdk.WithRateLimitMaxRetries(3),
		oktagosdk.WithRequestTimeout(30),
	)
	if err != nil {
		return nil, ErrClientConfigInvalid
	}

	return oktagosdk.NewAPIClient(cfg), nil
}

