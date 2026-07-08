package zitadel

import (
	"context"

	"github.com/zitadel/zitadel-go/v3/pkg/client"
	objectv2 "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/object/v2"
	userv2 "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/user/v2"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// protoJSON serializes Zitadel protobuf payloads into the JSON shape the mappings expect:
// proto field names (snake_case) and integer enum values, so nested oneof messages
// (human/machine) and timestamps serialize correctly. Standard encoding/json mangles proto
// messages (oneof wrappers, google.protobuf.Timestamp), so protojson is required here.
var protoJSON = protojson.MarshalOptions{UseProtoNames: true, UseEnumNumbers: true}

// DirectorySync collects Zitadel directory users for ingest
type DirectorySync struct{}

// IngestHandle adapts directory sync to the ingest operation registration boundary
func (d DirectorySync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(zitadelClient, func(ctx context.Context, request types.OperationRequest, c *client.Client) ([]types.IngestPayloadSet, error) {
		var cfg UserInput
		if request.Integration != nil {
			_ = jsonx.UnmarshalIfPresent(request.Integration.Config.ClientConfig, &cfg)
		}

		return d.Run(ctx, c, cfg)
	})
}

// Run collects Zitadel directory users
func (DirectorySync) Run(ctx context.Context, c *client.Client, cfg UserInput) ([]types.IngestPayloadSet, error) {
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

		raw, err := protoJSON.Marshal(user)
		if err != nil {
			return nil, ErrPayloadEncode
		}

		accountEnvelopes = append(accountEnvelopes, providerkit.RawEnvelope(resourceID, raw))
	}

	return []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			Envelopes: accountEnvelopes,
		},
	}, nil
}

// listDirectoryUsers pages through all Zitadel users using offset-based pagination
func listDirectoryUsers(ctx context.Context, c *client.Client) ([]*userv2.User, error) {
	users := make([]*userv2.User, 0)
	var offset uint64 = 0

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		resp, err := c.UserServiceV2().ListUsers(ctx, &userv2.ListUsersRequest{
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