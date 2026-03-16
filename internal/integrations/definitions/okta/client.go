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
	var cred credential
	if err := jsonx.UnmarshalIfPresent(req.Credential.ProviderData, &cred); err != nil {
		return nil, err
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
	)
	if err != nil {
		return nil, err
	}

	return oktagosdk.NewAPIClient(cfg), nil
}

// FromAny casts a registered client instance to the Okta API client type
func (Client) FromAny(value any) (*oktagosdk.APIClient, error) {
	c, ok := value.(*oktagosdk.APIClient)
	if !ok {
		return nil, ErrClientType
	}

	return c, nil
}
