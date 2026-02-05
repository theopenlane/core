package cloudflare

import (
	"context"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientCloudflareAPI identifies the Cloudflare HTTP API client.
	ClientCloudflareAPI types.ClientName = "api"
)

// cloudflareClientDescriptors returns the client descriptors published by Cloudflare.
func cloudflareClientDescriptors() []types.ClientDescriptor {
	return helpers.DefaultClientDescriptors(TypeCloudflare, ClientCloudflareAPI, "Cloudflare REST API client", buildCloudflareClient)
}

// buildCloudflareClient constructs an authenticated Cloudflare API client.
func buildCloudflareClient(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	token, err := helpers.APITokenFromPayload(payload, string(TypeCloudflare))
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	return helpers.NewAuthenticatedClient(token, headers), nil
}
