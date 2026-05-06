package slack

import (
	"context"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// resolveInstallationMetadata derives Slack workspace metadata from whichever credential is bound
// and merges any provider input (for example, a user-selected default channel) supplied at install time
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	token, err := resolveAccessToken(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	authTest, err := slackgo.New(token).AuthTestContext(ctx)
	if err != nil {
		return InstallationMetadata{}, false, ErrAuthTestFailed
	}

	if authTest.TeamID == "" && authTest.Team == "" {
		return InstallationMetadata{}, false, nil
	}

	var input InstallationInput
	if err := jsonx.UnmarshalIfPresent(req.Input, &input); err != nil {
		return InstallationMetadata{}, false, ErrInstallationInputDecode
	}

	return InstallationMetadata{
		TeamID:         authTest.TeamID,
		TeamName:       authTest.Team,
		DefaultChannel: input.DefaultChannel,
	}, true, nil
}
