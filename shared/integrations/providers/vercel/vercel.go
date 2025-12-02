package vercel

import (
	"github.com/theopenlane/shared/integrations/providers"
	"github.com/theopenlane/shared/integrations/providers/apikey"
	"github.com/theopenlane/shared/integrations/types"
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
