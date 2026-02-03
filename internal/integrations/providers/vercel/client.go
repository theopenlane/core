package vercel

import (
	"context"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientVercelAPI identifies the Vercel HTTP API client.
	ClientVercelAPI types.ClientName = "api"
)

// vercelClientDescriptors returns the client descriptors published by Vercel.
func vercelClientDescriptors() []types.ClientDescriptor {
	return []types.ClientDescriptor{
		{
			Provider:     TypeVercel,
			Name:         ClientVercelAPI,
			Description:  "Vercel REST API client",
			Build:        buildVercelClient,
			ConfigSchema: map[string]any{"type": "object"},
		},
	}
}

// buildVercelClient constructs an authenticated Vercel API client.
func buildVercelClient(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	token, err := helpers.APITokenFromPayload(payload, string(TypeVercel))
	if err != nil {
		return nil, err
	}
	return helpers.NewAuthenticatedClient(token, nil), nil
}
