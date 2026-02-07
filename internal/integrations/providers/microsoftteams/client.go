package microsoftteams

import (
	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientMicrosoftTeamsAPI identifies the Microsoft Graph API client.
	ClientMicrosoftTeamsAPI types.ClientName = "api"
)

// teamsClientDescriptors returns the client descriptors published by Microsoft Teams.
func teamsClientDescriptors() []types.ClientDescriptor {
	return auth.DefaultClientDescriptors(TypeMicrosoftTeams, ClientMicrosoftTeamsAPI, "Microsoft Graph API client", auth.OAuthClientBuilder(nil))
}
