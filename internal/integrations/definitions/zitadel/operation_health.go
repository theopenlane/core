package zitadel

import (
	"context"
	"encoding/json"

	"github.com/zitadel/zitadel-go/v3/pkg/client"
	objectv2 "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/object/v2"
	userv2 "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/user/v2"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// HealthCheck holds the result of a Zitadel health check
type HealthCheck struct {
	// Domain is the Zitadel instance domain that was checked
	Domain string `json:"domain"`
	// UserCount is the total number of users in the instance
	UserCount uint64 `json:"userCount"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClientRequest(zitadelClient, func(ctx context.Context, req types.OperationRequest, c *client.Client) (json.RawMessage, error) {
		return h.Run(ctx, c, req)
	})
}

// Run executes the Zitadel health check
func (HealthCheck) Run(ctx context.Context, c *client.Client, req types.OperationRequest) (json.RawMessage, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error resolving credentials")
		return nil, ErrHealthCheckFailed
	}

	resp, err := c.UserServiceV2().ListUsers(ctx, &userv2.ListUsersRequest{
	Query: &objectv2.ListQuery{
		Limit: 1,
	},
})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error listing users for health check")
		return nil, ErrHealthCheckFailed
	}

	return providerkit.EncodeResult(HealthCheck{
		Domain:    cred.Domain,
		UserCount: resp.GetDetails().GetTotalResult(),
	}, ErrResultEncode)
}