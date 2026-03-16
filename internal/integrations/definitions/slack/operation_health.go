package slack

import (
	"context"
	"encoding/json"

	slackgo "github.com/slack-go/slack"

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
func (h HealthCheck) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return h.Run(ctx, c)
	}
}

// Run executes the Slack auth.test health check
func (HealthCheck) Run(ctx context.Context, c *slackgo.Client) (json.RawMessage, error) {
	resp, err := c.AuthTestContext(ctx)
	if err != nil {
		return nil, ErrAuthTestFailed
	}

	return providerkit.EncodeResult(HealthCheck{
		Team: resp.Team,
		URL:  resp.URL,
		User: resp.User,
	}, ErrResultEncode)
}
