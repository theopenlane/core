package githubapp

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// resolveInstallationMetadata derives GitHub App installation metadata from callback input
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	var metadata InstallationMetadata
	if err := jsonx.UnmarshalIfPresent(req.Input, &metadata); err != nil {
		return InstallationMetadata{}, false, ErrInstallationMetadataDecode
	}

	if metadata.InstallationID == "" {
		return InstallationMetadata{}, false, nil
	}

	return metadata, true, nil
}
