package github

import (
	"github.com/theopenlane/shared/integrations/providers"
	"github.com/theopenlane/shared/integrations/providers/oauth"
	"github.com/theopenlane/shared/integrations/types"
)

// TypeGitHub identifies the GitHub provider
const TypeGitHub = types.ProviderType("github")

// Builder returns the GitHub provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeGitHub, oauth.WithOperations(githubOperations()))
}
