package googleworkspace

import (
	"context"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientGoogleWorkspaceAPI identifies the Google Workspace HTTP API client.
	ClientGoogleWorkspaceAPI types.ClientName = "api"
)

// googleWorkspaceClientDescriptors returns the client descriptors published by Google Workspace.
func googleWorkspaceClientDescriptors() []types.ClientDescriptor {
	return auth.DefaultClientDescriptors(TypeGoogleWorkspace, ClientGoogleWorkspaceAPI, "Google Workspace REST API client", buildGoogleWorkspaceClient)
}

// buildGoogleWorkspaceClient constructs an authenticated Google Workspace API client.
func buildGoogleWorkspaceClient(_ context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	token, err := auth.OAuthTokenFromPayload(payload, string(TypeGoogleWorkspace))
	if err != nil {
		return nil, err
	}

	return auth.NewAuthenticatedClient(token, nil), nil
}
