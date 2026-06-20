package zitadel

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
)

// resolveInstallationMetadata derives Zitadel instance metadata from the persisted credential
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	if cred.Domain == "" || cred.Token == "" {
		return InstallationMetadata{}, false, nil
	}

	return InstallationMetadata{
		Domain: cred.Domain,
	}, true, nil
}