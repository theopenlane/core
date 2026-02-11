package vercel

import (
	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientVercelAPI identifies the Vercel HTTP API client.
	ClientVercelAPI types.ClientName = "api"
)

// vercelClientDescriptors returns the client descriptors published by Vercel.
func vercelClientDescriptors() []types.ClientDescriptor {
	return auth.DefaultClientDescriptors(TypeVercel, ClientVercelAPI, "Vercel REST API client", auth.APITokenClientBuilder(nil))
}
