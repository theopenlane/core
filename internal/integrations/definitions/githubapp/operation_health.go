package githubapp

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck validates the GitHub App installation token
type HealthCheck struct{}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		githubClient, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return h.Run(ctx, githubClient)
	}
}

// Run executes the health check using the GitHub GraphQL client
func (HealthCheck) Run(ctx context.Context, client GraphQLClient) (json.RawMessage, error) {
	_, err := queryRepositories(ctx, client, 1)
	if err != nil {
		return nil, err
	}

	return providerkit.EncodeResult(map[string]any{}, ErrResultEncode)
}
