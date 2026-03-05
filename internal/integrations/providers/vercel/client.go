package vercel

import (
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// ClientVercelAPI identifies the Vercel HTTP API client.
	ClientVercelAPI types.ClientName = "api"
)

// vercelClientDescriptors returns the client descriptors published by Vercel.
func vercelClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeVercel, ClientVercelAPI, "Vercel REST API client", providerkit.TokenClientBuilder(auth.APITokenFromPayload, nil))
}
