package slack

import (
	"context"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/logx"
)

// channelsListPageSize is the number of conversations requested per page
const channelsListPageSize = 1000

// listTeams lists the connected workspace's non-archived public and private channels for choosing a message destination
func listTeams(ctx context.Context, c *Client) ([]operations.Team, error) {
	info, err := c.API.GetTeamInfoContext(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("slack: team info lookup failed")
		return nil, ErrTeamInfoFailed
	}

	team := operations.Team{ID: info.ID, DisplayName: info.Name, Channels: []operations.Channel{}}

	params := &slackgo.GetConversationsParameters{
		Types:           []string{"public_channel", "private_channel"},
		ExcludeArchived: true,
		Limit:           channelsListPageSize,
	}

	for {
		channels, cursor, err := c.API.GetConversationsContext(ctx, params)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("slack: conversations lookup failed")
			return nil, ErrChannelsListFailed
		}

		for _, channel := range channels {
			team.Channels = append(team.Channels, operations.Channel{ID: channel.ID, DisplayName: channel.Name})
		}

		if cursor == "" {
			break
		}

		params.Cursor = cursor
	}

	return []operations.Team{team}, nil
}
