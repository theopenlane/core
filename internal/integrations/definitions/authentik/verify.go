package authentik

import (
	"context"

	authentikSDK "goauthentik.io/api/v3"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// defaultBrandPageSize is the page size used when looking up the single default brand
const defaultBrandPageSize = int32(1)

// verify derives the Authentik instance metadata through the admin system call
func verify(ctx context.Context, req types.ConnectionRequest[CredentialSchema], c *authentikSDK.APIClient) (InstallationMetadata, error) {
	info, resp, err := c.AdminApi.AdminSystemRetrieve(ctx).Execute()
	if resp != nil {
		_ = resp.Body.Close()
	}

	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error checking admin")

		return InstallationMetadata{}, ErrHealthCheckFailed
	}

	brandID, err := resolveDefaultBrandID(ctx, c)
	if err != nil {
		return InstallationMetadata{}, err
	}

	return InstallationMetadata{
		Brand:   info.GetBrand(),
		BrandID: brandID,
		Host:    info.GetHttpHost(),
		BaseURL: req.Credential.BaseURL,
	}, nil
}

// resolveDefaultBrandID fetches the immutable uuid of the instance's default brand
func resolveDefaultBrandID(ctx context.Context, apiClient *authentikSDK.APIClient) (string, error) {
	brands, resp, err := apiClient.CoreApi.CoreBrandsList(ctx).Default_(true).PageSize(defaultBrandPageSize).Execute()
	if resp != nil {
		_ = resp.Body.Close()
	}

	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error listing default brand")

		return "", ErrBrandFetchFailed
	}

	results := brands.GetResults()
	if len(results) == 0 {
		return "", ErrDefaultBrandMissing
	}

	return results[0].GetBrandUuid(), nil
}
