package okta

import (
	"context"
	"fmt"

	oktagosdk "github.com/okta/okta-sdk-golang/v6/okta"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// resolveInstallationMetadata derives Okta tenant metadata from the persisted credential,
// resolving the immutable org id from the org settings API
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, _, err := oktaCredential.Resolve(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, ErrCredentialInvalid
	}

	if cred.OrgURL == "" {
		return InstallationMetadata{}, false, nil
	}

	built, err := Client{}.Build(ctx, types.ClientBuildRequest{Credentials: req.Credentials})
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	org, _, err := built.(*oktagosdk.APIClient).OrgSettingGeneralAPI.GetOrgSettings(ctx).Execute()
	if err != nil {
		return InstallationMetadata{}, false, fmt.Errorf("%w: %w", ErrOrgSettingsFetchFailed, err)
	}

	return InstallationMetadata{
		OrgURL: cred.OrgURL,
		OrgID:  org.GetId(),
	}, true, nil
}
