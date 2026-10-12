package oci

import (
	"context"

	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// verify reads the tenancy to confirm the API signing key, user, and region resolve and derives the installation metadata
func verify(ctx context.Context, _ types.ConnectionRequest[CredentialSchema], c Client) (InstallationMetadata, error) {
	_, err := c.Identity.GetTenancy(ctx, identity.GetTenancyRequest{
		TenancyId: lo.ToPtr(c.TenancyOCID),
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("region", c.Region).Msg("oci: healthcheck failed reading tenancy")
		return InstallationMetadata{}, ErrTenancyLookupFailed
	}

	return InstallationMetadata{
		TenancyOCID:     c.TenancyOCID,
		CompartmentOCID: c.CompartmentOCID,
		Region:          c.Region,
	}, nil
}
