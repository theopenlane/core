package buildkite

import (
	"github.com/theopenlane/shared/integrations/providers"
	"github.com/theopenlane/shared/integrations/providers/apikey"
	"github.com/theopenlane/shared/integrations/types"
)

// TypeBuildkite identifies the Buildkite provider
const TypeBuildkite = types.ProviderType("buildkite")

// Builder returns the Buildkite provider builder
func Builder() providers.Builder {
	return apikey.Builder(
		TypeBuildkite,
		apikey.WithTokenField("apiToken"),
		apikey.WithOperations(buildkiteOperations()),
	)
}
