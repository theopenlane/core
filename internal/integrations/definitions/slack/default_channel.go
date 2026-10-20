package slack

import (
	"context"

	generated "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/utils/jsonx"
)

// storeDefaultChannel saves the selected channel as the installation's default channel user input
func storeDefaultChannel(ctx context.Context, req types.OperationRequest, resources []types.SelectedResource) error {
	input, err := userInput.Decode(req.Integration)
	if err != nil {
		return ErrUserInputDecode
	}

	input.DefaultChannel = ""
	if len(resources) > 0 {
		input.DefaultChannel = resources[0].ID
	}

	return userInput.Save(ctx, req.DB, req.Integration, input)
}

// installationDefaultChannel returns the channel an installation's messages post to when none is given, preferring the user's choice over the channel chosen during app install
func installationDefaultChannel(integration *generated.Integration) (string, error) {
	// runtime operations execute without an installation, so there is no default to read
	if integration == nil {
		return "", nil
	}

	input, err := userInput.Decode(integration)
	if err != nil {
		return "", ErrUserInputDecode
	}

	if input.DefaultChannel != "" {
		return input.DefaultChannel, nil
	}

	var metadata InstallationMetadata
	if err := jsonx.UnmarshalIfPresent(integration.InstallationMetadata.Attributes, &metadata); err != nil {
		return "", ErrInstallationMetadataDecode
	}

	return metadata.DefaultChannel, nil
}
