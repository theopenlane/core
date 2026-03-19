package slack

import (
	"context"
	"encoding/json"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TeamInspect collects Slack workspace metadata via team.info
type TeamInspect struct {
	// TeamID is the Slack team identifier
	TeamID string `json:"teamId"`
	// Name is the Slack team display name
	Name string `json:"name"`
	// Domain is the Slack team domain
	Domain string `json:"domain"`
	// EmailDomain is the team's verified email domain
	EmailDomain string `json:"emailDomain"`
}

// Handle adapts team inspect to the generic operation registration boundary
func (t TeamInspect) Handle() types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := SlackClient.Cast(request.Client)
		if err != nil {
			return nil, err
		}

		return t.Run(ctx, c)
	}
}

// Run collects Slack workspace metadata via team.info
func (TeamInspect) Run(ctx context.Context, c *slackgo.Client) (json.RawMessage, error) {
	team, err := c.GetTeamInfoContext(ctx)
	if err != nil {
		return nil, ErrTeamInfoFailed
	}

	return providerkit.EncodeResult(TeamInspect{
		TeamID:      team.ID,
		Name:        team.Name,
		Domain:      team.Domain,
		EmailDomain: team.EmailDomain,
	}, ErrResultEncode)
}
