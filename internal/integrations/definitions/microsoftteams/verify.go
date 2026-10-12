package microsoftteams

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// verify confirms the Graph profile is reachable and resolves the tenant from the access token claims
func verify(ctx context.Context, req types.ConnectionRequest[teamsCred], c *msgraphsdk.GraphServiceClient) (InstallationMetadata, error) {
	if _, err := c.Me().Get(ctx, nil); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("microsoftteams: profile lookup failed")

		return InstallationMetadata{}, ErrProfileLookupFailed
	}

	token, _, err := new(jwt.Parser).ParseUnverified(req.Credential.AccessToken, jwt.MapClaims{})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("microsoftteams: error parsing JWT")

		return InstallationMetadata{}, ErrTenantUnresolved
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return InstallationMetadata{}, ErrTenantUnresolved
	}

	tenantID, _ := claims["tid"].(string)
	if tenantID == "" {
		return InstallationMetadata{}, ErrTenantUnresolved
	}

	return InstallationMetadata{TenantID: tenantID}, nil
}
