package cloudflare

import (
	"context"
	"encoding/json"

	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientCloudflareAPI identifies the Cloudflare HTTP API client.
	ClientCloudflareAPI types.ClientName = "api"
)

// cloudflareClientDescriptors returns the client descriptors published by Cloudflare.
func cloudflareClientDescriptors() []types.ClientDescriptor {
	return auth.DefaultClientDescriptors(TypeCloudflare, ClientCloudflareAPI, "Cloudflare REST API client", buildCloudflareClient)
}

// buildCloudflareClient constructs a Cloudflare SDK client from credential payload.
func buildCloudflareClient(_ context.Context, payload types.CredentialPayload, _ json.RawMessage) (types.ClientInstance, error) {
	token, err := auth.APITokenFromPayload(payload)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	return types.NewClientInstance(cf.NewClient(option.WithAPIToken(token))), nil
}
