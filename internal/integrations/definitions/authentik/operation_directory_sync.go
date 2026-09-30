package authentik

import (
	"context"

	authentikSDK "goauthentik.io/api/v3"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// directoryDefaultPageSize is the number of records requested per Authentik API page
const directoryDefaultPageSize = int32(100)

// directorySyncOperation is the Authentik directory sync operation
var directorySyncOperation = types.OperationRefOf[DirectorySync]().Ingests(authentikClient, runDirectorySync)

// runDirectorySync collects Authentik directory users, groups, and memberships
func runDirectorySync(ctx context.Context, _ types.OperationRequest, c *authentikSDK.APIClient, cfg DirectorySync) ([]types.IngestPayloadSet, error) {
	users, err := listDirectoryUsers(ctx, c)
	if err != nil {
		return nil, err
	}

	accountEnvelopes := make([]types.MappingEnvelope, 0, len(users))
	includedUsers := make(map[string]struct{}, len(users))

	for _, user := range users {
		resourceID := user.GetUid()

		envelope, err := providerkit.MarshalEnvelope(resourceID, user, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		accountEnvelopes = append(accountEnvelopes, envelope)
		includedUsers[resourceID] = struct{}{}
	}

	payloadSets := providerkit.DirectoryAccountPayloadSets(accountEnvelopes)

	if cfg.DisableGroupSync {
		return payloadSets, nil
	}

	groups, err := listDirectoryGroups(ctx, c)
	if err != nil {
		return nil, err
	}

	groupEnvelopes := make([]types.MappingEnvelope, 0, len(groups))
	membershipEnvelopes := make([]types.MappingEnvelope, 0)

	for _, group := range groups {
		envelope, err := providerkit.MarshalEnvelope(group.GetPk(), group, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		groupEnvelopes = append(groupEnvelopes, envelope)

		for _, member := range group.UsersObj {
			memberID := member.GetUid()

			if _, ok := includedUsers[memberID]; !ok {
				continue
			}

			envelope, err := providerkit.MarshalEnvelope(group.GetPk(), member, ErrPayloadEncode)
			if err != nil {
				return nil, err
			}

			membershipEnvelopes = append(membershipEnvelopes, envelope)
		}
	}

	return append(payloadSets, providerkit.DirectoryGroupPayloadSets(groupEnvelopes, membershipEnvelopes, true)...), nil
}

// listDirectoryUsers pages through all Authentik users
func listDirectoryUsers(ctx context.Context, c *authentikSDK.APIClient) ([]authentikSDK.User, error) {
	users := make([]authentikSDK.User, 0)
	page := int32(1)

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		result, resp, err := c.CoreApi.CoreUsersList(ctx).
			Page(page).
			PageSize(directoryDefaultPageSize).
			Execute()
		if resp != nil {
			_ = resp.Body.Close()
		}

		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error listing users")

			return nil, ErrDirectoryUsersFetchFailed
		}

		users = append(users, result.Results...)

		if result.Pagination.Next == 0 {
			break
		}

		page++
	}

	return users, nil
}

// listDirectoryGroups pages through all Authentik groups with embedded members
func listDirectoryGroups(ctx context.Context, c *authentikSDK.APIClient) ([]authentikSDK.Group, error) {
	groups := make([]authentikSDK.Group, 0)
	page := int32(1)

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		result, resp, err := c.CoreApi.CoreGroupsList(ctx).
			Page(page).
			PageSize(directoryDefaultPageSize).
			IncludeUsers(true).
			Execute()
		if resp != nil {
			_ = resp.Body.Close()
		}

		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error listing groups")

			return nil, ErrDirectoryGroupsFetchFailed
		}

		groups = append(groups, result.Results...)

		if result.Pagination.Next == 0 {
			break
		}

		page++
	}

	return groups, nil
}
