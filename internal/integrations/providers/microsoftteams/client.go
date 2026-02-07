package microsoftteams

import (
	"context"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientMicrosoftTeamsAPI identifies the Microsoft Graph API client.
	ClientMicrosoftTeamsAPI types.ClientName = "api"
)

// teamsClientDescriptors returns the client descriptors published by Microsoft Teams.
func teamsClientDescriptors() []types.ClientDescriptor {
	return auth.DefaultClientDescriptors(TypeMicrosoftTeams, ClientMicrosoftTeamsAPI, "Microsoft Graph API client", buildMicrosoftTeamsClient)
}

// buildMicrosoftTeamsClient constructs an authenticated Microsoft Graph client.
func buildMicrosoftTeamsClient(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	token, err := auth.OAuthTokenFromPayload(payload, string(TypeMicrosoftTeams))
	if err != nil {
		return nil, err
	}

	return auth.NewAuthenticatedClient(token, nil), nil
}
