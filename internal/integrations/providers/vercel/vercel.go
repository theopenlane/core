package vercel

import (
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/apikey"
)

// TypeVercel identifies the Vercel provider
const TypeVercel = types.ProviderType("vercel")

// Builder returns the Vercel provider builder
func Builder() providers.Builder {
	return apikey.Builder(
		TypeVercel,
		apikey.WithTokenField("apiToken"),
		apikey.WithOperations(vercelOperations()),
	)
}
