package onedrive

import (
	"context"

	"github.com/golang-jwt/jwt/v5"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/ssoutils"
)

// verify confirms the user's drive is reachable and resolves the tenant from the access token claims
func verify(ctx context.Context, req types.ConnectionRequest[oneDriveCred], c *DriveClient) (InstallationMetadata, error) {
	if _, err := c.Graph.Me().Drive().Get(ctx, nil); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("onedrive: drive lookup failed")

		return InstallationMetadata{}, ErrHealthCheckFailed
	}

	token, _, err := new(jwt.Parser).ParseUnverified(req.Credential.AccessToken, jwt.MapClaims{})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("onedrive: error parsing JWT")

		return InstallationMetadata{}, ErrTenantUnresolved
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return InstallationMetadata{}, ErrTenantUnresolved
	}

	tenantID, ok := claims["tid"].(string)
	if !ok || tenantID == "" {
		return InstallationMetadata{}, ErrTenantUnresolved
	}

	upn, ok := claims["upn"].(string)
	if !ok || upn == "" {
		upn, _ = claims["preferred_username"].(string)
	}

	return InstallationMetadata{
		TenantID: tenantID,
		Domain:   ssoutils.EmailDomain(upn),
	}, nil
}
