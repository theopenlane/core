package googleworkspace

import (
	"context"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientGoogleWorkspaceAPI identifies the Google Workspace HTTP API client.
	ClientGoogleWorkspaceAPI types.ClientName = "api"
)

// googleWorkspaceClientDescriptors returns the client descriptors published by Google Workspace.
func googleWorkspaceClientDescriptors() []types.ClientDescriptor {
	return []types.ClientDescriptor{
		{
			Provider:     TypeGoogleWorkspace,
			Name:         ClientGoogleWorkspaceAPI,
			Description:  "Google Workspace REST API client",
			Build:        buildGoogleWorkspaceClient,
			ConfigSchema: map[string]any{"type": "object"},
		},
	}
}

// buildGoogleWorkspaceClient constructs an authenticated Google Workspace API client.
func buildGoogleWorkspaceClient(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	token, err := helpers.OAuthTokenFromPayload(payload, string(TypeGoogleWorkspace))
	if err != nil {
		return nil, err
	}

	return helpers.NewAuthenticatedClient(token, nil), nil
}
