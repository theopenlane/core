package authentik

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// directoryDefaultPageSize is the number of records requested per Authentik API page
const directoryDefaultPageSize = 100

// DirectorySync collects Authentik directory users, groups, and memberships for ingest
type DirectorySync struct{}

// IngestHandle adapts directory sync to the ingest operation registration boundary
func (d DirectorySync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(authentikClient, func(ctx context.Context, request types.OperationRequest, c *Client) ([]types.IngestPayloadSet, error) {
		var cfg UserInput
		if request.Integration != nil {
			_ = jsonx.UnmarshalIfPresent(request.Integration.Config.ClientConfig, &cfg)
		}

		return d.Run(ctx, c, cfg)
	})
}

// Run collects Authentik directory users, groups, and memberships
func (DirectorySync) Run(ctx context.Context, c *Client, cfg UserInput) ([]types.IngestPayloadSet, error) {
	users, err := listDirectoryUsers(ctx, c)
	if err != nil {
		return nil, err
	}

	accountEnvelopes := make([]types.MappingEnvelope, 0, len(users))
	includedUsers := make(map[string]struct{}, len(users))

	for _, user := range users {
		resourceID := strconv.Itoa(user.PK)

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
		envelope, err := providerkit.MarshalEnvelope(group.PK, group, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		groupEnvelopes = append(groupEnvelopes, envelope)

		for _, member := range group.MembersObj {
			memberID := strconv.Itoa(member.PK)

			if _, ok := includedUsers[memberID]; !ok {
				continue
			}

			envelope, err := providerkit.MarshalEnvelope(group.PK, member, ErrPayloadEncode)
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
func listDirectoryUsers(ctx context.Context, c *Client) ([]UserResponse, error) {
	users := make([]UserResponse, 0)
	page := 1

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		url := fmt.Sprintf("%s%s?page=%d&page_size=%d", c.BaseURL, authentikUsersEndpoint, page, directoryDefaultPageSize)

		batch, err := fetchPage[UserResponse](ctx, c, url)
		if err != nil {
			return nil, ErrDirectoryUsersFetchFailed
		}

		users = append(users, batch.Results...)

		if batch.Pagination.Next == 0 {
			break
		}

		page++
	}

	return users, nil
}

// listDirectoryGroups pages through all Authentik groups with embedded members
func listDirectoryGroups(ctx context.Context, c *Client) ([]GroupResponse, error) {
	groups := make([]GroupResponse, 0)
	page := 1

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		url := fmt.Sprintf("%s%s?page=%d&page_size=%d&include_users=true", c.BaseURL, authentikGroupsEndpoint, page, directoryDefaultPageSize)

		batch, err := fetchPage[GroupResponse](ctx, c, url)
		if err != nil {
			return nil, ErrDirectoryGroupsFetchFailed
		}

		groups = append(groups, batch.Results...)

		if batch.Pagination.Next == 0 {
			break
		}

		page++
	}

	return groups, nil
}

// fetchPage executes a single paginated GET request and decodes the response
func fetchPage[T any](ctx context.Context, c *Client, url string) (PaginatedResponse[T], error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return PaginatedResponse[T]{}, ErrRequestBuildFailed
	}

	resp, err := c.do(ctx, req)
	if err != nil {
		return PaginatedResponse[T]{}, err
	}

	defer resp.Body.Close()

	var result PaginatedResponse[T]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return PaginatedResponse[T]{}, ErrPayloadEncode
	}

	return result, nil
}
