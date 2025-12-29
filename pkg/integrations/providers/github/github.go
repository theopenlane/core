package github

import (
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/pkg/integrations/providers"
	"github.com/theopenlane/core/pkg/integrations/providers/oauth"
)

// TypeGitHub identifies the GitHub provider
const TypeGitHub = types.ProviderType("github")

// Builder returns the GitHub provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeGitHub, oauth.WithOperations(githubOperations()))
}
