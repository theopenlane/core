package github

import (
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TypeGitHub identifies the GitHub provider
const TypeGitHub = types.ProviderType("github")

// Builder returns the GitHub provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeGitHub)
}
