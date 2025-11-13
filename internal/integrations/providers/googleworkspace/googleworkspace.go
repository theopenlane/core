package googleworkspace

import (
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TypeGoogleWorkspace identifies the Google Workspace provider
const TypeGoogleWorkspace = types.ProviderType("google_workspace")

// Builder returns the Google Workspace provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeGoogleWorkspace, oauth.WithOperations(googleWorkspaceOperations()))
}
