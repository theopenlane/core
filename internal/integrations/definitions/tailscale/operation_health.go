package tailscale

import (
	"context"
	"encoding/json"

	tsclient "github.com/tailscale/tailscale-client-go/v2"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// HealthCheck holds the result of a Tailscale API health check
type HealthCheck struct {
	// UserCount is the number of users visible to the credential
	UserCount int `json:"userCount,omitempty"`
}

// checkHealth validates Tailscale API access by listing users
func checkHealth(ctx context.Context, _ types.OperationRequest, client *tsclient.Client) (json.RawMessage, error) {
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
