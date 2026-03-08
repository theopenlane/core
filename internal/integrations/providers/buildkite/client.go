package buildkite

import (
	"context"
	"encoding/json"

	buildkitego "github.com/buildkite/go-buildkite/v3/buildkite"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// ClientBuildkiteAPI identifies the Buildkite HTTP API client.
	ClientBuildkiteAPI types.ClientName = "api"
)

// buildkiteClientDescriptors returns the client descriptors published by Buildkite.
func buildkiteClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeBuildkite, ClientBuildkiteAPI, "Buildkite REST API client", buildBuildkiteClient)
}

// buildBuildkiteClient constructs a Buildkite SDK client from credential payload.
func buildBuildkiteClient(_ context.Context, payload models.CredentialSet, _ json.RawMessage) (types.ClientInstance, error) {
	token, err := auth.APITokenFromPayload(payload)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	client, err := buildkitego.NewOpts(buildkitego.WithTokenAuth(token))
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	return types.NewClientInstance(client), nil
}
