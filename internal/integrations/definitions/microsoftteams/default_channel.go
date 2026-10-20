package microsoftteams

import (
	"context"

	generated "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// storeDefaultChannel saves the selected team and channel as the installation's default destination user input
func storeDefaultChannel(ctx context.Context, req types.OperationRequest, resources []types.SelectedResource) error {
	input, err := userInput.Decode(req.Integration)
	if err != nil {
		return ErrUserInputDecode
	}

	input.DefaultTeamID = ""
	input.DefaultChannel = ""

	if len(resources) > 0 {
		input.DefaultTeamID = resources[0].ParentID
		input.DefaultChannel = resources[0].ID
	}

	return userInput.Save(ctx, req.DB, req.Integration, input)
}

// installationUserInput returns the installation's user input, empty for runtime operations that execute without an installation
func installationUserInput(integration *generated.Integration) (UserInput, error) {
	if integration == nil {
		return UserInput{}, nil
	}

	input, err := userInput.Decode(integration)
	if err != nil {
		return UserInput{}, ErrUserInputDecode
	}

	return input, nil
}
