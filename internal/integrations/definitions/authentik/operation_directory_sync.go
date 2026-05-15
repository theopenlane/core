package authentik

import (
	"context"
	"time"

	authentikSDK "goauthentik.io/api/v3"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// directoryDefaultPageSize is the number of records requested per Authentik API page
const directoryDefaultPageSize = int32(100)

// DirectorySync collects Authentik directory users, groups, and memberships for ingest
type DirectorySync struct{}

// IngestHandle adapts directory sync to the ingest operation registration boundary
func (d DirectorySync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(authentikClient, func(ctx context.Context, request types.OperationRequest, c *authentikSDK.APIClient) ([]types.IngestPayloadSet, error) {
		var cfg UserInput
		if request.Integration != nil {
			_ = jsonx.UnmarshalIfPresent(request.Integration.Config.ClientConfig, &cfg)
		}

		return d.Run(ctx, c, cfg, request.LastRunAt)
	})
}

// Run collects Authentik directory users, groups, and memberships
func (DirectorySync) Run(ctx context.Context, c *authentikSDK.APIClient, cfg UserInput, lastRunAt *time.Time) ([]types.IngestPayloadSet, error) {
	users, err := listDirectoryUsers(ctx, c, lastRunAt)
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

	payloadSets := []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			Envelopes: accountEnvelopes,
		},
	}

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

	payloadSets = append(payloadSets,
		types.IngestPayloadSet{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryGroup,
			Envelopes: groupEnvelopes,
		},
		types.IngestPayloadSet{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryMembership,
			Envelopes: membershipEnvelopes,
		},
	)

	return payloadSets, nil
}

// listDirectoryUsers pages through all Authentik users
func listDirectoryUsers(ctx context.Context, c *authentikSDK.APIClient, lastRunAt *time.Time) ([]authentikSDK.User, error) {
	users := make([]authentikSDK.User, 0)
	page := int32(1)

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		req := c.CoreApi.CoreUsersList(ctx).
			Page(page).
			PageSize(directoryDefaultPageSize)

		if lastRunAt != nil {
			req = req.LastUpdatedGt(*lastRunAt)
		}

		result, resp, err := req.Execute()
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
