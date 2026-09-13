package githubapp

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// resolveInstallationMetadata derives GitHub App installation metadata from callback input, falling back to the metadata already stored on the installation when the input carries no installation id
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
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
