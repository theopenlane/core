package azuresecuritycenter

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// resolveInstallationMetadata derives Azure subscription metadata from the persisted credential
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	var cred CredentialSchema
	if err := jsonx.UnmarshalIfPresent(req.Credential.Data, &cred); err != nil {
		return InstallationMetadata{}, false, ErrCredentialInvalid
	}

	if cred.TenantID == "" && cred.ClientID == "" && cred.SubscriptionID == "" {
		return InstallationMetadata{}, false, nil
	}

	return InstallationMetadata{
		TenantID:       cred.TenantID,
		ClientID:       cred.ClientID,
		SubscriptionID: cred.SubscriptionID,
	}, true, nil
}
