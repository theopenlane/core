package okta

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
)

// resolveInstallationMetadata derives Okta tenant metadata from the persisted credential
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, _, err := oktaCredential.Resolve(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, ErrCredentialInvalid
	}

	if cred.OrgURL == "" {
		return InstallationMetadata{}, false, nil
	}

	return InstallationMetadata{
		OrgURL: cred.OrgURL,
	}, true, nil
}
