package googleworkspace

import (
	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientGoogleWorkspaceAPI identifies the Google Workspace HTTP API client.
	ClientGoogleWorkspaceAPI types.ClientName = "api"
)

// googleWorkspaceClientDescriptors returns the client descriptors published by Google Workspace.
func googleWorkspaceClientDescriptors() []types.ClientDescriptor {
	return auth.DefaultClientDescriptors(TypeGoogleWorkspace, ClientGoogleWorkspaceAPI, "Google Workspace REST API client", auth.TokenClientBuilder(auth.OAuthTokenFromPayload, nil))
}
