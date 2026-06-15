package azureentraid

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
)

// resolveInstallationMetadata derives Azure Entra installation metadata from the persisted credential
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, _, err := entraTenantCredential.Resolve(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	if cred.TenantID == "" {
		return InstallationMetadata{}, false, nil
	}

	return InstallationMetadata{TenantID: cred.TenantID}, true, nil
}
