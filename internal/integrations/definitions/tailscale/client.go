package tailscale

import (
	"context"

	tsclient "github.com/tailscale/tailscale-client-go/v2"

	"github.com/theopenlane/core/internal/integrations/types"
)

// Client builds Tailscale API clients for one installation
type Client struct{}

// Build constructs the Tailscale API client for one installation
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		return nil, err
	}

	if cred.ClientID == "" {
		return nil, ErrClientIDMissing
	}

	if cred.ClientSecret == "" {
		return nil, ErrClientSecretMissing
	}

	httpClient := tsclient.OAuthConfig{
		ClientID:     cred.ClientID,
		ClientSecret: cred.ClientSecret,
	}.HTTPClient()

	return &tsclient.Client{
		Tailnet: "-",
		HTTP:    httpClient,
	}, nil
}

// resolveCredential extracts the Tailscale credential from the binding set
func resolveCredential(bindings types.CredentialBindings) (CredentialSchema, error) {
	cred, ok, err := tailscaleCredential.Resolve(bindings)
	if err != nil {
		return CredentialSchema{}, ErrCredentialInvalid
	}

	if !ok {
		return CredentialSchema{}, ErrCredentialMetadataRequired
	}

	return cred, nil
}
