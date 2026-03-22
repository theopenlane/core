package azureentraid

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// resolveInstallationMetadata derives Azure Entra installation metadata from the persisted credential
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	var cred entraIDCred
	if err := jsonx.UnmarshalIfPresent(req.Credential.Data, &cred); err != nil {
		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	if cred.TenantID == "" {
		return InstallationMetadata{}, false, nil
	}

	return InstallationMetadata{
		TenantID: cred.TenantID,
	}, true, nil
}
