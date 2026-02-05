package microsoftteams

import (
	"context"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientMicrosoftTeamsAPI identifies the Microsoft Graph API client.
	ClientMicrosoftTeamsAPI types.ClientName = "api"
)

// teamsClientDescriptors returns the client descriptors published by Microsoft Teams.
func teamsClientDescriptors() []types.ClientDescriptor {
	return helpers.DefaultClientDescriptors(TypeMicrosoftTeams, ClientMicrosoftTeamsAPI, "Microsoft Graph API client", buildMicrosoftTeamsClient)
}

// buildMicrosoftTeamsClient constructs an authenticated Microsoft Graph client.
func buildMicrosoftTeamsClient(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	token, err := helpers.OAuthTokenFromPayload(payload, string(TypeMicrosoftTeams))
	if err != nil {
		return nil, err
	}

	return helpers.NewAuthenticatedClient(token, nil), nil
}
