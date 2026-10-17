package microsoftteams

import (
	"context"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/theopenlane/core/v2/internal/integrations/definitions/microsoft"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// verify confirms the Graph profile is reachable and resolves the tenant from the access token claims
func verify(ctx context.Context, req types.ConnectionRequest[teamsCred], c *msgraphsdk.GraphServiceClient) (InstallationMetadata, error) {
	if _, err := c.Me().Get(ctx, nil); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("microsoftteams: profile lookup failed")

		return InstallationMetadata{}, ErrProfileLookupFailed
	}

	tenantID := microsoft.TokenClaim(req.Credential.AccessToken, "tid")
	if tenantID == "" {
		logx.FromContext(ctx).Error().Msg("microsoftteams: tenant claim missing from access token")

		return InstallationMetadata{}, ErrTenantUnresolved
	}

	return InstallationMetadata{TenantID: tenantID}, nil
}
