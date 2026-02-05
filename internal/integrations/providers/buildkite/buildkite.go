package buildkite

import (
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/apikey"
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
