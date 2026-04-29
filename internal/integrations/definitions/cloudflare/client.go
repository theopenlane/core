package cloudflare

import (
	"context"

	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"

	"github.com/theopenlane/core/internal/integrations/types"
)

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

	return cf.NewClient(option.WithAPIToken(cred.APIToken)), nil
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
