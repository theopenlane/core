package cloudflare

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
)

// resolveInstallationMetadata derives Cloudflare account metadata from installation user input
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	meta, err := resolveCredential(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	if meta.AccountID == "" {
		return InstallationMetadata{}, false, nil
	}

	return InstallationMetadata{
		AccountID: meta.AccountID,
	}, true, nil
}
