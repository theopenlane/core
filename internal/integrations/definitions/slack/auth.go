package slack

import (
	"fmt"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// disconnectTeamID extracts the team ID from the installation metadata
func disconnectTeamID(req types.DisconnectRequest) (string, error) {
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
