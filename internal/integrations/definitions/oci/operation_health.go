package oci

import (
	"context"
	"encoding/json"

	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/logx"
)

// HealthCheck holds the result of an OCI health check
type HealthCheck struct {
	// TenancyName is the display name of the tenancy the credentials resolved to
	TenancyName string `json:"tenancyName,omitempty"`
	// HomeRegionKey is the region key of the tenancy home region
	HomeRegionKey string `json:"homeRegionKey,omitempty"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClientRequest(identityClient, func(ctx context.Context, request types.OperationRequest, client *identity.IdentityClient) (json.RawMessage, error) {
		return h.Run(ctx, request.Credentials, client)
	})
}

// Run reads the tenancy to verify the API signing key, user, and region all resolve
func (HealthCheck) Run(ctx context.Context, credentials types.CredentialBindings, c *identity.IdentityClient) (json.RawMessage, error) {
	meta, err := resolveCredential(credentials)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("oci: error attempting to resolve credentials")
		return nil, err
	}

	res, err := c.GetTenancy(ctx, identity.GetTenancyRequest{
		TenancyId: lo.ToPtr(meta.TenancyOCID),
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("region", meta.Region).Msg("oci: healthcheck failed reading tenancy")
		return nil, ErrTenancyLookupFailed
	}

	return providerkit.EncodeResult(HealthCheck{
		TenancyName:   lo.FromPtr(res.Name),
		HomeRegionKey: lo.FromPtr(res.HomeRegionKey),
	}, ErrResultEncode)
}
