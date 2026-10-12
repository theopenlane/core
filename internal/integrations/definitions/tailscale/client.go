package tailscale

import (
	"context"

	tsclient "github.com/tailscale/tailscale-client-go/v2"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// buildClient constructs the Tailscale API client for one installation
func buildClient(_ context.Context, req types.ConnectionRequest[CredentialSchema]) (*tsclient.Client, error) {
	cred := req.Credential

	switch {
	case cred.ClientID == "":
		return nil, ErrClientIDMissing
	case cred.ClientSecret == "":
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
