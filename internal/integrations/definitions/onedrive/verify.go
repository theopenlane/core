package onedrive

import (
	"context"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/definitions/microsoft"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/ssoutils"
)

// verify confirms the user's drive is reachable and resolves the tenant from the access token claims
func verify(ctx context.Context, req types.ConnectionRequest[oneDriveCred], c *DriveClient) (InstallationMetadata, error) {
	drive, err := c.Graph.Me().Drive().Get(ctx, nil)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("onedrive: drive lookup failed")

		return InstallationMetadata{}, ErrHealthCheckFailed
	}

	metadata := InstallationMetadata{
		DriveID:   lo.FromPtr(drive.GetId()),
		OwnerName: driveOwnerName(drive),
	}

	// personal Microsoft accounts receive opaque access tokens, so tenant metadata is best effort
	claims, err := microsoft.TokenClaims(req.Credential.AccessToken)
	if err != nil {
		logx.FromContext(ctx).Debug().Err(err).Msg("onedrive: access token is not a JWT, skipping tenant metadata")

		return metadata, nil
	}

	tenantID, ok := claims["tid"].(string)
	if !ok || tenantID == "" {
		return metadata, nil
	}

	upn, ok := claims["upn"].(string)
	if !ok || upn == "" {
		upn, _ = claims["preferred_username"].(string)
	}

	metadata.TenantID = tenantID
	metadata.Domain = ssoutils.EmailDomain(upn)

	return metadata, nil
}

// driveOwnerName returns the display name of the drive's owning user, empty when absent
func driveOwnerName(drive models.Driveable) string {
	owner := drive.GetOwner()
	if owner == nil || owner.GetUser() == nil {
		return ""
	}

	return lo.FromPtr(owner.GetUser().GetDisplayName())
}
