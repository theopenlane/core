package authentik

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
	authentikSDK "goauthentik.io/api/v3"
)

// resolveInstallationMetadata derives Authentik instance metadata from the persisted credential
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	if cred.Token == "" || cred.BaseURL == "" {
		return InstallationMetadata{}, false, nil
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

	return InstallationMetadata{
		Brand:   info.GetBrand(),
		Host:    info.GetHttpHost(),
		BaseURL: cred.BaseURL,
	}, true, nil
}
