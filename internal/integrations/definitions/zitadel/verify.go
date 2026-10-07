package zitadel

import (
	"context"

	"github.com/zitadel/zitadel-go/v3/pkg/client"
	instancev2 "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/instance/v2"
	objectv2 "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/object/v2"
	userv2 "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/user/v2"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// verify probes the instance with a user listing and returns the installation identity
func verify[T any](ctx context.Context, _ types.ConnectionRequest[T], c Client) (InstallationMetadata, error) {
	_, err := c.UserServiceV2().ListUsers(ctx, &userv2.ListUsersRequest{
		Query: &objectv2.ListQuery{
			Limit: 1,
		},
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error listing users for health check")
		return InstallationMetadata{}, ErrHealthCheckFailed
	}

	instanceID, err := resolveInstanceID(ctx, c.Client)
	if err != nil {
		return InstallationMetadata{}, err
	}

	return InstallationMetadata{Domain: c.Domain, InstanceID: instanceID}, nil
}

// resolveInstanceID fetches the immutable id of the Zitadel instance the credential is scoped to
func resolveInstanceID(ctx context.Context, c *client.Client) (string, error) {
	resp, err := c.InstanceServiceV2().GetInstance(ctx, &instancev2.GetInstanceRequest{})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error fetching zitadel instance")
		return "", ErrInstanceFetchFailed
	}

	return resp.GetInstance().GetId(), nil
}
