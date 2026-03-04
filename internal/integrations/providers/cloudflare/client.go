package cloudflare

import (
	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientCloudflareAPI identifies the Cloudflare HTTP API client.
	ClientCloudflareAPI types.ClientName = "api"
)

// cloudflareClientDescriptors returns the client descriptors published by Cloudflare.
func cloudflareClientDescriptors() []types.ClientDescriptor {
	return auth.DefaultClientDescriptors(TypeCloudflare, ClientCloudflareAPI, "Cloudflare REST API client", auth.TokenClientBuilder(auth.APITokenFromPayload, map[string]string{"Content-Type": "application/json"}))
}
