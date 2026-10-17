package microsoftteams

import (
	"context"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// listTeams lists the signed-in user's joined teams and their channels for choosing a message destination
func listTeams(ctx context.Context, c *msgraphsdk.GraphServiceClient) ([]operations.Team, error) {
	joined, err := c.Me().JoinedTeams().Get(ctx, nil)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("microsoftteams: joined teams lookup failed")
		return nil, ErrTeamsListFailed
	}

	teams := []operations.Team{}

	for _, joinedTeam := range joined.GetValue() {
		team := operations.Team{
			ID:          lo.FromPtr(joinedTeam.GetId()),
			DisplayName: lo.FromPtr(joinedTeam.GetDisplayName()),
			Channels:    []operations.Channel{},
		}

		channels, err := c.Teams().ByTeamId(team.ID).Channels().Get(ctx, nil)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("team_id", team.ID).Msg("microsoftteams: channels lookup failed")
			return nil, ErrChannelsListFailed
		}

		for _, channel := range channels.GetValue() {
			team.Channels = append(team.Channels, operations.Channel{
				ID:          lo.FromPtr(channel.GetId()),
				DisplayName: lo.FromPtr(channel.GetDisplayName()),
			})
		}

		teams = append(teams, team)
	}

	return teams, nil
}
