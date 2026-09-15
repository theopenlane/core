package githubapp

import (
	"context"
	"strconv"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// resolveInstallationMetadata derives GitHub App installation metadata from the bound credential first, then from
// callback input, falling back to the metadata already stored on the installation when neither carries an
// installation id
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, ok, err := gitHubAppCredential.Resolve(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	if ok && cred.InstallationID != 0 {
		return InstallationMetadata{
			InstallationID:   strconv.FormatInt(cred.InstallationID, 10),
			OrganizationName: cred.OrganizationName,
		}, true, nil
	}

	var metadata InstallationMetadata
	if err := jsonx.UnmarshalIfPresent(req.Input, &metadata); err != nil {
		return InstallationMetadata{}, false, ErrInstallationMetadataDecode
	}

	if metadata.InstallationID != "" {
		return metadata, true, nil
	}

	var stored InstallationMetadata
	if err := jsonx.UnmarshalIfPresent(req.Integration.InstallationMetadata.Attributes, &stored); err != nil {
		return InstallationMetadata{}, false, ErrInstallationMetadataDecode
	}

	if stored.InstallationID == "" {
		return InstallationMetadata{}, false, nil
	}

	return stored, true, nil
}
