package microsoftteams

import (
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// ClientMicrosoftTeamsAPI identifies the Microsoft Graph API client.
	ClientMicrosoftTeamsAPI types.ClientName = "api"
)

// teamsClientDescriptors returns the client descriptors published by Microsoft Teams.
func teamsClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeMicrosoftTeams, ClientMicrosoftTeamsAPI, "Microsoft Graph API client", providerkit.TokenClientBuilder(auth.OAuthTokenFromPayload, nil))
}
