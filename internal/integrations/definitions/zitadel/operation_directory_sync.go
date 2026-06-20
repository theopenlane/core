package zitadel

import (
	"context"

	objectv2 "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/object/v2"
	userv2 "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/user/v2"
	zitadelUser "github.com/zitadel/zitadel-go/v3/pkg/client/user/v2"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// DirectorySync collects Zitadel directory users for ingest
type DirectorySync struct{}

// IngestHandle adapts directory sync to the ingest operation registration boundary
func (d DirectorySync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(zitadelClient, func(ctx context.Context, request types.OperationRequest, c *zitadelUser.Client) ([]types.IngestPayloadSet, error) {
		var cfg UserInput
		if request.Integration != nil {
			_ = jsonx.UnmarshalIfPresent(request.Integration.Config.ClientConfig, &cfg)
		}

		return d.Run(ctx, c, cfg)
	})
}

// Run collects Zitadel directory users
func (DirectorySync) Run(ctx context.Context, c *zitadelUser.Client, cfg UserInput) ([]types.IngestPayloadSet, error) {
	users, err := listDirectoryUsers(ctx, c)
	if err != nil {
		return nil, err
	}

	accountEnvelopes := make([]types.MappingEnvelope, 0, len(users))

	for _, user := range users {
		if user.GetUserId() == "" {
			continue
		}

		resourceID := user.GetUserId()

		envelope, err := providerkit.MarshalEnvelope(resourceID, user, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		accountEnvelopes = append(accountEnvelopes, envelope)
	}

	return []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			Envelopes: accountEnvelopes,
		},
	}, nil
}

// listDirectoryUsers pages through all Zitadel users using offset-based pagination
func listDirectoryUsers(ctx context.Context, c *zitadelUser.Client) ([]*userv2.User, error) {
	users := make([]*userv2.User, 0)
	var offset uint64 = 0

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		resp, err := c.ListUsers(ctx, &userv2.ListUsersRequest{
			Query: &objectv2.ListQuery{
				Limit:  zitadelDefaultPageSize,
				Offset: offset,
			},
		})
		if err != nil {
			return nil, ErrDirectoryUsersFetchFailed
		}

		users = append(users, resp.Result...)

		if uint64(len(resp.Result)) < uint64(zitadelDefaultPageSize) {
			break
		}

		offset += uint64(zitadelDefaultPageSize)
	}

	return users, nil
}