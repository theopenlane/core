package slack

import (
	"context"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// resolveInstallationMetadata derives Slack workspace metadata from the persisted OAuth credential
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	var cred slackCred
	if err := jsonx.UnmarshalIfPresent(req.Credential.Data, &cred); err != nil {
		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	if cred.AccessToken == "" {
		return InstallationMetadata{}, false, ErrOAuthTokenMissing
	}

	authTest, err := slackgo.New(cred.AccessToken).AuthTestContext(ctx)
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
