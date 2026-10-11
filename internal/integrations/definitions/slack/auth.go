package slack

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// disconnectWorkspace returns the OAuth disconnect hook that points the user at the Slack app management page
func disconnectWorkspace(cfg Config) func(context.Context, types.DisconnectRequest[slackCred]) (types.DisconnectResult, error) {
	return func(_ context.Context, req types.DisconnectRequest[slackCred]) (types.DisconnectResult, error) {
		teamID, err := disconnectTeamID(req)
		if err != nil {
			return types.DisconnectResult{}, err
		}

		details, err := jsonx.ToRawMessage(disconnectDetails{
			TeamID: teamID,
		})
		if err != nil {
			return types.DisconnectResult{}, ErrInstallationMetadataEncode
		}

		return types.DisconnectResult{
			RedirectURL: getManageURL(teamID, cfg.AppID),
			Message:     "Uninstall the Openlane Slack App in Slack to finish disconnecting this integration.",
			Details:     details,
		}, nil
	}
}

// disconnectTeamID extracts the team ID from the installation metadata
func disconnectTeamID(req types.DisconnectRequest[slackCred]) (string, error) {
	var metadata InstallationMetadata
	if err := jsonx.UnmarshalIfPresent(req.Integration.InstallationMetadata.Attributes, &metadata); err != nil {
		return "", ErrInstallationMetadataDecode
	}

	if metadata.TeamID == "" {
		return "", ErrTeamIDMissing
	}

	return metadata.TeamID, nil
}

// disconnectDetails holds the metadata returned when initiating a Slack disconnect
type disconnectDetails struct {
	// TeamID is the Slack team ID being disconnected
	TeamID string `json:"teamId,omitempty"`
}

func getManageURL(teamID, appID string) string {
	// fall back to the full app list
	if appID == "" {
		return fmt.Sprintf("https://app.slack.com/apps-manage/%s/integrations", teamID)
	}

	return fmt.Sprintf("https://app.slack.com/apps-manage/%s/integrations/profile/%s", teamID, appID)
}
