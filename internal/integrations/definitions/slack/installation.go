package slack

import (
	"context"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/integrations/types"
)

// resolveInstallationMetadata derives Slack workspace metadata from whichever credential is bound
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

	return InstallationMetadata{
		TeamID:   authTest.TeamID,
		TeamName: authTest.Team,
	}, true, nil
}
