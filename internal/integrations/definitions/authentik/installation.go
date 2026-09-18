package authentik

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
	authentikSDK "goauthentik.io/api/v3"
)

// defaultBrandPageSize is the page size used when looking up the single default brand
const defaultBrandPageSize = int32(1)

// resolveInstallationMetadata derives Authentik instance metadata from the persisted credential
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	client, err := Client{}.Build(ctx, types.ClientBuildRequest{Credentials: req.Credentials})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error  during installation")

		return InstallationMetadata{}, false, err
	}

	apiClient := client.(*authentikSDK.APIClient)

	info, resp, err := apiClient.AdminApi.AdminSystemRetrieve(ctx).Execute()
	if resp != nil {
		_ = resp.Body.Close()
	}

	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error checking admin")

		return InstallationMetadata{}, false, ErrHealthCheckFailed
	}

	brandID, err := resolveDefaultBrandID(ctx, apiClient)
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	return InstallationMetadata{
		Brand:   info.GetBrand(),
		BrandID: brandID,
		Host:    info.GetHttpHost(),
		BaseURL: cred.BaseURL,
	}, true, nil
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
