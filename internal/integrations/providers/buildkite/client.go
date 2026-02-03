package buildkite

import (
	"context"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientBuildkiteAPI identifies the Buildkite HTTP API client.
	ClientBuildkiteAPI types.ClientName = "api"
)

// buildkiteClientDescriptors returns the client descriptors published by Buildkite.
func buildkiteClientDescriptors() []types.ClientDescriptor {
	return []types.ClientDescriptor{
		{
			Provider:     TypeBuildkite,
			Name:         ClientBuildkiteAPI,
			Description:  "Buildkite REST API client",
			Build:        buildBuildkiteClient,
			ConfigSchema: map[string]any{"type": "object"},
		},
	}
}

// buildBuildkiteClient constructs an authenticated Buildkite API client.
func buildBuildkiteClient(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	token, err := helpers.APITokenFromPayload(payload, string(TypeBuildkite))
	if err != nil {
		return nil, err
	}
	return helpers.NewAuthenticatedClient(token, nil), nil
}
