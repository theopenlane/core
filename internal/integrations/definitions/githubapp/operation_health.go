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
func (h HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClient(gitHubClient, h.Run)
}

// Run executes the health check using the GitHub GraphQL client
func (HealthCheck) Run(ctx context.Context, client GraphQLClient) (json.RawMessage, error) {
	_, err := queryRepositories(ctx, client, 1)
	if err != nil {
		return nil, err
	}

	return providerkit.EncodeResult(map[string]any{}, ErrResultEncode)
}
