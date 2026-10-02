package zitadel

import (
	"context"

	"github.com/zitadel/zitadel-go/v3/pkg/client"
	objectv2 "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/object/v2"
	userv2 "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/user/v2"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// protoJSON serializes Zitadel protobuf payloads into the JSON shape the mappings expect
var protoJSON = protojson.MarshalOptions{UseProtoNames: true, UseEnumNumbers: true}

// runDirectorySync collects Zitadel directory users
func runDirectorySync(ctx context.Context, _ types.OperationRequest, c *client.Client, _ DirectorySync) ([]types.IngestPayloadSet, error) {
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

	return providerkit.DirectoryAccountPayloadSets(accountEnvelopes), nil
}

// listDirectoryUsers pages through all Zitadel users using offset-based pagination
func listDirectoryUsers(ctx context.Context, c *client.Client) ([]*userv2.User, error) {
	users := make([]*userv2.User, 0)
	var offset uint64

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
