package slack

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// HealthCheck holds the result of a Slack health check
type HealthCheck struct {
	// Team is the Slack team name
	Team string `json:"team"`
	// URL is the Slack team URL
	URL string `json:"url"`
	// User is the authenticated Slack user
	User string `json:"user"`
}

// checkHealth executes the Slack auth.test health check
func checkHealth(ctx context.Context, _ types.OperationRequest, c *SlackClient) (json.RawMessage, error) {
	resp, err := c.API.AuthTestContext(ctx)
	if err != nil {
		return nil, ErrAuthTestFailed
	}

	return providerkit.EncodeResult(HealthCheck{
		Team: resp.Team,
		URL:  resp.URL,
		User: resp.User,
	}, ErrResultEncode)
}
