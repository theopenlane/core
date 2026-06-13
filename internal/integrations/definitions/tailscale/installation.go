package tailscale

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
)

// resolveInstallationMetadata derives Tailscale tailnet metadata from the installation credential
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	if cred.ClientID == "" {
		return InstallationMetadata{}, false, nil
	}

	return InstallationMetadata{
		ClientID: cred.ClientID,
	}, true, nil
}
