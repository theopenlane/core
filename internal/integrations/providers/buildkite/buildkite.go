package buildkite

import (
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TypeBuildkite identifies the Buildkite provider
const TypeBuildkite = types.ProviderType("buildkite")

// Builder returns the Buildkite provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeBuildkite)
}
