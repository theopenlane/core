package slack

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
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

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClient(slackClient, h.Run)
}

// Run executes the Slack auth.test health check
func (HealthCheck) Run(ctx context.Context, c *SlackClient) (json.RawMessage, error) {
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
