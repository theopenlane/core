package vercel

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const vercelAPIBaseURL = "https://api.vercel.com"

// Client builds Vercel REST API clients for one installation
type Client struct{}

// Build constructs the Vercel REST API client for one installation
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	var cred credential
	if err := jsonx.UnmarshalIfPresent(req.Credential.ProviderData, &cred); err != nil {
		return nil, err
	}

	if cred.APIToken == "" {
		return nil, ErrAPITokenMissing
	}

	return providerkit.NewAuthenticatedClient(vercelAPIBaseURL, cred.APIToken, nil), nil
}

// FromAny casts a registered client instance to the authenticated HTTP client type
func (Client) FromAny(value any) (*providerkit.AuthenticatedClient, error) {
	c, ok := value.(*providerkit.AuthenticatedClient)
	if !ok {
		return nil, ErrClientType
	}

	return c, nil
}
