package googleworkspace

import (
	"github.com/theopenlane/common/integrations/types"
	"github.com/theopenlane/core/pkg/integrations/providers"
	"github.com/theopenlane/core/pkg/integrations/providers/oauth"
)

// TypeGoogleWorkspace identifies the Google Workspace provider
const TypeGoogleWorkspace = types.ProviderType("google_workspace")

// Builder returns the Google Workspace provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeGoogleWorkspace, oauth.WithOperations(googleWorkspaceOperations()))
}
