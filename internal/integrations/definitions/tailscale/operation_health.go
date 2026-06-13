package tailscale

import (
	"context"
	"encoding/json"

	tsclient "github.com/tailscale/tailscale-client-go/v2"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// HealthCheck holds the result of a Tailscale API health check
type HealthCheck struct {
	// UserCount is the number of users visible to the credential
	UserCount int `json:"userCount,omitempty"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClientRequest(tailscaleClient, func(ctx context.Context, _ types.OperationRequest, client *tsclient.Client) (json.RawMessage, error) {
		return h.Run(ctx, client)
	})
}

// Run validates Tailscale API access by listing users
func (HealthCheck) Run(ctx context.Context, client *tsclient.Client) (json.RawMessage, error) {
	users, err := client.Users().List(ctx, nil, nil)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("tailscale: health check failed listing users")
		return nil, ErrHealthCheckFailed
	}

	details := HealthCheck{
		UserCount: len(users),
	}

	return providerkit.EncodeResult(details, ErrResultEncode)
}
