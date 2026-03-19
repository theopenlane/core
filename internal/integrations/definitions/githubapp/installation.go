package githubapp

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// ResolveInstallationMetadata derives GitHub App installation metadata from callback input.
func ResolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (types.IntegrationInstallationMetadata, error) {
	if len(req.Input) == 0 {
		return types.IntegrationInstallationMetadata{}, nil
	}

	var metadata InstallationMetadata
	if err := jsonx.UnmarshalIfPresent(req.Input, &metadata); err != nil {
		return types.IntegrationInstallationMetadata{}, ErrInstallationMetadataDecode
	}

	if metadata.InstallationID == "" {
		return types.IntegrationInstallationMetadata{}, nil
	}

	attributes, err := jsonx.ToRawMessage(metadata)
	if err != nil {
		return types.IntegrationInstallationMetadata{}, ErrAuthProviderDataEncode
	}

	return types.IntegrationInstallationMetadata{Attributes: attributes}, nil
}
