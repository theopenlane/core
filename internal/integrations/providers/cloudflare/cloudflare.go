package cloudflare

import (
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/apikey"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TypeCloudflare identifies the Cloudflare provider
const TypeCloudflare = types.ProviderType("cloudflare")

// Builder returns the Cloudflare provider builder
func Builder() providers.Builder {
	return apikey.Builder(
		TypeCloudflare,
		apikey.WithTokenField("apiToken"),
		apikey.WithOperations(cloudflareOperations()),
	)
}
