package buildkite

import (
	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientBuildkiteAPI identifies the Buildkite HTTP API client.
	ClientBuildkiteAPI types.ClientName = "api"
)

// buildkiteClientDescriptors returns the client descriptors published by Buildkite.
func buildkiteClientDescriptors() []types.ClientDescriptor {
	return auth.DefaultClientDescriptors(TypeBuildkite, ClientBuildkiteAPI, "Buildkite REST API client", auth.TokenClientBuilder(auth.APITokenFromPayload, nil))
}
