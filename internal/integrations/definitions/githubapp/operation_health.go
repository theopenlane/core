package githubapp

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// checkHealth executes the health check using the GitHub GraphQL client
func checkHealth(ctx context.Context, _ types.OperationRequest, client GraphQLClient) (json.RawMessage, error) {
	_, err := queryRepositories(ctx, client, 1, nil)
	if err != nil {
		return nil, err
	}

	return providerkit.EncodeResult(map[string]any{}, ErrResultEncode)
}
