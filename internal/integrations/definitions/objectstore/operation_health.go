package objectstore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// Handle adapts the health check to the generic operation registration boundary
func (HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClient(storageClient, func(ctx context.Context, client *Client) (json.RawMessage, error) {
		provider := string(client.Provider.ProviderType())

		if _, err := client.Provider.ListObjects(ctx, "", 1); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("provider", provider).Msg("objectstore: bucket listing failed")

			return nil, fmt.Errorf("%w: %w", ErrHealthCheckFailed, err)
		}

		return providerkit.EncodeResult(HealthCheck{Provider: provider}, ErrResultEncode)
	})
}
