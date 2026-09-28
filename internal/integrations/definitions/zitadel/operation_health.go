package zitadel

import (
	"context"
	"encoding/json"

	"github.com/zitadel/zitadel-go/v3/pkg/client"
	objectv2 "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/object/v2"
	userv2 "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/user/v2"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// HealthCheck holds the result of a Zitadel health check
type HealthCheck struct {
	// Domain is the Zitadel instance domain that was checked
	Domain string `json:"domain"`
	// UserCount is the total number of users in the instance
	UserCount uint64 `json:"userCount"`
}

// checkHealth executes the Zitadel health check
func checkHealth(ctx context.Context, req types.OperationRequest, c *client.Client) (json.RawMessage, error) {
	domain, ok := resolveDomain(req.Credentials)
	if !ok {
		logx.FromContext(ctx).Error().Msg("missing domain in credentials")
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
		Domain:    domain,
		UserCount: resp.GetDetails().GetTotalResult(),
	}, ErrResultEncode)
}
