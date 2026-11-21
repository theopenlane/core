package aikido

import (
	"context"
	"fmt"
	"strconv"

	"github.com/theopenlane/logx"
	"github.com/theopenlane/utils/mapx"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

const (
	// teamMemberCountKey is the team payload key the number of members is attached under
	teamMemberCountKey = "member_count"
	// membershipTeamIDKey is the membership payload key carrying the team id
	membershipTeamIDKey = "team_id"
	// membershipUserIDKey is the membership payload key carrying the user id
	membershipUserIDKey = "user_id"
)

// runDirectorySync collects Aikido users, and unless group sync is disabled their teams and team memberships
func runDirectorySync(ctx context.Context, _ types.OperationRequest, client *APIClient, cfg providerkit.DirectorySync) ([]types.IngestPayloadSet, error) {
	return CollectDirectory(ctx, client, cfg.DisableGroupSync)
}

// CollectDirectory lists workspace users as directory accounts and, unless skipGroups, teams as groups with their memberships
func CollectDirectory(ctx context.Context, c *APIClient, skipGroups bool) ([]types.IngestPayloadSet, error) {
	users, err := c.users(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDirectoryFetchFailed, err)
	}

	accounts := make([]types.MappingEnvelope, 0, len(users))

	for _, user := range users {
		envelope, err := recordEnvelope(ctx, recordID(user), user)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, envelope)
	}

	payloadSets := providerkit.DirectoryAccountPayloadSets(accounts)

	if skipGroups {
		return payloadSets, nil
	}

	teams, err := c.listAll(ctx, pathTeams, nil, teamPageSize)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDirectoryFetchFailed, err)
	}

	groups := make([]types.MappingEnvelope, 0, len(teams))
	memberships := make([]types.MappingEnvelope, 0)

	for _, team := range teams {
		teamID := recordID(team)

		members, err := c.users(ctx, map[string]string{"filter_team_id": teamID})
		if err != nil {
			return nil, fmt.Errorf("%w: team %s: %w", ErrDirectoryFetchFailed, teamID, err)
		}

		team[teamMemberCountKey] = len(members)

		envelope, err := recordEnvelope(ctx, teamID, team)
		if err != nil {
			return nil, err
		}

		groups = append(groups, envelope)

		for _, member := range members {
			userID := recordID(member)

			envelope, err := recordEnvelope(ctx, teamID+":"+userID, map[string]any{membershipTeamIDKey: teamID, membershipUserIDKey: userID})
			if err != nil {
				return nil, err
			}

			memberships = append(memberships, envelope)
		}
	}

	return append(payloadSets, providerkit.DirectoryGroupPayloadSets(groups, memberships, true)...), nil
}

// recordEnvelope marshals one Aikido record into an envelope named by resource
func recordEnvelope(ctx context.Context, resource string, record map[string]any) (types.MappingEnvelope, error) {
	envelope, err := providerkit.MarshalEnvelope(resource, mapx.PruneMapZeroAny(record), ErrPayloadEncode)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("resource", resource).Msg("aikido: failed to marshal record envelope")
	}

	return envelope, err
}

// recordID renders a record's numeric Aikido id as the string external ids are matched on
func recordID(record map[string]any) string {
	return strconv.FormatFloat(mapx.Value[float64](record, "id"), 'f', -1, 64)
}
