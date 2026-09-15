package zitadel

import (
	"context"

	"github.com/zitadel/zitadel-go/v3/pkg/client"
	instancev2 "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/instance/v2"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// resolveInstallationMetadata derives Zitadel instance metadata from the persisted credential
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	domain, ok := resolveDomain(req.Credentials)
	if !ok {
		logx.FromContext(ctx).Error().Msg("missing domain in credentials")
		return InstallationMetadata{}, false, ErrDomainMissing
	}

	api, err := Client{}.Build(ctx, types.ClientBuildRequest{Credentials: req.Credentials})
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	instanceID, err := resolveInstanceID(ctx, api.(*client.Client))
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	return InstallationMetadata{
		Domain:     domain,
		InstanceID: instanceID,
	}, true, nil
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
