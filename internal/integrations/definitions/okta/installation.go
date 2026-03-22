package okta

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// resolveInstallationMetadata derives Okta tenant metadata from the persisted credential
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	var cred CredentialSchema
	if err := jsonx.UnmarshalIfPresent(req.Credential.Data, &cred); err != nil {
		return InstallationMetadata{}, false, ErrCredentialInvalid
	}

	if cred.OrgURL == "" {
		return InstallationMetadata{}, false, nil
	}

	return InstallationMetadata{
		OrgURL: cred.OrgURL,
	}, true, nil
}
