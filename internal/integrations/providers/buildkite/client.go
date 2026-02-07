package buildkite

import (
	"context"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientBuildkiteAPI identifies the Buildkite HTTP API client.
	ClientBuildkiteAPI types.ClientName = "api"
)

// buildkiteClientDescriptors returns the client descriptors published by Buildkite.
func buildkiteClientDescriptors() []types.ClientDescriptor {
	return auth.DefaultClientDescriptors(TypeBuildkite, ClientBuildkiteAPI, "Buildkite REST API client", buildBuildkiteClient)
}

// buildBuildkiteClient constructs an authenticated Buildkite API client.
func buildBuildkiteClient(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	token, err := auth.APITokenFromPayload(payload, string(TypeBuildkite))
	if err != nil {
		return nil, err
	}
	return auth.NewAuthenticatedClient(token, nil), nil
}
